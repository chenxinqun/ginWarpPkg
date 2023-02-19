package httpDiscover

import (
	"encoding/json"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/businessCodex"
	"github.com/chenxinqun/ginWarpPkg/datax/etcdx"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/httpClient"
	"github.com/chenxinqun/ginWarpPkg/httpx/mux"
	"math/rand"
	"net/http"
	httpURL "net/url"
	"time"

	"github.com/chenxinqun/ginWarpPkg/convert"
	"github.com/chenxinqun/ginWarpPkg/loggerx"
	"go.uber.org/zap"
)

const (
	// 默认重试次数3次
	defaultRetryCount = 3
	// 默认请求超时时间60
	defaultTimeOut = 60
)

type ServiceClient struct {
	timeout time.Duration
	// 基于etcd的服务发现与注册接口
	etcdRepo etcdx.Repo
	// 基于zap的logger
	logger     *zap.Logger
	retryCount int
	retryDelay int
	// 负载均衡器, 只需要传一个进来, 多传无效
	balancer []etcdx.BalancerFunc
}

type Resource struct {
	EtcdAddrs  []string
	Env        string
	ServerName string
	// 超时时间, 单位秒
	Timeout int
	// 重试次数
	RetryCount int
	// 重试间隔时间, 单位秒. 会自动加上[10,100)毫秒的随机数, 同时重试时间会指数级增加.
	RetryDelay int
}

func New(rs ...Resource) *ServiceClient {
	// Resource 可以不传, 会设置默认值. 如果传请只传一个, 多传无效
	var r Resource
	if len(rs) > 0 {
		r = rs[0]
	} else {
		r = Resource{}
	}
	if r.RetryCount <= 0 {
		r.RetryCount = defaultRetryCount
	}
	// 默认超时时间
	if r.Timeout <= 0 {
		r.Timeout = defaultTimeOut
	}
	timeout := time.Duration(r.Timeout) * time.Second
	var ret = new(ServiceClient)
	ret.logger = loggerx.Default()
	ret.etcdRepo = etcdx.Default()
	ret.timeout = timeout
	ret.retryCount = r.RetryCount
	ret.retryDelay = r.RetryDelay

	return ret
}

func shouldRetry(httpCode int) bool {
	switch httpCode {
	case
		// 请求超时则重试
		http.StatusRequestTimeout:

		return true

	default:
		return false
	}
}

func (s *ServiceClient) serviceHealth(service etcdx.ServiceInfo) bool {
	url := service.Url()
	healthUrl := fmt.Sprintf("%s%s", url, service.HealthPath)
	// 快速健康检查
	ttl := service.TTL / 2
	if ttl < 1 {
		ttl = 1
	} else if ttl > 5 {
		ttl = 5
	}
	healthResp, httpCode, e := httpClient.GetJson(healthUrl, httpURL.Values{}, httpClient.WithTTL(time.Duration(ttl)*time.Second))
	if httpCode != http.StatusOK || e != nil {
		s.logger.Error("健康检查请求", zap.Error(e),
			zap.String("name", service.Name),
			zap.String("url", healthUrl),
			zap.Int("status", httpCode))
		return false
	}
	respData := &businessCodex.Response{}
	e = json.Unmarshal(healthResp, &respData)
	if e != nil {
		s.logger.Error("健康检查返回值解析错误", zap.Error(e),
			zap.String("name", service.Name),
			zap.String("url", healthUrl),
			zap.Int("status", httpCode),
			zap.ByteString("response", healthResp),
		)
		return false
	}
	healthData := respData.Data.(map[string]interface{})
	var health bool
	for k, v := range service.HealthVerdict {
		hv, ok := healthData[k]
		if !ok {
			continue
		}
		check, ok := hv.(string)
		if !ok {
			continue
		}
		health = check == v
		if health {
			break
		}
	}
	if !health {
		s.logger.Error("健康检查返回状态为不健康",
			zap.String("name", service.Name),
			zap.String("url", healthUrl),
			zap.Int("status", httpCode),
			zap.ByteString("response", healthResp),
		)
		return false
	}
	return true
}

type RetryVerify func(body []byte) (shouldRetry bool)

func (s *ServiceClient) retryRequest(ctx mux.Context, serviceName string, urlPath, scheme string, requestFunc httpClient.RequestFunc,
	requestData interface{}, retryVerify ...RetryVerify) (body []byte, httpCode int, err error) {
	defer func() {
		if httpCode == 0 {
			err = errno.Errorf("%s 服务找不到, 或者请求 %s 没有响应", serviceName, urlPath)
		}
	}()
	if scheme == "" {
		scheme = "http"
	}
	// 获得一个随机打乱了顺序的服务列表
	serviceArray := s.etcdRepo.GetServiceArray(etcdx.ServiceInfo{Name: serviceName, Scheme: scheme}, s.balancer...)
	// 如果没拿到服务列表, 重新发现一次服务, 再去拿
	if len(serviceArray) == 0 {
		_ = s.etcdRepo.Discover(s.etcdRepo.GetRepo().Service.Val)
		serviceArray = s.etcdRepo.GetServiceArray(etcdx.ServiceInfo{Name: serviceName, Scheme: scheme}, s.balancer...)
	}
	handlers := make([]httpClient.OptionHandler, 0)
	header := ctx.GinContext().Request.Header
	for k, _ := range header {
		handlers = append(handlers, httpClient.WithHeader(k, header.Get(k)))
	}

	handlers = append(handlers, []httpClient.OptionHandler{
		httpClient.WithLogger(s.logger), httpClient.WithTTL(s.timeout), httpClient.WithTrace(ctx.Trace()),
		httpClient.WithHeader(mux.UserID, fmt.Sprintf("%d", ctx.UserID())),
		httpClient.WithHeader(mux.UserName, fmt.Sprintf("%s", ctx.UserName())),
		httpClient.WithHeader(mux.RoleType, fmt.Sprintf("%d", ctx.RoleType())),
		httpClient.WithHeader(mux.TenantID, fmt.Sprintf("%d", ctx.TenantID())),
		httpClient.WithHeader(mux.IsAdmin, fmt.Sprintf("%v", ctx.IsAdmin())),
	}...)

	// 遍历这个服务, 先做健康检查再请求, 如果一遍过就跳出循环, 如果一遍不过, 就重试
	for _, service := range serviceArray {
		// 挑选一个健康的服务
		if !s.serviceHealth(service) {
			// 删除不可用的服务
			s.etcdRepo.GetRepo().DelServiceList(service.Name, service.ID)
			continue
		}
		// 合成请求的完整路径
		url := service.Url() + urlPath
		// 发起请求
		body, httpCode, err = requestFunc(url, requestData, handlers...)
		loggerx.Default().Info("请求服务", zap.String("service name", serviceName), zap.String("url", url), zap.String("request func", fmt.Sprintf("%T", requestFunc)))
		// 进入重试模式
		for k := 0; k < s.retryCount; k++ {
			var needRetry bool
			// 进行通用校验, 校验状态码
			needRetry = shouldRetry(httpCode)
			// 指数级重试时间
			retryTime := time.Second * time.Duration(s.retryDelay*(k+1))
			rand.Seed(time.Now().UnixNano())
			// 增加10-100毫秒的重试随机数
			randInt := rand.Intn(100)
			if randInt < 10 {
				randInt = 10
			}
			retryTime += time.Millisecond * time.Duration(randInt)
			// 如果从验证码判断, 需要重试
			if needRetry {
				time.Sleep(retryTime)
				continue
			} else {
				// 如果通过状态码判断不需要重试, 则进入body验证, 验证返回值
				for _, retry := range retryVerify {
					needRetry = retry(body)
					if needRetry {
						break
					}
				}
				// 如果专用验证判断, 需要重试, 则走这里
				if needRetry {
					time.Sleep(retryTime)
					continue
				} else {
					// 不需要重试, 直接打断循环
					break
				}
			}
		}
		return body, httpCode, err
	}
	return body, httpCode, err
}

func (s *ServiceClient) GetJson(ctx mux.Context, serviceName string, urlPath string, params interface{}, result interface{}) (code int, err error) {
	// serviceName 服务名称, urlPath URL路径, params 传一个结构体进来
	urlParams, err := convert.StructToQuery(params)
	if err != nil {
		return code, err
	}
	body, code, err := s.retryRequest(ctx, serviceName, urlPath, "", httpClient.GetJson, urlParams)
	if err == nil {
		if result != nil && body != nil {
			r := NewResult(result)
			if r.Code != 0 {
				return code, r
			}
			return code, json.Unmarshal(body, r)
		}
	}
	return code, err
}

func (s *ServiceClient) PostJson(ctx mux.Context, serviceName string, urlPath string, params interface{}, result interface{}) (code int, err error) {
	// serviceName 服务名称, urlPath URL路径, params 传一个结构体进来
	jsonBs, err := convert.StructToJSON(params)
	if err != nil {
		return code, err
	}
	body, code, err := s.retryRequest(ctx, serviceName, urlPath, "", httpClient.PostJSON, jsonBs)
	if err == nil {
		if result != nil && body != nil {
			r := NewResult(result)
			if r.Code != 0 {
				return code, r
			}
			return code, json.Unmarshal(body, r)
		}
	}
	return code, err
}

func (s *ServiceClient) PutJson(ctx mux.Context, serviceName string, urlPath string, params interface{}, result interface{}) (code int, err error) {
	// serviceName 服务名称, urlPath URL路径, params 传一个结构体进来
	jsonBs, err := convert.StructToJSON(params)
	if err != nil {
		return code, err
	}
	body, code, err := s.retryRequest(ctx, serviceName, urlPath, "", httpClient.PutJSON, jsonBs)
	if err == nil {
		if result != nil && body != nil {
			r := NewResult(result)
			if r.Code != 0 {
				return code, r
			}
			return code, json.Unmarshal(body, r)
		}
	}
	return code, err
}

func (s *ServiceClient) DeleteJson(ctx mux.Context, serviceName string, urlPath string, params interface{}, result interface{}) (code int, err error) {
	// serviceName 服务名称, urlPath URL路径, params 传一个结构体进来
	urlParams, err := convert.StructToQuery(params)
	if err != nil {
		return code, err
	}
	body, code, err := s.retryRequest(ctx, serviceName, urlPath, "", httpClient.DeleteJson, urlParams)
	if err == nil {
		if result != nil && body != nil {
			r := NewResult(result)
			if r.Code != 0 {
				return code, r
			}
			return code, json.Unmarshal(body, r)
		}
	}
	return code, err
}

func (s *ServiceClient) PatchJson(ctx mux.Context, serviceName string, urlPath string, params interface{}, result interface{}) (code int, err error) {
	// serviceName 服务名称, urlPath URL路径, params 传一个结构体进来
	jsonBs, err := convert.StructToJSON(params)
	if err != nil {
		return code, err
	}
	body, code, err := s.retryRequest(ctx, serviceName, urlPath, "", httpClient.PatchJSON, jsonBs)
	if err == nil {
		if result != nil && body != nil {
			r := NewResult(result)
			if r.Code != 0 {
				return code, r
			}
			return code, json.Unmarshal(body, r)
		}
	}
	return code, err
}

func NewResult(data interface{}) *Result {
	return &Result{Data: data}
}

type Result struct {
	Code int         `json:"code"` // 业务码
	Msg  string      `json:"msg"`  // 描述信息
	Data interface{} `json:"data"` // 返回值
}

func (r Result) Error() string {
	return fmt.Sprintf("[%d] %s", r.Code, r.Msg)
}

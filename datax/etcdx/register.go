package etcdx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/idGen"
	"github.com/chenxinqun/ginWarpPkg/sysx/environment"
	"github.com/chenxinqun/ginWarpPkg/sysx/shutdown"
	"strings"

	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ServiceInfo struct {
	// 一个uuid4, 存储在etcd中时, 作为key的后缀来使用.
	ID string `json:"ID"`
	// 命名空间, 用以区分环境.
	Namespace string `toml:"Namespace" json:"Namespace"`
	// 获取etcd域名
	K8sDomain string `toml:"K8sDomain" json:"K8sDomain"`
	// 服务名称.
	Name string `toml:"Name" json:"Name"`
	// 用来在etcd中标识这个key是一个服务而不是一个配置
	Prefix string `toml:"Prefix" json:"Prefix"`
	Addr   string `toml:"Addr" json:"Addr"`
	Port   int    `toml:"Port" json:"Port"`
	TTL    int64  `toml:"TTL" json:"TTL"`
	// 服务的协议, http, https, grpc
	Scheme string `toml:"Scheme" json:"Scheme"`
	// http用
	HealthPath string `toml:"HealthPath" json:"HealthPath"`
	// grpc用
	HealthMethod string `toml:"HealthMethod" json:"HealthMethod"`
	// 健康结果
	HealthVerdict map[string]string `toml:"HealthVerdict" json:"HealthVerdict"`
	Meta          map[string]string `toml:"Meta" json:"Meta"`
	// 自定义的api版本, 如v1, v2这样子.
	ApiVersion string `toml:"ApiVersion" json:"ApiVersion"`
	// etcd中的记录版本, 纯数字, 从1开始, 每修改一次则+1, 最大为1000
	RecordVersion int64
	// 轮询时的标记
	PollingCount int64
	// ProjectVersion 项目版本(注册中心用) ProjectVersion = "v2.0"
	ProjectVersion string `toml:"ProjectVersion" json:"ProjectVersion"`

	// ProjectName 项目名称(注册中心用) ProjectName = "xdr"
	ProjectName string `toml:"ProjectName" json:"ProjectName"`
	Listen      string `toml:"Listen" json:"Listen"`
}

func (i ServiceInfo) Url() string {
	var (
		ret  string
		addr string
	)
	if len(i.K8sDomain) > 0 || strings.Contains(i.Addr, "svc.cluster.local") {
		addr = fmt.Sprintf("%s.%s", i.Name, strings.TrimPrefix(i.K8sDomain, "."))
	} else {
		addr = fmt.Sprintf("%s:%v", i.Addr, i.Port)
	}
	if i.Scheme == "grpc" {
		ret = addr
	} else {
		ret = fmt.Sprintf("%s://%s", i.Scheme, addr)
	}
	return ret
}

func (i *ServiceInfo) SetID() error {
	if i.ID == "" {
		i.ID = idGen.NewUUID4()
		return nil
	} else {
		return errno.Errorf("设置失败, ID已经存在: %v", i.ID)
	}
}

func (i *ServiceInfo) SetPrefix() {
	prefix := fmt.Sprintf("/Service/%s/", i.Namespace)
	businessPlatform := environment.BusinessPlatform()
	if businessPlatform != "" {
		prefix = fmt.Sprintf("/%s%s", businessPlatform, prefix)
	}
	// 去除头尾多余的斜杆
	prefix = strings.TrimLeft(prefix, "/")
	prefix = strings.TrimRight(prefix, "/")
	i.Prefix = fmt.Sprintf("/%s/", prefix)
}

func (i *ServiceInfo) SetDefaultTTL() {
	if i.TTL <= 0 {
		i.TTL = 10
	}
}

type ServiceRegister struct {
	Key string      `json:"register_key"`
	Val ServiceInfo `json:"register_val"`
}

func (r *ServiceRegister) SetKey() {
	// 去除头尾多余的斜杆
	prefix := strings.TrimLeft(r.Val.Prefix, "/")
	prefix = strings.TrimRight(prefix, "/")
	r.Key = fmt.Sprintf("/%s/%s/%s", prefix, r.Val.Name, r.Val.ID)
}

var schemeMaping map[string]string

func init() {
	schemeMaping = map[string]string{"http": "支持", "https": "支持", "grpc": "支持"}
}

func initService(service *ServiceInfo) error {
	_ = service.SetID()
	service.SetPrefix()
	service.SetDefaultTTL()
	if _, ok := schemeMaping[service.Scheme]; !ok {
		keys := make([]string, 0, len(schemeMaping))
		for k := range schemeMaping {
			keys = append(keys, k)
		}
		return errno.NewError(fmt.Sprintf("Scheme 只支持 %v, 得到 %s", keys, service.Scheme))
	}
	return nil
}

// Register 使用etcd租约模式注册服务, meta可以不传, 如果传的话只传一个, 多传无效.
func (e *DB) Register(service *ServiceInfo) error {
	// 如果已经注册过了, 则不再往下执行
	if e.registered {
		return nil
	}
	err := initService(service)
	if err != nil {
		return err
	}
	register := ServiceRegister{Val: *service}
	register.SetKey()
	ctx, cancel := e.TimeOutCtx(int(service.TTL))
	defer cancel()
	resp, err := e.client.Grant(ctx, service.TTL)
	if err != nil {
		return err
	}
	val, err := json.Marshal(register.Val)
	_, err = e.client.Put(ctx, register.Key, string(val), clientV3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	leaseRespChan, err := e.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		//switch err {
		//case context.Canceled:
		//   log.Fatalf("ctx is canceled by another routine: %v", err)
		//case context.DeadlineExceeded:
		//   log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
		//case rpctypes.ErrEmptyKey:
		//   log.Fatalf("client-side error: %v", err)
		//default:
		//   log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
		//}
		return err
	}
	if register.Val.Scheme == "grpc" {
		e.GrpcService = register
	} else {
		e.Service = register
	}

	e.leaseID = resp.ID
	e.keepAliveChan = leaseRespChan
	go e.ListenKeepAlive()
	e.registered = true
	return nil
}

// ListenKeepAlive 监听续租情况.
func (e *DB) ListenKeepAlive() {
	for leaseKeepResp := range e.keepAliveChan {
		e.Logger.Debug("检查续约状态", zap.String("结果", fmt.Sprintf("续约成功: %v\n", leaseKeepResp)))
	}
	e.Logger.Info("检查续约状态", zap.String("结果", "停止续约"))
	shutdown.Default().CloseManual()
	e.Logger.Fatal("etcd服务注册异常, 注销服务, 退出程序")
}

// Deregister 注销服务.
func (e *DB) Deregister() error {
	ctx, cancel := e.TimeOutCtx(int(e.Service.Val.TTL))
	defer cancel()
	if _, err := e.client.Revoke(ctx, e.leaseID); err != nil {
		return err
	}
	e.registered = false
	_, err := e.client.Delete(ctx, e.Service.Key)
	return err
}

func (e *DB) ThisService() ServiceRegister {
	return e.Service
}

func (e *DB) IsRegistered() bool {
	return e.registered
}

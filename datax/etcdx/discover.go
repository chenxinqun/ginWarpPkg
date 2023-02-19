package etcdx

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ServiceListing map[string]map[string][]ServiceRegister
type BalancerFunc func([]ServiceRegister) interface{}

// Discover 服务发现, 核心参数是 ServiceInfo.Namespace, ServiceInfo.ProjectName, ServiceInfo.ProjectVersion, ServiceInfo.Version 这四个
func (e *DB) Discover(service ServiceInfo) error {
	service.SetPrefix()
	service.SetDefaultTTL()
	// 设置前缀 通过 ServiceInfo.Namespace, ServiceInfo.ProjectName, ServiceInfo.ProjectVersion, ServiceInfo.Version 这四个参数
	service.SetPrefix()
	ctx, cancel := e.TimeOutCtx(int(service.TTL))
	defer cancel()
	resp, err := e.client.Get(ctx, service.Prefix, clientV3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		err = e.SetServiceListing(kv.Value)
		if err != nil {
			e.Logger.Error("初始化服务发现时报错", zap.Error(err), zap.String("key", string(kv.Key)), zap.String("value", string(kv.Value)))
		}
	}
	if !e.discovered {
		go e.Watcher(service.Prefix, serviceWatcher)
		e.discovered = true
	}
	return err
}

func serviceWatcher(er Repo, wresp clientV3.WatchResponse) {
	for _, ev := range wresp.Events {
		switch ev.Type {
		case mvccpb.PUT: //修改或者新增
			msg := "服务变更"
			if ev.Kv.Version == 1 {
				msg = "新增服务"
			}
			er.GetRepo().Logger.Info(msg, zap.String("key", string(ev.Kv.Key)), zap.String("value", string(ev.Kv.Value)))
			err := er.SetServiceListing(ev.Kv.Value)
			if err != nil {
				er.GetRepo().Logger.Error(fmt.Sprintf("%s同步到内存时报错", msg), zap.Error(err), zap.String("key", string(ev.Kv.Key)), zap.String("value", string(ev.Kv.Value)))
			}
		case mvccpb.DELETE: //删除
			er.GetRepo().Logger.Info("服务注销", zap.String("key", string(ev.Kv.Key)), zap.String("value", string(ev.Kv.Value)))
			er.DelServiceListing(string(ev.Kv.Key))
		}
	}
}

func GetNameID(etcdKey string) (name, id string) {
	k := strings.Split(etcdKey, "/")
	sk := k[:len(k)-1]
	name = sk[len(sk)-1]
	id = k[len(k)-1]
	return
}

// SetServiceListing 新增服务地址
func (e *DB) SetServiceListing(val []byte) error {
	if val == nil {
		return nil
	}
	if string(val) == "" {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	service := ServiceInfo{}
	err := json.Unmarshal(val, &service)
	if err != nil {
		return err
	}
	sr := ServiceRegister{Val: service}
	sr.SetKey()
	sval := make(map[string][]ServiceRegister)
	if e.serviceListing[service.Name] == nil {
		e.serviceListing[service.Name] = sval
	}
	if e.serviceListing[service.Name][service.Scheme] == nil {
		e.serviceListing[service.Name][service.Scheme] = make([]ServiceRegister, 0)
	}
	e.serviceListing[service.Name][service.Scheme] = append(e.serviceListing[service.Name][service.Scheme], sr)
	return err
}

// DelServiceListing 删除服务地址
func (e *DB) DelServiceListing(key string) {
	name, id := GetNameID(key)
	e.DelServiceList(name, id)
}

func (e *DB) DelServiceList(name, id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Service.Val.Scheme == "" {
		e.Service.Val.Scheme = "http"
	}
	newList := make([]ServiceRegister, 0)
	serviceList := e.serviceListing[name][e.Service.Val.Scheme]
	for i := 0; i < len(serviceList); i++ {
		if !(serviceList[i].Val.ID == id) {
			newList = append(newList, serviceList[i])
		}
	}
	e.serviceListing[name][e.Service.Val.Scheme] = newList
}

func (e *DB) GetServiceListing() ServiceListing {
	return e.serviceListing
}

func RandomPolicy(l []ServiceRegister) interface{} {
	rand.Seed(time.Now().UnixNano())
	return l[rand.Intn(len(l))].Val
}

func RandomPermPolicy(l []ServiceRegister) interface{} {
	rand.Seed(time.Now().UnixNano())
	randIndex := rand.Perm(len(l))
	ret := make([]ServiceInfo, 0, len(l))
	for _, v := range randIndex {
		ret = append(ret, l[v].Val)
	}
	return ret
}

func MetaPermPolicy(meta map[string]string) BalancerFunc {
	return func(l []ServiceRegister) interface{} {
		ret := make([]ServiceInfo, 0)
		if rl, ok := RandomPermPolicy(l).([]ServiceInfo); ok {
			for _, val := range rl {
				for k, v := range meta {
					if val.Meta[k] == v {
						ret = append(ret, val)
					}
				}
			}
		}
		return ret
	}
}

func MetaPolicy(meta map[string]string) BalancerFunc {
	return func(l []ServiceRegister) interface{} {
		if ret, ok := MetaPermPolicy(meta)(l).([]ServiceInfo); ok {
			if len(ret) > 0 {
				return ret[0]
			}
		}
		return nil
	}
}

// GetService 一个负载均衡的服务获取, 默认为随机策略, 如果传了ServiceInfo.Meta参数, 则使用Meta匹配策略来寻找服务, 传进来的BalancerFunc会失效.
func (e *DB) GetService(s ServiceInfo, balancer ...BalancerFunc) ServiceInfo {
	var b BalancerFunc
	if s.Meta != nil {
		b = MetaPolicy(s.Meta)
	} else if len(balancer) == 1 {
		b = balancer[0]
	} else {
		b = RandomPolicy
	}
	if m, ok := e.serviceListing[s.Name]; ok {
		if l, ok := m[s.Scheme]; ok {
			if ret, ok := b(l).(ServiceInfo); ok {
				return ret
			}
		}
	}
	return ServiceInfo{}
}

// GetServiceArray 获取一个按负载俊航策略排序的服务列表, 默认为随机策略, 如果传了ServiceInfo.Meta参数, 则使用Meta匹配策略来寻找服务, 传进来的BalancerFunc会失效
func (e *DB) GetServiceArray(s ServiceInfo, balancer ...BalancerFunc) []ServiceInfo {
	var b BalancerFunc
	if s.Meta != nil {
		b = MetaPermPolicy(s.Meta)
	} else if len(balancer) == 1 {
		b = balancer[0]
	} else {
		b = RandomPermPolicy
	}
	if m, ok := e.serviceListing[s.Name]; ok {
		if l, ok := m[s.Scheme]; ok {
			if ret, ok := b(l).([]ServiceInfo); ok {
				return ret
			}
		}
	}
	return nil
}

func (e *DB) IsDiscovered() bool {
	return e.discovered
}

package etcdx

import "github.com/chenxinqun/ginWarpPkg/loggerx"

func CreateSpecialEtcd(addr ...string) Repo {
	l := loggerx.CreateSpecialLogger()
	cfg := Info{
		Endpoints: addr,
	}
	repo, _ := New(cfg, l)
	return repo
}

func CreateSpecialServiceInfo(dev, serviceName string) *ServiceInfo {
	ret := &ServiceInfo{Namespace: dev, Name: serviceName}
	_ = initService(ret)
	return ret
}

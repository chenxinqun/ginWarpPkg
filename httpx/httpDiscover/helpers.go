package httpDiscover

import (
	"github.com/chenxinqun/ginWarpPkg/datax/etcdx"
	"github.com/chenxinqun/ginWarpPkg/loggerx"
)

func CreateSpecialHttpClient(rs ...Resource) *ServiceClient {
	var r Resource
	if len(rs) > 0 {
		r = rs[0]
	}
	if r.ServerName == "" {
		r.ServerName = "test"
	}
	if r.Env == "" {
		r.Env = "dev"
	}
	if len(r.EtcdAddrs) == 0 {
		r.EtcdAddrs = []string{"127.0.0.1:2379"}
	}
	// 初始化
	etcdx.CreateSpecialEtcd(r.EtcdAddrs...)
	service := etcdx.CreateSpecialServiceInfo(r.Env, r.ServerName)
	_ = etcdx.Default().Discover(*service)
	loggerx.CreateSpecialLogger()
	return New(rs...)
}

package etcdx

import (
	"context"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var _ Repo = (*DB)(nil)

type Repo interface {
	GetRepo() *DB
	GetConn() *clientV3.Client
	GetServiceListing() ServiceListing
	GetConfigs() ConfigStringMap
	TimeOutCtx(ttl int) (context.Context, context.CancelFunc)
	Watcher(prefix string, watcher WatcherFunc)
	Close() error

	Register(service *ServiceInfo) error
	ListenKeepAlive()
	Deregister() error
	ThisService() ServiceRegister
	IsRegistered() bool

	Discover(service ServiceInfo) error
	SetServiceListing(val []byte) error
	DelServiceListing(key string)
	GetService(s ServiceInfo, balancer ...BalancerFunc) ServiceInfo
	GetServiceArray(s ServiceInfo, balancer ...BalancerFunc) []ServiceInfo
	IsDiscovered() bool

	Configuring(conf ConfigInfo, setFunc SetConfFunc, watcher ...WatcherFunc) error
	SetConfig(key string, val []byte, listing ...ConfigStringMap)
}

var defaultRepo Repo

func Default() Repo {
	return defaultRepo
}

func New(cfg Info, logger *zap.Logger) (Repo, error) {
	var repo Repo
	var (
		cli *clientV3.Client
		err error
	)
	conf := clientV3.Config{ //nolint:exhaustivestruct
		Endpoints:            cfg.Endpoints,
		DialTimeout:          time.Duration(cfg.DialTimeout) * time.Second,
		DialKeepAliveTimeout: cfg.DialKeepAliveTimeout,
		MaxCallSendMsgSize:   cfg.MaxCallSendMsgSize,
		MaxCallRecvMsgSize:   cfg.MaxCallRecvMsgSize,
		Username:             cfg.Username,
		Password:             cfg.Password,
		RejectOldCluster:     cfg.RejectOldCluster,
		PermitWithoutStream:  cfg.PermitWithoutStream,
		TLS:                  cfg.TLS,
	}
	cli, err = clientV3.New(conf)
	repo = &DB{ //nolint:exhaustivestruct
		client:         cli,
		serviceListing: make(ServiceListing),
		configListing:  make(ConfigStringMap),
		Logger:         logger,
	}
	// 第一次初始化设置为默认连接
	if defaultRepo == nil {
		defaultRepo = repo
	}

	return repo, err
}

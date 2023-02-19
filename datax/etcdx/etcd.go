package etcdx

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type Info struct {
	// Endpoints is a list of URLs.
	Endpoints []string `toml:"Endpoints" json:"Endpoints"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout int64 `toml:"DialTimeout" json:"DialTimeout"`

	// DialKeepAliveTimeout is the time that the client waits for a response for the
	// keep-alive probe. If the response is not received in this time, the connection is closed.
	DialKeepAliveTimeout time.Duration `toml:"DialKeepAliveTimeout" json:"DialKeepAliveTimeout"`

	// MaxCallSendMsgSize is the client-side request send limit in bytes.
	// If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
	// Make sure that "MaxCallSendMsgSize" < server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallSendMsgSize int `toml:"MaxCallSendMsgSize" json:"MaxCallSendMsgSize"`

	// MaxCallRecvMsgSize is the client-side response receive limit.
	// If 0, it defaults to "math.MaxInt32", because range response can
	// easily exceed request send limits.
	// Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallRecvMsgSize int `toml:"MaxCallRecvMsgSize" json:"MaxCallRecvMsgSize"`

	// Username is a user name for authentication.
	Username string `toml:"Username" json:"Username"`

	// Password is a password for authentication.
	Password string `toml:"pwd" json:"pwd"`

	// RejectOldCluster when set will refuse to create a client against an outdated cluster.
	RejectOldCluster bool `toml:"RejectOldCluster" json:"RejectOldCluster"`

	// PermitWithoutStream when set will allow client to send keepalive pings to server without any active streams(RPCs).
	PermitWithoutStream bool   `toml:"PermitWithoutStream" json:"PermitWithoutStream"`
	TLSCa               string `toml:"TLSCa" json:"TLSCa"`
	TLSCert             string `toml:"TLSCert" json:"TLSCert"`
	TLSCertKey          string `toml:"TLSCertKey" json:"TLSCertKey"`

	// TLS holds the client secure credentials, if any.
	TLS *tls.Config
}

type WatcherFunc func(Repo, clientV3.WatchResponse)

type DB struct {
	// 已经运行过服务发现
	discovered bool
	// 已经运行过服务注册
	registered     bool
	ConfigInfo     ConfigInfo
	Service        ServiceRegister
	GrpcService    ServiceRegister
	leaseID        clientV3.LeaseID
	keepAliveChan  <-chan *clientV3.LeaseKeepAliveResponse
	client         *clientV3.Client //etcd client
	serviceListing ServiceListing   //服务清单
	configListing  ConfigStringMap
	Logger         *zap.Logger
	mu             sync.Mutex
}

func (e *DB) GetRepo() *DB {
	return e
}

func (e *DB) GetConn() *clientV3.Client {
	return e.client
}

func (e *DB) TimeOutCtx(ttl int) (ctx context.Context, cancel context.CancelFunc) {
	if ttl <= 0 {
		ttl = 5
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Duration(ttl)*time.Second)
	return ctx, cancel
}

func (e *DB) Watcher(prefix string, watcherFunc WatcherFunc) {
	rch := e.client.Watch(context.Background(), prefix, clientV3.WithPrefix())
	for wresp := range rch {
		watcherFunc(e, wresp)
	}
}

func (e *DB) Close() error {
	_ = e.Deregister()
	return e.client.Close()
}

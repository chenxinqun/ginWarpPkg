package mongox

import (
	"context"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonoptions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Info struct {
	Addrs    []string `toml:"Addrs" json:"Addrs"`
	Database string   `toml:"Database" json:"Database"`
	Username string   `toml:"Username" json:"Username"`
	Password string   `toml:"pwd" json:"pwd"`
	MaxConn  int      `toml:"MaxConn" json:"MaxConn"`
	TimeOut  int      `toml:"TimeOut" json:"TimeOut"` // 超时时间, 单位秒
}

type Option func(*option)

type Trace = trace.T

type option struct {
	Trace *trace.Trace
	Mongo *trace.Mongo
}

var _ Repo = (*Mongo)(nil)

type Repo interface {
	Close()
	// GetDatabase 初始化一个新的 mongo-database
	GetDatabase(name string) DataBase

	// InitDefaultDatabase 初始化 当前应用默认 mongo-database
	InitDefaultDatabase(name string) error

	// DefaultDatabase 获取当前应用 默认 mongo-database
	DefaultDatabase() DataBase
}

type Mongo struct {
	//client *mgo.Session
	client    *mongo.Client
	timeout   time.Duration
	defaultDb DataBase
	mlock     sync.RWMutex
}

var defaultRepo Repo

func Default() Repo {
	return defaultRepo
}

func New(cfg Info) (Repo, error) {
	var repo Repo
	// client, err := mongoConnect(cfg)
	timeout := time.Second * time.Duration(cfg.TimeOut)
	client, err := dialMongo(cfg.Addrs, cfg.Username, cfg.Password, timeout)
	if err != nil {
		return nil, err
	}
	repo = &Mongo{
		client:  client,
		timeout: timeout,
	}
	if defaultRepo == nil {
		defaultRepo = repo
	}
	if len(cfg.Database) > 0 {
		repo.InitDefaultDatabase(cfg.Database)
	}
	return repo, nil
}

func (c *Mongo) GetDatabase(name string) DataBase {
	dataBase := NewDataBase(c.client, name)

	return dataBase
}

func (c *Mongo) InitDefaultDatabase(name string) error {
	c.mlock.Lock()
	defer c.mlock.Unlock()
	if c.defaultDb != nil {
		return errno.Errorf("default database has been Initialized")
	}
	database := c.GetDatabase(name)
	c.defaultDb = database
	return nil
}

func (c *Mongo) DefaultDatabase() DataBase {
	c.mlock.RLock()
	defer c.mlock.RUnlock()
	if c.defaultDb == nil {
		panic(errno.Errorf("default database not init, that is BUG"))
	}
	return c.defaultDb
}

// dialMongo will connection single server
func dialMongo(addr []string, user, passwd string, timeout time.Duration) (*mongo.Client, error) {
	defopts := options.Client()
	defopts.SetHosts(addr)
	if len(user) > 0 && len(passwd) > 0 {
		defopts.SetAuth(options.Credential{
			AuthMechanismProperties: nil,
			Username:                user,
			Password:                passwd,
			PasswordSet:             true,
		})
	}
	if timeout > 0 {
		defopts.SetServerSelectionTimeout(timeout)
		defopts.SetConnectTimeout(timeout)
		defopts.SetSocketTimeout(timeout)
	}

	/*
	 * primary:默认参数,只从主节点上进行读取操作;
	 * primaryPreferred:大部分从主节点上读取数据,只有主节点不可用时从secondary节点读取数据.
	 * secondary
	 * :只从secondary节点上进行读取操作,存在的问题是secondary节点的数据会比primary节点数据旧.
	 * secondaryPreferred:优先从secondary节点进行读取操作,secondary节点不可用时从主节点读取数据;
	 * nearest:不管是主节点、secondary节点,从网络延迟最低的节点上读取数据.
	 */
	defopts.SetReadPreference(readpref.PrimaryPreferred())
	opts := append(globalMongoOptions(), defopts)
	cli, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	var cancel context.CancelFunc
	ctx := context.Background()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	if e := cli.Connect(ctx); e != nil {
		return nil, e
	}
	if err := cli.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return nil, errors.Wrap(err, "ping mongo err")
	}

	return cli, nil
}

// 全局选项
func globalMongoOptions() []*options.ClientOptions {
	// opts := options.Client()
	return []*options.ClientOptions{
		options.Client().SetRegistry(NewCustomRegistry()),
	}
}

// NewCustomRegistry create custom registry
func NewCustomRegistry() *bsoncodec.Registry {
	rb := bsoncodec.NewRegistryBuilder()
	bsoncodec.DefaultValueEncoders{}.RegisterDefaultEncoders(rb)
	bsoncodec.DefaultValueDecoders{}.RegisterDefaultDecoders(rb)
	//
	rb.RegisterCodec(reflect.TypeOf(time.Time{}), bsoncodec.NewTimeCodec(
		bsonoptions.TimeCodec().SetUseLocalTimeZone(true)))

	return rb.Build()
}

// Close mongo client
func (c *Mongo) Close() {
	var cancel context.CancelFunc
	ctx := context.Background()
	if c.timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	_ = c.client.Disconnect(ctx)
}

// WithTrace 设置trace信息
func WithTrace(t Trace) Option {
	return func(opt *option) {
		if t != nil {
			opt.Trace = t.(*trace.Trace)
			opt.Mongo = new(trace.Mongo)
		}
	}
}

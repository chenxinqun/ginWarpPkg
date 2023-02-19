package redisx

import (
	"context"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/httpx/trace"
	"strings"
	"time"

	"github.com/chenxinqun/ginWarpPkg/loggerx"
	"github.com/chenxinqun/ginWarpPkg/timex"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DefaultTimeOut 默认超时时间10秒
const DefaultTimeOut = 10

type SentinelInfo struct {
	MasterName       string   `toml:"MasterName" json:"MasterName"`
	SentinelAddrs    []string `toml:"SentinelAddrs" json:"SentinelAddrs"`
	SentinelPassword string   `toml:"SentinelPassword" json:"SentinelPassword"`
}

type Info struct {
	Addr         string `toml:"Addr" json:"Addr"`
	Pass         string `toml:"Pass" json:"Pass"`
	Db           int    `toml:"Db" json:"Db"`
	MaxRetries   int    `toml:"MaxRetries" json:"MaxRetries"`
	PoolSize     int    `toml:"PoolSize" json:"PoolSize"`
	MinIDleConns int    `toml:"MinIDleConns" json:"MinIDleConns"`
	// 超时时间, 单位秒. 拨号超时, 读写超时采用同一个配置.
	TimeOut int64 `toml:"TimeOut" json:"TimeOut"`
	// redis哨兵相关配置
	Sentinel SentinelInfo `toml:"Sentinel" json:"Sentinel"`
}

type OptionHandler func(*option)

func IsNil(err error) bool {
	return strings.Contains(err.Error(), redis.Nil.Error())
}

type Trace = trace.T

type option struct {
	Trace *trace.Trace
	Redis *trace.Redis
}

func newOption() *option {
	return &option{}
}

var _ Repo = (*DB)(nil)

type Repo interface {
	Set(key, value string, ttl time.Duration, options ...OptionHandler) error
	Get(key string, options ...OptionHandler) (string, error)
	TTL(key string) (time.Duration, error)
	Expire(key string, ttl time.Duration) bool
	ExpireAt(key string, ttl time.Time) bool
	Del(key string, options ...OptionHandler) bool
	Exists(keys ...string) bool
	Incr(key string, options ...OptionHandler) int64
	Close() error
	GetConn() *redis.Client

	// Publish 发布消息
	Publish(channel string, message interface{}) (int64, error)
	// Subscribe 订阅消息
	Subscribe(handler SubHandler, channels ...string) error

	Keys(pattern string) *redis.StringSliceCmd
	Move(key string, db int) *redis.BoolCmd
	Sort(key string, sort *redis.Sort) *redis.StringSliceCmd
	SortStore(key string, store string, sort *redis.Sort) *redis.IntCmd
	SortInterfaces(key string, sort *redis.Sort) *redis.SliceCmd
	Decr(key string) *redis.IntCmd
	DecrBy(key string, decrement int64) *redis.IntCmd
	GetRange(key string, start int64, end int64) *redis.StringCmd
	GetSet(key string, value interface{}) *redis.StringCmd
	GetEx(key string, expiration time.Duration) *redis.StringCmd
	GetDel(key string) *redis.StringCmd
	IncrBy(key string, value int64) *redis.IntCmd
	IncrByFloat(key string, value float64) *redis.FloatCmd
	MGet(keys ...string) *redis.SliceCmd
	MSet(values ...interface{}) *redis.StatusCmd
	MSetNX(values ...interface{}) *redis.BoolCmd
	SetEX(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	GetBit(key string, offset int64) *redis.IntCmd
	SetBit(key string, offset int64, value int) *redis.IntCmd
	HDel(key string, fields ...string) *redis.IntCmd
	HExists(key string, field string) *redis.BoolCmd
	HGet(key string, field string) *redis.StringCmd
	HGetAll(key string) *redis.MapStringStringCmd
	HIncrBy(key string, field string, incr int64) *redis.IntCmd
	HIncrByFloat(key string, field string, incr float64) *redis.FloatCmd
	HKeys(key string) *redis.StringSliceCmd
	HLen(key string) *redis.IntCmd
	HMGet(key string, fields ...string) *redis.SliceCmd
	HSet(key string, values ...interface{}) *redis.IntCmd
	HMSet(key string, values ...interface{}) *redis.BoolCmd
	HSetNX(key string, field string, value interface{}) *redis.BoolCmd
	LIndex(key string, index int64) *redis.StringCmd
	LInsert(key string, op string, pivot interface{}, value interface{}) *redis.IntCmd
	LInsertBefore(key string, pivot interface{}, value interface{}) *redis.IntCmd
	LInsertAfter(key string, pivot interface{}, value interface{}) *redis.IntCmd
	LLen(key string) *redis.IntCmd
	LPop(key string) *redis.StringCmd
	LPopCount(key string, count int) *redis.StringSliceCmd
	LPos(key string, value string, a redis.LPosArgs) *redis.IntCmd
	LPosCount(key string, value string, count int64, a redis.LPosArgs) *redis.IntSliceCmd
	LPush(key string, values ...interface{}) *redis.IntCmd
	LPushX(key string, values ...interface{}) *redis.IntCmd
	LRange(key string, start int64, stop int64) *redis.StringSliceCmd
	LRem(key string, count int64, value interface{}) *redis.IntCmd
	LSet(key string, index int64, value interface{}) *redis.StatusCmd
	LTrim(key string, start int64, stop int64) *redis.StatusCmd
	RPop(key string) *redis.StringCmd
	RPopCount(key string, count int) *redis.StringSliceCmd
	RPopLPush(source string, destination string) *redis.StringCmd
	RPush(key string, values ...interface{}) *redis.IntCmd
	RPushX(key string, values ...interface{}) *redis.IntCmd
	LMove(source string, destination string, srcpos string, destpos string) *redis.StringCmd
	SAdd(key string, members ...interface{}) *redis.IntCmd
	SCard(key string) *redis.IntCmd
	SIsMember(key string, member interface{}) *redis.BoolCmd
	SMembers(key string) *redis.StringSliceCmd
	SPop(key string) *redis.StringCmd
	SPopN(key string, count int64) *redis.StringSliceCmd
	SRem(key string, members ...interface{}) *redis.IntCmd

	ZAddNX(key string, members ...redis.Z) *redis.IntCmd
	ZAddXX(key string, members ...redis.Z) *redis.IntCmd
	ZCard(key string) *redis.IntCmd
	ZRem(key string, members ...interface{}) *redis.IntCmd
	ZCount(key, min, max string) *redis.IntCmd
	ZLexCount(key, min, max string) *redis.IntCmd
	ZIncrBy(key string, increment float64, member string) *redis.FloatCmd
	ZInterStore(destination string, store *redis.ZStore) *redis.IntCmd
	ZInter(store *redis.ZStore) *redis.StringSliceCmd
	ZInterWithScores(store *redis.ZStore) *redis.ZSliceCmd
	ZMScore(key string, members ...string) *redis.FloatSliceCmd
	ZPopMax(key string, count ...int64) *redis.ZSliceCmd
	ZPopMin(key string, count ...int64) *redis.ZSliceCmd
}

type DB struct {
	CtxTimeOut int64
	Cfg        Info
	client     *redis.Client
	stopCh     chan struct{}
}

var defaultRepo Repo

func Default() Repo {
	return defaultRepo
}

func New(cfg Info) (Repo, error) {
	var repo Repo
	client, err := redisConnect(&cfg)
	if err != nil {
		return nil, err
	}

	repo = &DB{
		client:     client,
		Cfg:        cfg,
		CtxTimeOut: cfg.TimeOut + 1,
		stopCh:     make(chan struct{}),
	}
	if defaultRepo == nil {
		defaultRepo = repo
	}

	return repo, nil
}

func TimeOutCtx(timeout int64) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
	return ctx, cancel
}

func redisConnect(cfg *Info) (client *redis.Client, err error) {
	if cfg.TimeOut <= 0 {
		cfg.TimeOut = DefaultTimeOut
	}
	timeout := time.Duration(cfg.TimeOut) * time.Second
	if cfg.Sentinel.SentinelAddrs != nil {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.Sentinel.MasterName,
			SentinelAddrs:    cfg.Sentinel.SentinelAddrs,
			SentinelPassword: cfg.Sentinel.SentinelPassword,
			Password:         cfg.Pass,
			DB:               cfg.Db,
			MaxRetries:       cfg.MaxRetries,
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     cfg.MinIDleConns,
			WriteTimeout:     timeout,
			ReadTimeout:      timeout,
			DialTimeout:      timeout,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Pass,
			DB:           cfg.Db,
			MaxRetries:   cfg.MaxRetries,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIDleConns,
			WriteTimeout: timeout,
			ReadTimeout:  timeout,
			DialTimeout:  timeout,
		})
	}
	ctx, cancel := TimeOutCtx(cfg.TimeOut)
	defer cancel()
	if err = client.Ping(ctx).Err(); err != nil {
		return nil, errno.Wrap(err, "ping redis err")
	}

	return client, nil
}

func (c *DB) GetConn() *redis.Client {
	return c.client
}

// Set set some <key,value> into redis
func (c *DB) Set(key, value string, ttl time.Duration, options ...OptionHandler) error {
	ts := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timex.CSTLayoutString()
			opt.Redis.Handle = "set"
			opt.Redis.Key = key
			opt.Redis.Value = value
			opt.Redis.TTL = ttl.Minutes()
			opt.Redis.CostSeconds = time.Since(ts).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errno.Wrapf(err, "redis set key: %s err", key)
	}

	return nil
}

// Get get some key from redis
func (c *DB) Get(key string, options ...OptionHandler) (string, error) {
	ts := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timex.CSTLayoutString()
			opt.Redis.Handle = "get"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(ts).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", errno.Wrapf(err, "redis get key: %s err", key)
	}

	return value, nil
}

// TTL 获取某个key的过期时间
func (c *DB) TTL(key string) (time.Duration, error) {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return -1, errno.Wrapf(err, "redis get key: %s err", key)
	}

	return ttl, nil
}

// Expire 设置过期时间(多久之后过期)
func (c *DB) Expire(key string, ttl time.Duration) bool {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	ok, _ := c.client.Expire(ctx, key, ttl).Result()

	return ok
}

// ExpireAt 设置过期时间(在某一时刻过期)
func (c *DB) ExpireAt(key string, ttl time.Time) bool {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	ok, _ := c.client.ExpireAt(ctx, key, ttl).Result()

	return ok
}

func (c *DB) Exists(keys ...string) bool {
	if len(keys) == 0 {
		return true
	}
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	value, _ := c.client.Exists(ctx, keys...).Result()

	return value > 0
}

func (c *DB) Del(key string, options ...OptionHandler) bool {
	ts := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timex.CSTLayoutString()
			opt.Redis.Handle = "del"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(ts).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}

	if key == "" {
		return true
	}
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	value, _ := c.client.Del(ctx, key).Result()

	return value > 0
}

func (c *DB) Incr(key string, options ...OptionHandler) int64 {
	ts := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timex.CSTLayoutString()
			opt.Redis.Handle = "incr"
			opt.Redis.Key = key
			opt.Redis.CostSeconds = time.Since(ts).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()

	for _, f := range options {
		f(opt)
	}
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	value, _ := c.client.Incr(ctx, key).Result()

	return value
}

func (c *DB) Publish(channel string, message interface{}) (int64, error) {
	ts := time.Now()
	opt := newOption()
	defer func() {
		if opt.Trace != nil {
			opt.Redis.Timestamp = timex.CSTLayoutString()
			opt.Redis.Handle = "Publish"
			opt.Redis.Key = channel
			opt.Redis.CostSeconds = time.Since(ts).Seconds()
			opt.Trace.AppendRedis(opt.Redis)
		}
	}()
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	r := c.client.Publish(ctx, channel, message)
	return r.Result()
}

// Close close redis client
func (c *DB) Close() error {
	close(c.stopCh)
	return c.client.Close()
}

// WithTrace 设置trace信息
func WithTrace(t Trace) OptionHandler {
	return func(opt *option) {
		if t != nil {
			opt.Trace = t.(*trace.Trace)
			opt.Redis = new(trace.Redis)
		}
	}
}

func (c *DB) Keys(pattern string) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.Keys(ctx, pattern)
}
func (c *DB) Move(key string, db int) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.Move(ctx, key, db)
}
func (c *DB) Sort(key string, sort *redis.Sort) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.Sort(ctx, key, sort)
}
func (c *DB) SortStore(key string, store string, sort *redis.Sort) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SortStore(ctx, key, store, sort)
}
func (c *DB) SortInterfaces(key string, sort *redis.Sort) *redis.SliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SortInterfaces(ctx, key, sort)
}
func (c *DB) Decr(key string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.Decr(ctx, key)
}
func (c *DB) DecrBy(key string, decrement int64) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.DecrBy(ctx, key, decrement)
}
func (c *DB) GetRange(key string, start int64, end int64) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.GetRange(ctx, key, start, end)
}
func (c *DB) GetSet(key string, value interface{}) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.GetSet(ctx, key, value)
}
func (c *DB) GetEx(key string, expiration time.Duration) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.GetEx(ctx, key, expiration)
}
func (c *DB) GetDel(key string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.GetDel(ctx, key)
}
func (c *DB) IncrBy(key string, value int64) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.IncrBy(ctx, key, value)
}
func (c *DB) IncrByFloat(key string, value float64) *redis.FloatCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.IncrByFloat(ctx, key, value)
}
func (c *DB) MGet(keys ...string) *redis.SliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.MGet(ctx, keys...)
}
func (c *DB) MSet(values ...interface{}) *redis.StatusCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.MSet(ctx, values...)
}
func (c *DB) MSetNX(values ...interface{}) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.MSetNX(ctx, values...)
}
func (c *DB) SetEX(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SetEx(ctx, key, value, expiration)
}
func (c *DB) SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SetNX(ctx, key, value, expiration)
}
func (c *DB) GetBit(key string, offset int64) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.GetBit(ctx, key, offset)
}
func (c *DB) SetBit(key string, offset int64, value int) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SetBit(ctx, key, offset, value)
}
func (c *DB) HDel(key string, fields ...string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HDel(ctx, key, fields...)
}
func (c *DB) HExists(key string, field string) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HExists(ctx, key, field)
}
func (c *DB) HGet(key string, field string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HGet(ctx, key, field)
}
func (c *DB) HGetAll(key string) *redis.MapStringStringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HGetAll(ctx, key)
}
func (c *DB) HIncrBy(key string, field string, incr int64) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HIncrBy(ctx, key, field, incr)
}
func (c *DB) HIncrByFloat(key string, field string, incr float64) *redis.FloatCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HIncrByFloat(ctx, key, field, incr)
}
func (c *DB) HKeys(key string) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HKeys(ctx, key)
}
func (c *DB) HLen(key string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HLen(ctx, key)
}
func (c *DB) HMGet(key string, fields ...string) *redis.SliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HMGet(ctx, key, fields...)
}
func (c *DB) HSet(key string, values ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HSet(ctx, key, values...)
}
func (c *DB) HMSet(key string, values ...interface{}) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HMSet(ctx, key, values...)
}
func (c *DB) HSetNX(key string, field string, value interface{}) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.HSetNX(ctx, key, field, value)
}
func (c *DB) LIndex(key string, index int64) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LIndex(ctx, key, index)
}
func (c *DB) LInsert(key string, op string, pivot interface{}, value interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LInsert(ctx, key, op, pivot, value)
}
func (c *DB) LInsertBefore(key string, pivot interface{}, value interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LInsertBefore(ctx, key, pivot, value)
}
func (c *DB) LInsertAfter(key string, pivot interface{}, value interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LInsertAfter(ctx, key, pivot, value)
}
func (c *DB) LLen(key string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LLen(ctx, key)
}
func (c *DB) LPop(key string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPop(ctx, key)
}
func (c *DB) LPopCount(key string, count int) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPopCount(ctx, key, count)
}
func (c *DB) LPos(key string, value string, a redis.LPosArgs) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPos(ctx, key, value, a)
}
func (c *DB) LPosCount(key string, value string, count int64, a redis.LPosArgs) *redis.IntSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPosCount(ctx, key, value, count, a)
}
func (c *DB) LPush(key string, values ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPush(ctx, key, values...)
}
func (c *DB) LPushX(key string, values ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LPushX(ctx, key, values...)
}
func (c *DB) LRange(key string, start int64, stop int64) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LRange(ctx, key, start, stop)
}
func (c *DB) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LRem(ctx, key, count, value)
}
func (c *DB) LSet(key string, index int64, value interface{}) *redis.StatusCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LSet(ctx, key, index, value)
}
func (c *DB) LTrim(key string, start int64, stop int64) *redis.StatusCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LTrim(ctx, key, start, stop)
}
func (c *DB) RPop(key string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.RPop(ctx, key)
}
func (c *DB) RPopCount(key string, count int) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.RPopCount(ctx, key, count)
}
func (c *DB) RPopLPush(source string, destination string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.RPopLPush(ctx, source, destination)
}
func (c *DB) RPush(key string, values ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.RPush(ctx, key, values...)
}
func (c *DB) RPushX(key string, values ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.RPushX(ctx, key, values...)
}
func (c *DB) LMove(source string, destination string, srcpos string, destpos string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.LMove(ctx, source, destination, srcpos, destpos)
}
func (c *DB) SAdd(key string, members ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SAdd(ctx, key, members...)
}
func (c *DB) SCard(key string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SCard(ctx, key)
}
func (c *DB) SIsMember(key string, member interface{}) *redis.BoolCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SIsMember(ctx, key, member)
}
func (c *DB) SMembers(key string) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SMembers(ctx, key)
}
func (c *DB) SPop(key string) *redis.StringCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SPop(ctx, key)
}
func (c *DB) SPopN(key string, count int64) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SPopN(ctx, key, count)
}
func (c *DB) SRem(key string, members ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.SRem(ctx, key, members...)
}

func (c *DB) ZAddNX(key string, members ...redis.Z) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZAddNX(ctx, key, members...)
}

func (c *DB) ZAddXX(key string, members ...redis.Z) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZAddXX(ctx, key, members...)
}

func (c *DB) ZCard(key string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZCard(ctx, key)
}

func (c *DB) ZRem(key string, members ...interface{}) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZRem(ctx, key, members...)
}

func (c *DB) ZCount(key, min, max string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZCount(ctx, key, min, max)
}

func (c *DB) ZLexCount(key, min, max string) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZLexCount(ctx, key, min, max)
}

func (c *DB) ZIncrBy(key string, increment float64, member string) *redis.FloatCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZIncrBy(ctx, key, increment, member)
}

func (c *DB) ZInterStore(destination string, store *redis.ZStore) *redis.IntCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZInterStore(ctx, destination, store)
}

func (c *DB) ZInter(store *redis.ZStore) *redis.StringSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZInter(ctx, store)
}

func (c *DB) ZInterWithScores(store *redis.ZStore) *redis.ZSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZInterWithScores(ctx, store)
}

func (c *DB) ZMScore(key string, members ...string) *redis.FloatSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZMScore(ctx, key, members...)
}

func (c *DB) ZPopMax(key string, count ...int64) *redis.ZSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZPopMax(ctx, key, count...)
}

func (c *DB) ZPopMin(key string, count ...int64) *redis.ZSliceCmd {
	ctx, cancel := TimeOutCtx(c.Cfg.TimeOut)
	defer cancel()
	return c.client.ZPopMin(ctx, key, count...)
}

type SubHandler func(msg *redis.Message) error

func (c *DB) Subscribe(handler SubHandler, channels ...string) error {
	var needBreak bool
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		select {
		case <-c.stopCh:
			needBreak = true
			cancelFunc()
		}
	}()
	select {
	case <-ctx.Done():
		return nil
	default:
		go func() {
			for {
				if needBreak {
					break
				}
				cx, cancel := context.WithCancel(ctx)
				defer cancel()
				sub := c.client.Subscribe(cx, channels...)
				select {
				case msg := <-sub.Channel():
					e := handler(msg)
					if e != nil {
						loggerx.Default().Error("订阅消息处理时报错", zap.Any("msg", *msg), zap.Any("channels", channels), zap.Error(e))
					}
				case <-ctx.Done():
					needBreak = true
					_ = sub.Close()
					break
				case <-cx.Done():
					break
				}
				if e := cx.Err(); e != nil {
					loggerx.Default().Error("订阅消息处理时Context报错, 重新连接", zap.Any("channels", channels), zap.Error(e))
					cancel()
				}
			}
		}()
	}
	return nil
}

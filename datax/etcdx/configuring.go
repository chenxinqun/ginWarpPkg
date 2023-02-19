package etcdx

import (
	"encoding/json"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/sysx/environment"
	"strconv"
	"strings"

	"github.com/chenxinqun/ginWarpPkg/identify"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ConfigStringMap map[string]interface{}
type ConfigSlice []interface{}
type ConfigInfo ServiceInfo

type SetConfFunc func(Repo, clientV3.GetResponse)

func (c *ConfigInfo) SetPrefix() {
	prefix := fmt.Sprintf("/Config/%s/%s/", c.Namespace, c.Name)
	businessPlatform := environment.BusinessPlatform()
	if businessPlatform != "" {
		prefix = fmt.Sprintf("/%s%s", businessPlatform, prefix)
	}
	// 去除头尾多余的斜杆.
	prefix = strings.TrimLeft(prefix, "/")
	prefix = strings.TrimRight(prefix, "/")
	c.Prefix = fmt.Sprintf("/%s/", prefix)
}

func setConfFunc(e Repo, resp clientV3.GetResponse) {
	for _, kv := range resp.Kvs {
		e.SetConfig(string(kv.Key), kv.Value)
	}
}

// Configuring watcher 可以不传, 会有默认值. 如果要传的话, 只接受一个函数, 多传无效.
func (e *DB) Configuring(conf ConfigInfo, setFunc SetConfFunc, watcher ...WatcherFunc) error {
	conf.SetPrefix()
	e.ConfigInfo = conf
	// 监控配置变动的处理函数
	var wcr WatcherFunc
	if len(watcher) == 1 {
		wcr = watcher[0]
	} else {
		wcr = confWatcher
	}
	if setFunc == nil {
		setFunc = setConfFunc
	}

	ctx, cancel := e.TimeOutCtx(10)

	defer cancel()
	resp, err := e.client.Get(ctx, conf.Prefix, clientV3.WithPrefix())
	if err != nil {
		return err
	}
	// 从配置中心初始化配置
	setFunc(e, *resp)

	go e.Watcher(conf.Prefix, wcr)
	return err
}

// 监控配置变动的默认函数, 这个函数是一个范例. 真实项目中一般需要自定义.
// 配置文件, 默认只支持修改, 不支持删除, 因为删除配置项容易导致程序出问题, 尤其是使用的结构体做配置的时候.
// 配置文件修改的原则应该是只热更新与业务有关的配置项, 一般为开关类的配置以及短连接的地址配置. 数据库相关的需要连接池, 或者长连接配置一般不修改.
func confWatcher(er Repo, wresp clientV3.WatchResponse) {
	for _, ev := range wresp.Events {
		key := strings.TrimPrefix(string(ev.Kv.Key), er.GetRepo().ConfigInfo.Prefix)
		switch ev.Type {
		case mvccpb.PUT: //修改或者新增
			// 新增配置
			if ev.Kv.Version == 1 {
				er.GetRepo().Logger.Info("配置新增", zap.String(key, string(ev.Kv.Value)))
			} else {
				// 改动配置
				er.GetRepo().Logger.Info("配置改动", zap.String("key", key), zap.Any("原配置", er.GetConfigs()[key]), zap.String("新配置", string(ev.Kv.Value)))
			}
			er.SetConfig(key, ev.Kv.Value)
		}
	}
}

// SetConfig 新增配置项.
func (e *DB) SetConfig(key string, val []byte, listing ...ConfigStringMap) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var (
		value interface{}
		conf  ConfigStringMap
		v     ConfigStringMap
		vL    ConfigSlice
		err   error
	)
	sv := strings.TrimSpace(string(val))
	// 花括号开头, 转换成map
	if strings.HasPrefix(sv, "{") {
		v = make(ConfigStringMap)
		err = json.Unmarshal(val, &v)
		if err != nil {
			e.Logger.Error("etcd配置项 map[string]interface{} json序列化时出错", zap.String("key", key), zap.String("value", string(val)), zap.Error(err))
		} else {
			value = v
		}
		// 中括号开头, 转换为数组
	} else if strings.HasPrefix(sv, "[") {
		vL = make(ConfigSlice, 0)
		err = json.Unmarshal(val, &vL)
		if err != nil {
			e.Logger.Error("etcd配置项 []interface{} json序列化时出错", zap.String("key", key), zap.String("value", string(val)), zap.Error(err))
		} else {
			value = vL
		}
		// 不是花括号和中括号开头的处理方式
	} else {
		// 数字转换成int型
		if identify.IsDigit(sv) {
			sI, _ := strconv.Atoi(sv)
			value = sI
			// float 转换为 float64型
		} else if identify.IsFloat(sv) {
			sF, _ := strconv.ParseFloat(sv, 64)
			value = sF
			// bool转换为bool型
		} else if identify.IsBool(sv) {
			sB, _ := strconv.ParseBool(sv)
			value = sB
			// 否则原样保存为字符串
		} else {
			value = sv
		}
	}
	// 如果传了map进来, 用传进来的map存储数据
	if len(listing) == 1 {
		conf = listing[0]
		// 否则用默认map存储数据
	} else {
		conf = e.configListing
	}
	conf[key] = value
}

func (e *DB) GetConfigs() ConfigStringMap {
	return e.configListing
}

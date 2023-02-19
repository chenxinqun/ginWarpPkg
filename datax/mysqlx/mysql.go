package mysqlx

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/chenxinqun/ginWarpPkg/sysx/environment"
	"math/rand"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type Info struct {
	Read []struct {
		Addr string `toml:"Addr" json:"Addr"`
		User string `toml:"User" json:"User"`
		Pass string `toml:"Pass" json:"Pass"`
		Name string `toml:"Name" json:"Name"`
	} `toml:"Read" json:"Read"`

	Write []struct {
		Addr string `toml:"Addr" json:"Addr"`
		User string `toml:"User" json:"User"`
		Pass string `toml:"Pass" json:"Pass"`
		Name string `toml:"Name" json:"Name"`
	} `toml:"Write" json:"Write"`

	Base struct {
		MaxOpenConn     int           `toml:"MaxOpenConn" json:"MaxOpenConn"`
		MaxIDleConn     int           `toml:"MaxIDleConn" json:"MaxIDleConn"`
		ConnMaxLifeTime time.Duration `toml:"ConnMaxLifeTime" json:"ConnMaxLifeTime"` // 最大连接超时时间单位分钟
	} `toml:"Base"`
}

var _ Repo = (*DB)(nil)

type Repo interface {
	GetDb() *gorm.DB
	Close() error
}

type DB struct {
	Db *gorm.DB
}

var defaultRepo Repo

func Default() Repo {
	return defaultRepo
}

func New(cfg Info) (Repo, error) {
	var repo Repo
	base := cfg.Base
	write := cfg.Write[0]
	db, err := dbConnect(write.User, write.Pass, write.Addr, write.Name, base.MaxOpenConn, base.MaxIDleConn, base.ConnMaxLifeTime)
	if err != nil {
		return nil, err
	}

	// 根据MySQL热Read配置判断是否要开启读写分离
	if len(cfg.Read) > 0 {
		// 合成read库的连接地址
		readList := make([]string, 0)
		for _, r := range cfg.Read {
			readList = append(readList, Dsn(r.User, r.Pass, r.Addr, r.Name))
		}
		// 合成write库的连接地址
		writeList := make([]string, 0)
		for _, w := range cfg.Write {
			writeList = append(writeList, Dsn(w.User, w.Pass, w.Addr, w.Name))
		}
		// 获取写分分离插件
		resolver := readWriteResolver(readList, writeList, base.MaxOpenConn, base.MaxIDleConn, base.ConnMaxLifeTime)
		// 注册读写分离插件
		_ = db.Use(resolver)
	}

	// 注册链路追踪插件
	_ = db.Use(new(TracePlugin))

	if environment.Active() != nil {
		// 如果不是Pro和Pre环境, 开启db.Debug()模式
		if !environment.Active().IsPro() && !environment.Active().IsPre() {
			db.Logger = db.Logger.LogMode(logger.Info)
		}
	}

	repo = &DB{
		Db: db,
	}
	// 将第一次数据库连接设为默认值
	if defaultRepo == nil {
		defaultRepo = repo
	}
	return repo, nil
}

func (d *DB) GetDb() *gorm.DB {
	return d.Db
}

func (d *DB) Close() error {
	sqlDB, err := d.Db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func Dsn(user, pass, addr, dbName string) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		user,
		pass,
		addr,
		dbName,
		true,
		"Local")
	return dsn
}

type randomPolicy struct{}

func (randomPolicy) Resolve(connPools []gorm.ConnPool) gorm.ConnPool {
	rand.Seed(time.Now().UnixNano())
	return connPools[rand.Intn(len(connPools))]
}

func readWriteResolver(readList []string, writeList []string, maxOpenConn, maxIDleConn int, connMaxLifeTime time.Duration) *dbresolver.DBResolver {
	// 处理读的连接
	replicas := make([]gorm.Dialector, 0)
	for _, readDsn := range readList {
		replicas = append(replicas, mysql.Open(readDsn))
	}

	// 处理写的连接
	sources := make([]gorm.Dialector, 0)
	for _, writeDsn := range writeList {
		sources = append(sources, mysql.Open(writeDsn))
	}

	// 负载均衡器
	policy := new(randomPolicy)
	resolver := dbresolver.Register(dbresolver.Config{
		Sources:  sources,
		Replicas: replicas,
		Policy:   policy,
	}).
		// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
		SetMaxOpenConns(maxOpenConn).
		// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
		SetMaxIdleConns(maxIDleConn).
		// 设置最大连接超时
		SetConnMaxLifetime(time.Minute * connMaxLifeTime)
	return resolver

}

func dbConnect(user, pass, addr, dbName string, maxOpenConn, maxIDleConn int, connMaxLifeTime time.Duration) (*gorm.DB, error) {
	dsn := Dsn(user, pass, addr, dbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名(不使用复数形式单词)
		},
	})

	if err != nil {
		return nil, errno.Wrap(err, fmt.Sprintf("[mysql connection failed] Database DSN: %s", dsn))
	}

	// 获取原始SQL连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(maxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(maxIDleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * connMaxLifeTime)
	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	return db, nil
}

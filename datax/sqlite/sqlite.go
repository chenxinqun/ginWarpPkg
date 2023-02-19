package sqlite

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/sysx/environment"
	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type User struct {
	Name string
	Age  int
}

func Dsn(filename string) string {
	dsn := filename
	return dsn
}

type Info struct {
	FileName        string        `toml:"FileName" json:"FileName"`               // sqlite文件名
	MaxOpenConn     int           `toml:"MaxOpenConn" json:"MaxOpenConn"`         // 最大连接数
	MaxIDleConn     int           `toml:"MaxIDleConn" json:"MaxIDleConn"`         // 最大空闲连接数
	ConnMaxLifeTime time.Duration `toml:"ConnMaxLifeTime" json:"ConnMaxLifeTime"` // 最大连接超时时间, 单位分钟
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
	db, err := dbConnect(cfg.FileName, cfg.MaxOpenConn, cfg.MaxIDleConn, cfg.ConnMaxLifeTime)
	if err != nil {
		return nil, err
	}
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

func dbConnect(filename string, maxOpenConn, maxIDleConn int, connMaxLifeTime time.Duration) (*gorm.DB, error) {
	dsn := Dsn(filename)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名(不使用复数形式单词)
		},
		//Logger: loggerx.Default.LogMode(loggerx.Info), // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[sqlite connection failed] Database DSN: %s", dsn))
	}

	// 获取原始SQL连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接sqlite出现too many connections的错误。
	sqlDB.SetMaxOpenConns(maxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(maxIDleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * connMaxLifeTime)

	return db, nil
}

package postgresql

import (
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/sysx/environment"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type User struct {
	Name string
	Age  int
}

func Dsn(host, user, pass, dbName string, port uint16) string {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		host,
		user,
		pass,
		dbName,
		port,
	)
	return dsn
}

type Info struct {
	Host            string        `toml:"Host" json:"Host"`                       // "host:port", 必须使用postgresql的tcp端口, 不能使用http端口
	User            string        `toml:"User" json:"User"`                       // "用户名"
	Pass            string        `toml:"Pass" json:"Pass"`                       // 密码
	Name            string        `toml:"Name" json:"Name"`                       // 数据库名
	Port            uint16        `toml:"Port" json:"Port"`                       // 端口
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
	db, err := dbConnect(cfg.Host, cfg.User, cfg.Pass, cfg.Name, cfg.Port, cfg.MaxOpenConn, cfg.MaxIDleConn, cfg.ConnMaxLifeTime)
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

func dbConnect(host, user, pass, dbName string, port uint16, maxOpenConn, maxIDleConn int, connMaxLifeTime time.Duration) (*gorm.DB, error) {
	dsn := Dsn(host, user, pass, dbName, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名(不使用复数形式单词)
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[postgresql connection failed] Database DSN: %s", dsn))
	}

	// 获取原始SQL连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接postgresql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(maxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(maxIDleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * connMaxLifeTime)

	return db, nil
}

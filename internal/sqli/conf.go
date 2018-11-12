package sqli

import (
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	DefaultDBName = "db"
	HookSuffix    = "_with_hook"
)

type Conf struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MapperFunc      func(string) string // struct field name convert
	HookDisable     bool
}

func (c *Conf) Initialize(driverName string) *sqlx.DB {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 512
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 64
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = time.Minute * 15
	}
	if c.MapperFunc == nil {
		c.MapperFunc = snakecase
	}

	var db *sqlx.DB
	var err error
	if c.HookDisable {
		db, err = sqlx.Open(driverName, c.DSN)
	} else {
		// should register related driver in init func
		db, err = sqlx.Open(driverName+HookSuffix, c.DSN)
	}

	if err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(c.ConnMaxLifetime)
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.MapperFunc(c.MapperFunc)

	return db
}

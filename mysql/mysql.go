package mysql

import (
	"fmt"
	"sync"
	"time"

	// mysql driver
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var sqlPool = &struct {
	sync.RWMutex
	clients map[string]*Conf
}{
	clients: make(map[string]*Conf),
}

// Conf is sql conf
type Conf struct {
	DSN             string
	ConnMaxLifetime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	MapperFunc      func(string) string // struct tag convert
	db              *sqlx.DB
}

// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// each db should only register once
func Register(name string, conf Conf) error {
	if name == "" {
		name = "db"
	}

	sqlPool.Lock()
	defer sqlPool.Unlock()

	// check if DSN is valid
	_, err := mysql.ParseDSN(conf.DSN)
	if err != nil {
		return err
	}

	db, err := sqlx.Open("mysql", conf.DSN)
	if err != nil {
		return err
	}

	// set default 15min max conn, 512 max open, 64 max idl
	if conf.ConnMaxLifetime == 0 {
		conf.ConnMaxLifetime = time.Minute * 15
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxIdleConns = 512
	}
	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 64
	}

	db.SetConnMaxLifetime(conf.ConnMaxLifetime)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)

	// using snakecase default
	if conf.MapperFunc == nil {
		db.MapperFunc(snakecase)
	} else {
		db.MapperFunc(conf.MapperFunc)
	}

	conf.db = db

	sqlPool.clients[name] = &conf

	return nil
}

// Client returns mysql client
func Client(name string) (*sqlx.DB, error) {
	sqlPool.RLock()

	client, ok := sqlPool.clients[name]
	if ok {
		sqlPool.RUnlock()
		return client.db, nil
	}

	sqlPool.RUnlock()
	return nil, fmt.Errorf("db %q not registered", name)
}

// Deprecated, using DB() instead.
// Get return mysql client with given name or nil if not exist
func Get(name string) *sqlx.DB {
	cli, _ := Client(name)
	return cli
}

// DB is helper func returns default db
func DB(name ...string) *sqlx.DB {
	var cli *sqlx.DB
	if len(name) == 0 {
		cli, _ = Client("db")
	} else {
		cli, _ = Client(name[0])
	}
	return cli
}

// HealthCheck ping db
func HealthCheck() error {
	errs := make(map[string]error)

	sqlPool.RLock()

	for name, client := range sqlPool.clients {
		if err := client.db.Ping(); err != nil {
			errs[name] = err
		}
	}

	sqlPool.RUnlock()

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// Close closes all mysql conn
func Close() error {
	errs := make(map[string]error)

	sqlPool.Lock()

	for name, client := range sqlPool.clients {
		if err := client.db.Close(); err != nil {
			errs[name] = err
		}
	}

	sqlPool.Unlock()

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// MySQLErr try conver mysql err to *mysql.MySQLError
func MySQLErr(err error) *mysql.MySQLError {
	if err == nil {
		return nil
	}
	if e, ok := err.(*mysql.MySQLError); ok {
		return e
	}
	return nil
}

// IsDupErr check if mysql error is ER_DUP_ENTRY
// https://github.com/VividCortex/mysqlerr
func IsDupErr(err error) bool {
	e := MySQLErr(err)
	return e != nil && e.Number == 1062
}

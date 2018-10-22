package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gchaincl/sqlhooks"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var defaultDB = "db"

// alias
type (
	NullBool    = sql.NullBool
	NullInt64   = sql.NullInt64
	NullFloat64 = sql.NullFloat64
	NullString  = sql.NullString
	NullTime    = mysql.NullTime
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
	MapperFunc      func(string) string // struct field name convert
	db              *sqlx.DB
}

// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// each db should only register once
func Register(name string, conf Conf) error {
	if name == "" {
		name = defaultDB
	}

	// check if DSN is valid
	cfg, err := mysql.ParseDSN(conf.DSN)
	if err != nil {
		return err
	}

	var db *sqlx.DB

	if os.Getenv("mysql_hook_off") == "true" {
		db, err = sqlx.Open("mysql", cfg.FormatDSN())
	} else {
		sql.Register("mysql_hook", sqlhooks.Wrap(&mysql.MySQLDriver{}, &Hooks{}))
		db, err = sqlx.Open("mysql_hook", cfg.FormatDSN())
	}

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
		conf.MapperFunc = snakecase
	}
	db.MapperFunc(conf.MapperFunc)

	conf.db = db

	sqlPool.Lock()
	sqlPool.clients[name] = &conf
	sqlPool.Unlock()

	return nil
}

// Client returns mysql client, mostly, we use DB() func
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

// DB is helper func to get *sqlx.DB
func DB(name ...string) *sqlx.DB {
	var cli *sqlx.DB
	if len(name) == 0 {
		cli, _ = Client(defaultDB)
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

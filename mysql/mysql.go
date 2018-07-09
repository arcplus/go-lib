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
	db              *sqlx.DB
}

// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// each db should only register once
func Register(name string, conf Conf) error {
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

	db.SetConnMaxLifetime(conf.ConnMaxLifetime)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)

	conf.db = db

	sqlPool.clients[name] = &conf

	return nil
}

// Client returns mysql client
func Client(name string) (*sqlx.DB, error) {
	sqlPool.RLock()

	client, ok := sqlPool.clients[name]
	if !ok {
		sqlPool.RUnlock()
		return nil, fmt.Errorf("db %q not registered", name)
	}

	sqlPool.RUnlock()
	return client.db, nil
}

// HealthCheck ping sql
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

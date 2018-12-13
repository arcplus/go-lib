package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/gchaincl/sqlhooks"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/arcplus/go-lib/internal/sqli"
)

const driverName = "mysql"

func init() {
	sql.Register(driverName+sqli.HookSuffix, sqlhooks.Wrap(&mysql.MySQLDriver{}, &sqli.Hook{}))
}

// alias
type (
	NullBool    = sql.NullBool
	NullInt64   = sql.NullInt64
	NullFloat64 = sql.NullFloat64
	NullString  = sql.NullString
	NullTime    = mysql.NullTime
)

var pool = &struct {
	sync.RWMutex
	clients map[string]*sqlx.DB
}{
	clients: make(map[string]*sqlx.DB),
}

type Conf = sqli.Conf

// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// each db should only register once
func Register(name string, conf Conf) {
	db := conf.Initialize(driverName)

	pool.Lock()
	pool.clients[name] = db
	pool.Unlock()
}

// Client returns mysql client, mostly, we use DB() func
func Client(name string) (*sqlx.DB, error) {
	pool.RLock()

	db, ok := pool.clients[name]
	if ok {
		pool.RUnlock()
		return db, nil
	}

	pool.RUnlock()
	return nil, fmt.Errorf("mysql %q not registered", name)
}

// DB is helper func to get *sqlx.DB
func DB(name string) *sqlx.DB {
	cli, _ := Client(name)
	return cli
}

// HealthCheck ping db
func HealthCheck() error {
	errs := make(map[string]error)

	pool.RLock()

	for name, db := range pool.clients {
		if err := db.Ping(); err != nil {
			errs[name] = err
		}
	}

	pool.RUnlock()

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// Close closes all mysql conn
func Close() error {
	errs := make(map[string]error)

	pool.Lock()

	for name, db := range pool.clients {
		if err := db.Close(); err != nil {
			errs[name] = err
		}
	}

	pool.Unlock()

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

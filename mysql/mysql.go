package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/arcplus/go-lib/internal/sqli"
	"github.com/gchaincl/sqlhooks"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

var sqlPool = &struct {
	sync.RWMutex
	clients map[string]*sqlx.DB
}{
	clients: make(map[string]*sqlx.DB),
}

type Conf = sqli.Conf

// Register dsn format -> [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// each db should only register once
func Register(name string, conf Conf) {
	if name == "" {
		name = sqli.DefaultDBName
	}

	db := conf.Initialize(driverName)

	sqlPool.Lock()
	sqlPool.clients[name] = db
	sqlPool.Unlock()
}

// Client returns mysql client, mostly, we use DB() func
func Client(name string) (*sqlx.DB, error) {
	sqlPool.RLock()

	db, ok := sqlPool.clients[name]
	if ok {
		sqlPool.RUnlock()
		return db, nil
	}

	sqlPool.RUnlock()
	return nil, fmt.Errorf("database %q not registered", name)
}

// DB is helper func to get *sqlx.DB
func DB(name ...string) *sqlx.DB {
	var cli *sqlx.DB
	if len(name) == 0 {
		cli, _ = Client(sqli.DefaultDBName)
	} else {
		cli, _ = Client(name[0])
	}
	return cli
}

// HealthCheck ping db
func HealthCheck() error {
	errs := make(map[string]error)

	sqlPool.RLock()

	for name, db := range sqlPool.clients {
		if err := db.Ping(); err != nil {
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

	for name, db := range sqlPool.clients {
		if err := db.Close(); err != nil {
			errs[name] = err
		}
	}

	sqlPool.Unlock()

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

package pg

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/arcplus/go-lib/internal/sqli"
	"github.com/gchaincl/sqlhooks"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const driverName = "postgres"

func init() {
	sql.Register(driverName+sqli.HookSuffix, sqlhooks.Wrap(&pq.Driver{}, &sqli.Hook{}))
}

var pool = &struct {
	sync.RWMutex
	clients map[string]*sqlx.DB
}{
	clients: make(map[string]*sqlx.DB),
}

type Conf = sqli.Conf

func Register(name string, conf Conf) {
	if name == "" {
		name = sqli.DefaultDBName
	}

	db := conf.Initialize(driverName)

	pool.Lock()
	pool.clients[name] = db
	pool.Unlock()
}

// Client returns client, mostly, we use DB() func
func Client(name string) (*sqlx.DB, error) {
	pool.RLock()

	db, ok := pool.clients[name]
	if ok {
		pool.RUnlock()
		return db, nil
	}

	pool.RUnlock()
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

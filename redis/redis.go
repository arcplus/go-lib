package redis

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// Nil reply Redis returns when key does not exist.
const Nil = redis.Nil

var redisConfig = sync.Map{}

// Conf redis
type Conf struct {
	DSN string
}

// Register initialize redis
// dsn format -> redis://:password@url/dbNum[optional,default 0]
func Register(name string, conf Conf) error {
	opts, err := redis.ParseURL(conf.DSN)
	if err != nil {
		return err
	}

	// TODO pool size

	client := redis.NewClient(opts)

	redisConfig.Store(name, client)

	return client.Ping().Err()
}

// Client return redis client
func Client(name string) (*redis.Client, error) {
	v, ok := redisConfig.Load(name)
	if !ok {
		return nil, fmt.Errorf("rds %q not registered", name)
	}

	return v.(*redis.Client), nil
}

// Get return redis client with given name or nil if not exist
func Get(name string) *redis.Client {
	cli, _ := Client(name)
	return cli
}

// HealthCheck ping rds
func HealthCheck() error {
	errs := make(map[string]error)

	redisConfig.Range(func(key, val interface{}) bool {
		err := val.(*redis.Client).Ping().Err()
		if err != nil {
			errs[key.(string)] = err
		}
		return true
	})

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// Close close all redis conn
func Close() error {
	redisConfig.Range(func(k, v interface{}) bool {
		if c, ok := v.(*redis.Client); ok && c != nil {
			c.Close()
		}
		return false
	})
	return nil
}

// HelperGet get key from rds, if not nil, using  with v if exist. use initFunc to fetch result. v must be a pointer.
// using json as default marshal/unmarshal if is nil.
func HelperGet(name string, key string, v interface{}, ttl time.Duration, initFunc func() error, marshalFunc func(v interface{}) ([]byte, error), unmarshalFunc func(data []byte, v interface{}) error) error {
	cli, err := Client(name)
	if err != nil {
		return err
	}

	data, err := cli.Get(key).Bytes()
	if err != nil {
		if err == Nil {
			if initFunc != nil {
				err = initFunc()
				if err != nil {
					return err
				}
			}

			if marshalFunc == nil {
				marshalFunc = json.Marshal
			}

			data, err = marshalFunc(v)
			if err != nil {
				return err
			}

			return cli.Set(key, data, ttl).Err()
		}
		return err
	}

	if unmarshalFunc == nil {
		unmarshalFunc = json.Unmarshal
	}

	return unmarshalFunc(data, v)
}

package redis

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"

	"github.com/arcplus/go-lib/json"
	"github.com/arcplus/go-lib/safemap"
)

// Nil reply Redis returns when key does not exist.
const Nil = redis.Nil

var redisStore = safemap.New()

// Conf redis
type Conf struct {
	DSN string
}

// Register initialize redis
// dsn format -> redis://:password@url/dbNum[optional,default 0]
func Register(name string, conf Conf) error {
	opt, err := redis.ParseURL(conf.DSN)
	if err != nil {
		return err
	}

	// TODO opt set using config
	client := redis.NewClient(opt)
	redisStore.Set(name, client)

	return nil
}

// Client return redis client
func Client(name string) (*redis.Client, error) {
	v, ok := redisStore.Get(name)
	if !ok {
		return nil, fmt.Errorf("redis %q not registered", name)
	}

	return v.(*redis.Client), nil
}

// DB return redis client with given name or nil if not exist
func DB(name string) *redis.Client {
	cli, _ := Client(name)
	return cli
}

// Deprecated, using DB instead.
func Get(name string) *redis.Client {
	cli, _ := Client(name)
	return cli
}

// HealthCheck ping rds
func HealthCheck() error {
	errs := make(map[string]error)

	redisClients := redisStore.Items()

	for k, v := range redisClients {
		err := v.(*redis.Client).Ping().Err()
		if err != nil {
			errs[k.(string)] = err
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// Close close all redis conn
func Close() error {
	redisClients := redisStore.Items()

	for _, v := range redisClients {
		v.(*redis.Client).Close()
	}

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

package log

import (
	"io"
	"log"
	"time"

	"github.com/go-redis/redis"
)

// RedisConfig is conf for redis writer.
type RedisConfig struct {
	Level  Level
	DSN    string
	LogKey string
	Async  bool
	client *redis.Client
}

// RedisWriter redis writer.
func RedisWriter(conf RedisConfig) io.Writer {
	if conf.LogKey == "" {
		conf.LogKey = "log:basic"
	}

	opt, err := redis.ParseURL(conf.DSN)
	if err != nil {
		panic(err)
	}

	conf.client = redis.NewClient(opt)

	if conf.Async {
		wr := NewWriter(conf.Level, conf, 1000, 10*time.Millisecond, func(missed int) {
			log.Printf("Logger Dropped %d messages", missed)
		})
		asyncWaitList = append(asyncWaitList, wr.Close)
		return wr
	}

	return conf
}

// Write write data to writer
func (c RedisConfig) Write(p []byte) (n int, err error) {
	return len(p), c.client.RPush(c.LogKey, p).Err()
}

// WriteLevel write data to writer with level info provided
func (c RedisConfig) WriteLevel(level Level, p []byte) (n int, err error) {
	if level < c.Level {
		return len(p), nil
	}

	return c.Write(p)
}

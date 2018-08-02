package nsq

import (
	"github.com/youzan/go-nsq"
	"github.com/arcplus/go-lib/log"
)

// alias
type Message = nsq.Message

type HandlerFunc func(msg *Message) error

func Subscribe(addr string, config *nsq.Config, topic, channel string, handleFunc HandlerFunc, concurrency int) error {
	addr, config = getConfig(addr, config)

	if channel == "" {
		channel = "default"
	}

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	consumer.SetLogger(log.NSQLogger{}, nsq.LogLevelInfo)

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(handleFunc), concurrency)

	return consumer.ConnectToNSQLookupd(addr)
}

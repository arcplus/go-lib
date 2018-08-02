package nsq

import (
	"github.com/youzan/go-nsq"
	"github.com/arcplus/go-lib/log"
)

type Handler interface {
	HandleMessage(message *nsq.Message) error
}

func Subscribe(addr string, config *nsq.Config, topic, channel string, handler Handler, concurrency int) error {
	addr, config = getConfig(addr, config)

	if channel == "" {
		channel = "default"
	}

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	consumer.SetLogger(log.NSQLogger{}, nsq.LogLevelInfo)

	consumer.AddConcurrentHandlers(handler, concurrency)

	return consumer.ConnectToNSQLookupd(addr)
}

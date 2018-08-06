package nsq

import (
	"sync"

	"github.com/arcplus/go-lib/log"
	"github.com/youzan/go-nsq"
)

// alias
type Message = nsq.Message

type HandlerFunc func(msg *Message) error

var sbm = sync.Map{}

func Subscribe(addr string, config *nsq.Config, topic, channel string, handleFunc HandlerFunc, concurrency int) error {
	addr, config = getConfig(addr, config)

	if channel == "" {
		channel = "default"
	}

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	sbm.Store(topic+":"+channel, consumer)

	consumer.SetLogger(log.NSQLogger{}, nsq.LogLevelInfo)

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(handleFunc), concurrency)

	return consumer.ConnectToNSQLookupd(addr)
}

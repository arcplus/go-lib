package nsq

import (
	"sync"

	"github.com/youzan/go-nsq"
)

// alias
type Message = nsq.Message
type Handler = nsq.Handler

type HandlerFunc func(msg *Message) error

var sbm = sync.Map{}

// SubscribeHandler
func SubscribeHandler(lupdAddr string, config *Config, topic, channel string, handler Handler, concurrency int) error {
	config = getConfig(config)

	if channel == "" {
		channel = "default"
	}

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}

	sbm.Store(topic+":"+channel, consumer)

	consumer.SetLogger(logger{}, nsq.LogLevelInfo)

	consumer.AddConcurrentHandlers(handler, concurrency)

	return consumer.ConnectToNSQLookupd(getLupdAddr(lupdAddr))
}

// SubscribeHandleFunc
func SubscribeHandleFunc(addr string, config *Config, topic, channel string, handleFunc HandlerFunc, concurrency int) error {
	return SubscribeHandler(addr, config, topic, channel, nsq.HandlerFunc(handleFunc), concurrency)
}

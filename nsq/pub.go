package nsq

import (
	"sync"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/log"
	"github.com/youzan/go-nsq"
)

var psm = sync.Map{}

func getPubMgr(topic string) (*nsq.TopicProducerMgr, error) {
	pubMgr, ok := psm.Load(topic)
	if !ok {
		return nil, errs.New(401, "topic '%p' not register", topic)
	}
	return pubMgr.(*nsq.TopicProducerMgr), nil
}

func RegisterPub(addr string, config *nsq.Config, topics ...string) error {
	addr, config = getConfig(addr, config)

	pubMgr, err := nsq.NewTopicProducerMgr(topics, config)
	if err != nil {
		return err
	}

	pubMgr.SetLogger(log.NSQLogger{}, nsq.LogLevelInfo)

	err = pubMgr.ConnectToNSQLookupd(addr)
	if err != nil {
		return err
	}

	for i := range topics {
		psm.Store(topics[i], pubMgr)
	}

	return nil
}

func Publish(topic string, body []byte) error {
	pubMgr, err := getPubMgr(topic)
	if err != nil {
		return err
	}

	return pubMgr.Publish(topic, body)
}

func PublishOrdered(topic string, partitionKey []byte, body []byte) error {
	pubMgr, err := getPubMgr(topic)
	if err != nil {
		return err
	}

	_, _, _, err = pubMgr.PublishOrdered(topic, partitionKey, body)
	return err
}

// Close all conn
func Close() error {
	sbm.Range(func(key, val interface{}) bool {
		val.(*nsq.Consumer).Stop()
		return true
	})

	psm.Range(func(key, val interface{}) bool {
		val.(*nsq.TopicProducerMgr).Stop()
		return true
	})
	return nil
}

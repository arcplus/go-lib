package nsq

import (
	"sync"
	"errors"

	"github.com/youzan/go-nsq"
	"github.com/arcplus/go-lib/log"
)

var ErrTopicNotRegister = errors.New("topic not register")

var psm = sync.Map{}

func getPubMgr(topic string) (*nsq.TopicProducerMgr, error) {
	pubMgr, ok := psm.Load(topic)
	if !ok {
		return nil, ErrTopicNotRegister
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

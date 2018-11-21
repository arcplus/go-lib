package nsq

import (
	"sync"

	"github.com/youzan/go-nsq"

	"github.com/arcplus/go-lib/errs"
)

var psm = sync.Map{}

type Config = nsq.Config
type MsgExt = nsq.MsgExt

func getPubMgr(topic string) (*nsq.TopicProducerMgr, error) {
	pubMgr, ok := psm.Load(topic)
	if !ok {
		return nil, errs.New(errs.CodeInternal, "nsq topic '%p' not register", topic)
	}
	return pubMgr.(*nsq.TopicProducerMgr), nil
}

func RegisterPub(lupdAddr string, config *Config, topics ...string) error {
	config = getConfig(config)

	pubMgr, err := nsq.NewTopicProducerMgr(topics, config)
	if err != nil {
		return err
	}

	pubMgr.SetLogger(logger{}, nsq.LogLevelInfo)

	err = pubMgr.ConnectToNSQLookupd(lupdAddr)
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

func PublishWithJsonExt(topic string, body []byte, ext *MsgExt) error {
	pubMgr, err := getPubMgr(topic)
	if err != nil {
		return err
	}

	_, _, _, err = pubMgr.PublishWithJsonExt(topic, body, ext)
	return err
}

func PublishOrdered(topic string, partitionKey []byte, body []byte) error {
	pubMgr, err := getPubMgr(topic)
	if err != nil {
		return err
	}

	_, _, _, err = pubMgr.PublishOrdered(topic, partitionKey, body)
	return err
}

func PublishOrderedWithJsonExt(topic string, partitionKey []byte, body []byte, ext *MsgExt) error {
	pubMgr, err := getPubMgr(topic)
	if err != nil {
		return err
	}

	_, _, _, err = pubMgr.PublishOrderedWithJsonExt(topic, partitionKey, body, ext)
	return err
}

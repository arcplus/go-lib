package nsq

import (
	"github.com/youzan/go-nsq"
)

var defaultConfig = *nsq.NewConfig()

func init() {
	defaultConfig.EnableTrace = true
	defaultConfig.MaxAttempts = 128
}

func getConfig(config *Config) *Config {
	if config == nil {
		// copy of default
		d := defaultConfig
		config = &d
	}

	return config
}

// Close all conn
func Close() error {
	sbm.Range(func(key, val interface{}) bool {
		val.(*nsq.Consumer).Stop()
		return true
	})

	pm := map[*nsq.TopicProducerMgr]bool{}
	psm.Range(func(key, val interface{}) bool {
		m := val.(*nsq.TopicProducerMgr)
		if !pm[m] {
			m.Stop()
			pm[m] = true
		}
		return true
	})
	return nil
}

func getLupdAddr(lupd string) string {
	if lupd == "" || lupd == "default" {
		return "10.241.11.8:4161"
	}
	return lupd
}

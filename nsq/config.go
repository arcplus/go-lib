package nsq

import (
	"github.com/youzan/go-nsq"
)

var LookupdAddr = "10.241.11.8:4161"

var defaultConfig = *nsq.NewConfig()

func init() {
	defaultConfig.EnableTrace = true
	defaultConfig.MaxAttempts = 128
}

func getConfig(addr string, config *nsq.Config) (string, *nsq.Config) {
	if addr == "" || addr == "default" {
		addr = LookupdAddr
	}
	if config == nil {
		// copy of default
		d := defaultConfig
		config = &d
	}

	return addr, config
}

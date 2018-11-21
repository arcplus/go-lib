package nsq

import (
	"strings"

	"github.com/arcplus/go-lib/log"
)

type logger struct {
}

// nsq logger impl
func (logger) Output(calldepth int, s string) error {
	if len(s) < 5 {
		return nil
	}

	// TODO depth + caller
	l := log.Logger().KV("span", "nsq").Skip(calldepth)

	if strings.HasPrefix(s, "INF") {
		l.Info(s[5:])
		return nil
	}

	if strings.HasPrefix(s, "WRN") {
		l.Warn(s[5:])
		return nil
	}

	if strings.HasPrefix(s, "ERR") {
		l.Error(s[5:])
		return nil
	}

	l.Debug(s[5:])
	return nil
}

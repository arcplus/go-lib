package log

import (
	"strings"
)

type NSQLogger struct {
}

// nsq logger impl
func (NSQLogger) Output(calldepth int, s string) error {
	if len(s) < 5 {
		return nil
	}

	l := KV("plugin", "nsq").Skip(calldepth)

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

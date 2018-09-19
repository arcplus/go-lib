package log

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestSetAttachment(t *testing.T) {
	SetAttachment(map[string]string{"a1": "1", "a2": "2"})
	Debug("hello")
}

func TestBasic(t *testing.T) {
	Debug("msg debug")
	Debugf("msg debug %s", "format")
	Info("msg info")
	Infof("msg info %s", "format")
	Warn("msg warn")
	Warnf("msg warn %s", "format")
	Error("msg error")
	Errorf("msg error %s", "format")
}

func TestKV(t *testing.T) {
	KV("k1", "v1").KV("k2", "v2").Debug("hello")
	KV("k1", "v1").Debug("hello")
	KVPair(map[string]string{
		"p1": "1",
		"p2": "2",
	}).KV("k3", "3").Debug("hello")

	l1 := Logger()
	l1.SetKV("a1", "b1")
	l1.Debug("l1")
	Debug("x")

	wg := sync.WaitGroup{}
	go func() {
		logger1 := Logger()
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(x int) {
				defer wg.Done()
				seq := fmt.Sprint(x)
				logger1.SetKV("k"+seq, "v"+seq)
			}(i)
		}
		wg.Wait()
		logger1.SetKV("k", "v").Debug("logger1")
	}()

	go func() {
		logger2 := Logger()
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(x int) {
				defer wg.Done()
				seq := fmt.Sprint(x)
				logger2.SetKV("a"+seq, "b"+seq)
			}(i)
		}
		wg.Wait()
		logger2.SetKV("a", "b").Debug("logger2")
	}()

	wg.Wait()

	KV("k1", "v1").KV("k2", "v2").Debug("hello")

	<-time.After(time.Second)
}

func TestTrace(t *testing.T) {
	Trace("uuid-uuid-uuid-uuid").Info("trace test1")
	Trace("uuid-uuid-uuid-uuid").Error("trace test2")
	KV("x", "y").Trace("uuid-uuid-uuid-uuid").Error("trace test3")
}

func TestCaller(t *testing.T) {
	Caller().Info("caller test1")
	KV("x", "y").Caller().Errorf("trace test2")
}

var sLog = Sample(&BasicSampler{N: 2})

func TestSample(t *testing.T) {
	sLog.KV("x", "y").Debug("1")
	sLog.KV("m", "n").Debug("2")
	sLog.Debug("3")
	sLog.KV("k", "v").Debug("4")
	sLog.KV("k", "v").Debug("5")
	sLog.KV("k", "v").Debugf("6")
}

func TestWithStack(t *testing.T) {
	WithStack().Debug("hello")
	KV("stack", "x").WithStack().Debug("hello")
}

func TestLog2Redis(t *testing.T) {
	SetOutput(RedisWriter(RedisConfig{
		DSN:    "redis://:@127.0.0.1:6379",
		LogKey: "log:test",
		Level:  InfoLevel,
		Async:  true,
	}), ConsoleAsync)
	Debug("hidden")
	Info("hello info")
	Error("hello error")
	WithStack().Info("stack")
}

func TestSetLevel(t *testing.T) {
	time.Sleep(time.Second)
	SetLevel(InfoLevel)
}

func BenchmarkBase(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		log.Debug().Str("key", "v").Msg("MSG")
	}
}

func BenchmarkDebug(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Debug("MSG")
	}
}

func BenchmarkDebugf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Debugf("%s", "MSG")
	}
}

func BenchmarkKV(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		KV("a", "b").KV("x", "y").Debug("hello")
	}
}

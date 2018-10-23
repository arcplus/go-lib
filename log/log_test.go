package log

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

func TestMain(m *testing.M) {
	m.Run()
	Close()
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

func TestSetAttachment(t *testing.T) {
	SetAttachment(map[string]interface{}{"attach1": "1", "attach2": "2"})
	Debug("hello")
}

func TestEnabled(t *testing.T) {
	// reset
	defer SetLevel(DebugLevel)

	if !DebugEnabled() {
		t.Fatalf("debug level should be enabled.")
	}

	lg := Logger()
	if !lg.DebugEnabled() {
		t.Fatalf("debug level should be enabled.")
	}

	SetLevel(InfoLevel)

	if DebugEnabled() {
		t.Fatalf("debug should not be enabled.")
	}

	if lg.DebugEnabled() {
		t.Fatalf("debug should not be enabled.")
	}
}

func TestKV(t *testing.T) {
	KV("k1", "v1").KV("k2", "v2").Debug("k1,k2")
	KV("k1", "v1").Debug("k1")

	KVPair(map[string]interface{}{
		"p1": "1",
		"p2": "2",
	}).KV("k3", "3").Debug("p1,p2,k3")

	l1 := Logger()
	l1.SetKV("l1", 1)
	l1.Debug("l1")

	Debug("bare")

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

	KV("x1", "v1").KV("x2", "v2").Debug("x1,x2")

	<-time.After(time.Microsecond * 500)
}

func TestTrace(t *testing.T) {
	Trace("uuid-uuid-uuid-uuid1").Info("trace1")
	Trace("uuid-uuid-uuid-uuid2").Error("trace2")
	KV("x", "y").Trace("uuid-uuid-uuid-uuid").Error("trace test3")
}

func TestCaller(t *testing.T) {
	Caller().Info("caller1")
	KV("x", "y").Caller().Errorf("caller2")
}

var sLog = Sample(&BasicSampler{N: 2})

func TestSample(t *testing.T) {
	sLog.KV("x", "y").Debug("hidden1")
	sLog.KV("m", "n").Debug("shown2")
	sLog.Debug("hidden3")
	sLog.KV("k", "v").Debug("shown4")
	sLog.KV("k", "v").Debug("hidden5")
	sLog.KV("k", "v").Debugf("shown6")
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
	SetLevel(InfoLevel)
}

// go test -v --count=1 -test.bench=".*"

func BenchmarkBase(b *testing.B) {
	SetLevel(InfoLevel)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Debug().Str("key", "v").Msg("MSG")
	}
}

func BenchmarkDebug(b *testing.B) {
	SetLevel(InfoLevel)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Debug("MSG")
	}
}

func BenchmarkDebugf(b *testing.B) {
	SetLevel(InfoLevel)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debugf("%s", "MSG")
	}
}

func BenchmarkKV(b *testing.B) {
	SetLevel(InfoLevel)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KV("a", "b").KV("x", "y").Debugf("hello, %s", "world")
	}
}

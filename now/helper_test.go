package now

import (
	"testing"
	"time"
)

func TestNanoToMs(t *testing.T) {
	t1 := time.Now()
	time.Sleep(time.Millisecond * 100)
	td := time.Since(t1)
	t.Log(NanoToMs(td.Nanoseconds()))
}

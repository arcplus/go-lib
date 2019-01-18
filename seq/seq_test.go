package seq

import (
	"sync"
	"testing"
	"time"

	"github.com/sony/sonyflake"
)

func TestNextID(t *testing.T) {
	idc := make(chan string, 100000)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				idc <- NextID()
			}
		}()
	}

	wg.Wait()
	close(idc)

	idMap := map[string]bool{}
	for id := range idc {
		if idMap[id] {
			t.Fatal("dup hit")
		}
		idMap[id] = true
	}
	t.Log("ok")
}

func TestNextNumIDByTime(t *testing.T) {
	seqStart := NextNumIDByTime(time.Now())

	var seqList []uint64
	n := 10000
	for i := 0; i < n; i++ {
		seqList = append(seqList, NextNumID())
	}

	t.Log(sonyflake.Decompose(seqList[n-1]))

	st := int64(sonyflake.Decompose(seqList[n-1])["time"]) + startTimeNum
	sec := st / 100
	endTime := time.Unix(sec, 0).Add(time.Second)
	seqEnd := NextNumIDByTime(endTime)

	for i := range seqList {
		if seqList[i] <= seqStart {
			t.Fatal("all seq must gt seqStart")
		}
		if seqList[i] >= seqEnd {
			t.Fatal("all seq must le seqEnd")
		}
	}
}

package seq

import (
	"sync"
	"testing"
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

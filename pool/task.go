package pool

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultMaxRetries indicates the max retry count.
	DefaultMaxRetries = 3
	// RetryInterval indicates retry interval.
	RetryInterval uint64 = 500
)

// RunWithRetry will run the f with backoff and retry.
// retryCnt: Max retry count
// backoff: When run f failed, it will sleep backoff * triedCount time.Millisecond.
// Function f should have two return value. The first one is an bool which indicate if the err if retryable.
// The second is if the f meet any error.
func RunWithRetry(retryCnt int, backoff uint64, f func() (bool, error)) (err error) {
	for i := 1; i <= retryCnt; i++ {
		var retryAble bool
		retryAble, err = f()
		if err == nil || !retryAble {
			return err
		}
		sleepTime := time.Duration(backoff*uint64(i)) * time.Millisecond
		time.Sleep(sleepTime)
	}
	return err
}

// WorkFunc is simple work func
type WorkFunc func()

// WorkFuncWithCtx is simple work func
type WorkFuncWithCtx func(ctx context.Context) error

func MultiRun(fs ...func() error) error {
	if len(fs) == 0 {
		return nil
	}

	var err error
	var errOnce = &sync.Once{}

	wg := &sync.WaitGroup{}
	for i := range fs {
		if fs[i] == nil {
			continue
		}
		wg.Add(1)
		go func(i int) {
			if e := fs[i](); e != nil {
				errOnce.Do(func() {
					err = e
				})
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return err
}

// MultiRunWithCtx with ctx notify
func MultiRunWithCtx(ctx context.Context, fs ...WorkFuncWithCtx) error {
	if len(fs) == 0 {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	var errOnce = &sync.Once{}

	wg := &sync.WaitGroup{}
	for i := range fs {
		if fs[i] == nil {
			continue
		}
		wg.Add(1)
		go func(i int) {
			if e := fs[i](ctx); e != nil {
				errOnce.Do(func() {
					err = e
					cancel()
				})
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	return err
}

// MultiRunWithPool
func MultiRunWithPool(n int, fs ...WorkFunc) {
	l := len(fs)
	if l == 0 {
		return
	}

	c := make(chan WorkFunc, l)

	for i := range fs {
		c <- fs[i]
	}

	close(c)

	wg := sync.WaitGroup{}
	wg.Add(l)

	// shrink worker size
	if n > l || n <= 0 {
		n = l
	}

	// multi worker
	for i := 0; i < n; i++ {
		go func() {
			for w := range c {
				w()
				wg.Done()
			}
		}()
	}

	wg.Wait()
}

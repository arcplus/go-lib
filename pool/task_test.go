package pool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMultiRun(t *testing.T) {
	w1 := func() error {
		time.Sleep(time.Millisecond * 200)
		fmt.Println("second")
		return nil
	}

	w2 := func() error {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("last")
		return errors.New("w2 failed")
	}

	var w3e = errors.New("w2 failed")
	w3 := func() error {
		time.Sleep(time.Millisecond * 100)
		fmt.Println("first")
		return w3e
	}

	err := MultiRun(w1, w2, w3)
	if err != w3e {
		t.Fatal("multi be w3e", err)
	}
}

func TestMultiRunWithCtx(t *testing.T) {
	ctx := context.Background()
	w1 := func(ctx context.Context) error {
		time.Sleep(time.Millisecond * 200)
		if ctx.Err() != context.Canceled {
			t.Fatal("w1 ctx err should be canceled")
		}
		fmt.Println("last", ctx.Err())
		return errors.New("w1 error")
	}

	w2e := errors.New("w2 failed")
	w2 := func(ctx context.Context) error {
		fmt.Println("first", ctx.Err())
		time.Sleep(time.Millisecond * 100)
		return w2e
	}

	w3 := func(ctx context.Context) error {
		time.Sleep(time.Millisecond * 50)
		if ctx.Err() != nil {
			t.Fatal("w3 ctx err should be nil")
		}
		fmt.Println("second", ctx.Err())
		return nil
	}
	err := MultiRunWithCtx(ctx, w1, w2, w3)
	if err != w2e {
		t.Fatal("result should be w2e")
	}
}

func TestMultiRunWithPool(t *testing.T) {
	wl := func() []WorkFunc {
		x := []WorkFunc{}
		for i := 0; i < 10; i++ {
			f := func() {
				time.Sleep(time.Millisecond * 500)
				fmt.Println(i)
			}
			x = append(x, f)
		}
		return x
	}()

	MultiRunWithPool(5, wl...)
}

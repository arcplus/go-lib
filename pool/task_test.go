package pool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMultiRunWithCtx(t *testing.T) {
	ctx := context.Background()
	w1 := func(ctx context.Context) error {
		time.Sleep(time.Millisecond * 200)
		if ctx.Err() != context.Canceled {
			t.Fatal("w1 ctx err should be canceled")
		}
		fmt.Println("w1", ctx.Err())
		return nil
	}

	w2 := func(ctx context.Context) error {
		fmt.Println("w2", ctx.Err())
		time.Sleep(time.Millisecond * 100)
		return errors.New("w2 failed")
	}

	w3 := func(ctx context.Context) error {
		time.Sleep(time.Millisecond * 50)
		if ctx.Err() != nil {
			t.Fatal("w3 ctx err should be nil")
		}
		fmt.Println("w3", ctx.Err())
		return nil
	}
	err := MultiRunWithCtx(ctx, w1, w2, w3)
	if err == nil {
		t.Fatal("multi result should be err")
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

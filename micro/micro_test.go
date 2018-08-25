package micro

import (
	"errors"
	"testing"
)

func TestMicro_Close(t *testing.T) {
	m := New()
	m.Close()

	m.AddResCloseFunc(func() error {
		return errors.New("res1")
	})

	m.AddResCloseFunc(func() error {
		return errors.New("res2")
	})

	m.Close()
	m.Close()
}

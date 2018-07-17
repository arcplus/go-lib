package errs

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	e0 := New(0, "internal")
	t.Log(e0)
	e1 := New(404, "not found")
	t.Log(e1)
}

func TestWrap(t *testing.T) {
	e1 := Wrap(nil, 404, "not found")
	if e1 != nil {
		t.Fatal("e1 should be nil")
	}

	err := errors.New("hello")
	e2 := Wrap(err, 200, "ok")
	t.Log(e2)
	e3 := Wrap(e2, 302, "not modified")
	t.Log(e3)
}

func TestTrace(t *testing.T) {
	err := errors.New("hello")
	e1 := Trace(err)
	t.Log(e1)

	e2 := Trace(e1)
	t.Log(e2)

	e3 := Trace(e2)
	t.Log(e3)
}

func TestInternal(t *testing.T) {
	var err error
	if err != nil {
		t.Fatal("err should be nil")
	}
	err = Internal(err)
	if err != nil {
		t.Fatal("err should be nil")
	}
}

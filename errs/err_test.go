package errs

import (
	"errors"
	"testing"

	"github.com/arcplus/go-lib/json"
)

var errGo = errors.New("go error")
var errCodeTest uint32 = 9999
var errNew = New(errCodeTest, "errs.Error")

func TestNew(t *testing.T) {
	e1 := New(0, "ok")
	if e1 != nil {
		t.Fatalf("e0 should be nil")
	}
	t.Log(e1)
	e2 := New(404, "not found")
	if e2.Error() != "[404]not found" {
		t.Fatal(e2)
	}
	t.Log(e2)
}

func TestNewWithAlert(t *testing.T) {
	e1 := NewWithAlert(CodeBadRequest, "少参数", "missing params")
	t.Log(e1)
	data, _ := json.Marshal(e1)
	t.Log(string(data))
	t.Log(StackTrace(e1))
}

func TestWrap(t *testing.T) {
	e1 := Wrap(nil, 404, "not found")
	if e1 != nil {
		t.Fatal("e1 should be nil")
	}
	t.Log(e1)

	err := errors.New("hello")
	e2 := Wrap(err, 200)
	if e2.Error() != "[200]hello" {
		t.Fatal(e2)
	}
	t.Log(e2)
	e3 := Wrap(e2, 302, "not modify")
	if e3.Error() != "[302]not modify" {
		t.Fatal(e3)
	}
	t.Log(e3)
}

func TestTrace(t *testing.T) {
	e1 := Trace(errGo)
	t.Log(e1)

	e2 := Trace(errNew)
	t.Log(e2)

	e3 := Trace(e2)
	t.Log(e3)
}

func TestAnnotate(t *testing.T) {
	e1 := Annotate(errGo, "anno")
	if e1.Error() != "[1000]anno" {
		t.Fatal(e1)
	}
	e2 := Annotate(errNew, "new msg")
	if e2.Error() != "[9999]new msg" {
		t.Fatal(e2)
	}
}

func TestDeferredAnnotate(t *testing.T) {
	e1 := func() (err error) {
		err = errGo
		defer DeferredAnnotate(&err, "changed")
		return
	}()
	if e1.Error() != "[1000]changed" {
		t.Fatal(e1)
	}

	e2 := func() (err error) {
		err = errNew
		defer DeferredAnnotate(&err, "changed")
		return
	}()
	if e2.Error() != "[9999]changed" {
		t.Fatal(e2)
	}
}

package errs

import (
	"errors"
	"testing"
)

func TestIs(t *testing.T) {
	e1 := errors.New("e1")
	if !IsErr(e1, e1) {
		t.Fatal("e1 should equal e1")
	}
	e2 := Wrap(e1, 302)
	if !IsErr(e2, e1) {
		t.Fatal("e2 should equal e1")
	}
	e3 := Wrap(e2, 404)
	if !IsErr(e3, e1) {
		t.Fatal("e3 should equal e1")
	}
	if !IsErr(e3, e2) {
		t.Fatal("e3 should equal e2")
	}
}

func TestIsCode(t *testing.T) {
	if IsCode(errGo, CodeInternal) {
		t.Fatalf("%v should not code %d", errGo, CodeInternal)
	}
	if !IsCode(errNew, errCodeTest) {
		t.Fatalf("%v should not code %d", errNew, errCodeTest)
	}
	e1 := Wrap(errNew, 302)
	if !IsCode(e1, 302) {
		t.Fatalf("%v should not code %d", e1, 302)
	}
	e2 := Wrap(e1, 404)
	if !IsCode(e2, 302) {
		t.Fatalf("%v should not code %d", e2, 302)
	}
	if !IsCode(e2, 404) {
		t.Fatalf("%v should not code %d", e2, 404)
	}
}

func TestToError(t *testing.T) {

	//t.Log(ToError(e1.Err()))
}

func TestStackTrace(t *testing.T) {
	t.Logf("\n%s", StackTrace(errGo))
	t.Logf("\n%s", StackTrace(errNew))
	e1 := Trace(errGo)
	t.Logf("\n%s", StackTrace(e1))
	e2 := Trace(e1)
	t.Logf("\n%s", StackTrace(e2))
}

func TestCause(t *testing.T) {
	var origin error
	if Cause(origin) != nil {
		t.Fatal("should be nil")
	}
	origin = errors.New("cause")
	if Cause(origin) != origin {
		t.Fatal("should be origin")
	}

	e0 := Trace(origin)
	if Cause(e0) != origin {
		t.Fatal("e0 should be origin")
	}

	e1 := Wrap(e0, CodeBadRequest)
	if Cause(e1) != origin {
		t.Fatal("e1 should be origin")
	}

	e2 := Wrap(e1, CodeConflict)
	if Cause(e2) != origin {
		t.Fatal("e2 should be origin")
	}
}

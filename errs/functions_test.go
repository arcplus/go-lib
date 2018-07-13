package errs

import (
	"errors"
	"testing"
)

func TestStack(t *testing.T) {
	e1 := Trace(errors.New("hello"))
	e2 := Trace(e1)

	e3 := Wrap(e2, 200, "ok")
	e4 := Trace(e3)
	t.Log(StackTrace(e4))
}

func TestEqual(t *testing.T) {
	e1 := errors.New("hello")
	e2 := Trace(errors.New("hello"))

	t.Log(Equal(e1, e2))

	type args struct {
		e1 error
		e2 error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil test",
			args: args{e1: nil, e2: nil},
			want: true,
		},
		{
			name: "not equal test",
			args: args{e1: nil, e2: errors.New("test")},
			want: false,
		},
		{
			name: "not equal test",
			args: args{e1: errors.New("hello"), e2: errors.New("world")},
			want: false,
		},
		{
			name: "equal test",
			args: args{e1: errors.New("hello"), e2: errors.New("hello")},
			want: true,
		},
		{
			name: "equal testxyz",
			args: args{e1: errors.New("hello"), e2: Trace(errors.New("hello"))},
			want: true,
		},
		{
			name: "equal test",
			args: args{e1: &Error{}, e2: &Error{}},
			want: true,
		},
		{
			name: "equal test",
			args: args{e1: &Error{code: 123, message: "hello"}, e2: &Error{code: 123, message: "world"}},
			want: true,
		},
		{
			name: "not equal test",
			args: args{e1: &Error{code: 123, message: "hello"}, e2: &Error{code: 456, message: "hello"}},
			want: false,
		},
		{
			name: "not equal test",
			args: args{e1: &Error{code: 0, message: "hello"}, e2: &Error{code: 0, message: "world"}},
			want: false,
		},
		{
			name: "equal test",
			args: args{e1: &Error{code: 0, message: "hello"}, e2: &Error{code: 0, message: "hello"}},
			want: true,
		},
		{
			name: "equal test",
			args: args{e1: New(0, "hello world"), e2: New(0, "hello %s", "world")},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.e1, tt.args.e2); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCodeEqual(t *testing.T) {
	err := New(404, "not found")
	if !CodeEqual(404, err) {
		t.Fatal("should equal")
	}
}

func TestToGRPCErr(t *testing.T) {
	err := New(404, "not found")
	t.Log(ToGRPC(err))
}

func TestUnWrap(t *testing.T) {
	var err error = New(404, "not found")
	if UnWrap(err).Code() != 404 {
		t.Fatal("code should equal")
	}

	err = ToGRPC(err)
	if UnWrap(err).Code() != 404 {
		t.Fatal("code should equal")
	}
}

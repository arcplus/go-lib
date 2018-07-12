package errs

import (
	"testing"

	"github.com/juju/errors"
)

func Test(t *testing.T) {
	e1 := New(404, "not found")
	t.Log(e1)

	e2 := Annotate(e1, "hello")
	t.Log(ErrorStack(e2))
}

func TestEqual(t *testing.T) {
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
			name: "equal test",
			args: args{e1: errors.New("hello"), e2: errors.Trace(errors.New("hello"))},
			want: true,
		},
		{
			name: "equal test",
			args: args{e1: errors.New("hello"), e2: errors.Annotate(errors.New("hello"), "world")},
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

func TestNewErr(t *testing.T) {
	err := NewErr(errors.New("hello"))
	t.Log(err)
	err = NewErr(New(404, "not found"))
	t.Log(err)
}

func TestUnWrap(t *testing.T) {
	var err error = NewErr(New(404, "not found"))
	e := UnWrap(err)
	t.Log(e)
	err = NewErr(errors.New("hello"))
	e = UnWrap(err)
	t.Log(e)
}

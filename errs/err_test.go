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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.e1, tt.args.e2); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

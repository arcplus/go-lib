package errs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPC convert *Error to gRPC error
func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	if v, ok := err.(*Error); ok {
		return v.ToGRPC()
	}

	if s, ok := status.FromError(err); ok {
		return s.Err()
	}

	return status.Error(codes.Unknown, err.Error())
}

// UnWrap convert err or gRPC err to *Error if possible
func UnWrap(err error) *Error {
	if err == nil {
		return nil
	}

	v, ok := err.(*Error)
	if ok {
		return v
	}

	if s, ok := status.FromError(err); ok {
		return new(ErrCode(s.Code()), s.Message(), nil, nil, 1)
	}

	return new(0, err.Error(), nil, nil, 1)
}

// Equal checks if two errors is equal
func Equal(err1, err2 error) bool {
	originErr1, originErr2 := Cause(err1), Cause(err2)
	if originErr1 == originErr2 {
		return true
	}

	if originErr1 == nil || originErr2 == nil {
		return false
	}

	inErr1, _ := originErr1.(*Error)
	inErr2, _ := originErr2.(*Error)
	if inErr1 != nil && inErr2 != nil {
		if inErr1.code == 0 && inErr2.code == 0 {
			return inErr1.Message() == inErr2.Message()
		}
		return inErr1.code == inErr2.code
	}

	// same Error() return
	return originErr1.Error() == originErr2.Error()
}

// CodeEqual check err.(*Error).code==code
func CodeEqual(code ErrCode, err error) bool {
	v, ok := err.(*Error)
	if !ok {
		return false
	}
	return v.code == code
}

// Cause returns the cause of the given error.  This will be either the
// original error, or the result of a Wrap or Mask call.
//
// Cause is the usual way to diagnose errors that may have been wrapped by
// the other errors functions.
// Cause returns the cause of the given error.  This will be either the
// original error, or the result of a Wrap or Mask call.
//
// Cause is the usual way to diagnose errors that may have been wrapped by
// the other errors functions.
func Cause(err error) error {
	var diag error
	if err, ok := err.(causer); ok {
		diag = err.Cause()
	}
	if diag != nil {
		return diag
	}
	return err
}

type locationer interface {
	Location() (string, int)
}

type wrapper interface {
	Code() ErrCode
	// Message returns the top level error message,
	// not including the message from the Previous
	// error.
	Message() string

	// Underlying returns the Previous error, or nil
	// if there is none.
	Underlying() error
}

type causer interface {
	Cause() error
}

func StackTrace(err error) string {
	return strings.Join(Stack(err), "\n")
}

func Stack(err error) []string {
	if err == nil {
		return nil
	}

	var lines []string
	for {
		var buff []byte
		if err, ok := err.(locationer); ok {
			file, line := err.Location()
			// Strip off the leading GOPATH/src path elements.
			file = trimGoPath(file)
			if file != "" {
				buff = append(buff, fmt.Sprintf("%s:%d", file, line)...)
				buff = append(buff, ": "...)
			}
		}
		if cerr, ok := err.(wrapper); ok {
			if code := cerr.Code(); code != 0 {
				buff = append(buff, "["+strconv.Itoa(int(code))+"]"...)
			}

			message := cerr.Message()
			buff = append(buff, message...)
			// If there is a cause for this error, and it is different to the cause
			// of the underlying error, then output the error string in the stack trace.
			var cause error
			if err1, ok := err.(causer); ok {
				cause = err1.Cause()
			}
			err = cerr.Underlying()
			if cause != nil && !sameError(Cause(err), cause) {
				if message != "" {
					buff = append(buff, ": "...)
				}
				buff = append(buff, cause.Error()...)
			}
		} else {
			buff = append(buff, err.Error()...)
			err = nil
		}
		lines = append(lines, string(buff))
		if err == nil {
			break
		}
	}
	return lines
}

// Ideally we'd have a way to check identity, but deep equals will do.
func sameError(e1, e2 error) bool {
	return reflect.DeepEqual(e1, e2)
}

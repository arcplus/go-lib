package errs

import (
	"fmt"
	"strings"
)

var _ wrapper = Error{}

// A Wrapper is an error implementation
// wrapping context around another error.
type wrapper interface {
	// Unwrap returns the next error in the error chain.
	// If there is no next error, Unwrap returns nil.
	Unwrap() error
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type wrapper interface {
//		Unwrap() error
//	}
//
// If the error does not implement wrapper, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	for err != nil {
		w, ok := err.(wrapper)
		if !ok {
			break
		}
		cause := w.Unwrap()
		if cause == nil {
			break
		}
		err = cause
	}
	return err
}

// Is reports whether err or any of the errors in its chain is equal to target.
func IsErr(err, target error) bool {
	for {
		if err == target {
			return true
		}
		wrapper, ok := err.(wrapper)
		if !ok {
			return false
		}
		err = wrapper.Unwrap()
		if err == nil {
			return false
		}
	}
}

func IsCode(err error, code Code) bool {
	for {
		if err == nil {
			return false
		}
		e, ok := err.(*Error)
		if !ok {
			return false
		}
		if e.code == code {
			return true
		}
		err = e.Unwrap()
	}
}

// ToError try convert err to *Error without line info
func ToError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}

	return newError(CodeInternal, err.Error(), nil, err, 1)
}

// WithAlert change the *Error alert
func WithAlert(err error, alert string) {
	if e, ok := err.(*Error); ok {
		e.alert = alert
	}
}

var _ locationer = Error{}

type locationer interface {
	Location() (string, int)
}

func StackTrace(err error) string {
	return strings.Join(stack(err), "\n")
}

// Stack return all errs with line info if possible
// TODO stack buff optimize
func stack(err error) []string {
	if err == nil {
		return nil
	}

	var lines []string
	for {
		var buff []byte
		if err, ok := err.(locationer); ok {
			file, line := err.Location()
			// Strip off the leading GOPATH/src path elements.
			if file != "" {
				file = trimGoPath(file)
				buff = append(buff, fmt.Sprintf("%s:%d ", file, line)...)
			}
		}

		buff = append(buff, err.Error()...)
		if cerr, ok := err.(wrapper); ok {
			err = cerr.Unwrap()
		} else {
			err = nil
		}

		lines = append(lines, string(buff))
		if err == nil {
			break
		}
	}
	return lines
}

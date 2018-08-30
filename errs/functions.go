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

// Is reports whether err or any of the errors in its chain is equal to target.
func Is(err, target error) bool {
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

func IsCode(err error, code ErrCode) bool {
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

func ToError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}

	return new(ErrInternal, err.Error(), nil, err, 1)
}

var _ locationer = Error{}

type locationer interface {
	Location() (string, int)
}

func StackTrace(err error) string {
	return strings.Join(stack(err), "\n")
}

// Stack return all errs with line info if possible
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
			file = trimGoPath(file)
			if file != "" {
				buff = append(buff, fmt.Sprintf("%s:%d", file, line)...)
				buff = append(buff, ": "...)
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

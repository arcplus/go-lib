package errs

import (
	"fmt"
	"strings"
)

// Wrapper is an error implementation
// wrapping context around another error.
type Wrapper interface {
	// Unwrap returns the next error in the error chain.
	// If there is no next error, Unwrap returns nil.
	Unwrap() error
}

// Locationer indicate where error occured.
type Locationer interface {
	Location() (string, int)
}

// Errorer interface.
type Errorer interface {
	Code() uint32
	Message() string
}

// toError try convert to *err
func toError(e error) *Error {
	if e == nil {
		return nil
	}

	if er, ok := e.(*Error); ok {
		en := *er
		en.prev = er
		return &en
	}

	if er, ok := e.(Errorer); ok {
		return &Error{
			code: er.Code(),
			msg:  er.Message(),
			prev: e,
		}
	}

	return &Error{
		code: CodeInternal,
		msg:  e.Error(),
		prev: e,
	}
}

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errs.New(401, "missing id")
//
func New(code uint32, msg string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	er := &Error{
		code: code,
		msg:  msg,
		args: args,
	}

	er.setLocation(1)

	return er
}

// NewWithAlert is a drop in replacement for the standard library errors module that records
// the location that the error is created but with alert msg for show.
//
// For example:
//    return errs.New(404, "用户不存在", "user not found")
//
func NewWithAlert(code uint32, alert string, msg string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	er := &Error{
		code:  code,
		msg:   msg,
		args:  args,
		alert: alert,
	}

	er.setLocation(1)

	return er
}

// Trace add file and line info to err.
//
// For example:
//   err:=someFunc()
//   if err!=nil {
//	   return errs.Trace(err)
//   }
func Trace(e error) error {
	er := toError(e)
	if er == nil {
		return nil
	}

	er.setLocation(1)

	return er
}

// Wrap changes the code of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//    err:=someFunc()
//    if err!=nil {
//        return errs.Wrap(err, 500, "internal err")
//    }
func Wrap(e error, code uint32, v ...interface{}) error {
	er := toError(e)
	if er == nil {
		return nil
	}

	er.code = code
	if len(v) != 0 {
		if v0, ok := v[0].(string); ok {
			er.msg = v0
		} else {
			er.msg = fmt.Sprint(v[0])
		}
		er.args = v[1:]
	}
	er.setLocation(1)

	return er
}

// Annotate is used to add extra context to an existing error. The location of
// the Annotate call is recorded with the annotations. The file, line and
// function are also recorded.
//
// For example:
//   if err := SomeFunc(); err != nil {
//       return errs.Annotate(err, "failed to frombulate")
//   }
//
func Annotate(e error, msg string, args ...interface{}) error {
	er := toError(e)
	if er == nil {
		return nil
	}

	er.msg = msg
	er.args = args
	er.setLocation(1)

	return er
}

// DeferredAnnotate annotates the given error (when it is not nil) with the given
// format string and arguments (like fmt.Sprintf). If *err is nil, DeferredAnnotatef
// does nothing. This method is used in a defer statement in order to annotate any
// resulting error with the same message.
//
// For example:
//
//    defer DeferredAnnotate(&err, "failed to frombulate the %s", arg)
//
func DeferredAnnotate(e *error, msg string, args ...interface{}) {
	er := toError(*e)
	if er == nil {
		return
	}

	er.msg = msg
	er.args = args
	er.setLocation(1)

	*e = er
}

// ToError convert err to *err
func ToError(e error) *Error {
	er := toError(e)
	if er == nil {
		return nil
	}

	er.setLocation(1)

	return er
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
func Cause(e error) error {
	for e != nil {
		w, ok := e.(Wrapper)
		if !ok {
			break
		}
		c := w.Unwrap()
		if c == nil {
			break
		}
		e = c
	}
	return e
}

// Is reports whether err or any of the errors in its chain is equal to target.
func IsErr(err, target error) bool {
	for {
		if err == target {
			return true
		}
		wrapper, ok := err.(Wrapper)
		if !ok {
			return false
		}
		err = wrapper.Unwrap()
		if err == nil {
			return false
		}
	}
}

func IsCode(err error, code uint32) bool {
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
		if err, ok := err.(Locationer); ok {
			file, line := err.Location()
			// Strip off the leading GOPATH/src path elements.
			if file != "" {
				file = trimGoPath(file)
				buff = append(buff, fmt.Sprintf("%s:%d ", file, line)...)
			}
		}

		buff = append(buff, err.Error()...)
		if cerr, ok := err.(Wrapper); ok {
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

// BadRequest error
func BadRequest(msg interface{}, args ...interface{}) error {
	er := &Error{
		code: CodeBadRequest,
		args: args,
	}

	switch v := msg.(type) {
	case string:
		er.msg = v
	case error:
		er.prev = v
		if e, ok := v.(*Error); ok {
			er.msg = e.Message()
		} else {
			er.msg = v.Error()
		}
	default:
		er.msg = fmt.Sprint(msg)
	}

	er.setLocation(1)

	return er
}

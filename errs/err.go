package errs

import (
	"encoding/json"
	"fmt"
	"runtime"
)

type ErrCode uint32

const (
	ErrOK         ErrCode = 0
	ErrInternal   ErrCode = 1000
	ErrBadRequest ErrCode = 1400
	ErrUnAuth     ErrCode = 1401 // miss token
	ErrForbidden  ErrCode = 1403
	ErrNotFound   ErrCode = 1404
	ErrConflict   ErrCode = 1409 // conflict error
)

// Error implements error interface and add Code, so
// errors with different message can be compared.
type Error struct {
	code ErrCode // cause
	msg  string

	args []interface{}

	// previous holds the previous error in the error stack, if any.
	prev error

	// file and line hold the source code location where the error was
	// created.
	file string
	line int
}

// new returns Error
func new(code ErrCode, msg string, args []interface{}, prev error, skip int) *Error {
	err := Error{
		code: code,
		msg:  msg,
		args: args,
		prev: prev,
	}
	_, err.file, err.line, _ = runtime.Caller(skip + 1)
	return &err
}

func (e Error) Code() ErrCode {
	return e.code
}

// Message returns formatted message
func (e Error) Message() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.msg, e.args...)
	}
	return e.msg
}

var _ error = &Error{}

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code(), e.Message())
}

// Unwrap returns the next error in the error chain.
// If there is no next error, Unwrap returns nil.
func (e Error) Unwrap() error {
	return e.prev
}

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e Error) Location() (file string, line int) {
	return e.file, e.line
}

// MarshalJSON implements json.Marshaler interface.
func (e Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code ErrCode `json:"code"`
		Msg  string  `json:"message"`
	}{
		Code: e.Code(),
		Msg:  e.Message(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (e *Error) UnmarshalJSON(data []byte) error {
	info := &struct {
		Code ErrCode `json:"code"`
		Msg  string  `json:"message"`
	}{}

	if err := json.Unmarshal(data, &info); err != nil {
		return Trace(err)
	}

	e.code = info.Code
	e.msg = info.Msg
	return nil
}

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errs.New(401, "missing id")
//
func New(code ErrCode, msg string, args ...interface{}) error {
	if code == ErrOK {
		return nil
	}

	return new(code, msg, args, nil, 1)
}

// Wrap changes the code of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//    err:=someFunc()
//    if err!=nil {
//        return errs.Wrap(err, 500, "internal err")
//    }
func Wrap(err error, code ErrCode, v ...interface{}) error {
	if err == nil || code == ErrOK {
		return nil
	}

	var msg string
	var args []interface{}
	if len(v) == 0 {
		if e, ok := err.(*Error); ok {
			msg = e.Message()
		} else {
			msg = err.Error()
		}
	} else {
		if v0, ok := v[0].(string); ok {
			msg = v0
		} else {
			msg = fmt.Sprint(v[0])
		}
		args = v[1:]
	}

	return new(code, msg, args, err, 1)
}

// Trace add file and line info to err.
//
// For example:
//   err:=someFunc()
//   if err!=nil {
//	   return errs.Trace(err)
//   }
func Trace(err error) error {
	if err == nil {
		return nil
	}

	newErr := new(ErrInternal, "", nil, nil, 1)

	v, ok := err.(*Error)
	if ok {
		newErr.code = v.Code()
		newErr.msg = v.Message()
		newErr.prev = err
	} else {
		newErr.msg = err.Error()
	}

	return newErr
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
func Annotate(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	newErr := new(ErrInternal, msg, args, err, 1)
	newErr.prev = err

	if v, ok := err.(*Error); ok {
		newErr.code = v.Code()
	}

	return newErr
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
func DeferredAnnotate(err *error, msg string, args ...interface{}) {
	if *err == nil {
		return
	}

	newErr := new(ErrInternal, msg, args, *err, 1)
	newErr.prev = *err

	if v, ok := (*err).(*Error); ok {
		newErr.code = v.Code()
	}

	*err = newErr
}

// Internal for some internal error
func Internal(err error) error {
	if err == nil {
		return nil
	}
	return new(ErrInternal, err.Error(), nil, nil, 1)
}

// BadRequest error
func BadRequest(msg interface{}, args ...interface{}) error {
	var msgStr string
	switch v := msg.(type) {
	case string:
		msgStr = v
	case error:
		if e, ok := v.(*Error); ok {
			return new(ErrBadRequest, e.Message(), nil, e, 1)
		}
		return new(ErrBadRequest, v.Error(), nil, v, 1)
	default:
		msgStr = fmt.Sprint(msg)
	}

	return new(ErrBadRequest, msgStr, args, nil, 1)
}

// UnAuthorized error
func UnAuthorized(msg interface{}, args ...interface{}) error {
	var msgStr string
	switch v := msg.(type) {
	case string:
		msgStr = v
	case error:
		if e, ok := v.(*Error); ok {
			return new(ErrUnAuth, e.Message(), nil, e, 1)
		}
		return new(ErrUnAuth, v.Error(), nil, v, 1)
	default:
		msgStr = fmt.Sprint(msg)
	}

	return new(ErrUnAuth, msgStr, args, nil, 1)
}

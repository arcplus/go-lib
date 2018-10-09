package errs

import (
	"encoding/json"
	"fmt"
	"runtime"
)

type Code uint32

// Deprecated using Code instead
type ErrCode = Code

// Deprecated using CodeXXX instead
const (
	ErrOK         = CodeOK
	ErrInternal   = CodeInternal
	ErrBadRequest = CodeBadRequest
	ErrUnAuth     = CodeUnAuth // miss token
	ErrForbidden  = CodeForbidden
	ErrNotFound   = CodeNotFound
	ErrConflict   = CodeConflict // conflict error
)

const (
	CodeOK         Code = 0
	CodeInternal   Code = 1000
	CodeBadRequest Code = 1400
	CodeUnAuth     Code = 1401 // miss token
	CodeForbidden  Code = 1403
	CodeNotFound   Code = 1404
	CodeNotAllowed Code = 1405
	CodeConflict   Code = 1409 // conflict error
)

// Error implements error interface and add Code, so
// errors with different message can be compared.
type Error struct {
	code    Code          // cause
	message string        // message info
	args    []interface{} // fmt
	alert   string        // alert info

	// previous holds the previous error in the error stack, if any.
	prev error

	// file and line hold the source code location where the error was
	// created.
	file string
	line int
}

// new returns Error
func newError(code Code, msg string, args []interface{}, prev error, skip int) *Error {
	err := &Error{
		code:    code,
		message: msg,
		args:    args,
		prev:    prev,
	}

	// using skip -1 as no need line info
	if skip != -1 {
		_, err.file, err.line, _ = runtime.Caller(skip + 1)
	}

	return err
}

// Code return err casue code.
func (e Error) Code() Code {
	return e.code
}

// Message returns formatted message if args not empty
func (e Error) Message() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.message, e.args...)
	}
	return e.message
}

// Alert is used for err hint
func (e Error) Alert() string {
	return e.alert
}

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

type errJSON struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Alert   string `json:"alert"`
}

// MarshalJSON implements json.Marshaler interface.
func (e Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(errJSON{
		Code:    e.Code(),
		Message: e.Message(),
		Alert:   e.alert,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (e *Error) UnmarshalJSON(data []byte) error {
	info := &errJSON{}

	if err := json.Unmarshal(data, &info); err != nil {
		return Trace(err)
	}

	e.code = info.Code
	e.message = info.Message
	e.alert = info.Alert
	return nil
}

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errs.New(401, "missing id")
//
func New(code Code, msg string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	return newError(code, msg, args, nil, 1)
}

// NewRaw do the same as New but no line info added.
//
// For example:
//    return errs.NewRaw(401, "missing id")
//
func NewRaw(code Code, msg string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	return newError(code, msg, args, nil, -1)
}

// NewWithAlert is a drop in replacement for the standard library errors module that records
// the location that the error is created but with alert msg for show.
//
// For example:
//    return errs.New(404, "用户不存在", "user not found")
//
func NewWithAlert(code Code, alert string, msgf string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	e := newError(code, msgf, args, nil, 1)
	e.alert = alert
	return e
}

// NewRawWithAlert do the same as NewWithHint but no line info added.
//
// For example:
//    return errs.New(404, "用户不存在", "user not found")
//
func NewRawWithAlert(code Code, alert string, msgf string, args ...interface{}) error {
	if code == CodeOK {
		return nil
	}

	e := newError(code, msgf, args, nil, -1)
	e.alert = alert
	return e
}

// Wrap changes the code of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//    err:=someFunc()
//    if err!=nil {
//        return errs.Wrap(err, 500, "internal err")
//    }
func Wrap(err error, code Code, v ...interface{}) error {
	if err == nil || code == CodeOK {
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

	return newError(code, msg, args, err, 1)
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

	newErr := newError(CodeInternal, "", nil, nil, 1)

	v, ok := err.(*Error)
	if ok {
		newErr.code = v.Code()
		newErr.message = v.Message()
		newErr.prev = err
	} else {
		newErr.message = err.Error()
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

	newErr := newError(CodeInternal, msg, args, err, 1)
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

	newErr := newError(CodeInternal, msg, args, *err, 1)
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
	return newError(CodeInternal, err.Error(), nil, nil, 1)
}

// BadRequest error
func BadRequest(msg interface{}, args ...interface{}) error {
	var msgStr string
	switch v := msg.(type) {
	case string:
		msgStr = v
	case error:
		if e, ok := v.(*Error); ok {
			return newError(CodeBadRequest, e.Message(), nil, e, 1)
		}
		return newError(CodeBadRequest, v.Error(), nil, v, 1)
	default:
		msgStr = fmt.Sprint(msg)
	}

	return newError(CodeBadRequest, msgStr, args, nil, 1)
}

// UnAuthorized error
func UnAuthorized(msg interface{}, args ...interface{}) error {
	var msgStr string
	switch v := msg.(type) {
	case string:
		msgStr = v
	case error:
		if e, ok := v.(*Error); ok {
			return newError(CodeUnAuth, e.Message(), nil, e, 1)
		}
		return newError(CodeUnAuth, v.Error(), nil, v, 1)
	default:
		msgStr = fmt.Sprint(msg)
	}

	return newError(CodeUnAuth, msgStr, args, nil, 1)
}

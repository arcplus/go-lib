package errs

import (
	"encoding/json"
	"fmt"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// A Code is an unsigned 32-bit error code
type ErrCode uint32

const ErrUnknown ErrCode = 2

// Error implements error interface and add Code, so
// errors with different message can be compared.
type Error struct {
	code    ErrCode // cause
	message string

	args []interface{}

	// previous holds the previous error in the error stack, if any.
	prev error

	// file and line hold the source code location where the error was
	// created.
	file string
	line int
}

func new(code ErrCode, message string, args []interface{}, prev error, skip int) *Error {
	err := &Error{code: code, message: message, args: args, prev: prev}
	_, err.file, err.line, _ = runtime.Caller(skip + 1)
	return err
}

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errs.New(401, "missing id")
//
func New(code ErrCode, message string, args ...interface{}) *Error {
	return new(code, message, args, nil, 1)
}

// Wrap changes the code of the error. The location of the Wrap call is also
// stored in the error stack.
//
// For example:
//  err:=someFunc()
//  if err!=nil{
//    return errs.Wrap(err, 500, "internal err")
// }
func Wrap(err error, code ErrCode, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	// add trace info
	if _, ok := err.(*Error); !ok {
		err = new(ErrUnknown, err.Error(), nil, nil, 1)
	}

	return new(code, message, args, err, 1)
}

// Trace add file and line info to err
//
// For example:
// err:=someFunc()
// if err!=nil {
//	return errs.Trace(err)
// }
func Trace(err error) error {
	if err == nil {
		return nil
	}

	newErr := new(ErrUnknown, "", nil, nil, 1)

	v, ok := err.(*Error)
	if !ok {
		newErr.message = err.Error()
	} else {
		newErr.code = v.Code()
		newErr.message = v.Message()
		newErr.prev = err
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
func Annotate(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	newErr := new(ErrUnknown, message, args, err, 1)

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
func DeferredAnnotate(err *error, message string, args ...interface{}) {
	if *err == nil {
		return
	}

	newErr := new(ErrUnknown, message, args, *err, 1)

	if v, ok := (*err).(*Error); ok {
		newErr.code = v.Code()
	}

	*err = newErr
}

// Code returns ErrCode
func (e *Error) Code() ErrCode {
	return e.code
}

// Message returns formatted message
func (e *Error) Message() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.message, e.args...)
	}
	return e.message
}

// Underlying returns the Previous error, or nil
// if there is none.
func (e *Error) Underlying() error {
	return e.prev
}

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e *Error) Location() (file string, line int) {
	return e.file, e.line
}

// Error implements error interface.
func (e *Error) Error() string {
	msg := e.Message()
	if e.code == ErrUnknown {
		if msg != "" {
			return msg
		}
		if e.prev != nil {
			return e.prev.Error()
		}
	}
	return fmt.Sprintf("[%d]%s", e.code, msg)
}

// MarshalJSON implements json.Marshaler interface.
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Code ErrCode `json:"code"`
		Msg  string  `json:"message"`
	}{
		Code: e.code,
		Msg:  e.Message(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (e *Error) UnmarshalJSON(data []byte) error {
	err := &struct {
		Code ErrCode `json:"code"`
		Msg  string  `json:"message"`
	}{}

	if err := json.Unmarshal(data, &err); err != nil {
		return Trace(err)
	}

	e.code = err.Code
	e.message = err.Msg
	return nil
}

// Equal checks if err is equal to e.
func (e *Error) Equal(err error) bool {
	originErr := Cause(err)
	if originErr == nil {
		return false
	}

	if error(e) == originErr {
		return true
	}
	inErr, ok := originErr.(*Error)
	return ok && e.code == inErr.code
}

// ToGRPC convert *Error to gRPC error
func (e *Error) ToGRPC() error {
	return status.Error(codes.Code(e.code), e.Message())
}

// Internal for some internal error
func Internal(err error) error {
	if err == nil {
		return nil
	}
	return new(1000, err.Error(), nil, nil, 1)
}

// SQL error
func SQL(err error) error {
	if err == nil {
		return nil
	}
	return new(1001, err.Error(), nil, nil, 1)
}

// RDS error
func RDS(err error) error {
	if err == nil {
		return nil
	}
	return new(1002, err.Error(), nil, nil, 1)
}

// BadRequest error
func BadRequest(message string, args ...interface{}) error {
	return new(1003, message, args, nil, 1)
}

// UnAuthorized error
func UnAuthorized(message string, args ...interface{}) error {
	return new(1004, message, args, nil, 1)
}

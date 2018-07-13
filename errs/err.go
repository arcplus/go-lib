package errs

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// A Code is an unsigned 32-bit error code
type ErrCode uint32

// Error implements error interface and add Code, so
// errors with different message can be compared.
type Error struct {
	code    ErrCode
	message string

	args []interface{}

	// previous error
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

// Wrap wraps err with given code and message
//
// For example:
//  err:=someFunc()
//  if err!=nil{
//    return errs.Wrap(err, 500, "internal err")
// }
func Wrap(err error, code ErrCode, message string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}

	// add trace info
	if _, ok := err.(*Error); !ok {
		err = new(0, err.Error(), nil, nil, 1)
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
func Trace(err error) *Error {
	if err == nil {
		return nil
	}

	inErr := new(0, "", nil, nil, 1)
	_, ok := err.(*Error)
	if !ok {
		inErr.message = err.Error()
		return inErr
	}

	inErr.prev = err
	return inErr
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
	if e.code == 0 {
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

func SQL(err error) *Error {
	if err == nil {
		return nil
	}
	return new(1001, err.Error(), nil, nil, 1)
}

func RDS(err error) *Error {
	if err == nil {
		return nil
	}
	return new(1002, err.Error(), nil, nil, 1)
}

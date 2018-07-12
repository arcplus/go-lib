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

	// file and line hold the source code location where the error was
	// created.
	file string
	line int
}

// New is a drop in replacement for the standard library errors module that records
// the location that the error is created.
//
// For example:
//    return errors.New(401, "missing id")
//
func New(code ErrCode, message string, args ...interface{}) *Error {
	err := &Error{code: code, message: message, args: args}
	_, err.file, err.line, _ = runtime.Caller(1)
	return err
}

// NewErr warp err to *Error with code 0 if err is not *Error
func NewErr(err error) *Error {
	if err == nil {
		return nil
	}

	if v, ok := err.(*Error); ok {
		return v
	}

	inErr := &Error{code: 0, message: err.Error()}
	_, inErr.file, inErr.line, _ = runtime.Caller(1)
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

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e *Error) Location() (file string, line int) {
	return e.file, e.line
}

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.code, e.Message())
}

// StackTrace returns one string for each location recorded in the stack of
// errors. The first value is the originating error, with a line for each
// other annotation or tracing of the error.
func (e *Error) StackTrace() string {
	return ErrorStack(e)
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

// Equal checks if two errors is equal
func Equal(err1, err2 error) bool {
	originErr1, originErr2 := Cause(err1), Cause(err2)
	if originErr1 == originErr2 {
		return true
	}

	if originErr1 == nil || originErr2 == nil {
		return false
	}

	// same Error() return
	if originErr1.Error() == originErr2.Error() {
		return true
	}

	inErr1, _ := originErr1.(*Error)
	inErr2, _ := originErr2.(*Error)

	if inErr1 == nil || inErr2 == nil {
		return false
	}

	if inErr1.code == 0 && inErr2.code == 0 {
		return inErr1.Message() == inErr2.Message()
	}

	return inErr1.code == inErr2.code
}

// CodeEqual check err.(*Error).code==code
func CodeEqual(code ErrCode, err error) bool {
	v, ok := err.(*Error)
	if !ok {
		return false
	}
	return v.code == code
}

// UnWrap unwraps 1st layer err to *Error if possible
func UnWrap(err error) *Error {
	if err == nil {
		return nil
	}

	v, ok := err.(*Error)
	if ok {
		return v
	}

	return &Error{message: err.Error()}
}

package errs

import (
	"fmt"
	"runtime"

	"github.com/arcplus/go-lib/json"
)

const (
	CodeOK         uint32 = 0
	CodeInternal   uint32 = 1000
	CodeBadRequest uint32 = 1400
	CodeUnAuth     uint32 = 1401 // miss token
	CodeForbidden  uint32 = 1403
	CodeNotFound   uint32 = 1404
	CodeNotAllowed uint32 = 1405
	CodeConflict   uint32 = 1409 // conflict error
)

// Error implements error interface and add Code, so
// errors with different message can be compared.
type Error struct {
	code  uint32        // code
	msg   string        // msg
	args  []interface{} // fmt
	alert string        // alert info

	// previous holds the previous error in the error stack, if any.
	prev error

	// file and line hold the source code location where the error was
	// created.
	file string
	line int
}

// Code return err casue code.
func (e *Error) Code() uint32 {
	if e.code == 0 {
		e.code = CodeInternal
	}
	return e.code
}

// Message returns formatted message if args not empty
func (e *Error) Message() string {
	if len(e.args) > 0 {
		return fmt.Sprintf(e.msg, e.args...)
	}
	return e.msg
}

// Alert is used for err hint
func (e *Error) Alert() string {
	return e.alert
}

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code(), e.Message())
}

// Unwrap returns the next error in the error chain.
// If there is no next error, Unwrap returns nil.
func (e *Error) Unwrap() error {
	return e.prev
}

// Location returns the location where the error is created,
// implements juju/errors locationer interface.
func (e *Error) Location() (file string, line int) {
	return e.file, e.line
}

type errJSON struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Alert   string `json:"alert"`
}

// MarshalJSON implements json.Marshaler interface.
func (e *Error) MarshalJSON() ([]byte, error) {
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
	e.msg = info.Message
	e.alert = info.Alert
	return nil
}

func (e *Error) setLocation(callDepth int) {
	_, file, line, _ := runtime.Caller(callDepth + 1)
	e.file = trimGoPath(file)
	e.line = line
}

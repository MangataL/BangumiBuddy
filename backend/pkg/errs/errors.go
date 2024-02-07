package errs

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Error 异常
type Error struct {
	HTTPCode int
	Message  string
}

func (e *Error) Error() string {
	return e.Message
}

func NewBadRequest(message string) *Error {
	return &Error{
		HTTPCode: http.StatusBadRequest,
		Message:  message,
	}
}

func NewUnauthorized(message string) *Error {
	return &Error{
		HTTPCode: http.StatusUnauthorized,
		Message:  message,
	}
}

func NewUnauthorizedf(format string, args ...interface{}) *Error {
	return &Error{
		HTTPCode: http.StatusUnauthorized,
		Message:  fmt.Sprintf(format, args...),
	}
}

func NewNotFound(message string) *Error {
	return &Error{
		HTTPCode: http.StatusNotFound,
		Message:  message,
	}
}

func NewNotFoundf(format string, args ...interface{}) *Error {
	return &Error{
		HTTPCode: http.StatusNotFound,
		Message:  fmt.Sprintf(format, args...),
	}
}

func ParseError(err error) (int, string) {
	cause := errors.Cause(err)
	if e := (&Error{}); errors.As(cause, &e) {
		return e.HTTPCode, e.Error()
	}
	return http.StatusInternalServerError, err.Error()
}

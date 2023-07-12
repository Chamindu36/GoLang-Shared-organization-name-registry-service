package errors

import (
	"fmt"
	"strings"
)

// Service type errors
type serviceError struct {
	e *Error
}

// Create a custom Error object with srtring builder
func (s serviceError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`{"code":%d,"msg":%s}`, s.e.Code, s.e.Message))
	if len(s.e.Cause) > 0 {
		sb.WriteString(fmt.Sprintf(`", "cause": %s"`, s.e.Cause))
	}
	return sb.String()
}

// Custom Error structure
type Error struct {
	state   string    `json:"state"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   string    `json:"cause"`
}

// New method will define a new custom Error object
// @param code ErrorCode object which is already defined
// @param msg message string of the error
// @param err native error object
// @return error
func New(code ErrorCode, msg string, err error) error {
	return &serviceError{
		e: &Error{
			Code:    code,
			Message: msg,
			Cause: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		},
	}
}

// Newf method will initialize a new custom Error object
// @param code ErrorCode object which is already defined
// @param err native error object
// @param format format of the string builder
// @return error
func Newf(code ErrorCode, err error, format string, a ...interface{}) error {
	return New(code, fmt.Sprintf(format, a...), err)
}

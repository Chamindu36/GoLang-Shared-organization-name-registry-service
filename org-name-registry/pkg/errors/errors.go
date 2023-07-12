// Package that handle all the errors
package errors

import (
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type ErrorCode int32

// statusToError method will convert status values into Error values
// @param st Status
// @return error
func statusToError(st *status.Status) *Error {
	if st == nil || st.Code() == codes.OK {
		return nil
	}
	var errCode ErrorCode
	// fixme: add correct error codes
	switch st.Code() {
	case codes.Canceled:
		errCode = Error_INTERNAL
	case codes.InvalidArgument:
		errCode = Error_INVALID_REQUEST
	case codes.DeadlineExceeded:
		errCode = Error_INTERNAL
	case codes.NotFound:
		errCode = Error_NOT_FOUND
	case codes.PermissionDenied:
		errCode = Error_INTERNAL
	case codes.Unauthenticated:
		errCode = Error_NOT_AUTHORIZED
	case codes.ResourceExhausted:
		errCode = Error_INTERNAL
	case codes.FailedPrecondition:
		errCode = Error_INTERNAL
	case codes.Aborted:
		errCode = Error_INTERNAL
	case codes.OutOfRange:
		errCode = Error_INTERNAL
	case codes.Unimplemented:
		errCode = Error_UNIMPLEMENTED
	case codes.Internal:
		errCode = Error_INTERNAL
	case codes.Unavailable:
		errCode = Error_UNAVAILABLE
	case codes.DataLoss:
		errCode = Error_INTERNAL
	case codes.Unknown:
		errCode = Error_UNKNOWN_SERVER
	case codes.AlreadyExists:
		errCode = Error_DUPLICATE_RESERVATION
	default:
		errCode = Error_UNKNOWN
	}
	return &Error{
		Code:    errCode,
		Message: st.Message(),
	}
}

const (
	Error_UNKNOWN               ErrorCode = 0
	Error_INVALID_REQUEST       ErrorCode = 1002
	Error_NOT_FOUND             ErrorCode = 1003
	Error_NOT_AUTHORIZED        ErrorCode = 1004
	Error_DUPLICATE_RESERVATION ErrorCode = 1012
	Error_UNKNOWN_SERVER        ErrorCode = 2001
	Error_INTERNAL              ErrorCode = 2002
	Error_UNAVAILABLE           ErrorCode = 2003
	Error_UNIMPLEMENTED         ErrorCode = 2004
)

var (
	Error_Code_name = map[int32]string{
		0:    "UNKNOWN",
		1001: "UNKNOWN_CLIENT",
		1002: "INVALID_REQUEST",
		1003: "NOT_FOUND",
		1004: "NOT_AUTHORIZED",
		1012: "DUPLICATE_RESERVATION",
		2001: "UNKNOWN_SERVER",
		2002: "INTERNAL",
		2003: "UNAVAILABLE",
		2004: "UNIMPLEMENTED",
	}
	Error_Code_value = map[string]int32{
		"UNKNOWN":                 0,
		"UNKNOWN_CLIENT":          1001,
		"INVALID_REQUEST":         1002,
		"NOT_FOUND":               1003,
		"NOT_AUTHORIZED":          1004,
		"INVITATION_NOT_FOUND":    1010,
		"INVITATION_ALREADY_USED": 1011,
		"DUPLICATE_RESERVATION":   1012,
		"UNKNOWN_SERVER":          2001,
		"INTERNAL":                2002,
		"UNAVAILABLE":             2003,
		"UNIMPLEMENTED":           2004,
	}
)

func (x ErrorCode) Enum() *ErrorCode {
	p := new(ErrorCode)
	*p = x
	return p
}
func Is(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	e := FromError(err)
	if e == nil {
		return false
	}
	return e.Code == code
}

// FromError method will extract error and create a custom Error object
// @param err error object
// @return Error custom Error object
func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	if ae, ok := err.(*serviceError); ok {
		return ae.e
	}

	if st, ok := status.FromError(err); ok {
		// check grpc status details contains *apis.Error object
		for _, detail := range st.Details() {
			switch t := detail.(type) {
			case *Error:
				return t
			}
		}
		return statusToError(st)
	}
	return &Error{
		Code:    Error_UNKNOWN,
		Message: "unknown error",
		Cause:   err.Error(),
	}
}

// JsonError method will return the error type in Json string format
// @param w response object
// @param err Custom error object
// @param code custom error code
func JsonError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&err)
}

func NewInvalidRequest(msg string, err error) error {
	return New(Error_INVALID_REQUEST, msg, err)
}

func NewInternal(msg string, err error) error {
	return New(Error_INTERNAL, msg, err)
}

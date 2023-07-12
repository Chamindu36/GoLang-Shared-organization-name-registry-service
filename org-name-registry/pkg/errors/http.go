package errors

import (
	"github.com/golang/protobuf/jsonpb"
	"net/http"
)

type MarshalOption func(*jsonpb.Marshaler)

func WriteHttp(w http.ResponseWriter, err error) {
	e := FromError(err)

	var statusCode int
	switch e.Code {
	case Error_INVALID_REQUEST:
		statusCode = http.StatusBadRequest
	case Error_DUPLICATE_RESERVATION:
		statusCode = http.StatusConflict
	case Error_INTERNAL:
		statusCode = http.StatusInternalServerError
	case Error_NOT_FOUND:
		statusCode = http.StatusNotFound
	case Error_UNAVAILABLE:
		statusCode = http.StatusServiceUnavailable
	case Error_NOT_AUTHORIZED:
		statusCode = http.StatusUnauthorized
	case Error_UNIMPLEMENTED:
		statusCode = http.StatusNotFound // todo: should this be 500?
	default:
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)
}

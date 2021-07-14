package v3ioerrors

import (
	"errors"

	v3io "github.com/v3io/v3io-go/pkg/dataplane"
)

var ErrInvalidTypeConversion = errors.New("Invalid type conversion")
var ErrNotFound = errors.New("Not found")
var ErrStopped = errors.New("Stopped")
var ErrTimeout = errors.New("Timed out")

type ErrorWithStatusCode struct {
	error
	statusCode int
}

type ErrorWithStatusCodeAndResponse struct {
	ErrorWithStatusCode
	response *v3io.Response
}

func NewErrorWithStatusCode(err error, statusCode int) ErrorWithStatusCode {
	return ErrorWithStatusCode{
		error:      err,
		statusCode: statusCode,
	}
}

func (e ErrorWithStatusCode) StatusCode() int {
	return e.statusCode
}

func (e ErrorWithStatusCode) Error() string {
	return e.error.Error()
}

func NewErrorWithStatusCodeAndResponse(err error,
	statusCode int,
	response *v3io.Response) ErrorWithStatusCodeAndResponse {

	return ErrorWithStatusCodeAndResponse{
		ErrorWithStatusCode: NewErrorWithStatusCode(err, statusCode),
		response:            response,
	}
}

func (e ErrorWithStatusCodeAndResponse) Response() *v3io.Response {
	return e.response
}

package v3ioerrors

import (
	"errors"

	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/valyala/fasthttp"
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

	// Copy response struct since original response will be released if error is encountered
	var _copiedResponse *v3io.Response
	if response != nil {
		copiedHTTPResponse := &fasthttp.Response{}
		if response.HTTPResponse != nil {
			response.HTTPResponse.CopyTo(copiedHTTPResponse)
		}
		_copiedResponse = &v3io.Response{
			Output:       response.Output,
			Context:      response.Context,
			Error:        response.Error,
			HTTPResponse: copiedHTTPResponse,
		}
	}

	return ErrorWithStatusCodeAndResponse{
		ErrorWithStatusCode: NewErrorWithStatusCode(err, statusCode),
		response:            _copiedResponse,
	}
}

func (e ErrorWithStatusCodeAndResponse) Response() *v3io.Response {
	return e.response
}

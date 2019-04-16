package v3ioerrors

import (
	"errors"
	"fmt"
)

var ErrInvalidTypeConversion = errors.New("Invalid type conversion")

var ErrTimeout = errors.New("Timed out")

type ErrorWithStatusCode struct {
	error
	statusCode int
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
	return fmt.Sprintf("%s (%d response code)", e.error.Error(), e.statusCode)
}

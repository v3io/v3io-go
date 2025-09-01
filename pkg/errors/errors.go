/*
Copyright 2019 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License") with
an addition restriction as set forth herein. You may not use this
file except in compliance with the License. You may obtain a copy of
the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
implied. See the License for the specific language governing
permissions and limitations under the License.

In addition, you may not use the software for any purposes that are
illegal under applicable law, and the grant of the foregoing license
under the Apache 2.0 license is conditioned upon your compliance with
such restriction.
*/

package v3ioerrors

import (
	"errors"
)

var ErrInvalidTypeConversion = errors.New("Invalid type conversion") //nolint:staticcheck // ST1005
var ErrNotFound = errors.New("Not found") //nolint:staticcheck // ST1005
var ErrStopped = errors.New("Stopped")
var ErrTimeout = errors.New("Timed out") //nolint:staticcheck // ST1005

type ErrorWithStatusCode struct {
	error
	statusCode int
}

type ErrorWithStatusCodeAndResponse struct {
	ErrorWithStatusCode
	response interface{}
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
	response interface{}) ErrorWithStatusCodeAndResponse {

	return ErrorWithStatusCodeAndResponse{
		ErrorWithStatusCode: NewErrorWithStatusCode(err, statusCode),
		response:            response,
	}
}

func (e ErrorWithStatusCodeAndResponse) Response() interface{} {
	return e.response
}

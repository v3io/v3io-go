/*
Copyright 2025 Iguazio Systems Ltd.

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

package common

import (
	"fmt"
	"testing"

	"github.com/nuclio/errors"
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"

	"github.com/stretchr/testify/suite"
)

type helperTestSuite struct {
	suite.Suite
}

func (suite *helperTestSuite) TestEngineErrorIsNonFatalNestedErrorFromLog() {
	// Create the nested error structure with http 503 Service Temporarily Unavailable as the root cause

	// Create HTTP response string contains the 503 error
	httpResponseStr := `HTTP/1.1 503 Service Temporarily Unavailable\r\nServer: nginx\r\nDate: Tue, 03 Jun 2025 07:37:09 GMT\r\nContent-Type: application/json\r\nContent-Length: 89\r\nConnection: keep-alive\r\n\r\n{\n\t\"ErrorCode\": -117440512,\n\t\"ErrorMessage\": \"Failed to send a control message request\"\n}`
	sanitizedRequest := `PUT /projects/perform044wp2gqixmhkv1/datafetch_output_stream/20 HTTP/1.1\r\nUser-Agent: fasthttp\r\nHost: v3io-webapi:8081\r\nContent-Type: application/json\r\nContent-Length: 58\r\nX-V3io-Session-Key: SANITIZED\r\nX-V3io-Function: GetItem\r\n\r\n{\"AttributesToGet\": \"__serving_committed_sequence_number\"}`
	httpBody := fmt.Errorf("Expected a 2xx response status code: %s\nRequest details:\n%s",
		httpResponseStr, sanitizedRequest)

	// Wrap with ErrorWithStatusCode
	statusCodeErr := v3ioerrors.NewErrorWithStatusCode(httpBody, 503)

	// Further wrap the error to simulate the nested error chain
	shardItemErr := errors.Wrap(statusCodeErr, "Failed getting shard item")
	sequenceNumberErr := errors.Wrap(shardItemErr, "Failed to get shard sequenceNumber from item attributes")
	persistencyErr := errors.Wrap(sequenceNumberErr, "Failed to get shard location from persistency")
	locationErr := errors.Wrap(persistencyErr, "Failed to get shard location")
	finalErr := errors.Wrapf(locationErr, "Failed to get shard location state, attempts exhausted. shard id: %d", 20)

	// Test that EngineErrorIsNonFatal correctly unwraps the nested error chain
	// Since status code 503 is now in nonFatalStatusCodes, expect true
	result := EngineErrorIsNonFatal(finalErr)
	suite.Require().True(result, "Expected EngineErrorIsNonFatal to return true (503 is in nonFatalStatusCodes)")
}

func (suite *helperTestSuite) TestMatchErrorString() {
	for _, testCase := range []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "timeout error",
			errMsg:   "connection timeout occurred",
			expected: true,
		},
		{
			name:     "dial timeout error",
			errMsg:   "dialing to the given TCP address timed out",
			expected: true,
		},
		{
			name:     "connection refused",
			errMsg:   "connection refused",
			expected: true,
		},
		{
			name:     "generic error",
			errMsg:   "something went wrong",
			expected: false,
		},
		{
			name:     "empty error",
			errMsg:   "",
			expected: false,
		},
	} {
		suite.Run(testCase.name, func() {
			err := fmt.Errorf(testCase.errMsg)
			result := isNonFatalErrorString(err)
			suite.Require().Equal(testCase.expected, result)
		})
	}
}

func (suite *helperTestSuite) TestMatchErrorStatusCode() {
	baseErr := fmt.Errorf("test error")
	for _, testCase := range []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "non-fatal status code 503",
			statusCode: 503,
			expected:   true,
		},
		{
			name:       "fatal status code 500",
			statusCode: 500,
			expected:   false,
		},
		{
			name:       "fatal status code 404",
			statusCode: 404,
			expected:   false,
		},
		{
			name:       "fatal status code 200",
			statusCode: 200,
			expected:   false,
		},
	} {
		suite.Run(testCase.name, func() {
			err := v3ioerrors.NewErrorWithStatusCode(baseErr, testCase.statusCode)
			result := isNonFatalStatusCode(err)
			suite.Require().Equal(testCase.expected, result)
		})
	}
}

func (suite *helperTestSuite) TestMatchErrorStatusCodeNonV3ioError() {
	err := fmt.Errorf("regular error")
	result := isNonFatalStatusCode(err)
	suite.Require().False(result)
}

func (suite *helperTestSuite) TestEngineErrorIsNonFatalStringMatch() {
	for _, testCase := range []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "timeout error",
			errMsg:   "operation timeout",
			expected: true,
		},
		{
			name:     "connection refused",
			errMsg:   "connection refused by server",
			expected: true,
		},
		{
			name:     "generic error",
			errMsg:   "something went wrong",
			expected: false,
		},
	} {
		suite.Run(testCase.name, func() {
			err := fmt.Errorf(testCase.errMsg)
			result := EngineErrorIsNonFatal(err)
			suite.Require().Equal(testCase.expected, result)
		})
	}
}

func (suite *helperTestSuite) TestEngineErrorIsNonFatalStatusCodeMatch() {
	baseErr := fmt.Errorf("test error")
	for _, testCase := range []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "Service Temporarily Unavailable error",
			statusCode: 503,
			expected:   true,
		}, {
			name:       "no error",
			statusCode: 200,
			expected:   false,
		},
	} {
		suite.Run(testCase.name, func() {
			err := v3ioerrors.NewErrorWithStatusCode(baseErr, testCase.statusCode)
			result := EngineErrorIsNonFatal(err)
			suite.Require().Equal(testCase.expected, result)
		})
	}

}

func TestHelperTestSuite(t *testing.T) {
	suite.Run(t, new(helperTestSuite))
}

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

package v3iohttp

import (
	"encoding/base64"
	"fmt"

	v3io "github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/logger"
)

type session struct {
	logger              logger.Logger
	context             *context
	url                 string
	authenticationToken string
	accessKey           string
}

func newSession(parentLogger logger.Logger,
	context *context,
	url string,
	username string,
	password string,
	accessKey string) (v3io.Session, error) {

	authenticationToken := ""
	if username != "" && password != "" && accessKey == "" {
		authenticationToken = GenerateAuthenticationToken(username, password)
	}

	return &session{
		logger:              parentLogger.GetChild("session"),
		context:             context,
		url:                 url,
		authenticationToken: authenticationToken,
		accessKey:           accessKey,
	}, nil
}

// NewContainer creates a container
func (s *session) NewContainer(newContainerInput *v3io.NewContainerInput) (v3io.Container, error) {
	return newContainer(s.logger, s, newContainerInput.ContainerName)
}

func GenerateAuthenticationToken(username string, password string) string {

	// generate token for basic authentication
	usernameAndPassword := fmt.Sprintf("%s:%s", username, password)
	encodedUsernameAndPassword := base64.StdEncoding.EncodeToString([]byte(usernameAndPassword))

	return "Basic " + encodedUsernameAndPassword
}

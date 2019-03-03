/*
Copyright 2018 The v3io Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3io

import (
	"os"
	"testing"
	"time"

	"github.com/v3io/v3io-go"
	// import types
	_ "github.com/v3io/v3io-go/pkg/controlplane/http"
	_ "github.com/v3io/v3io-go/pkg/dataplane/http"

	"github.com/nuclio/logger"
	"github.com/nuclio/zap"
	"github.com/stretchr/testify/suite"
)

type clientSuite struct {
	suite.Suite
	logger logger.Logger
}

func (suite *clientSuite) SetupTest() {
	suite.logger, _ = nucliozap.NewNuclioZapTest("test")
}

func (suite *clientSuite) TestFlow() {
	client, err := v3io.NewClient(suite.logger, []string{
		"https://" + os.Getenv("HOST"),
	})
	suite.Require().NoError(err)

	testID := "32"

	controlPlaneSession := suite.createControlPlaneSession(client,
		os.Getenv("USER_CREATOR_USERNAME"),
		os.Getenv("USER_CREATOR_PASSWORD"))

	// prepare create user request
	createUserInput := v3io.CreateUserInput{}
	createUserInput.Username = "erand" + testID
	createUserInput.Password = "123456"
	createUserInput.Email = "erand@iguazio.com-" + testID
	createUserInput.DataAccessMode = "enabled"
	createUserInput.FirstName = "Eran-" + testID
	createUserInput.LastName = "Duchan-" + testID
	createUserInput.AssignedPolicies = []string{"Data", "Application Admin"}

	// create user
	createUserOutput, err := controlPlaneSession.CreateUserSync(&createUserInput)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(createUserOutput.ID)

	time.Sleep(10 * time.Second)

	controlPlaneSession = suite.createControlPlaneSession(client,
		createUserInput.Username,
		createUserInput.Password)

	// prepare create container request
	createContainerInput := v3io.CreateContainerInput{}
	createContainerInput.Name = "bigdata-" + testID

	// create user
	createContainerOutput, err := controlPlaneSession.CreateContainerSync(&createContainerInput)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(createContainerOutput.ID)
}

func (suite *clientSuite) createControlPlaneSession(client *v3io.Client,
	username string,
	password string) v3io.ControlplaneSession {

	// prepare create session request
	createSessionInput := v3io.CreateSessionInput{}
	createSessionInput.Username = username
	createSessionInput.Password = password
	createSessionInput.Planes = v3io.SessionPlaneControl

	// create session
	createSessionOutput, err := client.CreateSessionSync(&createSessionInput)

	suite.Require().NoError(err)

	// get control plane session
	controlPlaneSession := createSessionOutput.Session.ControlPlane()
	suite.Require().NotNil(controlPlaneSession)

	return controlPlaneSession
}

func TestWriterTestSuite(t *testing.T) {
	suite.Run(t, new(clientSuite))
}

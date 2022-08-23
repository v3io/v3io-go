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
package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/v3io/v3io-go/pkg/controlplane"
	"github.com/v3io/v3io-go/pkg/controlplane/http"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/logger"
	"github.com/nuclio/zap"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	logger          logger.Logger
	session         v3ioc.Session
	controlplaneURL string
	userID          string
	ctx             context.Context
}

func (suite *testSuite) SetupSuite() {
	suite.controlplaneURL = os.Getenv("V3IO_CONTROLPLANE_URL")
	suite.logger, _ = nucliozap.NewNuclioZapTest("test")

	// create a security admin session
	newSessionInput := v3ioc.NewSessionInput{}
	newSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	newSessionInput.Password = os.Getenv("V3IO_CONTROLPLANE_PASSWORD")
	newSessionInput.Endpoints = []string{suite.controlplaneURL}

	session, err := v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Require().NoError(err, fmt.Sprintf("\nInput: %v\n", newSessionInput))

	// create a unique user for the tests
	ts := time.Now().Unix()
	createUserInput := v3ioc.CreateUserInput{}
	createUserInput.Ctx = suite.ctx
	createUserInput.FirstName = fmt.Sprintf("Test-%d", ts)
	createUserInput.LastName = fmt.Sprintf("User-%d", ts)
	createUserInput.Username = fmt.Sprintf("testuser-%d", ts)
	createUserInput.Password = fmt.Sprintf("Testpasswd-%d!", ts)
	createUserInput.Email = fmt.Sprintf("testuser-%d@user.com", ts)
	createUserInput.Description = "A user created from tests"
	createUserInput.AssignedPolicies = []string{"Security Admin", "Data", "Application Admin"}

	// create a user with security session
	createUserOutput, err := session.CreateUserSync(&createUserInput)
	suite.Require().NoError(err)
	suite.Require().NotNil(createUserOutput.ID)
	suite.userID = createUserOutput.ID

	// create new user's session
	newSessionInput = v3ioc.NewSessionInput{}
	newSessionInput.Username = createUserInput.Username
	newSessionInput.Password = createUserInput.Password
	newSessionInput.Endpoints = []string{suite.controlplaneURL}

	newUserSession, err := v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Require().NoError(err, fmt.Sprintf("\nInput: %v\n", newSessionInput))

	getRunningUserAttributesInput := v3ioc.GetRunningUserAttributesInput{}
	getRunningUserAttributesInput.Ctx = suite.ctx

	getRunningUserAttributesOutput, err := newUserSession.GetRunningUserAttributesSync(&getRunningUserAttributesInput)
	suite.Require().NoError(err)
	suite.Require().Equal(getRunningUserAttributesOutput.Username, createUserInput.Username)

	// create a session with that user
	newSessionInput.Username = createUserInput.Username
	newSessionInput.Password = createUserInput.Password
	newSessionInput.Endpoints = []string{suite.controlplaneURL}

	suite.session, err = v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Require().NoError(err)

	time.Sleep(30 * time.Second)
}

func (suite *testSuite) TearDownSuite() {
	deleteUserInput := v3ioc.DeleteUserInput{}
	deleteUserInput.ID = suite.userID

	err := suite.session.DeleteUserSync(&deleteUserInput)
	suite.Require().NoError(err)
}

func (suite *testSuite) SetupTest() {
	suite.ctx = context.WithValue(context.TODO(), "RequestID", "test-0")
}

func (suite *testSuite) TestCreateContainer() {
	createContainerInput := v3ioc.CreateContainerInput{}
	createContainerInput.Ctx = suite.ctx
	createContainerInput.Name = "container-string"

	createContainerOutput, err := suite.session.CreateContainerSync(&createContainerInput)
	suite.Require().NoError(err)
	suite.Require().NotEqual(0, createContainerOutput.IDNumeric)

	time.Sleep(5 * time.Second)

	deleteContainerInput := v3ioc.DeleteContainerInput{}
	deleteContainerInput.Ctx = suite.ctx
	deleteContainerInput.IDNumeric = createContainerOutput.IDNumeric

	err = suite.session.DeleteContainerSync(&deleteContainerInput)
	suite.Require().NoError(err)
}

func (suite *testSuite) TestCreateSessionWithTimeout() {

	// create a security admin session
	newSessionInput := v3ioc.NewSessionInput{}
	newSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	newSessionInput.Password = os.Getenv("V3IO_CONTROLPLANE_PASSWORD")
	newSessionInput.Endpoints = []string{os.Getenv("V3IO_CONTROLPLANE_URL")}
	newSessionInput.Timeout = 1 * time.Millisecond

	session, err := v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Require().Equal(v3ioerrors.ErrTimeout, err)
	suite.Require().Nil(session)
}

func (suite *testSuite) TestCreateSessionWithBadPassword() {

	// create a security admin session
	newSessionInput := v3ioc.NewSessionInput{}
	newSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	newSessionInput.Password = "WRONG"
	newSessionInput.Endpoints = []string{os.Getenv("V3IO_CONTROLPLANE_URL")}

	session, err := v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Equal(401, err.(v3ioerrors.ErrorWithStatusCode).StatusCode())
	suite.Require().Nil(session)
}

func (suite *testSuite) TestCreateEventUsingAccessKey() {

	// Create new access key
	createAccessKeyInput := v3ioc.CreateAccessKeyInput{}
	createAccessKeyInput.Ctx = suite.ctx
	createAccessKeyInput.Label = "test_access_key_label"
	createAccessKeyInput.Plane = v3ioc.ControlPlane

	createAccessKeyOutput, err := suite.session.CreateAccessKeySync(&createAccessKeyInput)
	suite.Require().NoError(err)
	suite.Require().Equal(createAccessKeyOutput.Label, createAccessKeyInput.Label)

	// Create new session from access key
	newSessionInput := v3ioc.NewSessionInput{}
	newSessionInput.AccessKey = createAccessKeyOutput.ID
	newSessionInput.Endpoints = []string{os.Getenv("V3IO_CONTROLPLANE_URL")}
	accessKeySession, err := v3iochttp.NewSession(suite.logger, &newSessionInput)
	suite.Require().NoError(err)

	// Emit event
	createEventInput := v3ioc.CreateEventInput{}
	createEventInput.Ctx = suite.ctx
	createEventInput.Kind = "AppService.Test.Event"
	createEventInput.Source = "DummyService"

	err = accessKeySession.CreateEventSync(&createEventInput)
	suite.Require().NoError(err)

	// Delete access key
	deleteAccessKeyInput := v3ioc.DeleteAccessKeyInput{}
	deleteAccessKeyInput.ID = createAccessKeyOutput.ID
	deleteAccessKeyInput.Ctx = suite.ctx
	err = suite.session.DeleteAccessKeySync(&deleteAccessKeyInput)
	suite.Require().NoError(err)
}

func (suite *testSuite) TestReloadConfigurations() {

	// only igz_admin can reload configurations, it is a maintenance operation
	session := suite.createIGZAdminSession()
	retryInterval := 3 * time.Second
	timeout := 2 * time.Minute
	var err error

	suite.logger.InfoWith("Reloading cluster configuration")
	err = session.ReloadClusterConfigAndWaitForCompletion(context.TODO(), retryInterval, timeout)
	suite.Require().NoError(err)

	suite.logger.InfoWith("Reloading events configuration")
	err = session.ReloadEventsConfigAndWaitForCompletion(context.TODO(), retryInterval, timeout)
	suite.Require().NoError(err)

	suite.logger.InfoWith("Reloading app services configuration")
	err = session.ReloadAppServicesConfigAndWaitForCompletion(context.TODO(), retryInterval, timeout)
	suite.Require().NoError(err)

	suite.logger.InfoWith("Reloading artifact version manifest configuration")
	err = session.ReloadArtifactVersionManifestAndWaitForCompletion(context.TODO(), retryInterval, timeout)
	suite.Require().NoError(err)
}

func (suite *testSuite) createIGZAdminSession() v3ioc.Session {
	igzAdminSessionInput := v3ioc.NewSessionInput{}
	igzAdminSessionInput.Username = "igz_admin"
	igzAdminSessionInput.Password = os.Getenv("V3IO_CONTROLPLANE_IGZ_ADMIN_PASSWORD")
	igzAdminSessionInput.Endpoints = []string{suite.controlplaneURL}

	session, err := v3iochttp.NewSession(suite.logger, &igzAdminSessionInput)
	suite.Require().NoError(err, fmt.Sprintf("\nInput: %v\n", igzAdminSessionInput))
	suite.logger.InfoWith("Successfully created session for igz_admin", "session", session)
	return session
}

func TestControlPlaneTestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

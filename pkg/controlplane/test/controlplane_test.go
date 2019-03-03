package test

import (
	"context"
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

type githubClientSuite struct {
	suite.Suite
	logger  logger.Logger
	session v3ioc.Session
	userID  string
	ctx     context.Context
}

func (suite *githubClientSuite) SetupSuite() {
	controlplaneURL := os.Getenv("V3IO_CONTROLPLANE_URL")
	suite.logger, _ = nucliozap.NewNuclioZapTest("test")

	// create a security admin session
	createSessionInput := v3ioc.CreateSessionInput{}
	createSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	createSessionInput.Password = os.Getenv("V3IO_CONTROLPLANE_PASSWORD")
	createSessionInput.Endpoints = []string{controlplaneURL}

	session, err := v3iochttp.NewSession(suite.logger, &createSessionInput)
	suite.Require().NoError(err)

	// create a user for the tests
	createUserInput := v3ioc.CreateUserInput{}
	createUserInput.Ctx = suite.ctx
	createUserInput.FirstName = "Test"
	createUserInput.LastName = "User"
	createUserInput.Username = "testuser0"
	createUserInput.Password = "testpass0"
	createUserInput.Email = "test@user.com"
	createUserInput.Description = "A user created from tests"
	createUserInput.AssignedPolicies = []string{"Security Admin", "Data", "Application Admin", "Function Admin"}

	// create a user with security session
	createUserOutput, err := session.CreateUserSync(&createUserInput)
	suite.Require().NotNil(createUserOutput.ID)
	suite.userID = createUserOutput.ID

	// create a session with that user
	createSessionInput.Username = createUserInput.Username
	createSessionInput.Password = createUserInput.Password
	createSessionInput.Endpoints = []string{controlplaneURL}

	suite.session, err = v3iochttp.NewSession(suite.logger, &createSessionInput)
	suite.Require().NoError(err)

	time.Sleep(30 * time.Second)
}

func (suite *githubClientSuite) TearDownSuite() {
	deleteUserInput := v3ioc.DeleteUserInput{}
	deleteUserInput.ID = suite.userID

	err := suite.session.DeleteUserSync(&deleteUserInput)
	suite.Require().NoError(err)
}

func (suite *githubClientSuite) SetupTest() {
	suite.ctx = context.WithValue(nil, "RequestID", "test-0")
}

func (suite *githubClientSuite) TestCreateContainerStringID() {
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

func (suite *githubClientSuite) TestCreateContainerNumericID() {
	createContainerInput := v3ioc.CreateContainerInput{}
	createContainerInput.Ctx = suite.ctx
	createContainerInput.IDNumeric = 300
	createContainerInput.Name = "container-int"

	createContainerOutput, err := suite.session.CreateContainerSync(&createContainerInput)
	suite.Require().NoError(err)
	suite.Require().Equal(createContainerInput.IDNumeric, createContainerOutput.IDNumeric)

	time.Sleep(5 * time.Second)

	deleteContainerInput := v3ioc.DeleteContainerInput{}
	deleteContainerInput.Ctx = suite.ctx
	deleteContainerInput.IDNumeric = createContainerOutput.IDNumeric

	err = suite.session.DeleteContainerSync(&deleteContainerInput)
	suite.Require().NoError(err)
}

func (suite *githubClientSuite) TestCreateSessionWithTimeout() {

	// create a security admin session
	createSessionInput := v3ioc.CreateSessionInput{}
	createSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	createSessionInput.Password = os.Getenv("V3IO_CONTROLPLANE_PASSWORD")
	createSessionInput.Endpoints = []string{os.Getenv("V3IO_CONTROLPLANE_URL")}
	createSessionInput.Timeout = 1 * time.Millisecond

	session, err := v3iochttp.NewSession(suite.logger, &createSessionInput)
	suite.Require().Equal(v3ioerrors.ErrTimeout, err)
	suite.Require().Nil(session)
}

func (suite *githubClientSuite) TestCreateSessionWithBadPassword() {

	// create a security admin session
	createSessionInput := v3ioc.CreateSessionInput{}
	createSessionInput.Username = os.Getenv("V3IO_CONTROLPLANE_USERNAME")
	createSessionInput.Password = "WRONG"
	createSessionInput.Endpoints = []string{os.Getenv("V3IO_CONTROLPLANE_URL")}

	session, err := v3iochttp.NewSession(suite.logger, &createSessionInput)
	suite.Equal(401, err.(*v3ioerrors.ErrorWithStatusCode).StatusCode())
	suite.Require().Nil(session)
}

func TestGithubClientTestSuite(t *testing.T) {
	suite.Run(t, new(githubClientSuite))
}

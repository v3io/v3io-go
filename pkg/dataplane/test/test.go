package test

import (
	"github.com/nuclio/logger"
	"github.com/nuclio/zap"
	"github.com/stretchr/testify/suite"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/http"
	"os"
)

type testSuite struct {
	suite.Suite
	logger              logger.Logger
	container           v3io.Container
	containerName       string
	authenticationToken string
	accessKey           string
}

func (suite *testSuite) SetupSuite() {
	suite.logger, _ = nucliozap.NewNuclioZapTest("test")
}

func (suite *testSuite) populateDataPlaneInput(dataPlaneInput *v3io.DataPlaneInput) {
	dataPlaneInput.ContainerName = suite.containerName
	dataPlaneInput.AuthenticationToken = suite.authenticationToken
	dataPlaneInput.AccessKey = suite.accessKey
}

func (suite *testSuite) createContext() {
	var err error

	// create a context
	suite.container, err = v3iohttp.NewContext(suite.logger, []string{os.Getenv("V3IO_DATAPLANE_URL")}, 8)
	suite.Require().NoError(err)

	// populate fields that would have been populated by session/container
	suite.containerName = "bigdata"
	suite.authenticationToken = v3iohttp.GenerateAuthenticationToken(os.Getenv("V3IO_DATAPLANE_USERNAME"),
		os.Getenv("V3IO_DATAPLANE_PASSWORD"))
}

func (suite *testSuite) createContainer() {

	// create a context
	context, err := v3iohttp.NewContext(suite.logger, []string{os.Getenv("V3IO_DATAPLANE_URL")}, 8)
	suite.Require().NoError(err)

	session, err := context.NewSessionSync(&v3io.NewSessionInput{
		Username: os.Getenv("V3IO_DATAPLANE_USERNAME"),
		Password: os.Getenv("V3IO_DATAPLANE_PASSWORD"),
	})
	suite.Require().NoError(err)

	suite.container, err = session.NewContainer(&v3io.NewContainerInput{
		ContainerName: "bigdata",
	})
	suite.Require().NoError(err)
}

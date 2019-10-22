package test

import (
	"os"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/http"

	"github.com/nuclio/logger"
	"github.com/nuclio/zap"
	"github.com/stretchr/testify/suite"
)

type testSuite struct { // nolint: deadcode
	suite.Suite
	logger              logger.Logger
	container           v3io.Container
	url                 string
	containerName       string
	authenticationToken string
	accessKey           string
}

func (suite *testSuite) SetupSuite() {
	suite.logger, _ = nucliozap.NewNuclioZapTest("test")
}

func (suite *testSuite) populateDataPlaneInput(dataPlaneInput *v3io.DataPlaneInput) {
	dataPlaneInput.URL = suite.url
	dataPlaneInput.ContainerName = suite.containerName
	dataPlaneInput.AuthenticationToken = suite.authenticationToken
	dataPlaneInput.AccessKey = suite.accessKey
}

func (suite *testSuite) createContext() {
	var err error

	// create a context
	suite.container, err = v3iohttp.NewContext(suite.logger, &v3io.NewContextInput{})
	suite.Require().NoError(err)

	// populate fields that would have been populated by session/container
	suite.containerName = "bigdata"

	suite.url = os.Getenv("V3IO_DATAPLANE_URL")
	username := os.Getenv("V3IO_DATAPLANE_USERNAME")
	password := os.Getenv("V3IO_DATAPLANE_PASSWORD")

	if username != "" && password != "" {
		suite.authenticationToken = v3iohttp.GenerateAuthenticationToken(username, password)
	}

	suite.accessKey = os.Getenv("V3IO_DATAPLANE_ACCESS_KEY")
}

func (suite *testSuite) createContainer() {

	// create a context
	context, err := v3iohttp.NewContext(suite.logger, &v3io.NewContextInput{})
	suite.Require().NoError(err)

	session, err := context.NewSession(&v3io.NewSessionInput{
		URL:       os.Getenv("V3IO_DATAPLANE_URL"),
		Username:  os.Getenv("V3IO_DATAPLANE_USERNAME"),
		Password:  os.Getenv("V3IO_DATAPLANE_PASSWORD"),
		AccessKey: os.Getenv("V3IO_DATAPLANE_ACCESS_KEY"),
	})
	suite.Require().NoError(err)

	suite.container, err = session.NewContainer(&v3io.NewContainerInput{
		ContainerName: "bigdata",
	})
	suite.Require().NoError(err)
}

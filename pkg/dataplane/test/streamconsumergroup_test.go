package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/streamconsumergroup"

	"github.com/nuclio/logger"
	"github.com/stretchr/testify/suite"
)

type streamConsumerGroupTestSuite struct {
	StreamTestSuite
}

func (suite *streamConsumerGroupTestSuite) SetupSuite() {
	suite.StreamTestSuite.SetupSuite()
	suite.createContext()
}

func (suite *streamConsumerGroupTestSuite) TestShardsAssignment() {
	streamPath := fmt.Sprintf("%s/mystream/", suite.testPath)
	var dataPlaneInput v3io.DataPlaneInput
	suite.populateDataPlaneInput(&dataPlaneInput)

	createStreamInput := v3io.CreateStreamInput{
		DataPlaneInput:       dataPlaneInput,
		Path:                 streamPath,
		ShardCount:           8,
		RetentionPeriodHours: 1,
	}

	err := suite.container.CreateStreamSync(&createStreamInput)
	suite.Require().NoError(err, "Failed to create stream")

	streamConsumerGroup, err := streamconsumergroup.NewStreamConsumerGroup(
		"1",
		suite.logger,
		nil,
		dataPlaneInput,
		streamPath,
		1,
		suite.container)
	suite.Require().NoError(err, "Failed creating stream consumer group")

	memberID := "member1"
	streamConsumerGroupHandler, err := newStreamConsumerGroupHandler(suite, memberID)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(memberID, streamConsumerGroupHandler)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	time.Sleep(500 * time.Second)

	streamConsumerGroup.Close()
}

type streamConsumerGroupHandler struct {
	suite    *streamConsumerGroupTestSuite
	logger   logger.Logger
	memberID string
}

func newStreamConsumerGroupHandler(suite *streamConsumerGroupTestSuite, memberID string) (v3io.StreamConsumerGroupHandler, error) {
	return &streamConsumerGroupHandler{
		suite:    suite,
		logger:   suite.logger.GetChild(fmt.Sprintf("streamConsumerGroupHandler-%s", memberID)),
		memberID: memberID,
	}, nil
}

func (h *streamConsumerGroupHandler) Setup(session v3io.StreamConsumerGroupSession) error {
	assignedShardIDs, err := session.Claims()
	h.suite.Require().NoError(err, "Failed getting assigned claims")
	h.logger.DebugWith("Setup called", "assignedShardIDs", assignedShardIDs)
	return nil
}

func (h *streamConsumerGroupHandler) Cleanup(session v3io.StreamConsumerGroupSession) error {
	h.logger.DebugWith("Cleanup called")
	return nil
}

func (h *streamConsumerGroupHandler) ConsumeClaim(session v3io.StreamConsumerGroupSession, claim v3io.StreamConsumerGroupClaim) error {
	h.logger.DebugWith("Consume Claims called")
	return nil
}

func TestStreamConsumerGroupTestSuite(t *testing.T) {
	suite.Run(t, new(streamConsumerGroupTestSuite))
}

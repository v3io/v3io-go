package test

import (
	"fmt"
	"testing"

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
		8,
		suite.container)
	suite.Require().NoError(err, "Failed creating stream consumer group")

	memberID1 := "member1"
	streamConsumerGroupHandler1, err := newStreamConsumerGroupHandler(suite, memberID1)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(memberID1, streamConsumerGroupHandler1)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	memberID2 := "member2"
	streamConsumerGroupHandler2, err := newStreamConsumerGroupHandler(suite, memberID2)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(memberID2, streamConsumerGroupHandler2)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	memberID3 := "member3"
	streamConsumerGroupHandler3, err := newStreamConsumerGroupHandler(suite, memberID3)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(memberID3, streamConsumerGroupHandler3)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	// Put some records
	firstShardID := 1
	secondShardID := 2

	records := []*v3io.StreamRecord{
		{ShardID: &firstShardID, Data: []byte("first shard record #1")},
		{ShardID: &firstShardID, Data: []byte("first shard record #2")},
		{ShardID: &secondShardID, Data: []byte("second shard record #1")},
		{Data: []byte("some shard (will have ID=0) record #1")},
	}

	putRecordsInput := v3io.PutRecordsInput{
		Path:    streamPath,
		Records: records,
	}

	suite.populateDataPlaneInput(&putRecordsInput.DataPlaneInput)

	response, err := suite.container.PutRecordsSync(&putRecordsInput)
	suite.Require().NoError(err, "Failed to put records")

	putRecordsResponse := response.Output.(*v3io.PutRecordsOutput)
	suite.Require().Equal(0, putRecordsResponse.FailedRecordCount)

	//time.Sleep(10 * time.Second)

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

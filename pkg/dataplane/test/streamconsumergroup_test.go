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

	createStreamInput := v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           8,
		RetentionPeriodHours: 1,
	}

	err := suite.container.CreateStreamSync(&createStreamInput)
	suite.Require().NoError(err, "Failed to create stream")

	streamConsumerGroup, err := streamconsumergroup.NewStreamConsumerGroup(
		"1",
		"member1",
		suite.logger,
		nil,
		streamPath,
		8,
		suite.container)
	suite.Require().NoError(err, "Failed creating stream consumer group")

	memberID1 := "member1"
	streamConsumerGroupHandler1, err := newStreamConsumerGroupHandler(suite, memberID1)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(streamConsumerGroupHandler1)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	memberID2 := "member2"
	streamConsumerGroupHandler2, err := newStreamConsumerGroupHandler(suite, memberID2)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(streamConsumerGroupHandler2)
	suite.Require().NoError(err, "Failed consuming stream consumer group")

	memberID3 := "member3"
	streamConsumerGroupHandler3, err := newStreamConsumerGroupHandler(suite, memberID3)
	suite.Require().NoError(err, "Failed creating stream consumer group handler")

	err = streamConsumerGroup.Consume(streamConsumerGroupHandler3)
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

func newStreamConsumerGroupHandler(suite *streamConsumerGroupTestSuite, memberID string) (streamconsumergroup.Handler, error) {
	return &streamConsumerGroupHandler{
		suite:    suite,
		logger:   suite.logger.GetChild(fmt.Sprintf("streamConsumerGroupHandler-%s", memberID)),
		memberID: memberID,
	}, nil
}

func (h *streamConsumerGroupHandler) Setup(session streamconsumergroup.Session) error {
	h.logger.DebugWith("Setup called", "assignedShardIDs", session.GetClaims())
	return nil
}

func (h *streamConsumerGroupHandler) Cleanup(session streamconsumergroup.Session) error {
	h.logger.DebugWith("Cleanup called")
	return nil
}

func (h *streamConsumerGroupHandler) ConsumeClaim(session streamconsumergroup.Session, claim streamconsumergroup.Claim) error {
	h.logger.DebugWith("Consume Claims called")
	return nil
}

func TestStreamConsumerGroupTestSuite(t *testing.T) {
	suite.Run(t, new(streamConsumerGroupTestSuite))
}

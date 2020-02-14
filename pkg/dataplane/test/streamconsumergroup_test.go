package test

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/streamconsumergroup"

	"github.com/nuclio/logger"
	"github.com/stretchr/testify/suite"
)

type recordData struct {
	ShardID int `json:"shard_id"`
	Index   int `json:"index"`
}

type streamConsumerGroupTestSuite struct {
	StreamTestSuite
	streamPath string
}

func (suite *streamConsumerGroupTestSuite) SetupSuite() {
	suite.StreamTestSuite.SetupSuite()
	suite.createContainer()
	suite.streamPath = fmt.Sprintf("%s/test-stream-0/", suite.testPath)
}

func (suite *streamConsumerGroupTestSuite) TestLocationHandling() {
	consumerGroupName := "cg0"
	numShards := 8

	suite.createStream(suite.streamPath, numShards)

	memberGroup := newMemberGroup(suite,
		consumerGroupName,
		2,
		numShards,
		2,
		[]int{0, 0, 0, 0, 0, 0, 0, 0},
		[]int{5, 10, 10, 10, 15, 10, 10, 20})

	// wait a bit for things to happen - the members should all connect, get their partitions and start consuming
	// but not actually consume anything
	time.Sleep(3 * time.Second)

	// must have exactly 2 shards each, must all be consuming, must all have not processed any messages
	memberGroup.verifyClaimShards(numShards, []int{4})
	memberGroup.verifyNumActiveClaimConsumptions(numShards)
	memberGroup.verifyNumRecordsConsumed([]int{0, 0, 0, 0, 0, 0, 0, 0})

	suite.writeRecords([]int{30, 30, 30, 30, 30, 30, 30, 30})

	// wait a bit for things to happen - the members should read data from the shards up to the amount they were
	// told to read, verifying that each message is in order and the expected
	time.Sleep(15 * time.Second)

	memberGroup.verifyClaimShards(numShards, []int{4})
	memberGroup.verifyNumActiveClaimConsumptions(0)
	memberGroup.verifyNumRecordsConsumed([]int{5, 10, 10, 10, 15, 10, 10, 20})

	// stop the group
	memberGroup.stop()
	time.Sleep(3 * time.Second)

	memberGroup = newMemberGroup(suite,
		consumerGroupName,
		4,
		numShards,
		4,
		[]int{5, 10, 10, 10, 15, 10, 10, 20},
		[]int{50, 50, 50, 50, 50, 50, 50, 50})

	// wait a bit for things to happen
	time.Sleep(30 * time.Second)

	memberGroup.verifyClaimShards(numShards, []int{2})
	memberGroup.verifyNumActiveClaimConsumptions(8)
	memberGroup.verifyNumRecordsConsumed([]int{25, 20, 20, 20, 15, 20, 20, 10})

	memberGroup.stop()
	time.Sleep(3 * time.Second)
}

func (suite *streamConsumerGroupTestSuite) createStream(streamPath string, numShards int) {
	createStreamInput := v3io.CreateStreamInput{
		Path:                 streamPath,
		ShardCount:           numShards,
		RetentionPeriodHours: 1,
	}

	err := suite.container.CreateStreamSync(&createStreamInput)
	suite.Require().NoError(err, "Failed to create stream")
}

func (suite *streamConsumerGroupTestSuite) writeRecords(numRecordsPerShard []int) {
	var records []*v3io.StreamRecord

	suite.logger.DebugWith("Writing records", "numRecordsPerShard", numRecordsPerShard)

	for shardID, numRecordsPerShard := range numRecordsPerShard {

		// we're taking address
		shardIDCopy := shardID

		for recordIndex := 0; recordIndex < numRecordsPerShard; recordIndex++ {
			recordDataInstance := recordData{
				ShardID: shardIDCopy,
				Index:   recordIndex,
			}

			marshalledRecordDataInstance, err := json.Marshal(&recordDataInstance)
			suite.Require().NoError(err)

			records = append(records, &v3io.StreamRecord{
				ShardID: &shardIDCopy,
				Data:    marshalledRecordDataInstance,
			})
		}
	}

	putRecordsInput := v3io.PutRecordsInput{
		Path:    suite.streamPath,
		Records: records,
	}

	response, err := suite.container.PutRecordsSync(&putRecordsInput)
	suite.Require().NoError(err, "Failed to put records")

	putRecordsResponse := response.Output.(*v3io.PutRecordsOutput)
	suite.Require().Equal(0, putRecordsResponse.FailedRecordCount)

	suite.logger.DebugWith("Done writing records", "numRecordsPerShard", numRecordsPerShard)
}

//
// Orchestrates a group of members
//

type memberGroup struct {
	suite                   *streamConsumerGroupTestSuite
	members                 []*member
	numberOfRecordsConsumed []int
}

func newMemberGroup(suite *streamConsumerGroupTestSuite,
	consumerGroupName string,
	maxNumMembers int,
	numShards int,
	numMembers int,
	expectedInitialRecordIndex []int,
	numberOfRecordToConsume []int) *memberGroup {
	newMemberGroup := memberGroup{
		suite:                   suite,
		numberOfRecordsConsumed: make([]int, numShards),
	}

	memberChan := make(chan *member, numMembers)

	for memberIdx := 0; memberIdx < numMembers; memberIdx++ {
		go func() {
			memberInstance := newMember(suite,
				consumerGroupName,
				maxNumMembers,
				numShards,
				memberIdx,
				newMemberGroup.numberOfRecordsConsumed)

			// start
			memberInstance.start(expectedInitialRecordIndex, numberOfRecordToConsume)

			// shove to member chan
			memberChan <- memberInstance
		}()
	}

	for memberInstance := range memberChan {
		newMemberGroup.members = append(newMemberGroup.members, memberInstance)
		if len(newMemberGroup.members) >= numMembers {
			break
		}
	}

	return &newMemberGroup
}

func (mg *memberGroup) verifyClaimShards(expectedTotalNumShards int, expectedNumShardsPerMember []int) {
	totalNumShards := 0

	for _, member := range mg.members {
		numMemberShards := len(member.claims)

		mg.suite.Require().Contains(expectedNumShardsPerMember,
			numMemberShards,
			"Member %s doesn't have the required amount of shards. Has %d, expected %v",
			member.id,
			numMemberShards,
			expectedNumShardsPerMember)

		totalNumShards += numMemberShards
	}

	mg.suite.Require().Equal(expectedTotalNumShards, totalNumShards)
}

func (mg *memberGroup) verifyNumActiveClaimConsumptions(expectedNumActiveClaimConsumptions int) {
	totalNumActiveClaimConsumptions := 0

	for _, member := range mg.members {
		totalNumActiveClaimConsumptions += int(member.numActiveClaimConsumptions)
	}

	mg.suite.Require().Equal(expectedNumActiveClaimConsumptions, totalNumActiveClaimConsumptions)
}

func (mg *memberGroup) verifyNumRecordsConsumed(expectedNumRecordsConsumed []int) {
	mg.suite.Require().Equal(expectedNumRecordsConsumed, mg.numberOfRecordsConsumed)
}

func (mg *memberGroup) stop() {
	for _, member := range mg.members {
		member.stop()
	}

	mg.suite.logger.Info("Member group stopped")
}

//
// Simulates a member
//

type member struct {
	suite                      *streamConsumerGroupTestSuite
	logger                     logger.Logger
	id                         string
	expectedStartRecordIndex   []int
	numberOfRecordToConsume    []int
	numberOfRecordsConsumed    []int
	streamConsumerGroup        streamconsumergroup.StreamConsumerGroup
	claims                     []streamconsumergroup.Claim
	numActiveClaimConsumptions int64
}

func newMember(suite *streamConsumerGroupTestSuite,
	consumerGroupName string,
	maxNumMembers int,
	numShards int,
	index int,
	numberOfRecordsConsumed []int) *member {
	id := fmt.Sprintf("m%d", index)

	streamConsumerGroupConfig := streamconsumergroup.NewConfig()
	streamConsumerGroupConfig.Claim.RecordBatchFetch.NumRecordsInBatch = 1
	streamConsumerGroupConfig.Claim.RecordBatchFetch.Interval = 50 * time.Millisecond

	streamConsumerGroup, err := streamconsumergroup.NewStreamConsumerGroup(
		consumerGroupName,
		id,
		suite.logger,
		streamConsumerGroupConfig,
		suite.streamPath,
		maxNumMembers,
		suite.container)
	suite.Require().NoError(err, "Failed creating stream consumer group")

	return &member{
		suite:                    suite,
		logger:                   suite.logger.GetChild(id),
		id:                       id,
		streamConsumerGroup:      streamConsumerGroup,
		expectedStartRecordIndex: make([]int, numShards),
		numberOfRecordToConsume:  make([]int, numShards),
		numberOfRecordsConsumed:  numberOfRecordsConsumed,
	}
}

func (m *member) Setup(session streamconsumergroup.Session) error {
	m.claims = session.GetClaims()

	shardIDs := m.getShardIDs()
	m.logger.DebugWith("Setup called", "shardIDs", shardIDs)

	return nil
}

func (m *member) Cleanup(session streamconsumergroup.Session) error {
	m.logger.DebugWith("Cleanup called")
	return nil
}

func (m *member) ConsumeClaim(session streamconsumergroup.Session, claim streamconsumergroup.Claim) error {
	numActiveClaimConsumptions := atomic.AddInt64(&m.numActiveClaimConsumptions, 1)
	m.logger.DebugWith("Consume Claims called", "numActiveClaimConsumptions", numActiveClaimConsumptions)

	expectedRecordIndex := m.expectedStartRecordIndex[claim.GetShardID()]

	// reduce at the end
	defer func() {
		numActiveClaimConsumptions := atomic.AddInt64(&m.numActiveClaimConsumptions, -1)

		m.logger.DebugWith("Consume Claims done",
			"numRecordsConsumed", m.numberOfRecordsConsumed,
			"numActiveClaimConsumptions", numActiveClaimConsumptions)
	}()

	// start reading
	for recordBatch := range claim.GetRecordBatchChan() {

		// iterate over records
		for _, record := range recordBatch.Records {
			recordDataInstance := recordData{}

			// read the data into message
			err := json.Unmarshal(record.Data, &recordDataInstance)
			m.suite.Require().NoError(err)

			// make sure we're reading the proper shard
			m.suite.Require().Equal(recordDataInstance.ShardID, claim.GetShardID())

			// check we got the expected message index
			m.suite.Require().Equal(expectedRecordIndex, recordDataInstance.Index)

			expectedRecordIndex++
			m.numberOfRecordsConsumed[claim.GetShardID()]++

			if m.numberOfRecordsConsumed[claim.GetShardID()] >= m.numberOfRecordToConsume[claim.GetShardID()] {
				err := session.MarkRecordBatch(recordBatch)
				m.suite.Require().NoError(err)

				return nil
			}
		}

		err := session.MarkRecordBatch(recordBatch)
		m.suite.Require().NoError(err)
	}

	return nil
}

func (m *member) start(expectedStartRecordIndex []int, numberOfRecordToConsume []int) {
	m.expectedStartRecordIndex = expectedStartRecordIndex
	m.numberOfRecordToConsume = numberOfRecordToConsume

	// start consuming
	err := m.streamConsumerGroup.Consume(m)
	m.suite.Require().NoError(err)
}

func (m *member) stop() {
	err := m.streamConsumerGroup.Close()
	m.suite.Require().NoError(err)
}

func (m *member) getShardIDs() []int {
	var shardIDs []int

	for _, claim := range m.claims {
		shardIDs = append(shardIDs, claim.GetShardID())
	}

	return shardIDs
}

func TestStreamConsumerGroupTestSuite(t *testing.T) {
	suite.Run(t, new(streamConsumerGroupTestSuite))
}

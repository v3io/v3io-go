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
	"encoding/json"
	"fmt"
	"path"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/dataplane/streamconsumergroup"

	"github.com/nuclio/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type recordData struct {
	ShardID int `json:"shard_id"`
	Index   int `json:"index"`
}

type streamConsumerGroupTestSuite struct {
	streamTestSuite
	streamPath string
}

func (suite *streamConsumerGroupTestSuite) SetupSuite() {
	suite.streamTestSuite.SetupSuite()
	suite.createContainer()
}

func (suite *streamConsumerGroupTestSuite) SetupTest() {
	suite.streamTestSuite.SetupTest()
	suite.streamPath = fmt.Sprintf("%s/test-stream-0/", suite.testPath)
}

func (suite *streamConsumerGroupTestSuite) TestLocationHandling() {
	numShards := 8

	suite.createStream(suite.streamPath, numShards)

	streamConsumerGroup := suite.createStreamConsumerGroup(2)

	memberGroup := newMemberGroup(suite,
		streamConsumerGroup,
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

	// get the num shards from the observer
	observedNumShards, err := streamConsumerGroup.GetNumShards()
	suite.Require().NoError(err)
	suite.Require().Equal(numShards, observedNumShards)

	// get the num shards from the observer
	observedState, err := streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, 2)

	// iterate over shards to check their sequence numbers, they shouldn't exist
	for shardID := 0; shardID < numShards; shardID++ {
		shardSequenceNumber, err := streamConsumerGroup.GetShardSequenceNumber(shardID)
		suite.Require().Equal(err, streamconsumergroup.ErrShardNotFound)
		suite.Require().Equal(uint64(0), shardSequenceNumber)
	}

	suite.writeRecords([]int{30, 30, 30, 30, 30, 30, 30, 30})

	// wait a bit for things to happen - the members should read data from the shards up to the amount they were
	// told to read, verifying that each message is in order and the expected
	time.Sleep(15 * time.Second)

	memberGroup.verifyClaimShards(numShards, []int{4})
	memberGroup.verifyNumActiveClaimConsumptions(0)
	memberGroup.verifyNumRecordsConsumed([]int{5, 10, 10, 10, 15, 10, 10, 20})
	suite.verifyShardSequenceNumbers(numShards, streamConsumerGroup, []int{5, 10, 10, 10, 15, 10, 10, 20})

	// stop the group
	memberGroup.stop()
	time.Sleep(3 * time.Second)

	streamConsumerGroup = suite.createStreamConsumerGroup(4)

	memberGroup = newMemberGroup(suite,
		streamConsumerGroup,
		numShards,
		4,
		[]int{5, 10, 10, 10, 15, 10, 10, 20},
		[]int{50, 50, 50, 50, 50, 50, 50, 50})

	// wait a bit for things to happen
	time.Sleep(30 * time.Second)

	// get the num shards from the observer
	observedNumShards, err = streamConsumerGroup.GetNumShards()
	suite.Require().NoError(err)
	suite.Require().Equal(numShards, observedNumShards)

	// get the num shards from the observer
	observedState, err = streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, 4)

	// iterate over shards to check their sequence numbers
	suite.verifyShardSequenceNumbers(numShards, streamConsumerGroup, []int{30, 30, 30, 30, 30, 30, 30, 30})

	memberGroup.verifyClaimShards(numShards, []int{2})
	memberGroup.verifyNumActiveClaimConsumptions(8)
	memberGroup.verifyNumRecordsConsumed([]int{25, 20, 20, 20, 15, 20, 20, 10})

	memberGroup.stop()
	time.Sleep(3 * time.Second)
}

func (suite *streamConsumerGroupTestSuite) TestStateHandlerRetainShards() {
	numShards := 64
	members := 32
	expectedInitialRecordIndex := make([]int, numShards)
	numberOfRecordToConsume := make([]int, numShards)
	for i := 0; i < len(numberOfRecordToConsume); i++ {
		numberOfRecordToConsume[i] = 1
	}
	suite.createStream(suite.streamPath, numShards)
	suite.writeRecords(numberOfRecordToConsume)
	streamConsumerGroup := suite.createStreamConsumerGroup(members)
	memberGroup := newMemberGroup(suite,
		streamConsumerGroup,
		numShards,
		members,
		expectedInitialRecordIndex,
		numberOfRecordToConsume,
	)

	// wait a bit for things to happen - the members should all connect, get their partitions and start consuming
	// but not actually consume anything
	time.Sleep(25 * time.Second)

	memberGroup.verifyClaimShards(numShards, []int{numShards / members})
	memberGroup.verifyNumRecordsConsumed(numberOfRecordToConsume)

	// get the num shards from the observer
	observedNumShards, err := streamConsumerGroup.GetNumShards()
	suite.Require().NoError(err)
	suite.Require().Equal(numShards, observedNumShards)

	// get the state from the observer
	observedState, err := streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, members)

	stopMembersErrGroup, _ := errgroup.WithContext(context.TODO())

	originalShardGroups := map[string][]int{}

	// stop
	for i := 0; i < members/2; i++ {
		member := memberGroup.members[i]
		suite.logger.DebugWith("Stopping member", "memberID", member.streamConsumerGroupMember.GetID())

		// copy shardsToRetain list for every member
		if member.streamConsumerGroupMember.GetRetainShardFlag() {
			suite.logger.DebugWith("member needs to retain flag, saving shardsToRetain",
				"memberID", member.streamConsumerGroupMember.GetID(),
				"shardsToRetain", member.streamConsumerGroupMember.GetShardsToRetain())
			originalShardGroups[member.streamConsumerGroupMember.GetID()] = make([]int, len(member.streamConsumerGroupMember.GetShardsToRetain()))
			copy(originalShardGroups[member.streamConsumerGroupMember.GetID()], member.streamConsumerGroupMember.GetShardsToRetain())
		}

		stopMembersErrGroup.Go(func() error {

			// stall member from functioning
			suite.logger.DebugWith("stopping member", "memberID", member.id)
			member.stop()
			suite.Require().NoError(err)

			duration := 1 * time.Second
			return common.RetryFunc(context.TODO(),
				suite.logger,
				30,
				&duration,
				nil,
				func(attempt int) (bool, error) {
					observedState, err := streamConsumerGroup.GetState()
					suite.Require().NoError(err)
					for _, sessionState := range observedState.SessionStates {
						if sessionState.MemberID == member.streamConsumerGroupMember.GetID() {
							suite.logger.DebugWith("Session state was not removed just yet")
							return true, nil
						}
					}

					suite.logger.DebugWith("Session state was removed",
						"observedState", observedState,
						"memberID", member.id)
					return false, nil
				})
		})

	}

	suite.Require().NoError(stopMembersErrGroup.Wait())

	// get the state from the observer again
	observedState, err = streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, members/2)

	suite.logger.DebugWith("Observed state after stopping members", "observedState", observedState)

	// check retain shard flag
	for _, member := range memberGroup.members {
		suite.logger.DebugWith("Checking retainShard flags for member",
			"memberID", member.streamConsumerGroupMember.GetID(),
			"retainShardFlag", member.streamConsumerGroupMember.GetRetainShardFlag())
	}

	// start
	startMembersErrGroup, _ := errgroup.WithContext(context.TODO())
	for i := 0; i < members/2; i++ {
		member := memberGroup.members[i]
		suite.logger.DebugWith("Starting member", "memberID", member.streamConsumerGroupMember.GetID())
		startMembersErrGroup.Go(func() error {
			if err := member.streamConsumerGroupMember.Start(); err != nil {
				return err
			}

			// blocking
			go func() {
				suite.Require().NoError(member.streamConsumerGroupMember.Consume(member))
			}()

			duration := 1 * time.Second
			return common.RetryFunc(context.TODO(),
				suite.logger,
				30,
				&duration,
				nil,
				func(attempt int) (bool, error) {
					observedState, err := streamConsumerGroup.GetState()
					suite.Require().NoError(err)

					for _, sessionState := range observedState.SessionStates {
						if sessionState.MemberID == member.streamConsumerGroupMember.GetID() {
							suite.logger.DebugWith("retained shards",
								"shards", sessionState.Shards,
								"memberID", sessionState.MemberID)
							return false, nil
						}
					}

					suite.logger.DebugWith("Session state shards were no retained just yet",
						"sessionStates", observedState.SessionStates,
						"memberID", member.streamConsumerGroupMember.GetID())
					return true, nil

				})
		})
	}
	suite.Require().NoError(startMembersErrGroup.Wait())

	// get the state from the observer again
	observedState, err = streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, members)

	suite.logger.DebugWith("Observed state after starting members", "observedState", observedState)

	// check retain shard flag
	for _, member := range memberGroup.members {
		suite.logger.DebugWith("Checking retainShard flags for member",
			"memberID", member.streamConsumerGroupMember.GetID(),
			"retainShardFlag", member.streamConsumerGroupMember.GetRetainShardFlag())
	}

	// check if each member's shards are different from the original shards
	for i := 0; i < members/2; i++ {
		member := memberGroup.members[i]
		memberID := member.streamConsumerGroupMember.GetID()
		suite.logger.DebugWith("Shard comparison",
			"memberID", memberID,
			"memberShards", member.streamConsumerGroupMember.GetShardsToRetain(),
			"originalShards", originalShardGroups[memberID])
		suite.Require().ElementsMatch(originalShardGroups[memberID], member.streamConsumerGroupMember.GetShardsToRetain())
	}
}

func (suite *streamConsumerGroupTestSuite) TestStateHandlerAbort() {
	numShards := 32
	members := 4
	expectedInitialRecordIndex := make([]int, numShards)
	numberOfRecordToConsume := make([]int, numShards)
	for i := 0; i < len(numberOfRecordToConsume); i++ {
		numberOfRecordToConsume[i] = 1
	}
	suite.createStream(suite.streamPath, numShards)
	suite.writeRecords(numberOfRecordToConsume)
	streamConsumerGroup := suite.createStreamConsumerGroup(members)
	memberGroup := newMemberGroup(suite,
		streamConsumerGroup,
		numShards,
		members,
		expectedInitialRecordIndex,
		numberOfRecordToConsume,
	)

	// wait a bit for things to happen - the members should all connect, get their partitions and start consuming
	// but not actually consume anything
	time.Sleep(10 * time.Second)

	memberGroup.verifyClaimShards(numShards, []int{numShards / members})
	memberGroup.verifyNumRecordsConsumed(numberOfRecordToConsume)

	// get the num shards from the observer
	observedNumShards, err := streamConsumerGroup.GetNumShards()
	suite.Require().NoError(err)
	suite.Require().Equal(numShards, observedNumShards)

	// get the state from the observer
	observedState, err := streamConsumerGroup.GetState()
	suite.Require().NoError(err)
	suite.Require().Len(observedState.SessionStates, members)

	for _, member := range memberGroup.members {
		member.On("Abort", mock.Anything).Return(nil)
	}

	// read state file from persistency
	var state *streamconsumergroup.State
	consumerGroupName := "cg0"
	duration := 1 * time.Second
	err = common.RetryFunc(context.TODO(),
		suite.logger,
		10,
		&duration,
		nil,
		func(attempt int) (bool, error) {
			state, err = suite.getStateFromPersistency(suite.streamPath, consumerGroupName)
			if err != nil {
				suite.logger.DebugWith("State was not retrieved from persistency",
					"err", err)
				return true, err
			}
			return false, nil
		})
	suite.Require().NoError(err)

	suite.logger.DebugWith("Got state from persistency", "state", state)

	// change name of one session
	for i, sessionState := range state.SessionStates {
		sessionState.MemberID = fmt.Sprintf("some-member-%v", i)
	}

	err = common.RetryFunc(context.TODO(),
		suite.logger,
		10,
		&duration,
		nil,
		func(attempt int) (bool, error) {
			err = suite.setStateInPersistency(suite.streamPath, consumerGroupName, state)
			if err != nil {
				suite.logger.DebugWith("State was not set in persistency yet",
					"err", err)
				return true, err
			}
			return false, nil
		})
	suite.Require().NoError(err)

	// make sure abort was called for all members
	time.Sleep(5 * time.Second)
	for _, member := range memberGroup.members {
		member.AssertCalled(suite.T(), "Abort", mock.Anything)
	}

}

func (suite *streamConsumerGroupTestSuite) createStreamConsumerGroup(maxReplicas int) streamconsumergroup.StreamConsumerGroup {
	consumerGroupName := "cg0"

	streamConsumerGroupConfig := streamconsumergroup.NewConfig()
	streamConsumerGroupConfig.Claim.RecordBatchFetch.NumRecordsInBatch = 10
	streamConsumerGroupConfig.Claim.RecordBatchFetch.Interval = 50 * time.Millisecond
	streamConsumerGroupConfig.Session.Timeout = 10 * time.Second
	streamConsumerGroupConfig.Session.HeartbeatInterval = 50 * time.Millisecond

	streamConsumerGroup, err := streamconsumergroup.NewStreamConsumerGroup(suite.logger,
		consumerGroupName,
		streamConsumerGroupConfig,
		suite.container,
		suite.streamPath,
		maxReplicas)
	suite.Require().NoError(err, "Failed creating stream consumer group")

	return streamConsumerGroup
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

func (suite *streamConsumerGroupTestSuite) verifyShardSequenceNumbers(numShards int,
	streamConsumerGroup streamconsumergroup.StreamConsumerGroup,
	expectedShardSequenceNumbers []int) {

	// iterate over shards to check their sequence numbers
	for shardID := 0; shardID < numShards; shardID++ {
		shardSequenceNumber, err := streamConsumerGroup.GetShardSequenceNumber(shardID)
		suite.Require().NoError(err)
		suite.Require().Equal(uint64(expectedShardSequenceNumbers[shardID]), shardSequenceNumber)
	}
}

func (suite *streamConsumerGroupTestSuite) getStateFromPersistency(streamPath, consumerGroupName string) (*streamconsumergroup.State, error) {
	stateContentsAttributeKey := "state"

	response, err := suite.container.GetItemSync(&v3io.GetItemInput{
		Path: path.Join(streamPath, fmt.Sprintf("%s-state.json", consumerGroupName)),
		AttributeNames: []string{
			stateContentsAttributeKey,
		},
	})

	if err != nil {
		return nil, err
	}

	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	stateContents, err := getItemOutput.Item.GetFieldString(stateContentsAttributeKey)
	if err != nil {
		return nil, err
	}

	var state streamconsumergroup.State

	if err := json.Unmarshal([]byte(stateContents), &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (suite *testSuite) setStateInPersistency(streamPath, consumerGroupName string, state *streamconsumergroup.State) error {
	stateContents, err := json.Marshal(state)
	if err != nil {
		return err
	}

	if _, err := suite.container.UpdateItemSync(&v3io.UpdateItemInput{
		Path: path.Join(streamPath, fmt.Sprintf("%s-state.json", consumerGroupName)),
		Attributes: map[string]interface{}{
			"state": string(stateContents),
		},
	}); err != nil {
		return err
	}

	return nil
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
	streamConsumerGroup streamconsumergroup.StreamConsumerGroup,
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
		memberIdx := memberIdx
		go func() {
			memberInstance := newMember(suite,
				streamConsumerGroup,
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
	mock.Mock
	suite                      *streamConsumerGroupTestSuite
	logger                     logger.Logger
	streamConsumerGroupMember  streamconsumergroup.Member
	id                         string
	expectedStartRecordIndex   []int
	numberOfRecordToConsume    []int
	numberOfRecordsConsumed    []int
	streamConsumerGroup        streamconsumergroup.StreamConsumerGroup
	claims                     []streamconsumergroup.Claim
	numActiveClaimConsumptions int64
}

func newMember(suite *streamConsumerGroupTestSuite,
	streamConsumerGroup streamconsumergroup.StreamConsumerGroup,
	numShards int,
	index int,
	numberOfRecordsConsumed []int) *member {
	id := fmt.Sprintf("m%d", index)

	streamConsumerGroupMember, err := streamconsumergroup.NewMember(streamConsumerGroup, id)
	suite.Require().NoError(err)

	return &member{
		suite:                     suite,
		logger:                    suite.logger.GetChild(id),
		streamConsumerGroupMember: streamConsumerGroupMember,
		id:                        id,
		streamConsumerGroup:       streamConsumerGroup,
		expectedStartRecordIndex:  make([]int, numShards),
		numberOfRecordToConsume:   make([]int, numShards),
		numberOfRecordsConsumed:   numberOfRecordsConsumed,
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

			err = session.MarkRecord(&record)
			m.suite.Require().NoError(err)

			if m.numberOfRecordsConsumed[claim.GetShardID()] >= m.numberOfRecordToConsume[claim.GetShardID()] {
				return nil
			}
		}
	}

	return nil
}

func (m *member) Abort(session streamconsumergroup.Session) error {
	m.logger.DebugWith("Abort called", "memberID", m.streamConsumerGroupMember.GetID())
	m.Called(session)
	m.stop()
	return nil
}

func (m *member) start(expectedStartRecordIndex []int, numberOfRecordToConsume []int) {
	m.expectedStartRecordIndex = expectedStartRecordIndex
	m.numberOfRecordToConsume = numberOfRecordToConsume

	// start consuming
	err := m.streamConsumerGroupMember.Consume(m)
	m.suite.Require().NoError(err)
}

func (m *member) stop() {
	err := m.streamConsumerGroupMember.Close()
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

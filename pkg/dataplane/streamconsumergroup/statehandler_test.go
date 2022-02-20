package streamconsumergroup

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type stateHandlerSuite struct {
	suite.Suite
	stateHandler *stateHandler
}

func (suite *stateHandlerSuite) TestAssignShards() {

	for _, testCase := range []struct {
		name                string
		maxReplicas         int
		numShards           int
		existingShardGroups [][]int
		expectedShardGroup  []int
	}{
		{
			name:                "even, more shards than replicas",
			maxReplicas:         4,
			numShards:           8,
			existingShardGroups: [][]int{{0, 1}, {4, 5}},
			expectedShardGroup:  []int{2, 3},
		},
		{
			name:                "odd, more shards than replicas",
			maxReplicas:         3,
			numShards:           8,
			existingShardGroups: [][]int{{0, 1, 2}},
			expectedShardGroup:  []int{3, 4},
		},
		{
			name:                "equal number of shards and replicas",
			maxReplicas:         4,
			numShards:           4,
			existingShardGroups: [][]int{{0}, {1}, {3}},
			expectedShardGroup:  []int{2},
		},
		{
			name:                "more replicas than shards, no empty groups assigned",
			maxReplicas:         4,
			numShards:           2,
			existingShardGroups: [][]int{{0}, {1}},
			expectedShardGroup:  []int{},
		},
		{
			name:                "more replicas than shards, all empty groups assigned",
			maxReplicas:         4,
			numShards:           2,
			existingShardGroups: [][]int{{}, {}},
			expectedShardGroup:  []int{0},
		},
		{
			name:                "more replicas than shards, some empty groups assigned",
			maxReplicas:         4,
			numShards:           2,
			existingShardGroups: [][]int{{}, {0}},
			expectedShardGroup:  []int{},
		},
	} {
		// make state from shard groups
		state := State{}
		for _, existingShardGroup := range testCase.existingShardGroups {
			state.SessionStates = append(state.SessionStates, &SessionState{
				Shards: existingShardGroup,
			})
		}

		assignedShardGroup, err := suite.stateHandler.assignShards(testCase.maxReplicas, testCase.numShards, &state)
		suite.Require().NoError(err)
		suite.Require().Equal(testCase.expectedShardGroup, assignedShardGroup, testCase.name)
	}
}

func (suite *stateHandlerSuite) TestRetainShards() {
	for _, testCase := range []struct {
		name                string
		memberID            string
		existingShardGroups [][]int
		expectedShardGroup  []int
		expectedError       bool
	}{
		{
			name:                "successfulRetention",
			memberID:            "1",
			existingShardGroups: [][]int{{0, 1}, {2, 3}},
			expectedShardGroup:  []int{2, 3},
			expectedError:       false,
		},
		{
			name:                "failedRetention",
			memberID:            "2",
			existingShardGroups: [][]int{{0, 1}, {2, 3}, {4, 5}},
			expectedShardGroup:  []int{0, 1},
			expectedError:       true,
		},
		{
			name:                "unexpectedBehaviour",
			memberID:            "0",
			existingShardGroups: [][]int{{0, 1}, {2, 3}},
			expectedShardGroup:  []int{4, 5},
			expectedError:       true,
		},
	} {
		suite.Run(testCase.name, func() {

			// make state from shard groups
			state := State{}
			for i, existingShardGroup := range testCase.existingShardGroups {
				state.SessionStates = append(state.SessionStates, &SessionState{
					Shards:   existingShardGroup,
					MemberID: strconv.Itoa(i),
				})
			}

			shards, err := suite.stateHandler.retainShards(testCase.expectedShardGroup, testCase.memberID, &state)
			if testCase.expectedError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().ElementsMatch(testCase.expectedShardGroup, shards)
			}
		})
	}
}

func TestBinaryTestSuite(t *testing.T) {
	suite.Run(t, new(stateHandlerSuite))
}

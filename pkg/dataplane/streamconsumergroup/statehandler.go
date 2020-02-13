package streamconsumergroup

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

const stateContentsAttributeKey string = "state"

type stateHandler struct {
	logger              logger.Logger
	streamConsumerGroup *streamConsumerGroup
	stopChan            chan struct{}
	getStateChan        chan chan *State
}

func newStateHandler(streamConsumerGroup *streamConsumerGroup) (*stateHandler, error) {
	return &stateHandler{
		logger:              streamConsumerGroup.logger.GetChild("stateHandler"),
		streamConsumerGroup: streamConsumerGroup,
		stopChan:            make(chan struct{}),
		getStateChan:        make(chan chan *State),
	}, nil
}

func (sh *stateHandler) start() error {
	go sh.refreshStatePeriodically()

	return nil
}

func (sh *stateHandler) stop() error {
	sh.stopChan <- struct{}{}
	return nil
}

func (sh *stateHandler) getOrCreateSessionState(memberID string) (*SessionState, error) {

	// create a channel on which we'll request the state
	stateResponseChan := make(chan *State, 1)

	// send the channel to the refreshing goroutine. it'll post the state to this channel
	sh.getStateChan <- stateResponseChan

	// wait on it
	state := <-stateResponseChan

	// get the member's session state
	return sh.getSessionState(state, memberID)
}

func (sh *stateHandler) getSessionState(state *State, memberID string) (*SessionState, error) {
	for _, sessionState := range state.SessionStates {
		if sessionState.MemberID == memberID {
			return sessionState, nil
		}
	}

	return nil, errors.Errorf("Member state not found: %s", memberID)
}

func (sh *stateHandler) refreshStatePeriodically() {
	var err error

	// guaranteed to only be REPLACED by a new instance - not edited. as such, once this is initialized
	// it points to a read only state object
	var lastState *State

	for {
		select {

		// if we're asked to get state, get it
		case stateResponseChan := <-sh.getStateChan:
			if lastState != nil {
				stateResponseChan <- lastState
			} else {
				lastState, err = sh.refreshState()
				if err != nil {
					sh.logger.WarnWith("Failed getting state", "err", errors.GetErrorStackString(err, 10))
				}

				// lastState may be nil
				stateResponseChan <- lastState
			}

		// periodically get the state
		case <-time.After(sh.streamConsumerGroup.config.State.Heartbeat.Interval):
			lastState, err = sh.refreshState()
			if err != nil {
				sh.logger.WarnWith("Failed refreshing state", "err", errors.GetErrorStackString(err, 10))
				continue
			}

		// if we're told to stop, exit the loop
		case <-sh.stopChan:
			sh.logger.Debug("Stopping")
			return

		}
	}
}

func (sh *stateHandler) refreshState() (*State, error) {
	sh.logger.DebugWith("Refreshing state")

	return sh.modifyState(func(state *State) (*State, error) {

		// remove stale sessions from state
		if err := sh.removeStaleSessionStates(state); err != nil {
			return nil, errors.Wrap(err, "Failed to remove stale sessions")
		}

		// find our session by member ID
		sessionState := state.findSessionStateByMemberID(sh.streamConsumerGroup.memberID)

		// session already exists - just set the last heartbeat
		if sessionState != nil {
			sessionState.LastHeartbeat = time.Now()

			// we're done
			return state, nil
		}

		// session doesn't exist - create it
		if err := sh.createSessionState(state); err != nil {
			return nil, errors.Wrap(err, "Failed to create session state")
		}

		return state, nil
	})
}

func (sh *stateHandler) createSessionState(state *State) error {
	if state.SessionStates == nil {
		state.SessionStates = []*SessionState{}
	}

	shards, err := sh.assignShards(state)
	if err != nil {
		return errors.Wrap(err, "Failed resolving shards for session")
	}

	state.SessionStates = append(state.SessionStates, &SessionState{
		MemberID:      sh.streamConsumerGroup.memberID,
		LastHeartbeat: time.Now(),
		Shards:        shards,
	})

	sh.logger.DebugWith("Session state added to state", "shards", shards)

	return nil
}

func (sh *stateHandler) assignShards(state *State) ([]int, error) {
	numberOfShardsPerReplica, err := sh.getNumberOfShardsPerReplica(sh.streamConsumerGroup.totalNumShards, sh.streamConsumerGroup.maxReplicas)
	if err != nil {
		return nil, errors.Wrap(err, "Failed resolving max number of shards per session")
	}

	shardIDs := common.MakeRange(0, sh.streamConsumerGroup.totalNumShards-1)
	shardsToAssign := make([]int, 0)
	for _, shardID := range shardIDs {
		found := false
		for _, session := range state.SessionStates {
			if common.IntSliceContainsInt(session.Shards, shardID) {
				found = true
				break
			}
		}
		if found {

			// sanity - it gets inside when there was unassigned shard but an assigned shard found filling the shards
			// list for the session or reaching end of shards list
			if len(shardsToAssign) > 0 {
				return nil, errors.New("Shards assignment out of order")
			}
			continue
		}
		shardsToAssign = append(shardsToAssign, shardID)
		if len(shardsToAssign) == numberOfShardsPerReplica {
			return shardsToAssign, nil
		}
	}

	if len(shardsToAssign) == 0 {
		sh.logger.DebugWith("All shards assigned")
		// TODO: decide what to do
	}

	return shardsToAssign, nil
}

func (sh *stateHandler) getNumberOfShardsPerReplica(numberOfShards int, maxReplicas int) (int, error) {
	if numberOfShards%maxReplicas != 0 {
		return numberOfShards/maxReplicas + 1, nil
	}
	return numberOfShards / maxReplicas, nil
}

func (sh *stateHandler) modifyState(modifier stateModifier) (*State, error) {
	var modifiedState *State

	sh.logger.DebugWith("Modifying state with retries")

	backoff := sh.streamConsumerGroup.config.State.ModifyRetry.Backoff
	attempts := sh.streamConsumerGroup.config.State.ModifyRetry.Attempts

	err := common.RetryFunc(context.TODO(), sh.logger, attempts, nil, &backoff, func(int) (bool, error) {
		state, mtime, err := sh.getStateFromPersistency()
		if err != nil && err != v3ioerrors.ErrNotFound {
			return true, errors.Wrap(err, "Failed getting current state from persistency")
		}

		if state != nil {
			sh.logger.DebugWith("Got current state, modifying", "state", *state)
		} else {
			state, err = newState()
			if err != nil {
				return true, errors.Wrap(err, "Failed to create state")
			}
		}

		modifiedState, err = modifier(state)
		if err != nil {
			return true, errors.Wrap(err, "Failed modifying state")
		}

		sh.logger.DebugWith("Modified state, saving", "modifiedState", modifiedState)

		err = sh.setStateInPersistency(modifiedState, mtime)
		if err != nil {
			return true, errors.Wrap(err, "Failed setting state in persistency state")
		}

		return false, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "Failed modifying state, attempts exhausted")
	}
	sh.logger.DebugWith("State modified successfully")
	return modifiedState, nil
}

func (sh *stateHandler) getStateFilePath() (string, error) {
	return path.Join(sh.streamConsumerGroup.streamPath, fmt.Sprintf("%s-state.json", sh.streamConsumerGroup.name)), nil
}

func (sh *stateHandler) setStateInPersistency(state *State, mtime *int) error {
	stateFilePath, err := sh.getStateFilePath()
	if err != nil {
		return errors.Wrap(err, "Failed getting state file path")
	}

	stateContents, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "Failed marshaling state file contents")
	}

	var condition string
	if mtime != nil {
		condition = fmt.Sprintf("__mtime_nsecs == %v", *mtime)
	}

	err = sh.streamConsumerGroup.container.UpdateItemSync(&v3io.UpdateItemInput{
		Path:      stateFilePath,
		Condition: condition,
		Attributes: map[string]interface{}{
			stateContentsAttributeKey: string(stateContents),
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed setting state in persistency")
	}

	return nil
}

func (sh *stateHandler) getStateFromPersistency() (*State, *int, error) {
	stateFilePath, err := sh.getStateFilePath()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed getting state file path")
	}

	response, err := sh.streamConsumerGroup.container.GetItemSync(&v3io.GetItemInput{
		Path:           stateFilePath,
		AttributeNames: []string{"__mtime_nsecs", stateContentsAttributeKey},
	})

	if err != nil {
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return nil, nil, errors.Wrap(err, "Got error without status code")
		}

		if errWithStatusCode.StatusCode() != 404 {
			return nil, nil, errors.Wrap(err, "Failed getting state item")
		}

		return nil, nil, v3ioerrors.ErrNotFound
	}

	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	stateContents, err := getItemOutput.Item.GetFieldString(stateContentsAttributeKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed getting state attribute")
	}

	var state State

	err = json.Unmarshal([]byte(stateContents), &state)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed unmarshalling state contents: %s", stateContents)
	}

	stateMtime, err := getItemOutput.Item.GetFieldInt("__mtime_nsecs")
	if err != nil {
		return nil, nil, errors.New("Failed getting mtime attribute")
	}

	return &state, &stateMtime, nil
}

func (sh *stateHandler) removeStaleSessionStates(state *State) error {

	// clear out the sessions since we only want the valid sessions
	state.SessionStates = []*SessionState{}

	for _, sessionState := range state.SessionStates {

		// check if the last heartbeat happened prior to the session timeout
		if time.Since(sessionState.LastHeartbeat) < sh.streamConsumerGroup.config.Session.Timeout {
			state.SessionStates = append(state.SessionStates, sessionState)
		}
	}

	return nil
}

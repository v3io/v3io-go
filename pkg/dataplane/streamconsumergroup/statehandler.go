package streamconsumergroup

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

const stateContentsAttributeKey string = "state"

type streamConsumerGroupStateHandler struct {
	logger                     logger.Logger
	streamConsumerGroup        *streamConsumerGroup
	lastState                  *State
	stopStateRefreshingChannel chan bool
}

func newStreamConsumerGroupStateHandler(streamConsumerGroup *streamConsumerGroup) (StateHandler, error) {
	return &streamConsumerGroupStateHandler{
		logger:                     streamConsumerGroup.logger.GetChild("stateHandler"),
		streamConsumerGroup:        streamConsumerGroup,
		stopStateRefreshingChannel: make(chan bool),
	}, nil
}

func (sh *streamConsumerGroupStateHandler) Start() error {
	state, err := sh.refreshState()
	if err != nil {
		return errors.Wrap(err, "Failed first refreshing state")
	}
	sh.lastState = state

	go sh.refreshStatePeriodically(sh.stopStateRefreshingChannel, sh.streamConsumerGroup.config.State.Heartbeat.Interval)

	return nil
}

func (sh *streamConsumerGroupStateHandler) Stop() error {
	sh.stopStateRefreshingChannel <- true
	return nil
}

func (sh *streamConsumerGroupStateHandler) refreshStatePeriodically(stopStateRefreshingChannel chan bool,
	heartbeatInterval time.Duration) {
	ticker := time.NewTicker(heartbeatInterval)

	for {
		select {
		case <-stopStateRefreshingChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			state, err := sh.refreshState()
			if err != nil {
				sh.logger.WarnWith("Failed refreshing state", "err", errors.GetErrorStackString(err, 10))
				continue
			}
			sh.lastState = state
		}
	}
}

func (sh *streamConsumerGroupStateHandler) refreshState() (*State, error) {
	newState, err := sh.modifyState(func(state *State) (*State, error) {
		now := time.Now()
		if state == nil {
			state = &State{
				SchemasVersion: "0.0.1",
				Sessions:       make([]SessionState, 0),
			}
		}

		// remove stale sessions
		validSessions := make([]SessionState, 0)
		for index, sessionState := range state.Sessions {

			sessionTimeout := sh.streamConsumerGroup.config.Session.Timeout
			if !now.After(sessionState.LastHeartbeat.Add(sessionTimeout)) {
				validSessions = append(validSessions, state.Sessions[index])
			}
		}
		state.Sessions = validSessions

		// create or update sessions in the state
		for _, member := range sh.streamConsumerGroup.members {
			var sessionState *SessionState
			for index, session := range state.Sessions {
				if session.MemberID == member.ID {
					session = state.Sessions[index]
					break
				}
			}
			if sessionState == nil {
				if state.Sessions == nil {
					state.Sessions = make([]SessionState, 0)
				}
				shards, err := sh.resolveShardsToAssign(state)
				if err != nil {
					return nil, errors.Wrap(err, "Failed resolving shards for session")
				}
				state.Sessions = append(state.Sessions, SessionState{
					MemberID:      member.ID,
					LastHeartbeat: &now,
					Shards:        shards,
				})
			}
			sessionState.LastHeartbeat = &now
		}

		return state, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed modifying state")
	}

	return newState, nil
}

func (sh *streamConsumerGroupStateHandler) resolveShardsToAssign(state *State) ([]int, error) {
	numberOfShards, err := sh.resolveNumberOfShards()
	if err != nil {
		return nil, errors.Wrap(err, "Failed resolving number of shards")
	}

	maxNumberOfShardsPerSession, err := sh.resolveMaxNumberOfShardsPerSession(numberOfShards, sh.streamConsumerGroup.maxWorkers)
	if err != nil {
		return nil, errors.Wrap(err, "Failed resolving max number of shards per session")
	}

	shardIDs := common.MakeRange(0, numberOfShards-1)
	shardsToAssign := make([]int, 0)
	for _, shardID := range shardIDs {
		found := false
		for _, session := range state.Sessions {
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
		if len(shardsToAssign) == maxNumberOfShardsPerSession {
			return shardsToAssign, nil
		}
	}

	// all shards are assigned
	if len(shardsToAssign) == 0 {
		// TODO: decide what to do
	}

	return shardsToAssign, nil
}

func (sh *streamConsumerGroupStateHandler) resolveMaxNumberOfShardsPerSession(numberOfShards int, maxWorkers int) (int, error) {
	if numberOfShards%maxWorkers != 0 {
		return numberOfShards/maxWorkers + 1, nil
	}
	return numberOfShards / maxWorkers, nil
}

func (sh *streamConsumerGroupStateHandler) resolveNumberOfShards() (int, error) {
	// get all shards in the stream
	response, err := sh.streamConsumerGroup.container.GetContainerContentsSync(&v3io.GetContainerContentsInput{
		DataPlaneInput: sh.streamConsumerGroup.dataPlaneInput,
		Path:           sh.streamConsumerGroup.streamPath,
	})

	if err != nil {
		return 0, errors.Wrapf(err, "Failed getting stream shards: %s", sh.streamConsumerGroup.streamPath)
	}

	defer response.Release()

	return len(response.Output.(*v3io.GetContainerContentsOutput).Contents), nil
}

func (sh *streamConsumerGroupStateHandler) modifyState(modifier stateModifier) (*State, error) {
	var modifiedState *State

	backoff := sh.streamConsumerGroup.config.State.ModifyRetry.Backoff
	attempts := sh.streamConsumerGroup.config.State.ModifyRetry.Attempts
	ctx := context.TODO()

	err := common.RetryFunc(ctx, sh.logger, attempts, nil, &backoff, func(_ int) (bool, error) {
		state, mtime, err := sh.getStateFromPersistency()
		if err != nil {
			if err != common.ErrNotFound {
				return true, errors.Wrap(err, "failed getting current state from persistency")
			}
			state = nil
			mtime = nil
		}

		modifiedState, err := modifier(state)
		if err != nil {
			return true, errors.Wrap(err, "Failed modifying state")
		}

		err = sh.setStateInPersistency(modifiedState, mtime)
		if err != nil {
			return true, errors.Wrap(err, "Failed setting state in persistency state")
		}

		return false, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed modifying state, attempts exhausted")
	}
	return modifiedState, nil
}

func (sh *streamConsumerGroupStateHandler) getStateFilePath() (string, error) {
	return path.Join(sh.streamConsumerGroup.streamPath, fmt.Sprintf("%s-state.json", sh.streamConsumerGroup.ID)), nil
}

func (sh *streamConsumerGroupStateHandler) setStateInPersistency(state *State, mtime *time.Time) error {
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
		condition = fmt.Sprintf("__mtime == %s", *mtime)
	}

	err = sh.streamConsumerGroup.container.UpdateItemSync(&v3io.UpdateItemInput{
		DataPlaneInput: sh.streamConsumerGroup.dataPlaneInput,
		Path:           stateFilePath,
		Attributes: map[string]interface{}{
			stateContentsAttributeKey: stateContents,
		},
		Condition: condition,
	})
	if err != nil {
		return errors.Wrap(err, "Failed setting state in persistency")
	}

	return nil
}

func (sh *streamConsumerGroupStateHandler) getStateFromPersistency() (*State, *time.Time, error) {
	stateFilePath, err := sh.getStateFilePath()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed getting state file path")
	}

	response, err := sh.streamConsumerGroup.container.GetItemSync(&v3io.GetItemInput{
		DataPlaneInput: sh.streamConsumerGroup.dataPlaneInput,
		Path:           stateFilePath,
		AttributeNames: []string{"__mtime", stateContentsAttributeKey},
	})
	if err != nil {
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return nil, nil, errors.Wrap(err, "Got error without status code")
		}
		if errWithStatusCode.StatusCode() != 404 {
			return nil, nil, errors.Wrap(err, "Failed getting state item")
		}
		return nil, nil, common.ErrNotFound
	}
	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	stateContentsInterface, foundStateAttribute := getItemOutput.Item[stateContentsAttributeKey]
	if !foundStateAttribute {
		return nil, nil, errors.New("Failed getting state attribute")
	}
	stateContents, ok := stateContentsInterface.(string)
	if !ok {
		return nil, nil, errors.Errorf("Unexpected type for state attribute: %s", reflect.TypeOf(stateContentsInterface))
	}

	var state State

	err = json.Unmarshal([]byte(stateContents), state)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed unmarshaling state contents")
	}

	mtimeInterface, foundMtimeAttribute := getItemOutput.Item["__mtime"]
	if !foundMtimeAttribute {
		return nil, nil, errors.New("Failed getting mtime attribute")
	}
	mtime, ok := mtimeInterface.(time.Time)
	if !ok {
		return nil, nil, errors.Errorf("Unexpected type for mtime attribute: %s", reflect.TypeOf(mtimeInterface))
	}

	return &state, &mtime, nil
}

func (sh *streamConsumerGroupStateHandler) GetMemberState(memberID string) (*SessionState, error) {
	for index, sessionState := range sh.lastState.Sessions {
		if sessionState.MemberID == memberID {
			return &sh.lastState.Sessions[index], nil
		}
	}
	return nil, errors.Errorf("Member state not found: %s", memberID)
}

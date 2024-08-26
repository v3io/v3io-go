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

package streamconsumergroup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/v3io/v3io-go/pkg/common"
	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroup struct {
	logger         logger.Logger
	name           string
	config         *Config
	container      v3io.Container
	streamPath     string
	maxReplicas    int
	totalNumShards int
}

func NewStreamConsumerGroup(parentLogger logger.Logger,
	name string,
	config *Config,
	container v3io.Container,
	streamPath string,
	maxReplicas int) (StreamConsumerGroup, error) {
	var err error

	if config == nil {
		config = NewConfig()
	}

	newStreamConsumerGroup := streamConsumerGroup{
		logger:      parentLogger.GetChild(name),
		name:        name,
		config:      config,
		container:   container,
		streamPath:  streamPath,
		maxReplicas: maxReplicas,
	}

	// get the total number of shards for this stream
	newStreamConsumerGroup.totalNumShards, err = newStreamConsumerGroup.getTotalNumberOfShards()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get total number of shards")
	}

	return &newStreamConsumerGroup, nil
}

func (scg *streamConsumerGroup) GetState() (*State, error) {
	state, _, _, err := scg.getStateFromPersistency()
	return state, err
}

func (scg *streamConsumerGroup) GetShardSequenceNumber(shardID int) (uint64, error) {
	return scg.getShardSequenceNumberFromPersistency(shardID)
}

func (scg *streamConsumerGroup) GetNumShards() (int, error) {
	return scg.totalNumShards, nil
}

func (scg *streamConsumerGroup) getShardPath(shardID int) (string, error) {
	return path.Join(scg.streamPath, strconv.Itoa(shardID)), nil
}

func (scg *streamConsumerGroup) getTotalNumberOfShards() (int, error) {
	response, err := scg.container.DescribeStreamSync(&v3io.DescribeStreamInput{
		Path: scg.streamPath,
	})
	if err != nil {
		return 0, errors.Wrapf(err, "Failed describing stream: %s", scg.streamPath)
	}

	defer response.Release()

	return response.Output.(*v3io.DescribeStreamOutput).ShardCount, nil
}

func (scg *streamConsumerGroup) setState(modifier stateModifier,
	handlePostSetStateInPersistency postSetStateInPersistencyHandler) (*State, error) {
	var previousState, modifiedState *State

	backoff := scg.config.State.ModifyRetry.Backoff
	attempts := scg.config.State.ModifyRetry.Attempts

	err := common.RetryFunc(context.TODO(), scg.logger, attempts, nil, &backoff, func(attempt int) (bool, error) {
		state, stateMtimeNanoSeconds, stateMtimeSeconds, err := scg.getStateFromPersistency()
		if err != nil && !errors.Is(err, v3ioerrors.ErrNotFound) {
			return true, errors.Wrap(err, "Failed getting current state from persistency")
		}
		if common.EngineErrorIsNonFatal(err) {
			return true, errors.Wrap(err, "Failed getting current state from persistency due to a network error")
		}

		if state == nil {
			state, err = newState()
			if err != nil {
				return true, errors.Wrap(err, "Failed to create state")
			}
		}

		// for logging
		previousState = state.deepCopy()

		modifiedState, err = modifier(state)
		if err != nil {
			if errors.Is(errors.RootCause(err), errShardRetention) {

				// if shard retention failed the member needs to be aborted, so we can stop retrying
				return false, errors.Wrap(err, "Failed modifying state")
			}
			return true, errors.Wrap(err, "Failed modifying state")
		}

		// log only on change
		if !scg.statesEqual(previousState, modifiedState) {
			scg.logger.DebugWith("Modified state, saving",
				"stateMtimeNanoSeconds", stateMtimeNanoSeconds,
				"stateMtimeSeconds", stateMtimeSeconds,
				"previousState", previousState,
				"modifiedState", modifiedState)
		}

		if err := scg.setStateInPersistency(modifiedState, stateMtimeNanoSeconds, stateMtimeSeconds); err != nil {
			if attempt%10 == 0 {
				scg.logger.DebugWith("Failed to set state in persistency",
					"attempt", attempt,
					"err", errors.RootCause(err).Error())
			}
			return true, errors.Wrap(err, "Failed setting state in persistency state")
		}

		if err := handlePostSetStateInPersistency(); err != nil {
			return false, errors.Wrap(err, "Failed handling post set state in persistency")
		}

		return false, nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Failed modifying state, attempts exhausted. currentState(%s)", previousState.String())
	}

	return modifiedState, nil
}

func (scg *streamConsumerGroup) setStateInPersistency(
	state *State,
	stateMtimeNanoSeconds *int,
	stateMtimeSeconds *int) error {
	stateContents, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "Failed marshaling state file contents")
	}

	var condition string
	if stateMtimeNanoSeconds != nil && stateMtimeSeconds != nil {
		condition = fmt.Sprintf("(__mtime_nsecs == %v) AND (__mtime_secs == %v)",
			*stateMtimeNanoSeconds,
			*stateMtimeSeconds)
	} else {

		// mtime does not exist => file does not exist => create it
		// we want the file to be created by one replica only and thus
		// we condition the creation of it by checking if the state attribute
		// does not exist
		condition = fmt.Sprintf("not(exists(%s))", stateContentsAttributeKey)
	}

	if _, err := scg.container.UpdateItemSync(&v3io.UpdateItemInput{
		Path:      scg.getStateFilePath(),
		Condition: condition,
		Attributes: map[string]interface{}{
			stateContentsAttributeKey: string(stateContents),
		},
	}); err != nil {
		return errors.Wrap(err, "Failed setting state in persistency")
	}

	return nil
}

func (scg *streamConsumerGroup) getStateFromPersistency() (*State, *int, *int, error) {
	response, err := scg.container.GetItemSync(&v3io.GetItemInput{
		Path: scg.getStateFilePath(),
		AttributeNames: []string{
			"__mtime_nsecs",
			"__mtime_secs",
			stateContentsAttributeKey,
		},
	})

	if err != nil {
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return nil, nil, nil, errors.Wrap(err, "Got error without status code")
		}

		if errWithStatusCode.StatusCode() != 404 {
			return nil, nil, nil, errors.Wrap(err, "Failed getting state item")
		}

		return nil, nil, nil, v3ioerrors.ErrNotFound
	}

	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	stateContents, err := getItemOutput.Item.GetFieldString(stateContentsAttributeKey)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Failed getting state attribute")
	}

	var state State

	if err := json.Unmarshal([]byte(stateContents), &state); err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Failed unmarshalling state contents: %s", stateContents)
	}

	stateMtimeNanoSeconds, err := getItemOutput.Item.GetFieldInt("__mtime_nsecs")
	if err != nil {
		return nil, nil, nil, errors.New("Failed getting mtime attribute")
	}

	stateMtimeSeconds, err := getItemOutput.Item.GetFieldInt("__mtime_secs")
	if err != nil {
		return nil, nil, nil, errors.New("Failed getting mtime attribute")
	}

	return &state, &stateMtimeNanoSeconds, &stateMtimeSeconds, nil
}

func (scg *streamConsumerGroup) getStateFilePath() string {
	return path.Join(scg.streamPath, fmt.Sprintf("%s-state.json", scg.name))
}

func (scg *streamConsumerGroup) getShardLocationFromPersistency(shardID int,
	initialLocation v3io.SeekShardInputType) (string, error) {

	seekShardInput := v3io.SeekShardInput{}

	// get the shard sequenceNumber from the item
	shardSequenceNumber, err := scg.getShardSequenceNumberFromPersistency(shardID)
	if err != nil {

		// if the error is that the attribute wasn't found, but the shard was found - seek the shard
		// according to the configuration
		if !errors.Is(err, ErrShardSequenceNumberAttributeNotFound) {
			return "", errors.Wrap(err, "Failed to get shard sequenceNumber from item attributes")
		}

		seekShardInput.Type = initialLocation
	} else {

		// use sequence number
		seekShardInput.Type = v3io.SeekShardInputTypeSequence
		seekShardInput.StartingSequenceNumber = shardSequenceNumber + 1
	}

	seekShardInput.Path, err = scg.getShardPath(shardID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}

	return scg.getShardLocationWithSeek(&seekShardInput)
}

// returns the sequenceNumber, an error re: the shard itself and an error re: the attribute in the shard
func (scg *streamConsumerGroup) getShardSequenceNumberFromPersistency(shardID int) (uint64, error) {
	shardPath, err := scg.getShardPath(shardID)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}

	response, err := scg.container.GetItemSync(&v3io.GetItemInput{
		Path:           shardPath,
		AttributeNames: []string{scg.getShardCommittedSequenceNumberAttributeName()},
	})

	if err != nil {
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return 0, errors.Wrap(err, "Got error without status code")
		}

		if errWithStatusCode.StatusCode() != http.StatusNotFound {
			return 0, errors.Wrap(err, "Failed getting shard item")
		}

		return 0, ErrShardNotFound
	}

	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	// return the attribute name
	sequenceNumber, err := getItemOutput.Item.GetFieldUint64(scg.getShardCommittedSequenceNumberAttributeName())
	if err != nil && errors.Is(err, v3ioerrors.ErrNotFound) {
		return 0, ErrShardSequenceNumberAttributeNotFound
	}

	// return the sequenceNumber we found
	return sequenceNumber, nil
}

func (scg *streamConsumerGroup) getShardLocationWithSeek(seekShardInput *v3io.SeekShardInput) (string, error) {
	scg.logger.DebugWith("Seeking shard", "shardPath", seekShardInput.Path, "seekShardInput", seekShardInput)

	response, err := scg.container.SeekShardSync(seekShardInput)
	if err != nil {
		return "", errors.Wrap(err, "Failed to seek shard")
	}
	defer response.Release()

	location := response.Output.(*v3io.SeekShardOutput).Location

	scg.logger.DebugWith("Seek shard succeeded",
		"startingSequenceNumber", seekShardInput.StartingSequenceNumber,
		"shardPath", seekShardInput.Path,
		"location", location)

	return location, nil
}

func (scg *streamConsumerGroup) getShardCommittedSequenceNumberAttributeName() string {
	return fmt.Sprintf("__%s_committed_sequence_number", scg.name)
}

func (scg *streamConsumerGroup) setShardSequenceNumberInPersistency(shardID int, sequenceNumber uint64) error {
	scg.logger.DebugWith("Setting shard sequenceNumber in persistency", "shardID", shardID, "sequenceNumber", sequenceNumber)
	shardPath, err := scg.getShardPath(shardID)
	if err != nil {
		return errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}

	_, err = scg.container.UpdateItemSync(&v3io.UpdateItemInput{
		Path: shardPath,
		Attributes: map[string]interface{}{
			scg.getShardCommittedSequenceNumberAttributeName(): sequenceNumber,
		},
		Condition: "__obj_type == 3",
	})
	return err
}

// returns true if the states are equal, ignoring heartbeat times
func (scg *streamConsumerGroup) statesEqual(state0 *State, state1 *State) bool {
	if state0.SchemasVersion != state1.SchemasVersion {
		return false
	}

	if len(state0.SessionStates) != len(state1.SessionStates) {
		return false
	}

	// since we compared lengths, we can only do this from state0
	for _, state0SessionState := range state0.SessionStates {
		session1SessionState := state1.findSessionStateByMemberID(state0SessionState.MemberID)

		// if couldn't find session state
		if session1SessionState == nil {
			return false
		}

		if !common.IntSlicesEqual(state0SessionState.Shards, session1SessionState.Shards) {
			return false
		}
	}

	return true
}

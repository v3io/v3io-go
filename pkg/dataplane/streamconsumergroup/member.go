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
	"fmt"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/rs/xid"
)

type member struct {
	id                    string
	logger                logger.Logger
	streamConsumerGroup   *streamConsumerGroup
	stateHandler          *stateHandler
	sequenceNumberHandler *sequenceNumberHandler
	handler               Handler
	session               Session
	retainShards          bool
	shardGroupToRetain    []int
}

func NewMember(streamConsumerGroupInterface StreamConsumerGroup, name string) (Member, error) {
	var err error

	streamConsumerGroupInstance, ok := streamConsumerGroupInterface.(*streamConsumerGroup)
	if !ok {
		return nil, errors.Errorf("Expected streamConsumerGroupInterface of type streamConsumerGroup, got %T", streamConsumerGroupInterface)
	}

	// add uniqueness
	id := fmt.Sprintf("%s-%s", name, xid.New().String())

	newMember := member{
		logger:              streamConsumerGroupInstance.logger.GetChild(id),
		id:                  id,
		streamConsumerGroup: streamConsumerGroupInstance,
	}

	// create & start a state handler for the stream
	newMember.stateHandler, err = newStateHandler(&newMember)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating stream consumer group state handler")
	}

	// create & start a location handler for the stream
	newMember.sequenceNumberHandler, err = newSequenceNumberHandler(&newMember)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating stream consumer group location handler")
	}

	if err := newMember.Start(); err != nil {
		return nil, errors.Wrap(err, "Failed starting new member")
	}

	return &newMember, nil
}

func (m *member) Consume(handler Handler) error {
	m.logger.DebugWith("Starting consumption of consumer group")

	m.handler = handler

	// get the state (holding our shards)
	sessionState, err := m.stateHandler.getOrCreateSessionState(m.id)
	if err != nil {
		return errors.Wrap(err, "Failed getting stream consumer group member state")
	}

	// create a session object from our state
	m.session, err = newSession(m, sessionState)
	if err != nil {
		return errors.Wrap(err, "Failed creating stream consumer group session")
	}

	// start it
	return m.session.start()
}

func (m *member) Close() error {
	m.logger.DebugWith("Closing consumer group")

	if err := m.stateHandler.stop(); err != nil {
		return errors.Wrapf(err, "Failed stopping state handler")
	}
	if err := m.sequenceNumberHandler.stop(); err != nil {
		return errors.Wrapf(err, "Failed stopping location handler")
	}

	if m.session != nil {
		if err := m.session.stop(); err != nil {
			return errors.Wrap(err, "Failed stopping member session")
		}
	}

	return nil
}

func (m *member) Start() error {
	if err := m.stateHandler.start(); err != nil {
		return errors.Wrap(err, "Failed starting stream consumer group state handler")
	}

	if err := m.sequenceNumberHandler.start(); err != nil {
		return errors.Wrap(err, "Failed starting stream consumer group state handler")
	}

	return nil
}

func (m *member) GetID() string {
	return m.id
}

func (m *member) GetRetainShardFlag() bool {
	return m.retainShards
}

func (m *member) GetShardsToRetain() []int {
	return m.shardGroupToRetain
}

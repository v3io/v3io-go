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
	"time"

	v3io "github.com/v3io/v3io-go/pkg/dataplane"
)

type stateModifier func(*State) (*State, error)

type postSetStateInPersistencyHandler func() error

type SessionState struct {
	MemberID      string    `json:"member_id"`
	LastHeartbeat time.Time `json:"last_heartbeat_time"`
	Shards        []int     `json:"shards"`
}

type Handler interface {

	// Setup is run at the beginning of a new session, before ConsumeClaim.
	Setup(Session) error

	// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
	// but before the locations are committed for the very last time.
	Cleanup(Session) error

	// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
	// Once the Messages() channel is closed, the Handler must finish its processing
	// loop and exit.
	ConsumeClaim(Session, Claim) error

	// Abort signals the handler to start abort procedure
	Abort(Session) error
}

type RecordBatch struct {
	Records      []v3io.StreamRecord
	Location     string
	NextLocation string
	ShardID      int
}

type StreamConsumerGroup interface {
	GetState() (*State, error)
	GetShardSequenceNumber(int) (uint64, error)
	GetNumShards() (int, error)
}

type Member interface {
	Consume(Handler) error
	Close() error
	Start() error
	GetID() string
	GetRetainShardFlag() bool
	GetShardsToRetain() []int
}

type Session interface {
	GetClaims() []Claim
	GetMemberID() string
	MarkRecord(*v3io.StreamRecord) error

	start() error
	stop() error
}

type Claim interface {
	GetStreamPath() string
	GetShardID() int
	GetCurrentLocation() string
	GetRecordBatchChan() <-chan *RecordBatch

	start() error
	stop() error
}

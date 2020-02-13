package streamconsumergroup

import (
	"fmt"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type session struct {
	logger              logger.Logger
	streamConsumerGroup *streamConsumerGroup
	state               *SessionState
	claims              []Claim
}

func newSession(streamConsumerGroup *streamConsumerGroup,
	sessionState *SessionState) (Session, error) {

	return &session{
		logger:              streamConsumerGroup.logger.GetChild(fmt.Sprintf("session")),
		streamConsumerGroup: streamConsumerGroup,
		state:               sessionState,
		claims:              make([]Claim, 0),
	}, nil
}

func (s *session) Start() error {
	s.logger.DebugWith("Starting session")

	// for each shard we need handle, create a StreamConsumerGroupClaim object and start it
	for _, shardID := range s.state.Shards {
		claim, err := newClaim(s.streamConsumerGroup, shardID)
		if err != nil {
			return errors.Wrapf(err, "Failed creating stream consumer group claim for shard: %d", shardID)
		}

		// add to claims
		s.claims = append(s.claims, claim)
	}

	// tell the consumer group handler to set up
	s.logger.DebugWith("Triggering given handler Setup")
	if err := s.streamConsumerGroup.handler.Setup(s); err != nil {
		return errors.Wrap(err, "Failed to set up session")
	}

	s.logger.DebugWith("Starting claim consumption")
	for _, claim := range s.claims {
		if err := claim.Start(); err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

func (s *session) Stop() error {
	s.logger.DebugWith("Stopping session, triggering given handler cleanup")

	// tell the consumer group handler to set up
	if err := s.streamConsumerGroup.handler.Cleanup(s); err != nil {
		return errors.Wrap(err, "Failed to cleanup")
	}

	s.logger.DebugWith("Stopping claims")

	for _, claim := range s.claims {
		err := claim.Stop()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

func (s *session) GetClaims() []Claim {
	return s.claims
}

func (s *session) GetMemberID() string {
	return s.streamConsumerGroup.memberID
}

func (s *session) MarkRecordBatch(recordBatch *RecordBatch) error {
	s.logger.DebugWith("Marking record batch as consumed")
	err := s.streamConsumerGroup.locationHandler.markShardLocation(recordBatch.ShardID, recordBatch.NextLocation)
	if err != nil {
		return errors.Wrap(err, "Failed marking record batch as consumed")
	}

	return nil
}

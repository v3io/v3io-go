package streamconsumergroup

import (
	"fmt"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroupStateSession struct {
	logger              logger.Logger
	streamConsumerGroup *streamConsumerGroup
	state               *SessionState
	member              *streamConsumerGroupMember
	claims              []v3io.StreamConsumerGroupClaim
}

func newStreamConsumerGroupSession(streamConsumerGroup *streamConsumerGroup,
	state *SessionState,
	member *streamConsumerGroupMember) (v3io.StreamConsumerGroupSession, error) {

	return &streamConsumerGroupStateSession{
		logger:              streamConsumerGroup.logger.GetChild(fmt.Sprintf("session-%s", member.ID)),
		streamConsumerGroup: streamConsumerGroup,
		state:               state,
		member:              member,
		claims:              make([]v3io.StreamConsumerGroupClaim, 0),
	}, nil
}

func (sh *streamConsumerGroupStateSession) Start() error {
	sh.logger.DebugWith("Starting session")
	// for each shard we need handle, create a StreamConsumerGroupClaim object and start it
	for _, shardID := range sh.state.Shards {
		streamConsumerGroupClaim, err := newStreamConsumerGroupClaim(sh.streamConsumerGroup, shardID, sh.member)
		if err != nil {
			return errors.Wrapf(err, "Failed creating stream consumer group claim for shard: %v", shardID)
		}

		// add to claims
		sh.claims = append(sh.claims, streamConsumerGroupClaim)
	}

	// tell the consumer group handler to set up
	sh.logger.DebugWith("Triggering given handler Setup")
	sh.member.handler.Setup(sh)

	sh.logger.DebugWith("Starting claims")
	for _, claim := range sh.claims {
		err := claim.Start()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

func (sh *streamConsumerGroupStateSession) Stop() error {
	sh.logger.DebugWith("Stopping session, triggering given handler cleanup")
	// tell the consumer group handler to set up
	sh.member.handler.Cleanup(sh)

	sh.logger.DebugWith("Stopping claims")
	for _, claim := range sh.claims {
		err := claim.Stop()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

func (sh *streamConsumerGroupStateSession) Claims() ([]int, error) {
	claimedShardIDs := make([]int, 0)
	for _, claim := range sh.claims {
		shardID, err := claim.Shard()
		if err != nil {
			return nil, errors.Wrap(err, "Failed getting claim shard")
		}
		claimedShardIDs = append(claimedShardIDs, shardID)
	}
	return claimedShardIDs, nil
}

func (sh *streamConsumerGroupStateSession) MemberID() (string, error) {
	return sh.member.ID, nil
}

func (sh *streamConsumerGroupStateSession) MarkChunk(streamChunk *v3io.StreamChunk) error {
	sh.logger.DebugWith("Marking chunk as consumed")
	err := sh.streamConsumerGroup.locationHandler.MarkLocation(streamChunk.ShardID, streamChunk.NextLocation)
	if err != nil {
		errors.Wrap(err, "Failed marking chunk as consumed")
	}
	return nil
}

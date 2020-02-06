package streamconsumergroup

import (
	"fmt"
	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/v3io/v3io-go/pkg/dataplane"
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
	sh.member.handler.Setup(sh)

	// start the claims
	for _, claim := range sh.claims {
		err := claim.Start()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

func (sh *streamConsumerGroupStateSession) Stop() error {
	// tell the consumer group handler to set up
	sh.member.handler.Cleanup(sh)

	// stop the claims
	for _, claim := range sh.claims {
		err := claim.Stop()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group claim")
		}
	}

	return nil
}

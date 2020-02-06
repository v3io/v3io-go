package streamconsumergroup

import (
	"path"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroup struct {
	ID              string
	logger          logger.Logger
	config          *Config
	dataPlaneInput  v3io.DataPlaneInput
	members         []streamConsumerGroupMember
	streamPath      string
	maxWorkers      int
	stateHandler    StateHandler
	locationHandler LocationHandler
	container       v3io.Container
}

type streamConsumerGroupMember struct {
	ID      string
	handler v3io.StreamConsumerGroupHandler
	session v3io.StreamConsumerGroupSession
}

func NewStreamConsumerGroup(id string,
	parentLogger logger.Logger,
	config *Config,
	dataPlaneInput v3io.DataPlaneInput,
	streamPath string,
	maxWorkers int,
	container v3io.Container) (v3io.StreamConsumerGroup, error) {

	if config == nil {
		config = NewConfig()
	}

	return &streamConsumerGroup{
		ID:             id,
		logger:         parentLogger.GetChild("streamConsumerGroup"),
		config:         config,
		dataPlaneInput: dataPlaneInput,
		members:        make([]streamConsumerGroupMember, 0),
		streamPath:     streamPath,
		maxWorkers:     maxWorkers,
		container:      container,
	}, nil
}

func (scg *streamConsumerGroup) Consume(memberID string, streamConsumerGroupHandler v3io.StreamConsumerGroupHandler) error {

	member := streamConsumerGroupMember{
		ID:      memberID,
		handler: streamConsumerGroupHandler,
	}

	if scg.stateHandler == nil {

		// create & start a state handler for the stream
		stateHandler, err := newStreamConsumerGroupStateHandler(scg)
		if err != nil {
			return errors.Wrap(err, "Failed creating stream consumer group state handler")
		}
		err = stateHandler.Start()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group state handler")
		}

		scg.stateHandler = stateHandler
	}

	if scg.locationHandler == nil {

		// create & start an location handler for the stream
		locationHandler, err := newStreamConsumerGroupLocationHandler(scg)
		if err != nil {
			return errors.Wrap(err, "Failed creating stream consumer group location handler")
		}
		err = locationHandler.Start()
		if err != nil {
			return errors.Wrap(err, "Failed starting stream consumer group state handler")
		}
		scg.locationHandler = locationHandler
	}

	// get the state (holding our shards)
	state, err := scg.stateHandler.GetMemberState(memberID)
	if err != nil {
		return errors.Wrap(err, "Failed getting stream consumer group state")
	}

	// create a session object from our state
	streamConsumerGroupSession, err := newStreamConsumerGroupSession(scg, state, &member)
	if err != nil {
		return errors.Wrap(err, "Failed creating stream consumer group session")
	}

	member.session = streamConsumerGroupSession

	scg.members = append(scg.members, member)

	// start it
	streamConsumerGroupSession.Start()

	return nil
}

func (scg *streamConsumerGroup) Close() error {
	for _, member := range scg.members {
		if err := member.session.Stop(); err != nil {
			return errors.Wrapf(err, "Failed stopping member session: %s", member.ID)
		}
	}
	return nil
}

func (scg *streamConsumerGroup) seekShard(shardID int, inputType v3io.SeekShardInputType) (string, error) {
	shardPath, err := scg.getShardPath(shardID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}
	seekShardInput := v3io.SeekShardInput{
		DataPlaneInput: scg.dataPlaneInput,
		Path:           shardPath,
		Type:           inputType,
	}

	response, err := scg.container.SeekShardSync(&seekShardInput)
	if err != nil {
		return "", errors.Wrapf(err, "Failed seeking shard: %v", shardID)
	}
	defer response.Release()

	return response.Output.(*v3io.SeekShardOutput).Location, nil
}

func (scg *streamConsumerGroup) getShardPath(shardID int) (string, error) {
	return path.Join(scg.streamPath, string(shardID)), nil
}

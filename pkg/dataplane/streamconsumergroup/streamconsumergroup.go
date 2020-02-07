package streamconsumergroup

import (
	"fmt"
	"path"
	"strconv"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroup struct {
	ID              string
	logger          logger.Logger
	config          *Config
	dataPlaneInput  v3io.DataPlaneInput
	members         map[string]*streamConsumerGroupMember
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
		logger:         parentLogger.GetChild(fmt.Sprintf("streamConsumerGroup-%s", id)),
		config:         config,
		dataPlaneInput: dataPlaneInput,
		members:        make(map[string]*streamConsumerGroupMember, 0),
		streamPath:     streamPath,
		maxWorkers:     maxWorkers,
		container:      container,
	}, nil
}

func (scg *streamConsumerGroup) Consume(memberID string, streamConsumerGroupHandler v3io.StreamConsumerGroupHandler) error {
	scg.logger.DebugWith("Member requesting to consume", "memberID", memberID)

	member := streamConsumerGroupMember{
		ID:      memberID,
		handler: streamConsumerGroupHandler,
	}

	scg.members[memberID] = &member

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
	state, err := scg.stateHandler.GetOrCreateMemberState(memberID)
	if err != nil {
		return errors.Wrap(err, "Failed getting stream consumer group state")
	}

	// create a session object from our state
	streamConsumerGroupSession, err := newStreamConsumerGroupSession(scg, state, &member)
	if err != nil {
		return errors.Wrap(err, "Failed creating stream consumer group session")
	}

	member.session = streamConsumerGroupSession

	// start it
	streamConsumerGroupSession.Start()

	return nil
}

func (scg *streamConsumerGroup) Close() error {
	scg.logger.DebugWith("Closing consumer group")
	if scg.stateHandler != nil {
		if err := scg.stateHandler.Stop(); err != nil {
			return errors.Wrapf(err, "Failed stopping state handler")
		}
	}
	if scg.locationHandler != nil {
		if err := scg.locationHandler.Stop(); err != nil {
			return errors.Wrapf(err, "Failed stopping location handler")
		}
	}
	for _, member := range scg.members {
		if member.session != nil {
			if err := member.session.Stop(); err != nil {
				return errors.Wrapf(err, "Failed stopping member session: %s", member.ID)
			}
		}
	}
	return nil
}

func (scg *streamConsumerGroup) seekShard(shardID int, seekShardType v3io.SeekShardInputType) (string, error) {
	shardPath, err := scg.getShardPath(shardID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}
	seekShardInput := v3io.SeekShardInput{
		DataPlaneInput: scg.dataPlaneInput,
		Path:           shardPath,
		Type:           seekShardType,
	}

	scg.logger.DebugWith("Seeking shard", "shardID", shardID, "seekShardType", seekShardType)

	response, err := scg.container.SeekShardSync(&seekShardInput)
	if err != nil {
		//TODO: understand why it is failing with bad request (instead of not found) when shard does not exist
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return "", errors.Wrap(err, "Got error without status code")
		}
		if errWithStatusCode.StatusCode() != 404 {
			return "", errors.Wrapf(err, "Failed seeking shard: %v", shardID)
		}
		scg.logger.DebugWith("Seeking shard failed on not found, it is ok")
		return "", common.ErrNotFound
	}
	defer response.Release()

	location := response.Output.(*v3io.SeekShardOutput).Location

	scg.logger.DebugWith("Seek shard succeeded", "shardID", shardID, "location", location)

	return location, nil
}

func (scg *streamConsumerGroup) getShardPath(shardID int) (string, error) {
	return path.Join(scg.streamPath, strconv.Itoa(shardID)), nil
}

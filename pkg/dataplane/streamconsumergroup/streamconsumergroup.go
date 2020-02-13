package streamconsumergroup

import (
	"fmt"
	"path"
	"strconv"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
	"github.com/v3io/v3io-go/pkg/dataplane"
)

type streamConsumerGroup struct {
	name            string
	memberID        string
	logger          logger.Logger
	config          *Config
	streamPath      string
	maxReplicas     int
	stateHandler    *stateHandler
	locationHandler *locationHandler
	container       v3io.Container
	handler         Handler
	session         Session
	totalNumShards  int
}

func NewStreamConsumerGroup(name string,
	memberID string,
	parentLogger logger.Logger,
	config *Config,
	streamPath string,
	maxReplicas int,
	container v3io.Container) (StreamConsumerGroup, error) {
	var err error

	if config == nil {
		config = NewConfig()
	}

	newStreamConsumerGroup := streamConsumerGroup{
		name:        name,
		memberID:    memberID,
		logger:      parentLogger.GetChild(fmt.Sprintf("streamConsumerGroup-%s", name)),
		config:      config,
		streamPath:  streamPath,
		maxReplicas: maxReplicas,
		container:   container,
	}

	// get the total number of shards for this stream
	newStreamConsumerGroup.totalNumShards, err = newStreamConsumerGroup.getTotalNumberOfShards()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get total number of shards")
	}

	// create & start a state handler for the stream
	newStreamConsumerGroup.stateHandler, err = newStateHandler(&newStreamConsumerGroup)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating stream consumer group state handler")
	}

	err = newStreamConsumerGroup.stateHandler.start()
	if err != nil {
		return nil, errors.Wrap(err, "Failed starting stream consumer group state handler")
	}

	// create & start an location handler for the stream
	newStreamConsumerGroup.locationHandler, err = newLocationHandler(&newStreamConsumerGroup)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating stream consumer group location handler")
	}

	err = newStreamConsumerGroup.locationHandler.start()
	if err != nil {
		return nil, errors.Wrap(err, "Failed starting stream consumer group state handler")
	}

	return &newStreamConsumerGroup, nil
}

func (scg *streamConsumerGroup) Consume(handler Handler) error {
	scg.logger.DebugWith("Member requesting to consume")

	scg.handler = handler

	// get the state (holding our shards)
	sessionState, err := scg.stateHandler.getOrCreateSessionState(scg.memberID)
	if err != nil {
		return errors.Wrap(err, "Failed getting stream consumer group member state")
	}

	// create a session object from our state
	scg.session, err = newSession(scg, sessionState)
	if err != nil {
		return errors.Wrap(err, "Failed creating stream consumer group session")
	}

	// start it
	return scg.session.Start()
}

func (scg *streamConsumerGroup) Close() error {
	scg.logger.DebugWith("Closing consumer group")

	if err := scg.stateHandler.stop(); err != nil {
		return errors.Wrapf(err, "Failed stopping state handler")
	}
	if err := scg.locationHandler.stop(); err != nil {
		return errors.Wrapf(err, "Failed stopping location handler")
	}

	if scg.session != nil {
		if err := scg.session.Stop(); err != nil {
			return errors.Wrap(err, "Failed stopping member session")
		}
	}

	return nil
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

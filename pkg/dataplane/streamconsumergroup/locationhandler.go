package streamconsumergroup

import (
	"fmt"
	"reflect"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroupLocationHandler struct {
	logger                     logger.Logger
	streamConsumerGroup        *streamConsumerGroup
	shardLocationsCache        map[int]string
	stopCacheCommittingChannel chan bool
}

func newStreamConsumerGroupLocationHandler(streamConsumerGroup *streamConsumerGroup) (LocationHandler, error) {

	return &streamConsumerGroupLocationHandler{
		logger:                     streamConsumerGroup.logger.GetChild("locationHandler"),
		streamConsumerGroup:        streamConsumerGroup,
		shardLocationsCache:        make(map[int]string, 0),
		stopCacheCommittingChannel: make(chan bool),
	}, nil
}

func (lh *streamConsumerGroupLocationHandler) Start() error {
	commitCacheInterval := lh.streamConsumerGroup.config.Location.CommitCache.Interval
	go lh.commitCachePeriodically(lh.stopCacheCommittingChannel, commitCacheInterval)
	return nil
}

func (lh *streamConsumerGroupLocationHandler) Stop() error {
	lh.stopCacheCommittingChannel <- true
	return nil
}

func (lh *streamConsumerGroupLocationHandler) MarkLocation(shardID int, location string) error {
	lh.shardLocationsCache[shardID] = location
	return nil
}

func (lh *streamConsumerGroupLocationHandler) GetLocation(shardID int) (string, error) {
	location, found := lh.shardLocationsCache[shardID]
	if !found {
		location, err := lh.getShardLocationFromPersistency(shardID)
		if err != nil {
			return "", errors.Wrap(err, "Failed getting shard location from persistency")
		}
		lh.shardLocationsCache[shardID] = location
	}
	return location, nil
}

func (lh *streamConsumerGroupLocationHandler) getShardLocationFromPersistency(shardID int) (string, error) {
	shardPath, err := lh.streamConsumerGroup.getShardPath(shardID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}
	shardLocationAttribute, err := lh.getShardLocationAttributeName()
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard location attribute")
	}
	response, err := lh.streamConsumerGroup.container.GetItemSync(&v3io.GetItemInput{
		DataPlaneInput: lh.streamConsumerGroup.dataPlaneInput,
		Path:           shardPath,
		AttributeNames: []string{shardLocationAttribute},
	})
	if err != nil {
		return "", errors.Wrap(err, "Failed getting shard location item from persistency")
	}
	defer response.Release()
	getItemOutput := response.Output.(*v3io.GetItemOutput)

	shardLocationInterface, foundShardLocationAttribute := getItemOutput.Item[shardLocationAttribute]
	if !foundShardLocationAttribute {
		inputType := lh.streamConsumerGroup.config.Shard.InputType
		shardLocation, err := lh.streamConsumerGroup.seekShard(shardID, inputType)
		if err != nil {
			return "", errors.Wrapf(err, "Failed seeking shard: %v", shardID)
		}
		return shardLocation, nil
	}
	shardLocation, ok := shardLocationInterface.(string)
	if !ok {
		return "", errors.Errorf("Unexpected type for state attribute: %s", reflect.TypeOf(shardLocationInterface))
	}

	return shardLocation, nil
}

func (lh *streamConsumerGroupLocationHandler) getShardLocationAttributeName() (string, error) {
	return fmt.Sprintf("__%s_location", lh.streamConsumerGroup.ID), nil
}

func (lh *streamConsumerGroupLocationHandler) setSharedLocationInPersistency(shardID int, location string) error {
	shardPath, err := lh.streamConsumerGroup.getShardPath(shardID)
	if err != nil {
		return errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}
	shardLocationAttribute, err := lh.getShardLocationAttributeName()
	if err != nil {
		return errors.Wrapf(err, "Failed getting shard location attribute")
	}
	err = lh.streamConsumerGroup.container.UpdateItemSync(&v3io.UpdateItemInput{
		DataPlaneInput: lh.streamConsumerGroup.dataPlaneInput,
		Path:           shardPath,
		Attributes: map[string]interface{}{
			shardLocationAttribute: location,
		},
	})
	return nil
}

func (lh *streamConsumerGroupLocationHandler) commitCachePeriodically(stopCacheCommittingChannel chan bool,
	commitCacheInterval time.Duration) {
	ticker := time.NewTicker(commitCacheInterval)

	for {
		select {
		case <-stopCacheCommittingChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			err := lh.commitCache()
			if err != nil {
				lh.logger.WarnWith("Failed committing cache", "err", errors.GetErrorStackString(err, 10))
				continue
			}
		}
	}
}

func (lh *streamConsumerGroupLocationHandler) commitCache() error {
	failedShardIDs := make([]int, 0)
	for shardID, location := range lh.shardLocationsCache {
		err := lh.setSharedLocationInPersistency(shardID, location)
		if err != nil {
			lh.logger.WarnWith("Failed committing shard location", "shardID", shardID,
				"location", location,
				"err", errors.GetErrorStackString(err, 10))
			failedShardIDs = append(failedShardIDs, shardID)
		}
	}
	if len(failedShardIDs) > 0 {
		return errors.Errorf("Failed committing cache in shards: %v", failedShardIDs)
	}
	return nil
}

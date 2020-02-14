package streamconsumergroup

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

var errShardNotFound = errors.New("Shard not found")
var errShardLocationAttributeNotFound = errors.New("Shard location attribute")

type locationHandler struct {
	logger                               logger.Logger
	streamConsumerGroup                  *streamConsumerGroup
	markedShardLocations                 []string
	markedShardLocationsLock             sync.RWMutex
	stopMarkedShardLocationCommitterChan chan struct{}
	lastCommittedShardLocations          []string
}

func newLocationHandler(streamConsumerGroup *streamConsumerGroup) (*locationHandler, error) {

	return &locationHandler{
		logger:                               streamConsumerGroup.logger.GetChild("locationHandler"),
		streamConsumerGroup:                  streamConsumerGroup,
		markedShardLocations:                 make([]string, streamConsumerGroup.totalNumShards),
		stopMarkedShardLocationCommitterChan: make(chan struct{}, 1),
	}, nil
}

func (lh *locationHandler) start() error {
	lh.logger.DebugWith("Starting location handler")

	// stopped on stop()
	go lh.markedShardLocationsCommitter(lh.streamConsumerGroup.config.Location.Commit.Interval,
		lh.stopMarkedShardLocationCommitterChan)

	return nil
}

func (lh *locationHandler) stop() error {
	lh.logger.DebugWith("Stopping location handler")

	select {
	case lh.stopMarkedShardLocationCommitterChan <- struct{}{}:
	default:
	}

	return nil
}

func (lh *locationHandler) markShardLocation(shardID int, location string) error {

	// lock semantics are reverse - it's OK to write in parallel since each write goes
	// to a different cell in the array, but once a read is happening we need to stop the world
	lh.markedShardLocationsLock.RLock()
	lh.markedShardLocations[shardID] = location
	lh.markedShardLocationsLock.RUnlock()

	return nil
}

func (lh *locationHandler) getShardLocationFromPersistency(shardID int) (string, error) {
	lh.logger.DebugWith("Getting shard location from persistency", "shardID", shardID)

	shardPath, err := lh.streamConsumerGroup.getShardPath(shardID)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}

	// get the shard location from the item
	shardLocation, err := lh.getShardLocationFromItemAttributes(shardPath)
	if err != nil {

		// if the error is that the attribute wasn't found, but the shard was found - seek the shard
		// according to the configuration
		if err == errShardLocationAttributeNotFound {
			return lh.getShardLocationWithSeek(shardPath, lh.streamConsumerGroup.config.Claim.RecordBatchFetch.InitialLocation)
		}

		return "", errors.Wrap(err, "Failed to get shard location from item attributes")
	}

	return shardLocation, nil
}

// returns the location, an error re: the shard itself and an error re: the attribute in the shard
func (lh *locationHandler) getShardLocationFromItemAttributes(shardPath string) (string, error) {
	response, err := lh.streamConsumerGroup.container.GetItemSync(&v3io.GetItemInput{
		Path:           shardPath,
		AttributeNames: []string{lh.getShardLocationAttributeName()},
	})

	if err != nil {
		errWithStatusCode, errHasStatusCode := err.(v3ioerrors.ErrorWithStatusCode)
		if !errHasStatusCode {
			return "", errors.Wrap(err, "Got error without status code")
		}

		if errWithStatusCode.StatusCode() != http.StatusNotFound {
			return "", errors.Wrap(err, "Failed getting shard item")
		}

		// TODO: remove after errors.Is support added
		lh.logger.DebugWith("Could not find shard, probably doesn't exist yet", "path", shardPath)

		return "", errShardNotFound
	}

	defer response.Release()

	getItemOutput := response.Output.(*v3io.GetItemOutput)

	// return the attribute name
	location, err := getItemOutput.Item.GetFieldString(lh.getShardLocationAttributeName())
	if err != nil && err == v3ioerrors.ErrNotFound {
		return "", errShardLocationAttributeNotFound
	}

	// return the location we found
	return location, nil
}

func (lh *locationHandler) getShardLocationWithSeek(shardPath string, seekType v3io.SeekShardInputType) (string, error) {
	seekShardInput := v3io.SeekShardInput{
		Path: shardPath,
		Type: seekType,
	}

	lh.logger.DebugWith("Seeking shard", "shardPath", shardPath, "seekShardType", seekType)

	response, err := lh.streamConsumerGroup.container.SeekShardSync(&seekShardInput)
	if err != nil {
		return "", errors.Wrap(err, "Failed to seek shard")
	}
	defer response.Release()

	location := response.Output.(*v3io.SeekShardOutput).Location

	lh.logger.DebugWith("Seek shard succeeded", "shardPath", shardPath, "location", location)

	return location, nil
}

func (lh *locationHandler) getShardLocationAttributeName() string {
	return fmt.Sprintf("__%s_location", lh.streamConsumerGroup.name)
}

func (lh *locationHandler) setShardLocationInPersistency(shardID int, location string) error {
	lh.logger.DebugWith("Setting shard location in persistency", "shardID", shardID, "location", location)
	shardPath, err := lh.streamConsumerGroup.getShardPath(shardID)
	if err != nil {
		return errors.Wrapf(err, "Failed getting shard path: %v", shardID)
	}

	return lh.streamConsumerGroup.container.UpdateItemSync(&v3io.UpdateItemInput{
		Path: shardPath,
		Attributes: map[string]interface{}{
			lh.getShardLocationAttributeName(): location,
		},
	})
}

func (lh *locationHandler) markedShardLocationsCommitter(interval time.Duration, stopChan chan struct{}) {
	for {
		select {
		case <-time.After(interval):
			if err := lh.commitMarkedShardLocations(); err != nil {
				lh.logger.WarnWith("Failed committing marked shard locations", "err", errors.GetErrorStackString(err, 10))
				continue
			}
		case <-stopChan:
			lh.logger.Debug("Stopped committing marked shard locations")

			// do the last commit
			if err := lh.commitMarkedShardLocations(); err != nil {
				lh.logger.WarnWith("Failed committing marked shard locations on stop", "err", errors.GetErrorStackString(err, 10))
			}
			return
		}
	}
}

func (lh *locationHandler) commitMarkedShardLocations() error {
	var markedShardLocationsCopy []string

	// create a copy of the marked shard locations
	lh.markedShardLocationsLock.Lock()
	markedShardLocationsCopy = append(markedShardLocationsCopy, lh.markedShardLocations...)
	lh.markedShardLocationsLock.Unlock()

	// if there was no chance since last, do nothing
	if common.StringSlicesEqual(lh.lastCommittedShardLocations, markedShardLocationsCopy) {
		return nil
	}

	lh.logger.DebugWith("Committing marked shard locations", "markedShardLocationsCopy", markedShardLocationsCopy)

	var failedShardIDs []int
	for shardID, location := range markedShardLocationsCopy {

		// the location array holds a location for all partitions, indexed by their id to allow for
		// faster writes (using a rw lock) only the relevant shards ever get populated
		if location == "" {
			continue
		}

		if err := lh.setShardLocationInPersistency(shardID, location); err != nil {
			lh.logger.WarnWith("Failed committing shard location", "shardID", shardID,
				"location", location,
				"err", errors.GetErrorStackString(err, 10))

			failedShardIDs = append(failedShardIDs, shardID)
		}
	}

	if len(failedShardIDs) > 0 {
		return errors.Errorf("Failed committing marked shard locations in shards: %v", failedShardIDs)
	}

	lh.lastCommittedShardLocations = markedShardLocationsCopy

	return nil
}

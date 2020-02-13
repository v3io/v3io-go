package streamconsumergroup

import (
	"fmt"
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"
	"path"
	"strconv"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type claim struct {
	logger                   logger.Logger
	streamConsumerGroup      *streamConsumerGroup
	shardID                  int
	recordBatchChan          chan *RecordBatch
	stopRecordBatchFetchChan chan struct{}
	currentShardLocation     string
}

func newClaim(streamConsumerGroup *streamConsumerGroup, shardID int) (*claim, error) {
	return &claim{
		logger:                   streamConsumerGroup.logger.GetChild(fmt.Sprintf("claim-%d", shardID)),
		streamConsumerGroup:      streamConsumerGroup,
		shardID:                  shardID,
		recordBatchChan:          make(chan *RecordBatch, streamConsumerGroup.config.Claim.RecordBatchChanSize),
		stopRecordBatchFetchChan: make(chan struct{}),
	}, nil
}

func (c *claim) Start() error {
	c.logger.DebugWith("Starting claim")

	go c.fetchRecordBatches(c.stopRecordBatchFetchChan,
		c.streamConsumerGroup.config.Claim.RecordBatchFetch.Interval)

	// tell the consumer group handler to consume the claim
	c.logger.DebugWith("Triggering given handler ConsumeClaim")
	return c.streamConsumerGroup.handler.ConsumeClaim(c.streamConsumerGroup.session, c)
}

func (c *claim) Stop() error {
	c.logger.DebugWith("Stopping claim")
	c.stopRecordBatchFetchChan <- struct{}{}
	return nil
}

func (c *claim) GetStreamPath() string {
	return c.streamConsumerGroup.streamPath
}

func (c *claim) GetShardID() int {
	return c.shardID
}

func (c *claim) GetCurrentLocation() string {
	return c.currentShardLocation
}

func (c *claim) GetRecordBatchChan() <-chan *RecordBatch {
	return c.recordBatchChan
}

func (c *claim) fetchRecordBatches(stopChannel chan struct{}, fetchInterval time.Duration) error {
	var err error

	// read initial location. use config if error. might need to wait until shard actually exists
	c.currentShardLocation, err = c.getCurrentShardLocation(c.shardID)
	if err != nil {
		return errors.Wrap(err, "Failed to get shard location")
	}

	for {
		select {
		case <-time.After(fetchInterval):
			c.currentShardLocation, err = c.fetchRecordBatch(c.currentShardLocation)
			if err != nil {
				c.logger.WarnWith("Failed fetching record batch", "err", errors.GetErrorStackString(err, 10))
				continue
			}

		case <-stopChannel:
			return nil
		}
	}
}

func (c *claim) fetchRecordBatch(location string) (string, error) {
	c.logger.DebugWith("Fetching record batch", "location", location)

	getRecordsInput := v3io.GetRecordsInput{
		Path:     path.Join(c.streamConsumerGroup.streamPath, strconv.Itoa(c.shardID)),
		Location: location,
		Limit:    c.streamConsumerGroup.config.Claim.RecordBatchFetch.NumRecordsInBatch,
	}

	response, err := c.streamConsumerGroup.container.GetRecordsSync(&getRecordsInput)
	if err != nil {
		return "", errors.Wrapf(err, "Failed fetching record batch: %s", location)
	}

	defer response.Release()

	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	if len(getRecordsOutput.Records) == 0 {
		return getRecordsOutput.NextLocation, nil
	}

	records := make([]v3io.StreamRecord, len(getRecordsOutput.Records))

	for _, getRecordResult := range getRecordsOutput.Records {
		record := v3io.StreamRecord{
			ShardID:      &c.shardID,
			Data:         getRecordResult.Data,
			ClientInfo:   getRecordResult.ClientInfo,
			PartitionKey: getRecordResult.PartitionKey,
		}

		records = append(records, record)
	}

	recordBatch := RecordBatch{
		Records:      records,
		NextLocation: getRecordsOutput.NextLocation,
		ShardID:      c.shardID,
	}

	// write into chunks channel, blocking if there's no space
	c.recordBatchChan <- &recordBatch

	return getRecordsOutput.NextLocation, nil
}

func (c *claim) getCurrentShardLocation(shardID int) (string, error) {

	// get the location from persistency
	currentShardLocation, err := c.streamConsumerGroup.locationHandler.getShardLocationFromPersistency(shardID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get shard location")
	}

	// if shard wasn't found, try again periodically
	if err == errShardNotFound {
		for {
			select {

			// TODO: from configuration
			case <-time.After(1 * time.Second):

				// get the location from persistency
				currentShardLocation, err = c.streamConsumerGroup.locationHandler.getShardLocationFromPersistency(shardID)
				if err != nil {
					if err == errShardNotFound {

						// shard doesn't exist yet, try again
						continue
					}

					return "", errors.Wrap(err, "Failed to get shard location")
				}

				return currentShardLocation, nil
			case <-c.stopRecordBatchFetchChan:
				return "", v3ioerrors.ErrStopped
			}
		}
	}

	return currentShardLocation, nil
}

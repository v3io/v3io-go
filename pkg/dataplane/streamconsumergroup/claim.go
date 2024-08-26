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
	"context"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	v3io "github.com/v3io/v3io-go/pkg/dataplane"
	v3ioerrors "github.com/v3io/v3io-go/pkg/errors"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type claim struct {
	logger                   logger.Logger
	member                   *member
	shardID                  int
	recordBatchChan          chan *RecordBatch
	stopRecordBatchFetchChan chan struct{}
	currentShardLocation     string

	// get shard location configuration
	getShardLocationAttempts int
	getShardLocationBackoff  common.Backoff
}

func newClaim(member *member, shardID int) (*claim, error) {
	return &claim{
		logger:                   member.streamConsumerGroup.logger.GetChild(fmt.Sprintf("claim-%d", shardID)),
		member:                   member,
		shardID:                  shardID,
		recordBatchChan:          make(chan *RecordBatch, member.streamConsumerGroup.config.Claim.RecordBatchChanSize),
		stopRecordBatchFetchChan: make(chan struct{}, 1),
		getShardLocationAttempts: member.streamConsumerGroup.config.Claim.GetShardLocationRetry.Attempts,
		getShardLocationBackoff:  member.streamConsumerGroup.config.Claim.GetShardLocationRetry.Backoff,
	}, nil
}

func (c *claim) start() error {
	c.logger.DebugWith("Starting claim", "shardID", c.shardID)

	go func() {
		err := c.fetchRecordBatches(c.stopRecordBatchFetchChan,
			c.member.streamConsumerGroup.config.Claim.RecordBatchFetch.Interval)

		if err != nil {
			c.logger.ErrorWith("Failed to fetch record batches", "err", errors.GetErrorStackString(err, 10))
		}
	}()

	go func() {

		// tell the consumer group handler to consume the claim
		c.logger.DebugWith("Calling ConsumeClaim on handler")
		if err := c.member.handler.ConsumeClaim(c.member.session, c); err != nil {
			c.logger.ErrorWith("ConsumeClaim returned with error", "err", errors.GetErrorStackString(err, 10))
		}

		if err := c.stop(); err != nil {
			c.logger.ErrorWith("Failed to stop claim after consumption", "err", errors.GetErrorStackString(err, 10))
		}
	}()

	return nil
}

func (c *claim) stop() error {
	c.logger.DebugWith("Stopping claim")

	// don't block
	select {
	case c.stopRecordBatchFetchChan <- struct{}{}:
	default:
	}

	return nil
}

func (c *claim) GetStreamPath() string {
	return c.member.streamConsumerGroup.streamPath
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
	if err := common.RetryFunc(context.TODO(),
		c.logger,
		c.getShardLocationAttempts,
		nil,
		&c.getShardLocationBackoff, func(attempt int) (bool, error) {
			c.currentShardLocation, err = c.getCurrentShardLocation(c.shardID)
			if err != nil {
				if common.EngineErrorIsNonFatal(err) {
					return true, errors.Wrap(err, "Failed to get shard location due to a network error")
				}
				// requested for an immediate stop
				if err == v3ioerrors.ErrStopped {
					return false, nil
				}

				switch errors.RootCause(err).(type) {

				// in case of a network error, retry to avoid transient errors
				case *net.OpError:
					return true, errors.Wrap(err, "Failed to get shard location due to a network error")

				// unknown error, fail now
				default:
					return false, errors.Wrap(err, "Failed to get shard location")
				}
			}

			// we have shard location
			return false, nil
		}); err != nil {
		return errors.Wrapf(err,
			"Failed to get shard location state, attempts exhausted. shard id: %d",
			c.shardID)
	}

	for {
		select {
		case <-time.After(fetchInterval):
			location, err := c.fetchRecordBatch(c.currentShardLocation)
			if err != nil {
				c.logger.WarnWith("Failed fetching record batch",
					"shardId", c.shardID,
					"err", errors.GetErrorStackString(err, 10))

				// if the error is that the location we asked for is bad, try to seek earliest
				if c.checkIllegalLocationErr(err) {
					location, err := c.recoverLocationAfterIllegalLocationErr()

					if err != nil {
						c.logger.WarnWith("Failed to recover after illegal location",
							"shardId", c.shardID,
							"err", errors.GetErrorStackString(err, 10))

						continue
					}

					// set next location for next time
					c.currentShardLocation = location
				}

				continue
			}

			c.currentShardLocation = location

		case <-stopChannel:
			close(c.recordBatchChan)
			c.logger.Debug("Stopping fetch")
			return nil
		}
	}
}

func (c *claim) fetchRecordBatch(location string) (string, error) {
	getRecordsInput := v3io.GetRecordsInput{
		Path:     path.Join(c.member.streamConsumerGroup.streamPath, strconv.Itoa(c.shardID)),
		Location: location,
		Limit:    c.member.streamConsumerGroup.config.Claim.RecordBatchFetch.NumRecordsInBatch,
	}

	response, err := c.member.streamConsumerGroup.container.GetRecordsSync(&getRecordsInput)
	if err != nil {
		return "", errors.Wrapf(err, "Failed fetching record batch: %s", location)
	}

	defer response.Release()

	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	if len(getRecordsOutput.Records) == 0 {
		return getRecordsOutput.NextLocation, nil
	}

	records := make([]v3io.StreamRecord, len(getRecordsOutput.Records))

	for receivedRecordIndex, receivedRecord := range getRecordsOutput.Records {
		record := v3io.StreamRecord{
			ShardID:         &c.shardID,
			Data:            receivedRecord.Data,
			ClientInfo:      receivedRecord.ClientInfo,
			PartitionKey:    receivedRecord.PartitionKey,
			SequenceNumber:  receivedRecord.SequenceNumber,
			ArrivalTimeSec:  receivedRecord.ArrivalTimeSec,
			ArrivalTimeNSec: receivedRecord.ArrivalTimeNSec,
		}

		records[receivedRecordIndex] = record
	}

	recordBatch := RecordBatch{
		Location:     location,
		Records:      records,
		NextLocation: getRecordsOutput.NextLocation,
		ShardID:      c.shardID,
	}

	// write into chunks channel, blocking if there's no space
	c.recordBatchChan <- &recordBatch

	return getRecordsOutput.NextLocation, nil
}

func (c *claim) getCurrentShardLocation(shardID int) (string, error) {

	initialLocation := c.member.streamConsumerGroup.config.Claim.RecordBatchFetch.InitialLocation

	// get the location from persistency
	currentShardLocation, err := c.member.streamConsumerGroup.getShardLocationFromPersistency(shardID, initialLocation)
	if err != nil && errors.RootCause(err) != ErrShardNotFound {
		return "", errors.Wrap(err, "Failed to get shard location from persistency")
	}

	// if shard wasn't found, try again periodically
	if errors.RootCause(err) == ErrShardNotFound {

		// if the shard does not exist yet, turn seek-latest to seek-earliest (IG-19366)
		if initialLocation == v3io.SeekShardInputTypeLatest {
			initialLocation = v3io.SeekShardInputTypeEarliest
		}

		for {
			select {
			case <-time.After(c.member.streamConsumerGroup.config.SequenceNumber.ShardWaitInterval):

				// get the location from persistency
				currentShardLocation, err = c.member.streamConsumerGroup.getShardLocationFromPersistency(shardID, initialLocation)
				if err != nil {
					if errors.RootCause(err) == ErrShardNotFound {

						// shard doesn't exist yet, try again
						continue
					}

					return "", errors.Wrap(err, "Failed to get shard location from persistency")
				}

				return currentShardLocation, nil
			case <-c.stopRecordBatchFetchChan:
				return "", v3ioerrors.ErrStopped
			}
		}
	}

	return currentShardLocation, nil
}

func (c *claim) checkIllegalLocationErr(err error) bool {

	// sanity
	if err == nil {
		return false
	}

	// try to get the cause
	causeErr := errors.Cause(err)
	if causeErr == nil {

		// if we can't determind the cause, don't assume location (sanity)
		return false
	}

	return strings.Contains(causeErr.Error(), "IllegalLocationException")
}

func (c *claim) recoverLocationAfterIllegalLocationErr() (string, error) {
	c.logger.InfoWith("Location requested is invalid. Trying to seek to earliest",
		"shardId", c.shardID,
		"location", c.currentShardLocation)

	streamPath, err := c.member.streamConsumerGroup.getShardPath(c.shardID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to get shard path")
	}

	location, err := c.member.streamConsumerGroup.getShardLocationWithSeek(&v3io.SeekShardInput{
		Path: streamPath,
		Type: v3io.SeekShardInputTypeEarliest,
	})

	if err != nil {
		return "", errors.Wrap(err, "Failed to seek to earliest")
	}

	c.logger.InfoWith("Got new location after seeking to earliest",
		"shardId", c.shardID,
		"location", location)

	return location, nil
}

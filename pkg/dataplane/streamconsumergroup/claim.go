package streamconsumergroup

import (
	"fmt"
	"path"
	"time"

	"github.com/v3io/v3io-go/pkg/dataplane"

	"github.com/nuclio/errors"
	"github.com/nuclio/logger"
)

type streamConsumerGroupClaim struct {
	logger              logger.Logger
	streamConsumerGroup *streamConsumerGroup
	shardID             int
	member              *streamConsumerGroupMember
	recordsChannel      chan v3io.StreamRecord
	stopPollingChannel  chan bool
}

func newStreamConsumerGroupClaim(streamConsumerGroup *streamConsumerGroup,
	shardID int, member *streamConsumerGroupMember) (v3io.StreamConsumerGroupClaim, error) {

	recordsChannelSize := streamConsumerGroup.config.Claim.RecordsChannelSize

	return &streamConsumerGroupClaim{
		logger:              streamConsumerGroup.logger.GetChild(fmt.Sprintf("claim-%v", shardID)),
		streamConsumerGroup: streamConsumerGroup,
		shardID:             shardID,
		member:              member,
		recordsChannel:      make(chan v3io.StreamRecord, recordsChannelSize),
		stopPollingChannel:  make(chan bool),
	}, nil
}

func (c *streamConsumerGroupClaim) Start() error {
	pollingInterval := c.streamConsumerGroup.config.Claim.Polling.Interval
	go c.pollMessagesPeriodically(c.stopPollingChannel, pollingInterval)

	// tell the consumer group handler to consume the claim
	c.member.handler.ConsumeClaim(c)
	return nil
}

func (c *streamConsumerGroupClaim) Stop() error {
	c.stopPollingChannel <- true
	return nil
}

func (c *streamConsumerGroupClaim) pollMessagesPeriodically(stopChannel chan bool, pollingInterval time.Duration) {
	ticker := time.NewTicker(pollingInterval)

	// read initial location. use config if error
	location, err := c.streamConsumerGroup.locationHandler.GetLocation(c.shardID)
	if err != nil {
		c.logger.DebugWith("Failed getting location", "err", errors.GetErrorStackString(err, 10))
		location = ""
	}

	for {
		select {
		case <-stopChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			location, err = c.pollMessages(location)
			if err != nil {
				c.logger.WarnWith("Failed polling messages", "err", errors.GetErrorStackString(err, 10))
				continue
			}
		}
	}
}

func (c *streamConsumerGroupClaim) pollMessages(location string) (string, error) {
	if location == "" {
		inputType := c.streamConsumerGroup.config.Shard.InputType
		var err error
		location, err = c.streamConsumerGroup.seekShard(c.shardID, inputType)
		if err != nil {
			return "", errors.Wrapf(err, "Failed seeking shard: %v", c.shardID)
		}
	}

	getRecordsLimit := c.streamConsumerGroup.config.Claim.Polling.GetRecordsLimit

	getRecordsInput := v3io.GetRecordsInput{
		DataPlaneInput: c.streamConsumerGroup.dataPlaneInput,
		Path:           path.Join(c.streamConsumerGroup.streamPath, string(c.shardID)),
		Location:       location,
		Limit:          getRecordsLimit,
	}

	response, err := c.streamConsumerGroup.container.GetRecordsSync(&getRecordsInput)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting records: %s", location)
	}
	defer response.Release()

	getRecordsOutput := response.Output.(*v3io.GetRecordsOutput)

	for _, getRecordResult := range getRecordsOutput.Records {
		record := v3io.StreamRecord{
			ShardID:      &c.shardID,
			Data:         getRecordResult.Data,
			ClientInfo:   getRecordResult.ClientInfo,
			PartitionKey: getRecordResult.PartitionKey,
		}

		// write into messages channel, blocking if there's no space
		c.recordsChannel <- record
	}

	return getRecordsOutput.NextLocation, nil
}

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
	initialLocation     string
	member              *streamConsumerGroupMember
	chunksChannel       chan *v3io.StreamChunk
	stopPollingChannel  chan bool
}

func newStreamConsumerGroupClaim(streamConsumerGroup *streamConsumerGroup,
	shardID int, member *streamConsumerGroupMember) (v3io.StreamConsumerGroupClaim, error) {

	chunksChannelSize := streamConsumerGroup.config.Claim.ChunksChannelSize

	return &streamConsumerGroupClaim{
		logger:              streamConsumerGroup.logger.GetChild(fmt.Sprintf("claim-%s-%v", member.ID, shardID)),
		streamConsumerGroup: streamConsumerGroup,
		shardID:             shardID,
		member:              member,
		chunksChannel:       make(chan *v3io.StreamChunk, chunksChannelSize),
		stopPollingChannel:  make(chan bool),
	}, nil
}

func (c *streamConsumerGroupClaim) Start() error {
	c.logger.DebugWith("Starting claim", "shardID", c.shardID, "memberID", c.member.ID)
	pollingInterval := c.streamConsumerGroup.config.Claim.Polling.Interval
	go c.pollRecordsPeriodically(c.stopPollingChannel, pollingInterval)

	// tell the consumer group handler to consume the claim
	c.logger.DebugWith("Triggering given handler ConsumeClaim")
	c.member.handler.ConsumeClaim(c.member.session, c)
	c.logger.DebugWith("Handler consuming claims")
	return nil
}

func (c *streamConsumerGroupClaim) Stop() error {
	c.stopPollingChannel <- true
	return nil
}

func (c *streamConsumerGroupClaim) Stream() (string, error) {
	// TODO: maybe differentiate between stream and stream path
	return c.streamConsumerGroup.streamPath, nil
}

func (c *streamConsumerGroupClaim) Shard() (int, error) {
	return c.shardID, nil
}

func (c *streamConsumerGroupClaim) InitialLocation() (string, error) {
	if c.initialLocation == "" {
		return "", errors.New("Initial location has not been populated yet")
	}
	return c.initialLocation, nil
}

func (c *streamConsumerGroupClaim) Chunks() <-chan *v3io.StreamChunk {
	return c.chunksChannel
}

func (c *streamConsumerGroupClaim) pollRecordsPeriodically(stopChannel chan bool, pollingInterval time.Duration) {
	ticker := time.NewTicker(pollingInterval)

	// read initial location. use config if error
	location, err := c.streamConsumerGroup.locationHandler.GetLocation(c.shardID)
	if err != nil {

		// TODO: Ideally use errors.Is(ErrNotFound) and only ignore on not found - might be problematic cause we can't
		// ensure we're running on Go 1.13 or above (in which errors.Is added)
		c.logger.DebugWith("Location not found, It is ok")
		location = ""
	}

	for {
		select {
		case <-stopChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			if c.initialLocation == "" && location != "" {
				c.logger.DebugWith("Found initial location", "initialLocation", location)
				c.initialLocation = location
			}
			location, err = c.pollRecords(location)
			if err != nil {
				c.logger.WarnWith("Failed polling messages", "err", errors.GetErrorStackString(err, 10))
				continue
			}
		}
	}
}

func (c *streamConsumerGroupClaim) pollRecords(location string) (string, error) {
	if location == "" {
		inputType := c.streamConsumerGroup.config.Shard.InputType
		var err error
		location, err = c.streamConsumerGroup.seekShard(c.shardID, inputType)
		if err != nil {
			return "", errors.Wrapf(err, "Failed seeking shard: %v", c.shardID)
		}
	}

	c.logger.DebugWith("Polling records", "location", location)

	chunkSize := c.streamConsumerGroup.config.Claim.Polling.ChunkSize

	getRecordsInput := v3io.GetRecordsInput{
		DataPlaneInput: c.streamConsumerGroup.dataPlaneInput,
		Path:           path.Join(c.streamConsumerGroup.streamPath, string(c.shardID)),
		Location:       location,
		Limit:          chunkSize,
	}

	response, err := c.streamConsumerGroup.container.GetRecordsSync(&getRecordsInput)
	if err != nil {
		return "", errors.Wrapf(err, "Failed getting records: %s", location)
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

	chunk := v3io.StreamChunk{
		Records:      records,
		NextLocation: getRecordsOutput.NextLocation,
		ShardID:      c.shardID,
	}

	// write into chunks channel, blocking if there's no space
	c.chunksChannel <- &chunk

	return getRecordsOutput.NextLocation, nil
}

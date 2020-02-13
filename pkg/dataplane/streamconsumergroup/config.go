package streamconsumergroup

import (
	"time"

	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
)

type Config struct {
	Session struct {
		Timeout time.Duration
	}
	State struct {
		ModifyRetry struct {
			Attempts int
			Backoff  common.Backoff
		}
		Heartbeat struct {
			Interval time.Duration
		}
	}
	Location struct {
		CommitCache struct {
			Interval time.Duration
		}
	}
	Claim struct {
		RecordBatchChanSize int
		RecordBatchFetch    struct {
			Interval          time.Duration
			NumRecordsInBatch int
			InitialLocation   v3io.SeekShardInputType
		}
	}
}

// NewConfig returns a new configuration instance with sane defaults.
func NewConfig() *Config {
	c := &Config{}
	c.Session.Timeout = 30 * time.Second
	c.State.ModifyRetry.Attempts = 5
	c.State.ModifyRetry.Backoff = common.Backoff{
		Min: 50 * time.Millisecond,
		Max: 1 * time.Second,
	}
	c.State.Heartbeat.Interval = 3 * time.Second
	c.Location.CommitCache.Interval = 10 * time.Second
	c.Claim.RecordBatchChanSize = 100
	c.Claim.RecordBatchFetch.Interval = 3 * time.Second
	c.Claim.RecordBatchFetch.NumRecordsInBatch = 5
	c.Claim.RecordBatchFetch.InitialLocation = v3io.SeekShardInputTypeEarliest

	return c
}

package streamconsumergroup

import (
	"github.com/v3io/v3io-go/pkg/common"
	"github.com/v3io/v3io-go/pkg/dataplane"
	"time"
)

type Config struct {
	Shard struct {
		InputType v3io.SeekShardInputType
	}
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
		RecordsChannelSize int
		Polling            struct {
			Interval        time.Duration
			GetRecordsLimit int
		}
	}
}

// NewConfig returns a new configuration instance with sane defaults.
func NewConfig() *Config {
	c := &Config{}
	c.Shard.InputType = v3io.SeekShardInputTypeEarliest
	c.Session.Timeout = 30 * time.Second
	c.State.ModifyRetry.Attempts = 5
	c.State.ModifyRetry.Backoff = common.Backoff{
		Min: 50 * time.Millisecond,
		Max: 1 * time.Second,
	}
	c.State.Heartbeat.Interval = 5 * time.Second
	c.Location.CommitCache.Interval = 10 * time.Second
	c.Claim.RecordsChannelSize = 100
	c.Claim.Polling.Interval = 1 * time.Second
	c.Claim.Polling.GetRecordsLimit = 20

	return c
}

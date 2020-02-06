package streamconsumergroup

import "time"

type LocationHandler interface {
	Start() error
	Stop() error
	GetLocation(shardID int) (string, error)
	MarkLocation(shardID int, location string) error
}

type StateHandler interface {
	Start() error
	GetState() (*State, error)
	GetMemberState(string) (*SessionState, error)
}

type State struct {
	SchemasVersion string         `json:"schemaVersion"`
	Sessions       []SessionState `json:"sessions"`
}

type SessionState struct {
	MemberID      string     `json:"memberId"`
	LastHeartbeat *time.Time `json:"lastHeartbeat"`
	Shards        []int      `json:"shards"`
}

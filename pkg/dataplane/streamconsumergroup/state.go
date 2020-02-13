package streamconsumergroup

type State struct {
	SchemasVersion string          `json:"schema_version"`
	SessionStates  []*SessionState `json:"session_states"`
}

func newState() (*State, error) {
	return &State{
		SchemasVersion: "0.0.1",
		SessionStates:  []*SessionState{},
	}, nil
}

func (s *State) findSessionStateByMemberID(memberID string) *SessionState {
	for _, sessionState := range s.SessionStates {
		if sessionState.MemberID == memberID {
			return sessionState
		}
	}

	return nil
}

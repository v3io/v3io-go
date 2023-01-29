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
	"encoding/json"
)

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

func (s *State) String() string {
	marshalledState, err := json.Marshal(s)
	if err != nil {
		return err.Error()
	}

	return string(marshalledState)
}

func (s *State) deepCopy() *State {
	stateCopy := State{}
	stateCopy.SchemasVersion = s.SchemasVersion
	for _, stateSession := range s.SessionStates {
		stateSessionCopy := stateSession
		stateCopy.SessionStates = append(stateCopy.SessionStates, stateSessionCopy)
	}

	return &stateCopy
}

func (s *State) findSessionStateByMemberID(memberID string) *SessionState {
	for _, sessionState := range s.SessionStates {
		if sessionState.MemberID == memberID {
			return sessionState
		}
	}

	return nil
}

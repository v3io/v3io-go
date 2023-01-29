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

package registry

import (
	"fmt"
	"sync"
)

type Registry struct {
	className  string
	Lock       sync.Locker
	Registered map[string]interface{}
}

func NewRegistry(className string) *Registry {
	return &Registry{
		className:  className,
		Lock:       &sync.Mutex{},
		Registered: map[string]interface{}{},
	}
}

func (r *Registry) Register(kind string, registeree interface{}) {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	_, found := r.Registered[kind]
	if found {

		// registries register things on package initialization; no place for error handling
		panic(fmt.Sprintf("Already registered: %s", kind))
	}

	r.Registered[kind] = registeree
}

func (r *Registry) Get(kind string) (interface{}, error) {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	registree, found := r.Registered[kind]
	if !found {

		// registries register things on package initialization; no place for error handling
		return nil, fmt.Errorf("Registry for %s failed to find: %s", r.className, kind)
	}

	return registree, nil
}

func (r *Registry) GetKinds() []string {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	keys := make([]string, 0, len(r.Registered))

	for key := range r.Registered {
		keys = append(keys, key)
	}

	return keys
}

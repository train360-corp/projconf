/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package state

import (
	"go.uber.org/zap"
	"sync"
)

type State struct {
	postgres  bool
	postgrest bool
	mutex     sync.Mutex
	anonKey   string
	logger    *zap.SugaredLogger
}

var state *State
var once sync.Once

func Get() *State {
	once.Do(func() {
		state = &State{
			postgres:  false,
			postgrest: false,
		}
	})
	return state
}

func (s *State) SetLogger(logger *zap.SugaredLogger) {
	s.logger = logger
}

func (s *State) GetLogger() *zap.SugaredLogger {
	return s.logger
}

func (s *State) IsAlive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.postgres && s.postgrest
}

func (s *State) IsPostgresAlive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.postgres
}

func (s *State) SetPostgresAlive(alive bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.postgres = alive
}

func (s *State) IsPostgrestAlive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.postgrest
}

func (s *State) SetPostgrestAlive(alive bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.postgrest = alive
}

func (s *State) GetAnonymousKey() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.anonKey
}

func (s *State) SetAnonymousKey(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.anonKey = key
}

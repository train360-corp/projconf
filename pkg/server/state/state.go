/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package state

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/train360-corp/projconf/internal/utils"
	"sync"
	"time"
)

type State struct {
	postgres  bool
	postgrest bool
	jwtSecret string
	anonKey   string
	mutex     sync.Mutex
}

var state *State
var once sync.Once

func Get() *State {
	once.Do(func() {

		jwtSecret := utils.RandomString(32)
		iat := time.Now().Unix()
		exp := time.Now().AddDate(10, 0, 0).Unix()

		pubKey, pubKeyErr := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(map[string]any{
			"role": "anon",
			"iss":  "supabase",
			"iat":  iat,
			"exp":  exp,
		})).SignedString([]byte(jwtSecret))
		if pubKeyErr != nil {
			panic(fmt.Sprintf("failed to generate runtime environment - public key error: %s", pubKeyErr.Error()))
		}

		state = &State{
			postgres:  false,
			postgrest: false,
			jwtSecret: jwtSecret,
			anonKey:   pubKey,
		}
	})
	return state
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

func (s *State) JwtSecret() string {
	return s.jwtSecret
}

func (s *State) AnonKey() string {
	return s.anonKey
}

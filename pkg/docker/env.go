/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package docker

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/train360-corp/projconf/internal/utils"
	"time"
)

type Env struct {
	PGPASSWORD             string
	JWT_SECRET             string
	PROJCONF_ADMIN_API_KEY string
	SUPABASE_PUBLIC_KEY    string
}

func NewEnv(PGPASSWORD string, PROJCONF_ADMIN_API_KEY string) (*Env, error) {

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
		return nil, fmt.Errorf("failed to generate runtime environment - public key error: %s", pubKeyErr.Error())
	}

	return &Env{
		PGPASSWORD:             PGPASSWORD,
		JWT_SECRET:             jwtSecret,
		PROJCONF_ADMIN_API_KEY: PROJCONF_ADMIN_API_KEY,
		SUPABASE_PUBLIC_KEY:    pubKey,
	}, nil
}

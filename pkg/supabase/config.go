/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package supabase

type AuthConfig struct {
	Id          string
	Secret      string
	AdminAPIKey string
}

type Config struct {
	Url     string
	AnonKey string
}

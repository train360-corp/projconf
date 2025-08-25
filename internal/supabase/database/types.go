/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

//go:generate sh -c "supabase gen types --lang=go --local > types.gen.go"
package database

type PublicRpcCreateClientAndSecretRequest struct {
	Display       string `json:"p_display"`
	EnvironmentId string `json:"p_env_id"`
}

type PublicRpcCreateClientAndSecretResponse struct {
	ClientId string `json:"client_id"`
	SecretId string `json:"secret_id"`
	Secret   string `json:"secret"`
}

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package supabase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/pkg/consts"
	"github.com/train360-corp/projconf/pkg/supabase/database"
	"io"
	"net/http"
	"strings"
)

func post[T any](c *Client, table string, requestData any) (*T, error) {

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.Url, "/"), strings.TrimPrefix(table, "/")), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.AnonKey))
	req.Header.Add("apikey", c.config.AnonKey)
	req.Header.Add("Prefer", "return=representation")
	req.Header.Add("Accept", "application/vnd.pgrst.object+json")

	if c.auth.AdminAPIKey != "" {
		req.Header.Add(consts.X_ADMIN_API_KEY, c.auth.AdminAPIKey)
	} else {
		req.Header.Add(consts.X_CLIENT_SECRET_ID, c.auth.Id)
		req.Header.Add(consts.X_CLIENT_SECRET, c.auth.Secret)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		var out T
		if err := json.Unmarshal(responseBody, &out); err != nil {
			return nil, fmt.Errorf("decode %s: %w", table, err)
		}
		return &out, nil
	} else { // handle bad return code
		return nil, errors.New(string(responseBody))
	}
}

func (c *Client) PostProject(row database.PublicProjectsInsert) (*database.PublicProjectsSelect, error) {
	return post[database.PublicProjectsSelect](c, "projects", row)
}

func (c *Client) PostEnvironment(row database.PublicEnvironmentsInsert) (*database.PublicEnvironmentsSelect, error) {
	return post[database.PublicEnvironmentsSelect](c, "environments", row)
}

func (c *Client) PostVariable(row database.PublicVariablesInsert) (*database.PublicVariablesSelect, error) {
	return post[database.PublicVariablesSelect](c, "variables", row)
}

func (c *Client) CreateClient(req database.PublicRpcCreateClientAndSecretRequest) (*database.PublicRpcCreateClientAndSecretResponse, error) {
	return post[database.PublicRpcCreateClientAndSecretResponse](c, "rpc/create_client_and_secret", req)
}

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package supabase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/train360-corp/projconf/internal/consts"
	"github.com/train360-corp/projconf/internal/supabase/database"
	"io"
	"net/http"
	"strings"
)

func (c *Client) get(config *request) (*http.Response, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.Url, "/"), strings.TrimPrefix(config.endpoint, "/")), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.AnonKey))
	req.Header.Add("apikey", c.config.AnonKey)

	if c.auth.AdminAPIKey != "" {
		req.Header.Add(consts.X_ADMIN_API_KEY, c.auth.AdminAPIKey)
	} else {
		req.Header.Add(consts.X_CLIENT_SECRET_ID, c.auth.Id)
		req.Header.Add(consts.X_CLIENT_SECRET, c.auth.Secret)
	}

	if config.single {
		req.Header.Add("Accept", "application/vnd.pgrst.object+json")
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Generic list fetcher
func getList[T any](client *Client, table string, filter string) (*[]T, error) {

	endpoint := fmt.Sprintf("/rest/v1/%s", table)
	if filter != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, filter)
	}

	res, err := client.get(&request{
		endpoint: endpoint,
		single:   false,
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		// include status + a snippet of body to help debugging
		return nil, fmt.Errorf("GET %s: status %d: %s", table, res.StatusCode, string(body))
	}

	var out []T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", table, err)
	}
	return &out, nil
}

func (c *Client) GetEnvironments(projectID *uuid.UUID) (*[]database.PublicEnvironmentsSelect, error) {

	var filter string
	if projectID != nil {
		filter = fmt.Sprintf("project_id=eq.%s", projectID.String())
	}

	return getList[database.PublicEnvironmentsSelect](c, "environments", filter)
}

func (c *Client) GetProjects() (*[]database.PublicProjectsSelect, error) {
	return getList[database.PublicProjectsSelect](c, "projects", "")
}

func (c *Client) GetSelf() (*database.PublicClientsSelect, error) {

	if c.self == nil {
		res, err := c.get(&request{
			endpoint: "/rest/v1/clients",
			single:   true,
		})
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			//log.Println(string(body))
			return nil, errors.New("unable to find client")
		}

		var clientObj database.PublicClientsSelect
		if err := json.Unmarshal(body, &clientObj); err != nil {
			return nil, err
		}

		c.self = &clientObj
	}

	if c.self == nil {
		return nil, errors.New("c.self unexpectedly nil")
	}

	return c.self, nil
}

type GetSecretsVariable struct {
	Key string `json:"key"`
}

type GetSecretsSecret struct {
	Value     string             `json:"value"`
	Variables GetSecretsVariable `json:"variables"`
}

func (c *Client) GetSecrets(projectId string, environmentId string) ([]GetSecretsSecret, error) {

	endpoint := fmt.Sprintf("/rest/v1/secrets?select=value%%2Cvariables(key)&project_id=eq.%s&environment_id=eq.%s", projectId, environmentId)
	res, err := c.get(&request{
		endpoint: endpoint,
		single:   false,
	})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println(string(body))
		return nil, errors.New("unable to load secrets")
	}

	var secrets []GetSecretsSecret
	if err := json.Unmarshal(body, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

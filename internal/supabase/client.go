package supabase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/supabase/database"
	"io"
	"net/http"
)

type Client struct {
	config *Config
	auth   *AuthConfig
	self   *database.PublicClientsSelect
}

func Get(config *Config) *Client {
	client := &Client{config: config}
	return client
}

func GetWithAuth(config *Config, authConfig *AuthConfig) *Client {
	client := Get(config)
	client.SetAuth(authConfig)
	return client
}

func (c *Client) SetAuth(config *AuthConfig) {
	c.auth = config
}

func (c *Client) GetSelf() (*database.PublicClientsSelect, error) {

	if c.self == nil {
		url := fmt.Sprintf("%s/rest/v1/clients", c.config.Url)

		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Accept", "application/vnd.pgrst.object+json")
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.AnonKey))

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
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

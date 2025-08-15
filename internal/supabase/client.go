package supabase

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/supabase/database"
	"io"
	"log"
	"net/http"
	"strings"
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

func GetFromContext(ctx *gin.Context) *Client {

	appCfg, err := config.Load()
	if err != nil {
		panic(errors.New(fmt.Sprintf("read config error: %v", err)))
	}

	return GetWithAuth(&Config{
		Url:     appCfg.Supabase.Url,
		AnonKey: appCfg.Supabase.Keys.Public,
	}, &AuthConfig{
		Id:     ctx.GetHeader("x-client-secret-id"),
		Secret: ctx.GetHeader("x-client-secret"),
	})
}

func (c *Client) SetAuth(config *AuthConfig) {
	c.auth = config
}

type requestConfig struct {
	endpoint string
	single   bool
}

func (c *Client) request(config *requestConfig) (*http.Response, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", strings.TrimSuffix(c.config.Url, "/"), strings.TrimPrefix(config.endpoint, "/")), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.config.AnonKey))
	req.Header.Add("apikey", c.config.AnonKey)
	req.Header.Add("x-client-secret-id", c.auth.Id)
	req.Header.Add("x-client-secret", c.auth.Secret)
	if config.single {
		req.Header.Add("Accept", "application/vnd.pgrst.object+json")
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) GetSelf() (*database.PublicClientsSelect, error) {

	if c.self == nil {
		res, err := c.request(&requestConfig{
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
			log.Println(string(body))
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
	res, err := c.request(&requestConfig{
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

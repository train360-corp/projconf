package supabase

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	credentials *models.Credentials
	token       *AuthResponse
}

func GetClient() (*Client, error) {

	secret, err := secrets.GetSecret()
	if err != nil {
		return nil, err
	}

	return &Client{
		credentials: secret,
		token:       nil,
	}, nil
}

// Authenticate authenticates the Supabase client.
func (c *Client) Authenticate() (bool, error) {

	// Prepare the JSON payload
	payload := map[string]string{
		"email":    c.credentials.Email,
		"password": c.credentials.Password,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, NewClientError("error creating authentication payload", err)
	}

	// Open HTTP connection
	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", c.getBaseURL())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, NewClientError("error creating request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.credentials.Supabase.Keys.Anon)
	req.Header.Set("apikey", c.credentials.Supabase.Keys.Anon)
	req.Header.Set("X-JWT-AUD", string(c.credentials.Type))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, NewClientError("error during authentication request", err)
	}
	defer resp.Body.Close()

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, NewClientError("error reading response body", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.Unmarshal(body, &c.token); err != nil {
			return false, NewClientError("error parsing authentication response", err)
		}
		return true, nil
	} else {
		return false, NewClientError("error response during supabase authentication", errors.New(string(body)))
	}
}

// Close invalidates the local session.
func (c *Client) Close() error {
	url := fmt.Sprintf("%s/auth/v1/logout", c.getBaseURL())
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return NewClientError("error creating logout request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.credentials.Supabase.Keys.Anon)
	req.Header.Set("apikey", c.credentials.Supabase.Keys.Anon)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return NewClientError("error during logout request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return NewClientError("logout failed", errors.New(fmt.Sprintf("unexpected status code: %d", resp.StatusCode)))
}

func (c *Client) IsAuthenticated() bool {
	return c.token != nil
}

func (c *Client) GetUser() (*User, error) {
	if !c.IsAuthenticated() {
		return nil, errors.New("supabase client not authenticated")
	}

	return &c.token.User, nil
}

func (c *Client) execute(endpoint string, output interface{}, extraHeaders map[string]string, extraBody map[string]string) error {

	var body io.Reader
	if extraBody != nil {
		jsonData, err := json.Marshal(extraBody)
		if err != nil {
			return fmt.Errorf("failed to marshal extraBody: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	method := "GET"
	if body != nil {
		method = "POST"
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("apikey", c.credentials.Supabase.Keys.Anon)
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	if extraHeaders != nil {
		for k, v := range extraHeaders {
			req.Header.Set(k, v)
		}
	}

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		if strings.Contains(string(body), "The result contains 0 rows") {
			// no rows found
			return nil
		}

		return fmt.Errorf("failed to retrieve rows (HTTP response code: %d): %s", resp.StatusCode, string(body))
	}

	// Read the entire response bodyData into a buffer
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response bodyData: %w", err)
	}

	// Decode the bodyData into the output object
	if err := json.Unmarshal(bodyBytes, &output); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *Client) RPC(name string, params map[string]string, output interface{}) error {
	if !c.IsAuthenticated() {
		return errors.New("supabase client not authenticated")
	}
	endpoint := fmt.Sprintf("%s/rest/v1/rpc/%v", c.getBaseURL(), name)
	return c.execute(endpoint, output, map[string]string{
		"Content-Type": "application/json",
	}, params)
}

func (c *Client) GetById(tableName string, id string, output interface{}) error {
	if !c.IsAuthenticated() {
		return errors.New("supabase client not authenticated")
	}
	endpoint := fmt.Sprintf("%s/rest/v1/%s?select=*&id=eq.%v", c.getBaseURL(), tableName, id)
	return c.execute(endpoint, output, map[string]string{
		"Accept": "application/vnd.pgrst.object+json",
	}, nil)
}

// Get retrieves rows from the specified table endpoint and maps them to the provided type
func (c *Client) Get(tableName string, output interface{}) error {
	if !c.IsAuthenticated() {
		return errors.New("supabase client not authenticated")
	}
	endpoint := fmt.Sprintf("%s/rest/v1/%s", c.getBaseURL(), tableName)
	return c.execute(endpoint, output, nil, nil)
}

// getBaseURL returns the base URL of the Supabase client.
func (c *Client) getBaseURL() string {
	return strings.TrimSuffix(c.credentials.Supabase.Url, "/")
}

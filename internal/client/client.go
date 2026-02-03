package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsNotFoundError checks if an error is a 404 Not Found error
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsNotFound()
	}
	return false
}

// Client is the API client for Environment Server
type Client struct {
	BaseURL          string
	HTTPClient       *http.Client
	Token            string
	DefaultProjectID string
	DefaultEnvID     string
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Endpoint         string
	Token            string
	APIKey           string
	Username         string
	Password         string
	OktaToken        string
	Timeout          int
	DefaultProjectID string
	DefaultEnvID     string
}

// NewClient creates a new Environment Server API client
func NewClient(config *AuthConfig) (*Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	client := &Client{
		BaseURL: config.Endpoint,
		HTTPClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		DefaultProjectID: config.DefaultProjectID,
		DefaultEnvID:     config.DefaultEnvID,
	}

	// Authenticate based on provided credentials (token takes highest priority)
	if config.Token != "" {
		// Use pre-authenticated token directly
		client.Token = config.Token
	} else if config.APIKey != "" {
		client.Token = config.APIKey
	} else if config.Username != "" && config.Password != "" {
		token, err := client.loginWithCredentials(config.Username, config.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
		client.Token = token
	} else if config.OktaToken != "" {
		token, err := client.loginWithOkta(config.OktaToken)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with Okta: %w", err)
		}
		client.Token = token
	} else {
		return nil, fmt.Errorf("no authentication method provided")
	}

	return client, nil
}
// loginWithCredentials authenticates using username and password via OAuth2 form data
// The endpoint expects application/x-www-form-urlencoded format as per OAuth2 spec
func (c *Client) loginWithCredentials(username, password string) (string, error) {
	// Prepare form data for OAuth2PasswordRequestForm
	formData := url.Values{}
	formData.Set("username", username)
	formData.Set("password", password)

	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		c.BaseURL+"/api/login/token",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// OAuth2 requires application/x-www-form-urlencoded
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.AccessToken, nil
}

func (c *Client) loginWithOkta(oktaToken string) (string, error) {
	data := map[string]string{
		"okta_token": oktaToken,
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	err := c.doRequest(context.Background(), "POST", "/api/login/okta-login", data, &result, false)
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, useAuth bool) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if useAuth && c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, "GET", path, nil, result, true)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, "POST", path, body, result, true)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, "PATCH", path, body, result, true)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doRequest(ctx, "DELETE", path, nil, nil, true)
}
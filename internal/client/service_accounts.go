package client

import (
	"context"
	"fmt"
	"time"
)

// FlexibleTime is a custom time type that can parse multiple datetime formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for FlexibleTime
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	// Remove quotes
	if len(s) > 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Try multiple datetime formats
	formats := []string{
		time.RFC3339,                 // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05",        // Without timezone
		"2006-01-02T15:04:05.999999", // With microseconds, no timezone
		time.RFC3339Nano,             // With nanoseconds and timezone
	}

	var err error
	for _, format := range formats {
		ft.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to parse time %q: %w", s, err)
}

// ServiceAccount represents an Environment Server service account
type ServiceAccount struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	CreatedTime FlexibleTime `json:"created_time"`
	CreatedBy   string       `json:"created_by"`
	Policies    []Policy     `json:"policies,omitempty"`
}

// ServiceAccountAPIKey represents an API key for a service account
type ServiceAccountAPIKey struct {
	KeyID     string `json:"key_id"`
	FullKey   string `json:"full_key,omitempty"` // Only available on creation
	LastLogin string `json:"last_login,omitempty"`
}

// CreateServiceAccountRequest represents the request to create a service account
type CreateServiceAccountRequest struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
}

// PatchServiceAccountRequest represents the request to update a service account
type PatchServiceAccountRequest struct {
	Description string `json:"description,omitempty"`
}

// CreateServiceAccount creates a new service account
func (c *Client) CreateServiceAccount(ctx context.Context, req *CreateServiceAccountRequest) (*ServiceAccount, error) {
	var result ServiceAccount
	err := c.Post(ctx, "/api/auth/service_account/create", req, &result)
	return &result, err
}

// GetServiceAccount retrieves a service account by ID
func (c *Client) GetServiceAccount(ctx context.Context, id string) (*ServiceAccount, error) {
	var result ServiceAccount
	err := c.Get(ctx, fmt.Sprintf("/api/auth/service_account/get/%s", id), &result)

	return &result, err
}

// UpdateServiceAccount updates an existing service account
func (c *Client) UpdateServiceAccount(ctx context.Context, id string, req *PatchServiceAccountRequest) (*ServiceAccount, error) {
	var result ServiceAccount
	err := c.Patch(ctx, fmt.Sprintf("/api/auth/service_account/%s", id), req, &result)
	return &result, err
}

// DeleteServiceAccount deletes a service account
func (c *Client) DeleteServiceAccount(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("/api/auth/service_account/%s", id))
	
	return err
}

// CreateServiceAccountAPIKey creates a new API key for a service account
func (c *Client) CreateServiceAccountAPIKey(ctx context.Context, serviceAccountID string) (*ServiceAccountAPIKey, error) {
	var result ServiceAccountAPIKey
	err := c.Post(ctx, fmt.Sprintf("/api/auth/service_account/api_key/%s", serviceAccountID), map[string]interface{}{}, &result)

	return &result, err
}

// GetServiceAccountAPIKey retrieves an API key by key ID
// Returns an APIError with 404 status if the key is not found
func (c *Client) GetServiceAccountAPIKey(ctx context.Context, keyID string) (*ServiceAccountAPIKey, error) {
	var result ServiceAccountAPIKey
	err := c.Get(ctx, fmt.Sprintf("/api/auth/service_account/api_key/%s", keyID), &result)
	return &result, err
}

// DeleteServiceAccountAPIKey deletes an API key by key ID
func (c *Client) DeleteServiceAccountAPIKey(ctx context.Context, keyID string) error {
	return c.Delete(ctx, fmt.Sprintf("/api/auth/service_account/api_key/%s", keyID))
}
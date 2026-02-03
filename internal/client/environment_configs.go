package client

import (
	"context"
	"fmt"
)

// EnvironmentConfig represents a single environment configuration key-value pair
type EnvironmentConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetEnvironmentConfig retrieves a single environment configuration by key
func (c *Client) GetEnvironmentConfig(ctx context.Context, projectID, envID, key string) (*EnvironmentConfig, error) {
	path := fmt.Sprintf("/api/environments/%s/%s/configs/get/%s", projectID, envID, key)

	var result EnvironmentConfig
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetEnvironmentConfigs retrieves all environment configurations
func (c *Client) GetEnvironmentConfigs(ctx context.Context, projectID, envID string) ([]EnvironmentConfig, error) {
	path := fmt.Sprintf("/api/environments/%s/%s/configs/get", projectID, envID)

	var result []EnvironmentConfig
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

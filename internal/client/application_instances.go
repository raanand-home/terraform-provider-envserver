package client

import (
	"context"
	"fmt"
	"time"
)

// ApplicationInstance represents an application instance in an environment
type ApplicationInstance struct {
	AppID      string            `json:"app_id"`
	InstanceID string            `json:"instance_id"`
	Tags       map[string]string `json:"tags,omitempty"`
	CreatedAt  *time.Time        `json:"created_at,omitempty"`
	UpdatedAt  *time.Time        `json:"updated_at,omitempty"`
}

// ApplicationInstanceCreate represents the request body for creating an application instance
type ApplicationInstanceCreate struct {
	AppID      string            `json:"app_id"`
	InstanceID string            `json:"instance_id"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// ApplicationInstancePatch represents the request body for updating an application instance
type ApplicationInstancePatch struct {
	Tags map[string]string `json:"tags"`
}

// CreateApplicationInstance creates a new application instance in an environment
func (c *Client) CreateApplicationInstance(ctx context.Context, projectID, envID string, instance *ApplicationInstanceCreate) (*ApplicationInstance, error) {
	var result ApplicationInstance
	path := fmt.Sprintf("/api/environments/%s/%s/instances/add", projectID, envID)
	err := c.Post(ctx, path, instance, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create application instance: %w", err)
	}
	return &result, nil
}

// GetApplicationInstance retrieves an application instance by app_id and instance_id
func (c *Client) GetApplicationInstance(ctx context.Context, projectID, envID, appID, instanceID string) (*ApplicationInstance, error) {
	var result ApplicationInstance
	path := fmt.Sprintf("/api/environments/%s/%s/instances/%s/%s", projectID, envID, appID, instanceID)
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get application instance: %w", err)
	}
	return &result, nil
}

// UpdateApplicationInstance updates an existing application instance
func (c *Client) UpdateApplicationInstance(ctx context.Context, projectID, envID, appID, instanceID string, patch *ApplicationInstancePatch) (*ApplicationInstance, error) {
	var result ApplicationInstance
	path := fmt.Sprintf("/api/environments/%s/%s/instances/%s/%s", projectID, envID, appID, instanceID)
	err := c.Patch(ctx, path, patch, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update application instance: %w", err)
	}
	return &result, nil
}

// DeleteApplicationInstance deletes an application instance
func (c *Client) DeleteApplicationInstance(ctx context.Context, projectID, envID, appID, instanceID string) error {
	path := fmt.Sprintf("/api/environments/%s/%s/instances/%s/%s", projectID, envID, appID, instanceID)
	err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete application instance: %w", err)
	}
	return nil
}

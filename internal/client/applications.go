package client

import (
	"context"
	"fmt"
)

// Application represents an application in the system
// Based on AppsPublic schema from OpenAPI spec
type Application struct {
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags,omitempty"`
}

// GetVersionRequest represents the request body for getting an application version
type GetVersionRequest struct {
	RefID     string `json:"ref_id"`
	VersionID string `json:"version_id"`
}

// ApplicationVersion represents a version of an application
// Based on PublicApplicationVersion schema from OpenAPI spec
type ApplicationVersion struct {
	ID          string            `json:"id"`
	Ref         string            `json:"ref"`
	VersionData map[string]string `json:"version_data"`
	CreatedAt   FlexibleTime      `json:"created_at"`
	Operations  []OperationPublic `json:"operations,omitempty"`
}

// OperationPublic represents an operation in the version
type OperationPublic struct {
	Key  string                 `json:"key"`
	Data map[string]interface{} `json:"data"`
	Tags map[string]string      `json:"tags"`
}

// ApplicationCreate represents the request body for creating an application
// Based on AppsCreate schema from OpenAPI spec
type ApplicationCreate struct {
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags,omitempty"`
}

// ApplicationPatch represents the request body for updating an application
// Based on ApplicationPatch schema from OpenAPI spec
type ApplicationPatch struct {
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateApplication creates a new application
func (c *Client) CreateApplication(ctx context.Context, app *ApplicationCreate) (*Application, error) {
	var result Application
	path := "/api/apps/create"
	err := c.Post(ctx, path, app, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return &result, nil
}

// GetApplication retrieves an application by app_id
func (c *Client) GetApplication(ctx context.Context, appID string) (*Application, error) {
	var result Application
	path := fmt.Sprintf("/api/apps/get/%s", appID)
	err := c.Get(ctx, path, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	return &result, nil
}

// UpdateApplication updates an existing application
func (c *Client) UpdateApplication(ctx context.Context, appID string, patch *ApplicationPatch) (*Application, error) {
	var result Application
	path := fmt.Sprintf("/api/apps/%s", appID)
	err := c.Patch(ctx, path, patch, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}
	return &result, nil
}

// DeleteApplication deletes an application
func (c *Client) DeleteApplication(ctx context.Context, appID string) error {
	path := fmt.Sprintf("/api/apps/%s", appID)
	err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	return nil
}

// GetApplicationVersion retrieves a specific version of an application
func (c *Client) GetApplicationVersion(ctx context.Context, appID string, refID string, versionID string) (*ApplicationVersion, error) {
	var result ApplicationVersion
	path := fmt.Sprintf("/api/apps/get_version/%s", appID)
	req := &GetVersionRequest{
		RefID:     refID,
		VersionID: versionID,
	}
	err := c.Post(ctx, path, req, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get application version: %w", err)
	}
	return &result, nil
}

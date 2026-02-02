package client

import (
	"context"
	"fmt"
)

// Project represents an Environment Server project
type Project struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, project *Project) (*Project, error) {
	var result Project
	err := c.Post(ctx, "/api/projects/create", project, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return &result, nil
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var result Project
	err := c.Get(ctx, fmt.Sprintf("/api/projects/get/%s", id), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return &result, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, id string, project *Project) (*Project, error) {
	var result Project
	err := c.Patch(ctx, fmt.Sprintf("/api/projects/%s", id), project, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	return &result, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("/api/projects/%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// ListProjects lists all projects
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var result []Project
	err := c.Get(ctx, "/api/projects/list", &result)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return result, nil
}
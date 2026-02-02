package client

import (
	"context"
	"fmt"
)

// Policy represents an Environment Server access policy
type Policy struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Policy      string `json:"policy"` // YAML or JSON policy document
	Managed     bool   `json:"managed"`
}

// CreatePolicyRequest represents the request to create a policy
type CreatePolicyRequest struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Policy      string `json:"policy"`
}

// PatchPolicyRequest represents the request to update a policy
type PatchPolicyRequest struct {
	Description *string `json:"description,omitempty"`
	Policy      *string `json:"policy,omitempty"`
}

// PolicyAttachment represents a policy attachment to a user or service account
type PolicyAttachment struct {
	UserEmail          string `json:"email,omitempty"`
	ServiceAccountID   string `json:"service_account_id,omitempty"`
	PolicyID           string `json:"policy_id"`
}

// CreatePolicy creates a new policy
func (c *Client) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
	var result Policy
	err := c.Post(ctx, "/api/auth/policies/create", req, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}
	return &result, nil
}

// GetPolicy retrieves a policy by ID
func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	var result Policy
	err := c.Get(ctx, fmt.Sprintf("/api/auth/policies/get/%s", id), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePolicy updates an existing policy
func (c *Client) UpdatePolicy(ctx context.Context, id string, req *PatchPolicyRequest) (*Policy, error) {
	var result Policy
	err := c.Patch(ctx, fmt.Sprintf("/api/auth/policies/%s", id), req, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}
	return &result, nil
}

// DeletePolicy deletes a policy
func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	err := c.Delete(ctx, fmt.Sprintf("/api/auth/policies/%s", id))
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	return nil
}

// AttachPolicyToUser attaches a policy to a user
func (c *Client) AttachPolicyToUser(ctx context.Context, userEmail, policyID string) error {
	req := map[string]string{
		"email":     userEmail,
		"policy_id": policyID,
	}
	err := c.Post(ctx, "/api/auth/user/attach_policy", req, nil)
	if err != nil {
		return fmt.Errorf("failed to attach policy to user: %w", err)
	}
	return nil
}

// DetachPolicyFromUser detaches a policy from a user
func (c *Client) DetachPolicyFromUser(ctx context.Context, userEmail, policyID string) error {
	req := map[string]string{
		"email":     userEmail,
		"policy_id": policyID,
	}
	err := c.Post(ctx, "/api/auth/user/deatach_policy", req, nil)
	if err != nil {
		return fmt.Errorf("failed to detach policy from user: %w", err)
	}
	return nil
}

// AttachPolicyToServiceAccount attaches a policy to a service account
func (c *Client) AttachPolicyToServiceAccount(ctx context.Context, serviceAccountID, policyID string) error {
	req := map[string]string{
		"service_account_id": serviceAccountID,
		"policy_id":          policyID,
	}
	err := c.Post(ctx, "/api/auth/service_account/attach_policy", req, nil)
	if err != nil {
		return fmt.Errorf("failed to attach policy to service account: %w", err)
	}
	return nil
}

// DetachPolicyFromServiceAccount detaches a policy from a service account
func (c *Client) DetachPolicyFromServiceAccount(ctx context.Context, serviceAccountID, policyID string) error {
	req := map[string]string{
		"service_account_id": serviceAccountID,
		"policy_id":          policyID,
	}
	err := c.Post(ctx, "/api/auth/service_account/deatach_policy", req, nil)
	if err != nil {
		return fmt.Errorf("failed to detach policy from service account: %w", err)
	}
	return nil
}
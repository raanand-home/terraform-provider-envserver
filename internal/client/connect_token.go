package client

import (
	"context"
	"fmt"
)

// ConnectTokenRequest represents the request body for generating a connect token
type ConnectTokenRequest struct {
	Audience string `json:"audience"`
	Exp      int    `json:"exp,omitempty"`
}

// ConnectTokenResponse represents the response from the connect token endpoint
type ConnectTokenResponse struct {
	Token string `json:"access_token"`
}

// GetConnectToken generates an OIDC token for connecting to external services from an environment
func (c *Client) GetConnectToken(ctx context.Context, projectID, envID string, request *ConnectTokenRequest) (*ConnectTokenResponse, error) {
	path := fmt.Sprintf("/api/environments/%s/%s/connect_token", projectID, envID)

	var result ConnectTokenResponse
	err := c.Post(ctx, path, request, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

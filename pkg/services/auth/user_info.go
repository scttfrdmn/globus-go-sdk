// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserInfo represents information about a user from the userinfo endpoint
type UserInfo struct {
	Sub               string    `json:"sub"`
	Name              string    `json:"name"`
	PreferredUsername string    `json:"preferred_username"`
	Email             string    `json:"email"`
	EmailVerified     bool      `json:"email_verified"`
	IdentityProvider  string    `json:"identity_provider"`
	OrganizationID    string    `json:"organization"`
	LastAuthenticated time.Time `json:"last_authentication"`
}

// GetUserInfo retrieves information about the user associated with the token
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// Create a request to the userinfo endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Client.BaseURL+"oauth2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	// Add the access token to the request
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %s", resp.Status)
	}

	// Decode the response
	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

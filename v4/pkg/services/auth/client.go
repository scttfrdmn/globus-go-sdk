// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Auth service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Auth client
// In v4, config is required and must include explicit scopes
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	// Set default Auth service URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://auth.globus.org/v2"
	}

	// Create base client
	baseClient, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    config.BaseURL,
	}, nil
}

// GetUserInfo retrieves user information (requires openid scope)
// v4: Context is always first parameter
func (c *Client) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	var userInfo UserInfo
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/oauth2/userinfo", nil, nil, &userInfo)
	if err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// IntrospectToken introspects an OAuth2 token
// v4: Context is always first parameter
func (c *Client) IntrospectToken(ctx context.Context, token string) (*TokenIntrospection, error) {
	if token == "" {
		return nil, &core.ValidationError{
			Field:   "token",
			Message: "token is required",
		}
	}

	body := map[string]interface{}{
		"token": token,
	}

	var introspection TokenIntrospection
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token/introspect", nil, body, &introspection)
	if err != nil {
		return nil, err
	}

	return &introspection, nil
}

// RevokeToken revokes an OAuth2 token
// v4: Context is always first parameter
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return &core.ValidationError{
			Field:   "token",
			Message: "token is required",
		}
	}

	body := map[string]interface{}{
		"token": token,
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token/revoke", nil, body, nil)
}

// ExchangeAuthorizationCode exchanges an authorization code for tokens
// v4: Context is always first parameter
func (c *Client) ExchangeAuthorizationCode(ctx context.Context, code, clientID, clientSecret, redirectURI string) (*TokenResponse, error) {
	if code == "" {
		return nil, &core.ValidationError{
			Field:   "code",
			Message: "authorization code is required",
		}
	}
	if clientID == "" {
		return nil, &core.ValidationError{
			Field:   "clientID",
			Message: "client ID is required",
		}
	}
	if redirectURI == "" {
		return nil, &core.ValidationError{
			Field:   "redirectURI",
			Message: "redirect URI is required",
		}
	}

	// Build form data for token exchange
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("redirect_uri", redirectURI)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	var tokenResp TokenResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token", nil, data, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an OAuth2 access token using a refresh token
// v4: Context is always first parameter
func (c *Client) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, &core.ValidationError{
			Field:   "refreshToken",
			Message: "refresh token is required",
		}
	}
	if clientID == "" {
		return nil, &core.ValidationError{
			Field:   "clientID",
			Message: "client ID is required",
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	var tokenResp TokenResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token", nil, data, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// GetProjects retrieves the user's projects (requires manage_projects scope)
// v4: Context is always first parameter
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var projects []Project
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/api/projects", nil, nil, &projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// CreateProject creates a new project (requires manage_projects scope)
// v4: Context is always first parameter
func (c *Client) CreateProject(ctx context.Context, project *ProjectCreate) (*Project, error) {
	if project == nil {
		return nil, &core.ValidationError{
			Field:   "project",
			Message: "project data is required",
		}
	}
	if project.DisplayName == "" {
		return nil, &core.ValidationError{
			Field:   "DisplayName",
			Message: "project display name is required",
		}
	}

	var result Project
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v2/api/projects", nil, project, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetProject retrieves a specific project by ID
// v4: Context is always first parameter
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	if projectID == "" {
		return nil, &core.ValidationError{
			Field:   "projectID",
			Message: "project ID is required",
		}
	}

	var project Project
	endpoint := fmt.Sprintf("/v2/api/projects/%s", projectID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject deletes a project
// v4: Context is always first parameter
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return &core.ValidationError{
			Field:   "projectID",
			Message: "project ID is required",
		}
	}

	endpoint := fmt.Sprintf("/v2/api/projects/%s", projectID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}
// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}


// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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

// IntrospectToken introspects an OAuth2 token. The request is form-encoded per
// the OAuth2 introspection spec. Pass opts.Include (e.g. "identity_set") to
// request extra fields; opts may be nil.
// v4: Context is always first parameter
func (c *Client) IntrospectToken(ctx context.Context, token string, opts *IntrospectOptions) (*TokenIntrospection, error) {
	if token == "" {
		return nil, &core.ValidationError{
			Field:   "token",
			Message: "token is required",
		}
	}

	data := url.Values{}
	data.Set("token", token)
	if opts != nil && opts.Include != "" {
		data.Set("include", opts.Include)
	}

	var introspection TokenIntrospection
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token/introspect", nil, data, &introspection)
	if err != nil {
		return nil, err
	}

	return &introspection, nil
}

// RevokeToken revokes an OAuth2 token. The request is form-encoded.
// v4: Context is always first parameter
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return &core.ValidationError{
			Field:   "token",
			Message: "token is required",
		}
	}

	data := url.Values{}
	data.Set("token", token)

	return c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token/revoke", nil, data, nil)
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
	var envelope struct {
		Projects []Project `json:"projects"`
	}
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/projects", nil, nil, &envelope)
	if err != nil {
		return nil, err
	}
	return envelope.Projects, nil
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

	body := map[string]interface{}{"project": project}
	var result struct {
		Project Project `json:"project"`
	}
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/api/projects", nil, body, &result)
	if err != nil {
		return nil, err
	}
	return &result.Project, nil
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

	endpoint := fmt.Sprintf("/api/projects/%s", projectID)
	var result struct {
		Project Project `json:"project"`
	}
	err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result.Project, nil
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

	endpoint := fmt.Sprintf("/api/projects/%s", projectID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// GetAuthorizationURL constructs the Globus Auth authorization URL that the
// user should be redirected to as the first step of the authorization code
// flow.  It does not make an HTTP request.
func (c *Client) GetAuthorizationURL(opts *AuthorizationURLOptions) (string, error) {
	if opts == nil {
		return "", &core.ValidationError{Field: "opts", Message: "options are required"}
	}
	if opts.ClientID == "" {
		return "", &core.ValidationError{Field: "ClientID", Message: "client ID is required"}
	}
	if opts.RedirectURI == "" {
		return "", &core.ValidationError{Field: "RedirectURI", Message: "redirect URI is required"}
	}

	base := c.baseURL
	if base == "" {
		base = "https://auth.globus.org/v2"
	}

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", opts.ClientID)
	q.Set("redirect_uri", opts.RedirectURI)
	if len(opts.Scopes) > 0 {
		q.Set("scope", strings.Join(opts.Scopes, " "))
	}
	if opts.State != "" {
		q.Set("state", opts.State)
	}
	if opts.AccessType != "" {
		q.Set("access_type", opts.AccessType)
	}
	if opts.Prompt != "" {
		q.Set("prompt", opts.Prompt)
	}
	if len(opts.SessionRequiredIdentities) > 0 {
		q.Set("session_required_identities", strings.Join(opts.SessionRequiredIdentities, ","))
	}
	if len(opts.SessionRequiredSingleDomain) > 0 {
		q.Set("session_required_single_domain", strings.Join(opts.SessionRequiredSingleDomain, ","))
	}
	if len(opts.SessionRequiredPolicies) > 0 {
		q.Set("session_required_policies", strings.Join(opts.SessionRequiredPolicies, ","))
	}
	if opts.SessionRequiredMFA {
		q.Set("session_required_mfa", "true")
	}
	if opts.SessionMessage != "" {
		q.Set("session_message", opts.SessionMessage)
	}

	return base + "/oauth2/authorize?" + q.Encode(), nil
}

// StartDeviceAuthorization initiates the RFC 8628 device authorization grant.
func (c *Client) StartDeviceAuthorization(ctx context.Context, clientID string, scopes []string) (*DeviceAuthorizationResponse, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if len(scopes) == 0 {
		return nil, &core.ValidationError{Field: "scopes", Message: "at least one scope is required"}
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", strings.Join(scopes, " "))

	var resp DeviceAuthorizationResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/device/code", nil, data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ErrAuthorizationPending is returned by PollDeviceAuthorization when the user
// has not yet approved the request.
var ErrAuthorizationPending = errors.New("auth: authorization pending — user has not yet approved the device request")

// ErrSlowDown is returned by PollDeviceAuthorization when the server requests
// a reduced polling rate.
var ErrSlowDown = errors.New("auth: slow down — reduce polling frequency")

// PollDeviceAuthorization polls /oauth2/token for a device code grant.
func (c *Client) PollDeviceAuthorization(ctx context.Context, clientID, deviceCode string) (*TokenResponse, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if deviceCode == "" {
		return nil, &core.ValidationError{Field: "deviceCode", Message: "device code is required"}
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("client_id", clientID)
	data.Set("device_code", deviceCode)

	var tokenResp TokenResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token", nil, data, &tokenResp)
	if err != nil {
		var apiErr *core.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.Code {
			case "authorization_pending":
				return nil, ErrAuthorizationPending
			case "slow_down":
				return nil, ErrSlowDown
			}
		}
		return nil, err
	}
	return &tokenResp, nil
}

// WaitForDeviceAuthorization polls until approval or ctx cancellation.
func (c *Client) WaitForDeviceAuthorization(ctx context.Context, clientID string, resp *DeviceAuthorizationResponse) (*TokenResponse, error) {
	if resp == nil {
		return nil, &core.ValidationError{Field: "resp", Message: "device authorization response is required"}
	}

	interval := time.Duration(resp.Interval) * time.Second
	if interval < time.Second {
		interval = 5 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
			tokenResp, err := c.PollDeviceAuthorization(ctx, clientID, resp.DeviceCode)
			if err == nil {
				return tokenResp, nil
			}
			if errors.Is(err, ErrSlowDown) {
				interval += 5 * time.Second
				continue
			}
			if errors.Is(err, ErrAuthorizationPending) {
				continue
			}
			return nil, err
		}
	}
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

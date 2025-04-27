// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// Constants for Globus Auth
const (
	DefaultBaseURL = "https://auth.globus.org/v2/"
	AuthScope      = "openid profile email"
)

// Client provides methods for interacting with Globus Auth
type Client struct {
	Client       *core.Client
	Transport    *transport.Transport
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewClient creates a new Auth client
func NewClient(clientID, clientSecret string, options ...core.ClientOption) *Client {
	// Apply default options specific to Auth
	defaultOptions := []core.ClientOption{
		core.WithBaseURL(DefaultBaseURL),
		core.WithAuthorizer(authorizers.NullCoreAuthorizer()),
	}

	// Merge with user options, letting user options override defaults
	options = append(defaultOptions, options...)

	// Create the base client
	baseClient := core.NewClient(options...)

	return &Client{
		Client:       baseClient,
		Transport:    transport.NewTransport(baseClient),
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

// SetRedirectURL sets the redirect URL for OAuth flows
func (c *Client) SetRedirectURL(redirectURL string) {
	c.RedirectURL = redirectURL
}

// GetAuthorizationURL returns a URL for user authorization
func (c *Client) GetAuthorizationURL(state string, scopes ...string) string {
	// Use default scope if none provided
	if len(scopes) == 0 {
		scopes = []string{AuthScope}
	}

	// Build the scopes string
	scopesStr := strings.Join(scopes, " ")

	// Build the query parameters
	query := url.Values{}
	query.Set("client_id", c.ClientID)
	query.Set("redirect_uri", c.RedirectURL)
	query.Set("scope", scopesStr)
	query.Set("state", state)
	query.Set("response_type", "code")

	// Build the authorization URL
	authURL := fmt.Sprintf("%soauth2/authorize?%s", c.Client.BaseURL, query.Encode())

	return authURL
}

// ExchangeAuthorizationCode exchanges an authorization code for tokens
func (c *Client) ExchangeAuthorizationCode(ctx context.Context, code string) (*TokenResponse, error) {
	if c.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required for code exchange")
	}

	// Build the request body
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURL)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	return c.tokenRequest(ctx, form)
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Build the request body
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	return c.tokenRequest(ctx, form)
}

// tokenRequest makes a token request with the given form data
func (c *Client) tokenRequest(ctx context.Context, form url.Values) (*TokenResponse, error) {
	// Set the headers
	headers := http.Header{}
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create the request body
	body := strings.NewReader(form.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL+"oauth2/token", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Calculate expiry time
	tokenResponse.ExpiryTime = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return &tokenResponse, nil
}

// IntrospectToken gets information about a token
func (c *Client) IntrospectToken(ctx context.Context, token string) (*TokenInfo, error) {
	// Build the request body
	form := url.Values{}
	form.Set("token", token)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	// Set the headers
	headers := http.Header{}
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create the request body
	body := strings.NewReader(form.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL+"oauth2/token/introspect", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create introspect request: %w", err)
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("introspect request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("introspect request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse introspect response: %w", err)
	}

	return &tokenInfo, nil
}

// RevokeToken revokes a token
func (c *Client) RevokeToken(ctx context.Context, token string) error {
	// Build the request body
	form := url.Values{}
	form.Set("token", token)
	form.Set("client_id", c.ClientID)

	// Add client secret if available
	if c.ClientSecret != "" {
		form.Set("client_secret", c.ClientSecret)
	}

	// Set the headers
	headers := http.Header{}
	headers.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create the request body
	body := strings.NewReader(form.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL+"oauth2/token/revoke", body)
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	// Set headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetClientCredentialsToken gets a token using client credentials
func (c *Client) GetClientCredentialsToken(ctx context.Context, scopes ...string) (*TokenResponse, error) {
	if c.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required for client credentials flow")
	}

	// Use default scope if none provided
	if len(scopes) == 0 {
		scopes = []string{AuthScope}
	}

	// Build the scopes string
	scopesStr := strings.Join(scopes, " ")

	// Build the request body
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("scope", scopesStr)

	return c.tokenRequest(ctx, form)
}

// CreateClientCredentialsAuthorizer creates an authorizer that uses client credentials
func (c *Client) CreateClientCredentialsAuthorizer(scopes ...string) *authorizers.ClientCredentialsAuthorizer {
	authFunc := func(ctx context.Context, clientID, clientSecret string, scopes []string) (string, time.Time, error) {
		// Create a temporary client for this request
		tempClient := NewClient(clientID, clientSecret)

		// Get the token
		tokenResp, err := tempClient.GetClientCredentialsToken(ctx, scopes...)
		if err != nil {
			return "", time.Time{}, err
		}

		return tokenResp.AccessToken, tokenResp.ExpiryTime, nil
	}

	return authorizers.NewClientCredentialsAuthorizer(c.ClientID, c.ClientSecret, scopes, authFunc)
}

// CreateRefreshableTokenAuthorizer creates an authorizer that can refresh tokens
func (c *Client) CreateRefreshableTokenAuthorizer(accessToken, refreshToken string, expiresIn int) *authorizers.RefreshableTokenAuthorizer {
	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		// Refresh the token
		tokenResp, err := c.RefreshToken(ctx, refreshToken)
		if err != nil {
			return "", "", time.Time{}, err
		}

		// Return the new tokens
		return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiryTime, nil
	}

	return authorizers.NewRefreshableTokenAuthorizer(accessToken, refreshToken, expiresIn, refreshFunc)
}

// CreateStaticTokenAuthorizer creates an authorizer with a static token
func (c *Client) CreateStaticTokenAuthorizer(accessToken string) *authorizers.StaticTokenAuthorizer {
	return authorizers.NewStaticTokenAuthorizer(accessToken)
}

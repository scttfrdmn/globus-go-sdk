// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// ClientCredentialsAuthorizer obtains access tokens via the OAuth2 client
// credentials grant. No user interaction is required. It implements core.Authorizer.
type ClientCredentialsAuthorizer struct {
	renewingAuthorizer
	clientID     string
	clientSecret string
	scopes       []string
	authBaseURL  string
	httpClient   *http.Client
}

// ClientCredentialsOption is a functional option for NewClientCredentialsAuthorizer.
type ClientCredentialsOption func(*ClientCredentialsAuthorizer)

// WithClientCredentialsAuthBaseURL overrides the Globus Auth base URL.
func WithClientCredentialsAuthBaseURL(baseURL string) ClientCredentialsOption {
	return func(a *ClientCredentialsAuthorizer) {
		a.authBaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithClientCredentialsHTTPClient sets the HTTP client used for token requests.
func WithClientCredentialsHTTPClient(c *http.Client) ClientCredentialsOption {
	return func(a *ClientCredentialsAuthorizer) {
		a.httpClient = c
	}
}

// NewClientCredentialsAuthorizer creates an authorizer that uses the client
// credentials grant to obtain access tokens automatically.
func NewClientCredentialsAuthorizer(clientID, clientSecret string, scopes []string, opts ...ClientCredentialsOption) *ClientCredentialsAuthorizer {
	a := &ClientCredentialsAuthorizer{
		clientID:     clientID,
		clientSecret: clientSecret,
		scopes:       scopes,
		authBaseURL:  "https://auth.globus.org",
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(a)
	}
	a.fetchNewTokens = a.doClientCredentials
	return a
}

// GetAuthorizationHeader returns the Bearer authorization header, obtaining a
// new token via client credentials if necessary.
func (a *ClientCredentialsAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	return a.getAuthorizationHeader(ctx)
}

// HandleMissingAuthorization forces a token refresh on 401 responses.
func (a *ClientCredentialsAuthorizer) HandleMissingAuthorization(ctx context.Context) bool {
	a.mu.Lock()
	a.accessToken = "" // force refresh
	a.mu.Unlock()
	return a.ensureValidToken(ctx) == nil
}

// doClientCredentials calls the Globus Auth token endpoint using the client
// credentials grant. Must be called with a.mu held (via ensureValidToken).
func (a *ClientCredentialsAuthorizer) doClientCredentials(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	if len(a.scopes) > 0 {
		form.Set("scope", strings.Join(a.scopes, " "))
	}

	tokenURL := a.authBaseURL + "/v2/oauth2/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("authorizer: create client credentials request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authorizer: client credentials request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &core.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("authorizer: client credentials returned HTTP %d", resp.StatusCode),
		}
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("authorizer: decode client credentials response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	a.setToken(result.AccessToken, expiresAt)
	return nil
}

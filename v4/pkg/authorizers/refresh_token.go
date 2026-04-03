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

// RefreshTokenAuthorizer automatically refreshes an access token using a
// refresh token. It implements core.Authorizer.
type RefreshTokenAuthorizer struct {
	renewingAuthorizer
	refreshToken string
	clientID     string
	clientSecret string
	authBaseURL  string
	httpClient   *http.Client
	onRefresh    func(accessToken, refreshToken string, expiresAt time.Time)
}

// RefreshTokenOption is a functional option for NewRefreshTokenAuthorizer.
type RefreshTokenOption func(*RefreshTokenAuthorizer)

// WithInitialAccessToken seeds the authorizer with an existing access token
// so it does not refresh immediately on first use.
func WithInitialAccessToken(accessToken string, expiresAt time.Time) RefreshTokenOption {
	return func(a *RefreshTokenAuthorizer) {
		a.accessToken = accessToken
		a.expiresAt = expiresAt
	}
}

// WithOnRefresh registers a callback invoked after every successful token refresh.
func WithOnRefresh(fn func(accessToken, refreshToken string, expiresAt time.Time)) RefreshTokenOption {
	return func(a *RefreshTokenAuthorizer) {
		a.onRefresh = fn
	}
}

// WithAuthBaseURL overrides the Globus Auth base URL (default: https://auth.globus.org).
func WithAuthBaseURL(baseURL string) RefreshTokenOption {
	return func(a *RefreshTokenAuthorizer) {
		a.authBaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient sets the HTTP client used for token refresh calls.
func WithHTTPClient(c *http.Client) RefreshTokenOption {
	return func(a *RefreshTokenAuthorizer) {
		a.httpClient = c
	}
}

// NewRefreshTokenAuthorizer creates an authorizer that uses the given refresh
// token to obtain new access tokens automatically.
func NewRefreshTokenAuthorizer(refreshToken, clientID, clientSecret string, opts ...RefreshTokenOption) *RefreshTokenAuthorizer {
	a := &RefreshTokenAuthorizer{
		refreshToken: refreshToken,
		clientID:     clientID,
		clientSecret: clientSecret,
		authBaseURL:  "https://auth.globus.org",
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(a)
	}
	a.fetchNewTokens = a.doRefresh
	return a
}

// GetAuthorizationHeader returns the Bearer authorization header, refreshing
// the token first if it is close to expiry.
func (a *RefreshTokenAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	return a.getAuthorizationHeader(ctx)
}

// HandleMissingAuthorization attempts a token refresh on 401 responses.
func (a *RefreshTokenAuthorizer) HandleMissingAuthorization(ctx context.Context) bool {
	a.mu.Lock()
	a.accessToken = "" // force refresh
	a.mu.Unlock()
	return a.ensureValidToken(ctx) == nil
}

// doRefresh calls the Globus Auth token endpoint to obtain a new access token.
// Must be called with a.mu held (via ensureValidToken).
func (a *RefreshTokenAuthorizer) doRefresh(ctx context.Context) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	if a.clientSecret != "" {
		form.Set("client_secret", a.clientSecret)
	}

	tokenURL := a.authBaseURL + "/v2/oauth2/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("authorizer: create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authorizer: refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &core.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("authorizer: token refresh returned HTTP %d", resp.StatusCode),
		}
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("authorizer: decode refresh response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	a.setToken(result.AccessToken, expiresAt)

	if result.RefreshToken != "" {
		a.refreshToken = result.RefreshToken
	}

	if a.onRefresh != nil {
		a.onRefresh(result.AccessToken, a.refreshToken, expiresAt)
	}
	return nil
}

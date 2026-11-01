// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

const (
	defaultAuthBaseURL = "https://auth.globus.org"
	revokePath         = "/v2/oauth2/token/revoke"
)

// revokeToken revokes one access or refresh token at Globus Auth, matching
// upstream AuthLoginClient.oauth2_revoke_token: a form-encoded "token" field,
// authenticated with HTTP Basic auth for confidential clients, or carrying
// "client_id" in the body for native (public) clients.
func (c *AppConfig) revokeToken(ctx context.Context, clientID, clientSecret, token string) error {
	form := url.Values{"token": {token}}
	if clientSecret == "" {
		form.Set("client_id", clientID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.AuthBaseURL+revokePath, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revoke endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// clearTokens revokes then removes each token, as upstream GlobusApp.logout
// does. Unlike upstream (which stops at the first revocation error), this is
// best-effort: a token is removed from storage even if revoking it failed, so
// logging out always clears local credentials. All failures are joined into
// the returned error.
func (c *AppConfig) clearTokens(ctx context.Context, clientID, clientSecret string, tokens []*tokenstorage.TokenData) error {
	var errs []error
	for _, td := range tokens {
		if td == nil {
			continue
		}
		for _, tok := range []string{td.AccessToken, td.RefreshToken} {
			if tok == "" {
				continue
			}
			if err := c.revokeToken(ctx, clientID, clientSecret, tok); err != nil {
				errs = append(errs, fmt.Errorf("app: revoke token for %s: %w", td.ResourceServer, err))
			}
		}
		if err := c.TokenStorage.Remove(td.ResourceServer); err != nil {
			errs = append(errs, fmt.Errorf("app: remove token for %s: %w", td.ResourceServer, err))
		}
	}
	return errors.Join(errs...)
}

func defaultHTTPClient() *http.Client { return &http.Client{Timeout: 30 * time.Second} }

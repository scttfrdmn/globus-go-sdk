// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package login

import (
	"context"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// AuthParams contains the parameters for initiating an OAuth2 login flow.
type AuthParams struct {
	// Scopes is the list of OAuth2 scopes to request.
	Scopes []string

	// RedirectURI overrides the manager's default redirect URI.
	// Leave empty to use the manager default.
	RedirectURI string

	// RequestRefresh requests a refresh token (offline_access scope).
	// The manager appends "offline_access" to Scopes when true.
	RequestRefresh bool

	// State is an optional CSRF token embedded in the authorization URL.
	State string
}

// LoginResult is returned by LoginFlowManager.RunLoginFlow after a successful
// OAuth2 authorization code exchange. It contains token data for one or more
// resource servers.
type LoginResult struct {
	// Tokens holds the token data for each resource server returned by the
	// authorization server. Globus Auth may return tokens for multiple
	// resource servers via the `other_tokens` extension field.
	Tokens []*tokenstorage.TokenData
}

// LoginFlowManager drives an OAuth2 authorization code flow.
// Implementations are responsible for obtaining user consent and exchanging
// the resulting authorization code for tokens.
type LoginFlowManager interface {
	RunLoginFlow(ctx context.Context, params AuthParams) (*LoginResult, error)
}

// tokenExchangeResponse is the JSON structure returned by the token endpoint.
// It mirrors auth.TokenResponse but is private to avoid an import cycle.
type tokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`

	// ResourceServer is a Globus Auth extension field.
	ResourceServer string `json:"resource_server,omitempty"`

	// OtherTokens is a Globus Auth extension: additional tokens for dependent
	// resource servers, returned alongside the primary token.
	OtherTokens []tokenExchangeResponse `json:"other_tokens,omitempty"`
}

// toTokenData converts a tokenExchangeResponse to a tokenstorage.TokenData.
// resourceServer defaults to "auth.globus.org" when the field is absent.
func (r *tokenExchangeResponse) toTokenData() *tokenstorage.TokenData {
	rs := r.ResourceServer
	if rs == "" {
		rs = "auth.globus.org"
	}
	return &tokenstorage.TokenData{
		ResourceServer: rs,
		AccessToken:    r.AccessToken,
		RefreshToken:   r.RefreshToken,
		Scope:          r.Scope,
		TokenType:      r.TokenType,
		ExpiresAt:      time.Now().Add(time.Duration(r.ExpiresIn) * time.Second),
	}
}

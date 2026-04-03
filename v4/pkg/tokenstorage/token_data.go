// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package tokenstorage

import "time"

// TokenData holds the OAuth2 token set for a single resource server.
type TokenData struct {
	// ResourceServer is the resource server identifier (e.g. "transfer.api.globus.org").
	ResourceServer string `json:"resource_server"`

	// IdentityID is the Globus identity that owns the tokens, if known.
	IdentityID string `json:"identity_id,omitempty"`

	// Scope is the space-separated list of scopes granted.
	Scope string `json:"scope,omitempty"`

	// AccessToken is the current access token.
	AccessToken string `json:"access_token"`

	// RefreshToken is the refresh token, if one was granted.
	RefreshToken string `json:"refresh_token,omitempty"`

	// ExpiresAt is the wall-clock time when the access token expires.
	ExpiresAt time.Time `json:"expires_at"`

	// TokenType is the token type (usually "Bearer").
	TokenType string `json:"token_type,omitempty"`
}

// IsExpired returns true if the access token has expired.
func (t *TokenData) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// ExpiresIn returns the duration remaining until the token expires.
// Returns 0 if already expired.
func (t *TokenData) ExpiresIn() time.Duration {
	d := time.Until(t.ExpiresAt)
	if d < 0 {
		return 0
	}
	return d
}

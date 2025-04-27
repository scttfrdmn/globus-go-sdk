// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"time"
)

// TokenResponse represents a response containing tokens
type TokenResponse struct {
	AccessToken     string            `json:"access_token"`
	RefreshToken    string            `json:"refresh_token,omitempty"`
	ExpiresIn       int               `json:"expires_in"`
	ResourceServer  string            `json:"resource_server"`
	TokenType       string            `json:"token_type"`
	OtherTokens     []json.RawMessage `json:"other_tokens,omitempty"`     // Raw to avoid recursion
	DependentTokens []json.RawMessage `json:"dependent_tokens,omitempty"` // Raw to avoid recursion
	Scope           string            `json:"scope"`
	State           string            `json:"state,omitempty"`
	ExpiryTime      time.Time         `json:"-"` // Calculated expiry time
}

// HasRefreshToken returns true if the response contains a refresh token
func (t *TokenResponse) HasRefreshToken() bool {
	return t.RefreshToken != ""
}

// IsValid returns true if the token is still valid (not expired)
func (t *TokenResponse) IsValid() bool {
	// Add a buffer of 30 seconds to avoid edge cases
	return time.Now().Add(30 * time.Second).Before(t.ExpiryTime)
}

// GetOtherTokens parses the other_tokens field into actual TokenResponse objects
func (t *TokenResponse) GetOtherTokens() ([]*TokenResponse, error) {
	if len(t.OtherTokens) == 0 {
		return nil, nil
	}

	tokens := make([]*TokenResponse, 0, len(t.OtherTokens))
	for _, raw := range t.OtherTokens {
		var token TokenResponse
		if err := json.Unmarshal(raw, &token); err != nil {
			return nil, err
		}
		// Calculate expiry time for each token
		token.ExpiryTime = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		tokens = append(tokens, &token)
	}
	return tokens, nil
}

// GetDependentTokens parses the dependent_tokens field into actual TokenResponse objects
func (t *TokenResponse) GetDependentTokens() ([]*TokenResponse, error) {
	if len(t.DependentTokens) == 0 {
		return nil, nil
	}

	tokens := make([]*TokenResponse, 0, len(t.DependentTokens))
	for _, raw := range t.DependentTokens {
		var token TokenResponse
		if err := json.Unmarshal(raw, &token); err != nil {
			return nil, err
		}
		// Calculate expiry time for each token
		token.ExpiryTime = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		tokens = append(tokens, &token)
	}
	return tokens, nil
}

// Identity represents a Globus Auth identity
type Identity struct {
	IdentityID       string `json:"identity_id"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Status           string `json:"status"`
	IdentityProvider string `json:"identity_provider"`
	Organization     string `json:"organization"`
}

// IdentitySet represents a collection of identities
type IdentitySet struct {
	Identities []Identity `json:"identities"`
}

// TokenInfo represents information about a token
type TokenInfo struct {
	Active      bool     `json:"active"`
	Scope       string   `json:"scope"`
	ClientID    string   `json:"client_id"`
	Username    string   `json:"username"`
	Exp         int64    `json:"exp"`
	SubjectType string   `json:"sub_type"`
	Subject     string   `json:"sub"`
	IdentitySet []string `json:"identity_set,omitempty"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
}

// IsActive returns true if the token is active
func (t *TokenInfo) IsActive() bool {
	return t.Active
}

// ExpiresAt returns the expiry time as a Time object
func (t *TokenInfo) ExpiresAt() time.Time {
	return time.Unix(t.Exp, 0)
}

// IsExpired returns true if the token is expired
func (t *TokenInfo) IsExpired() bool {
	// Add a buffer of 30 seconds to avoid edge cases
	return time.Now().Add(30 * time.Second).After(t.ExpiresAt())
}

// AuthorizeResponse represents a response from the authorize endpoint
type AuthorizeResponse struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

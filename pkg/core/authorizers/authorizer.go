// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"time"
)

// Authorizer defines the interface for authorization mechanisms
type Authorizer interface {
	// GetAuthorizationHeader returns the authorization header value
	GetAuthorizationHeader(ctx context.Context) (string, error)

	// HandleMissingAuthorization is called when authorization is missing or expired
	// Returns true if the authorization was refreshed and the request should be retried
	HandleMissingAuthorization(ctx context.Context) bool

	// IsExpired checks if the authorization is expired
	IsExpired() bool
}

// NullAuthorizer implements Authorizer with no authentication
type NullAuthorizer struct{}

// GetAuthorizationHeader for NullAuthorizer returns an empty string
func (a *NullAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	return "", nil
}

// HandleMissingAuthorization for NullAuthorizer always returns false
func (a *NullAuthorizer) HandleMissingAuthorization(_ context.Context) bool {
	return false
}

// IsExpired for NullAuthorizer always returns false
func (a *NullAuthorizer) IsExpired() bool {
	return false
}

// StaticTokenAuthorizer implements Authorizer with a static bearer token
type StaticTokenAuthorizer struct {
	Token string
}

// NewStaticTokenAuthorizer creates a new StaticTokenAuthorizer
func NewStaticTokenAuthorizer(token string) *StaticTokenAuthorizer {
	return &StaticTokenAuthorizer{Token: token}
}

// GetAuthorizationHeader for StaticTokenAuthorizer returns a bearer token header
func (a *StaticTokenAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	if a.Token == "" {
		return "", nil
	}
	return "Bearer " + a.Token, nil
}

// HandleMissingAuthorization for StaticTokenAuthorizer always returns false
func (a *StaticTokenAuthorizer) HandleMissingAuthorization(_ context.Context) bool {
	return false
}

// IsExpired for StaticTokenAuthorizer always returns false
func (a *StaticTokenAuthorizer) IsExpired() bool {
	return false
}

// RefreshableTokenAuthorizer implements Authorizer with a refreshable token
type RefreshableTokenAuthorizer struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	RefreshFunc  func(ctx context.Context, refreshToken string) (string, string, time.Time, error)
}

// NewRefreshableTokenAuthorizer creates a new RefreshableTokenAuthorizer
func NewRefreshableTokenAuthorizer(
	accessToken string,
	refreshToken string,
	expiresIn int,
	refreshFunc func(ctx context.Context, refreshToken string) (string, string, time.Time, error),
) *RefreshableTokenAuthorizer {
	return &RefreshableTokenAuthorizer{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		RefreshFunc:  refreshFunc,
	}
}

// GetAuthorizationHeader returns the authorization header with the access token
func (a *RefreshableTokenAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	if a.AccessToken == "" {
		return "", nil
	}
	return "Bearer " + a.AccessToken, nil
}

// HandleMissingAuthorization refreshes the token if expired
func (a *RefreshableTokenAuthorizer) HandleMissingAuthorization(ctx context.Context) bool {
	// If no refresh function or token, can't handle
	if a.RefreshFunc == nil || a.RefreshToken == "" {
		return false
	}

	// If not expired, no need to refresh
	if !a.IsExpired() {
		return false
	}

	// Refresh the token
	accessToken, refreshToken, expiresAt, err := a.RefreshFunc(ctx, a.RefreshToken)
	if err != nil {
		return false
	}

	// Update tokens
	a.AccessToken = accessToken
	if refreshToken != "" {
		a.RefreshToken = refreshToken
	}
	a.ExpiresAt = expiresAt

	return true
}

// IsExpired checks if the token is expired
func (a *RefreshableTokenAuthorizer) IsExpired() bool {
	// Add a buffer of 30 seconds to avoid edge cases
	return time.Now().Add(30 * time.Second).After(a.ExpiresAt)
}

// ClientCredentialsAuthorizer implements Authorizer using client credentials flow
type ClientCredentialsAuthorizer struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	AccessToken  string
	ExpiresAt    time.Time
	AuthFunc     func(ctx context.Context, clientID, clientSecret string, scopes []string) (string, time.Time, error)
}

// NewClientCredentialsAuthorizer creates a new ClientCredentialsAuthorizer
func NewClientCredentialsAuthorizer(
	clientID string,
	clientSecret string,
	scopes []string,
	authFunc func(ctx context.Context, clientID, clientSecret string, scopes []string) (string, time.Time, error),
) *ClientCredentialsAuthorizer {
	return &ClientCredentialsAuthorizer{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		AuthFunc:     authFunc,
	}
}

// GetAuthorizationHeader returns the authorization header with the client credentials token
func (a *ClientCredentialsAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	// If token is expired or not set, get a new one
	if a.IsExpired() {
		if err := a.refreshToken(ctx); err != nil {
			return "", err
		}
	}

	return "Bearer " + a.AccessToken, nil
}

// HandleMissingAuthorization refreshes the token
func (a *ClientCredentialsAuthorizer) HandleMissingAuthorization(ctx context.Context) bool {
	// If no authFunc, can't handle
	if a.AuthFunc == nil {
		return false
	}
	err := a.refreshToken(ctx)
	return err == nil
}

// IsExpired checks if the token is expired
func (a *ClientCredentialsAuthorizer) IsExpired() bool {
	return a.AccessToken == "" || time.Now().Add(30*time.Second).After(a.ExpiresAt)
}

// refreshToken gets a new token using client credentials
func (a *ClientCredentialsAuthorizer) refreshToken(ctx context.Context) error {
	if a.AuthFunc == nil {
		return nil
	}

	accessToken, expiresAt, err := a.AuthFunc(ctx, a.ClientID, a.ClientSecret, a.Scopes)
	if err != nil {
		return err
	}

	a.AccessToken = accessToken
	a.ExpiresAt = expiresAt
	return nil
}

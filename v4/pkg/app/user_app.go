// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/login"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// UserApp is a GlobusApp implementation for applications that authenticate
// on behalf of a human user via an interactive login flow.
type UserApp struct {
	clientID     string
	clientSecret string
	config       *AppConfig

	mu     sync.RWMutex
	scopes map[string][]string // resourceServer → scopes
}

// NewUserApp creates a UserApp. clientSecret may be empty for public (native) clients.
// config may be nil; defaults are applied.
func NewUserApp(clientID, clientSecret string, config *AppConfig) (*UserApp, error) {
	if clientID == "" {
		return nil, fmt.Errorf("app: clientID is required")
	}
	if config == nil {
		config = &AppConfig{}
	}
	return &UserApp{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config.withDefaults(clientID, clientSecret),
		scopes:       make(map[string][]string),
	}, nil
}

// AddScopeRequirements registers scopes needed for the given resource server.
func (a *UserApp) AddScopeRequirements(resourceServer string, scopes ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scopes[resourceServer] = append(a.scopes[resourceServer], scopes...)
}

// Login drives an interactive login flow and stores the resulting tokens.
func (a *UserApp) Login(ctx context.Context) error {
	a.mu.RLock()
	allScopes := a.collectScopes()
	a.mu.RUnlock()

	params := login.AuthParams{
		Scopes:         allScopes,
		RequestRefresh: a.config.RequestRefreshTokens,
	}

	result, err := a.config.LoginFlowManager.RunLoginFlow(ctx, params)
	if err != nil {
		return fmt.Errorf("app: login flow failed: %w", err)
	}

	for _, td := range result.Tokens {
		if err := a.config.TokenStorage.Store(td); err != nil {
			return fmt.Errorf("app: store token for %s: %w", td.ResourceServer, err)
		}
	}
	return nil
}

// LoginRequired returns true if any registered resource server has no stored token.
func (a *UserApp) LoginRequired() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for rs := range a.scopes {
		td, err := a.config.TokenStorage.Get(rs)
		if err != nil || td == nil {
			return true
		}
	}
	return false
}

// GetAuthorizer returns an Authorizer for the given resource server.
// Returns RefreshTokenAuthorizer when a refresh token is stored, otherwise
// AccessTokenAuthorizer.
func (a *UserApp) GetAuthorizer(_ context.Context, resourceServer string) (core.Authorizer, error) {
	td, err := a.config.TokenStorage.Get(resourceServer)
	if err != nil {
		return nil, fmt.Errorf("app: token storage error: %w", err)
	}
	if td == nil {
		return nil, fmt.Errorf("app: no token for resource server %q — call Login first", resourceServer)
	}

	if td.RefreshToken != "" {
		return authorizers.NewRefreshTokenAuthorizer(
			td.RefreshToken, a.clientID, a.clientSecret,
			authorizers.WithInitialAccessToken(td.AccessToken, td.ExpiresAt),
		), nil
	}
	return authorizers.NewAccessTokenAuthorizer(td.AccessToken), nil
}

// Logout revokes and removes stored tokens for all registered resource
// servers. Equivalent to LogoutWithOptions(ctx, nil).
func (a *UserApp) Logout(ctx context.Context) error {
	return a.LogoutWithOptions(ctx, nil)
}

// LogoutWithOptions revokes each stored access and refresh token at Globus
// Auth, then removes it from storage (upstream GlobusApp.logout). By default
// only tokens for the app's registered resource servers are cleared; with
// opts.Sweep every token in storage is (upstream logout(sweep=True), v4.9.0).
//
// Revocation is best-effort: tokens are removed locally even when revoking
// them fails, and any failures are returned joined in a single error.
func (a *UserApp) LogoutWithOptions(ctx context.Context, opts *LogoutOptions) error {
	var tokens []*tokenstorage.TokenData
	if opts != nil && opts.Sweep {
		all, err := a.config.TokenStorage.GetAll()
		if err != nil {
			return fmt.Errorf("app: list stored tokens: %w", err)
		}
		tokens = all
	} else {
		a.mu.RLock()
		for rs := range a.scopes {
			td, err := a.config.TokenStorage.Get(rs)
			if err != nil {
				a.mu.RUnlock()
				return fmt.Errorf("app: read token for %s: %w", rs, err)
			}
			if td != nil {
				tokens = append(tokens, td)
			}
		}
		a.mu.RUnlock()
	}
	return a.config.clearTokens(ctx, a.clientID, a.clientSecret, tokens)
}

// Close releases resources held by the app.
func (a *UserApp) Close() error {
	return a.config.TokenStorage.Close()
}

// collectScopes returns a deduplicated list of all registered scopes.
// Must be called with a.mu held (at least read).
func (a *UserApp) collectScopes() []string {
	seen := make(map[string]struct{})
	var out []string
	for _, ss := range a.scopes {
		for _, s := range ss {
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				out = append(out, s)
			}
		}
	}
	return out
}

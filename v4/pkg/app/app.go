// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package app

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/login"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
)

// AppConfig controls the behaviour of UserApp and ClientApp.
// All fields are optional; sensible defaults are applied.
type AppConfig struct {
	// TokenStorage persists OAuth2 tokens between calls.
	// Default: MemoryTokenStorage (tokens are lost when the process exits).
	TokenStorage tokenstorage.TokenStorage

	// LoginFlowManager drives the OAuth2 authorization code flow for UserApp.
	// Default: CommandLineLoginFlowManager.
	LoginFlowManager login.LoginFlowManager

	// RequestRefreshTokens requests refresh tokens during login so that
	// access tokens can be renewed without user interaction.
	RequestRefreshTokens bool

	// Environment is the Globus environment name (e.g. "production", "sandbox").
	// Default: "production".
	Environment string
}

// withDefaults returns a copy of c with nil-fields populated by defaults.
// It also initialises the LoginFlowManager when clientID and clientSecret are known.
func (c *AppConfig) withDefaults(clientID, clientSecret string) *AppConfig {
	cfg := *c

	if cfg.TokenStorage == nil {
		cfg.TokenStorage = tokenstorage.NewMemoryTokenStorage()
	}
	if cfg.Environment == "" {
		cfg.Environment = "production"
	}
	// LoginFlowManager is initialised by the caller (UserApp/ClientApp) because
	// it needs clientID/clientSecret which are not in AppConfig.
	if cfg.LoginFlowManager == nil {
		cfg.LoginFlowManager = login.NewCommandLineLoginFlowManager(clientID, clientSecret)
	}
	return &cfg
}

// GlobusApp is the interface implemented by UserApp and ClientApp.
type GlobusApp interface {
	// Login initiates user authentication (UserApp) or is a no-op (ClientApp).
	Login(ctx context.Context) error

	// Logout removes stored tokens for all registered resource servers.
	Logout(ctx context.Context) error

	// LoginRequired returns true when stored credentials are missing for one
	// or more registered resource servers.
	LoginRequired() bool

	// GetAuthorizer returns an Authorizer for the given resource server.
	// Returns an error when no credentials are available.
	GetAuthorizer(ctx context.Context, resourceServer string) (core.Authorizer, error)

	// AddScopeRequirements registers the scopes needed for a resource server.
	// Multiple calls accumulate; duplicates are deduplicated at login time.
	AddScopeRequirements(resourceServer string, scopes ...string)

	// Close releases any resources held by the app (e.g. token storage).
	Close() error
}

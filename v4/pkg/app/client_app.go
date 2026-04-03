// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// ClientApp is a GlobusApp implementation for service-account / machine-to-machine
// applications that authenticate using the OAuth2 client credentials grant.
// No user interaction is required.
type ClientApp struct {
	clientID     string
	clientSecret string
	config       *AppConfig

	mu     sync.RWMutex
	scopes map[string][]string // resourceServer → scopes
}

// NewClientApp creates a ClientApp. config may be nil; defaults are applied.
func NewClientApp(clientID, clientSecret string, config *AppConfig) (*ClientApp, error) {
	if clientID == "" {
		return nil, fmt.Errorf("app: clientID is required")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("app: clientSecret is required for ClientApp")
	}
	if config == nil {
		config = &AppConfig{}
	}
	return &ClientApp{
		clientID:     clientID,
		clientSecret: clientSecret,
		config:       config.withDefaults(clientID, clientSecret),
		scopes:       make(map[string][]string),
	}, nil
}

// AddScopeRequirements registers scopes needed for the given resource server.
func (a *ClientApp) AddScopeRequirements(resourceServer string, scopes ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.scopes[resourceServer] = append(a.scopes[resourceServer], scopes...)
}

// Login is a no-op for ClientApp — client credentials do not require user interaction.
func (a *ClientApp) Login(_ context.Context) error {
	return nil
}

// LoginRequired always returns false for ClientApp.
func (a *ClientApp) LoginRequired() bool {
	return false
}

// GetAuthorizer returns a ClientCredentialsAuthorizer for the given resource server.
// The scopes sent to the token endpoint are those registered via AddScopeRequirements.
func (a *ClientApp) GetAuthorizer(_ context.Context, resourceServer string) (core.Authorizer, error) {
	a.mu.RLock()
	scopes := append([]string{}, a.scopes[resourceServer]...)
	a.mu.RUnlock()

	return authorizers.NewClientCredentialsAuthorizer(a.clientID, a.clientSecret, scopes), nil
}

// Logout is a no-op for ClientApp — client credentials tokens are not stored.
func (a *ClientApp) Logout(_ context.Context) error {
	return nil
}

// Close releases resources held by the app.
func (a *ClientApp) Close() error {
	return a.config.TokenStorage.Close()
}

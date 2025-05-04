// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// RefreshHandler defines the interface for token refresh operations
type RefreshHandler interface {
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, error)
}

// Ensure that auth.Client implements RefreshHandler
var _ RefreshHandler = (*auth.Client)(nil)

// Manager handles token storage, retrieval, and automatic refreshing
type Manager struct {
	Storage          Storage
	RefreshThreshold time.Duration
	RefreshHandler   RefreshHandler
	refreshMutex     sync.Mutex
}

// NewManager creates a new token manager with the provided options
func NewManager(opts ...ClientOption) (*Manager, error) {
	// Apply default options
	options := defaultOptions()
	
	// Apply user options
	for _, opt := range opts {
		opt(options)
	}
	
	// Validate required options
	if options.storage == nil {
		return nil, fmt.Errorf("no storage provided")
	}

	return &Manager{
		Storage:          options.storage,
		RefreshThreshold: options.refreshThreshold,
		RefreshHandler:   options.refreshHandler,
	}, nil
}

// GetToken retrieves a token, automatically refreshing it if needed
func (m *Manager) GetToken(ctx context.Context, resource string) (*Entry, error) {
	// Get the token from storage
	entry, err := m.Storage.Lookup(resource)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("no token found for resource: %s", resource)
	}

	// Create TokenSet if it doesn't exist
	if entry.TokenSet == nil {
		entry.TokenSet = &TokenSet{
			AccessToken:  entry.AccessToken,
			RefreshToken: entry.RefreshToken,
			ExpiresAt:    entry.ExpiresAt,
			Scope:        entry.Scope,
			ResourceID:   entry.Resource,
		}
	}

	// If the token is valid and not close to expiry, return it
	if !entry.TokenSet.IsExpired() && time.Until(entry.TokenSet.ExpiresAt) > m.RefreshThreshold {
		return entry, nil
	}

	// If the token can't be refreshed, return it as-is (it might be valid but close to expiry)
	if !entry.TokenSet.CanRefresh() {
		return entry, nil
	}

	// Refresh the token
	refreshedEntry, err := m.refreshToken(ctx, resource, entry)
	if err != nil {
		// If refresh fails but the token is still valid, return the original token
		if !entry.TokenSet.IsExpired() {
			return entry, nil
		}
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return refreshedEntry, nil
}

// StoreToken stores a token
func (m *Manager) StoreToken(ctx context.Context, entry *Entry) error {
	return m.Storage.Store(entry)
}

// refreshToken refreshes a token and stores it
func (m *Manager) refreshToken(ctx context.Context, resource string, entry *Entry) (*Entry, error) {
	// Use a mutex to prevent multiple simultaneous refreshes for the same token
	m.refreshMutex.Lock()
	defer m.refreshMutex.Unlock()

	// Check if another goroutine already refreshed the token while we were waiting
	latestEntry, err := m.Storage.Lookup(resource)
	if err == nil && latestEntry != nil && latestEntry.AccessToken != entry.AccessToken &&
		latestEntry.TokenSet != nil && !latestEntry.TokenSet.IsExpired() &&
		time.Until(latestEntry.TokenSet.ExpiresAt) > m.RefreshThreshold {
		return latestEntry, nil
	}

	// Refresh the token
	tokenResponse, err := m.RefreshHandler.RefreshToken(ctx, entry.TokenSet.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Calculate expiry time if not set directly in the response
	expiryTime := tokenResponse.ExpiryTime
	if expiryTime.IsZero() {
		expiryTime = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	// Create a new entry with the refreshed values
	refreshedEntry := &Entry{
		Resource:     resource,
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresAt:    expiryTime,
		Scope:        tokenResponse.Scope,
	}

	// If the refresh token wasn't updated, use the original one
	if refreshedEntry.RefreshToken == "" {
		refreshedEntry.RefreshToken = entry.RefreshToken
	}

	// Create TokenSet for convenience
	refreshedEntry.TokenSet = &TokenSet{
		AccessToken:  refreshedEntry.AccessToken,
		RefreshToken: refreshedEntry.RefreshToken,
		ExpiresAt:    refreshedEntry.ExpiresAt,
		Scope:        refreshedEntry.Scope,
		ResourceID:   refreshedEntry.Resource,
	}

	// Store the refreshed token
	if err := m.Storage.Store(refreshedEntry); err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return refreshedEntry, nil
}

// SetRefreshThreshold sets the threshold for automatic refreshing
func (m *Manager) SetRefreshThreshold(threshold time.Duration) {
	m.RefreshThreshold = threshold
}

// StartBackgroundRefresh starts a background goroutine that periodically refreshes tokens
// It returns a stop function that should be called to stop the background refresh
func (m *Manager) StartBackgroundRefresh(refreshInterval time.Duration) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.refreshAllTokens(ctx)
			}
		}
	}()

	return cancel
}

// refreshAllTokens refreshes all tokens that are close to expiry
func (m *Manager) refreshAllTokens(ctx context.Context) {
	// List all tokens
	resources, err := m.Storage.List()
	if err != nil {
		return
	}

	// Refresh each token that needs it
	for _, resource := range resources {
		// Use a timeout for each token refresh
		refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		// Get the token
		entry, err := m.Storage.Lookup(resource)
		if err == nil && entry != nil && entry.TokenSet != nil && entry.TokenSet.CanRefresh() {
			// If the token is close to expiry, refresh it
			if entry.TokenSet.IsExpired() || time.Until(entry.TokenSet.ExpiresAt) < m.RefreshThreshold {
				_, _ = m.refreshToken(refreshCtx, resource, entry) // Ignore errors
			}
		}

		cancel()
	}
}
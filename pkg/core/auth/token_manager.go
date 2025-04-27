// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// RefreshFunc is a function that refreshes a token
type RefreshFunc func(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, expiresAt time.Time, err error)

// TokenManager handles token storage, retrieval, and automatic refreshing
type TokenManager struct {
	Storage          TokenStorage
	RefreshThreshold time.Duration
	RefreshFunc      RefreshFunc
	refreshMutex     sync.Mutex
}

// NewTokenManager creates a new token manager
func NewTokenManager(storage TokenStorage, refreshFunc RefreshFunc) *TokenManager {
	return &TokenManager{
		Storage:          storage,
		RefreshThreshold: 5 * time.Minute, // Default refresh threshold
		RefreshFunc:      refreshFunc,
	}
}

// GetToken retrieves a token, automatically refreshing it if needed
func (tm *TokenManager) GetToken(ctx context.Context, key string) (TokenInfo, error) {
	// Get the token from storage
	token, err := tm.Storage.GetToken(ctx, key)
	if err != nil {
		return TokenInfo{}, err
	}

	// If the token is valid and not close to expiry, return it
	if token.IsValid() && time.Until(token.ExpiresAt) > tm.RefreshThreshold {
		return token, nil
	}

	// If the token can't be refreshed, return it as-is (it might be valid but close to expiry)
	if !token.CanRefresh() {
		return token, nil
	}

	// Refresh the token
	refreshedToken, err := tm.refreshToken(ctx, key, token)
	if err != nil {
		// If refresh fails but the token is still valid, return the original token
		if token.IsValid() {
			return token, nil
		}
		return TokenInfo{}, fmt.Errorf("failed to refresh token: %w", err)
	}

	return refreshedToken, nil
}

// StoreToken stores a token
func (tm *TokenManager) StoreToken(ctx context.Context, key string, token TokenInfo) error {
	return tm.Storage.StoreToken(ctx, key, token)
}

// RefreshToken explicitly refreshes a token
func (tm *TokenManager) RefreshToken(ctx context.Context, key string) (TokenInfo, error) {
	// Get the token from storage
	token, err := tm.Storage.GetToken(ctx, key)
	if err != nil {
		return TokenInfo{}, err
	}

	// If the token can't be refreshed, return an error
	if !token.CanRefresh() {
		return TokenInfo{}, errors.New("token cannot be refreshed (no refresh token)")
	}

	// Refresh the token
	return tm.refreshToken(ctx, key, token)
}

// refreshToken refreshes a token and stores it
func (tm *TokenManager) refreshToken(ctx context.Context, key string, token TokenInfo) (TokenInfo, error) {
	// Use a mutex to prevent multiple simultaneous refreshes for the same token
	tm.refreshMutex.Lock()
	defer tm.refreshMutex.Unlock()

	// Check if another goroutine already refreshed the token while we were waiting
	latestToken, err := tm.Storage.GetToken(ctx, key)
	if err == nil && latestToken.AccessToken != token.AccessToken &&
		latestToken.IsValid() && time.Until(latestToken.ExpiresAt) > tm.RefreshThreshold {
		return latestToken, nil
	}

	// Refresh the token
	accessToken, newRefreshToken, expiresAt, err := tm.RefreshFunc(ctx, token.RefreshToken)
	if err != nil {
		return TokenInfo{}, err
	}

	// Create a new token with the refreshed values
	refreshedToken := TokenInfo{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		Scopes:       token.Scopes,
		ResourceID:   token.ResourceID,
	}

	// If the refresh token wasn't updated, use the original one
	if refreshedToken.RefreshToken == "" {
		refreshedToken.RefreshToken = token.RefreshToken
	}

	// Store the refreshed token
	if err := tm.Storage.StoreToken(ctx, key, refreshedToken); err != nil {
		return TokenInfo{}, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return refreshedToken, nil
}

// SetRefreshThreshold sets the threshold for automatic refreshing
func (tm *TokenManager) SetRefreshThreshold(threshold time.Duration) {
	tm.RefreshThreshold = threshold
}

// StartBackgroundRefresh starts a background goroutine that periodically refreshes tokens
// It returns a stop function that should be called to stop the background refresh
func (tm *TokenManager) StartBackgroundRefresh(refreshInterval time.Duration) func() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tm.refreshAllTokens(ctx)
			}
		}
	}()

	return cancel
}

// refreshAllTokens refreshes all tokens that are close to expiry
func (tm *TokenManager) refreshAllTokens(ctx context.Context) {
	// List all tokens
	keys, err := tm.Storage.ListTokens(ctx)
	if err != nil {
		return
	}

	// Refresh each token that needs it
	for _, key := range keys {
		// Use a timeout for each token refresh
		refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		// Get the token
		token, err := tm.Storage.GetToken(refreshCtx, key)
		if err == nil && token.CanRefresh() {
			// If the token is close to expiry, refresh it
			if time.Until(token.ExpiresAt) < tm.RefreshThreshold {
				_, _ = tm.refreshToken(refreshCtx, key, token) // Ignore errors
			}
		}

		cancel()
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"testing"
	"time"
)

func TestTokenManager(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryTokenStorage()

	// Create a mock refresh function
	refreshCount := 0
	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		refreshCount++
		return "new-access-token", "new-refresh-token", time.Now().Add(time.Hour), nil
	}

	// Create a token manager
	manager := NewTokenManager(storage, refreshFunc)

	// Store a valid token
	validToken := TokenInfo{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	err := manager.StoreToken(ctx, "user1", validToken)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Get the token (should not trigger refresh)
	token, err := manager.GetToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}
	if token.AccessToken != "valid-token" {
		t.Errorf("Got unexpected token: %s, want valid-token", token.AccessToken)
	}
	if refreshCount != 0 {
		t.Errorf("Refresh count = %d, want 0", refreshCount)
	}

	// Store a token that will expire soon
	expiringToken := TokenInfo{
		AccessToken:  "expiring-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(time.Minute),
	}
	err = manager.StoreToken(ctx, "user2", expiringToken)
	if err != nil {
		t.Fatalf("Failed to store expiring token: %v", err)
	}

	// Get the expiring token (should trigger refresh)
	token, err = manager.GetToken(ctx, "user2")
	if err != nil {
		t.Fatalf("Failed to get expiring token: %v", err)
	}
	if token.AccessToken != "new-access-token" {
		t.Errorf("Got unexpected token: %s, want new-access-token", token.AccessToken)
	}
	if refreshCount != 1 {
		t.Errorf("Refresh count = %d, want 1", refreshCount)
	}

	// Store a token without a refresh token
	nonRefreshableToken := TokenInfo{
		AccessToken: "non-refreshable-token",
		ExpiresAt:   time.Now().Add(time.Minute),
	}
	err = manager.StoreToken(ctx, "user3", nonRefreshableToken)
	if err != nil {
		t.Fatalf("Failed to store non-refreshable token: %v", err)
	}

	// Get the non-refreshable token (should not trigger refresh)
	token, err = manager.GetToken(ctx, "user3")
	if err != nil {
		t.Fatalf("Failed to get non-refreshable token: %v", err)
	}
	if token.AccessToken != "non-refreshable-token" {
		t.Errorf("Got unexpected token: %s, want non-refreshable-token", token.AccessToken)
	}
	if refreshCount != 1 {
		t.Errorf("Refresh count = %d, want 1", refreshCount)
	}

	// Manually refresh a token
	token, err = manager.RefreshToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to manually refresh token: %v", err)
	}
	if token.AccessToken != "new-access-token" {
		t.Errorf("Got unexpected token after manual refresh: %s, want new-access-token", token.AccessToken)
	}
	if refreshCount != 2 {
		t.Errorf("Refresh count after manual refresh = %d, want 2", refreshCount)
	}

	// Verify the token is updated in storage
	storedToken, err := storage.GetToken(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get stored token after refresh: %v", err)
	}
	if storedToken.AccessToken != "new-access-token" {
		t.Errorf("Stored token after manual refresh = %s, want new-access-token", storedToken.AccessToken)
	}

	// Try to manually refresh a non-refreshable token (should fail)
	_, err = manager.RefreshToken(ctx, "user3")
	if err == nil {
		t.Error("Expected error when manually refreshing non-refreshable token")
	}
}

func TestBackgroundRefresh(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryTokenStorage()

	// Create a mock refresh function with a channel to track refreshes
	refreshed := make(chan string, 10)
	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		refreshed <- refreshToken
		return "new-access-token", "new-refresh-token", time.Now().Add(time.Hour), nil
	}

	// Create a token manager with a short refresh threshold
	manager := NewTokenManager(storage, refreshFunc)
	manager.SetRefreshThreshold(10 * time.Minute)

	// Store tokens with different expiration times
	validToken := TokenInfo{
		AccessToken:  "valid-token",
		RefreshToken: "refresh-token-1",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	err := manager.StoreToken(ctx, "user1", validToken)
	if err != nil {
		t.Fatalf("Failed to store valid token: %v", err)
	}

	expiringToken := TokenInfo{
		AccessToken:  "expiring-token",
		RefreshToken: "refresh-token-2",
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Will expire soon
	}
	err = manager.StoreToken(ctx, "user2", expiringToken)
	if err != nil {
		t.Fatalf("Failed to store expiring token: %v", err)
	}

	expiredToken := TokenInfo{
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token-3",
		ExpiresAt:    time.Now().Add(-5 * time.Minute), // Already expired
	}
	err = manager.StoreToken(ctx, "user3", expiredToken)
	if err != nil {
		t.Fatalf("Failed to store expired token: %v", err)
	}

	// Start background refresh with a short interval
	stop := manager.StartBackgroundRefresh(50 * time.Millisecond)
	defer stop()

	// Wait for refreshes
	refreshed2 := false
	refreshed3 := false

	// Wait for refreshes (with timeout)
	timeout := time.After(2 * time.Second)
	for {
		select {
		case refreshToken := <-refreshed:
			switch refreshToken {
			case "refresh-token-1":
				// This shouldn't happen as the token is not close to expiry
				t.Error("Valid token should not be refreshed")
			case "refresh-token-2":
				refreshed2 = true
			case "refresh-token-3":
				refreshed3 = true
			}

			// Check if we've seen all expected refreshes
			if refreshed2 && refreshed3 {
				// Success - both expiring and expired tokens were refreshed
				return
			}
		case <-timeout:
			// Check what was refreshed
			t.Logf("Refreshed: user2=%v, user3=%v", refreshed2, refreshed3)

			if !refreshed2 {
				t.Error("Expiring token was not refreshed")
			}
			if !refreshed3 {
				t.Error("Expired token was not refreshed")
			}
			return
		}
	}
}

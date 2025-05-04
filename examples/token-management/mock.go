// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

// MockRefreshHandler implements the tokens.RefreshHandler interface for demonstration purposes
type MockRefreshHandler struct {
	// For tracking refresh calls
	refreshCount int
}

// NewMockRefreshHandler creates a new mock refresh handler
func NewMockRefreshHandler() *MockRefreshHandler {
	return &MockRefreshHandler{
		refreshCount: 0,
	}
}

// RefreshToken implements the tokens.RefreshHandler interface.
// This is a mock implementation that simulates token refreshing.
func (m *MockRefreshHandler) RefreshToken(_ context.Context, refreshToken string) (*auth.TokenResponse, error) {
	m.refreshCount++
	
	// For demonstration purposes, we just generate a new token
	return &auth.TokenResponse{
		AccessToken:  fmt.Sprintf("refreshed-access-token-%d", m.refreshCount),
		RefreshToken: fmt.Sprintf("refreshed-refresh-token-%d", m.refreshCount),
		ExpiresIn:    3600, // 1 hour in seconds
		ExpiryTime:   time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
		Scope:        "openid profile email",
	}, nil
}

// DemonstrateWithMockHandler shows how to use the token manager with a mock handler
func DemonstrateWithMockHandler() {
	fmt.Println("=== Token Manager with Mock RefreshHandler ===")

	// Create a context
	ctx := context.Background()

	// Create a memory storage
	storage := tokens.NewMemoryStorage()

	// Create a mock refresh handler
	mockHandler := NewMockRefreshHandler()

	// Create a token manager with the mock handler using options
	manager, err := tokens.NewManager(
		tokens.WithStorage(storage),
		tokens.WithRefreshHandler(mockHandler),
		tokens.WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create token manager: %v\n", err)
		return
	}
	fmt.Printf("Set refresh threshold to %s\n", manager.RefreshThreshold)

	// Create an example token that's about to expire
	soonToExpireToken := &tokens.TokenSet{
		AccessToken:  "almost-expired-access-token",
		RefreshToken: "sample-refresh-token",
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Will expire soon
		Scope:        "openid profile email",
	}

	entry := &tokens.Entry{
		Resource:     "demo-user",
		AccessToken:  soonToExpireToken.AccessToken,
		RefreshToken: soonToExpireToken.RefreshToken,
		ExpiresAt:    soonToExpireToken.ExpiresAt,
		Scope:        soonToExpireToken.Scope,
		TokenSet:     soonToExpireToken,
	}

	// Store the token
	err := storage.Store(entry)
	if err != nil {
		fmt.Printf("Failed to store token: %v\n", err)
		return
	}
	fmt.Println("Token stored successfully")

	// Get the token - this should trigger a refresh because it's close to expiry
	refreshedEntry, err := manager.GetToken(ctx, "demo-user")
	if err != nil {
		fmt.Printf("Failed to get token: %v\n", err)
		return
	}

	// Check if the token was refreshed
	if refreshedEntry.TokenSet.AccessToken != soonToExpireToken.AccessToken {
		fmt.Println("✅ Token was automatically refreshed!")
		fmt.Printf("Old token: %s\n", soonToExpireToken.AccessToken)
		fmt.Printf("New token: %s\n", refreshedEntry.TokenSet.AccessToken)
		fmt.Printf("New expiry time: %s\n", refreshedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))
	} else {
		fmt.Println("❌ Token was not refreshed as expected")
	}

	// Trigger another refresh, this time explicitly
	beforeToken, err := storage.Lookup("demo-user")
	if err != nil || beforeToken == nil {
		fmt.Println("Failed to lookup token before explicit refresh")
		return
	}

	// Force a refresh by setting a very high threshold
	manager.SetRefreshThreshold(24 * time.Hour)
	afterRefreshEntry, err := manager.GetToken(ctx, "demo-user")
	if err != nil {
		fmt.Printf("Failed to refresh token: %v\n", err)
		return
	}

	if afterRefreshEntry.TokenSet.AccessToken != beforeToken.TokenSet.AccessToken {
		fmt.Println("\n✅ Token was explicitly refreshed with high threshold!")
		fmt.Printf("Previous token: %s\n", beforeToken.TokenSet.AccessToken)
		fmt.Printf("New token: %s\n", afterRefreshEntry.TokenSet.AccessToken)
	} else {
		fmt.Println("\n❌ Explicit refresh did not generate a new token")
	}

	fmt.Println()
}
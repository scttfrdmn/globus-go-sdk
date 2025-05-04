// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// TestTokenRefreshIntegration tests the token refresh functionality with real Globus credentials.
// This test requires the following environment variables:
// - GLOBUS_CLIENT_ID: Globus client ID
// - GLOBUS_CLIENT_SECRET: Globus client secret
func TestTokenRefreshIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create in-memory storage for this test
	storage := NewMemoryStorage()

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager with short refresh threshold for testing
	shortRefreshThreshold := 2 * time.Second
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
		WithRefreshThreshold(shortRefreshThreshold),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsToken(ctx, auth.ScopeOpenID)
	if err != nil {
		t.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	// Store the tokens with a short expiry time to test refresh
	entry := &Entry{
		Resource: "refresh-test",
		TokenSet: &TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			// Set expiry very close to trigger refresh
			ExpiresAt: time.Now().Add(shortRefreshThreshold / 2),
			Scope:     tokenResponse.Scope,
		},
	}

	err = storage.Store(entry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Make sure the token is stored
	storedEntry, err := storage.Lookup("refresh-test")
	if err != nil {
		t.Fatalf("Failed to lookup token: %v", err)
	}
	if storedEntry == nil {
		t.Fatal("Token not found after storing")
	}

	// Original access token
	originalAccessToken := storedEntry.TokenSet.AccessToken

	// Try to get the token, which should trigger a refresh
	refreshedEntry, err := manager.GetToken(ctx, "refresh-test")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Verify the token was refreshed
	if refreshedEntry.TokenSet.AccessToken == originalAccessToken {
		t.Error("Token was not refreshed when it should have been")
	}

	// Verify the refresh token was preserved
	if refreshedEntry.TokenSet.RefreshToken == "" {
		t.Error("Refresh token was lost during refresh")
	}

	// Make sure the refreshed token was stored
	storedRefreshedEntry, err := storage.Lookup("refresh-test")
	if err != nil {
		t.Fatalf("Failed to lookup refreshed token: %v", err)
	}
	if storedRefreshedEntry.TokenSet.AccessToken != refreshedEntry.TokenSet.AccessToken {
		t.Error("Refreshed token was not stored correctly")
	}
}

// TestBackgroundRefreshIntegration tests the background refresh functionality with real Globus credentials.
func TestBackgroundRefreshIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create in-memory storage for this test
	storage := NewMemoryStorage()

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager with short refresh threshold for testing
	shortRefreshThreshold := 2 * time.Second
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
		WithRefreshThreshold(shortRefreshThreshold),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsToken(ctx, auth.ScopeOpenID)
	if err != nil {
		t.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	// Store the tokens with a short expiry time to test refresh
	entry := &Entry{
		Resource: "background-refresh-test",
		TokenSet: &TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			// Set expiry very close to trigger refresh
			ExpiresAt: time.Now().Add(shortRefreshThreshold / 2),
			Scope:     tokenResponse.Scope,
		},
	}

	err = storage.Store(entry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Original access token
	originalAccessToken := entry.TokenSet.AccessToken

	// Start background refresh with a short interval
	stop := manager.StartBackgroundRefresh(1 * time.Second)
	defer stop()

	// Wait for background refresh to occur
	time.Sleep(3 * time.Second)

	// Check if token was refreshed
	refreshedEntry, err := storage.Lookup("background-refresh-test")
	if err != nil {
		t.Fatalf("Failed to lookup token after background refresh: %v", err)
	}

	// Verify token was refreshed
	if refreshedEntry.TokenSet.AccessToken == originalAccessToken {
		t.Error("Token was not refreshed by background process")
	}
}

// TestMultipleTokensRefreshIntegration tests refreshing multiple tokens with real Globus credentials.
func TestMultipleTokensRefreshIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create in-memory storage for this test
	storage := NewMemoryStorage()

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager with short refresh threshold for testing
	shortRefreshThreshold := 2 * time.Second
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
		WithRefreshThreshold(shortRefreshThreshold),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsToken(ctx, auth.ScopeOpenID)
	if err != nil {
		t.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	// Store multiple tokens with short expiry
	tokenResources := []string{"multi-refresh-1", "multi-refresh-2", "multi-refresh-3"}
	originalTokens := make(map[string]string)

	for _, resource := range tokenResources {
		entry := &Entry{
			Resource: resource,
			TokenSet: &TokenSet{
				AccessToken:  tokenResponse.AccessToken,
				RefreshToken: tokenResponse.RefreshToken,
				// Set expiry very close to trigger refresh
				ExpiresAt: time.Now().Add(shortRefreshThreshold / 2),
				Scope:     tokenResponse.Scope,
			},
		}

		err = storage.Store(entry)
		if err != nil {
			t.Fatalf("Failed to store token %s: %v", resource, err)
		}

		originalTokens[resource] = entry.TokenSet.AccessToken
	}

	// Start background refresh with a short interval
	stop := manager.StartBackgroundRefresh(1 * time.Second)
	defer stop()

	// Wait for background refresh to occur
	time.Sleep(5 * time.Second)

	// Check if all tokens were refreshed
	for _, resource := range tokenResources {
		refreshedEntry, err := storage.Lookup(resource)
		if err != nil {
			t.Fatalf("Failed to lookup token %s after background refresh: %v", resource, err)
		}

		if refreshedEntry.TokenSet.AccessToken == originalTokens[resource] {
			t.Errorf("Token %s was not refreshed by background process", resource)
		}
	}
}

// TestGetTokenWithExpiredTokenIntegration tests getting a token with an expired token but no refresh token.
func TestGetTokenWithExpiredTokenIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create in-memory storage for this test
	storage := NewMemoryStorage()

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store an expired token without a refresh token
	expiredEntry := &Entry{
		Resource: "expired-no-refresh",
		TokenSet: &TokenSet{
			AccessToken: "expired-access-token",
			// No refresh token
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Scope:     "test-scope",
		},
	}

	err = storage.Store(expiredEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Try to get the token, which should fail
	_, err = manager.GetToken(ctx, "expired-no-refresh")
	if err == nil {
		t.Error("Expected error when getting expired token without refresh token, but got nil")
	}
}

// TestGetTokenWithNearExpiryTokenIntegration tests getting a token that's close to expiry but has no refresh token.
func TestGetTokenWithNearExpiryTokenIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create in-memory storage for this test
	storage := NewMemoryStorage()

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager with short refresh threshold
	shortRefreshThreshold := 5 * time.Minute
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
		WithRefreshThreshold(shortRefreshThreshold),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Store a token that's close to expiry but has no refresh token
	nearExpiryEntry := &Entry{
		Resource: "near-expiry-no-refresh",
		TokenSet: &TokenSet{
			AccessToken: "near-expiry-access-token",
			// No refresh token
			ExpiresAt: time.Now().Add(1 * time.Minute), // Close to expiry
			Scope:     "test-scope",
		},
	}

	err = storage.Store(nearExpiryEntry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Try to get the token, which should succeed but with the same token
	retrievedEntry, err := manager.GetToken(ctx, "near-expiry-no-refresh")
	if err != nil {
		t.Fatalf("Failed to get near-expiry token: %v", err)
	}

	// Verify token was not refreshed (since it can't be)
	if retrievedEntry.TokenSet.AccessToken != "near-expiry-access-token" {
		t.Error("Token was refreshed when it shouldn't have been")
	}
}

// TestFileStorageIntegration tests token persistence using file storage.
func TestFileStorageIntegration(t *testing.T) {
	// Skip this test if running short tests only
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Skip if credentials are not available
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	// Check for prefixed versions
	if clientID == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: No client credentials found in environment variables")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tokens-file-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager using the WithFileStorage option
	manager, err := NewManager(
		WithFileStorage(tempDir),
		WithAuthClient(authClient),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsToken(ctx, auth.ScopeOpenID)
	if err != nil {
		t.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	// Store the token
	entry := &Entry{
		Resource: "file-storage-test",
		TokenSet: &TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
			Scope:        tokenResponse.Scope,
		},
	}

	err = manager.StoreToken(ctx, entry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Create a new manager with the same file storage to verify persistence
	newManager, err := NewManager(
		WithFileStorage(tempDir),
		WithAuthClient(authClient),
	)
	if err != nil {
		t.Fatalf("Failed to create new token manager: %v", err)
	}

	// Get the token from the new manager
	retrievedEntry, err := newManager.GetToken(ctx, "file-storage-test")
	if err != nil {
		t.Fatalf("Failed to get token from new manager: %v", err)
	}

	// Verify token was retrieved correctly
	if retrievedEntry.TokenSet.AccessToken != tokenResponse.AccessToken {
		t.Error("Retrieved token does not match stored token")
	}
}
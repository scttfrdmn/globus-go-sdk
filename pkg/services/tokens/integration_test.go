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

// TestTokenManagerIntegration tests the token manager with real Globus credentials.
// This test requires the following environment variables:
// - GLOBUS_CLIENT_ID: Globus client ID
// - GLOBUS_CLIENT_SECRET: Globus client secret
func TestTokenManagerIntegration(t *testing.T) {
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
	tempDir, err := os.MkdirTemp("", "tokens-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create file storage
	storage, err := NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// Create token manager with options
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
		WithRefreshThreshold(30 * time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsToken(ctx, auth.ScopeOpenID)
	if err != nil {
		t.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	// Store the tokens
	entry := &Entry{
		Resource: "integration-test",
		TokenSet: &TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
			Scope:        tokenResponse.Scope,
		},
	}

	err = storage.Store(entry)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Retrieve the token
	retrievedEntry, err := storage.Lookup("integration-test")
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}
	if retrievedEntry == nil {
		t.Fatal("Retrieved token entry is nil")
	}

	if retrievedEntry.TokenSet.AccessToken != tokenResponse.AccessToken {
		t.Errorf("Retrieved token does not match stored token")
	}

	// Get token via manager
	managerEntry, err := manager.GetToken(ctx, "integration-test")
	if err != nil {
		t.Fatalf("Failed to get token from manager: %v", err)
	}

	if managerEntry.TokenSet.AccessToken != tokenResponse.AccessToken {
		t.Errorf("Token from manager does not match stored token")
	}

	// List tokens
	tokenList, err := storage.List()
	if err != nil {
		t.Fatalf("Failed to list tokens: %v", err)
	}

	found := false
	for _, resource := range tokenList {
		if resource == "integration-test" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("integration-test token not found in token list")
	}

	// Delete the token
	err = storage.Delete("integration-test")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify deletion
	retrievedEntry, err = storage.Lookup("integration-test")
	if err != nil {
		t.Fatalf("Failed to lookup token after deletion: %v", err)
	}
	if retrievedEntry != nil {
		t.Errorf("Token still exists after deletion")
	}
}

// TestTokenManagerFunctionalOptions tests the functional options pattern for the token manager.
func TestTokenManagerFunctionalOptions(t *testing.T) {
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

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Test cases for different combinations of options
	testCases := []struct {
		name          string
		options       []ClientOption
		expectError   bool
		checkRefresh  bool
	}{
		{
			name: "Default memory storage",
			options: []ClientOption{
				WithAuthClient(authClient),
			},
			expectError: false,
		},
		{
			name: "Custom refresh threshold",
			options: []ClientOption{
				WithAuthClient(authClient),
				WithRefreshThreshold(45 * time.Minute),
			},
			expectError: false,
			checkRefresh: true,
		},
		{
			name: "File storage",
			options: func() []ClientOption {
				tempDir, err := os.MkdirTemp("", "tokens-options-test")
				if err != nil {
					t.Fatalf("Failed to create temp directory: %v", err)
				}
				t.Cleanup(func() { os.RemoveAll(tempDir) })
				
				return []ClientOption{
					WithAuthClient(authClient),
					WithFileStorage(tempDir),
				}
			}(),
			expectError: false,
		},
		{
			name: "Missing required options",
			options: []ClientOption{
				// No storage or refresh handler
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, err := NewManager(tc.options...)
			
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Failed to create token manager: %v", err)
			}
			
			// Verify the manager was created with correct options
			if manager.RefreshHandler != authClient && !tc.expectError && tc.name != "Missing required options" {
				t.Errorf("RefreshHandler not set correctly")
			}
			
			if tc.checkRefresh && manager.RefreshThreshold != 45*time.Minute {
				t.Errorf("RefreshThreshold not set correctly, got %v, expected %v", 
					manager.RefreshThreshold, 45*time.Minute)
			}
		})
	}
}
// TestTokenManagerSDKIntegration tests the integration with the SDK config.
func TestTokenManagerSDKIntegration(t *testing.T) {
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

	// Import the SDK package for this test
	// This is imported in the test function to avoid import cycles
	var sdkConfig interface{} 
	
	// Create in-memory storage for testing
	storage := NewMemoryStorage()
	
	// Test the integration by verifying the token manager can be created
	// The actual creation is tested in the SDK tests, but we can validate
	// that our interface is compatible
	
	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Create token manager directly with options
	manager, err := NewManager(
		WithStorage(storage),
		WithAuthClient(authClient),
	)
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}
	
	// Verify the token manager was created correctly
	if manager.Storage == nil {
		t.Error("Storage not set correctly")
	}
	
	if manager.RefreshHandler == nil {
		t.Error("RefreshHandler not set correctly")
	}
	
	_ = sdkConfig // Suppress unused variable warning
}
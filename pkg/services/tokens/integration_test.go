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
	if clientID == "" || clientSecret == "" {
		t.Skip("Skipping integration test: GLOBUS_CLIENT_ID or GLOBUS_CLIENT_SECRET not set")
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

	// Create auth client
	authClient := auth.NewClient(clientID, clientSecret)

	// Create token manager
	manager := NewManager(storage, authClient)
	manager.SetRefreshThreshold(30 * time.Minute)

	// Get tokens using client credentials flow
	tokenResponse, err := authClient.GetClientCredentialsTokens(ctx, []string{auth.ScopeOpenID})
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
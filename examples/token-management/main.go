// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Check for Globus credentials
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		// Initialize the Globus Auth client
		authClient := auth.NewClient(clientID, clientSecret)
		fmt.Println("Found Globus credentials, demonstrating with real auth client")
		
		// Demonstration of different token storage mechanisms
		demonstrateMemoryStorage(ctx, authClient)
		demonstrateFileStorage(ctx, authClient)
		demonstrateTokenManager(ctx, authClient)
	} else {
		fmt.Println("No Globus credentials found, using mock implementations")
		fmt.Println("To use real Globus authentication, set GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables")
		fmt.Println()

		// Demonstrate with mock handler
		demonstrateMemoryStorage(ctx, nil)
		demonstrateFileStorage(ctx, nil)
		DemonstrateWithMockHandler()
	}
}

// demonstrateMemoryStorage shows how to use the in-memory token storage
func demonstrateMemoryStorage(ctx context.Context, authClient *auth.Client) {
	fmt.Println("=== In-Memory Token Storage ===")

	// Create a memory storage
	storage := tokens.NewMemoryStorage()

	// Create a token entry
	tokenSet := &tokens.TokenSet{
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email",
		ResourceID:   "user-123",
	}

	entry := &tokens.Entry{
		Resource:     "user-123",
		AccessToken:  tokenSet.AccessToken,
		RefreshToken: tokenSet.RefreshToken,
		ExpiresAt:    tokenSet.ExpiresAt,
		Scope:        tokenSet.Scope,
		TokenSet:     tokenSet,
	}

	// Store the token
	err := storage.Store(entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}
	fmt.Println("Token stored successfully")

	// Lookup the token
	retrievedEntry, err := storage.Lookup("user-123")
	if err != nil {
		log.Fatalf("Failed to lookup token: %v", err)
	}
	if retrievedEntry == nil {
		log.Fatal("Token not found")
	}

	fmt.Printf("Retrieved token: %s (expires at: %s)\n",
		retrievedEntry.TokenSet.AccessToken,
		retrievedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))

	// List all tokens
	tokenList, err := storage.List()
	if err != nil {
		log.Fatalf("Failed to list tokens: %v", err)
	}
	fmt.Printf("Token list: %v\n", tokenList)

	// Delete the token
	err = storage.Delete("user-123")
	if err != nil {
		log.Fatalf("Failed to delete token: %v", err)
	}
	fmt.Println("Token deleted successfully")

	// Verify deletion
	retrievedEntry, _ = storage.Lookup("user-123")
	if retrievedEntry != nil {
		log.Fatal("Token was not deleted")
	}
	fmt.Println("Token verified as deleted")
	fmt.Println()
}

// demonstrateFileStorage shows how to use the file-based token storage
func demonstrateFileStorage(ctx context.Context, authClient *auth.Client) {
	fmt.Println("=== File-Based Token Storage ===")

	// Create a temporary directory for token storage
	tempDir, err := os.MkdirTemp("", "globus-tokens")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up afterwards

	fmt.Printf("Using temporary token directory: %s\n", tempDir)

	// Create a file storage
	storage, err := tokens.NewFileStorage(tempDir)
	if err != nil {
		log.Fatalf("Failed to create file storage: %v", err)
	}

	// Create a token entry
	tokenSet := &tokens.TokenSet{
		AccessToken:  "sample-access-token-file",
		RefreshToken: "sample-refresh-token-file",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email",
		ResourceID:   "user-456",
	}

	entry := &tokens.Entry{
		Resource:     "user-456",
		AccessToken:  tokenSet.AccessToken,
		RefreshToken: tokenSet.RefreshToken,
		ExpiresAt:    tokenSet.ExpiresAt,
		Scope:        tokenSet.Scope,
		TokenSet:     tokenSet,
	}

	// Store the token
	err = storage.Store(entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}
	fmt.Println("Token stored successfully to file")

	// Lookup the token
	retrievedEntry, err := storage.Lookup("user-456")
	if err != nil {
		log.Fatalf("Failed to lookup token: %v", err)
	}
	if retrievedEntry == nil {
		log.Fatal("Token not found")
	}

	fmt.Printf("Retrieved token from file: %s (expires at: %s)\n",
		retrievedEntry.TokenSet.AccessToken,
		retrievedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))

	// List all tokens
	tokenList, err := storage.List()
	if err != nil {
		log.Fatalf("Failed to list tokens: %v", err)
	}
	fmt.Printf("Token list from file storage: %v\n", tokenList)

	// Verify the token file exists
	files, err := os.ReadDir(tempDir)
	if err != nil {
		log.Fatalf("Failed to read temp directory: %v", err)
	}
	fmt.Printf("Files in token directory: %d\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file.Name())
	}

	// Delete the token
	err = storage.Delete("user-456")
	if err != nil {
		log.Fatalf("Failed to delete token: %v", err)
	}
	fmt.Println("Token file deleted successfully")
	fmt.Println()
}

// demonstrateTokenManager shows how to use the token manager for automatic refreshing
func demonstrateTokenManager(ctx context.Context, authClient *auth.Client) {
	fmt.Println("=== Token Manager with Automatic Refresh ===")

	// Create a memory storage
	storage := tokens.NewMemoryStorage()

	// Create a token manager
	manager := tokens.NewManager(storage, authClient)

	// Set refresh threshold (when to refresh tokens)
	manager.SetRefreshThreshold(30 * time.Minute)
	fmt.Printf("Set refresh threshold to %s\n", manager.RefreshThreshold)

	// Create an example token that's about to expire
	soonToExpireToken := &tokens.TokenSet{
		AccessToken:  "soon-to-expire-token",
		RefreshToken: "refresh-token-for-expiring-token",
		ExpiresAt:    time.Now().Add(5 * time.Minute), // Will expire soon
		Scope:        "openid profile email",
		ResourceID:   "user-789",
	}

	entry := &tokens.Entry{
		Resource:     "user-789",
		AccessToken:  soonToExpireToken.AccessToken,
		RefreshToken: soonToExpireToken.RefreshToken,
		ExpiresAt:    soonToExpireToken.ExpiresAt,
		Scope:        soonToExpireToken.Scope,
		TokenSet:     soonToExpireToken,
	}

	// Store the token
	err := storage.Store(entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}
	fmt.Println("Token stored successfully")

	// In a real application, you would now call manager.GetToken() which would
	// automatically refresh the token if it's close to expiry
	fmt.Println("In a real application with credentials, GetToken() would:")
	fmt.Println("1. Check if the token is close to expiry")
	fmt.Println("2. Automatically refresh it using the auth client if needed")
	fmt.Println("3. Store the refreshed token")
	fmt.Println("4. Return the refreshed token")

	// Start the background refresh
	fmt.Println("\nStarting background refresh (would run every 15 minutes)")
	stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)

	// In a real application, you would let this run until the application exits
	// For demonstration, we'll just stop it immediately
	fmt.Println("For demonstration, stopping background refresh immediately")
	stopRefresh()

	fmt.Println("\nToken Manager usage notes:")
	fmt.Println("- Use TokenManager in long-running applications")
	fmt.Println("- StartBackgroundRefresh() returns a function to stop the refresh")
	fmt.Println("- Call the stop function with defer when your application exits")
	fmt.Println("- Set an appropriate RefreshThreshold for your use case")
	fmt.Println()
}
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	// Parse lines
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Skip comments and empty lines
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		// Set environment variable
		os.Setenv(key, value)
	}

	return nil
}

func main() {
	fmt.Println("=== Testing Tokens Package with Real Credentials ===")

	// Find the .env.test file (starting from the current directory and going up)
	envPath := ""
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Try current directory first
	if _, err := os.Stat(".env.test"); err == nil {
		envPath = ".env.test"
	} else {
		// Try to find the file in parent directories
		for dir := currentDir; dir != "/"; dir = filepath.Dir(dir) {
			path := filepath.Join(dir, ".env.test")
			if _, err := os.Stat(path); err == nil {
				envPath = path
				break
			}
		}
	}

	if envPath == "" {
		log.Fatal("Could not find .env.test file. Please ensure it exists in the project root or current directory.")
	}

	fmt.Printf("Using credentials from: %s\n", envPath)

	// Load environment variables from .env.test
	err = loadEnvFile(envPath)
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	// Get credentials
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Missing GLOBUS_CLIENT_ID or GLOBUS_CLIENT_SECRET in .env.test file")
	}

	fmt.Printf("Found credentials for client ID: %s\n", clientID)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create temporary directory for token storage
	tempDir, err := os.MkdirTemp("", "globus-tokens-test")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fmt.Printf("Using temporary token directory: %s\n", tempDir)

	// Create a file storage
	storage, err := tokens.NewFileStorage(tempDir)
	if err != nil {
		log.Fatalf("Failed to create file storage: %v", err)
	}

	// Create an auth client with the credentials
	authClient := auth.NewClient(clientID, clientSecret)

	// Test client credentials flow to get tokens
	fmt.Println("\nTesting client credentials flow...")
	
	// Get tokens using client credentials
	tokenResponse, err := authClient.GetClientCredentialsTokens(ctx, []string{auth.ScopeOpenID})
	if err != nil {
		log.Fatalf("Failed to get client credentials tokens: %v", err)
	}

	fmt.Printf("Successfully obtained tokens for client\n")
	fmt.Printf("Access token: %s...\n", tokenResponse.AccessToken[:20])
	fmt.Printf("Expires in: %d seconds\n", tokenResponse.ExpiresIn)
	fmt.Printf("Scope: %s\n", tokenResponse.Scope)

	// Create token manager
	manager := tokens.NewManager(storage, authClient)
	manager.SetRefreshThreshold(30 * time.Minute)

	// Store the tokens
	entry := &tokens.Entry{
		Resource: "client-credentials",
		TokenSet: &tokens.TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken, // Usually empty for client credentials
			ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
			Scope:        tokenResponse.Scope,
		},
	}

	err = storage.Store(entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}
	fmt.Println("Successfully stored token")

	// Retrieve the token
	retrievedEntry, err := storage.Lookup("client-credentials")
	if err != nil {
		log.Fatalf("Failed to retrieve token: %v", err)
	}
	if retrievedEntry == nil {
		log.Fatal("Retrieved token entry is nil")
	}

	fmt.Printf("Successfully retrieved token: %s...\n", retrievedEntry.TokenSet.AccessToken[:20])
	fmt.Printf("Expires at: %s\n", retrievedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))

	// List tokens
	tokenList, err := storage.List()
	if err != nil {
		log.Fatalf("Failed to list tokens: %v", err)
	}
	fmt.Printf("Token list: %v\n", tokenList)

	// Get token via manager (should return without refreshing since it's not close to expiry)
	managerEntry, err := manager.GetToken(ctx, "client-credentials")
	if err != nil {
		log.Fatalf("Failed to get token from manager: %v", err)
	}

	fmt.Printf("Successfully got token from manager: %s...\n", managerEntry.TokenSet.AccessToken[:20])

	// If the token has a refresh token, test refreshing
	if tokenResponse.RefreshToken != "" {
		fmt.Println("\nTesting token refresh...")
		
		// Create a token that's close to expiry
		expiringEntry := &tokens.Entry{
			Resource: "expiring-token",
			TokenSet: &tokens.TokenSet{
				AccessToken:  tokenResponse.AccessToken,
				RefreshToken: tokenResponse.RefreshToken,
				ExpiresAt:    time.Now().Add(5 * time.Minute), // Close to expiry
				Scope:        tokenResponse.Scope,
			},
		}

		err = storage.Store(expiringEntry)
		if err != nil {
			log.Fatalf("Failed to store expiring token: %v", err)
		}

		// Set a short refresh threshold
		manager.SetRefreshThreshold(30 * time.Minute)

		// Get the token (should trigger a refresh)
		refreshedEntry, err := manager.GetToken(ctx, "expiring-token")
		if err != nil {
			log.Fatalf("Failed to refresh token: %v", err)
		}

		if refreshedEntry.TokenSet.AccessToken != expiringEntry.TokenSet.AccessToken {
			fmt.Println("✅ Token was successfully refreshed!")
			fmt.Printf("Old token: %s...\n", expiringEntry.TokenSet.AccessToken[:20])
			fmt.Printf("New token: %s...\n", refreshedEntry.TokenSet.AccessToken[:20])
		} else {
			fmt.Println("❌ Token was not refreshed as expected")
		}
	} else {
		fmt.Println("\nSkipping refresh test as client credentials tokens don't have refresh tokens")
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("The tokens package is working correctly with real credentials!")
}
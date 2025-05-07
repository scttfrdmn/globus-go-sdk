// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create SDK config
	config := pkg.NewConfig()

	// Add client ID and secret from environment variables
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables must be set")
		fmt.Println("Using mock implementations for demonstration")
		demonstrateWithMock()
		return
	}

	// Create auth client with options
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
		auth.WithHTTPDebugging(os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"),
	)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Println("=== Token Manager with Functional Options ===")

	// Create a token storage directory in temp folder
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	tokenDir := filepath.Join(homeDir, ".globus-token-example")
	err = os.MkdirAll(tokenDir, 0700)
	if err != nil {
		log.Fatalf("Failed to create token directory: %v", err)
	}

	fmt.Printf("Using token storage directory: %s\n", tokenDir)

	// Example 1: Create token manager with individual options
	fmt.Println("\n1. Creating token manager with individual options")
	manager1, err := tokens.NewManager(
		tokens.WithFileStorage(tokenDir),
		tokens.WithAuthClient(authClient),
		tokens.WithRefreshThreshold(15*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created with refresh threshold: %s\n", manager1.RefreshThreshold)

	// Example 2: Create token manager using SDK helper method
	fmt.Println("\n2. Creating token manager using SDK helper method")
	manager2, err := config.WithClientID(clientID).
		WithClientSecret(clientSecret).
		NewTokenManagerWithAuth(tokenDir)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created with storage directory: %s\n", tokenDir)

	// Example 3: In-memory storage with custom options
	fmt.Println("\n3. Creating token manager with in-memory storage")
	memoryStorage := tokens.NewMemoryStorage()

	manager3, err := tokens.NewManager(
		tokens.WithStorage(memoryStorage),
		tokens.WithAuthClient(authClient),
		tokens.WithRefreshThreshold(30*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created with in-memory storage and refresh threshold: %s\n",
		manager3.RefreshThreshold)

	// Store a sample token
	sampleToken := &tokens.TokenSet{
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "openid profile email",
		ResourceID:   "example-resource",
	}

	entry := &tokens.Entry{
		Resource:     "example-resource",
		AccessToken:  sampleToken.AccessToken,
		RefreshToken: sampleToken.RefreshToken,
		ExpiresAt:    sampleToken.ExpiresAt,
		Scope:        sampleToken.Scope,
		TokenSet:     sampleToken,
	}

	// Store the token
	err = manager3.StoreToken(ctx, entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}

	fmt.Println("Token successfully stored")

	// Get the token
	retrievedEntry, err := manager3.GetToken(ctx, "example-resource")
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Retrieved token: %s (expires at: %s)\n",
		retrievedEntry.TokenSet.AccessToken,
		retrievedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))

	// Example 4: Start background refresh
	fmt.Println("\n4. Starting background token refresh")
	stopRefresh := manager3.StartBackgroundRefresh(15 * time.Minute)
	defer stopRefresh() // Stop the refresh when the application exits

	fmt.Println("Background refresh started (interval: 15 minutes)")
	fmt.Println("In a real application, this would continue running")
	fmt.Println("For demonstration, we'll stop it with defer when the program ends")

	fmt.Println("\n=== End of Token Manager Demonstration ===")
}

// Mock implementation for demonstration without credentials
func demonstrateWithMock() {
	ctx := context.Background()

	fmt.Println("=== Token Manager with Mock Implementation ===")

	// Create a mock refresh handler
	mockHandler := &MockRefreshHandler{refreshCount: 0}

	// Create an in-memory storage
	storage := tokens.NewMemoryStorage()

	// Create a token manager with options
	manager, err := tokens.NewManager(
		tokens.WithStorage(storage),
		tokens.WithRefreshHandler(mockHandler),
		tokens.WithRefreshThreshold(15*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	fmt.Printf("Token manager created with refresh threshold: %s\n", manager.RefreshThreshold)

	// Create a sample token that will expire soon
	sampleToken := &tokens.TokenSet{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(10 * time.Minute), // Will expire soon
		Scope:        "openid profile email",
		ResourceID:   "mock-resource",
	}

	entry := &tokens.Entry{
		Resource:     "mock-resource",
		AccessToken:  sampleToken.AccessToken,
		RefreshToken: sampleToken.RefreshToken,
		ExpiresAt:    sampleToken.ExpiresAt,
		Scope:        sampleToken.Scope,
		TokenSet:     sampleToken,
	}

	// Store the token
	err = storage.Store(entry)
	if err != nil {
		log.Fatalf("Failed to store token: %v", err)
	}

	fmt.Println("Token stored successfully")

	// Get the token - this should trigger a refresh since it's close to expiry
	refreshedEntry, err := manager.GetToken(ctx, "mock-resource")
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	// Check if the token was refreshed
	if refreshedEntry.TokenSet.AccessToken != sampleToken.AccessToken {
		fmt.Println("✅ Token was automatically refreshed!")
		fmt.Printf("Old token: %s\n", sampleToken.AccessToken)
		fmt.Printf("New token: %s\n", refreshedEntry.TokenSet.AccessToken)
		fmt.Printf("New expiry time: %s\n", refreshedEntry.TokenSet.ExpiresAt.Format(time.RFC3339))
	} else {
		fmt.Println("❌ Token was not refreshed as expected")
	}

	fmt.Println("\n=== End of Mock Demonstration ===")
}

// MockRefreshHandler implements the tokens.RefreshHandler interface for demonstration
type MockRefreshHandler struct {
	refreshCount int
}

// RefreshToken implements the tokens.RefreshHandler interface
func (m *MockRefreshHandler) RefreshToken(_ context.Context, refreshToken string) (*auth.TokenResponse, error) {
	m.refreshCount++

	return &auth.TokenResponse{
		AccessToken:  fmt.Sprintf("refreshed-token-%d", m.refreshCount),
		RefreshToken: fmt.Sprintf("refreshed-refresh-token-%d", m.refreshCount),
		ExpiresIn:    3600, // 1 hour in seconds
		ExpiryTime:   time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
		Scope:        "openid profile email",
	}, nil
}

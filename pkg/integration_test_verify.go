//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// TestIntegration_VerifySetup verifies that the test environment is correctly configured
// This is a helper test that should be run before other integration tests to validate
// that the test credentials and resources are properly set up
func TestIntegration_VerifySetup(t *testing.T) {
	// Get required credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	// Check for required credentials
	if clientID == "" || clientSecret == "" {
		t.Fatal("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}

	fmt.Println("✅ Found required credentials")

	// Create SDK config
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Verify Auth credentials
	fmt.Println("Verifying authentication credentials...")
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	ctx := context.Background()

	// Try to get client credentials token
	token, err := authClient.GetClientCredentialsToken(ctx, "openid email profile")
	if err != nil {
		t.Fatalf("Failed to get client credentials token: %v", err)
	}

	fmt.Printf("✅ Successfully obtained client credentials token (expires: %s)\n",
		token.ExpiresAt().Format(time.RFC3339))

	// Introspect the token
	tokenInfo, err := authClient.IntrospectToken(ctx, token.AccessToken)
	if err != nil {
		t.Fatalf("Failed to introspect token: %v", err)
	}

	if !tokenInfo.Active {
		t.Fatal("Token is not active")
	}

	fmt.Println("✅ Token introspection successful")
	fmt.Printf("   Token scopes: %s\n", tokenInfo.Scope)

	// Check for transfer endpoints if specified
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DESTINATION_ENDPOINT_ID")

	if sourceEndpointID != "" && destEndpointID != "" {
		fmt.Println("Verifying transfer endpoints...")
		verifyTransferEndpoints(t, config, sourceEndpointID, destEndpointID)
	} else {
		fmt.Println("⚠️  Transfer endpoint IDs not specified; skipping transfer endpoint verification")
		fmt.Println("   Set GLOBUS_TEST_SOURCE_ENDPOINT_ID and GLOBUS_TEST_DEST_ENDPOINT_ID to enable transfer tests")
	}

	// Check for group ID if specified
	groupID := os.Getenv("GLOBUS_TEST_GROUP_ID")
	if groupID != "" {
		fmt.Printf("✅ Group ID specified: %s\n", groupID)
	} else {
		fmt.Println("⚠️  GLOBUS_TEST_GROUP_ID not specified; some group tests may be skipped")
	}

	// Check for user ID if specified
	userID := os.Getenv("GLOBUS_TEST_USER_ID")
	if userID != "" {
		fmt.Printf("✅ User ID specified: %s\n", userID)
	} else {
		fmt.Println("⚠️  GLOBUS_TEST_USER_ID not specified; some membership tests may be skipped")
	}

	fmt.Println("✅ Test environment verification complete")
	fmt.Println("🚀 You are ready to run integration tests!")
}

// verifyTransferEndpoints checks that the specified endpoints are accessible
func verifyTransferEndpoints(t *testing.T, config *Config, sourceEndpointID, destEndpointID string) {
	ctx := context.Background()

	// We need a token with transfer scope
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	token, err := authClient.GetClientCredentialsToken(ctx, TransferScope)
	if err != nil {
		t.Fatalf("Failed to get token with transfer scope: %v", err)
	}

	// Create transfer client
	transferClient := config.NewTransferClient(token.AccessToken)

	// Verify source endpoint
	fmt.Printf("Checking source endpoint (%s)...\n", sourceEndpointID)
	sourceEndpoint, err := transferClient.GetEndpoint(ctx, sourceEndpointID)
	if err != nil {
		t.Fatalf("Failed to get source endpoint: %v", err)
	}
	fmt.Printf("✅ Source endpoint accessible: %s\n", sourceEndpoint.DisplayName)

	// Verify destination endpoint
	fmt.Printf("Checking destination endpoint (%s)...\n", destEndpointID)
	destEndpoint, err := transferClient.GetEndpoint(ctx, destEndpointID)
	if err != nil {
		t.Fatalf("Failed to get destination endpoint: %v", err)
	}
	fmt.Printf("✅ Destination endpoint accessible: %s\n", destEndpoint.DisplayName)

	// Check if endpoints are activated
	sourceActivated, err := transferClient.EndpointIsActivated(ctx, sourceEndpointID)
	if err != nil {
		fmt.Printf("⚠️  Couldn't check if source endpoint is activated: %v\n", err)
	} else if !sourceActivated {
		fmt.Println("⚠️  Source endpoint is not activated. Some transfer tests may fail.")
		fmt.Println("   Activate the endpoint in the Globus web interface or using the SDK.")
	} else {
		fmt.Println("✅ Source endpoint is activated")
	}

	destActivated, err := transferClient.EndpointIsActivated(ctx, destEndpointID)
	if err != nil {
		fmt.Printf("⚠️  Couldn't check if destination endpoint is activated: %v\n", err)
	} else if !destActivated {
		fmt.Println("⚠️  Destination endpoint is not activated. Some transfer tests may fail.")
		fmt.Println("   Activate the endpoint in the Globus web interface or using the SDK.")
	} else {
		fmt.Println("✅ Destination endpoint is activated")
	}

	// Check test paths if specified
	sourcePath := os.Getenv("GLOBUS_TEST_SOURCE_PATH")
	if sourcePath == "" {
		sourcePath = "/globus-test"
		fmt.Printf("GLOBUS_TEST_SOURCE_PATH not specified, using default: %s\n", sourcePath)
	}

	destPath := os.Getenv("GLOBUS_TEST_DEST_PATH")
	if destPath == "" {
		destPath = "/globus-test"
		fmt.Printf("GLOBUS_TEST_DEST_PATH not specified, using default: %s\n", destPath)
	}

	// Verify source path exists
	_, err = transferClient.ListDirectoryContents(ctx, sourceEndpointID, sourcePath, &transfer.ListOptions{})
	if err != nil {
		fmt.Printf("⚠️  Could not access source path %s: %v\n", sourcePath, err)
		fmt.Println("   Consider creating this directory for transfer tests")
	} else {
		fmt.Printf("✅ Source path exists: %s\n", sourcePath)
	}

	// Verify destination path exists
	_, err = transferClient.ListDirectoryContents(ctx, destEndpointID, destPath, &transfer.ListOptions{})
	if err != nil {
		fmt.Printf("⚠️  Could not access destination path %s: %v\n", destPath, err)
		fmt.Println("   Consider creating this directory for transfer tests")
	} else {
		fmt.Printf("✅ Destination path exists: %s\n", destPath)
	}
}

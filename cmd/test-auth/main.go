// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

func main() {
	// Get credentials from environment
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET environment variables must be set")
	}

	// Create a client
	client, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Test client credentials flow
	ctx := context.Background()
	tokenResp, err := client.GetClientCredentialsToken(ctx, auth.AuthScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Successfully got token: %s... (expires in %d seconds)\n", 
		tokenResp.AccessToken[:10], tokenResp.ExpiresIn)
}
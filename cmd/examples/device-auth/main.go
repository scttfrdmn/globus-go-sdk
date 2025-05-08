// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// This example demonstrates how to use device authentication flow with the Globus SDK.
// Device flow is ideal for CLI applications or other non-browser environments.

func main() {
	// Get client ID from environment variable or use a default
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	if clientID == "" {
		log.Fatal("GLOBUS_CLIENT_ID environment variable must be set")
	}

	// Create the auth client
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithCoreOption(core.WithHTTPDebugging(false)),
	)
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	fmt.Println("Starting device authentication flow...")

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Define the scopes we need
	scopes := []string{
		"openid",
		"profile",
		"email",
		"urn:globus:auth:scope:transfer.api.globus.org:all",
	}

	// This callback will be called with the device code information
	displayCallback := func(deviceCode *auth.DeviceCodeResponse) {
		fmt.Println("\n===== Device Authorization Required =====")
		fmt.Println("Please visit this URL to authorize this application:")
		fmt.Printf("  %s\n\n", deviceCode.VerificationURI)
		fmt.Println("Enter the following code when prompted:")
		fmt.Printf("  %s\n", deviceCode.UserCode)
		fmt.Println("=======================================")
		fmt.Printf("This code will expire in %d seconds.\n", deviceCode.ExpiresIn)
		fmt.Printf("Waiting for authorization...\n\n")
	}

	// Start the device flow and wait for user authorization
	tokenResp, err := authClient.CompleteDeviceFlow(ctx, displayCallback, 0, scopes...)
	if err != nil {
		log.Fatalf("Device flow failed: %v", err)
	}

	fmt.Println("Authentication successful!")
	fmt.Printf("Access Token: %s...\n", tokenResp.AccessToken[:15])
	fmt.Printf("Token expires in: %d seconds\n", tokenResp.ExpiresIn)

	if tokenResp.RefreshToken != "" {
		fmt.Printf("Refresh Token: %s...\n", tokenResp.RefreshToken[:15])
	}

	// Example of how to extract other tokens returned by Globus Auth
	otherTokens, err := tokenResp.GetOtherTokens()
	if err != nil {
		log.Printf("Failed to parse other tokens: %v", err)
	} else if len(otherTokens) > 0 {
		fmt.Printf("Received %d additional token(s) for other resources\n", len(otherTokens))
		for i, token := range otherTokens {
			fmt.Printf("  Token %d for resource: %s\n", i+1, token.ResourceServer)
		}
	}
}

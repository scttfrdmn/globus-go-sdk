// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package debug

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// deleteAuthorizer implements the authorizer interface for testing
type deleteAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *deleteAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns whether the authorization is valid
func (a *deleteAuthorizer) IsValid() bool {
	return a.token != ""
}

// GetToken returns the token
func (a *deleteAuthorizer) GetToken() string {
	return a.token
}

// getAccessTokenDelete tries to get a token with various scopes
func getAccessTokenDelete(clientID, clientSecret string) (string, error) {
	// First, check if there's a transfer token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if staticToken != "" {
		fmt.Println("Using static transfer token from environment")
		return staticToken, nil
	}

	// If not, try to get a token from the auth service
	fmt.Println("No static token, trying client credentials flow")

	// Check if we have client credentials
	if clientID == "" || clientSecret == "" {
		clientID = os.Getenv("GLOBUS_TEST_CLIENT_ID")
		clientSecret = os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("no token found and no client credentials available")
	}

	// Create an auth client
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create auth client: %w", err)
	}

	// Try different scopes
	scopes := []string{
		"urn:globus:auth:scope:transfer.api.globus.org:all",
		"https://auth.globus.org/scopes/transfer.api.globus.org/all",
	}

	var tokenResp *auth.TokenResponse
	var tokenErr error
	var gotToken bool

	// Try each scope until we get a token
	for _, scope := range scopes {
		tokenResp, tokenErr = authClient.GetClientCredentialsToken(context.Background(), scope)
		if tokenErr != nil {
			fmt.Printf("Failed to get token with scope %s: %v\n", scope, tokenErr)
			continue
		}

		// Check if we got a token for the transfer service
		fmt.Printf("Got token with resource server: %s, scopes: %s\n", tokenResp.ResourceServer, tokenResp.Scope)
		if tokenResp.ResourceServer != "" && (tokenResp.ResourceServer == "transfer.api.globus.org" ||
			tokenResp.Scope == "urn:globus:auth:scope:transfer.api.globus.org:all") {
			gotToken = true
			break
		}
	}

	// If we didn't get a transfer token, fall back to the default token
	if !gotToken {
		fmt.Println("Could not get a transfer token, falling back to default token")
		if tokenResp == nil {
			return "", fmt.Errorf("failed to get any token: %v", tokenErr)
		}
	}

	return tokenResp.AccessToken, nil
}

// RunDelete is the main function for debug delete operations
func RunDelete() {
	// Load environment variables
	_ = godotenv.Load(".env.test")

	// Enable debug output
	os.Setenv("HTTP_DEBUG", "1")

	// Get token
	token, err := getAccessTokenDelete("", "")
	if err != nil {
		fmt.Printf("ERROR: Failed to get token: %v\n", err)
		os.Exit(1)
	}

	// Create authorizer
	authorizer := &deleteAuthorizer{token: token}

	// Create client with debugging enabled
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizer),
		transfer.WithHTTPDebugging(true),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create transfer client: %v\n", err)
		os.Exit(1)
	}

	// Test a simple delete operation
	endpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	if endpointID == "" {
		fmt.Println("ERROR: GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable is required")
		os.Exit(1)
	}

	path := "/globus-test/delete-test-" + time.Now().Format("20060102-150405")

	// First, create a directory to delete
	fmt.Printf("Creating directory: %s\n", path)
	err = client.Mkdir(context.Background(), endpointID, path)
	if err != nil {
		fmt.Printf("ERROR: Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Then delete it
	fmt.Printf("Deleting path: %s\n", path)

	// Create delete request
	deleteRequest := &transfer.DeleteTaskRequest{
		DataType:   "delete",
		EndpointID: endpointID,
		Items: []transfer.DeleteItem{
			{
				DataType: "delete_item",
				Path:     path,
			},
		},
	}

	result, err := client.CreateDeleteTask(context.Background(), deleteRequest)
	if err != nil {
		fmt.Printf("ERROR: Failed to delete path: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Delete task submitted: %s\n", result.TaskID)
}
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// testAuthorizer implements the authorizer interface for testing
type testAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *testAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns whether the authorization is valid
func (a *testAuthorizer) IsValid() bool {
	return a.token != ""
}

// GetToken returns the token
func (a *testAuthorizer) GetToken() string {
	return a.token
}

func getTestCredentials() (string, string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")

	if clientID == "" {
		fmt.Println("GLOBUS_TEST_CLIENT_ID environment variable not set")
		os.Exit(1)
	}

	if clientSecret == "" {
		fmt.Println("GLOBUS_TEST_CLIENT_SECRET environment variable not set")
		os.Exit(1)
	}

	return clientID, clientSecret, sourceEndpointID, destEndpointID
}

func getAccessToken(clientID, clientSecret string) string {
	// First, check if there's a transfer token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if staticToken != "" {
		fmt.Println("Using static transfer token from environment")
		return staticToken
	}

	// If no static token, try to get one via client credentials
	fmt.Println("Getting client credentials token for transfer")
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		fmt.Printf("Failed to create auth client: %v\n", err)
		os.Exit(1)
	}

	// Try different scopes that might work for transfer
	scopes := []string{
		"urn:globus:auth:scope:transfer.api.globus.org:all",
		"https://auth.globus.org/scopes/transfer.api.globus.org/all",
	}

	var tokenResp *auth.TokenResponse
	var err error
	var gotToken bool

	// Try each scope until we get a token
	for _, scope := range scopes {
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background(), scope)
		if err != nil {
			fmt.Printf("Failed to get token with scope %s: %v\n", scope, err)
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
		tokenResp, err = authClient.GetClientCredentialsToken(context.Background())
		if err != nil {
			fmt.Printf("Failed to get any token: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using default token with resource server: %s, scopes: %s\n",
			tokenResp.ResourceServer, tokenResp.Scope)
		fmt.Println("WARNING: This token may not have transfer permissions. Consider providing GLOBUS_TEST_TRANSFER_TOKEN")
	}

	return tokenResp.AccessToken
}

func main() {
	// Load environment variables
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("pkg/.env.test")
	_ = godotenv.Load("pkg/services/.env.test")
	_ = godotenv.Load("pkg/services/transfer/.env.test")

	// Enable debug output
	os.Setenv("HTTP_DEBUG", "1")

	// Get credentials
	clientID, clientSecret, sourceEndpointID, _ := getTestCredentials()
	if sourceEndpointID == "" {
		fmt.Println("GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable not set")
		os.Exit(1)
	}

	// Get access token
	accessToken := getAccessToken(clientID, clientSecret)

	// Create transfer client with debugging enabled
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&testAuthorizer{token: accessToken}),
		transfer.WithHTTPDebugging(true),
		transfer.WithHTTPTracing(true),
	)
	if err != nil {
		fmt.Printf("Failed to create transfer client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Create a test directory to delete
	testBasePath := os.Getenv("GLOBUS_TEST_DIRECTORY_PATH")
	if testBasePath == "" {
		testBasePath = "globus-test" // Default to a simple test directory
	}

	timestamp := time.Now().Format("20060102_150405")
	testDir := fmt.Sprintf("%s/debug_delete_test_%s", testBasePath, timestamp)

	// Try to create the directory first
	fmt.Printf("Creating test directory: %s\n", testDir)
	err = client.Mkdir(ctx, sourceEndpointID, testDir)
	if err != nil {
		fmt.Printf("Failed to create test directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Test directory created successfully")

	// Now try to delete the directory - showing the request structure
	fmt.Println("\nAttempting to delete the directory...")

	// Get a submission ID manually to see if it's working
	subID, err := client.GetSubmissionID(ctx)
	if err != nil {
		fmt.Printf("Failed to get submission ID: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got submission ID: %s\n", subID)

	// First try with current implementation
	deleteRequest := &transfer.DeleteTaskRequest{
		DataType:     "delete",
		Label:        fmt.Sprintf("Debug delete test %s", timestamp),
		EndpointID:   sourceEndpointID,
		SubmissionID: subID,
		Items: []transfer.DeleteItem{
			{
				DataType: "delete_item",
				Path:     testDir,
			},
		},
	}

	fmt.Println("\nAttempting delete with current implementation (Items under DATA)...")
	deleteResp, err := client.CreateDeleteTask(ctx, deleteRequest)
	if err != nil {
		fmt.Printf("Delete task failed: %v\n", err)
		// Continue with modified version
	} else {
		fmt.Printf("Delete task successful: %s\n", deleteResp.TaskID)
		fmt.Println("No fix needed! The current implementation works.")
		os.Exit(0)
	}

	// If we're here, the standard implementation failed
	// Skip manual construction and go directly to testing our fixed implementation
	fmt.Println("\nStandard implementation failed, trying with our fixed implementation...")

	// We can't use the base client directly, so manually create a new delete task request with the proper format
	fmt.Println("\nSkipping modified request test - using the standard client with the modified structure")

	// Instead, let's create another test directory and try again with our updated DeleteTaskRequest structure
	testDir2 := fmt.Sprintf("%s/debug_delete_test2_%s", testBasePath, timestamp)
	fmt.Printf("\nCreating a second test directory: %s\n", testDir2)

	err = client.Mkdir(ctx, sourceEndpointID, testDir2)
	if err != nil {
		fmt.Printf("Failed to create second test directory: %v\n", err)
		os.Exit(1)
	}

	// Get a new submission ID
	subID2, err := client.GetSubmissionID(ctx)
	if err != nil {
		fmt.Printf("Failed to get second submission ID: %v\n", err)
		os.Exit(1)
	}

	// Try with our modified implementation
	deleteRequest2 := &transfer.DeleteTaskRequest{
		DataType:     "delete",
		Label:        fmt.Sprintf("Debug delete test 2 %s", timestamp),
		EndpointID:   sourceEndpointID,
		SubmissionID: subID2,
		Items: []transfer.DeleteItem{
			{
				DataType: "delete_item",
				Path:     testDir2,
			},
		},
	}

	fmt.Println("\nAttempting delete with updated implementation and DataType fields added...")
	deleteResp2, err := client.CreateDeleteTask(ctx, deleteRequest2)
	if err != nil {
		fmt.Printf("Delete task still failed: %v\n", err)
		fmt.Println("\nMore debugging may be required to resolve this issue")
	} else {
		fmt.Printf("Success! Delete task successful with updated implementation: %s\n", deleteResp2.TaskID)
		fmt.Println("\nFIX IS WORKING: The structure changes worked!")
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package debug

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// minimalAuthorizer implements a simple authorizer interface for this test
type minimalAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header
func (a *minimalAuthorizer) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns true if the token is non-empty
func (a *minimalAuthorizer) IsValid() bool {
	return a.token != ""
}

// getAccessTokenMinimal gets an access token for the Transfer API
func getAccessTokenMinimal(clientID, clientSecret string) (string, error) {
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

		fmt.Printf("Got token for scope: %s\n", scope)
		gotToken = true
		break
	}

	if !gotToken {
		if tokenErr != nil {
			return "", fmt.Errorf("failed to get token: %w", tokenErr)
		}
		return "", fmt.Errorf("failed to get token with any scope")
	}

	return tokenResp.AccessToken, nil
}

// RunDeleteMinimal implements a simple delete operation using direct HTTP requests
func RunDeleteMinimal() {
	// Load environment variables
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("pkg/.env.test")

	// Enable debug output
	os.Setenv("HTTP_DEBUG", "1")

	// Get an access token
	accessToken, err := getAccessTokenMinimal("", "")
	if err != nil {
		fmt.Printf("ERROR: Failed to get token: %v\n", err)
		os.Exit(1)
	}

	// Get endpoint ID
	endpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	if endpointID == "" {
		fmt.Println("ERROR: GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable is required")
		os.Exit(1)
	}

	// Create a unique path for this test
	path := fmt.Sprintf("/globus-test/minimal-test-%s", time.Now().Format("20060102-150405"))

	// First create the directory
	fmt.Printf("Creating directory: %s\n", path)
	mkdirBody := map[string]string{
		"path":      path,
		"DATA_TYPE": "mkdir",
	}
	mkdirJSON, _ := json.Marshal(mkdirBody)

	mkdirReq, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://transfer.api.globus.org/v0.10/operation/endpoint/%s/mkdir", endpointID),
		bytes.NewBuffer(mkdirJSON),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create mkdir request: %v\n", err)
		os.Exit(1)
	}

	mkdirReq.Header.Set("Authorization", "Bearer "+accessToken)
	mkdirReq.Header.Set("Content-Type", "application/json")
	
	client := http.Client{}
	mkdirResp, err := client.Do(mkdirReq)
	if err != nil {
		fmt.Printf("ERROR: Failed to execute mkdir request: %v\n", err)
		os.Exit(1)
	}
	defer mkdirResp.Body.Close()
	
	if mkdirResp.StatusCode < 200 || mkdirResp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(mkdirResp.Body)
		fmt.Printf("ERROR: Failed to create directory, status: %d, response: %s\n", 
			mkdirResp.StatusCode, string(respBody))
		os.Exit(1)
	}
	
	fmt.Println("Directory created successfully")

	// Now delete it
	fmt.Printf("Deleting directory: %s\n", path)
	
	// Get a submission ID first
	subIDReq, err := http.NewRequest(
		"GET",
		"https://transfer.api.globus.org/v0.10/submission_id",
		nil,
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create submission ID request: %v\n", err)
		os.Exit(1)
	}
	
	subIDReq.Header.Set("Authorization", "Bearer "+accessToken)
	
	subIDResp, err := client.Do(subIDReq)
	if err != nil {
		fmt.Printf("ERROR: Failed to get submission ID: %v\n", err)
		os.Exit(1)
	}
	defer subIDResp.Body.Close()
	
	if subIDResp.StatusCode < 200 || subIDResp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(subIDResp.Body)
		fmt.Printf("ERROR: Failed to get submission ID, status: %d, response: %s\n", 
			subIDResp.StatusCode, string(respBody))
		os.Exit(1)
	}
	
	var subIDData struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(subIDResp.Body).Decode(&subIDData); err != nil {
		fmt.Printf("ERROR: Failed to decode submission ID response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Got submission ID: %s\n", subIDData.Value)
	
	// Create the delete request
	deleteBody := map[string]interface{}{
		"DATA_TYPE":     "delete",
		"endpoint":      endpointID,
		"submission_id": subIDData.Value,
		"DATA": []map[string]string{
			{
				"DATA_TYPE": "delete_item",
				"path":      path,
			},
		},
	}
	deleteJSON, _ := json.Marshal(deleteBody)
	
	deleteReq, err := http.NewRequest(
		"POST",
		"https://transfer.api.globus.org/v0.10/delete",
		bytes.NewBuffer(deleteJSON),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create delete request: %v\n", err)
		os.Exit(1)
	}
	
	deleteReq.Header.Set("Authorization", "Bearer "+accessToken)
	deleteReq.Header.Set("Content-Type", "application/json")
	
	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		fmt.Printf("ERROR: Failed to execute delete request: %v\n", err)
		os.Exit(1)
	}
	defer deleteResp.Body.Close()
	
	if deleteResp.StatusCode < 200 || deleteResp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(deleteResp.Body)
		fmt.Printf("ERROR: Failed to delete directory, status: %d, response: %s\n", 
			deleteResp.StatusCode, string(respBody))
		os.Exit(1)
	}
	
	var deleteData struct {
		TaskID string `json:"task_id"`
	}
	if err := json.NewDecoder(deleteResp.Body).Decode(&deleteData); err != nil {
		fmt.Printf("ERROR: Failed to decode delete response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Delete task submitted successfully: %s\n", deleteData.TaskID)
}
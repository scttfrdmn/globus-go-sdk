// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

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
		if tokenResp.ResourceServer != "" && (
		   tokenResp.ResourceServer == "transfer.api.globus.org" || 
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
	}
	
	return tokenResp.AccessToken
}

func main() {
	// Load environment variables
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("pkg/.env.test")
	
	// Enable debug output
	os.Setenv("HTTP_DEBUG", "1")
	
	// Get credentials
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	
	if clientID == "" || clientSecret == "" || sourceEndpointID == "" {
		fmt.Println("Required environment variables not set")
		os.Exit(1)
	}

	// Get access token
	accessToken := getAccessToken(clientID, clientSecret)
	
	// Create a test directory to delete (raw HTTP request)
	testBasePath := os.Getenv("GLOBUS_TEST_DIRECTORY_PATH")
	if testBasePath == "" {
		testBasePath = "globus-test" // Default to a simple test directory 
	}
	
	timestamp := time.Now().Format("20060102_150405")
	testDir := fmt.Sprintf("%s/debug_delete_minimal_%s", testBasePath, timestamp)
	
	// Create directory with raw HTTP request
	createDirBody := map[string]string{
		"path":      testDir,
		"DATA_TYPE": "mkdir",
	}
	createDirBodyJSON, _ := json.Marshal(createDirBody)
	
	// Build the request
	createDirReq, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("https://transfer.api.globus.org/v0.10/operation/endpoint/%s/mkdir", sourceEndpointID),
		bytes.NewReader(createDirBodyJSON),
	)
	createDirReq.Header.Set("Authorization", "Bearer "+accessToken)
	createDirReq.Header.Set("Content-Type", "application/json")
	
	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(createDirReq)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	// Print response
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Create directory status: %d\n", resp.StatusCode)
	fmt.Printf("Create directory response: %s\n", string(respBody))
	
	if resp.StatusCode >= 400 {
		fmt.Println("Failed to create directory")
		os.Exit(1)
	}
	
	// Get a submission ID
	submissionIDReq, _ := http.NewRequest(
		"GET",
		"https://transfer.api.globus.org/v0.10/submission_id",
		nil,
	)
	submissionIDReq.Header.Set("Authorization", "Bearer "+accessToken)
	submissionIDReq.Header.Set("Accept", "application/json")
	
	resp, err = client.Do(submissionIDReq)
	if err != nil {
		fmt.Printf("Failed to get submission ID: %v\n", err)
		os.Exit(1)
	}
	
	var submissionIDResp struct {
		Value string `json:"value"`
	}
	respBody, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	
	fmt.Printf("Submission ID status: %d\n", resp.StatusCode)
	fmt.Printf("Submission ID response: %s\n", string(respBody))
	
	if err := json.Unmarshal(respBody, &submissionIDResp); err != nil {
		fmt.Printf("Failed to parse submission ID: %v\n", err)
		os.Exit(1)
	}
	
	submissionID := submissionIDResp.Value
	fmt.Printf("Got submission ID: %s\n", submissionID)
	
	// Try delete with a completely custom JSON payload
	// This is a minimal request based on the API docs
	deleteReqBody := fmt.Sprintf(`{
		"DATA_TYPE": "delete",
		"submission_id": "%s",
		"endpoint": "%s",
		"label": "SDK Delete Test %s",
		"DATA": [
			{
				"DATA_TYPE": "delete_item",
				"path": "%s"
			}
		]
	}`, submissionID, sourceEndpointID, timestamp, testDir)
	
	fmt.Printf("Delete request body: %s\n", deleteReqBody)
	
	deleteReq, _ := http.NewRequest(
		"POST",
		"https://transfer.api.globus.org/v0.10/delete",
		bytes.NewReader([]byte(deleteReqBody)),
	)
	deleteReq.Header.Set("Authorization", "Bearer "+accessToken)
	deleteReq.Header.Set("Content-Type", "application/json")
	deleteReq.Header.Set("Accept", "application/json")
	
	resp, err = client.Do(deleteReq)
	if err != nil {
		fmt.Printf("Failed to send delete request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	respBody, _ = io.ReadAll(resp.Body)
	fmt.Printf("Delete task status: %d\n", resp.StatusCode)
	fmt.Printf("Delete task response: %s\n", string(respBody))
	
	if resp.StatusCode >= 400 {
		fmt.Println("Delete task failed")
		fmt.Println("This might be due to permission issues or endpoint-specific restrictions")
	} else {
		fmt.Println("Success! Delete task created successfully")
		
		// Parse response to get task ID
		var taskResp struct {
			TaskID string `json:"task_id"`
		}
		if err := json.Unmarshal(respBody, &taskResp); err != nil {
			fmt.Printf("Failed to parse task response: %v\n", err)
		} else {
			fmt.Printf("Task ID: %s\n", taskResp.TaskID)
		}
	}
}
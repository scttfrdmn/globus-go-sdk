// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

// EnhancedDeleteItem adds a DATA_TYPE field to DeleteItem
type EnhancedDeleteItem struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
	DataType  string `json:"DATA_TYPE,omitempty"`
}

// EnhancedDeleteTaskRequest adds DATA_TYPE field to DeleteItem
type EnhancedDeleteTaskRequest struct {
	DataType          string               `json:"DATA_TYPE,omitempty"`
	Label             string               `json:"label,omitempty"`
	EndpointID        string               `json:"endpoint"`
	Deadline          *time.Time           `json:"deadline,omitempty"`
	NotifyOnSucceeded bool                 `json:"notify_on_succeeded,omitempty"`
	NotifyOnFailed    bool                 `json:"notify_on_failed,omitempty"`
	NotifyOnInactive  bool                 `json:"notify_on_inactive,omitempty"`
	SubmissionID      string               `json:"submission_id,omitempty"`
	Items             []EnhancedDeleteItem `json:"DATA"`
}

func main() {
	// Load environment variables
	_ = godotenv.Load(".env.test")

	// Get credentials from environment
	accessToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if accessToken == "" {
		fmt.Println("ERROR: GLOBUS_TEST_TRANSFER_TOKEN environment variable is required")
		os.Exit(1)
	}

	endpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	if endpointID == "" {
		fmt.Println("ERROR: GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable is required")
		os.Exit(1)
	}

	// Enable HTTP debugging
	os.Setenv("HTTP_DEBUG", "true")

	// Create Transfer client with debugging enabled
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&testAuthorizer{token: accessToken}),
		transfer.WithHTTPDebugging(true),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create transfer client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	timestamp := time.Now().Format("20060102_150405")

	// Create test directories for our two tests
	testPath1 := fmt.Sprintf("globus-test/debug_original_%s", timestamp)
	testPath2 := fmt.Sprintf("globus-test/debug_enhanced_%s", timestamp)

	// Create the first test directory
	fmt.Printf("Creating test directory 1: %s\n", testPath1)
	err = client.Mkdir(ctx, endpointID, testPath1)
	if err != nil {
		fmt.Printf("ERROR: Failed to create test directory 1: %v\n", err)
		os.Exit(1)
	}

	// Create the second test directory
	fmt.Printf("Creating test directory 2: %s\n", testPath2)
	err = client.Mkdir(ctx, endpointID, testPath2)
	if err != nil {
		fmt.Printf("ERROR: Failed to create test directory 2: %v\n", err)
		os.Exit(1)
	}

	// Get submission ID for our requests
	submissionID, err := client.GetSubmissionID(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to get submission ID: %v\n", err)
		os.Exit(1)
	}

	// Test 1: Original implementation
	fmt.Println("\n=== TEST 1: ORIGINAL IMPLEMENTATION ===")
	fmt.Println("Using the current SDK implementation without DATA_TYPE field in delete items")

	deleteRequest1 := &transfer.DeleteTaskRequest{
		DataType:     "delete",
		Label:        "Debug original implementation",
		EndpointID:   endpointID,
		SubmissionID: submissionID,
		Items: []transfer.DeleteItem{
			{
				Path:      testPath1,
				Recursive: true,
			},
		},
	}

	// Print the JSON that will be sent
	reqJSON1, _ := json.MarshalIndent(deleteRequest1, "", "  ")
	fmt.Printf("Request JSON:\n%s\n\n", string(reqJSON1))

	// Try to create the delete task
	fmt.Println("Sending delete request with SDK...")
	resp1, err := client.CreateDeleteTask(ctx, deleteRequest1)
	
	if err != nil {
		fmt.Printf("ERROR: Original delete task failed: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Original delete task created with task ID: %s\n", resp1.TaskID)
	}

	// Test 2: Enhanced implementation with DATA_TYPE field
	fmt.Println("\n=== TEST 2: ENHANCED IMPLEMENTATION ===")
	fmt.Println("Adding DATA_TYPE field to each delete item")

	// Create an enhanced request with DATA_TYPE field for each item
	enhancedRequest := EnhancedDeleteTaskRequest{
		DataType:     "delete",
		Label:        "Debug with DATA_TYPE for items",
		EndpointID:   endpointID,
		SubmissionID: submissionID,
		Items: []EnhancedDeleteItem{
			{
				Path:      testPath2,
				Recursive: true,
				DataType:  "delete_item", // This is the key difference
			},
		},
	}

	// Print the enhanced JSON
	reqJSON2, _ := json.MarshalIndent(enhancedRequest, "", "  ")
	fmt.Printf("Enhanced Request JSON:\n%s\n\n", string(reqJSON2))

	// We can't use the SDK directly with our enhanced struct, so we'll send the request manually
	reqBody, err := json.Marshal(enhancedRequest)
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal enhanced request: %v\n", err)
		os.Exit(1)
	}

	// Create an HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, 
		"https://transfer.api.globus.org/v0.10/delete", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Printf("ERROR: Failed to create HTTP request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request
	fmt.Println("Sending enhanced delete request directly...")
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: HTTP request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read and parse the response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ERROR: Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Print response status and body
	fmt.Printf("Response status: %s\n", resp.Status)
	fmt.Printf("Response body: %s\n", string(respBody))

	// Parse the response
	var taskResp transfer.TaskResponse
	if err := json.Unmarshal(respBody, &taskResp); err != nil {
		fmt.Printf("ERROR: Failed to parse response: %v\n", err)
	} else if taskResp.TaskID != "" {
		fmt.Printf("SUCCESS: Enhanced delete task created with task ID: %s\n", taskResp.TaskID)
	}

	fmt.Println("\n=== DEBUGGING COMPLETE ===")
	fmt.Println("If the enhanced implementation succeeded but the original failed,")
	fmt.Println("then adding the DATA_TYPE field to delete items is likely the solution.")
	fmt.Println("This would require updating the DeleteItem struct in models.go to include DataType")
	fmt.Println("and modifying CreateDeleteTask to set it to \"delete_item\" for each item.")
}
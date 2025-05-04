// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
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

func main() {
	// Enable HTTP debugging
	os.Setenv("HTTP_DEBUG", "true")

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

	// Create a test directory to delete
	timestamp := time.Now().Format("20060102_150405")
	testPath := fmt.Sprintf("globus-test/debug_delete_%s", timestamp)
	
	fmt.Printf("Creating test directory: %s\n", testPath)
	err = client.Mkdir(ctx, endpointID, testPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to create test directory: %v\n", err)
		os.Exit(1)
	}

	// Now attempt to delete the directory using DeleteTaskRequest
	fmt.Println("\nAttempting to delete directory using DeleteTaskRequest...")
	
	deleteRequest := &transfer.DeleteTaskRequest{
		DataType:   "delete",
		Label:      fmt.Sprintf("Debug Delete Task %s", timestamp),
		EndpointID: endpointID,
		Items: []transfer.DeleteItem{
			{
				Path:      testPath,
				Recursive: true,
			},
		},
	}

	// Execute the delete task
	resp, err := client.CreateDeleteTask(ctx, deleteRequest)
	if err != nil {
		fmt.Printf("ERROR: Delete task failed: %v\n", err)
		fmt.Println("\nPossible cause: DeleteItem might need a DATA_TYPE field similar to TransferItem")
		fmt.Println("In the TransferTaskRequest implementation, each TransferItem gets a DataType if not set:")
		fmt.Println("    if request.Items[i].DataType == \"\" {\n        request.Items[i].DataType = \"transfer_item\"\n    }")
		fmt.Println("A similar fix might be needed for DeleteItem in CreateDeleteTask")
	} else {
		fmt.Printf("SUCCESS: Delete task created with task ID: %s\n", resp.TaskID)
		
		// Wait a moment and check task status
		time.Sleep(2 * time.Second)
		task, err := client.GetTask(ctx, resp.TaskID)
		if err != nil {
			fmt.Printf("ERROR: Failed to get task status: %v\n", err)
		} else {
			fmt.Printf("Task status: %s\n", task.Status)
		}
	}
}
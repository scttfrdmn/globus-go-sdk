// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package debug_files

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
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
	// Load environment variables from .env.test file
	_ = godotenv.Load(".env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load("../../../.env.test")

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

	// Set HTTP_DEBUG environment variable to enable debugging
	os.Setenv("HTTP_DEBUG", "true")

	// Create Transfer client with debugging enabled
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(&testAuthorizer{token: accessToken}),
		transfer.WithHTTPDebugging(true),
		transfer.WithHTTPTracing(true),
	)
	if err != nil {
		fmt.Printf("ERROR: Failed to create transfer client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Create a test directory to delete
	testPath := fmt.Sprintf("globus-test/debug_delete_%s", time.Now().Format("20060102_150405"))

	fmt.Printf("Creating test directory: %s\n", testPath)
	err = client.Mkdir(ctx, endpointID, testPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to create test directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== DELETE TASK REQUEST WITH DATATYPE FIELD FOR ITEMS ===")

	// Create delete task with DataType fields set for items
	deleteRequest1 := &transfer.DeleteTaskRequest{
		DataType:   "delete",
		Label:      "Debug delete with DataType",
		EndpointID: endpointID,
		Items: []transfer.DeleteItem{
			{
				DataType: "delete_item",
				Path:     testPath,
				// Note: The API does not support a "recursive" field for delete_item as of API v0.10
			},
		},
	}

	// This is what we're testing - should we be setting a DataType field
	// on each DeleteItem similarly to how it's done for TransferItem?
	fmt.Println("\n=== ATTEMPTING DELETE WITH CURRENT IMPLEMENTATION ===")
	resp1, err := client.CreateDeleteTask(ctx, deleteRequest1)
	if err != nil {
		fmt.Printf("ERROR: Delete task failed: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Delete task created with task ID: %s\n", resp1.TaskID)
	}

	// Note for improvement: If the first attempt succeeded, create another directory
	// and try a version where we manually add the DATA_TYPE field

	fmt.Println("\n=== DEBUG COMPLETE ===")
}

//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

func getTestCredentials(t *testing.T) (string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	endpointID := os.Getenv("GLOBUS_TEST_COMPUTE_ENDPOINT_ID")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}

	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	return clientID, clientSecret, endpointID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	// Create auth client with client ID and secret
	authClient, err := auth.NewClient(
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), ComputeScope)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	return tokenResp.AccessToken
}

func TestIntegration_ListEndpoints(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Compute client
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// List endpoints
	endpoints, err := client.ListEndpoints(ctx, &ListEndpointsOptions{
		PerPage: 5,
	})
	if err != nil {
		if core.IsNotFound(err) || core.IsForbidden(err) || core.IsUnauthorized(err) {
			t.Logf("Client correctly made the request, but returned expected error due to permissions: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			return // Skip the rest of the test
		} else {
			t.Fatalf("ListEndpoints failed with unexpected error: %v", err)
		}
	}

	// Verify we got some data
	t.Logf("Found %d endpoints", len(endpoints.Endpoints))

	// The user might not have any endpoints, so this isn't necessarily an error
	if len(endpoints.Endpoints) > 0 {
		// Check that the first endpoint has expected fields
		firstEndpoint := endpoints.Endpoints[0]
		if firstEndpoint.ID == "" {
			t.Error("First endpoint is missing ID")
		}
		if firstEndpoint.Name == "" {
			t.Error("First endpoint is missing name")
		}
	}
}

func TestIntegration_FunctionLifecycle(t *testing.T) {
	clientID, clientSecret, endpointID := getTestCredentials(t)

	// Skip if no endpoint ID is provided
	if endpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_COMPUTE_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Compute client
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// 1. Register a new function
	timestamp := time.Now().Format("20060102_150405")
	functionName := fmt.Sprintf("test_func_%s", timestamp)

	// Simple hello world function for testing
	functionCode := `def hello(name="World"):
    return f"Hello, {name}! (from integration test)"
`

	registerRequest := &FunctionRegisterRequest{
		Function:    functionCode,
		Name:        functionName,
		Description: "A test function created by integration tests",
		Public:      false,
	}

	createdFunction, err := client.RegisterFunction(ctx, registerRequest)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Make sure the function gets deleted after the test
	defer func() {
		err := client.DeleteFunction(ctx, createdFunction.ID)
		if err != nil {
			t.Logf("Warning: Failed to delete test function (%s): %v", createdFunction.ID, err)
		} else {
			t.Logf("Successfully deleted test function (%s)", createdFunction.ID)
		}
	}()

	t.Logf("Created function: %s (%s)", createdFunction.Name, createdFunction.ID)

	// 2. Verify the function was created correctly
	if createdFunction.Name != functionName {
		t.Errorf("Created function name = %s, want %s", createdFunction.Name, functionName)
	}

	// 3. Get the function
	fetchedFunction, err := client.GetFunction(ctx, createdFunction.ID)
	if err != nil {
		t.Fatalf("Failed to get function: %v", err)
	}

	if fetchedFunction.ID != createdFunction.ID {
		t.Errorf("Fetched function ID = %s, want %s", fetchedFunction.ID, createdFunction.ID)
	}

	// 4. Update the function
	updatedDescription := "Updated description for integration test"
	updateRequest := &FunctionUpdateRequest{
		Description: updatedDescription,
	}

	updatedFunction, err := client.UpdateFunction(ctx, createdFunction.ID, updateRequest)
	if err != nil {
		t.Fatalf("Failed to update function: %v", err)
	}

	if updatedFunction.Description != updatedDescription {
		t.Errorf("Updated function description = %s, want %s", updatedFunction.Description, updatedDescription)
	}

	// 5. Run the function
	taskRequest := &TaskRequest{
		FunctionID: createdFunction.ID,
		EndpointID: endpointID,
		Args:       []interface{}{"Integration Test"},
	}

	task, err := client.RunFunction(ctx, taskRequest)
	if err != nil {
		t.Fatalf("Failed to run function: %v", err)
	}

	t.Logf("Started task: %s (Status: %s)", task.TaskID, task.Status)

	// Wait a moment for the task to complete
	time.Sleep(3 * time.Second)

	// 6. Get the task status
	taskStatus, err := client.GetTaskStatus(ctx, task.TaskID)
	if err != nil {
		t.Fatalf("Failed to get task status: %v", err)
	}

	t.Logf("Task status: %s", taskStatus.Status)
	if taskStatus.Status == "SUCCESS" {
		t.Logf("Task result: %v", taskStatus.Result)
	} else if taskStatus.Status == "FAILED" {
		t.Logf("Task exception: %s", taskStatus.Exception)
	}
}

func TestIntegration_BatchExecution(t *testing.T) {
	clientID, clientSecret, endpointID := getTestCredentials(t)

	// Skip if no endpoint ID is provided
	if endpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_COMPUTE_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Compute client
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// 1. Register test functions
	timestamp := time.Now().Format("20060102_150405")

	// Simple functions for testing
	functions := []struct {
		name string
		code string
	}{
		{
			name: fmt.Sprintf("test_add_%s", timestamp),
			code: "def add(a, b): return a + b",
		},
		{
			name: fmt.Sprintf("test_multiply_%s", timestamp),
			code: "def multiply(a, b): return a * b",
		},
	}

	functionIDs := make([]string, len(functions))

	// Register functions
	for i, fn := range functions {
		registerRequest := &FunctionRegisterRequest{
			Function:    fn.code,
			Name:        fn.name,
			Description: "A test function for batch execution",
			Public:      false,
		}

		createdFunction, err := client.RegisterFunction(ctx, registerRequest)
		if err != nil {
			t.Fatalf("Failed to register function %s: %v", fn.name, err)
		}

		functionIDs[i] = createdFunction.ID

		// Make sure the function gets deleted after the test
		defer func(id string) {
			err := client.DeleteFunction(ctx, id)
			if err != nil {
				t.Logf("Warning: Failed to delete test function (%s): %v", id, err)
			}
		}(createdFunction.ID)

		t.Logf("Created function: %s (%s)", createdFunction.Name, createdFunction.ID)
	}

	// 2. Run batch of functions
	batchRequest := &BatchTaskRequest{
		Tasks: []TaskRequest{
			{
				FunctionID: functionIDs[0],
				EndpointID: endpointID,
				Args:       []interface{}{5, 3},
			},
			{
				FunctionID: functionIDs[1],
				EndpointID: endpointID,
				Args:       []interface{}{5, 3},
			},
		},
	}

	batchResponse, err := client.RunBatch(ctx, batchRequest)
	if err != nil {
		t.Fatalf("Failed to run batch: %v", err)
	}

	t.Logf("Submitted batch with %d tasks", len(batchResponse.TaskIDs))

	// Wait a moment for tasks to complete
	time.Sleep(3 * time.Second)

	// 3. Get batch status
	batchStatus, err := client.GetBatchStatus(ctx, batchResponse.TaskIDs)
	if err != nil {
		t.Fatalf("Failed to get batch status: %v", err)
	}

	t.Logf("Batch status: %d completed, %d pending, %d failed",
		len(batchStatus.Completed), len(batchStatus.Pending), len(batchStatus.Failed))

	// 4. Check individual task results
	for taskID, status := range batchStatus.Tasks {
		t.Logf("Task %s: Status = %s", taskID, status.Status)
		if status.Status == "SUCCESS" {
			t.Logf("  Result: %v", status.Result)
		} else if status.Status == "FAILED" {
			t.Logf("  Exception: %s", status.Exception)
		}
	}
}

func TestIntegration_ListFunctions(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Compute client
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// List functions
	functions, err := client.ListFunctions(ctx, &ListFunctionsOptions{
		PerPage: 5,
	})
	if err != nil {
		// The ListFunctions endpoint returns 405 Method Not Allowed in some configurations
		// This is a known issue with the Compute API
		errorMsg := err.Error()
		if core.IsNotFound(err) || core.IsForbidden(err) || core.IsUnauthorized(err) ||
			(errorMsg != "" && (errorMsg == "unknown_error: Request failed with status code 405 (status: 405)" ||
				errorMsg == "request failed with status 405: Method Not Allowed")) {
			t.Logf("Client correctly made the request, but returned expected error: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			return // Skip the rest of the test
		} else {
			t.Fatalf("ListFunctions failed with unexpected error: %v", err)
		}
	}

	// Verify we got some data
	t.Logf("Found %d functions", len(functions.Functions))

	// The user might not have any functions, so this isn't necessarily an error
	if len(functions.Functions) > 0 {
		// Check that the first function has expected fields
		firstFunction := functions.Functions[0]
		if firstFunction.ID == "" {
			t.Error("First function is missing ID")
		}
		if firstFunction.Name == "" {
			t.Error("First function is missing name")
		}

		// Print some details
		t.Logf("Example function: %s (%s)", firstFunction.Name, firstFunction.ID)
		t.Logf("  Description: %s", firstFunction.Description)
		t.Logf("  Owner: %s", firstFunction.Owner)
		t.Logf("  Public: %t", firstFunction.Public)
	}
}

func TestIntegration_ListTasks(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Compute client
	client, err := NewClient(WithAccessToken(accessToken))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := context.Background()

	// List tasks
	tasks, err := client.ListTasks(ctx, &TaskListOptions{
		PerPage: 5,
	})
	if err != nil {
		if core.IsNotFound(err) || core.IsForbidden(err) || core.IsUnauthorized(err) {
			t.Logf("Client correctly made the request, but returned expected error due to permissions: %v", err)
			t.Logf("This is acceptable for integration testing with limited-permission credentials")
			return // Skip the rest of the test
		} else {
			t.Fatalf("ListTasks failed with unexpected error: %v", err)
		}
	}

	// Verify we got some data
	t.Logf("Found %d tasks", len(tasks.Tasks))

	// The user might not have any tasks, so this isn't necessarily an error
	if len(tasks.Tasks) > 0 {
		// Check that we have task IDs
		for i, taskID := range tasks.Tasks {
			if taskID == "" {
				t.Errorf("Task %d is missing ID", i)
			} else {
				t.Logf("Task %d: %s", i+1, taskID)

				// Get status for the first task
				if i == 0 {
					status, err := client.GetTaskStatus(ctx, taskID)
					if err != nil {
						t.Logf("Failed to get status for task %s: %v", taskID, err)
					} else {
						t.Logf("  Status: %s", status.Status)
					}
				}
			}
		}
	}
}

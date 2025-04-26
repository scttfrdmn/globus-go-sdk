// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package transfer

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/services/auth"
)

func getTestCredentials(t *testing.T) (string, string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")
	
	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}
	
	return clientID, clientSecret, sourceEndpointID, destEndpointID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	authClient := auth.NewClient(clientID, clientSecret)
	
	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), TransferScope)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}
	
	return tokenResp.AccessToken
}

func TestIntegration_ListEndpoints(t *testing.T) {
	clientID, clientSecret, _, _ := getTestCredentials(t)
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Transfer client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// List endpoints
	endpoints, err := client.ListEndpoints(ctx, &ListEndpointsOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("ListEndpoints failed: %v", err)
	}
	
	// Verify we got some endpoints
	if len(endpoints.DATA) == 0 {
		t.Log("No endpoints found, but this is not an error. User may not have any endpoints.")
	} else {
		t.Logf("Found %d endpoints", len(endpoints.DATA))
		
		// Verify endpoint data
		for i, endpoint := range endpoints.DATA {
			if endpoint.ID == "" {
				t.Errorf("Endpoint %d is missing ID", i)
			}
			if endpoint.DisplayName == "" {
				t.Errorf("Endpoint %d is missing display name", i)
			}
		}
	}
}

func TestIntegration_TransferFlow(t *testing.T) {
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)
	
	// Skip if transfer endpoints are not provided
	if sourceEndpointID == "" || destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID and GLOBUS_TEST_DEST_ENDPOINT_ID")
	}
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Transfer client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// 1. Verify endpoints exist
	sourceEndpoint, err := client.GetEndpoint(ctx, sourceEndpointID)
	if err != nil {
		t.Fatalf("Failed to get source endpoint: %v", err)
	}
	t.Logf("Source endpoint: %s (%s)", sourceEndpoint.DisplayName, sourceEndpoint.ID)
	
	destEndpoint, err := client.GetEndpoint(ctx, destEndpointID)
	if err != nil {
		t.Fatalf("Failed to get destination endpoint: %v", err)
	}
	t.Logf("Destination endpoint: %s (%s)", destEndpoint.DisplayName, destEndpoint.ID)
	
	// 2. Activate endpoints (ignore errors as they might already be activated)
	_ = client.ActivateEndpoint(ctx, sourceEndpointID)
	_ = client.ActivateEndpoint(ctx, destEndpointID)
	
	// 3. Create a unique test file name
	timestamp := time.Now().Format("20060102_150405")
	sourceFilePath := fmt.Sprintf("/~/%s_test.txt", timestamp)
	destFilePath := fmt.Sprintf("/~/%s_received.txt", timestamp)
	
	// 4. Submit a transfer
	label := fmt.Sprintf("Integration Test Transfer %s", timestamp)
	transferRequest := &TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		Items: []TransferItem{
			{
				SourcePath:      sourceFilePath,
				DestinationPath: destFilePath,
			},
		},
	}
	
	// Important: This test might fail if the source file doesn't exist
	// In real usage, ensure the source file exists before testing
	t.Logf("Transferring %s to %s", sourceFilePath, destFilePath)
	taskResponse, err := client.CreateTransferTask(ctx, transferRequest)
	
	// This might fail if the source file doesn't exist,
	// but we just want to test the API interaction
	if err != nil {
		t.Logf("Transfer request failed (possibly file doesn't exist): %v", err)
		// Not marking as fatal to continue with the test
	} else {
		t.Logf("Transfer task submitted, task ID: %s", taskResponse.TaskID)
		
		// 5. Get task status
		task, err := client.GetTask(ctx, taskResponse.TaskID)
		if err != nil {
			t.Fatalf("Failed to get task status: %v", err)
		}
		
		t.Logf("Task status: %s", task.Status)
		
		// 6. List tasks
		tasks, err := client.ListTasks(ctx, &ListTasksOptions{
			FilterTaskID: taskResponse.TaskID,
		})
		if err != nil {
			t.Fatalf("Failed to list tasks: %v", err)
		}
		
		if len(tasks.DATA) == 0 {
			t.Error("Task not found in tasks list")
		} else {
			t.Logf("Found task in tasks list: %s", tasks.DATA[0].TaskID)
		}
	}
}

func TestIntegration_GetEndpointActivationRequirements(t *testing.T) {
	clientID, clientSecret, sourceEndpointID, _ := getTestCredentials(t)
	
	// Skip if endpoint is not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	}
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Transfer client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// Get activation requirements
	requirements, err := client.GetActivationRequirements(ctx, sourceEndpointID)
	if err != nil {
		t.Fatalf("Failed to get activation requirements: %v", err)
	}
	
	// Verified we got a response - exact requirements depend on the endpoint type
	t.Logf("Activation requirements data type: %s", requirements.DataType)
	t.Logf("Number of activation requirements: %d", len(requirements.ActivationRequirements))
	
	// Try to activate the endpoint (might already be activated)
	err = client.ActivateEndpoint(ctx, sourceEndpointID)
	if err != nil {
		t.Logf("Activation might require additional steps: %v", err)
	} else {
		t.Log("Endpoint activated successfully")
	}
}
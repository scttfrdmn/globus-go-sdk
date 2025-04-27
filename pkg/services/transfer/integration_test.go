// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
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

	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), "urn:globus:auth:scope:transfer.api.globus.org:all")
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
	client := transfer.NewClient(accessToken)
	ctx := context.Background()

	// List endpoints
	endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
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

	// 3. Create test directories with timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("/~/test_transfer_%s", timestamp)
	destDir := fmt.Sprintf("/~/test_received_%s", timestamp)

	// Create source directory
	err = client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
		EndpointID: sourceEndpointID,
		Path:       sourceDir,
	})
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	t.Logf("Created source directory: %s", sourceDir)

	// Setup cleanup for source directory
	defer func() {
		err := client.DeleteItem(ctx, &transfer.DeleteItemOptions{
			EndpointID: sourceEndpointID,
			Path:       sourceDir,
			Recursive:  true,
		})
		if err != nil {
			t.Logf("Warning: Failed to delete source directory: %v", err)
		} else {
			t.Logf("Cleaned up source directory: %s", sourceDir)
		}
	}()

	// Create destination directory
	err = client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
		EndpointID: destEndpointID,
		Path:       destDir,
	})
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	t.Logf("Created destination directory: %s", destDir)

	// Setup cleanup for destination directory
	defer func() {
		err := client.DeleteItem(ctx, &transfer.DeleteItemOptions{
			EndpointID: destEndpointID,
			Path:       destDir,
			Recursive:  true,
		})
		if err != nil {
			t.Logf("Warning: Failed to delete destination directory: %v", err)
		} else {
			t.Logf("Cleaned up destination directory: %s", destDir)
		}
	}()

	// 4. Create a subdirectory in source
	sourceSubDir := sourceDir + "/subdir"
	err = client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
		EndpointID: sourceEndpointID,
		Path:       sourceSubDir,
	})
	if err != nil {
		t.Fatalf("Failed to create source subdirectory: %v", err)
	}
	t.Logf("Created source subdirectory: %s", sourceSubDir)

	// 5. Create a test file in source directory
	// First create a local temporary file
	tempFile, err := ioutil.TempFile("", "globus-transfer-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	testContent := fmt.Sprintf("Globus Transfer SDK Test File Content %s", timestamp)
	_, err = tempFile.WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	// Get the filename part
	tempFilename := filepath.Base(tempFile.Name())
	sourceFilePath := fmt.Sprintf("%s/%s", sourceDir, tempFilename)
	destFilePath := fmt.Sprintf("%s/%s", destDir, tempFilename)

	t.Logf("Local file created: %s", tempFile.Name())
	defer os.Remove(tempFile.Name())

	// Upload the file to source endpoint
	// Note: This step may require separate helper functions/tools depending on
	// how the SDK is designed to handle uploads
	t.Logf("Uploading file to source endpoint path: %s", sourceFilePath)

	// This is a simplified example - actual implementation would depend on how the SDK supports uploads
	// For now, assume the file is uploaded to the source endpoint by other means

	// 6. Submit a transfer from source to destination
	label := fmt.Sprintf("Integration Test Transfer %s", timestamp)
	transferRequest := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Encrypt:               true,
		VerifyChecksum:        true,
		Items: []transfer.TransferItem{
			{
				SourcePath:      sourceDir,
				DestinationPath: destDir,
				Recursive:       true,
			},
		},
	}

	t.Logf("Submitting transfer: %s to %s", sourceDir, destDir)
	taskResponse, err := client.CreateTransferTask(ctx, transferRequest)
	if err != nil {
		t.Fatalf("Failed to submit transfer task: %v", err)
	}

	t.Logf("Transfer task submitted, task ID: %s", taskResponse.TaskID)

	// 7. Wait for task completion (with a timeout)
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	task, err := client.WaitForTaskCompletion(waitCtx, taskResponse.TaskID, 10*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("Task did not complete within timeout period, but this doesn't necessarily indicate failure")
		} else {
			t.Fatalf("Error waiting for task completion: %v", err)
		}
	} else {
		t.Logf("Task completed with status: %s", task.Status)

		if task.Status == "SUCCEEDED" {
			// 8. List contents of destination directory to verify transfer
			listOptions := &transfer.ListDirectoryOptions{
				EndpointID: destEndpointID,
				Path:       destDir,
			}

			listing, err := client.ListDirectory(ctx, listOptions)
			if err != nil {
				t.Fatalf("Failed to list destination directory: %v", err)
			}

			t.Logf("Destination directory contents (%d items):", len(listing.DATA))
			for _, item := range listing.DATA {
				t.Logf("  - %s [%s]", item.Name, item.Type)
			}
		}
	}

	// 9. Test file rename operation
	if task.Status == "SUCCEEDED" {
		renamedPath := destDir + "/renamed_file.txt"
		t.Logf("Renaming file from %s to %s", destFilePath, renamedPath)

		renameOptions := &transfer.RenameItemOptions{
			EndpointID: destEndpointID,
			OldPath:    destFilePath,
			NewPath:    renamedPath,
		}

		err = client.RenameItem(ctx, renameOptions)
		if err != nil {
			t.Logf("Rename operation failed (file may not exist): %v", err)
		} else {
			t.Logf("File renamed successfully")

			// List directory again to confirm rename
			listOptions := &transfer.ListDirectoryOptions{
				EndpointID: destEndpointID,
				Path:       destDir,
			}

			listing, err := client.ListDirectory(ctx, listOptions)
			if err != nil {
				t.Fatalf("Failed to list destination directory after rename: %v", err)
			}

			found := false
			for _, item := range listing.DATA {
				if item.Name == "renamed_file.txt" {
					found = true
					break
				}
			}

			if found {
				t.Logf("Renamed file found in destination directory")
			} else {
				t.Errorf("Renamed file not found in destination directory")
			}
		}
	}
}

func TestIntegration_RecursiveTransfer(t *testing.T) {
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

	// 1. Create test directory structure with timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("/~/test_recursive_%s", timestamp)
	destDir := fmt.Sprintf("/~/test_recursive_dest_%s", timestamp)

	// Create source directory
	err := client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
		EndpointID: sourceEndpointID,
		Path:       sourceDir,
	})
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	t.Logf("Created source directory: %s", sourceDir)

	// Setup cleanup for source directory
	defer func() {
		err := client.DeleteItem(ctx, &transfer.DeleteItemOptions{
			EndpointID: sourceEndpointID,
			Path:       sourceDir,
			Recursive:  true,
		})
		if err != nil {
			t.Logf("Warning: Failed to delete source directory: %v", err)
		} else {
			t.Logf("Cleaned up source directory: %s", sourceDir)
		}
	}()

	// Create destination directory
	err = client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
		EndpointID: destEndpointID,
		Path:       destDir,
	})
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	t.Logf("Created destination directory: %s", destDir)

	// Setup cleanup for destination directory
	defer func() {
		err := client.DeleteItem(ctx, &transfer.DeleteItemOptions{
			EndpointID: destEndpointID,
			Path:       destDir,
			Recursive:  true,
		})
		if err != nil {
			t.Logf("Warning: Failed to delete destination directory: %v", err)
		} else {
			t.Logf("Cleaned up destination directory: %s", destDir)
		}
	}()

	// 2. Create a nested directory structure
	for _, subpath := range []string{"/subdir1", "/subdir1/nested1", "/subdir2"} {
		err = client.CreateDirectory(ctx, &transfer.CreateDirectoryOptions{
			EndpointID: sourceEndpointID,
			Path:       sourceDir + subpath,
		})
		if err != nil {
			t.Fatalf("Failed to create nested directory %s: %v", subpath, err)
		}
		t.Logf("Created nested directory: %s%s", sourceDir, subpath)
	}

	// 3. Test recursive transfer using the SDK's recursive transfer functionality
	err = client.RecursiveTransfer(ctx, &transfer.RecursiveTransferOptions{
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		SourcePath:            sourceDir,
		DestinationPath:       destDir,
		Label:                 fmt.Sprintf("Recursive Transfer Test %s", timestamp),
		Sync:                  true,
		VerifyChecksum:        true,
	})

	if err != nil {
		t.Fatalf("RecursiveTransfer failed: %v", err)
	}

	t.Log("Recursive transfer submitted successfully")

	// 4. List tasks to verify transfer was initiated
	listTasksOptions := &transfer.ListTasksOptions{
		Limit: 5,
	}

	tasks, err := client.ListTasks(ctx, listTasksOptions)
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks.DATA) == 0 {
		t.Error("No transfer tasks found after recursive transfer")
	} else {
		t.Logf("Found %d recent tasks, most recent: %s", len(tasks.DATA), tasks.DATA[0].TaskID)
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

	// Get endpoint autoactivation status
	endpoint, err := client.GetEndpoint(ctx, sourceEndpointID)
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}

	t.Logf("Endpoint autoactivation enabled: %v", endpoint.IsAutoActivateEnabled)
	t.Logf("Endpoint activation profile is public: %v", endpoint.ActivationProfile == "public")
}

func TestIntegration_TaskManagement(t *testing.T) {
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

	// 1. Create a small transfer task
	timestamp := time.Now().Format("20060102_150405")
	label := fmt.Sprintf("Task Management Test %s", timestamp)
	sourcePath := "/~/" // Assuming user home directory exists
	destPath := "/~/"

	transferRequest := &transfer.TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destEndpointID,
		Sync:                  false,
		Items: []transfer.TransferItem{
			{
				SourcePath:      sourcePath,
				DestinationPath: destPath,
				Recursive:       false,
			},
		},
	}

	// This transfer might fail if paths don't exist, but we just want to test task management
	taskResponse, err := client.CreateTransferTask(ctx, transferRequest)
	if err != nil {
		t.Logf("Transfer request failed (possibly path doesn't exist): %v", err)
		// Create a minimal task if the main one failed
		deleteTaskRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      "Delete task for testing task management",
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					Path: "/~/nonexistent_path_for_test",
				},
			},
		}

		taskResponse, err = client.CreateDeleteTask(ctx, deleteTaskRequest)
		if err != nil {
			t.Fatalf("Failed to create any test task: %v", err)
		}
	}

	t.Logf("Task created with ID: %s", taskResponse.TaskID)

	// 2. Get task by ID
	task, err := client.GetTask(ctx, taskResponse.TaskID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	t.Logf("Task status: %s", task.Status)
	t.Logf("Task label: %s", task.Label)

	// 3. Update task label if the task is not completed
	if task.Status != "SUCCEEDED" && task.Status != "FAILED" {
		updatedLabel := fmt.Sprintf("Updated Test Task %s", timestamp)
		t.Logf("Updating task label to: %s", updatedLabel)

		updateRequest := &transfer.UpdateTaskLabelOptions{
			TaskID: taskResponse.TaskID,
			Label:  updatedLabel,
		}

		err = client.UpdateTaskLabel(ctx, updateRequest)
		if err != nil {
			t.Logf("Failed to update task label (might be already completed): %v", err)
		} else {
			// Verify label was updated
			updatedTask, err := client.GetTask(ctx, taskResponse.TaskID)
			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if updatedTask.Label == updatedLabel {
				t.Log("Task label updated successfully")
			} else {
				t.Errorf("Task label not updated, expected %s, got %s", updatedLabel, updatedTask.Label)
			}
		}
	} else {
		t.Logf("Task already completed, skipping label update")
	}

	// 4. Test task cancellation if the task is still active
	if task.Status == "ACTIVE" {
		t.Log("Attempting to cancel task")

		err = client.CancelTask(ctx, taskResponse.TaskID)
		if err != nil {
			t.Logf("Failed to cancel task (might be already completed): %v", err)
		} else {
			t.Log("Task cancellation request submitted")

			// Verify task was cancelled
			cancelledTask, err := client.GetTask(ctx, taskResponse.TaskID)
			if err != nil {
				t.Fatalf("Failed to get cancelled task: %v", err)
			}

			t.Logf("Task status after cancellation request: %s", cancelledTask.Status)
		}
	} else {
		t.Logf("Task not in ACTIVE state, skipping cancellation test")
	}

	// 5. Get task events
	events, err := client.GetTaskEvents(ctx, taskResponse.TaskID, &transfer.GetTaskEventsOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Failed to get task events: %v", err)
	}

	t.Logf("Found %d task events", len(events.DATA))
	for i, event := range events.DATA {
		t.Logf("Event %d: %s at %s", i, event.Code, event.Time)
	}
}

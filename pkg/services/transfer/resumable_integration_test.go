// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package transfer_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

// getTestCredentials retrieves test credentials from environment variables
func getTestCredentials(t *testing.T) (string, string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	sourceEndpointID := os.Getenv("GLOBUS_TEST_SOURCE_ENDPOINT_ID")
	destEndpointID := os.Getenv("GLOBUS_TEST_DEST_ENDPOINT_ID")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}

	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	return clientID, clientSecret, sourceEndpointID, destEndpointID
}

// getAccessToken gets an access token for testing
func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	// First, check if there's a transfer token provided directly
	staticToken := os.Getenv("GLOBUS_TEST_TRANSFER_TOKEN")
	if staticToken != "" {
		t.Log("Using static transfer token from environment")
		return staticToken
	}

	// If no static token, try to get one via client credentials
	t.Log("No static token found, trying to get token via client credentials")
	authClient, err := auth.NewClient(
		auth.WithClientCredentials(clientID, clientSecret),
	)
	if err != nil {
		t.Skipf("Failed to create auth client: %v", err)
		return ""
	}

	// Get token via client credentials
	resp, err := authClient.GetClientCredentialsToken(context.Background(), []string{"urn:globus:auth:scope:transfer.api.globus.org:all"})
	if err != nil {
		t.Skipf("Failed to get token via client credentials: %v", err)
		return ""
	}

	return resp.AccessToken
}

// TestIntegration_ResumableTransfer tests the resumable transfer functionality
func TestIntegration_ResumableTransfer(t *testing.T) {
	// Skip this test during automated CI runs if credentials aren't provided
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if endpoints are not provided
	if sourceEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID environment variable")
	}

	if destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_DEST_ENDPOINT_ID environment variable")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client with new pattern
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Set up test directories with timestamp to avoid conflicts
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("/~/resumable-test-source-%s", timestamp)
	destDir := fmt.Sprintf("/~/resumable-test-dest-%s", timestamp)

	// Create source directory with retry for rate limiting
	t.Logf("Creating source directory: %s", sourceDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Set up cleanup for source directory
	defer func() {
		t.Logf("Cleaning up source test directory: %s", sourceDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup source test dir %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     sourceDir,
				},
			},
		}

		// Use retry for cleanup to handle rate limiting
		_ = ratelimit.RetryWithBackoff(
			context.Background(), // Use a new context for cleanup
			func(ctx context.Context) error {
				_, err := client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
	}()

	// Create destination directory with retry
	t.Logf("Creating destination directory: %s", destDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, destEndpointID, destDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Set up cleanup for destination directory
	defer func() {
		t.Logf("Cleaning up destination test directory: %s", destDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup dest test dir %s", timestamp),
			EndpointID: destEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     destDir,
				},
			},
		}

		// Use retry for cleanup to handle rate limiting
		_ = ratelimit.RetryWithBackoff(
			context.Background(), // Use a new context for cleanup
			func(ctx context.Context) error {
				_, err := client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
	}()

	// Create test files in the source directory
	fileCount := 5
	t.Logf("Creating %d test files in source directory", fileCount)

	for i := 1; i <= fileCount; i++ {
		fileName := fmt.Sprintf("%s/file%d.txt", sourceDir, i)
		fileContent := fmt.Sprintf("Test file %d for resumable transfer test", i)

		// Create file with retry for rate limiting
		var taskID string
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				taskID, err = client.SubmitTextFileCreation(
					ctx,
					sourceEndpointID,
					fileName,
					fileContent,
					fmt.Sprintf("Test file %d creation", i),
				)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		// Wait for task completion with retry
		t.Logf("Waiting for file %d creation to complete (task: %s)", i, taskID)
		var task *transfer.Task
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				task, err = client.WaitForTaskCompletion(ctx, taskID, 1*time.Second)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
		if err != nil {
			t.Fatalf("Failed to wait for file creation task: %v", err)
		}

		if task.Status != "SUCCEEDED" {
			t.Fatalf("File creation task failed with status: %s", task.Status)
		}
	}

	// Set up progress tracking
	var progressCalled bool
	var lastProgress float64

	progressCallback := func(state *transfer.CheckpointState) {
		progressCalled = true
		if state.Stats.TotalItems > 0 {
			lastProgress = float64(state.Stats.CompletedItems) / float64(state.Stats.TotalItems) * 100
			t.Logf("Transfer progress: %.2f%% (%d/%d files)",
				lastProgress,
				state.Stats.CompletedItems,
				state.Stats.TotalItems)
		}
	}

	// Configure resumable transfer options
	options := transfer.DefaultResumableTransferOptions()
	options.BatchSize = 2 // Small batch size for testing
	options.ProgressCallback = progressCallback
	options.CheckpointInterval = 2 * time.Second
	options.SyncLevel = 1 // Verify file size
	options.PreserveMtime = true

	// Create resumable transfer with retry
	t.Log("Creating resumable transfer")
	var checkpointID string
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var err error
			checkpointID, err = client.CreateResumableTransfer(
				ctx,
				sourceEndpointID, sourceDir,
				destEndpointID, destDir,
				options,
			)
			return err
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create resumable transfer: %v", err)
	}

	t.Logf("Created resumable transfer with checkpoint ID: %s", checkpointID)

	// Get initial checkpoint state
	var state *transfer.CheckpointState
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var err error
			state, err = client.GetTransferCheckpoint(ctx, checkpointID)
			return err
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to get initial checkpoint state: %v", err)
	}

	t.Logf("Initial checkpoint state: %d files to transfer", state.Stats.TotalItems)

	if state.Stats.TotalItems != fileCount {
		t.Errorf("Expected %d files to transfer, got %d", fileCount, state.Stats.TotalItems)
	}

	// Resume the transfer with retry
	t.Log("Resuming transfer")
	var result *transfer.ResumableTransferResult
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var err error
			result, err = client.ResumeTransfer(ctx, checkpointID, nil)
			return err
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to resume transfer: %v", err)
	}

	// Verify the result
	t.Logf("Transfer result: Completed=%v, CompletedItems=%d, FailedItems=%d, TaskIDs=%v",
		result.Completed, result.CompletedItems, result.FailedItems, result.TaskIDs)

	if !result.Completed {
		t.Errorf("Expected transfer to complete, but it did not")
	}

	if result.FailedItems > 0 {
		t.Errorf("Expected 0 failed items, got %d", result.FailedItems)
	}

	if result.CompletedItems != fileCount {
		t.Errorf("Expected %d completed items, got %d", fileCount, result.CompletedItems)
	}

	// Verify progress callback was called
	if !progressCalled {
		t.Errorf("Progress callback was not called")
	} else {
		t.Logf("Final progress: %.2f%%", lastProgress)
	}

	// List files in destination directory to verify the transfer
	t.Log("Verifying files in destination directory")
	var files *transfer.FileList
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var err error
			files, err = client.ListFiles(ctx, destEndpointID, destDir, nil)
			return err
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to list files in destination directory: %v", err)
	}

	if len(files.Data) != fileCount {
		t.Errorf("Expected %d files in destination directory, got %d", fileCount, len(files.Data))
	} else {
		t.Logf("Successfully verified %d files in destination directory", fileCount)
		for i, file := range files.Data {
			t.Logf("  File %d: %s (%d bytes)", i+1, file.Name, file.Size)
		}
	}

	// Clean up the checkpoint
	t.Log("Cleaning up checkpoint")
	storage, err := transfer.NewFileCheckpointStorage("")
	if err != nil {
		t.Fatalf("Failed to create checkpoint storage: %v", err)
	}

	err = storage.DeleteCheckpoint(ctx, checkpointID)
	if err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	t.Log("Resumable transfer test completed successfully")
}

// TestIntegration_ResumableTransferCancellation tests cancellation of resumable transfers
func TestIntegration_ResumableTransferCancellation(t *testing.T) {
	// Skip this test during automated CI runs if credentials aren't provided
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if endpoints are not provided
	if sourceEndpointID == "" || destEndpointID == "" {
		t.Skip("Integration test requires source and destination endpoint IDs")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client with new pattern
	client, err := transfer.NewClient(
		transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
	)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Set up test directories with timestamp
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("/~/resumable-cancel-source-%s", timestamp)
	destDir := fmt.Sprintf("/~/resumable-cancel-dest-%s", timestamp)

	// Create source directory with retry
	t.Logf("Creating source directory: %s", sourceDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, sourceEndpointID, sourceDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Set up cleanup for source directory
	defer func() {
		t.Logf("Cleaning up source test directory: %s", sourceDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup source test dir %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     sourceDir,
				},
			},
		}

		_ = ratelimit.RetryWithBackoff(
			context.Background(),
			func(ctx context.Context) error {
				_, err := client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
	}()

	// Create destination directory with retry
	t.Logf("Creating destination directory: %s", destDir)
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			return client.Mkdir(ctx, destEndpointID, destDir)
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Set up cleanup for destination directory
	defer func() {
		t.Logf("Cleaning up destination test directory: %s", destDir)

		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Cleanup dest test dir %s", timestamp),
			EndpointID: destEndpointID,
			Items: []transfer.DeleteItem{
				{
					DataType: "delete_item",
					Path:     destDir,
				},
			},
		}

		_ = ratelimit.RetryWithBackoff(
			context.Background(),
			func(ctx context.Context) error {
				_, err := client.CreateDeleteTask(ctx, deleteRequest)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
	}()

	// Create a few test files
	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("%s/cancel-file%d.txt", sourceDir, i)
		fileContent := fmt.Sprintf("Cancel test file %d", i)

		// Create file with retry
		var taskID string
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				taskID, err = client.SubmitTextFileCreation(
					ctx,
					sourceEndpointID,
					fileName,
					fileContent,
					fmt.Sprintf("Cancel test file %d creation", i),
				)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
		if err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		// Wait for completion with retry
		var task *transfer.Task
		err = ratelimit.RetryWithBackoff(
			ctx,
			func(ctx context.Context) error {
				var err error
				task, err = client.WaitForTaskCompletion(ctx, taskID, 1*time.Second)
				return err
			},
			ratelimit.DefaultBackoff(),
			transfer.IsRetryableTransferError,
		)
		if err != nil {
			t.Fatalf("Failed to wait for file creation task: %v", err)
		}

		if task.Status != "SUCCEEDED" {
			t.Fatalf("File creation task failed with status: %s", task.Status)
		}
	}

	// Configure options for cancellation test
	options := transfer.DefaultResumableTransferOptions()
	options.BatchSize = 1 // Small batch size for testing

	// Create resumable transfer with retry
	t.Log("Creating resumable transfer for cancellation test")
	var checkpointID string
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var err error
			checkpointID, err = client.CreateResumableTransfer(
				ctx,
				sourceEndpointID, sourceDir,
				destEndpointID, destDir,
				options,
			)
			return err
		},
		ratelimit.DefaultBackoff(),
		transfer.IsRetryableTransferError,
	)
	if err != nil {
		t.Fatalf("Failed to create resumable transfer: %v", err)
	}

	t.Logf("Created resumable transfer with checkpoint ID: %s", checkpointID)

	// Start the transfer in a goroutine, then immediately cancel it
	go func() {
		// Wait a short time to ensure the checkpoint is created
		time.Sleep(1 * time.Second)

		// Delete the checkpoint to simulate cancellation
		t.Log("Cancelling transfer by deleting checkpoint")
		storage, err := transfer.NewFileCheckpointStorage("")
		if err != nil {
			t.Errorf("Failed to create checkpoint storage: %v", err)
			return
		}

		err = storage.DeleteCheckpoint(ctx, checkpointID)
		if err != nil {
			t.Errorf("Failed to delete checkpoint: %v", err)
		}
	}()

	// Try to resume the transfer, which should fail after the checkpoint is deleted
	_, err = client.ResumeTransfer(ctx, checkpointID, nil)

	// We expect an error since the checkpoint should be deleted
	if err == nil {
		t.Errorf("Expected error after cancellation, but got none")
	} else {
		t.Logf("Got expected error after cancellation: %v", err)
	}

	// Verify the checkpoint is gone
	storage, err := transfer.NewFileCheckpointStorage("")
	if err != nil {
		t.Fatalf("Failed to create checkpoint storage: %v", err)
	}

	checkpoints, err := storage.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	for _, id := range checkpoints {
		if id == checkpointID {
			t.Errorf("Checkpoint %s still exists after cancellation", checkpointID)
			return
		}
	}

	t.Log("Checkpoint was successfully deleted, cancellation test passed")
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
package transfer_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// TestIntegration_ResumableTransfer tests the resumable transfer functionality
func TestIntegration_ResumableTransfer(t *testing.T) {
	// Skip this test during automated CI runs if credentials aren't provided
	clientID, clientSecret, sourceEndpointID, destEndpointID := getTestCredentials(t)

	// Skip if endpoints are not provided 
	if sourceEndpointID == "" || destEndpointID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_SOURCE_ENDPOINT_ID and GLOBUS_TEST_DEST_ENDPOINT_ID")
	}

	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)

	// Create Transfer client
	client := transfer.NewClient(accessToken)
	ctx := context.Background()

	// Set up test directories
	timestamp := time.Now().Format("20060102_150405")
	sourceDir := fmt.Sprintf("/~/resumable-test-source-%s", timestamp)
	destDir := fmt.Sprintf("/~/resumable-test-dest-%s", timestamp)

	// Create source directory
	err := client.Mkdir(ctx, sourceEndpointID, sourceDir)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	t.Logf("Created source directory: %s", sourceDir)

	// Create destination directory
	err = client.Mkdir(ctx, destEndpointID, destDir)
	if err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	t.Logf("Created destination directory: %s", destDir)

	// Clean up directories when test completes
	defer func() {
		// Delete the source and destination directories
		deleteRequest := &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Delete test directories %s", timestamp),
			EndpointID: sourceEndpointID,
			Items: []transfer.DeleteItem{
				{
					Path: sourceDir,
				},
			},
		}
		_, err := client.CreateDeleteTask(ctx, deleteRequest)
		if err != nil {
			t.Logf("Warning: Failed to delete source directory: %v", err)
		} else {
			t.Logf("Cleaned up source directory: %s", sourceDir)
		}

		deleteRequest = &transfer.DeleteTaskRequest{
			DataType:   "delete",
			Label:      fmt.Sprintf("Delete test directories %s", timestamp),
			EndpointID: destEndpointID,
			Items: []transfer.DeleteItem{
				{
					Path: destDir,
				},
			},
		}
		_, err = client.CreateDeleteTask(ctx, deleteRequest)
		if err != nil {
			t.Logf("Warning: Failed to delete destination directory: %v", err)
		} else {
			t.Logf("Cleaned up destination directory: %s", destDir)
		}
	}()

	// Create nested directory structure
	for _, subpath := range []string{"/subdir1", "/subdir1/nested1", "/subdir2", "/subdir3"} {
		err = client.Mkdir(ctx, sourceEndpointID, sourceDir+subpath)
		if err != nil {
			t.Fatalf("Failed to create nested directory %s: %v", subpath, err)
		}
		t.Logf("Created nested directory: %s%s", sourceDir, subpath)
	}

	// Set up progress callback for testing
	var lastProgress float64
	progressCalled := false
	progressCallback := func(state *transfer.CheckpointState) {
		progressCalled = true
		progress := float64(0)
		if state.Stats.TotalItems > 0 {
			progress = float64(state.Stats.CompletedItems) / float64(state.Stats.TotalItems) * 100
		}
		t.Logf("Progress: %.2f%% (%d/%d files)", 
			progress, 
			state.Stats.CompletedItems,
			state.Stats.TotalItems)
		lastProgress = progress
	}

	// Set up options for resumable transfer
	options := transfer.DefaultResumableTransferOptions()
	options.BatchSize = 10 // Small batch size for testing
	options.ProgressCallback = progressCallback
	options.CheckpointInterval = time.Second * 5

	// Start the resumable transfer
	t.Log("Starting resumable transfer")
	checkpointID, err := client.SubmitResumableTransfer(
		ctx,
		sourceEndpointID, sourceDir,
		destEndpointID, destDir,
		options,
	)
	if err != nil {
		t.Fatalf("Failed to submit resumable transfer: %v", err)
	}
	t.Logf("Started resumable transfer with checkpoint ID: %s", checkpointID)

	// Wait a moment to allow the transfer to initialize
	time.Sleep(time.Second * 3)

	// Get transfer status
	state, err := client.GetResumableTransferStatus(ctx, checkpointID)
	if err != nil {
		t.Fatalf("Failed to get transfer status: %v", err)
	}
	
	t.Logf("Transfer status: Total items: %d, Completed: %d, Failed: %d, Pending: %d",
		state.Stats.TotalItems, state.Stats.CompletedItems, 
		state.Stats.FailedItems, state.Stats.RemainingItems)

	// Verify checkpoint storage is working
	storage, err := transfer.NewFileCheckpointStorage("")
	if err != nil {
		t.Fatalf("Failed to create checkpoint storage: %v", err)
	}

	checkpoints, err := storage.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	// Check if our checkpoint is in the list
	found := false
	for _, id := range checkpoints {
		if id == checkpointID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Checkpoint %s not found in checkpoint storage", checkpointID)
	} else {
		t.Logf("Checkpoint verified in storage: %s", checkpointID)
	}

	// Resume the transfer
	result, err := client.ResumeResumableTransfer(ctx, checkpointID, options)
	if err != nil {
		t.Fatalf("Failed to resume transfer: %v", err)
	}

	t.Logf("Transfer completed: %d completed, %d failed, %d tasks created",
		result.CompletedItems, result.FailedItems, len(result.TaskIDs))

	// Verify progress callback was called
	if !progressCalled {
		t.Errorf("Progress callback was not called")
	} else {
		t.Logf("Progress callback was called, final progress: %.2f%%", lastProgress)
	}

	// Verify directory structure was transferred
	listOptions := &transfer.ListFileOptions{
		ShowHidden: true,
	}
	
	// List the destination directory to verify files were transferred
	destFiles, err := client.ListFiles(ctx, destEndpointID, destDir, listOptions)
	if err != nil {
		t.Fatalf("Failed to list destination directory: %v", err)
	}

	// Check if subdirectories were transferred
	subdirCount := 0
	for _, file := range destFiles.Data {
		if file.Type == "dir" {
			subdirCount++
			t.Logf("Found directory in destination: %s", file.Name)
		}
	}

	// We expect subdir1, subdir2, and subdir3 at minimum
	if subdirCount < 3 {
		t.Errorf("Expected at least 3 subdirectories, found %d", subdirCount)
	} else {
		t.Logf("Found %d subdirectories in destination", subdirCount)
	}

	// Test canceling a transfer
	// Start a new transfer for testing cancellation
	cancelCheckpointID, err := client.SubmitResumableTransfer(
		ctx,
		sourceEndpointID, sourceDir,
		destEndpointID, destDir+"-cancel-test",
		options,
	)
	if err != nil {
		t.Fatalf("Failed to submit transfer for cancellation test: %v", err)
	}
	t.Logf("Started transfer for cancellation test: %s", cancelCheckpointID)

	// Wait a moment to allow the transfer to initialize
	time.Sleep(time.Second * 2)

	// Cancel the transfer
	err = client.CancelResumableTransfer(ctx, cancelCheckpointID)
	if err != nil {
		t.Fatalf("Failed to cancel transfer: %v", err)
	}
	t.Logf("Successfully canceled transfer: %s", cancelCheckpointID)

	// Verify the checkpoint was deleted
	checkpoints, err = storage.ListCheckpoints(ctx)
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	// Check if the canceled checkpoint is still in the list
	found = false
	for _, id := range checkpoints {
		if id == cancelCheckpointID {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Checkpoint %s still exists after cancellation", cancelCheckpointID)
	} else {
		t.Logf("Checkpoint was successfully deleted after cancellation")
	}

	// Clean up the checkpoint from the completed transfer
	err = client.CancelResumableTransfer(ctx, checkpointID)
	if err != nil {
		t.Logf("Warning: Failed to clean up checkpoint: %v", err)
	} else {
		t.Logf("Successfully cleaned up checkpoint: %s", checkpointID)
	}
}
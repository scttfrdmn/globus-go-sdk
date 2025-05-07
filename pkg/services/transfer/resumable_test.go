// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

func TestCreateResumableTransfer(t *testing.T) {
	// Create a temporary directory for checkpoint files
	tempDir, err := os.MkdirTemp("", "globus-go-sdk-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/operation/endpoint/src-endpoint/ls" {
			// Return a sample file listing
			files := FileList{
				Data: []FileListItem{
					{
						DataType:     "file",
						Name:         "file1.txt",
						Type:         "file",
						Size:         1024,
						LastModified: time.Now().Format(time.RFC3339),
					},
					{
						DataType:     "file",
						Name:         "file2.txt",
						Type:         "file",
						Size:         2048,
						LastModified: time.Now().Format(time.RFC3339),
					},
				},
				EndpointID:  "src-endpoint",
				Path:        "/source",
				HasNextPage: false,
			}
			json.NewEncoder(w).Encode(files)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	// Create client with the new pattern
	client, err := NewClient(
		WithAuthorizer(mockAuthorizer("fake-token")),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test creating a resumable transfer
	options := DefaultResumableTransferOptions()
	options.BatchSize = 10

	checkpointID, err := client.CreateResumableTransfer(
		context.Background(),
		"src-endpoint", "/source",
		"dest-endpoint", "/destination",
		options,
	)

	if err != nil {
		t.Fatalf("Failed to create resumable transfer: %v", err)
	}

	if checkpointID == "" {
		t.Fatalf("Expected a checkpoint ID, got empty string")
	}

	// Check if checkpoint file was created
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		t.Fatalf("Failed to create checkpoint storage: %v", err)
	}

	state, err := storage.LoadCheckpoint(context.Background(), checkpointID)
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify checkpoint state
	if state.TaskInfo.SourceEndpointID != "src-endpoint" {
		t.Errorf("Expected source endpoint 'src-endpoint', got '%s'", state.TaskInfo.SourceEndpointID)
	}

	if state.TaskInfo.DestinationEndpointID != "dest-endpoint" {
		t.Errorf("Expected destination endpoint 'dest-endpoint', got '%s'", state.TaskInfo.DestinationEndpointID)
	}

	if len(state.PendingItems) != 2 {
		t.Errorf("Expected 2 pending items, got %d", len(state.PendingItems))
	}

	// Check if DATA_TYPE is properly set for transfer items
	for i, item := range state.PendingItems {
		if item.DataType != "transfer_item" {
			t.Errorf("Pending item %d has incorrect DATA_TYPE: expected 'transfer_item', got '%s'", i, item.DataType)
		}
	}

	// Clean up checkpoint
	if err := storage.DeleteCheckpoint(context.Background(), checkpointID); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}
}

func TestCheckpointStorage(t *testing.T) {
	// Create a temporary directory for checkpoint files
	tempDir, err := os.MkdirTemp("", "globus-go-sdk-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	storage, err := NewFileCheckpointStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create checkpoint storage: %v", err)
	}

	// Create checkpoint state
	state := &CheckpointState{
		CheckpointID: "test-checkpoint",
	}
	state.TaskInfo.SourceEndpointID = "src-endpoint"
	state.TaskInfo.DestinationEndpointID = "dest-endpoint"
	state.TaskInfo.SourceBasePath = "/source"
	state.TaskInfo.DestinationBasePath = "/destination"
	state.TaskInfo.Label = "Test Transfer"
	state.TaskInfo.StartTime = time.Now()
	state.TaskInfo.LastUpdated = time.Now()

	// Add some pending items
	state.PendingItems = []TransferItem{
		{
			DataType:        "transfer_item",
			SourcePath:      "/source/file1.txt",
			DestinationPath: "/destination/file1.txt",
		},
		{
			DataType:        "transfer_item",
			SourcePath:      "/source/file2.txt",
			DestinationPath: "/destination/file2.txt",
		},
	}

	// Save checkpoint
	if err := storage.SaveCheckpoint(context.Background(), state); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify file was created
	checkpointFile := filepath.Join(tempDir, "test-checkpoint.json")
	if _, err := os.Stat(checkpointFile); os.IsNotExist(err) {
		t.Fatalf("Checkpoint file was not created")
	}

	// List checkpoints
	checkpoints, err := storage.ListCheckpoints(context.Background())
	if err != nil {
		t.Fatalf("Failed to list checkpoints: %v", err)
	}

	if len(checkpoints) != 1 || checkpoints[0] != "test-checkpoint" {
		t.Errorf("Expected ['test-checkpoint'], got %v", checkpoints)
	}

	// Load checkpoint
	loadedState, err := storage.LoadCheckpoint(context.Background(), "test-checkpoint")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify loaded state
	if loadedState.CheckpointID != "test-checkpoint" {
		t.Errorf("Expected checkpoint ID 'test-checkpoint', got '%s'", loadedState.CheckpointID)
	}

	if loadedState.TaskInfo.SourceEndpointID != "src-endpoint" {
		t.Errorf("Expected source endpoint 'src-endpoint', got '%s'", loadedState.TaskInfo.SourceEndpointID)
	}

	if len(loadedState.PendingItems) != 2 {
		t.Errorf("Expected 2 pending items, got %d", len(loadedState.PendingItems))
	}

	// Verify DATA_TYPE field in loaded items
	for i, item := range loadedState.PendingItems {
		if item.DataType != "transfer_item" {
			t.Errorf("Loaded pending item %d has incorrect DATA_TYPE: expected 'transfer_item', got '%s'", i, item.DataType)
		}
	}

	// Delete checkpoint
	if err := storage.DeleteCheckpoint(context.Background(), "test-checkpoint"); err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(checkpointFile); !os.IsNotExist(err) {
		t.Errorf("Checkpoint file was not deleted")
	}
}

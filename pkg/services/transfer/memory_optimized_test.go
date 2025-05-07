// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// WaitOptions contains options for waiting on transfer tasks
type WaitOptions struct {
	// PollInterval is the time to wait between status checks
	PollInterval time.Duration
	// Timeout is the maximum time to wait
	Timeout time.Duration
	// ProgressCallback is called with progress updates
	ProgressCallback func(completed, total int, message string)
}

// WaitForMemoryOptimizedTransfer waits for all tasks in a memory-optimized transfer to complete
func (c *Client) WaitForMemoryOptimizedTransfer(
	ctx context.Context,
	result *MemoryOptimizedTransferResult,
	options *WaitOptions,
) error {
	if len(result.TaskIDs) == 0 {
		return nil // Nothing to wait for
	}

	// Create a context with timeout if specified
	var cancel context.CancelFunc
	if options != nil && options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Use default poll interval if not specified
	pollInterval := 1 * time.Second
	if options != nil && options.PollInterval > 0 {
		pollInterval = options.PollInterval
	}

	// Create a ticker for polling
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	total := len(result.TaskIDs)
	completed := 0
	taskStatus := make(map[string]string)

	// Report initial progress
	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(completed, total, "Starting to wait for tasks")
	}

	// Check each task until all are complete or context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check each task that's not yet complete
			for _, taskID := range result.TaskIDs {
				if isDone(taskStatus[taskID]) {
					continue
				}

				// Get task status
				task, err := c.GetTask(ctx, taskID)
				if err != nil {
					return fmt.Errorf("error checking task %s: %w", taskID, err)
				}

				// Update status
				taskStatus[taskID] = task.Status

				// Check if newly completed
				if isDone(task.Status) && !isDone(taskStatus[taskID]) {
					completed++
					// Report progress
					if options != nil && options.ProgressCallback != nil {
						options.ProgressCallback(completed, total,
							fmt.Sprintf("Task %s completed with status: %s", taskID, task.Status))
					}
				}
			}

			// Check if all tasks are complete
			allDone := true
			for _, taskID := range result.TaskIDs {
				if !isDone(taskStatus[taskID]) {
					allDone = false
					break
				}
			}

			if allDone {
				return nil
			}
		}
	}
}

// Helper function to check if a task status indicates it's done
func isDone(status string) bool {
	return status == "SUCCEEDED" || status == "FAILED" || status == "CANCELLED"
}

func TestMemoryOptimizedTransfer(t *testing.T) {
	// Create a mock server to simulate Transfer API
	var submittedTasks []TransferTaskRequest
	var taskIDCounter int
	var serverMutex sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle different paths
		switch {
		case r.URL.Path == "/v0.10/submission_id" && r.Method == http.MethodGet:
			// Return a mock submission ID
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"value": "submission-id-123456"}`)

		case r.URL.Path == "/v0.10/operation/endpoint/mock-source/ls" && r.Method == http.MethodGet:
			// Handle listing files in source directory
			// This is called by the memory optimized transfer function to get the file list
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"data": [
					{"data_type": "file", "name": "file1.txt", "type": "file", "size": 1024, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file2.txt", "type": "file", "size": 2048, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file3.txt", "type": "file", "size": 3072, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file4.txt", "type": "file", "size": 4096, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file5.txt", "type": "file", "size": 5120, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file6.txt", "type": "file", "size": 6144, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file7.txt", "type": "file", "size": 7168, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file8.txt", "type": "file", "size": 8192, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file9.txt", "type": "file", "size": 9216, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file10.txt", "type": "file", "size": 10240, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file11.txt", "type": "file", "size": 11264, "last_modified": "2021-01-01T00:00:00Z"},
					{"data_type": "file", "name": "file12.txt", "type": "file", "size": 12288, "last_modified": "2021-01-01T00:00:00Z"}
				],
				"endpoint_id": "mock-source",
				"path": "/source",
				"data_type": "file_list",
				"has_next_page": false
			}`)

		case r.URL.Path == "/v0.10/transfer" && r.Method == http.MethodPost:
			// Handle transfer submission
			var options TransferTaskRequest
			err := json.NewDecoder(r.Body).Decode(&options)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"code": "BadRequest", "message": "Invalid request body"}`)
				return
			}

			serverMutex.Lock()
			taskIDCounter++
			taskID := fmt.Sprintf("task-%d", taskIDCounter)
			// Store for verification
			submittedTasks = append(submittedTasks, options)
			serverMutex.Unlock()

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"task_id": "%s", "message": "The transfer has been accepted and a task has been created", "code": "Accepted"}`, taskID)

		case strings.HasPrefix(r.URL.Path, "/v0.10/task/") && r.Method == http.MethodGet:
			// Handle task status checks
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 4 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			taskID := parts[3]

			// Simulate task progress - all tasks complete immediately for testing
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"task_id": "%s",
				"status": "SUCCEEDED",
				"bytes_transferred": 1024000,
				"subtasks_succeeded": 1,
				"subtasks_total": 1,
				"files_transferred": 10,
				"type": "TRANSFER"
			}`, taskID)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code": "NotFound", "message": "Unknown endpoint: %s"}`, r.URL.Path)
		}
	}))
	defer server.Close()

	// Create a client using the new pattern
	client, err := NewClient(
		WithAuthorizer(mockAuthorizer("fake-token")),
		WithCoreOption(core.WithBaseURL(server.URL+"/v0.10/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("Memory optimization batching", func(t *testing.T) {
		// Create a test context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// We don't need to manually create a list of files
		// The SubmitMemoryOptimizedTransfer will get the file list from the source endpoint

		// Submit memory-optimized transfer
		result, err := client.SubmitMemoryOptimizedTransfer(
			ctx,
			"mock-source", "/source",
			"mock-dest", "/dest",
			&MemoryOptimizedOptions{
				BatchSize:          5,
				MaxConcurrentTasks: 2,
				Label:              "Test Memory-Optimized Transfer",
				SyncLevel:          SyncLevelChecksum,
				VerifyChecksum:     true,
				ProgressCallback: func(processed, total int, bytes int64, message string) {
					t.Logf("Progress: %d files, %d bytes, %s", processed, bytes, message)
				},
			},
		)

		if err != nil {
			t.Fatalf("SubmitMemoryOptimizedTransfer failed: %v", err)
		}

		// Wait for transfers to complete
		err = client.WaitForMemoryOptimizedTransfer(ctx, result, &WaitOptions{
			PollInterval: 100 * time.Millisecond,
			Timeout:      10 * time.Second,
			ProgressCallback: func(completed, total int, message string) {
				t.Logf("Wait progress: %d/%d tasks, %s", completed, total, message)
			},
		})
		if err != nil {
			t.Fatalf("WaitForMemoryOptimizedTransfer failed: %v", err)
		}

		// Verify results
		if len(result.TaskIDs) == 0 {
			t.Errorf("Expected at least one task ID, got none")
		}

		// Verify that no more than 5 files were included in each batch
		serverMutex.Lock()
		for i, task := range submittedTasks {
			if len(task.Items) > 5 {
				t.Errorf("Task %d has %d items, exceeding batch size of 5", i+1, len(task.Items))
			}
		}
		serverMutex.Unlock()
	})
}

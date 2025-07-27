// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
	
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer/testutils"
)

func TestMemoryOptimizedTransfer(t *testing.T) {
	// Create a mock server to simulate Transfer API
	var submittedTasks []SubmitTransferOptions
	var taskIDCounter int
	var serverMutex sync.Mutex
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Handle different paths
		switch {
		case r.URL.Path == "/transfer" && r.Method == http.MethodPost:
			// Handle transfer submission
			var options SubmitTransferOptions
			// Parse request body (omitted for brevity)
			
			serverMutex.Lock()
			taskIDCounter++
			taskID := fmt.Sprintf("task-%d", taskIDCounter)
			submittedTasks = append(submittedTasks, options)
			serverMutex.Unlock()
			
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{"task_id": "%s", "message": "The transfer has been accepted and a task has been created"}`, taskID)
			
		case strings.HasPrefix(r.URL.Path, "/task/") && r.Method == http.MethodGet:
			// Handle task status checks
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 3 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			
			taskID := parts[2]
			
			// Simulate task progress - all tasks complete immediately for testing
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"task_id": "%s",
				"status": "SUCCEEDED",
				"bytes_transferred": 1024000,
				"subtasks_succeeded": 1,
				"subtasks_total": 1,
				"files_transferred": 10
			}`, taskID)
			
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	
	// Create a client that uses the test server
	httpClient := &http.Client{}
	transport := core.NewHTTPTransport(server.URL, httpClient)
	client := &Client{
		Transport: transport,
	}
	
	t.Run("Memory optimization benchmarking", func(t *testing.T) {
		// Create a memory tracker to measure usage
		memoryTracker := testutils.NewMemoryTracker()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		// Start tracking memory usage
		go memoryTracker.TrackMemoryUsage(ctx, t, 100)
		
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Create a list of test files
		files := make([]FileTransfer, 0, 15)
		for i := 0; i < 15; i++ {
			files = append(files, FileTransfer{
				SourcePath:      fmt.Sprintf("/source/file%d.txt", i),
				DestinationPath: fmt.Sprintf("/dest/file%d.txt", i),
				Size:            1024 * 1024, // 1MB files
			})
		}
		
		// Submit memory-optimized transfer
		result, err := client.SubmitMemoryOptimizedTransfer(
			ctx,
			"mock-source", "/source",
			"mock-dest", "/dest",
			&MemoryOptimizedOptions{
				BatchSize:        5,
				MaxConcurrentTasks: 2,
				Label:             "Test Memory-Optimized Transfer",
				SyncLevel:         SyncChecksum,
				VerifyChecksum:    true,
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
		
		// Print memory usage summary
		cancel() // Stop memory tracking
		
		// Log memory usage
		t.Logf("Peak memory usage: %.2f MB", float64(memoryTracker.MaxUsage)/(1024*1024))
		
		// Verify results
		if len(result.TaskIDs) == 0 {
			t.Errorf("Expected at least one task ID, got none")
		}
		
		// Verify that no more than 5 files were included in each batch
		serverMutex.Lock()
		for i, task := range submittedTasks {
			if len(task.TransferItems) > 5 {
				t.Errorf("Task %d has %d items, exceeding batch size of 5", i+1, len(task.TransferItems))
			}
		}
		serverMutex.Unlock()
	})
}
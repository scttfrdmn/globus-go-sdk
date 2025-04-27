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
	
	"github.com/scttfrdmn/globus-go-sdk/pkg/benchmark"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
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
		case strings.Contains(r.URL.Path, "/endpoint/mock-source/ls"):
			// Return directory listing based on path parameter
			path := r.URL.Query().Get("path")
			
			switch path {
			case "/source":
				// Return root directory with 10 files and 2 subdirectories
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file1.txt", "type": "file", "size": 1024, "last_modified": "2021-01-01T00:00:00Z", "path": "file1.txt"},
						{"name": "file2.txt", "type": "file", "size": 2048, "last_modified": "2021-01-01T00:00:00Z", "path": "file2.txt"},
						{"name": "file3.txt", "type": "file", "size": 3072, "last_modified": "2021-01-01T00:00:00Z", "path": "file3.txt"},
						{"name": "file4.txt", "type": "file", "size": 4096, "last_modified": "2021-01-01T00:00:00Z", "path": "file4.txt"},
						{"name": "file5.txt", "type": "file", "size": 5120, "last_modified": "2021-01-01T00:00:00Z", "path": "file5.txt"},
						{"name": "dir1", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1"},
						{"name": "dir2", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2"}
					],
					"endpoint": "mock-source",
					"path": "/source",
					"DATA_TYPE": "file_list"
				}`)
			case "/source/dir1":
				// Return dir1 with 5 files
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file6.txt", "type": "file", "size": 6144, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1/file6.txt"},
						{"name": "file7.txt", "type": "file", "size": 7168, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1/file7.txt"},
						{"name": "file8.txt", "type": "file", "size": 8192, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1/file8.txt"},
						{"name": "file9.txt", "type": "file", "size": 9216, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1/file9.txt"},
						{"name": "file10.txt", "type": "file", "size": 10240, "last_modified": "2021-01-01T00:00:00Z", "path": "dir1/file10.txt"}
					],
					"endpoint": "mock-source",
					"path": "/source/dir1",
					"DATA_TYPE": "file_list"
				}`)
			case "/source/dir2":
				// Return dir2 with 5 more files
				fmt.Fprint(w, `{
					"DATA": [
						{"name": "file11.txt", "type": "file", "size": 11264, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2/file11.txt"},
						{"name": "file12.txt", "type": "file", "size": 12288, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2/file12.txt"},
						{"name": "file13.txt", "type": "file", "size": 13312, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2/file13.txt"},
						{"name": "file14.txt", "type": "file", "size": 14336, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2/file14.txt"},
						{"name": "file15.txt", "type": "file", "size": 15360, "last_modified": "2021-01-01T00:00:00Z", "path": "dir2/file15.txt"}
					],
					"endpoint": "mock-source",
					"path": "/source/dir2",
					"DATA_TYPE": "file_list"
				}`)
			default:
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, `{"code": "NotFound", "message": "Path %s not found"}`, path)
			}
			
		case r.URL.Path == "/v0.10/submission_id":
			// Return a submission ID
			fmt.Fprint(w, `{"value": "mock-submission-id", "DATA_TYPE": "submission_id"}`)
			
		case r.URL.Path == "/v0.10/transfer":
			// Handle transfer submission
			serverMutex.Lock()
			taskIDCounter++
			taskID := fmt.Sprintf("TASK-%d", taskIDCounter)
			
			// Track the transfer options
			var transferOptions SubmitTransferOptions
			if err := core.JSONDecode(r.Body, &transferOptions); err != nil {
				serverMutex.Unlock()
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"code": "BadRequest", "message": "Invalid transfer request: %v"}`, err)
				return
			}
			
			submittedTasks = append(submittedTasks, transferOptions)
			serverMutex.Unlock()
			
			// Return task result
			fmt.Fprintf(w, `{"task_id": "%s", "DATA_TYPE": "transfer_result"}`, taskID)
			
		case strings.Contains(r.URL.Path, "/task/"):
			// Return task status
			parts := strings.Split(r.URL.Path, "/")
			taskID := parts[len(parts)-1]
			
			// Always return success in tests
			fmt.Fprintf(w, `{
				"DATA_TYPE": "task",
				"task_id": "%s",
				"status": "SUCCEEDED",
				"bytes_transferred": 123456,
				"bytes_expected": 123456,
				"files_transferred": 15,
				"files_expected": 15
			}`, taskID)
			
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"code": "NotFound", "message": "Unknown endpoint: %s"}`, r.URL.Path)
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
		// Create a memory sampler to measure usage
		memorySampler := benchmark.NewMemorySampler(100 * time.Millisecond)
		memorySampler.Start()
		defer memorySampler.Stop()
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Execute memory-optimized transfer
		result, err := client.SubmitMemoryOptimizedTransfer(
			ctx,
			"mock-source", "/source",
			"mock-dest", "/dest",
			&MemoryOptimizedOptions{
				BatchSize:         5,
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
		memorySampler.Stop()
		memorySampler.PrintSummary()
		
		// Verify results
		if len(result.TaskIDs) == 0 {
			t.Errorf("Expected at least one task ID, got none")
		}
		
		// Expect 15 files
		if result.FilesTransferred != 15 {
			t.Errorf("Expected 15 files transferred, got %d", result.FilesTransferred)
		}
		
		// Verify that no more than 5 files were included in each batch
		serverMutex.Lock()
		for i, task := range submittedTasks {
			if len(task.TransferItems) > 5 {
				t.Errorf("Task %d has %d items, exceeding batch size of 5", i+1, len(task.TransferItems))
			}
		}
		serverMutex.Unlock()
		
		// Log peak memory usage
		peakMemory := memorySampler.GetPeakMemory()
		t.Logf("Peak memory usage: %.2f MB", peakMemory)
	})
}

// TestMemoryComparison performs a comparison of memory usage between regular and optimized transfers
func TestMemoryComparison(t *testing.T) {
	// This test simulates a large file list to compare memory usage
	// between the regular and optimized implementations
	
	// Skip in short test mode as this is more of a benchmark
	if testing.Short() {
		t.Skip("Skipping memory comparison test in short mode")
	}
	
	// Create a mock server that generates a large directory listing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if strings.Contains(r.URL.Path, "/endpoint/mock-source/ls") {
			// Generate a large listing dynamically
			path := r.URL.Query().Get("path")
			depth := 0
			
			// Extract depth from path (e.g., /source/dir1/dir2 has depth 2)
			if path != "/source" {
				depth = len(strings.Split(strings.TrimPrefix(path, "/source/"), "/"))
			}
			
			// Don't go deeper than 3 levels
			if depth > 3 {
				fmt.Fprint(w, `{"DATA": [], "endpoint": "mock-source", "path": "`+path+`", "DATA_TYPE": "file_list"}`)
				return
			}
			
			// Start building response
			w.Write([]byte(`{"DATA": [`))
			
			// Generate 100 files
			for i := 0; i < 100; i++ {
				if i > 0 {
					w.Write([]byte(","))
				}
				
				filePath := fmt.Sprintf("%s/file%d.txt", path, i)
				filePath = strings.TrimPrefix(filePath, "/source/")
				if filePath == "/sourcefile0.txt" {
					filePath = "file0.txt"
				}
				
				fileEntry := fmt.Sprintf(`
					{"name": "file%d.txt", "type": "file", "size": %d, "last_modified": "2021-01-01T00:00:00Z", "path": "%s"}`,
					i, i*1024, filePath)
				w.Write([]byte(fileEntry))
			}
			
			// Add subdirectories (but only if we're not too deep)
			if depth < 3 {
				for i := 0; i < 10; i++ {
					dirEntry := fmt.Sprintf(`,
						{"name": "dir%d", "type": "dir", "size": 0, "last_modified": "2021-01-01T00:00:00Z", "path": "%s/dir%d"}`,
						i, strings.TrimPrefix(path, "/source/"), i)
					w.Write([]byte(dirEntry))
				}
			}
			
			// Close the response
			w.Write([]byte(`], "endpoint": "mock-source", "path": "`+path+`", "DATA_TYPE": "file_list"}`))
		} else if r.URL.Path == "/v0.10/submission_id" {
			fmt.Fprint(w, `{"value": "mock-submission-id", "DATA_TYPE": "submission_id"}`)
		} else if r.URL.Path == "/v0.10/transfer" {
			fmt.Fprint(w, `{"task_id": "mock-task-id", "DATA_TYPE": "transfer_result"}`)
		} else {
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
	
	ctx := context.Background()
	
	// Measure memory for standard recursive transfer
	standardMemorySampler := benchmark.NewMemorySampler(10 * time.Millisecond)
	standardMemorySampler.Start()
	
	_, err := client.SubmitRecursiveTransfer(
		ctx,
		"mock-source", "/source",
		"mock-dest", "/dest",
		&RecursiveTransferOptions{
			Recursive:         true,
			PreserveTimestamp: true,
			VerifyChecksum:    true,
			EncryptData:       true,
		},
	)
	if err != nil {
		t.Fatalf("SubmitRecursiveTransfer failed: %v", err)
	}
	
	standardMemorySampler.Stop()
	standardPeakMemory := standardMemorySampler.GetPeakMemory()
	
	// Wait a bit to let memory be released
	time.Sleep(1 * time.Second)
	
	// Measure memory for optimized transfer
	optimizedMemorySampler := benchmark.NewMemorySampler(10 * time.Millisecond)
	optimizedMemorySampler.Start()
	
	_, err = client.SubmitMemoryOptimizedTransfer(
		ctx,
		"mock-source", "/source",
		"mock-dest", "/dest",
		&MemoryOptimizedOptions{
			BatchSize:         50,
			MaxConcurrentTasks: 2,
		},
	)
	if err != nil {
		t.Fatalf("SubmitMemoryOptimizedTransfer failed: %v", err)
	}
	
	optimizedMemorySampler.Stop()
	optimizedPeakMemory := optimizedMemorySampler.GetPeakMemory()
	
	// Print comparison
	fmt.Printf("\nMemory Usage Comparison:\n")
	fmt.Printf("Standard Implementation:  %.2f MB\n", standardPeakMemory)
	fmt.Printf("Optimized Implementation: %.2f MB\n", optimizedPeakMemory)
	fmt.Printf("Memory Savings:           %.2f MB (%.1f%%)\n",
		standardPeakMemory-optimizedPeakMemory,
		(1-(optimizedPeakMemory/standardPeakMemory))*100)
	
	// The optimized version should use significantly less memory
	if optimizedPeakMemory >= standardPeakMemory {
		t.Errorf("Expected optimized implementation to use less memory, but it used %.2f MB vs. %.2f MB",
			optimizedPeakMemory, standardPeakMemory)
	}
}
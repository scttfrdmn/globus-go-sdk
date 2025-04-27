// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatchIngestDocuments(t *testing.T) {
	// Create test documents
	docs := make([]SearchDocument, 250)
	for i := 0; i < 250; i++ {
		docs[i] = SearchDocument{
			Subject: fmt.Sprintf("doc%d", i),
			Content: map[string]interface{}{
				"title": fmt.Sprintf("Document %d", i),
				"data":  fmt.Sprintf("Content %d", i),
			},
		}
	}

	// Counter for requests
	var requestCount int32

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment request count
		atomic.AddInt32(&requestCount, 1)

		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/ingest" {
			t.Errorf("Expected path /ingest, got %s", r.URL.Path)
		}

		// Decode request body
		var request IngestRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.IndexID != "test-index-id" {
			t.Errorf("Expected index ID = test-index-id, got %s", request.IndexID)
		}

		// Return mock response
		response := IngestResponse{
			Task: IngestTask{
				TaskID:          fmt.Sprintf("test-task-id-%d", atomic.LoadInt32(&requestCount)),
				ProcessingState: "SUCCESS",
				CreatedAt:       time.Now().Format(time.RFC3339),
				CompletedAt:     time.Now().Format(time.RFC3339),
			},
			Succeeded: len(request.Documents),
			Failed:    0,
			Total:     len(request.Documents),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-token",
		WithBaseURL(server.URL+"/"),
	)

	// Prepare batch options
	options := &BatchIngestOptions{
		BatchSize:     100,
		MaxConcurrent: 2,
		TaskIDPrefix:  "test-batch",
		ProgressCallback: func(processed, total int) {
			t.Logf("Progress: %d/%d documents", processed, total)
		},
	}

	// Test batch ingest
	result, err := client.BatchIngestDocuments(context.Background(), "test-index-id", docs, options)
	if err != nil {
		t.Fatalf("BatchIngestDocuments() error = %v", err)
	}

	// Check result
	if result.TotalDocuments != 250 {
		t.Errorf("Expected total documents = 250, got %d", result.TotalDocuments)
	}
	if result.SuccessDocuments != 250 {
		t.Errorf("Expected success documents = 250, got %d", result.SuccessDocuments)
	}
	if result.FailedDocuments != 0 {
		t.Errorf("Expected failed documents = 0, got %d", result.FailedDocuments)
	}

	// Check number of batches
	if len(result.TaskIDs) != 3 {
		t.Errorf("Expected 3 task IDs, got %d", len(result.TaskIDs))
	}

	// Check request count
	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("Expected 3 requests, got %d", atomic.LoadInt32(&requestCount))
	}
}

func TestBatchDeleteDocuments(t *testing.T) {
	// Create test subjects
	subjects := make([]string, 250)
	for i := 0; i < 250; i++ {
		subjects[i] = fmt.Sprintf("doc%d", i)
	}

	// Counter for requests
	var requestCount int32

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment request count
		atomic.AddInt32(&requestCount, 1)

		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/delete" {
			t.Errorf("Expected path /delete, got %s", r.URL.Path)
		}

		// Decode request body
		var request DeleteDocumentsRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.IndexID != "test-index-id" {
			t.Errorf("Expected index ID = test-index-id, got %s", request.IndexID)
		}

		// Return mock response
		response := DeleteDocumentsResponse{
			Task: IngestTask{
				TaskID:          fmt.Sprintf("test-task-id-%d", atomic.LoadInt32(&requestCount)),
				ProcessingState: "SUCCESS",
				CreatedAt:       time.Now().Format(time.RFC3339),
				CompletedAt:     time.Now().Format(time.RFC3339),
			},
			Succeeded: len(request.Subjects),
			Failed:    0,
			Total:     len(request.Subjects),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-token",
		WithBaseURL(server.URL+"/"),
	)

	// Prepare batch options
	options := &BatchDeleteOptions{
		BatchSize:     100,
		MaxConcurrent: 2,
		TaskIDPrefix:  "test-batch",
		ProgressCallback: func(processed, total int) {
			t.Logf("Progress: %d/%d subjects", processed, total)
		},
	}

	// Test batch delete
	result, err := client.BatchDeleteDocuments(context.Background(), "test-index-id", subjects, options)
	if err != nil {
		t.Fatalf("BatchDeleteDocuments() error = %v", err)
	}

	// Check result
	if result.TotalSubjects != 250 {
		t.Errorf("Expected total subjects = 250, got %d", result.TotalSubjects)
	}
	if result.SuccessSubjects != 250 {
		t.Errorf("Expected success subjects = 250, got %d", result.SuccessSubjects)
	}
	if result.FailedSubjects != 0 {
		t.Errorf("Expected failed subjects = 0, got %d", result.FailedSubjects)
	}

	// Check number of batches
	if len(result.TaskIDs) != 3 {
		t.Errorf("Expected 3 task IDs, got %d", len(result.TaskIDs))
	}

	// Check request count
	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("Expected 3 requests, got %d", atomic.LoadInt32(&requestCount))
	}
}

func TestWaitForTasks(t *testing.T) {
	// Setup test server
	taskStates := map[string]string{
		"task1": "PROCESSING",
		"task2": "SUCCESS",
		"task3": "PROCESSING",
	}

	// Request counter
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment request count
		count := atomic.AddInt32(&requestCount, 1)

		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		taskID := r.URL.Path[6:] // Strip /task/ prefix
		if _, ok := taskStates[taskID]; !ok {
			t.Errorf("Unexpected task ID: %s", taskID)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Change state after a few requests
		if count > 2 {
			taskStates["task1"] = "SUCCESS"
			taskStates["task3"] = "SUCCESS"
		}

		// Return mock response
		response := TaskStatusResponse{
			TaskID:         taskID,
			State:          taskStates[taskID],
			CreatedAt:      time.Now().Format(time.RFC3339),
			CompletedAt:    "",
			TotalDocuments: 100,
		}

		if taskStates[taskID] == "SUCCESS" {
			response.CompletedAt = time.Now().Format(time.RFC3339)
			response.SuccessDocuments = 100
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-token",
		WithBaseURL(server.URL+"/"),
	)

	// Test wait for tasks
	taskIDs := []string{"task1", "task2", "task3"}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := client.WaitForTasks(ctx, taskIDs, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForTasks() error = %v", err)
	}

	// Check results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.TaskID != taskIDs[i] {
			t.Errorf("Expected task ID = %s, got %s", taskIDs[i], result.TaskID)
		}
		if result.State != "SUCCESS" {
			t.Errorf("Expected state = SUCCESS, got %s", result.State)
		}
	}

	// Check request count
	if atomic.LoadInt32(&requestCount) <= 3 {
		t.Errorf("Expected more than 3 requests, got %d", atomic.LoadInt32(&requestCount))
	}
}

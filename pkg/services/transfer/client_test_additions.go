// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestGetTask(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet || r.URL.Path != "/task/task-123456" {
			t.Errorf("Expected GET /task/task-123456, got %s %s", r.Method, r.URL.Path)
		}

		// Return mock response
		response := Task{
			TaskID:                "task-123456",
			Type:                  "TRANSFER",
			Status:                "ACTIVE",
			Label:                 "Test Transfer",
			Owner:                 "test-user@example.com",
			SourceEndpointID:      "source-endpoint",
			DestinationEndpointID: "dest-endpoint",
			BytesTransferred:      1024,
			FilesTransferred:      2,
			FilesSkipped:          0,
			SubtasksPending:       1,
			SubtasksSucceeded:     3,
			SubtasksFailed:        0,
			SubtasksCanceled:      0,
			SubmissionTime:        time.Now().Add(-time.Hour).Format(time.RFC3339),
			CompletionTime:        "",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test getting a task
	result, err := client.GetTask(context.Background(), "task-123456")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	// Check response
	if result.TaskID != "task-123456" {
		t.Errorf("TaskID = %s, want 'task-123456'", result.TaskID)
	}

	if result.Status != "ACTIVE" {
		t.Errorf("Status = %s, want 'ACTIVE'", result.Status)
	}

	if result.Label != "Test Transfer" {
		t.Errorf("Label = %s, want 'Test Transfer'", result.Label)
	}

	if result.BytesTransferred != 1024 {
		t.Errorf("BytesTransferred = %d, want 1024", result.BytesTransferred)
	}

	if result.FilesTransferred != 2 {
		t.Errorf("FilesTransferred = %d, want 2", result.FilesTransferred)
	}

	// Test with empty task ID
	_, err = client.GetTask(context.Background(), "")
	if err == nil {
		t.Error("GetTask() with empty task ID should return error")
	}
}

func TestCancelTask(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost || r.URL.Path != "/task/task-123456/cancel" {
			t.Errorf("Expected POST /task/task-123456/cancel, got %s %s", r.Method, r.URL.Path)
		}

		// Return mock response
		response := TaskResponse{
			Code:    "Canceled",
			Message: "The task has been cancelled",
			TaskID:  "task-123456",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test canceling a task
	result, err := client.CancelTask(context.Background(), "task-123456")
	if err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}

	// Check response
	if result.Code != "Canceled" {
		t.Errorf("Code = %s, want 'Canceled'", result.Code)
	}

	if result.Message != "The task has been cancelled" {
		t.Errorf("Message = %s, want 'The task has been cancelled'", result.Message)
	}

	if result.TaskID != "task-123456" {
		t.Errorf("TaskID = %s, want 'task-123456'", result.TaskID)
	}

	// Test with empty task ID
	_, err = client.CancelTask(context.Background(), "")
	if err == nil {
		t.Error("CancelTask() with empty task ID should return error")
	}
}

func TestListTasks(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet || r.URL.Path != "/task_list" {
			t.Errorf("Expected GET /task_list, got %s %s", r.Method, r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("filter") != "type:TRANSFER/status:ACTIVE" {
			t.Errorf("Expected filter=type:TRANSFER/status:ACTIVE, got %s", r.URL.Query().Get("filter"))
		}

		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
		}

		// Return mock response
		response := TaskList{
			Data: []Task{
				{
					TaskID: "task-123456",
					Type:   "TRANSFER",
					Status: "ACTIVE",
					Label:  "Test Transfer 1",
				},
				{
					TaskID: "task-123457",
					Type:   "TRANSFER",
					Status: "ACTIVE",
					Label:  "Test Transfer 2",
				},
			},
			NextPageToken: "next-page-token",
			HasNextPage:   true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test listing tasks
	options := &ListTasksOptions{
		Filter: "type:TRANSFER/status:ACTIVE",
		Limit:  10,
	}

	result, err := client.ListTasks(context.Background(), options)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	// Check response
	if len(result.Data) != 2 {
		t.Errorf("ListTasks() returned %d tasks, want 2", len(result.Data))
	}

	if result.Data[0].TaskID != "task-123456" {
		t.Errorf("First task ID = %s, want 'task-123456'", result.Data[0].TaskID)
	}

	if result.Data[1].Label != "Test Transfer 2" {
		t.Errorf("Second task Label = %s, want 'Test Transfer 2'", result.Data[1].Label)
	}

	if !result.HasNextPage {
		t.Error("HasNextPage = false, want true")
	}

	if result.NextPageToken != "next-page-token" {
		t.Errorf("NextPageToken = %s, want 'next-page-token'", result.NextPageToken)
	}
}

func TestWaitTask(t *testing.T) {
	callCount := 0

	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet || r.URL.Path != "/task/task-123456" {
			t.Errorf("Expected GET /task/task-123456, got %s %s", r.Method, r.URL.Path)
		}

		// Return different status based on call count
		status := "ACTIVE"
		if callCount >= 2 {
			status = "SUCCEEDED"
		}
		callCount++

		// Return mock response
		response := Task{
			TaskID: "task-123456",
			Type:   "TRANSFER",
			Status: status,
			Label:  "Test Transfer",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Set a shorter poll interval for the test
	pollInterval := 50 * time.Millisecond

	// Test waiting for a task
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	task, err := client.WaitTask(ctx, "task-123456", pollInterval)
	if err != nil {
		t.Fatalf("WaitTask() error = %v", err)
	}

	// Check final task status
	if task.Status != "SUCCEEDED" {
		t.Errorf("Final task status = %s, want 'SUCCEEDED'", task.Status)
	}

	// Test with empty task ID
	_, err = client.WaitTask(ctx, "", pollInterval)
	if err == nil {
		t.Error("WaitTask() with empty task ID should return error")
	}

	// Test with context timeout
	// Create a server that always returns ACTIVE
	handler = func(w http.ResponseWriter, r *http.Request) {
		response := Task{
			TaskID: "task-123456",
			Type:   "TRANSFER",
			Status: "ACTIVE",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err = client.WaitTask(ctx, "task-123456", pollInterval)
	if err == nil {
		t.Error("WaitTask() with context timeout should return error")
	}
}

func TestErrorHandling(t *testing.T) {
	// Setup test server that returns an error
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		errorResp := map[string]interface{}{
			"code":       "ClientError.NotFound",
			"message":    "The requested endpoint was not found",
			"request_id": "abc123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(errorResp)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test getting a non-existent endpoint
	_, err := client.GetEndpoint(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetEndpoint() should return an error for non-existent endpoint")
	}

	// Check that the error contains the expected information
	errorStr := err.Error()
	if errorStr == "" {
		t.Error("Error message should not be empty")
	}

	// Check that the error contains either the status code or error message
	if !contains(errorStr, "404") && !contains(errorStr, "NotFound") {
		t.Errorf("Error should mention NotFound or 404 status, got: %s", errorStr)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

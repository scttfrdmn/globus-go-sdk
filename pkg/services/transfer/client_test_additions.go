// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
)

// setupMockServerForTests creates a test server and client for specific tests
func setupMockServerForTests(handler http.HandlerFunc) (*httptest.Server, *Client) {
	// Create a test server
	server := httptest.NewServer(handler)

	// Create a client that talks to the test server
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	client, err := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		panic(err)
	}

	return server, client
}

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
			SourceEndpointID:      "source-endpoint",
			DestinationEndpointID: "dest-endpoint",
			BytesTransferred:      1024,
			FilesTransferred:      2,
			FilesSkipped:          0,
			SubtasksPending:       1,
			SubtasksSucceeded:     3,
			SubtasksFailed:        0,
			SubtasksCanceled:      0,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServerForTests(handler)
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

	server, client := setupMockServerForTests(handler)
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
			NextMarker: "next-page-token",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServerForTests(handler)
	defer server.Close()

	// Test listing tasks
	options := &ListTasksOptions{
		TaskType: "TRANSFER",
		Status:   "ACTIVE",
		Limit:    10,
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

	if result.NextMarker == "" {
		t.Error("NextMarker is empty, expected a value")
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

	server, client := setupMockServerForTests(handler)
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

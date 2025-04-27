// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Test mock server
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)

	// Create a client that uses the test server
	client := NewClient("test-token",
		core.WithBaseURL(server.URL+"/"),
	)

	return server, client
}

func TestListEndpoints(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoints" {
			t.Errorf("Expected path /endpoints, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if perPage := queryParams.Get("per_page"); perPage != "10" {
			t.Errorf("Expected per_page=10, got %s", perPage)
		}

		// Return mock response
		endpointTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := ComputeEndpointList{
			EndpointIDs: []string{"test-endpoint-id"},
			Endpoints: []ComputeEndpoint{
				{
					ID:           "test-endpoint-id",
					UUID:         "test-uuid",
					Status:       "online",
					Name:         "Test Endpoint",
					Description:  "A test endpoint",
					Owner:        "test-user",
					CreatedAt:    endpointTime,
					LastModified: endpointTime,
					Connected:    true,
					Type:         "container",
					Public:       false,
				},
			},
			Total:       1,
			Offset:      0,
			Limit:       10,
			HasMorePage: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list endpoints
	options := &ListEndpointsOptions{
		PerPage: 10,
	}

	endpointList, err := client.ListEndpoints(context.Background(), options)
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}

	// Check response
	if len(endpointList.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(endpointList.Endpoints))
	}

	endpoint := endpointList.Endpoints[0]
	if endpoint.ID != "test-endpoint-id" {
		t.Errorf("Expected endpoint ID = test-endpoint-id, got %s", endpoint.ID)
	}
	if endpoint.Name != "Test Endpoint" {
		t.Errorf("Expected endpoint name = Test Endpoint, got %s", endpoint.Name)
	}
}

func TestGetEndpoint(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/endpoints/test-endpoint-id" {
			t.Errorf("Expected path /endpoints/test-endpoint-id, got %s", r.URL.Path)
		}

		// Return mock response
		endpointTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := ComputeEndpoint{
			ID:           "test-endpoint-id",
			UUID:         "test-uuid",
			Status:       "online",
			Name:         "Test Endpoint",
			Description:  "A test endpoint",
			Owner:        "test-user",
			CreatedAt:    endpointTime,
			LastModified: endpointTime,
			Connected:    true,
			Type:         "container",
			Public:       false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get endpoint
	endpoint, err := client.GetEndpoint(context.Background(), "test-endpoint-id")
	if err != nil {
		t.Fatalf("GetEndpoint() error = %v", err)
	}

	// Check response
	if endpoint.ID != "test-endpoint-id" {
		t.Errorf("Expected endpoint ID = test-endpoint-id, got %s", endpoint.ID)
	}
	if endpoint.Name != "Test Endpoint" {
		t.Errorf("Expected endpoint name = Test Endpoint, got %s", endpoint.Name)
	}

	// Test empty endpoint ID
	_, err = client.GetEndpoint(context.Background(), "")
	if err == nil {
		t.Error("GetEndpoint() with empty ID should return error")
	}
}

func TestRegisterFunction(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/functions" {
			t.Errorf("Expected path /functions, got %s", r.URL.Path)
		}

		// Decode request body
		var request FunctionRegisterRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.Function != "def hello(name='World'): return f'Hello, {name}!'" {
			t.Errorf("Expected function code, got %s", request.Function)
		}

		// Return mock response
		funcTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := FunctionResponse{
			ID:          "test-function-id",
			Function:    request.Function,
			Name:        request.Name,
			Description: request.Description,
			Status:      "ACTIVE",
			Owner:       "test-user",
			Public:      request.Public,
			CreatedAt:   funcTime,
			ModifiedAt:  funcTime,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test register function
	registerRequest := &FunctionRegisterRequest{
		Function:    "def hello(name='World'): return f'Hello, {name}!'",
		Name:        "hello",
		Description: "A simple greeting function",
		Public:      true,
	}

	function, err := client.RegisterFunction(context.Background(), registerRequest)
	if err != nil {
		t.Fatalf("RegisterFunction() error = %v", err)
	}

	// Check response
	if function.ID != "test-function-id" {
		t.Errorf("Expected function ID = test-function-id, got %s", function.ID)
	}
	if function.Name != "hello" {
		t.Errorf("Expected function name = hello, got %s", function.Name)
	}

	// Test nil request
	_, err = client.RegisterFunction(context.Background(), nil)
	if err == nil {
		t.Error("RegisterFunction() with nil request should return error")
	}

	// Test empty function code
	_, err = client.RegisterFunction(context.Background(), &FunctionRegisterRequest{
		Name: "empty",
	})
	if err == nil {
		t.Error("RegisterFunction() with empty function code should return error")
	}
}

func TestGetFunction(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/functions/test-function-id" {
			t.Errorf("Expected path /functions/test-function-id, got %s", r.URL.Path)
		}

		// Return mock response
		funcTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := FunctionResponse{
			ID:          "test-function-id",
			Function:    "def hello(name='World'): return f'Hello, {name}!'",
			Name:        "hello",
			Description: "A simple greeting function",
			Status:      "ACTIVE",
			Owner:       "test-user",
			Public:      true,
			CreatedAt:   funcTime,
			ModifiedAt:  funcTime,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get function
	function, err := client.GetFunction(context.Background(), "test-function-id")
	if err != nil {
		t.Fatalf("GetFunction() error = %v", err)
	}

	// Check response
	if function.ID != "test-function-id" {
		t.Errorf("Expected function ID = test-function-id, got %s", function.ID)
	}
	if function.Name != "hello" {
		t.Errorf("Expected function name = hello, got %s", function.Name)
	}

	// Test empty function ID
	_, err = client.GetFunction(context.Background(), "")
	if err == nil {
		t.Error("GetFunction() with empty ID should return error")
	}
}

func TestListFunctions(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/functions" {
			t.Errorf("Expected path /functions, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if perPage := queryParams.Get("per_page"); perPage != "10" {
			t.Errorf("Expected per_page=10, got %s", perPage)
		}

		// Return mock response
		funcTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := FunctionList{
			Functions: []FunctionResponse{
				{
					ID:          "test-function-id",
					Function:    "def hello(name='World'): return f'Hello, {name}!'",
					Name:        "hello",
					Description: "A simple greeting function",
					Status:      "ACTIVE",
					Owner:       "test-user",
					Public:      true,
					CreatedAt:   funcTime,
					ModifiedAt:  funcTime,
				},
			},
			Total:       1,
			HasNextPage: false,
			Offset:      0,
			Limit:       10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list functions
	options := &ListFunctionsOptions{
		PerPage: 10,
	}

	functionList, err := client.ListFunctions(context.Background(), options)
	if err != nil {
		t.Fatalf("ListFunctions() error = %v", err)
	}

	// Check response
	if len(functionList.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(functionList.Functions))
	}

	function := functionList.Functions[0]
	if function.ID != "test-function-id" {
		t.Errorf("Expected function ID = test-function-id, got %s", function.ID)
	}
	if function.Name != "hello" {
		t.Errorf("Expected function name = hello, got %s", function.Name)
	}
}

func TestUpdateFunction(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/functions/test-function-id" {
			t.Errorf("Expected path /functions/test-function-id, got %s", r.URL.Path)
		}

		// Decode request body
		var request FunctionUpdateRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.Description != "Updated description" {
			t.Errorf("Expected updated description, got %s", request.Description)
		}

		// Return mock response
		funcTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := FunctionResponse{
			ID:          "test-function-id",
			Function:    "def hello(name='World'): return f'Hello, {name}!'",
			Name:        "hello",
			Description: request.Description,
			Status:      "ACTIVE",
			Owner:       "test-user",
			Public:      true,
			CreatedAt:   funcTime,
			ModifiedAt:  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test update function
	updateRequest := &FunctionUpdateRequest{
		Description: "Updated description",
	}

	function, err := client.UpdateFunction(context.Background(), "test-function-id", updateRequest)
	if err != nil {
		t.Fatalf("UpdateFunction() error = %v", err)
	}

	// Check response
	if function.ID != "test-function-id" {
		t.Errorf("Expected function ID = test-function-id, got %s", function.ID)
	}
	if function.Description != "Updated description" {
		t.Errorf("Expected description = Updated description, got %s", function.Description)
	}

	// Test empty function ID
	_, err = client.UpdateFunction(context.Background(), "", updateRequest)
	if err == nil {
		t.Error("UpdateFunction() with empty ID should return error")
	}

	// Test nil request
	_, err = client.UpdateFunction(context.Background(), "test-function-id", nil)
	if err == nil {
		t.Error("UpdateFunction() with nil request should return error")
	}
}

func TestDeleteFunction(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/functions/test-function-id" {
			t.Errorf("Expected path /functions/test-function-id, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusNoContent)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test delete function
	err := client.DeleteFunction(context.Background(), "test-function-id")
	if err != nil {
		t.Fatalf("DeleteFunction() error = %v", err)
	}

	// Test empty function ID
	err = client.DeleteFunction(context.Background(), "")
	if err == nil {
		t.Error("DeleteFunction() with empty ID should return error")
	}
}

func TestRunFunction(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/run" {
			t.Errorf("Expected path /run, got %s", r.URL.Path)
		}

		// Decode request body
		var request TaskRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.FunctionID != "test-function-id" {
			t.Errorf("Expected function ID = test-function-id, got %s", request.FunctionID)
		}
		if request.EndpointID != "test-endpoint-id" {
			t.Errorf("Expected endpoint ID = test-endpoint-id, got %s", request.EndpointID)
		}

		// Return mock response
		response := TaskResponse{
			TaskID:  "test-task-id",
			Status:  "PENDING",
			Message: "Task submitted",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test run function
	taskRequest := &TaskRequest{
		FunctionID: "test-function-id",
		EndpointID: "test-endpoint-id",
		Args:       []interface{}{"test-arg"},
	}

	taskResponse, err := client.RunFunction(context.Background(), taskRequest)
	if err != nil {
		t.Fatalf("RunFunction() error = %v", err)
	}

	// Check response
	if taskResponse.TaskID != "test-task-id" {
		t.Errorf("Expected task ID = test-task-id, got %s", taskResponse.TaskID)
	}
	if taskResponse.Status != "PENDING" {
		t.Errorf("Expected status = PENDING, got %s", taskResponse.Status)
	}

	// Test nil request
	_, err = client.RunFunction(context.Background(), nil)
	if err == nil {
		t.Error("RunFunction() with nil request should return error")
	}

	// Test empty function ID
	_, err = client.RunFunction(context.Background(), &TaskRequest{
		EndpointID: "test-endpoint-id",
	})
	if err == nil {
		t.Error("RunFunction() with empty function ID should return error")
	}

	// Test empty endpoint ID
	_, err = client.RunFunction(context.Background(), &TaskRequest{
		FunctionID: "test-function-id",
	})
	if err == nil {
		t.Error("RunFunction() with empty endpoint ID should return error")
	}
}

func TestRunBatch(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/batch" {
			t.Errorf("Expected path /batch, got %s", r.URL.Path)
		}

		// Decode request body
		var request BatchTaskRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if len(request.Tasks) != 2 {
			t.Errorf("Expected 2 tasks, got %d", len(request.Tasks))
		}

		// Return mock response
		response := BatchTaskResponse{
			TaskIDs: []string{"test-task-id-1", "test-task-id-2"},
			Status:  "PENDING",
			Message: "Batch tasks submitted",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test run batch
	batchRequest := &BatchTaskRequest{
		Tasks: []TaskRequest{
			{
				FunctionID: "test-function-id-1",
				EndpointID: "test-endpoint-id",
				Args:       []interface{}{"test-arg-1"},
			},
			{
				FunctionID: "test-function-id-2",
				EndpointID: "test-endpoint-id",
				Args:       []interface{}{"test-arg-2"},
			},
		},
	}

	batchResponse, err := client.RunBatch(context.Background(), batchRequest)
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}

	// Check response
	if len(batchResponse.TaskIDs) != 2 {
		t.Errorf("Expected 2 task IDs, got %d", len(batchResponse.TaskIDs))
	}
	if batchResponse.Status != "PENDING" {
		t.Errorf("Expected status = PENDING, got %s", batchResponse.Status)
	}

	// Test nil request
	_, err = client.RunBatch(context.Background(), nil)
	if err == nil {
		t.Error("RunBatch() with nil request should return error")
	}

	// Test empty tasks
	_, err = client.RunBatch(context.Background(), &BatchTaskRequest{})
	if err == nil {
		t.Error("RunBatch() with empty tasks should return error")
	}
}

func TestGetTaskStatus(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/status/test-task-id" {
			t.Errorf("Expected path /status/test-task-id, got %s", r.URL.Path)
		}

		// Return mock response
		completedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := TaskStatus{
			TaskID:      "test-task-id",
			Status:      "SUCCESS",
			CompletedAt: completedTime,
			Result:      "Hello, test-arg!",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get task status
	status, err := client.GetTaskStatus(context.Background(), "test-task-id")
	if err != nil {
		t.Fatalf("GetTaskStatus() error = %v", err)
	}

	// Check response
	if status.TaskID != "test-task-id" {
		t.Errorf("Expected task ID = test-task-id, got %s", status.TaskID)
	}
	if status.Status != "SUCCESS" {
		t.Errorf("Expected status = SUCCESS, got %s", status.Status)
	}

	// Test empty task ID
	_, err = client.GetTaskStatus(context.Background(), "")
	if err == nil {
		t.Error("GetTaskStatus() with empty ID should return error")
	}
}

func TestGetBatchStatus(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/batch_status" {
			t.Errorf("Expected path /batch_status, got %s", r.URL.Path)
		}

		// Decode request body
		var request map[string][]string
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		taskIDs, ok := request["task_ids"]
		if !ok {
			t.Errorf("Expected task_ids in request")
		}
		if len(taskIDs) != 2 {
			t.Errorf("Expected 2 task IDs, got %d", len(taskIDs))
		}

		// Return mock response
		completedTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := BatchTaskStatus{
			Tasks: map[string]TaskStatus{
				"test-task-id-1": {
					TaskID:      "test-task-id-1",
					Status:      "SUCCESS",
					CompletedAt: completedTime,
					Result:      "Hello, test-arg-1!",
				},
				"test-task-id-2": {
					TaskID:      "test-task-id-2",
					Status:      "SUCCESS",
					CompletedAt: completedTime,
					Result:      "Hello, test-arg-2!",
				},
			},
			Completed: []string{"test-task-id-1", "test-task-id-2"},
			Message:   "All tasks completed",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get batch status
	status, err := client.GetBatchStatus(context.Background(), []string{"test-task-id-1", "test-task-id-2"})
	if err != nil {
		t.Fatalf("GetBatchStatus() error = %v", err)
	}

	// Check response
	if len(status.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(status.Tasks))
	}
	if len(status.Completed) != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", len(status.Completed))
	}

	// Test empty task IDs
	_, err = client.GetBatchStatus(context.Background(), []string{})
	if err == nil {
		t.Error("GetBatchStatus() with empty task IDs should return error")
	}
}

func TestListTasks(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/tasks" {
			t.Errorf("Expected path /tasks, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if perPage := queryParams.Get("per_page"); perPage != "10" {
			t.Errorf("Expected per_page=10, got %s", perPage)
		}

		// Return mock response
		response := TaskList{
			Tasks:       []string{"test-task-id-1", "test-task-id-2"},
			Total:       2,
			HasNextPage: false,
			Offset:      0,
			Limit:       10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list tasks
	options := &TaskListOptions{
		PerPage: 10,
	}

	taskList, err := client.ListTasks(context.Background(), options)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	// Check response
	if len(taskList.Tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(taskList.Tasks))
	}
	if taskList.Total != 2 {
		t.Errorf("Expected total = 2, got %d", taskList.Total)
	}
}

func TestCancelTask(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/tasks/test-task-id/cancel" {
			t.Errorf("Expected path /tasks/test-task-id/cancel, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test cancel task
	err := client.CancelTask(context.Background(), "test-task-id")
	if err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}

	// Test empty task ID
	err = client.CancelTask(context.Background(), "")
	if err == nil {
		t.Error("CancelTask() with empty ID should return error")
	}
}

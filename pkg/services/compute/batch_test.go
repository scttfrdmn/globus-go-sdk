// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateWorkflow(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workflows", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req WorkflowCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-workflow", req.Name)
		assert.Equal(t, 2, len(req.Tasks))

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := WorkflowResponse{
			ID:   "workflow123",
			Name: req.Name,
			Tasks: []WorkflowTask{
				{
					ID:         "task1",
					FunctionID: "func1",
					EndpointID: "endpoint1",
				},
				{
					ID:         "task2",
					FunctionID: "func2",
					EndpointID: "endpoint1",
				},
			},
			Dependencies: map[string][]string{
				"task2": {"task1"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &WorkflowCreateRequest{
		Name:        "test-workflow",
		Description: "Test workflow for unit tests",
		Tasks: []WorkflowTask{
			{
				ID:         "task1",
				FunctionID: "func1",
				EndpointID: "endpoint1",
			},
			{
				ID:         "task2",
				FunctionID: "func2",
				EndpointID: "endpoint1",
			},
		},
		Dependencies: map[string][]string{
			"task2": {"task1"},
		},
		ErrorHandling: "continue",
	}

	resp, err := client.CreateWorkflow(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "workflow123", resp.ID)
	assert.Equal(t, "test-workflow", resp.Name)
	assert.Len(t, resp.Tasks, 2)
	assert.Contains(t, resp.Dependencies, "task2")
	assert.Contains(t, resp.Dependencies["task2"], "task1")
}

func TestRunWorkflow(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workflows/workflow123/run", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req WorkflowRunRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, 5, req.Priority)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := WorkflowRunResponse{
			RunID:      "run123",
			WorkflowID: "workflow123",
			Status:     "ACTIVE",
			Message:    "Workflow started",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &WorkflowRunRequest{
		Priority:    5,
		Description: "Test run",
		GlobalArgs: map[string]interface{}{
			"data_url": "https://example.com/data.json",
		},
	}

	resp, err := client.RunWorkflow(ctx, "workflow123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "run123", resp.RunID)
	assert.Equal(t, "workflow123", resp.WorkflowID)
	assert.Equal(t, "ACTIVE", resp.Status)
}

func TestGetWorkflowStatus(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workflows/runs/run123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := WorkflowStatusResponse{
			RunID:      "run123",
			WorkflowID: "workflow123",
			Status:     "RUNNING",
			TaskStatus: map[string]TaskStatusInfo{
				"task1": {
					Status:      "COMPLETED",
					TaskID:      "task1-run-id",
					StartedAt:   time.Now().Add(-5 * time.Minute),
					CompletedAt: time.Now().Add(-2 * time.Minute),
					Result:      "Task 1 result",
				},
				"task2": {
					Status:    "RUNNING",
					TaskID:    "task2-run-id",
					StartedAt: time.Now().Add(-1 * time.Minute),
				},
			},
			Progress: WorkflowProgressInfo{
				TotalTasks:  2,
				Completed:   1,
				Running:     1,
				PercentDone: 50.0,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.GetWorkflowStatus(ctx, "run123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "run123", resp.RunID)
	assert.Equal(t, "workflow123", resp.WorkflowID)
	assert.Equal(t, "RUNNING", resp.Status)
	assert.Len(t, resp.TaskStatus, 2)
	assert.Equal(t, "COMPLETED", resp.TaskStatus["task1"].Status)
	assert.Equal(t, "RUNNING", resp.TaskStatus["task2"].Status)
	assert.Equal(t, 50.0, resp.Progress.PercentDone)
}

func TestRunDependencyGraph(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependency_graph/run", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req DependencyGraphRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Len(t, req.Nodes, 3)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyGraphResponse{
			RunID:   "graph123",
			Status:  "ACTIVE",
			Message: "Dependency graph execution started",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &DependencyGraphRequest{
		Nodes: map[string]DependencyGraphNode{
			"node1": {
				Task: TaskRequest{
					FunctionID: "func1",
					EndpointID: "endpoint1",
				},
			},
			"node2": {
				Task: TaskRequest{
					FunctionID: "func2",
					EndpointID: "endpoint1",
				},
				Dependencies: []string{"node1"},
			},
			"node3": {
				Task: TaskRequest{
					FunctionID: "func3",
					EndpointID: "endpoint1",
				},
				Dependencies: []string{"node1", "node2"},
				RetryPolicy: &RetryPolicy{
					MaxRetries: 3,
				},
			},
		},
		ErrorPolicy: "fail-fast",
	}

	resp, err := client.RunDependencyGraph(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "graph123", resp.RunID)
	assert.Equal(t, "ACTIVE", resp.Status)
}

func TestGetDependencyGraphStatus(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/dependency_graph/runs/graph123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := DependencyGraphStatusResponse{
			RunID:  "graph123",
			Status: "RUNNING",
			NodeStatus: map[string]TaskStatusInfo{
				"node1": {
					Status:      "COMPLETED",
					TaskID:      "node1-run-id",
					StartedAt:   time.Now().Add(-10 * time.Minute),
					CompletedAt: time.Now().Add(-8 * time.Minute),
					Result:      "Node 1 result",
				},
				"node2": {
					Status:      "COMPLETED",
					TaskID:      "node2-run-id",
					StartedAt:   time.Now().Add(-7 * time.Minute),
					CompletedAt: time.Now().Add(-4 * time.Minute),
					Result:      "Node 2 result",
				},
				"node3": {
					Status:    "RUNNING",
					TaskID:    "node3-run-id",
					StartedAt: time.Now().Add(-3 * time.Minute),
				},
			},
			Progress: DependencyGraphProgressInfo{
				TotalNodes:  3,
				Completed:   2,
				Running:     1,
				PercentDone: 66.67,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	resp, err := client.GetDependencyGraphStatus(ctx, "graph123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "graph123", resp.RunID)
	assert.Equal(t, "RUNNING", resp.Status)
	assert.Len(t, resp.NodeStatus, 3)
	assert.Equal(t, "COMPLETED", resp.NodeStatus["node1"].Status)
	assert.Equal(t, "COMPLETED", resp.NodeStatus["node2"].Status)
	assert.Equal(t, "RUNNING", resp.NodeStatus["node3"].Status)
	assert.InDelta(t, 66.67, resp.Progress.PercentDone, 0.01)
}

func TestWaitForWorkflowCompletion(t *testing.T) {
	// Setup test server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workflows/runs/run123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		callCount++
		status := "RUNNING"
		if callCount >= 3 {
			status = "COMPLETED"
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := WorkflowStatusResponse{
			RunID:      "run123",
			WorkflowID: "workflow123",
			Status:     status,
			Progress: WorkflowProgressInfo{
				TotalTasks:  2,
				Completed:   callCount,
				PercentDone: float64(callCount) * 50.0,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method with short poll interval for test
	ctx := context.Background()
	resp, err := client.WaitForWorkflowCompletion(ctx, "run123", 5*time.Second, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "COMPLETED", resp.Status)

	// The struct fields are typed as int/float64 but JSON unmarshaling can produce different numeric types
	// So we need to make the assertion more flexible
	completedVal := resp.Progress.Completed
	percentDone := resp.Progress.PercentDone
	assert.Equal(t, 3, completedVal)
	assert.Equal(t, 150.0, percentDone)

	assert.Equal(t, 3, callCount) // Check that we polled status 3 times
}

func TestCreateTaskGroup(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/task_groups", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req TaskGroupCreateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-group", req.Name)
		assert.Equal(t, 3, len(req.Tasks))
		assert.Equal(t, 5, req.Concurrency)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := TaskGroupResponse{
			ID:          "group123",
			Name:        req.Name,
			Concurrency: req.Concurrency,
			Tasks:       req.Tasks,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &TaskGroupCreateRequest{
		Name:        "test-group",
		Description: "Test task group for unit tests",
		Tasks: []TaskRequest{
			{
				FunctionID: "func1",
				EndpointID: "endpoint1",
				Args:       []interface{}{"arg1", "arg2"},
			},
			{
				FunctionID: "func1",
				EndpointID: "endpoint1",
				Args:       []interface{}{"arg3", "arg4"},
			},
			{
				FunctionID: "func1",
				EndpointID: "endpoint1",
				Args:       []interface{}{"arg5", "arg6"},
			},
		},
		Concurrency: 5,
		RetryPolicy: &RetryPolicy{
			MaxRetries: 2,
		},
	}

	resp, err := client.CreateTaskGroup(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "group123", resp.ID)
	assert.Equal(t, "test-group", resp.Name)
	assert.Equal(t, 5, resp.Concurrency)
	assert.Len(t, resp.Tasks, 3)
}

func TestRunTaskGroup(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/task_groups/group123/run", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var req TaskGroupRunRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, 3, req.Priority)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := TaskGroupRunResponse{
			RunID:       "tgr123",
			TaskGroupID: "group123",
			Status:      "ACTIVE",
			Message:     "Task group started",
			TaskIDs:     []string{"task1", "task2", "task3"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(
		WithBaseURL(server.URL + "/"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test the method
	ctx := context.Background()
	req := &TaskGroupRunRequest{
		Priority:    3,
		Description: "Test task group run",
	}

	resp, err := client.RunTaskGroup(ctx, "group123", req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "tgr123", resp.RunID)
	assert.Equal(t, "group123", resp.TaskGroupID)
	assert.Equal(t, "ACTIVE", resp.Status)
	assert.Len(t, resp.TaskIDs, 3)
}

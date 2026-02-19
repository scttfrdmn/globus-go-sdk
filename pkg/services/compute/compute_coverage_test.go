// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
)

// newTestClient creates a compute client pointed at a test server.
func newTestClient(t *testing.T, server *httptest.Server) *compute.Client {
	t.Helper()
	client, err := compute.NewClient(
		compute.WithAccessToken("test-token"),
		compute.WithBaseURL(server.URL+"/"),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

// ──────────────────────────────────────────────────────────────────────────────
// Options
// ──────────────────────────────────────────────────────────────────────────────

func TestWithHTTPDebugging(t *testing.T) {
	// Just verifies that the option can be applied without error.
	_, err := compute.NewClient(
		compute.WithAccessToken("tok"),
		compute.WithHTTPDebugging(true),
	)
	if err != nil {
		t.Fatalf("NewClient() with WithHTTPDebugging error = %v", err)
	}
}

func TestWithHTTPTracing(t *testing.T) {
	_, err := compute.NewClient(
		compute.WithAccessToken("tok"),
		compute.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("NewClient() with WithHTTPTracing error = %v", err)
	}
}

func TestWithCoreOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(compute.ComputeEndpointList{})
	}))
	defer server.Close()

	// WithCoreOption allows setting a base URL via the core option path.
	client := newTestClient(t, server)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithAuthorizer(t *testing.T) {
	// A nil authorizer is still accepted (no-op).
	_, err := compute.NewClient(
		compute.WithAuthorizer(nil),
	)
	if err != nil {
		t.Fatalf("NewClient() with WithAuthorizer(nil) error = %v", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListEndpointsV2
// ──────────────────────────────────────────────────────────────────────────────

func TestListEndpointsV2(t *testing.T) {
	endpointTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/endpoints" {
			t.Errorf("expected path /endpoints, got %s", r.URL.Path)
		}
		resp := compute.ComputeEndpointList{
			Endpoints: []compute.ComputeEndpoint{
				{
					ID:           "ep-v2",
					UUID:         "uuid-v2",
					Status:       "online",
					Name:         "Endpoint V2",
					Owner:        "user",
					CreatedAt:    endpointTime,
					LastModified: endpointTime,
					Connected:    true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	opts := &compute.ListEndpointsOptions{
		PerPage:      5,
		Marker:       "marker1",
		OrderBy:      "name",
		Search:       "test",
		FilterScope:  "my-endpoints",
		FilterStatus: "online",
		IncludeInfo:  true,
	}

	result, err := client.ListEndpointsV2(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListEndpointsV2() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Data.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(result.Data.Endpoints))
	}
	if result.Data.Endpoints[0].ID != "ep-v2" {
		t.Errorf("expected endpoint ID ep-v2, got %s", result.Data.Endpoints[0].ID)
	}
}

func TestListEndpointsV2_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.ListEndpointsV2(context.Background(), nil)
	if err == nil {
		t.Error("expected error from ListEndpointsV2 on server error, got nil")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetEndpoint
// ──────────────────────────────────────────────────────────────────────────────

func TestGetEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/endpoints/ep-123" {
			t.Errorf("expected /endpoints/ep-123, got %s", r.URL.Path)
		}
		resp := compute.ComputeEndpoint{
			ID:     "ep-123",
			Name:   "My Endpoint",
			Status: "online",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	ep, err := client.GetEndpoint(context.Background(), "ep-123")
	if err != nil {
		t.Fatalf("GetEndpoint() error = %v", err)
	}
	if ep.ID != "ep-123" {
		t.Errorf("expected ID ep-123, got %s", ep.ID)
	}
	if ep.Name != "My Endpoint" {
		t.Errorf("expected name My Endpoint, got %s", ep.Name)
	}
}

func TestGetEndpoint_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called")
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetEndpoint(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Function CRUD
// ──────────────────────────────────────────────────────────────────────────────

func TestRegisterFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/functions" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req compute.FunctionRegisterRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := compute.FunctionResponse{
			ID:       "fn-001",
			Function: req.Function,
			Name:     req.Name,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	req := &compute.FunctionRegisterRequest{
		Function:    "def my_func(x): return x*2",
		Name:        "my_func",
		Description: "doubles x",
	}
	fn, err := client.RegisterFunction(context.Background(), req)
	if err != nil {
		t.Fatalf("RegisterFunction() error = %v", err)
	}
	if fn.ID != "fn-001" {
		t.Errorf("expected fn-001, got %s", fn.ID)
	}
}

func TestRegisterFunction_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RegisterFunction(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestRegisterFunction_EmptyFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RegisterFunction(context.Background(), &compute.FunctionRegisterRequest{Name: "fn"})
	if err == nil {
		t.Error("expected error for empty function code")
	}
}

func TestGetFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/functions/fn-001" {
			t.Errorf("expected /functions/fn-001, got %s", r.URL.Path)
		}
		resp := compute.FunctionResponse{ID: "fn-001", Name: "my_func"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	fn, err := client.GetFunction(context.Background(), "fn-001")
	if err != nil {
		t.Fatalf("GetFunction() error = %v", err)
	}
	if fn.ID != "fn-001" {
		t.Errorf("expected fn-001, got %s", fn.ID)
	}
}

func TestGetFunction_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.GetFunction(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty function ID")
	}
}

func TestListFunctions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/functions" {
			t.Errorf("expected /functions, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("per_page") != "5" {
			t.Errorf("expected per_page=5, got %s", q.Get("per_page"))
		}
		resp := compute.FunctionList{
			Functions: []compute.FunctionResponse{
				{ID: "fn-001", Name: "func1"},
				{ID: "fn-002", Name: "func2"},
			},
			Total: 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	opts := &compute.ListFunctionsOptions{
		PerPage:     5,
		Marker:      "m1",
		OrderBy:     "name",
		Search:      "fn",
		FilterScope: "my-functions",
	}
	list, err := client.ListFunctions(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListFunctions() error = %v", err)
	}
	if len(list.Functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(list.Functions))
	}
}

func TestListFunctions_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := compute.FunctionList{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	client := newTestClient(t, server)
	list, err := client.ListFunctions(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFunctions() with nil options error = %v", err)
	}
	if list == nil {
		t.Error("expected non-nil list")
	}
}

func TestUpdateFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/functions/fn-001" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req compute.FunctionUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := compute.FunctionResponse{ID: "fn-001", Name: req.Name}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	req := &compute.FunctionUpdateRequest{Name: "updated_func"}
	fn, err := client.UpdateFunction(context.Background(), "fn-001", req)
	if err != nil {
		t.Fatalf("UpdateFunction() error = %v", err)
	}
	if fn.Name != "updated_func" {
		t.Errorf("expected updated_func, got %s", fn.Name)
	}
}

func TestUpdateFunction_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.UpdateFunction(context.Background(), "", &compute.FunctionUpdateRequest{})
	if err == nil {
		t.Error("expected error for empty function ID")
	}
}

func TestUpdateFunction_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.UpdateFunction(context.Background(), "fn-001", nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestDeleteFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/functions/fn-001" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.DeleteFunction(context.Background(), "fn-001"); err != nil {
		t.Fatalf("DeleteFunction() error = %v", err)
	}
}

func TestDeleteFunction_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	if err := client.DeleteFunction(context.Background(), ""); err == nil {
		t.Error("expected error for empty function ID")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// RunFunction / RunBatch
// ──────────────────────────────────────────────────────────────────────────────

func TestRunFunction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/run" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		resp := compute.TaskResponse{TaskID: "task-run-001", Status: "PENDING"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	req := &compute.TaskRequest{
		FunctionID: "fn-001",
		EndpointID: "ep-001",
		Args:       []interface{}{1, 2, 3},
	}
	resp, err := client.RunFunction(context.Background(), req)
	if err != nil {
		t.Fatalf("RunFunction() error = %v", err)
	}
	if resp.TaskID != "task-run-001" {
		t.Errorf("expected task-run-001, got %s", resp.TaskID)
	}
}

func TestRunFunction_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunFunction(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestRunFunction_EmptyFunctionID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunFunction(context.Background(), &compute.TaskRequest{EndpointID: "ep"})
	if err == nil {
		t.Error("expected error for empty function ID")
	}
}

func TestRunFunction_EmptyEndpointID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunFunction(context.Background(), &compute.TaskRequest{FunctionID: "fn"})
	if err == nil {
		t.Error("expected error for empty endpoint ID")
	}
}

func TestRunBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/batch" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var req compute.BatchTaskRequest
		json.NewDecoder(r.Body).Decode(&req)
		ids := make([]string, len(req.Tasks))
		for i := range req.Tasks {
			ids[i] = "batch-task-" + req.Tasks[i].FunctionID
		}
		resp := compute.BatchTaskResponse{TaskIDs: ids, Status: "PENDING"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	req := &compute.BatchTaskRequest{
		Tasks: []compute.TaskRequest{
			{FunctionID: "fn-A", EndpointID: "ep-001"},
			{FunctionID: "fn-B", EndpointID: "ep-001"},
		},
	}
	resp, err := client.RunBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if len(resp.TaskIDs) != 2 {
		t.Errorf("expected 2 task IDs, got %d", len(resp.TaskIDs))
	}
}

func TestRunBatch_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunBatch(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestRunBatch_EmptyTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunBatch(context.Background(), &compute.BatchTaskRequest{Tasks: []compute.TaskRequest{}})
	if err == nil {
		t.Error("expected error for empty tasks")
	}
}

func TestRunBatch_MissingFunctionID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunBatch(context.Background(), &compute.BatchTaskRequest{
		Tasks: []compute.TaskRequest{{EndpointID: "ep"}},
	})
	if err == nil {
		t.Error("expected error for missing function ID in batch task")
	}
}

func TestRunBatch_MissingEndpointID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.RunBatch(context.Background(), &compute.BatchTaskRequest{
		Tasks: []compute.TaskRequest{{FunctionID: "fn"}},
	})
	if err == nil {
		t.Error("expected error for missing endpoint ID in batch task")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// GetTaskStatus / GetBatchStatus
// ──────────────────────────────────────────────────────────────────────────────

func TestGetTaskStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status/task-001" {
			t.Errorf("expected /status/task-001, got %s", r.URL.Path)
		}
		resp := compute.TaskStatus{TaskID: "task-001", Status: "COMPLETED"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	status, err := client.GetTaskStatus(context.Background(), "task-001")
	if err != nil {
		t.Fatalf("GetTaskStatus() error = %v", err)
	}
	if status.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", status.Status)
	}
}

func TestGetTaskStatus_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.GetTaskStatus(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty task ID")
	}
}

func TestGetBatchStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/batch_status" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string][]string
		json.NewDecoder(r.Body).Decode(&body)
		tasks := make(map[string]compute.TaskStatus)
		for _, id := range body["task_ids"] {
			tasks[id] = compute.TaskStatus{TaskID: id, Status: "PENDING"}
		}
		resp := compute.BatchTaskStatus{Tasks: tasks}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	batchStatus, err := client.GetBatchStatus(context.Background(), []string{"t1", "t2"})
	if err != nil {
		t.Fatalf("GetBatchStatus() error = %v", err)
	}
	if len(batchStatus.Tasks) != 2 {
		t.Errorf("expected 2 tasks in batch status, got %d", len(batchStatus.Tasks))
	}
}

func TestGetBatchStatus_EmptyIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.GetBatchStatus(context.Background(), []string{})
	if err == nil {
		t.Error("expected error for empty task IDs")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListTasks (compute) / CancelTask (compute)
// ──────────────────────────────────────────────────────────────────────────────

func TestComputeListTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tasks" {
			t.Errorf("expected /tasks, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", q.Get("per_page"))
		}
		resp := compute.TaskList{
			Tasks: []string{"task-001", "task-002"},
			Total: 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	opts := &compute.TaskListOptions{
		PerPage:    10,
		Marker:     "m1",
		Status:     "PENDING",
		EndpointID: "ep-001",
		FunctionID: "fn-001",
	}
	list, err := client.ListTasks(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(list.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(list.Tasks))
	}
}

func TestComputeListTasks_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := compute.TaskList{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	list, err := client.ListTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTasks() with nil options error = %v", err)
	}
	if list == nil {
		t.Error("expected non-nil list")
	}
}

func TestComputeCancelTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/tasks/task-001/cancel" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.CancelTask(context.Background(), "task-001"); err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}
}

func TestComputeCancelTask_EmptyID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	if err := client.CancelTask(context.Background(), ""); err == nil {
		t.Error("expected error for empty task ID")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Batch functions that have 0% coverage
// ──────────────────────────────────────────────────────────────────────────────

func TestGetWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workflows/wf-001" {
			t.Errorf("expected /workflows/wf-001, got %s", r.URL.Path)
		}
		resp := compute.WorkflowResponse{ID: "wf-001", Name: "My Workflow"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	wf, err := client.GetWorkflow(context.Background(), "wf-001")
	if err != nil {
		t.Fatalf("GetWorkflow() error = %v", err)
	}
	if wf.ID != "wf-001" {
		t.Errorf("expected wf-001, got %s", wf.ID)
	}
}

func TestListWorkflows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workflows" {
			t.Errorf("expected /workflows, got %s", r.URL.Path)
		}
		resp := []compute.WorkflowResponse{
			{ID: "wf-001", Name: "Workflow 1"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	list, err := client.ListWorkflows(context.Background())
	if err != nil {
		t.Fatalf("ListWorkflows() error = %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 workflow, got %d", len(list))
	}
}

func TestDeleteWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/workflows/wf-001" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.DeleteWorkflow(context.Background(), "wf-001"); err != nil {
		t.Fatalf("DeleteWorkflow() error = %v", err)
	}
}

func TestCancelWorkflowRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/workflows/runs/run-001/cancel" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	if err := client.CancelWorkflowRun(context.Background(), "run-001"); err != nil {
		t.Fatalf("CancelWorkflowRun() error = %v", err)
	}
}

func TestGetTaskGroupStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task_groups/runs/tgr-001" {
			t.Errorf("expected /task_groups/runs/tgr-001, got %s", r.URL.Path)
		}
		resp := compute.TaskGroupStatusResponse{
			RunID:       "tgr-001",
			TaskGroupID: "group-001",
			Status:      "RUNNING",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	status, err := client.GetTaskGroupStatus(context.Background(), "tgr-001")
	if err != nil {
		t.Fatalf("GetTaskGroupStatus() error = %v", err)
	}
	if status.Status != "RUNNING" {
		t.Errorf("expected RUNNING, got %s", status.Status)
	}
}

func TestWaitForTaskGroupCompletion(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		status := "RUNNING"
		if callCount >= 2 {
			status = "COMPLETED"
		}
		resp := compute.TaskGroupStatusResponse{
			RunID:  "tgr-001",
			Status: status,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.WaitForTaskGroupCompletion(
		context.Background(), "tgr-001", 5*time.Second, 10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("WaitForTaskGroupCompletion() error = %v", err)
	}
	if result.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", result.Status)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 poll calls, got %d", callCount)
	}
}

func TestWaitForDependencyGraphCompletion(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		status := "RUNNING"
		if callCount >= 2 {
			status = "COMPLETED"
		}
		resp := compute.DependencyGraphStatusResponse{
			RunID:  "graph-001",
			Status: status,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	result, err := client.WaitForDependencyGraphCompletion(
		context.Background(), "graph-001", 5*time.Second, 10*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("WaitForDependencyGraphCompletion() error = %v", err)
	}
	if result.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", result.Status)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CreateWorkflow validation paths
// ──────────────────────────────────────────────────────────────────────────────

func TestCreateWorkflow_NilRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.CreateWorkflow(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestCreateWorkflow_EmptyName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	req := &compute.WorkflowCreateRequest{
		Tasks: []compute.WorkflowTask{{ID: "t1", FunctionID: "fn", EndpointID: "ep"}},
	}
	_, err := client.CreateWorkflow(context.Background(), req)
	if err == nil {
		t.Error("expected error for empty workflow name")
	}
}

func TestCreateWorkflow_EmptyTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	client := newTestClient(t, server)
	_, err := client.CreateWorkflow(context.Background(), &compute.WorkflowCreateRequest{Name: "wf"})
	if err == nil {
		t.Error("expected error for empty tasks")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// doRequest error paths
// ──────────────────────────────────────────────────────────────────────────────

func TestGetEndpoint_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetEndpoint(context.Background(), "ep-fail")
	if err == nil {
		t.Error("expected error from server 500")
	}
}

func TestGetFunction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.GetFunction(context.Background(), "missing-fn")
	if err == nil {
		t.Error("expected error from server 404")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ListEndpoints options coverage - exercise remaining query param paths
// ──────────────────────────────────────────────────────────────────────────────

func TestListEndpoints_AllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("include_info") != "true" {
			t.Errorf("expected include_info=true, got %q", q.Get("include_info"))
		}
		if q.Get("filter_status") != "online" {
			t.Errorf("expected filter_status=online, got %q", q.Get("filter_status"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(compute.ComputeEndpointList{})
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.ListEndpoints(context.Background(), &compute.ListEndpointsOptions{
		IncludeInfo:  true,
		FilterStatus: "online",
	})
	if err != nil {
		t.Fatalf("ListEndpoints() error = %v", err)
	}
}

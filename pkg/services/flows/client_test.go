// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

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

func TestListFlows(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/flows" {
			t.Errorf("Expected path /flows, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if limit := queryParams.Get("limit"); limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
		}

		// Return mock response
		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := FlowList{
			Flows: []Flow{
				{
					ID:          "test-flow-id",
					Title:       "Test Flow",
					Description: "A test flow",
					FlowOwner:   "test-user",
					CreatedAt:   flowTime,
					UpdatedAt:   flowTime,
					Definition: map[string]interface{}{
						"Comment": "This is a test flow definition",
					},
					RunCount: 0,
					Public:   false,
				},
			},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list flows
	options := &ListFlowsOptions{
		Limit: 10,
	}

	flowList, err := client.ListFlows(context.Background(), options)
	if err != nil {
		t.Fatalf("ListFlows() error = %v", err)
	}

	// Check response
	if len(flowList.Flows) != 1 {
		t.Errorf("Expected 1 flow, got %d", len(flowList.Flows))
	}

	flow := flowList.Flows[0]
	if flow.ID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", flow.ID)
	}
	if flow.Title != "Test Flow" {
		t.Errorf("Expected flow title = Test Flow, got %s", flow.Title)
	}
}

func TestGetFlow(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/flows/test-flow-id" {
			t.Errorf("Expected path /flows/test-flow-id, got %s", r.URL.Path)
		}

		// Return mock response
		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Flow{
			ID:          "test-flow-id",
			Title:       "Test Flow",
			Description: "A test flow",
			FlowOwner:   "test-user",
			CreatedAt:   flowTime,
			UpdatedAt:   flowTime,
			Definition: map[string]interface{}{
				"Comment": "This is a test flow definition",
			},
			RunCount: 0,
			Public:   false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get flow
	flow, err := client.GetFlow(context.Background(), "test-flow-id")
	if err != nil {
		t.Fatalf("GetFlow() error = %v", err)
	}

	// Check response
	if flow.ID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", flow.ID)
	}
	if flow.Title != "Test Flow" {
		t.Errorf("Expected flow title = Test Flow, got %s", flow.Title)
	}

	// Test empty flow ID
	_, err = client.GetFlow(context.Background(), "")
	if err == nil {
		t.Error("GetFlow() with empty ID should return error")
	}
}

func TestCreateFlow(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/flows" {
			t.Errorf("Expected path /flows, got %s", r.URL.Path)
		}

		// Decode request body
		var request FlowCreateRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.Title != "New Test Flow" {
			t.Errorf("Expected flow title = New Test Flow, got %s", request.Title)
		}

		// Return mock response
		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Flow{
			ID:          "new-test-flow-id",
			Title:       request.Title,
			Description: request.Description,
			FlowOwner:   "test-user",
			CreatedAt:   flowTime,
			UpdatedAt:   flowTime,
			Definition:  request.Definition,
			RunCount:    0,
			Public:      request.Public,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test create flow
	createRequest := &FlowCreateRequest{
		Title:       "New Test Flow",
		Description: "A new test flow",
		Definition: map[string]interface{}{
			"Comment": "This is a test flow definition",
		},
	}

	flow, err := client.CreateFlow(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("CreateFlow() error = %v", err)
	}

	// Check response
	if flow.ID != "new-test-flow-id" {
		t.Errorf("Expected flow ID = new-test-flow-id, got %s", flow.ID)
	}
	if flow.Title != "New Test Flow" {
		t.Errorf("Expected flow title = New Test Flow, got %s", flow.Title)
	}

	// Test nil request
	_, err = client.CreateFlow(context.Background(), nil)
	if err == nil {
		t.Error("CreateFlow() with nil request should return error")
	}

	// Test empty title
	_, err = client.CreateFlow(context.Background(), &FlowCreateRequest{
		Definition: map[string]interface{}{},
	})
	if err == nil {
		t.Error("CreateFlow() with empty title should return error")
	}

	// Test nil definition
	_, err = client.CreateFlow(context.Background(), &FlowCreateRequest{
		Title: "Test Flow",
	})
	if err == nil {
		t.Error("CreateFlow() with nil definition should return error")
	}
}

func TestUpdateFlow(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/flows/test-flow-id" {
			t.Errorf("Expected path /flows/test-flow-id, got %s", r.URL.Path)
		}

		// Decode request body
		var request FlowUpdateRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.Title != "Updated Test Flow" {
			t.Errorf("Expected flow title = Updated Test Flow, got %s", request.Title)
		}

		// Return mock response
		flowTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := Flow{
			ID:          "test-flow-id",
			Title:       request.Title,
			Description: "Updated description",
			FlowOwner:   "test-user",
			CreatedAt:   flowTime,
			UpdatedAt:   time.Now(),
			Definition: map[string]interface{}{
				"Comment": "This is an updated test flow definition",
			},
			RunCount: 0,
			Public:   false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test update flow
	updateRequest := &FlowUpdateRequest{
		Title:       "Updated Test Flow",
		Description: "Updated description",
		Definition: map[string]interface{}{
			"Comment": "This is an updated test flow definition",
		},
	}

	flow, err := client.UpdateFlow(context.Background(), "test-flow-id", updateRequest)
	if err != nil {
		t.Fatalf("UpdateFlow() error = %v", err)
	}

	// Check response
	if flow.ID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", flow.ID)
	}
	if flow.Title != "Updated Test Flow" {
		t.Errorf("Expected flow title = Updated Test Flow, got %s", flow.Title)
	}

	// Test empty flow ID
	_, err = client.UpdateFlow(context.Background(), "", updateRequest)
	if err == nil {
		t.Error("UpdateFlow() with empty ID should return error")
	}

	// Test nil request
	_, err = client.UpdateFlow(context.Background(), "test-flow-id", nil)
	if err == nil {
		t.Error("UpdateFlow() with nil request should return error")
	}
}

func TestDeleteFlow(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/flows/test-flow-id" {
			t.Errorf("Expected path /flows/test-flow-id, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusNoContent)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test delete flow
	err := client.DeleteFlow(context.Background(), "test-flow-id")
	if err != nil {
		t.Fatalf("DeleteFlow() error = %v", err)
	}

	// Test empty flow ID
	err = client.DeleteFlow(context.Background(), "")
	if err == nil {
		t.Error("DeleteFlow() with empty ID should return error")
	}
}

func TestRunFlow(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs" {
			t.Errorf("Expected path /runs, got %s", r.URL.Path)
		}

		// Decode request body
		var request RunRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.FlowID != "test-flow-id" {
			t.Errorf("Expected flow ID = test-flow-id, got %s", request.FlowID)
		}

		// Return mock response
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunResponse{
			RunID:     "test-run-id",
			FlowID:    request.FlowID,
			Status:    "ACTIVE",
			CreatedAt: runTime,
			StartedAt: runTime,
			Label:     request.Label,
			Tags:      request.Tags,
			UserID:    "test-user",
			RunOwner:  "test-user",
			Input:     request.Input,
			FlowTitle: "Test Flow",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test run flow
	runRequest := &RunRequest{
		FlowID: "test-flow-id",
		Label:  "Test Run",
		Tags:   []string{"test"},
		Input: map[string]interface{}{
			"param1": "value1",
		},
	}

	run, err := client.RunFlow(context.Background(), runRequest)
	if err != nil {
		t.Fatalf("RunFlow() error = %v", err)
	}

	// Check response
	if run.RunID != "test-run-id" {
		t.Errorf("Expected run ID = test-run-id, got %s", run.RunID)
	}
	if run.FlowID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", run.FlowID)
	}

	// Test nil request
	_, err = client.RunFlow(context.Background(), nil)
	if err == nil {
		t.Error("RunFlow() with nil request should return error")
	}

	// Test empty flow ID
	_, err = client.RunFlow(context.Background(), &RunRequest{
		Input: map[string]interface{}{},
	})
	if err == nil {
		t.Error("RunFlow() with empty flow ID should return error")
	}

	// Test nil input
	_, err = client.RunFlow(context.Background(), &RunRequest{
		FlowID: "test-flow-id",
	})
	if err == nil {
		t.Error("RunFlow() with nil input should return error")
	}
}

func TestListRuns(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs" {
			t.Errorf("Expected path /runs, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if flowID := queryParams.Get("flow_id"); flowID != "test-flow-id" {
			t.Errorf("Expected flow_id=test-flow-id, got %s", flowID)
		}

		// Return mock response
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunList{
			Runs: []RunResponse{
				{
					RunID:     "test-run-id",
					FlowID:    "test-flow-id",
					Status:    "ACTIVE",
					CreatedAt: runTime,
					StartedAt: runTime,
					Label:     "Test Run",
					Tags:      []string{"test"},
					UserID:    "test-user",
					RunOwner:  "test-user",
					Input: map[string]interface{}{
						"param1": "value1",
					},
					FlowTitle: "Test Flow",
				},
			},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list runs
	options := &ListRunsOptions{
		FlowID: "test-flow-id",
		Limit:  10,
	}

	runList, err := client.ListRuns(context.Background(), options)
	if err != nil {
		t.Fatalf("ListRuns() error = %v", err)
	}

	// Check response
	if len(runList.Runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runList.Runs))
	}

	run := runList.Runs[0]
	if run.RunID != "test-run-id" {
		t.Errorf("Expected run ID = test-run-id, got %s", run.RunID)
	}
	if run.FlowID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", run.FlowID)
	}
}

func TestGetRun(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs/test-run-id" {
			t.Errorf("Expected path /runs/test-run-id, got %s", r.URL.Path)
		}

		// Return mock response
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunResponse{
			RunID:     "test-run-id",
			FlowID:    "test-flow-id",
			Status:    "ACTIVE",
			CreatedAt: runTime,
			StartedAt: runTime,
			Label:     "Test Run",
			Tags:      []string{"test"},
			UserID:    "test-user",
			RunOwner:  "test-user",
			Input: map[string]interface{}{
				"param1": "value1",
			},
			FlowTitle: "Test Flow",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get run
	run, err := client.GetRun(context.Background(), "test-run-id")
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}

	// Check response
	if run.RunID != "test-run-id" {
		t.Errorf("Expected run ID = test-run-id, got %s", run.RunID)
	}
	if run.FlowID != "test-flow-id" {
		t.Errorf("Expected flow ID = test-flow-id, got %s", run.FlowID)
	}

	// Test empty run ID
	_, err = client.GetRun(context.Background(), "")
	if err == nil {
		t.Error("GetRun() with empty ID should return error")
	}
}

func TestCancelRun(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs/test-run-id/cancel" {
			t.Errorf("Expected path /runs/test-run-id/cancel, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusNoContent)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test cancel run
	err := client.CancelRun(context.Background(), "test-run-id")
	if err != nil {
		t.Fatalf("CancelRun() error = %v", err)
	}

	// Test empty run ID
	err = client.CancelRun(context.Background(), "")
	if err == nil {
		t.Error("CancelRun() with empty ID should return error")
	}
}

func TestUpdateRun(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs/test-run-id" {
			t.Errorf("Expected path /runs/test-run-id, got %s", r.URL.Path)
		}

		// Decode request body
		var request RunUpdateRequest
		json.NewDecoder(r.Body).Decode(&request)

		// Check request body
		if request.Label != "Updated Test Run" {
			t.Errorf("Expected run label = Updated Test Run, got %s", request.Label)
		}

		// Return mock response
		runTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunResponse{
			RunID:     "test-run-id",
			FlowID:    "test-flow-id",
			Status:    "ACTIVE",
			CreatedAt: runTime,
			StartedAt: runTime,
			Label:     request.Label,
			Tags:      request.Tags,
			UserID:    "test-user",
			RunOwner:  "test-user",
			Input: map[string]interface{}{
				"param1": "value1",
			},
			FlowTitle: "Test Flow",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test update run
	updateRequest := &RunUpdateRequest{
		Label: "Updated Test Run",
		Tags:  []string{"test", "updated"},
	}

	run, err := client.UpdateRun(context.Background(), "test-run-id", updateRequest)
	if err != nil {
		t.Fatalf("UpdateRun() error = %v", err)
	}

	// Check response
	if run.RunID != "test-run-id" {
		t.Errorf("Expected run ID = test-run-id, got %s", run.RunID)
	}
	if run.Label != "Updated Test Run" {
		t.Errorf("Expected run label = Updated Test Run, got %s", run.Label)
	}

	// Test empty run ID
	_, err = client.UpdateRun(context.Background(), "", updateRequest)
	if err == nil {
		t.Error("UpdateRun() with empty ID should return error")
	}

	// Test nil request
	_, err = client.UpdateRun(context.Background(), "test-run-id", nil)
	if err == nil {
		t.Error("UpdateRun() with nil request should return error")
	}
}

func TestGetRunLogs(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/runs/test-run-id/log" {
			t.Errorf("Expected path /runs/test-run-id/log, got %s", r.URL.Path)
		}

		// Check query parameters
		queryParams := r.URL.Query()
		if limit := queryParams.Get("limit"); limit != "10" {
			t.Errorf("Expected limit=10, got %s", limit)
		}

		// Return mock response
		logTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := RunLogList{
			Entries: []RunLogEntry{
				{
					Code:        "STARTED",
					RunID:       "test-run-id",
					CreatedAt:   logTime,
					Description: "Flow run started",
				},
				{
					Code:        "STEP_STARTED",
					RunID:       "test-run-id",
					CreatedAt:   logTime.Add(time.Second),
					Description: "Flow step started",
					Details: map[string]interface{}{
						"step_id": "step1",
					},
				},
			},
			Total:   2,
			HadMore: false,
			Offset:  0,
			Limit:   10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get run logs
	logs, err := client.GetRunLogs(context.Background(), "test-run-id", 10, 0)
	if err != nil {
		t.Fatalf("GetRunLogs() error = %v", err)
	}

	// Check response
	if len(logs.Entries) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(logs.Entries))
	}

	entry := logs.Entries[0]
	if entry.Code != "STARTED" {
		t.Errorf("Expected log code = STARTED, got %s", entry.Code)
	}
	if entry.RunID != "test-run-id" {
		t.Errorf("Expected run ID = test-run-id, got %s", entry.RunID)
	}

	// Test empty run ID
	_, err = client.GetRunLogs(context.Background(), "", 10, 0)
	if err == nil {
		t.Error("GetRunLogs() with empty ID should return error")
	}
}

func TestListActionProviders(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/action_providers" {
			t.Errorf("Expected path /action_providers, got %s", r.URL.Path)
		}

		// Return mock response
		providerTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := ActionProviderList{
			ActionProviders: []ActionProvider{
				{
					ID:          "test-provider-id",
					DisplayName: "Test Provider",
					Description: "A test action provider",
					Owner:       "globus",
					CreatedAt:   providerTime,
					UpdatedAt:   providerTime,
					Type:        "action",
					Globus:      true,
					Visible:     true,
				},
			},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list action providers
	options := &ListActionProvidersOptions{
		Limit: 10,
	}

	providerList, err := client.ListActionProviders(context.Background(), options)
	if err != nil {
		t.Fatalf("ListActionProviders() error = %v", err)
	}

	// Check response
	if len(providerList.ActionProviders) != 1 {
		t.Errorf("Expected 1 action provider, got %d", len(providerList.ActionProviders))
	}

	provider := providerList.ActionProviders[0]
	if provider.ID != "test-provider-id" {
		t.Errorf("Expected provider ID = test-provider-id, got %s", provider.ID)
	}
	if provider.DisplayName != "Test Provider" {
		t.Errorf("Expected provider display name = Test Provider, got %s", provider.DisplayName)
	}
}

func TestGetActionProvider(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/action_providers/test-provider-id" {
			t.Errorf("Expected path /action_providers/test-provider-id, got %s", r.URL.Path)
		}

		// Return mock response
		providerTime, _ := time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
		response := ActionProvider{
			ID:          "test-provider-id",
			DisplayName: "Test Provider",
			Description: "A test action provider",
			Owner:       "globus",
			CreatedAt:   providerTime,
			UpdatedAt:   providerTime,
			Type:        "action",
			Globus:      true,
			Visible:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get action provider
	provider, err := client.GetActionProvider(context.Background(), "test-provider-id")
	if err != nil {
		t.Fatalf("GetActionProvider() error = %v", err)
	}

	// Check response
	if provider.ID != "test-provider-id" {
		t.Errorf("Expected provider ID = test-provider-id, got %s", provider.ID)
	}
	if provider.DisplayName != "Test Provider" {
		t.Errorf("Expected provider display name = Test Provider, got %s", provider.DisplayName)
	}

	// Test empty provider ID
	_, err = client.GetActionProvider(context.Background(), "")
	if err == nil {
		t.Error("GetActionProvider() with empty ID should return error")
	}
}

func TestListActionRoles(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/action_providers/test-provider-id/roles" {
			t.Errorf("Expected path /action_providers/test-provider-id/roles, got %s", r.URL.Path)
		}

		// Return mock response
		response := ActionRoleList{
			ActionRoles: []ActionRole{
				{
					ID:          "test-role-id",
					Name:        "Test Role",
					Description: "A test action role",
					Visible:     true,
				},
			},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   10,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list action roles
	roleList, err := client.ListActionRoles(context.Background(), "test-provider-id", 10, 0)
	if err != nil {
		t.Fatalf("ListActionRoles() error = %v", err)
	}

	// Check response
	if len(roleList.ActionRoles) != 1 {
		t.Errorf("Expected 1 action role, got %d", len(roleList.ActionRoles))
	}

	role := roleList.ActionRoles[0]
	if role.ID != "test-role-id" {
		t.Errorf("Expected role ID = test-role-id, got %s", role.ID)
	}
	if role.Name != "Test Role" {
		t.Errorf("Expected role name = Test Role, got %s", role.Name)
	}

	// Test empty provider ID
	_, err = client.ListActionRoles(context.Background(), "", 10, 0)
	if err == nil {
		t.Error("ListActionRoles() with empty provider ID should return error")
	}
}

func TestGetActionRole(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/action_providers/test-provider-id/roles/test-role-id" {
			t.Errorf("Expected path /action_providers/test-provider-id/roles/test-role-id, got %s", r.URL.Path)
		}

		// Return mock response
		response := ActionRole{
			ID:          "test-role-id",
			Name:        "Test Role",
			Description: "A test action role",
			Visible:     true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get action role
	role, err := client.GetActionRole(context.Background(), "test-provider-id", "test-role-id")
	if err != nil {
		t.Fatalf("GetActionRole() error = %v", err)
	}

	// Check response
	if role.ID != "test-role-id" {
		t.Errorf("Expected role ID = test-role-id, got %s", role.ID)
	}
	if role.Name != "Test Role" {
		t.Errorf("Expected role name = Test Role, got %s", role.Name)
	}

	// Test empty provider ID
	_, err = client.GetActionRole(context.Background(), "", "test-role-id")
	if err == nil {
		t.Error("GetActionRole() with empty provider ID should return error")
	}

	// Test empty role ID
	_, err = client.GetActionRole(context.Background(), "test-provider-id", "")
	if err == nil {
		t.Error("GetActionRole() with empty role ID should return error")
	}
}

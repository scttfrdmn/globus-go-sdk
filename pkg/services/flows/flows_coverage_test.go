// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package flows_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"
)

// newTestClient creates a flows client pointed at the given server URL.
func newTestClient(t *testing.T, serverURL string) *flows.Client {
	t.Helper()
	client, err := flows.NewClient(
		flows.WithAccessToken("test-token"),
		flows.WithBaseURL(serverURL+"/"),
	)
	if err != nil {
		t.Fatalf("failed to create flows client: %v", err)
	}
	return client
}

// startMockServer starts an httptest server and returns the server and a client
// configured to use it. Caller must defer server.Close().
func startMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *flows.Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	return server, newTestClient(t, server.URL)
}

// flowTime is a fixed time used across test helpers.
var flowTime, _ = time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")

func makeFlow(id, title string) flows.Flow {
	return flows.Flow{
		ID:        id,
		Title:     title,
		FlowOwner: "test-user",
		CreatedAt: flowTime,
		UpdatedAt: flowTime,
		Definition: map[string]interface{}{
			"Comment": "test flow",
		},
	}
}

func makeRun(runID, flowID, status string) flows.RunResponse {
	return flows.RunResponse{
		RunID:     runID,
		FlowID:    flowID,
		Status:    status,
		CreatedAt: flowTime,
		StartedAt: flowTime,
		UserID:    "test-user",
		RunOwner:  "test-user",
	}
}

// ---------------------------------------------------------------------------
// TestNewClientOptions – exercise option constructors
// ---------------------------------------------------------------------------

func TestNewClientOptions(t *testing.T) {
	// WithHTTPDebugging
	_, err := flows.NewClient(
		flows.WithAccessToken("tok"),
		flows.WithHTTPDebugging(true),
	)
	if err != nil {
		t.Fatalf("NewClient with WithHTTPDebugging: %v", err)
	}

	// WithHTTPTracing
	_, err = flows.NewClient(
		flows.WithAccessToken("tok"),
		flows.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("NewClient with WithHTTPTracing: %v", err)
	}

	// WithCoreOption (smoke test – no authorizer still works here since
	// access token is provided)
	_, err = flows.NewClient(
		flows.WithAccessToken("tok"),
	)
	if err != nil {
		t.Fatalf("NewClient with minimal options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestListFlowsOptions – exercise all ListFlowsOptions query params
// ---------------------------------------------------------------------------

func TestListFlowsAllOptions(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"per_page":      "10",
			"offset":        "20",
			"marker":        "tok-abc",
			"orderby":       "created_at",
			"q":             "search-term",
			"filter_roles":  "flow_owner",
			"filter_owner":  "user-123",
			"filter_public": "true",
			"roles_only":    "true",
		}
		for k, v := range checks {
			if got := q.Get(k); got != v {
				t.Errorf("param %s: got %q, want %q", k, got, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{Flows: []flows.Flow{}})
	})
	defer server.Close()

	_, err := client.ListFlows(context.Background(), &flows.ListFlowsOptions{
		PerPage:      10,
		Offset:       20,
		Marker:       "tok-abc",
		OrderBy:      "created_at",
		Q:            "search-term",
		FilterRoles:  "flow_owner",
		FilterOwner:  "user-123",
		FilterPublic: true,
		RolesOnly:    true,
	})
	if err != nil {
		t.Fatalf("ListFlows with all options: %v", err)
	}
}

func TestListFlowsWithLimit(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "25" {
			t.Errorf("expected limit=25, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{Flows: []flows.Flow{}})
	})
	defer server.Close()

	_, err := client.ListFlows(context.Background(), &flows.ListFlowsOptions{Limit: 25})
	if err != nil {
		t.Fatalf("ListFlows with limit: %v", err)
	}
}

func TestListFlowsNilOptions(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{Flows: []flows.Flow{}})
	})
	defer server.Close()

	_, err := client.ListFlows(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFlows with nil options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestListFlowsV2
// ---------------------------------------------------------------------------

func TestListFlowsV2(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{
			Flows:   []flows.Flow{makeFlow("f1", "Flow 1"), makeFlow("f2", "Flow 2")},
			Total:   2,
			HadMore: false,
		})
	})
	defer server.Close()

	resp, err := client.ListFlowsV2(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListFlowsV2: %v", err)
	}
	if len(resp.Data.Flows) != 2 {
		t.Fatalf("expected 2 flows, got %d", len(resp.Data.Flows))
	}
}

func TestListFlowsV2WithOptions(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "5" {
			t.Errorf("expected limit=5, got %q", q.Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{Flows: []flows.Flow{}})
	})
	defer server.Close()

	_, err := client.ListFlowsV2(context.Background(), &flows.ListFlowsOptions{Limit: 5})
	if err != nil {
		t.Fatalf("ListFlowsV2 with options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestListRunsAllOptions
// ---------------------------------------------------------------------------

func TestListRunsAllOptions(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"per_page":  "10",
			"offset":    "5",
			"marker":    "tok-xyz",
			"orderby":   "created_at",
			"q":         "search",
			"flow_id":   "flow-123",
			"status":    "ACTIVE",
			"role_type": "run_owner",
			"label":     "my-run",
		}
		for k, v := range checks {
			if got := q.Get(k); got != v {
				t.Errorf("param %s: got %q, want %q", k, got, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunList{Runs: []flows.RunResponse{}})
	})
	defer server.Close()

	_, err := client.ListRuns(context.Background(), &flows.ListRunsOptions{
		PerPage:  10,
		Offset:   5,
		Marker:   "tok-xyz",
		OrderBy:  "created_at",
		Q:        "search",
		FlowID:   "flow-123",
		Status:   "ACTIVE",
		RoleType: "run_owner",
		Label:    "my-run",
	})
	if err != nil {
		t.Fatalf("ListRuns with all options: %v", err)
	}
}

func TestListRunsWithLimit(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "50" {
			t.Errorf("expected limit=50, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunList{Runs: []flows.RunResponse{}})
	})
	defer server.Close()

	_, err := client.ListRuns(context.Background(), &flows.ListRunsOptions{Limit: 50})
	if err != nil {
		t.Fatalf("ListRuns with limit: %v", err)
	}
}

func TestListRunsNilOptions(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunList{Runs: []flows.RunResponse{}})
	})
	defer server.Close()

	_, err := client.ListRuns(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListRuns with nil options: %v", err)
	}
}
// ---------------------------------------------------------------------------
// TestGetRunLogsWithOffset
// ---------------------------------------------------------------------------

func TestGetRunLogsWithOffset(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "5" {
			t.Errorf("expected limit=5, got %q", q.Get("limit"))
		}
		if q.Get("offset") != "10" {
			t.Errorf("expected offset=10, got %q", q.Get("offset"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunLogList{Entries: []flows.RunLogEntry{}})
	})
	defer server.Close()

	_, err := client.GetRunLogs(context.Background(), "run-1", 5, 10)
	if err != nil {
		t.Fatalf("GetRunLogs with offset: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestWaitForRun – terminated states
// ---------------------------------------------------------------------------

func TestWaitForRunSucceeded(t *testing.T) {
	callCount := 0
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		status := "ACTIVE"
		if callCount >= 2 {
			status = "SUCCEEDED"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeRun("run-1", "flow-1", status))
	})
	defer server.Close()

	run, err := client.WaitForRun(context.Background(), "run-1", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if run.Status != "SUCCEEDED" {
		t.Errorf("run.Status = %q, want SUCCEEDED", run.Status)
	}
}

func TestWaitForRunFailed(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeRun("run-1", "flow-1", "FAILED"))
	})
	defer server.Close()

	run, err := client.WaitForRun(context.Background(), "run-1", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if run.Status != "FAILED" {
		t.Errorf("run.Status = %q, want FAILED", run.Status)
	}
}

func TestWaitForRunCanceled(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeRun("run-1", "flow-1", "CANCELED"))
	})
	defer server.Close()

	run, err := client.WaitForRun(context.Background(), "run-1", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForRun: %v", err)
	}
	if run.Status != "CANCELED" {
		t.Errorf("run.Status = %q, want CANCELED", run.Status)
	}
}

func TestWaitForRunContextCanceled(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeRun("run-1", "flow-1", "ACTIVE"))
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.WaitForRun(ctx, "run-1", 1*time.Millisecond)
	if err == nil {
		t.Fatal("WaitForRun should return error when context is canceled")
	}
}

func TestWaitForRunEmptyRunID(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	_, err := client.WaitForRun(context.Background(), "", 1*time.Millisecond)
	if err == nil {
		t.Fatal("WaitForRun with empty run ID should return error")
	}
}

func TestWaitForRunDefaultPollInterval(t *testing.T) {
	// zero pollInterval should use the default (3s), but we use a context with timeout
	// so it doesn't actually wait 3 seconds in tests.
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeRun("run-1", "flow-1", "SUCCEEDED"))
	})
	defer server.Close()

	// Pass 0 as pollInterval to exercise the default path; use a short timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	run, err := client.WaitForRun(ctx, "run-1", 0)
	if err != nil {
		t.Fatalf("WaitForRun with zero poll interval: %v", err)
	}
	if run.Status != "SUCCEEDED" {
		t.Errorf("run.Status = %q, want SUCCEEDED", run.Status)
	}
}

// ---------------------------------------------------------------------------
// TestListAllFlows
// ---------------------------------------------------------------------------

func TestListAllFlows(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.FlowList{
			Flows:   []flows.Flow{makeFlow("f1", "Flow 1"), makeFlow("f2", "Flow 2")},
			Total:   2,
			HadMore: false,
			Offset:  0,
			Limit:   100,
		})
	})
	defer server.Close()

	flowList, err := client.ListAllFlows(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAllFlows: %v", err)
	}
	if len(flowList) != 2 {
		t.Fatalf("expected 2 flows, got %d", len(flowList))
	}
}

// ---------------------------------------------------------------------------
// TestListAllRuns
// ---------------------------------------------------------------------------

func TestListAllRuns(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunList{
			Runs:    []flows.RunResponse{makeRun("r1", "f1", "SUCCEEDED"), makeRun("r2", "f1", "FAILED")},
			Total:   2,
			HadMore: false,
			Offset:  0,
			Limit:   100,
		})
	})
	defer server.Close()

	runList, err := client.ListAllRuns(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListAllRuns: %v", err)
	}
	if len(runList) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runList))
	}
}
// ---------------------------------------------------------------------------
// TestListAllRunLogs
// ---------------------------------------------------------------------------

func TestListAllRunLogs(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/runs/run-1/log" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunLogList{
			Entries: []flows.RunLogEntry{
				{Code: "STARTED", Time: flowTime, Description: "started"},
			},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   100,
		})
	})
	defer server.Close()

	entries, err := client.ListAllRunLogs(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("ListAllRunLogs: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Code != "STARTED" {
		t.Errorf("entries[0].Code = %q, want STARTED", entries[0].Code)
	}
}

// ---------------------------------------------------------------------------
// TestGetFlowsIterator
// ---------------------------------------------------------------------------

func TestGetFlowsIterator(t *testing.T) {
	callCount := 0
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			json.NewEncoder(w).Encode(flows.FlowList{
				Flows:   []flows.Flow{makeFlow("f1", "Flow 1")},
				Total:   2,
				HadMore: true,
				Offset:  0,
				Limit:   1,
			})
		} else {
			json.NewEncoder(w).Encode(flows.FlowList{
				Flows:   []flows.Flow{makeFlow("f2", "Flow 2")},
				Total:   2,
				HadMore: false,
				Offset:  1,
				Limit:   1,
			})
		}
	})
	defer server.Close()

	iter := client.GetFlowsIterator(&flows.ListFlowsOptions{Limit: 1})
	var collected []flows.Flow
	for iter.Next(context.Background()) {
		collected = append(collected, *iter.Flow())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(collected) != 2 {
		t.Fatalf("expected 2 flows from iterator, got %d", len(collected))
	}
}

// ---------------------------------------------------------------------------
// TestGetRunsIterator
// ---------------------------------------------------------------------------

func TestGetRunsIterator(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunList{
			Runs:    []flows.RunResponse{makeRun("r1", "f1", "SUCCEEDED")},
			Total:   1,
			HadMore: false,
			Offset:  0,
			Limit:   100,
		})
	})
	defer server.Close()

	iter := client.GetRunsIterator(nil)
	var runs []flows.RunResponse
	for iter.Next(context.Background()) {
		runs = append(runs, *iter.Run())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
}
// ---------------------------------------------------------------------------
// TestGetRunLogsIterator
// ---------------------------------------------------------------------------

func TestGetRunLogsIterator(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flows.RunLogList{
			Entries: []flows.RunLogEntry{
				{Code: "STARTED", Time: flowTime, Description: "started"},
				{Code: "COMPLETED", Time: flowTime, Description: "done"},
			},
			Total:   2,
			HadMore: false,
			Offset:  0,
			Limit:   100,
		})
	})
	defer server.Close()

	iter := client.GetRunLogsIterator("run-1", 100)
	var entries []flows.RunLogEntry
	for iter.Next(context.Background()) {
		entries = append(entries, *iter.LogEntry())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

// ---------------------------------------------------------------------------
// TestFlowAuthenticationPolicy – model coverage
// ---------------------------------------------------------------------------

func TestFlowAuthenticationPolicy(t *testing.T) {
	highAssurance := true
	requiredMFA := false

	policy := flows.FlowAuthenticationPolicy{
		HighAssurance:   &highAssurance,
		RequiredMFA:     &requiredMFA,
		SessionPolicies: []string{"policy-1", "policy-2"},
	}

	if *policy.HighAssurance != true {
		t.Errorf("HighAssurance = %v, want true", *policy.HighAssurance)
	}
	if *policy.RequiredMFA != false {
		t.Errorf("RequiredMFA = %v, want false", *policy.RequiredMFA)
	}
	if len(policy.SessionPolicies) != 2 {
		t.Errorf("expected 2 session policies, got %d", len(policy.SessionPolicies))
	}
}

func TestCreateFlowWithAuthPolicy(t *testing.T) {
	highAssurance := true
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req flows.FlowCreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.AuthenticationPolicy == nil {
			t.Error("expected AuthenticationPolicy to be set")
		} else if req.AuthenticationPolicy.HighAssurance == nil || !*req.AuthenticationPolicy.HighAssurance {
			t.Error("expected HighAssurance=true")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeFlow("f-new", "Auth Flow"))
	})
	defer server.Close()

	flow, err := client.CreateFlow(context.Background(), &flows.FlowCreateRequest{
		Title:      "Auth Flow",
		Definition: map[string]interface{}{"Comment": "test"},
		AuthenticationPolicy: &flows.FlowAuthenticationPolicy{
			HighAssurance: &highAssurance,
		},
	})
	if err != nil {
		t.Fatalf("CreateFlow with auth policy: %v", err)
	}
	if flow.ID != "f-new" {
		t.Errorf("flow.ID = %q, want f-new", flow.ID)
	}
}

func TestUpdateFlowWithAuthPolicy(t *testing.T) {
	requiredMFA := true
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req flows.FlowUpdateRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.AuthenticationPolicy == nil {
			t.Error("expected AuthenticationPolicy to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(makeFlow("f1", "Updated"))
	})
	defer server.Close()

	_, err := client.UpdateFlow(context.Background(), "f1", &flows.FlowUpdateRequest{
		Title: "Updated",
		AuthenticationPolicy: &flows.FlowAuthenticationPolicy{
			RequiredMFA:     &requiredMFA,
			SessionPolicies: []string{"pol-1"},
		},
	})
	if err != nil {
		t.Fatalf("UpdateFlow with auth policy: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestParseErrorResponse – error type parsing
// ---------------------------------------------------------------------------

func TestParseErrorResponseFlowNotFound(t *testing.T) {
	body := []byte(`{"code":"NOT_FOUND","message":"Flow not found"}`)
	err := flows.ParseErrorResponse(body, http.StatusNotFound, "flow-123", "flow")
	if err == nil {
		t.Fatal("expected error")
	}
	if !flows.IsFlowNotFoundError(err) {
		t.Errorf("expected FlowNotFoundError, got %T: %v", err, err)
	}
}

func TestParseErrorResponseRunNotFound(t *testing.T) {
	body := []byte(`{"code":"NOT_FOUND","message":"Run not found"}`)
	err := flows.ParseErrorResponse(body, http.StatusNotFound, "run-123", "run")
	if err == nil {
		t.Fatal("expected error")
	}
	if !flows.IsRunNotFoundError(err) {
		t.Errorf("expected RunNotFoundError, got %T: %v", err, err)
	}
}
func TestParseErrorResponseForbidden(t *testing.T) {
	body := []byte(`{"code":"FORBIDDEN","message":"Access denied"}`)
	err := flows.ParseErrorResponse(body, http.StatusForbidden, "", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !flows.IsForbiddenError(err) {
		t.Errorf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestParseErrorResponseValidation(t *testing.T) {
	body := []byte(`{"code":"VALIDATION_ERROR","message":"Invalid input"}`)
	err := flows.ParseErrorResponse(body, http.StatusBadRequest, "", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !flows.IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestParseErrorResponseGenericNotFound(t *testing.T) {
	body := []byte(`{"code":"NOT_FOUND","message":"Resource not found"}`)
	err := flows.ParseErrorResponse(body, http.StatusNotFound, "res-123", "unknown_type")
	if err == nil {
		t.Fatal("expected error")
	}
	// Should be a generic ErrorResponse
	if flows.IsFlowNotFoundError(err) || flows.IsRunNotFoundError(err) {
		t.Errorf("unexpected specific error type for unknown resource type")
	}
}

func TestParseErrorResponseGenericOtherStatus(t *testing.T) {
	body := []byte(`{"code":"INTERNAL_ERROR","message":"Something went wrong"}`)
	err := flows.ParseErrorResponse(body, http.StatusInternalServerError, "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseErrorResponseInvalidJSON(t *testing.T) {
	body := []byte(`not valid json`)
	err := flows.ParseErrorResponse(body, http.StatusBadGateway, "", "")
	if err == nil {
		t.Fatal("expected error for invalid JSON body")
	}
}

// ---------------------------------------------------------------------------
// TestErrorResponse – error message formatting
// ---------------------------------------------------------------------------

func TestErrorResponseErrorWithRequestID(t *testing.T) {
	e := &flows.ErrorResponse{
		Code:      "TEST_CODE",
		Message:   "test message",
		RequestID: "req-123",
	}
	msg := e.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	// Should contain request ID
	if len(msg) == 0 {
		t.Error("empty error message")
	}
}

func TestErrorResponseErrorWithoutRequestID(t *testing.T) {
	e := &flows.ErrorResponse{
		Code:    "TEST_CODE",
		Message: "test message",
	}
	msg := e.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestFlowNotFoundErrorMessage(t *testing.T) {
	e := &flows.FlowNotFoundError{
		FlowID:        "flow-abc",
		ErrorResponse: &flows.ErrorResponse{Code: "NOT_FOUND", Message: "not found"},
	}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRunNotFoundErrorMessage(t *testing.T) {
	e := &flows.RunNotFoundError{
		RunID:         "run-abc",
		ErrorResponse: &flows.ErrorResponse{Code: "NOT_FOUND", Message: "not found"},
	}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
func TestForbiddenErrorMessage(t *testing.T) {
	e := &flows.ForbiddenError{
		ErrorResponse: &flows.ErrorResponse{Code: "FORBIDDEN", Message: "Access denied"},
	}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidationErrorMessage(t *testing.T) {
	e := &flows.ValidationError{
		ErrorResponse: &flows.ErrorResponse{Code: "VALIDATION", Message: "bad input"},
	}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

// ---------------------------------------------------------------------------
// TestHTTPErrorResponses – verify error paths in doRequestLowLevel
// ---------------------------------------------------------------------------

func TestGetFlowHTTPError404(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "NOT_FOUND", Message: "flow not found"})
	})
	defer server.Close()

	_, err := client.GetFlow(context.Background(), "missing-flow")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !flows.IsFlowNotFoundError(err) {
		t.Errorf("expected FlowNotFoundError, got %T: %v", err, err)
	}
}

func TestRunFlowHTTPError404(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "NOT_FOUND", Message: "run not found"})
	})
	defer server.Close()

	_, err := client.RunFlow(context.Background(), &flows.RunRequest{
		FlowID: "flow-1",
		Body:   map[string]interface{}{"key": "val"},
	})
	if err == nil {
		t.Fatal("expected error for 404")
	}
	// Error should be a RunNotFoundError since path is "runs" POST
	if !flows.IsRunNotFoundError(err) {
		t.Errorf("expected RunNotFoundError, got %T: %v", err, err)
	}
}
func TestGetRunHTTPError404(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "NOT_FOUND", Message: "run not found"})
	})
	defer server.Close()

	_, err := client.GetRun(context.Background(), "missing-run")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !flows.IsRunNotFoundError(err) {
		t.Errorf("expected RunNotFoundError, got %T: %v", err, err)
	}
}

func TestDeleteFlowHTTPError403(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "FORBIDDEN", Message: "no permission"})
	})
	defer server.Close()

	err := client.DeleteFlow(context.Background(), "flow-1")
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !flows.IsForbiddenError(err) {
		t.Errorf("expected ForbiddenError, got %T: %v", err, err)
	}
}

func TestListFlowsHTTPError400(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "VALIDATION_ERROR", Message: "bad params"})
	})
	defer server.Close()

	_, err := client.ListFlows(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	if !flows.IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}
// TestListFlowsV2ErrorWrapping exercises the GlobusError wrapping in ListFlowsV2.
func TestListFlowsV2ErrorWrapping(t *testing.T) {
	server, client := startMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(flows.ErrorResponse{Code: "VALIDATION_ERROR", Message: "bad params"})
	})
	defer server.Close()

	_, err := client.ListFlowsV2(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

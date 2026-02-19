// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewClient_Default verifies that NewClient creates a client with defaults.
func TestNewClient_Default(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}
	if client.Client == nil {
		t.Fatal("NewClient() returned client with nil inner Client")
	}
}

// TestNewClient_WithAccessToken verifies the WithAccessToken option.
func TestNewClient_WithAccessToken(t *testing.T) {
	client, err := NewClient(WithAccessToken("tok123"))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

// TestNewClient_WithBaseURL verifies the WithBaseURL option.
func TestNewClient_WithBaseURL(t *testing.T) {
	client, err := NewClient(WithBaseURL("https://example.com/api/"))
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if client.Client.BaseURL != "https://example.com/api/" {
		t.Errorf("expected BaseURL 'https://example.com/api/', got %q", client.Client.BaseURL)
	}
}

// TestNewClient_WithHTTPDebugging verifies that debugging option does not error.
func TestNewClient_WithHTTPDebugging(t *testing.T) {
	_, err := NewClient(WithHTTPDebugging(true))
	if err != nil {
		t.Fatalf("NewClient(WithHTTPDebugging) error: %v", err)
	}
}

// TestNewClient_WithHTTPTracing verifies that tracing option does not error.
func TestNewClient_WithHTTPTracing(t *testing.T) {
	_, err := NewClient(WithHTTPTracing(true))
	if err != nil {
		t.Fatalf("NewClient(WithHTTPTracing) error: %v", err)
	}
}

// TestCreateTimer_NilRequest checks that a nil request returns an error.
func TestCreateTimer_NilRequest(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.CreateTimer(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

// TestGetTimer_EmptyID checks that an empty timer ID returns an error.
func TestGetTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.GetTimer(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestUpdateTimer_EmptyID checks that an empty timer ID returns an error.
func TestUpdateTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.UpdateTimer(context.Background(), "", &UpdateTimerRequest{})
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestUpdateTimer_NilRequest checks that a nil update request returns an error.
func TestUpdateTimer_NilRequest(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.UpdateTimer(context.Background(), "abc", nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

// TestUpdateTimer_Success verifies a successful update operation.
func TestUpdateTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		timer := Timer{ID: "timer-1", Name: "Updated", Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	newName := "Updated"
	result, err := client.UpdateTimer(context.Background(), "timer-1", &UpdateTimerRequest{Name: &newName})
	if err != nil {
		t.Fatalf("UpdateTimer() error: %v", err)
	}
	if result.Name != "Updated" {
		t.Errorf("expected Name 'Updated', got %q", result.Name)
	}
}

// TestDeleteTimer_EmptyID checks that an empty timer ID returns an error.
func TestDeleteTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	err := client.DeleteTimer(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestDeleteTimer_Success verifies a successful delete operation.
func TestDeleteTimer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	err := client.DeleteTimer(context.Background(), "timer-1")
	if err != nil {
		t.Fatalf("DeleteTimer() error: %v", err)
	}
}

// TestDeleteTimer_ErrorResponse verifies that a 4xx from the server causes an error.
func TestDeleteTimer_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	err := client.DeleteTimer(context.Background(), "missing-timer")
	if err == nil {
		t.Fatal("expected error from 404 response, got nil")
	}
}

// TestListTimers_Nil verifies ListTimers with nil options succeeds.
func TestListTimers_Nil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := TimerList{Timers: []Timer{}, Total: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	list, err := client.ListTimers(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTimers(nil) error: %v", err)
	}
	if list.Total != 0 {
		t.Errorf("expected 0 total, got %d", list.Total)
	}
}

// TestListTimers_AllOptions verifies that all ListTimersOptions fields are properly forwarded.
func TestListTimers_AllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "5" {
			t.Errorf("expected limit=5, got %q", q.Get("limit"))
		}
		if q.Get("marker") != "page2" {
			t.Errorf("expected marker=page2, got %q", q.Get("marker"))
		}
		if q.Get("status") != "active" {
			t.Errorf("expected status=active, got %q", q.Get("status"))
		}
		if q.Get("schedule_type") != "once" {
			t.Errorf("expected schedule_type=once, got %q", q.Get("schedule_type"))
		}
		if q.Get("callback_type") != "flow" {
			t.Errorf("expected callback_type=flow, got %q", q.Get("callback_type"))
		}
		list := TimerList{Total: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	limit := 5
	marker := "page2"
	status := "active"
	schedType := "once"
	cbType := "flow"
	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	_, err := client.ListTimers(context.Background(), &ListTimersOptions{
		Limit:        &limit,
		Marker:       &marker,
		Status:       &status,
		ScheduleType: &schedType,
		CallbackType: &cbType,
	})
	if err != nil {
		t.Fatalf("ListTimers() error: %v", err)
	}
}

// TestListTimersV2_Success verifies the V2 list method with the unified response.
func TestListTimersV2_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := TimerList{
			Timers: []Timer{
				{ID: "t1", Name: "Timer1", Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now},
			},
			Total: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	resp, err := client.ListTimersV2(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTimersV2() error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestListTimersV2_AllOptions verifies all options are forwarded in V2.
func TestListTimersV2_AllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "3" {
			t.Errorf("expected limit=3, got %q", q.Get("limit"))
		}
		list := TimerList{Total: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	limit := 3
	marker := "mark"
	status := "paused"
	schedType := "cron"
	cbType := "web"
	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	_, err := client.ListTimersV2(context.Background(), &ListTimersOptions{
		Limit:        &limit,
		Marker:       &marker,
		Status:       &status,
		ScheduleType: &schedType,
		CallbackType: &cbType,
	})
	if err != nil {
		t.Fatalf("ListTimersV2() error: %v", err)
	}
}

// TestPauseTimer_EmptyID checks that an empty timer ID returns an error.
func TestPauseTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.PauseTimer(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestPauseTimer_Success verifies a successful pause operation.
func TestPauseTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		timer := Timer{ID: "timer-1", Status: string(TimerStatusPaused), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.PauseTimer(context.Background(), "timer-1")
	if err != nil {
		t.Fatalf("PauseTimer() error: %v", err)
	}
	if result.Status != string(TimerStatusPaused) {
		t.Errorf("expected status %q, got %q", TimerStatusPaused, result.Status)
	}
}

// TestResumeTimer_EmptyID checks that an empty timer ID returns an error.
func TestResumeTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.ResumeTimer(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestResumeTimer_Success verifies a successful resume operation.
func TestResumeTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := Timer{ID: "timer-1", Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.ResumeTimer(context.Background(), "timer-1")
	if err != nil {
		t.Fatalf("ResumeTimer() error: %v", err)
	}
	if result.ID != "timer-1" {
		t.Errorf("expected ID 'timer-1', got %q", result.ID)
	}
}

// TestRunTimer_EmptyID checks that an empty timer ID returns an error.
func TestRunTimer_EmptyID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.RunTimer(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestRunTimer_Success verifies a successful manual run trigger.
func TestRunTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		run := TimerRun{ID: "run-1", TimerID: "timer-1", Status: string(RunStatusPending), StartTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.RunTimer(context.Background(), "timer-1")
	if err != nil {
		t.Fatalf("RunTimer() error: %v", err)
	}
	if result.ID != "run-1" {
		t.Errorf("expected run ID 'run-1', got %q", result.ID)
	}
}

// TestListRuns_EmptyTimerID checks that an empty timer ID returns an error.
func TestListRuns_EmptyTimerID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.ListRuns(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestListRuns_Success verifies listing runs for a timer.
func TestListRuns_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/timers/timer-1/runs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		runList := TimerRunList{
			Runs:  []TimerRun{{ID: "run-1", TimerID: "timer-1", Status: string(RunStatusSuccess), StartTime: now}},
			Total: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runList)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.ListRuns(context.Background(), "timer-1", nil)
	if err != nil {
		t.Fatalf("ListRuns() error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 run, got %d", result.Total)
	}
}

// TestListRuns_AllOptions verifies that all ListRunsOptions fields are properly forwarded.
func TestListRuns_AllOptions(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "2" {
			t.Errorf("expected limit=2, got %q", q.Get("limit"))
		}
		if q.Get("marker") != "next" {
			t.Errorf("expected marker=next, got %q", q.Get("marker"))
		}
		if q.Get("status") != "success" {
			t.Errorf("expected status=success, got %q", q.Get("status"))
		}
		if q.Get("start_after") == "" {
			t.Error("expected start_after to be set")
		}
		if q.Get("start_before") == "" {
			t.Error("expected start_before to be set")
		}
		runList := TimerRunList{Total: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runList)
	}))
	defer server.Close()

	limit := 2
	marker := "next"
	status := "success"
	startAfter := now.Add(-24 * time.Hour)
	startBefore := now
	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	_, err := client.ListRuns(context.Background(), "timer-1", &ListRunsOptions{
		Limit:       &limit,
		Marker:      &marker,
		Status:      &status,
		StartAfter:  &startAfter,
		StartBefore: &startBefore,
	})
	if err != nil {
		t.Fatalf("ListRuns() with options error: %v", err)
	}
}

// TestGetRun_EmptyTimerID checks that an empty timer ID returns an error.
func TestGetRun_EmptyTimerID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.GetRun(context.Background(), "", "run-1")
	if err == nil {
		t.Fatal("expected error for empty timerID, got nil")
	}
}

// TestGetRun_EmptyRunID checks that an empty run ID returns an error.
func TestGetRun_EmptyRunID(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	_, err := client.GetRun(context.Background(), "timer-1", "")
	if err == nil {
		t.Fatal("expected error for empty runID, got nil")
	}
}

// TestGetRun_Success verifies retrieving a specific run.
func TestGetRun_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/timers/timer-1/runs/run-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		run := TimerRun{ID: "run-1", TimerID: "timer-1", Status: string(RunStatusSuccess), StartTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(run)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.GetRun(context.Background(), "timer-1", "run-1")
	if err != nil {
		t.Fatalf("GetRun() error: %v", err)
	}
	if result.ID != "run-1" {
		t.Errorf("expected run ID 'run-1', got %q", result.ID)
	}
}

// TestGetCurrentUser_Success verifies retrieving the current user info.
func TestGetCurrentUser_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		name := "Test User"
		user := CurrentUserInfo{ID: "user-1", Username: "testuser", Email: "test@example.com", Name: &name}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	result, err := client.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("GetCurrentUser() error: %v", err)
	}
	if result.ID != "user-1" {
		t.Errorf("expected user ID 'user-1', got %q", result.ID)
	}
	if result.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", result.Username)
	}
}

// TestCreateOnceTimer_Success verifies the helper for one-shot timers.
func TestCreateOnceTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Schedule.Type != string(ScheduleTypeOnce) {
			t.Errorf("expected schedule type 'once', got %q", req.Schedule.Type)
		}
		timer := Timer{ID: "once-1", Name: req.Name, Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	startTime := now.Add(time.Hour)
	callback := Callback{Type: string(CallbackTypeWeb), URL: strPtr("https://example.com")}
	result, err := client.CreateOnceTimer(context.Background(), "Once Timer", startTime, callback, nil)
	if err != nil {
		t.Fatalf("CreateOnceTimer() error: %v", err)
	}
	if result.ID != "once-1" {
		t.Errorf("expected 'once-1', got %q", result.ID)
	}
}

// TestCreateRecurringTimer_Success verifies the helper for recurring timers.
func TestCreateRecurringTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Schedule.Type != string(ScheduleTypeRecurring) {
			t.Errorf("expected schedule type 'recurring', got %q", req.Schedule.Type)
		}
		timer := Timer{ID: "recurring-1", Name: req.Name, Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	startTime := now
	endTime := now.Add(7 * 24 * time.Hour)
	callback := Callback{Type: string(CallbackTypeWeb), URL: strPtr("https://example.com")}
	result, err := client.CreateRecurringTimer(context.Background(), "Recurring Timer", startTime, "P1D", &endTime, callback, nil)
	if err != nil {
		t.Fatalf("CreateRecurringTimer() error: %v", err)
	}
	if result.ID != "recurring-1" {
		t.Errorf("expected 'recurring-1', got %q", result.ID)
	}
}

// TestCreateCronTimer_Success verifies the helper for cron-scheduled timers.
func TestCreateCronTimer_Success(t *testing.T) {
	now := time.Now()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Schedule.Type != string(ScheduleTypeCron) {
			t.Errorf("expected schedule type 'cron', got %q", req.Schedule.Type)
		}
		timer := Timer{ID: "cron-1", Name: req.Name, Status: string(TimerStatusActive), LastUpdate: now, CreateTime: now}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	callback := Callback{Type: string(CallbackTypeWeb), URL: strPtr("https://example.com")}
	result, err := client.CreateCronTimer(context.Background(), "Cron Timer", "0 * * * *", "UTC", nil, callback, nil)
	if err != nil {
		t.Fatalf("CreateCronTimer() error: %v", err)
	}
	if result.ID != "cron-1" {
		t.Errorf("expected 'cron-1', got %q", result.ID)
	}
}

// TestClose verifies Close does not panic.
func TestClose(t *testing.T) {
	client, _ := NewClient(WithAccessToken("tok"))
	// Should not panic
	client.Close()
}

// TestFlowUserScope verifies the scope string format.
func TestFlowUserScope(t *testing.T) {
	flowID := "abc123"
	expected := "https://auth.globus.org/scopes/abc123/flow_abc123_user"
	got := FlowUserScope(flowID)
	if got != expected {
		t.Errorf("FlowUserScope() = %q, want %q", got, expected)
	}
}

// TestCreateFlowCallback verifies the flow callback constructor.
func TestCreateFlowCallback(t *testing.T) {
	flowID := "flow-1"
	flowLabel := "My Flow"
	flowInput := map[string]interface{}{"key": "value"}

	cb := CreateFlowCallback(flowID, flowLabel, flowInput)
	if cb.Type != string(CallbackTypeFlow) {
		t.Errorf("expected type %q, got %q", CallbackTypeFlow, cb.Type)
	}
	if cb.FlowID == nil || *cb.FlowID != flowID {
		t.Errorf("expected FlowID %q, got %v", flowID, cb.FlowID)
	}
	if cb.FlowLabel == nil || *cb.FlowLabel != flowLabel {
		t.Errorf("expected FlowLabel %q, got %v", flowLabel, cb.FlowLabel)
	}
}

// TestCreateWebCallback verifies the web callback constructor.
func TestCreateWebCallback(t *testing.T) {
	rawURL := "https://example.com/hook"
	method := "POST"
	headers := map[string]string{"X-Custom": "value"}
	body := "request body"

	cb := CreateWebCallback(rawURL, method, headers, &body)
	if cb.Type != string(CallbackTypeWeb) {
		t.Errorf("expected type %q, got %q", CallbackTypeWeb, cb.Type)
	}
	if cb.URL == nil || *cb.URL != rawURL {
		t.Errorf("expected URL %q, got %v", rawURL, cb.URL)
	}
	if cb.Method == nil || *cb.Method != method {
		t.Errorf("expected method %q, got %v", method, cb.Method)
	}
	if cb.Body == nil || *cb.Body != body {
		t.Errorf("expected body %q, got %v", body, cb.Body)
	}
}

// TestCreateWebCallback_NilBody verifies the web callback with nil body.
func TestCreateWebCallback_NilBody(t *testing.T) {
	cb := CreateWebCallback("https://example.com", "GET", nil, nil)
	if cb.Body != nil {
		t.Errorf("expected nil body, got %v", cb.Body)
	}
}

// TestDoRequestLowLevel_EmptyResponseBody verifies behavior with empty response body.
func TestDoRequestLowLevel_EmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No body written
	}))
	defer server.Close()

	client, _ := NewClient(WithAccessToken("tok"), WithBaseURL(server.URL+"/"))
	var result map[string]interface{}
	err := client.doRequestLowLevel(context.Background(), http.MethodGet, "test", nil, nil, &result)
	if err != nil {
		t.Fatalf("doRequestLowLevel with empty body error: %v", err)
	}
}

// TestBuildURLLowLevel_WithoutTrailingSlash verifies URL building when base URL has no trailing slash.
func TestBuildURLLowLevel_WithoutTrailingSlash(t *testing.T) {
	client, _ := NewClient(WithBaseURL("https://example.com/api"))
	url := client.buildURLLowLevel("timers", nil)
	expected := "https://example.com/api/timers"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

// TestTimerStatusConstants verifies all timer status constants are defined.
func TestTimerStatusConstants(t *testing.T) {
	statuses := []TimerStatus{
		TimerStatusActive,
		TimerStatusPaused,
		TimerStatusExpired,
		TimerStatusFailed,
		TimerStatusComplete,
	}
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("timer status constant is empty")
		}
	}
}

// TestScheduleTypeConstants verifies all schedule type constants are defined.
func TestScheduleTypeConstants(t *testing.T) {
	types := []ScheduleType{ScheduleTypeOnce, ScheduleTypeRecurring, ScheduleTypeCron}
	for _, s := range types {
		if string(s) == "" {
			t.Errorf("schedule type constant is empty")
		}
	}
}

// TestCallbackTypeConstants verifies all callback type constants are defined.
func TestCallbackTypeConstants(t *testing.T) {
	types := []CallbackType{CallbackTypeFlow, CallbackTypeWeb}
	for _, s := range types {
		if string(s) == "" {
			t.Errorf("callback type constant is empty")
		}
	}
}

// TestRunStatusConstants verifies all run status constants are defined.
func TestRunStatusConstants(t *testing.T) {
	statuses := []RunStatus{RunStatusPending, RunStatusInProgress, RunStatusSuccess, RunStatusFailure}
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("run status constant is empty")
		}
	}
}

// strPtr is a local helper to avoid conflict with the existing one in client_test.go
func strPtr(s string) *string {
	return &s
}

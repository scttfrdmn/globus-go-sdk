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

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

func TestCreateFlowTimerOnce(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/timers" {
			t.Errorf("Expected path /timers, got %s", r.URL.Path)
		}

		// Verify request body
		var req CreateTimerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Verify schedule
		if req.Schedule.Type != string(ScheduleTypeOnce) {
			t.Errorf("Expected schedule type 'once', got %s", req.Schedule.Type)
		}

		// Verify callback
		if req.Callback.Type != string(CallbackTypeFlow) {
			t.Errorf("Expected callback type 'flow', got %s", req.Callback.Type)
		}
		if req.Callback.FlowID == nil || *req.Callback.FlowID != "test-flow-id" {
			t.Errorf("Expected FlowID 'test-flow-id', got %v", req.Callback.FlowID)
		}

		// Verify flow_scope in data
		if scope, ok := req.Data["flow_scope"].(string); !ok || scope != "test-scope" {
			t.Errorf("Expected flow_scope 'test-scope' in data, got %v", req.Data["flow_scope"])
		}

		// Return mock response
		response := Timer{
			ID:     "test-timer-id",
			Name:   req.Name,
			Status: string(TimerStatusActive),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	baseClient := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithAuthorizer(authorizer),
	)
	client := &Client{Client: baseClient}

	// Test FlowTimer creation
	flowTimer := &FlowTimer{
		FlowID:    "test-flow-id",
		FlowScope: "test-scope",
		FlowInput: map[string]interface{}{
			"test_key": "test_value",
		},
		FlowLabel: "Test Flow Run",
	}

	startTime := time.Now().Add(1 * time.Hour)
	timer, err := client.CreateFlowTimerOnce(
		context.Background(),
		"Test Flow Timer",
		startTime,
		flowTimer,
		map[string]interface{}{"custom_data": "value"},
	)

	if err != nil {
		t.Fatalf("CreateFlowTimerOnce() error = %v", err)
	}

	if timer.ID != "test-timer-id" {
		t.Errorf("Expected timer ID 'test-timer-id', got %s", timer.ID)
	}
}

func TestCreateFlowTimerOnce_ValidationErrors(t *testing.T) {
	// Create minimal client (won't make actual requests)
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	baseClient := core.NewClient(
		core.WithBaseURL("http://localhost:9999/"),
		core.WithAuthorizer(authorizer),
	)
	client := &Client{Client: baseClient}

	startTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name      string
		flowTimer *FlowTimer
		wantErr   string
	}{
		{
			name:      "nil FlowTimer",
			flowTimer: nil,
			wantErr:   "flowTimer is required",
		},
		{
			name: "missing FlowID",
			flowTimer: &FlowTimer{
				FlowScope: "test-scope",
			},
			wantErr: "flowTimer.FlowID is required",
		},
		{
			name: "missing FlowScope",
			flowTimer: &FlowTimer{
				FlowID: "test-flow-id",
			},
			wantErr: "flowTimer.FlowScope is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateFlowTimerOnce(
				context.Background(),
				"Test Timer",
				startTime,
				tt.flowTimer,
				nil,
			)
			if err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.wantErr)
			} else if err.Error() != tt.wantErr {
				t.Errorf("Expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCreateFlowTimerRecurring(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify schedule type
		if req.Schedule.Type != string(ScheduleTypeRecurring) {
			t.Errorf("Expected schedule type 'recurring', got %s", req.Schedule.Type)
		}

		// Verify interval
		if req.Schedule.Interval == nil || *req.Schedule.Interval != "P1D" {
			t.Errorf("Expected interval 'P1D', got %v", req.Schedule.Interval)
		}

		// Return mock response
		response := Timer{
			ID:     "test-recurring-timer-id",
			Name:   req.Name,
			Status: string(TimerStatusActive),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	baseClient := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithAuthorizer(authorizer),
	)
	client := &Client{Client: baseClient}

	// Test recurring FlowTimer creation
	flowTimer := &FlowTimer{
		FlowID:    "test-flow-id",
		FlowScope: "test-scope",
	}

	startTime := time.Now().Add(1 * time.Hour)
	timer, err := client.CreateFlowTimerRecurring(
		context.Background(),
		"Test Recurring Flow Timer",
		startTime,
		"P1D", // Daily
		nil,   // No end time
		flowTimer,
		nil,
	)

	if err != nil {
		t.Fatalf("CreateFlowTimerRecurring() error = %v", err)
	}

	if timer.ID != "test-recurring-timer-id" {
		t.Errorf("Expected timer ID 'test-recurring-timer-id', got %s", timer.ID)
	}
}

func TestCreateFlowTimerCron(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)

		// Verify schedule type
		if req.Schedule.Type != string(ScheduleTypeCron) {
			t.Errorf("Expected schedule type 'cron', got %s", req.Schedule.Type)
		}

		// Verify cron expression
		if req.Schedule.CronExpression == nil || *req.Schedule.CronExpression != "0 0 * * *" {
			t.Errorf("Expected cron expression '0 0 * * *', got %v", req.Schedule.CronExpression)
		}

		// Verify timezone
		if req.Schedule.Timezone == nil || *req.Schedule.Timezone != "America/New_York" {
			t.Errorf("Expected timezone 'America/New_York', got %v", req.Schedule.Timezone)
		}

		// Return mock response
		response := Timer{
			ID:     "test-cron-timer-id",
			Name:   req.Name,
			Status: string(TimerStatusActive),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	baseClient := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithAuthorizer(authorizer),
	)
	client := &Client{Client: baseClient}

	// Test cron FlowTimer creation
	flowTimer := &FlowTimer{
		FlowID:    "test-flow-id",
		FlowScope: "test-scope",
		FlowInput: map[string]interface{}{
			"param1": "value1",
		},
	}

	timer, err := client.CreateFlowTimerCron(
		context.Background(),
		"Test Cron Flow Timer",
		"0 0 * * *",        // Daily at midnight
		"America/New_York", // EST/EDT
		nil,                // No end time
		flowTimer,
		nil,
	)

	if err != nil {
		t.Fatalf("CreateFlowTimerCron() error = %v", err)
	}

	if timer.ID != "test-cron-timer-id" {
		t.Errorf("Expected timer ID 'test-cron-timer-id', got %s", timer.ID)
	}
}

func TestFlowTimer_ScopeHandling(t *testing.T) {
	// Mock server that captures the request
	var capturedData map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateTimerRequest
		json.NewDecoder(r.Body).Decode(&req)
		capturedData = req.Data

		response := Timer{ID: "test-id", Name: req.Name, Status: string(TimerStatusActive)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	baseClient := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithAuthorizer(authorizer),
	)
	client := &Client{Client: baseClient}

	flowTimer := &FlowTimer{
		FlowID:    "test-flow-id",
		FlowScope: "test-scope",
	}

	// Test 1: No existing data - should add flow_scope
	startTime := time.Now().Add(1 * time.Hour)
	_, err := client.CreateFlowTimerOnce(context.Background(), "Test", startTime, flowTimer, nil)
	if err != nil {
		t.Fatalf("CreateFlowTimerOnce() error = %v", err)
	}
	if capturedData["flow_scope"] != "test-scope" {
		t.Errorf("Expected flow_scope 'test-scope', got %v", capturedData["flow_scope"])
	}

	// Test 2: Existing data without flow_scope - should add it
	customData := map[string]interface{}{"custom": "value"}
	_, err = client.CreateFlowTimerOnce(context.Background(), "Test", startTime, flowTimer, customData)
	if err != nil {
		t.Fatalf("CreateFlowTimerOnce() error = %v", err)
	}
	if capturedData["flow_scope"] != "test-scope" {
		t.Errorf("Expected flow_scope 'test-scope', got %v", capturedData["flow_scope"])
	}
	if capturedData["custom"] != "value" {
		t.Errorf("Expected custom data 'value', got %v", capturedData["custom"])
	}

	// Test 3: Existing data with flow_scope - should NOT override
	customData = map[string]interface{}{"flow_scope": "custom-scope"}
	_, err = client.CreateFlowTimerOnce(context.Background(), "Test", startTime, flowTimer, customData)
	if err != nil {
		t.Fatalf("CreateFlowTimerOnce() error = %v", err)
	}
	if capturedData["flow_scope"] != "custom-scope" {
		t.Errorf("Expected flow_scope 'custom-scope' (should not override), got %v", capturedData["flow_scope"])
	}
}

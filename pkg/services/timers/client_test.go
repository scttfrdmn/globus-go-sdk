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

func TestCreateTimer(t *testing.T) {
	// Create a test server that returns a mock timer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected method %s, got %s", http.MethodPost, r.Method)
		}
		if r.URL.Path != "/timers" {
			t.Errorf("Expected path /timers, got %s", r.URL.Path)
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var request CreateTimerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check request fields
		if request.Name != "Test Timer" {
			t.Errorf("Expected timer name 'Test Timer', got '%s'", request.Name)
		}
		if request.Schedule.Type != string(ScheduleTypeOnce) {
			t.Errorf("Expected schedule type '%s', got '%s'", ScheduleTypeOnce, request.Schedule.Type)
		}
		if request.Callback.Type != string(CallbackTypeFlow) {
			t.Errorf("Expected callback type '%s', got '%s'", CallbackTypeFlow, request.Callback.Type)
		}

		// Create response
		now := time.Now()
		future := now.Add(24 * time.Hour)
		timer := Timer{
			ID:         "test-timer-id",
			Name:       request.Name,
			Owner:      "test-user",
			Schedule:   &request.Schedule,
			Callback:   &request.Callback,
			Status:     string(TimerStatusActive),
			LastUpdate: now,
			NextDue:    &future,
			CreateTime: now,
			Data:       request.Data,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	// Create client
	client := NewClient("test-token", WithBaseURL(server.URL+"/"))

	// Create a test timer request
	flowID := "test-flow-id"
	flowLabel := "Test Flow"
	startTime := time.Now().Add(1 * time.Hour)
	callback := Callback{
		Type:      string(CallbackTypeFlow),
		FlowID:    &flowID,
		FlowLabel: &flowLabel,
		FlowInput: map[string]interface{}{
			"key": "value",
		},
	}
	request := &CreateTimerRequest{
		Name: "Test Timer",
		Schedule: Schedule{
			Type:      string(ScheduleTypeOnce),
			StartTime: &startTime,
		},
		Callback: callback,
		Data: map[string]interface{}{
			"note": "This is a test timer",
		},
	}

	// Create timer
	timer, err := client.CreateTimer(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	// Check response
	if timer.ID != "test-timer-id" {
		t.Errorf("Expected timer ID 'test-timer-id', got '%s'", timer.ID)
	}
	if timer.Name != "Test Timer" {
		t.Errorf("Expected timer name 'Test Timer', got '%s'", timer.Name)
	}
	if timer.Status != string(TimerStatusActive) {
		t.Errorf("Expected timer status '%s', got '%s'", TimerStatusActive, timer.Status)
	}
}

func TestGetTimer(t *testing.T) {
	// Create a test server that returns a mock timer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected method %s, got %s", http.MethodGet, r.Method)
		}
		if r.URL.Path != "/timers/test-timer-id" {
			t.Errorf("Expected path /timers/test-timer-id, got %s", r.URL.Path)
		}

		// Create response
		now := time.Now()
		future := now.Add(24 * time.Hour)
		flowID := "test-flow-id"
		flowLabel := "Test Flow"
		schedule := Schedule{
			Type:      string(ScheduleTypeOnce),
			StartTime: &future,
		}
		callback := Callback{
			Type:      string(CallbackTypeFlow),
			FlowID:    &flowID,
			FlowLabel: &flowLabel,
			FlowInput: map[string]interface{}{
				"key": "value",
			},
		}
		timer := Timer{
			ID:         "test-timer-id",
			Name:       "Test Timer",
			Owner:      "test-user",
			Schedule:   &schedule,
			Callback:   &callback,
			Status:     string(TimerStatusActive),
			LastUpdate: now,
			NextDue:    &future,
			CreateTime: now,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timer)
	}))
	defer server.Close()

	// Create client
	client := NewClient("test-token", WithBaseURL(server.URL+"/"))

	// Get timer
	timer, err := client.GetTimer(context.Background(), "test-timer-id")
	if err != nil {
		t.Fatalf("Failed to get timer: %v", err)
	}

	// Check response
	if timer.ID != "test-timer-id" {
		t.Errorf("Expected timer ID 'test-timer-id', got '%s'", timer.ID)
	}
	if timer.Name != "Test Timer" {
		t.Errorf("Expected timer name 'Test Timer', got '%s'", timer.Name)
	}
	if timer.Status != string(TimerStatusActive) {
		t.Errorf("Expected timer status '%s', got '%s'", TimerStatusActive, timer.Status)
	}
}

func TestListTimers(t *testing.T) {
	// Create a test server that returns a mock timer list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected method %s, got %s", http.MethodGet, r.Method)
		}
		if r.URL.Path != "/timers" {
			t.Errorf("Expected path /timers, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit 10, got %s", query.Get("limit"))
		}
		if query.Get("status") != string(TimerStatusActive) {
			t.Errorf("Expected status %s, got %s", TimerStatusActive, query.Get("status"))
		}

		// Create response
		now := time.Now()
		future := now.Add(24 * time.Hour)
		flowID := "test-flow-id"
		flowLabel := "Test Flow"
		nextPage := "next-page-token"
		
		// Create two timers
		timer1 := Timer{
			ID:         "test-timer-id-1",
			Name:       "Test Timer 1",
			Owner:      "test-user",
			Schedule: &Schedule{
				Type:      string(ScheduleTypeOnce),
				StartTime: &future,
			},
			Callback: &Callback{
				Type:      string(CallbackTypeFlow),
				FlowID:    &flowID,
				FlowLabel: &flowLabel,
			},
			Status:     string(TimerStatusActive),
			LastUpdate: now,
			NextDue:    &future,
			CreateTime: now,
		}
		
		timer2 := Timer{
			ID:         "test-timer-id-2",
			Name:       "Test Timer 2",
			Owner:      "test-user",
			Schedule: &Schedule{
				Type:      string(ScheduleTypeRecurring),
				StartTime: &now,
			},
			Callback: &Callback{
				Type:   string(CallbackTypeWeb),
				URL:    stringPtr("https://example.com"),
				Method: stringPtr("POST"),
			},
			Status:     string(TimerStatusActive),
			LastUpdate: now,
			NextDue:    &future,
			CreateTime: now,
		}
		
		// Create timer list
		timerList := TimerList{
			Timers:      []Timer{timer1, timer2},
			Total:       2,
			HasNextPage: true,
			NextPage:    &nextPage,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(timerList)
	}))
	defer server.Close()

	// Create client
	client := NewClient("test-token", WithBaseURL(server.URL+"/"))

	// Set up list options
	limit := 10
	status := string(TimerStatusActive)
	options := &ListTimersOptions{
		Limit:  &limit,
		Status: &status,
	}

	// List timers
	list, err := client.ListTimers(context.Background(), options)
	if err != nil {
		t.Fatalf("Failed to list timers: %v", err)
	}

	// Check response
	if list.Total != 2 {
		t.Errorf("Expected 2 timers, got %d", list.Total)
	}
	if !list.HasNextPage {
		t.Errorf("Expected HasNextPage to be true")
	}
	if *list.NextPage != "next-page-token" {
		t.Errorf("Expected NextPage 'next-page-token', got '%s'", *list.NextPage)
	}
	if len(list.Timers) != 2 {
		t.Errorf("Expected 2 timers in list, got %d", len(list.Timers))
	}
	if list.Timers[0].ID != "test-timer-id-1" {
		t.Errorf("Expected first timer ID 'test-timer-id-1', got '%s'", list.Timers[0].ID)
	}
	if list.Timers[1].ID != "test-timer-id-2" {
		t.Errorf("Expected second timer ID 'test-timer-id-2', got '%s'", list.Timers[1].ID)
	}
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}
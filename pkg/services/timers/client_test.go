// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-token")
	client, _ := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	return server, client
}

func TestCreateTimer(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/timer" {
			t.Errorf("%s %s, want POST /v2/timer", r.Method, r.URL.Path)
		}
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["timer"]; !ok {
			t.Error("create body must be wrapped in a \"timer\" key")
		}
		_ = json.NewEncoder(w).Encode(Timer{JobID: "job-1", Name: "T"})
	})
	defer server.Close()

	timer, err := client.CreateTimer(context.Background(),
		NewTransferTimer("T", NewRecurringSchedule(1800, "", nil), map[string]interface{}{"DATA_TYPE": "transfer"}))
	if err != nil {
		t.Fatalf("CreateTimer() error = %v", err)
	}
	if timer.JobID != "job-1" {
		t.Errorf("JobID = %s", timer.JobID)
	}
}

func TestCreateJob(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/jobs/" {
			t.Errorf("%s %s, want POST /jobs/", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Timer{JobID: "job-2"})
	})
	defer server.Close()

	timer, err := client.CreateJob(context.Background(), &TimerJob{
		CallbackURL:  "https://actions.example/run",
		CallbackBody: map[string]interface{}{"k": "v"},
		Start:        "2026-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateJob() error = %v", err)
	}
	if timer.JobID != "job-2" {
		t.Errorf("JobID = %s", timer.JobID)
	}
}

func TestGetTimer(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/job-1" {
			t.Errorf("path = %s, want /jobs/job-1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Timer{JobID: "job-1", Name: "My Timer"})
	})
	defer server.Close()

	timer, err := client.GetTimer(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetTimer() error = %v", err)
	}
	if timer.Name != "My Timer" {
		t.Errorf("Name = %s", timer.Name)
	}

	if _, err := client.GetTimer(context.Background(), ""); err == nil {
		t.Error("expected error for empty timer ID")
	}
}

func TestUpdateTimer(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/jobs/job-1" {
			t.Errorf("%s %s, want PATCH /jobs/job-1", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Timer{JobID: "job-1", Name: "Renamed"})
	})
	defer server.Close()

	timer, err := client.UpdateTimer(context.Background(), "job-1", map[string]interface{}{"name": "Renamed"})
	if err != nil {
		t.Fatalf("UpdateTimer() error = %v", err)
	}
	if timer.Name != "Renamed" {
		t.Errorf("Name = %s", timer.Name)
	}
}

func TestDeleteTimer(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/jobs/job-1" {
			t.Errorf("%s %s, want DELETE /jobs/job-1", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	if err := client.DeleteTimer(context.Background(), "job-1"); err != nil {
		t.Fatalf("DeleteTimer() error = %v", err)
	}
}

func TestListTimers(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/" {
			t.Errorf("path = %s, want /jobs/", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(TimerList{Timers: []Timer{{JobID: "t-1"}}})
	})
	defer server.Close()

	list, err := client.ListTimers(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTimers() error = %v", err)
	}
	if len(list.Timers) != 1 {
		t.Errorf("got %d timers, want 1", len(list.Timers))
	}
}

func TestPauseResumeTimer(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/jobs/job-1/pause":
			w.WriteHeader(http.StatusOK)
		case "/jobs/job-1/resume":
			var body map[string]bool
			_ = json.NewDecoder(r.Body).Decode(&body)
			if !body["update_credentials"] {
				t.Error("expected update_credentials=true in resume body")
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	defer server.Close()

	if err := client.PauseTimer(context.Background(), "job-1"); err != nil {
		t.Fatalf("PauseTimer() error = %v", err)
	}
	yes := true
	if err := client.ResumeTimer(context.Background(), "job-1", &yes); err != nil {
		t.Fatalf("ResumeTimer() error = %v", err)
	}
}

func TestScheduleBuilders(t *testing.T) {
	once := NewOnceSchedule("2026-01-01T00:00:00Z")
	if once.Type != "once" || once.Datetime == "" {
		t.Errorf("once schedule: %+v", once)
	}
	rec := NewRecurringSchedule(3600, "2026-01-01T00:00:00Z", &ScheduleEnd{Condition: "iterations", Iterations: 5})
	if rec.Type != "recurring" || rec.IntervalSeconds != 3600 || rec.End.Iterations != 5 {
		t.Errorf("recurring schedule: %+v", rec)
	}
}

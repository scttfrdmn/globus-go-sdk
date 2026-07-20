// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/timers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"https://auth.globus.org/scopes/524230d7-ea86-4a52-8312-86065a9e0417/timer"}}
		_, err := timers.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetTimer(t *testing.T) {
	t.Run("empty timer ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTimer(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "timerID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &timers.Timer{
			JobID:    "timer-123",
			Name:     "My Timer",
			Schedule: timers.NewOnceSchedule("2026-01-01T00:00:00Z"),
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/jobs/timer-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetTimer(context.Background(), "timer-123")
		require.NoError(t, err)
		assert.Equal(t, "timer-123", result.JobID)
		assert.Equal(t, "My Timer", result.Name)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "timer not found", "NOT_FOUND")
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetTimer(context.Background(), "missing")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestCreateTimer(t *testing.T) {
	t.Run("nil timer returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateTimer(context.Background(), nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("success posts wrapped document to /v2/timer", func(t *testing.T) {
		timer := timers.NewTransferTimer(
			"Test Timer",
			timers.NewRecurringSchedule(1800, "2026-01-01T00:00:00Z", nil),
			map[string]interface{}{"DATA_TYPE": "transfer"},
		)
		expected := &timers.Timer{JobID: "new-timer", Name: "Test Timer"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/timer", r.URL.Path)

			var body map[string]json.RawMessage
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			_, hasTimerEnvelope := body["timer"]
			assert.True(t, hasTimerEnvelope, "create body must be wrapped in a \"timer\" key")

			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateTimer(context.Background(), timer)
		require.NoError(t, err)
		assert.Equal(t, "new-timer", result.JobID)
	})
}

func TestCreateJob(t *testing.T) {
	t.Run("nil job returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateJob(context.Background(), nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})

	t.Run("success posts to /jobs/", func(t *testing.T) {
		expected := &timers.Timer{JobID: "job-1"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/jobs/", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateJob(context.Background(), &timers.TimerJob{
			CallbackURL:  "https://actions.example/run",
			CallbackBody: map[string]interface{}{"k": "v"},
			Start:        "2026-01-01T00:00:00Z",
		})
		require.NoError(t, err)
		assert.Equal(t, "job-1", result.JobID)
	})
}

func TestUpdateTimer(t *testing.T) {
	t.Run("success uses PATCH /jobs/{id}", func(t *testing.T) {
		expected := &timers.Timer{JobID: "timer-123", Name: "Renamed"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/jobs/timer-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.UpdateTimer(context.Background(), "timer-123", map[string]interface{}{"name": "Renamed"})
		require.NoError(t, err)
		assert.Equal(t, "Renamed", result.Name)
	})
}

func TestDeleteTimer(t *testing.T) {
	t.Run("empty timer ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteTimer(context.Background(), "")
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("success", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/jobs/timer-123", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.DeleteTimer(context.Background(), "timer-123")
		assert.NoError(t, err)
	})
}

func TestPauseResumeTimer(t *testing.T) {
	t.Run("pause empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.PauseTimer(context.Background(), "")
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})

	t.Run("resume empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.ResumeTimer(context.Background(), "", nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})

	t.Run("pause success", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/jobs/timer-123/pause", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.PauseTimer(context.Background(), "timer-123")
		assert.NoError(t, err)
	})

	t.Run("resume sends update_credentials when set", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/jobs/timer-123/resume", r.URL.Path)
			var body map[string]bool
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, true, body["update_credentials"])
			w.WriteHeader(http.StatusOK)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		yes := true
		err = client.ResumeTimer(context.Background(), "timer-123", &yes)
		assert.NoError(t, err)
	})
}

func TestListTimers(t *testing.T) {
	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &timers.TimerList{Timers: []timers.Timer{{JobID: "t-1", Name: "Timer 1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/jobs/", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListTimers(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Timers, 1)
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

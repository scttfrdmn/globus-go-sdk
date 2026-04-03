// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timers_test

import (
	"context"
	"net/http"
	"testing"
	"time"

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
			ID:   "timer-123",
			Name: "My Timer",
			Schedule: &timers.Schedule{
				Type:      "once",
				StartTime: time.Now().Add(time.Hour),
			},
			Callback: &timers.Callback{Type: "https", URL: "https://example.com"},
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "timer-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetTimer(context.Background(), "timer-123")
		require.NoError(t, err)
		assert.Equal(t, "timer-123", result.ID)
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

	t.Run("success", func(t *testing.T) {
		timer := &timers.Timer{
			Name:     "Test Timer",
			Schedule: &timers.Schedule{Type: "once", StartTime: time.Now().Add(time.Hour)},
			Callback: &timers.Callback{Type: "https", URL: "https://example.com/cb"},
		}
		expected := &timers.Timer{ID: "new-timer", Name: "Test Timer"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateTimer(context.Background(), timer)
		require.NoError(t, err)
		assert.Equal(t, "new-timer", result.ID)
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
			assert.Contains(t, r.URL.Path, "timer-123")
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

		err = client.ResumeTimer(context.Background(), "")
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})

	t.Run("pause success", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Path, "timer-123")
			assert.Contains(t, r.URL.Path, "pause")
			w.WriteHeader(http.StatusOK)
		})
		client, err := timers.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.PauseTimer(context.Background(), "timer-123")
		assert.NoError(t, err)
	})
}

func TestListTimers(t *testing.T) {
	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &timers.TimerList{Timers: []timers.Timer{{ID: "t-1", Name: "Timer 1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.Query().Get("limit"))
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

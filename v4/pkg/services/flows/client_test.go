// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"https://auth.globus.org/scopes/eec9b274-0c81-4334-bdc2-54e90e689b9a/manage_flows"}}
		_, err := flows.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetFlow(t *testing.T) {
	t.Run("empty flow ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetFlow(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "flowID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &flows.Flow{ID: "flow-123", Title: "My Flow"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "flow-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetFlow(context.Background(), "flow-123")
		require.NoError(t, err)
		assert.Equal(t, "flow-123", result.ID)
		assert.Equal(t, "My Flow", result.Title)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "flow not found", "NOT_FOUND")
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetFlow(context.Background(), "missing-flow")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestRunFlow(t *testing.T) {
	t.Run("empty flow ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.RunFlow(context.Background(), "", &flows.FlowInput{})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "flowID", valErr.Field)
	})

	t.Run("nil input returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.RunFlow(context.Background(), "flow-123", nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("success", func(t *testing.T) {
		expected := &flows.FlowRun{RunID: "run-123", FlowID: "flow-123", Status: "ACTIVE"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "flow-123")
			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.RunFlow(context.Background(), "flow-123",
			&flows.FlowInput{Input: map[string]interface{}{"key": "value"}})
		require.NoError(t, err)
		assert.Equal(t, "run-123", result.RunID)
		assert.Equal(t, "ACTIVE", result.Status)
	})
}

func TestListRuns(t *testing.T) {
	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &flows.RunList{Runs: []flows.FlowRun{{RunID: "run-1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListRuns(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Runs, 1)
	})

	t.Run("limit and offset passed as query params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "5", r.URL.Query().Get("limit"))
			assert.Equal(t, "10", r.URL.Query().Get("offset"))
			testhelpers.RespondJSON(w, http.StatusOK, &flows.RunList{})
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListRuns(context.Background(), &flows.ListRunsOptions{Limit: 5, Offset: 10})
		assert.NoError(t, err)
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

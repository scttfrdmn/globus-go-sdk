// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"https://auth.globus.org/scopes/facd7ccc-c5f4-42aa-916b-a0e270e2c2a9/all"}}
		_, err := compute.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetEndpoint(t *testing.T) {
	t.Run("empty endpoint ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetEndpoint(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "endpointID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &compute.Endpoint{ID: "ep-123", Name: "Test Endpoint", Status: "online"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "ep-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetEndpoint(context.Background(), "ep-123")
		require.NoError(t, err)
		assert.Equal(t, "ep-123", result.ID)
		assert.Equal(t, "online", result.Status)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "endpoint not found", "NOT_FOUND")
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetEndpoint(context.Background(), "missing-ep")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestSubmitFunction(t *testing.T) {
	t.Run("empty endpoint ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitFunction(context.Background(), "", &compute.FunctionSubmission{FunctionID: "fn-1"})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "endpointID", valErr.Field)
	})

	t.Run("nil submission returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitFunction(context.Background(), "ep-123", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "submission", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &compute.FunctionRun{ID: "run-123", Status: "running"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "ep-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.SubmitFunction(context.Background(), "ep-123",
			&compute.FunctionSubmission{FunctionID: "fn-1"})
		require.NoError(t, err)
		assert.Equal(t, "run-123", result.ID)
	})
}

func TestListFunctions(t *testing.T) {
	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &compute.FunctionList{Functions: []compute.FunctionRun{{ID: "run-1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListFunctions(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Functions, 1)
	})

	t.Run("limit and offset are passed as query params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "10", r.URL.Query().Get("limit"))
			assert.Equal(t, "20", r.URL.Query().Get("offset"))
			testhelpers.RespondJSON(w, http.StatusOK, &compute.FunctionList{})
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListFunctions(context.Background(), &compute.ListFunctionsOptions{Limit: 10, Offset: 20})
		assert.NoError(t, err)
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

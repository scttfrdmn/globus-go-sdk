// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute_test

import (
	"context"
	"encoding/json"
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

	t.Run("success hits /v2/endpoints/{id}", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/endpoints/ep-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"uuid": "ep-123", "status": "online"})
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetEndpoint(context.Background(), "ep-123")
		require.NoError(t, err)
		assert.Equal(t, "ep-123", result["uuid"])
		assert.Equal(t, "online", result["status"])
	})
}

func TestGetEndpoints(t *testing.T) {
	t.Run("role param sent, no limit/offset", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/endpoints", r.URL.Path)
			assert.Equal(t, "owner", r.URL.Query().Get("role"))
			assert.Empty(t, r.URL.Query().Get("limit"))
			// /v2/endpoints returns a top-level JSON array.
			testhelpers.RespondJSON(w, http.StatusOK, []map[string]interface{}{
				{"uuid": "ep-1"},
				{"uuid": "ep-2"},
			})
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		eps, err := client.GetEndpoints(context.Background(), &compute.GetEndpointsOptions{Role: "owner"})
		require.NoError(t, err)
		require.Len(t, eps, 2)
		assert.Equal(t, "ep-1", eps[0]["uuid"])
	})
}

func TestSubmit(t *testing.T) {
	t.Run("nil data returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Submit(context.Background(), nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "data", valErr.Field)
	})

	t.Run("v2 submit hits /v2/submit", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/submit", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"request_id": "r1"})
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		res, err := client.Submit(context.Background(), map[string]interface{}{"tasks": []interface{}{}})
		require.NoError(t, err)
		assert.Equal(t, "r1", res["request_id"])
	})

	t.Run("v3 submit hits /v3/endpoints/{id}/submit", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v3/endpoints/ep-1/submit", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"request_id": "r2"})
		})
		client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		res, err := client.SubmitV3(context.Background(), "ep-1", map[string]interface{}{"tasks": []interface{}{}})
		require.NoError(t, err)
		assert.Equal(t, "r2", res["request_id"])
	})
}

func TestGetTaskBatchBody(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/batch_status", r.URL.Path)
		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		ids, ok := body["task_ids"].([]interface{})
		require.True(t, ok)
		assert.Len(t, ids, 2)
		testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"results": map[string]interface{}{}})
	})
	client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	_, err = client.GetTaskBatch(context.Background(), []string{"t1", "t2"})
	require.NoError(t, err)
}

func TestGetVersionServiceParam(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/version", r.URL.Path)
		assert.Equal(t, "web", r.URL.Query().Get("service"))
		testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"version": "1.0"})
	})
	client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	res, err := client.GetVersion(context.Background(), &compute.GetVersionOptions{Service: "web"})
	require.NoError(t, err)
	// With a service, /v2/version returns a JSON object.
	obj, ok := res.(map[string]interface{})
	require.True(t, ok, "GetVersion should return an object when a service is given")
	assert.Equal(t, "1.0", obj["version"])
}

func TestGetVersionScalarResponse(t *testing.T) {
	// With no service, /v2/version returns a bare JSON string.
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/version", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("service"))
		testhelpers.RespondJSON(w, http.StatusOK, "2.34.0")
	})
	client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	res, err := client.GetVersion(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "2.34.0", res)
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := compute.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

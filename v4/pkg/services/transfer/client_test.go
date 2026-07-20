// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"urn:globus:auth:scope:transfer.api.globus.org:all"}}
		_, err := transfer.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetEndpoint(t *testing.T) {
	t.Run("empty endpoint ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetEndpoint(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "endpointID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &transfer.Endpoint{ID: "ep-abc", DisplayName: "My Endpoint"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "ep-abc")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetEndpoint(context.Background(), "ep-abc")
		require.NoError(t, err)
		assert.Equal(t, "ep-abc", result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "endpoint not found", "NOT_FOUND")
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetEndpoint(context.Background(), "missing")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestSubmitTransfer(t *testing.T) {
	t.Run("nil transfer returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitTransfer(context.Background(), nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("missing source endpoint returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitTransfer(context.Background(), &transfer.Transfer{
			DestinationEndpoint: "dst-ep",
			Items:               []transfer.TransferItem{{SourcePath: "/src", DestinationPath: "/dst"}},
		})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "SourceEndpoint", valErr.Field)
	})

	t.Run("missing destination endpoint returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitTransfer(context.Background(), &transfer.Transfer{
			SourceEndpoint: "src-ep",
			Items:          []transfer.TransferItem{{SourcePath: "/src", DestinationPath: "/dst"}},
		})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "DestinationEndpoint", valErr.Field)
	})

	t.Run("empty items returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.SubmitTransfer(context.Background(), &transfer.Transfer{
			SourceEndpoint:      "src-ep",
			DestinationEndpoint: "dst-ep",
		})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "Items", valErr.Field)
	})

	t.Run("success auto-fetches submission_id and posts to /v0.10/transfer", func(t *testing.T) {
		expected := &transfer.TaskSubmitResponse{TaskID: "task-123", Code: "Accepted"}
		var submittedID string
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/v0.10/submission_id":
				assert.Equal(t, http.MethodGet, r.Method)
				testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"DATA_TYPE": "submission_id", "value": "sub-xyz"})
			case "/v0.10/transfer":
				assert.Equal(t, http.MethodPost, r.Method)
				var body map[string]interface{}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				submittedID, _ = body["submission_id"].(string)
				testhelpers.RespondJSON(w, http.StatusAccepted, expected)
			default:
				t.Errorf("unexpected path %s", r.URL.Path)
			}
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.SubmitTransfer(context.Background(), &transfer.Transfer{
			SourceEndpoint:      "src-ep",
			DestinationEndpoint: "dst-ep",
			Items:               []transfer.TransferItem{{SourcePath: "/src", DestinationPath: "/dst"}},
		})
		require.NoError(t, err)
		assert.Equal(t, "task-123", result.TaskID)
		assert.Equal(t, "sub-xyz", submittedID, "submission_id should be auto-fetched and sent")
	})
}

func TestListTasks(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		expected := &transfer.TaskList{Data: []transfer.Task{{TaskID: "t-1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Empty(t, r.URL.Query().Get("filter_status"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListTasks(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Data, 1)
	})

	t.Run("filter_status comma-joined into a single param", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v0.10/task_list", r.URL.Path)
			statuses := r.URL.Query()["filter_status"]
			require.Len(t, statuses, 1)
			assert.Equal(t, "ACTIVE,INACTIVE", statuses[0])
			testhelpers.RespondJSON(w, http.StatusOK, &transfer.TaskList{})
		})
		client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListTasks(context.Background(), &transfer.ListTasksOptions{
			FilterStatus: []string{"ACTIVE", "INACTIVE"},
		})
		assert.NoError(t, err)
	})
}

func TestEndpointSearch(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v0.10/endpoint_search", r.URL.Path)
		assert.Equal(t, "tutorial", r.URL.Query().Get("filter_fulltext"))
		testhelpers.RespondJSON(w, http.StatusOK, &transfer.EndpointSearchResult{
			Data: []transfer.Endpoint{{ID: "ep-1"}},
		})
	})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	res, err := client.EndpointSearch(context.Background(), &transfer.EndpointSearchOptions{FilterFulltext: "tutorial"})
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	assert.Equal(t, "ep-1", res.Data[0].ID)
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := transfer.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

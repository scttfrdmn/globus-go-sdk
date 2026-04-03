// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"urn:globus:auth:scope:search.api.globus.org:all"}}
		_, err := search.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetIndex(t *testing.T) {
	t.Run("empty index ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetIndex(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "indexID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &search.Index{ID: "idx-123", DisplayName: "My Index", Status: "open"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "idx-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetIndex(context.Background(), "idx-123")
		require.NoError(t, err)
		assert.Equal(t, "idx-123", result.ID)
		assert.Equal(t, "My Index", result.DisplayName)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "index not found", "NOT_FOUND")
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetIndex(context.Background(), "missing-idx")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestSearch(t *testing.T) {
	t.Run("empty index ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Search(context.Background(), "", &search.SearchQuery{Q: "test"})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok)
		assert.Equal(t, "indexID", valErr.Field)
	})

	t.Run("nil query returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.Search(context.Background(), "idx-123", nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("success", func(t *testing.T) {
		expected := &search.SearchResults{Count: 1, Total: 1}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "idx-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.Search(context.Background(), "idx-123", &search.SearchQuery{Q: "hello"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Total)
	})
}

// TestIndexList covers issue #22 — SearchClient.IndexList() added in upstream v4.5.0.
func TestIndexList(t *testing.T) {
	t.Run("nil options — no query params", func(t *testing.T) {
		expected := &search.IndexList{Indexes: []search.Index{{ID: "idx-1"}, {ID: "idx-2"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/index_list", r.URL.Path)
			assert.Empty(t, r.URL.Query().Get("filter_roles"))
			assert.Empty(t, r.URL.Query().Get("limit"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.IndexList(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Indexes, 2)
	})

	t.Run("single filter_roles param", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			roles := r.URL.Query()["filter_roles"]
			require.Len(t, roles, 1)
			assert.Equal(t, "owner", roles[0])
			testhelpers.RespondJSON(w, http.StatusOK, &search.IndexList{})
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.IndexList(context.Background(), &search.ListIndexesOptions{
			FilterRoles: []string{"owner"},
		})
		assert.NoError(t, err)
	})

	t.Run("multiple filter_roles params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			roles := r.URL.Query()["filter_roles"]
			assert.ElementsMatch(t, []string{"owner", "writer"}, roles)
			testhelpers.RespondJSON(w, http.StatusOK, &search.IndexList{})
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.IndexList(context.Background(), &search.ListIndexesOptions{
			FilterRoles: []string{"owner", "writer"},
		})
		assert.NoError(t, err)
	})

	t.Run("limit and offset params", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "25", r.URL.Query().Get("limit"))
			assert.Equal(t, "50", r.URL.Query().Get("offset"))
			testhelpers.RespondJSON(w, http.StatusOK, &search.IndexList{})
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.IndexList(context.Background(), &search.ListIndexesOptions{
			Limit: 25, Offset: 50,
		})
		assert.NoError(t, err)
	})

	t.Run("forbidden returns APIError", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusForbidden, "permission denied", "FORBIDDEN")
		})
		client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.IndexList(context.Background(), nil)
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
		assert.True(t, apiErr.IsAuthError())
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := search.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close())
}

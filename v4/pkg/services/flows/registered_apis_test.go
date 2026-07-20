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

// Tests for the Flows registered APIs — upstream v4.6.0 (GET /registered_apis)
// and v4.7.0 (per_page query param).

func TestGetRegisteredAPI(t *testing.T) {
	t.Run("empty ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetRegisteredAPI(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "registeredAPIID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &flows.RegisteredAPI{
			ID:     "api-123",
			Name:   "registered-api-1",
			Status: "ACTIVE",
			Roles:  &flows.RegisteredAPIRoles{Owners: []string{"urn:globus:auth:identity:owner"}},
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/registered_apis/api-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetRegisteredAPI(context.Background(), "api-123")
		require.NoError(t, err)
		assert.Equal(t, "api-123", result.ID)
		assert.Equal(t, "registered-api-1", result.Name)
		assert.Equal(t, "ACTIVE", result.Status)
		require.NotNil(t, result.Roles)
		assert.Equal(t, []string{"urn:globus:auth:identity:owner"}, result.Roles.Owners)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "registered API not found", "NOT_FOUND")
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetRegisteredAPI(context.Background(), "missing")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok, "expected APIError, got %T", err)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	})
}

func TestListRegisteredAPIs(t *testing.T) {
	t.Run("success with nil options", func(t *testing.T) {
		expected := &flows.RegisteredAPIList{
			RegisteredAPIs: []flows.RegisteredAPI{{ID: "api-1"}, {ID: "api-2"}},
			HasNextPage:    false,
		}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/registered_apis", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListRegisteredAPIs(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.RegisteredAPIs, 2)
		assert.False(t, result.HasNextPage)
	})

	t.Run("query params are set including comma-joined filter_roles", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			assert.Equal(t, "owner,administrator", q.Get("filter_roles"))
			assert.Equal(t, "name ASC", q.Get("orderby"))
			assert.Equal(t, "50", q.Get("per_page"))
			assert.Equal(t, "fake_marker_0", q.Get("marker"))
			testhelpers.RespondJSON(w, http.StatusOK, &flows.RegisteredAPIList{})
		})
		client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListRegisteredAPIs(context.Background(), &flows.ListRegisteredAPIsOptions{
			FilterRoles: []string{"owner", "administrator"},
			OrderBy:     "name ASC",
			PerPage:     50,
			Marker:      "fake_marker_0",
		})
		assert.NoError(t, err)
	})
}

func TestNewRegisteredAPIsPager(t *testing.T) {
	// Three pages driven by marker/has_next_page, mirroring the upstream
	// list_registered_apis fixture (markers fake_marker_0, fake_marker_1, then done).
	pages := map[string]*flows.RegisteredAPIList{
		"": {
			RegisteredAPIs: []flows.RegisteredAPI{{ID: "api-1"}},
			HasNextPage:    true,
			Marker:         "fake_marker_0",
		},
		"fake_marker_0": {
			RegisteredAPIs: []flows.RegisteredAPI{{ID: "api-2"}},
			HasNextPage:    true,
			Marker:         "fake_marker_1",
		},
		"fake_marker_1": {
			RegisteredAPIs: []flows.RegisteredAPI{{ID: "api-3"}},
			HasNextPage:    false,
		},
	}

	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		marker := r.URL.Query().Get("marker")
		page, ok := pages[marker]
		require.True(t, ok, "unexpected marker %q", marker)
		testhelpers.RespondJSON(w, http.StatusOK, page)
	})
	client, err := flows.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	pager := client.NewRegisteredAPIsPager(nil)
	var ids []string
	for pager.HasNext() {
		items, err := pager.NextPage(context.Background())
		require.NoError(t, err)
		for _, a := range items {
			ids = append(ids, a.ID)
		}
	}
	assert.Equal(t, []string{"api-1", "api-2", "api-3"}, ids)
}

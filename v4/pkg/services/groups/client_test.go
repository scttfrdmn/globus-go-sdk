// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"urn:globus:auth:scope:groups.api.globus.org:all"}}
		_, err := groups.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetGroup(t *testing.T) {
	t.Run("empty group ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetGroup(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "groupID", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &groups.Group{ID: "grp-123", Name: "Test Group"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "grp-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetGroup(context.Background(), "grp-123")
		require.NoError(t, err)
		assert.Equal(t, "grp-123", result.ID)
		assert.Equal(t, "Test Group", result.Name)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "group not found", "NOT_FOUND")
		})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetGroup(context.Background(), "missing-grp")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestCreateGroup(t *testing.T) {
	t.Run("nil group returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateGroup(context.Background(), nil)
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("missing name returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.CreateGroup(context.Background(), &groups.GroupCreate{})
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "Name", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &groups.Group{ID: "new-grp", Name: "New Group"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			testhelpers.RespondJSON(w, http.StatusCreated, expected)
		})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.CreateGroup(context.Background(), &groups.GroupCreate{Name: "New Group"})
		require.NoError(t, err)
		assert.Equal(t, "new-grp", result.ID)
	})
}

func TestListGroups(t *testing.T) {
	t.Run("nil options succeeds", func(t *testing.T) {
		expected := &groups.GroupList{Groups: []groups.Group{{ID: "grp-1"}}}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.ListGroups(context.Background(), nil)
		require.NoError(t, err)
		assert.Len(t, result.Groups, 1)
	})

	t.Run("my groups filter", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "true", r.URL.Query().Get("my_groups"))
			testhelpers.RespondJSON(w, http.StatusOK, &groups.GroupList{})
		})
		client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.ListGroups(context.Background(), &groups.ListGroupsOptions{MyGroups: true})
		assert.NoError(t, err)
	})
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close()) // idempotent
}

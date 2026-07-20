// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// groupsTestConfig points at the test server but keeps the "/v2" suffix the real
// Groups base URL carries, so tests exercise the base-path join.
func groupsTestConfig(serverURL string) *core.Config {
	cfg := testhelpers.NewTestConfig(serverURL)
	cfg.BaseURL = serverURL + "/v2"
	return cfg
}

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

		_, err = client.GetGroup(context.Background(), "", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "groupID", valErr.Field)
	})

	t.Run("success with include comma-joined", func(t *testing.T) {
		expected := &groups.Group{ID: "grp-123", Name: "Test Group"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/groups/grp-123", r.URL.Path)
			assert.Equal(t, "memberships,policies", r.URL.Query().Get("include"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := groups.NewClient(context.Background(), groupsTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetGroup(context.Background(), "grp-123", &groups.GetGroupOptions{Include: []string{"memberships", "policies"}})
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

		_, err = client.GetGroup(context.Background(), "missing-grp", nil)
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

func TestGetMyGroups(t *testing.T) {
	t.Run("hits my_groups and returns top-level array", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/groups/my_groups", r.URL.Path)
			assert.Equal(t, "active,invited", r.URL.Query().Get("statuses"))
			testhelpers.RespondJSON(w, http.StatusOK, []groups.Group{{ID: "grp-1"}, {ID: "grp-2"}})
		})
		client, err := groups.NewClient(context.Background(), groupsTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetMyGroups(context.Background(), []string{"active", "invited"})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "grp-1", result[0].ID)
	})
}

func TestBatchMembershipAction(t *testing.T) {
	t.Run("posts action document to /groups/{id}", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/groups/grp-1", r.URL.Path)
			var body map[string]json.RawMessage
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			_, hasAdd := body["add"]
			assert.True(t, hasAdd, "expected an \"add\" action key")
			testhelpers.RespondJSON(w, http.StatusOK, &groups.Group{ID: "grp-1"})
		})
		client, err := groups.NewClient(context.Background(), groupsTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.BatchMembershipAction(context.Background(), "grp-1", &groups.BatchMembershipActions{
			Add: []groups.MemberWithRole{{IdentityID: "id-1", Role: groups.RoleMember}},
		})
		require.NoError(t, err)
	})
}

func TestSetSubscriptionAdminVerified(t *testing.T) {
	t.Run("nil subscription sends JSON null", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "/v2/groups/grp-1/subscription_admin_verified", r.URL.Path)
			var body map[string]interface{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			v, present := body["subscription_admin_verified_id"]
			assert.True(t, present)
			assert.Nil(t, v)
			w.WriteHeader(http.StatusOK)
		})
		client, err := groups.NewClient(context.Background(), groupsTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.SetSubscriptionAdminVerified(context.Background(), "grp-1", nil)
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

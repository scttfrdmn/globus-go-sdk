// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// authTestConfig points the client at the test server but keeps the "/v2" path
// suffix that the real Auth base URL carries, so tests exercise the base-path
// join (upstream paths like /v2/oauth2/... must survive).
func authTestConfig(serverURL string) *core.Config {
	cfg := testhelpers.NewTestConfig(serverURL)
	cfg.BaseURL = serverURL + "/v2"
	return cfg
}

func TestNewClient(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		config := testhelpers.NewTestConfig(server.URL)
		client, err := auth.NewClient(context.Background(), config)
		require.NoError(t, err)
		require.NotNil(t, client)
		_ = client.Close()
	})

	t.Run("missing access token", func(t *testing.T) {
		config := &core.Config{Scopes: []string{"openid"}}
		_, err := auth.NewClient(context.Background(), config)
		assert.Error(t, err)
	})
}

func TestGetUserInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &auth.UserInfo{Sub: "user-123", Email: "user@example.com"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/oauth2/userinfo", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetUserInfo(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "user-123", result.Sub)
		assert.Equal(t, "user@example.com", result.Email)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusUnauthorized, "invalid token", "UNAUTHORIZED")
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetUserInfo(context.Background())
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok, "expected APIError, got %T", err)
		assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
		assert.True(t, apiErr.IsAuthError())
	})
}

func TestIntrospectToken(t *testing.T) {
	t.Run("empty token returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.IntrospectToken(context.Background(), "", nil)
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "token", valErr.Field)
	})

	t.Run("success form-encoded with include", func(t *testing.T) {
		expected := &auth.TokenIntrospection{Active: true, Sub: "user-123"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v2/oauth2/token/introspect", r.URL.Path)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "some-token", r.PostForm.Get("token"))
			assert.Equal(t, "identity_set", r.PostForm.Get("include"))
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.IntrospectToken(context.Background(), "some-token", &auth.IntrospectOptions{Include: "identity_set"})
		require.NoError(t, err)
		assert.True(t, result.Active)
		assert.Equal(t, "user-123", result.Sub)
	})
}

func TestRevokeToken(t *testing.T) {
	t.Run("empty token returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.RevokeToken(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "token", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusOK)
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		err = client.RevokeToken(context.Background(), "some-token")
		assert.NoError(t, err)
	})
}

func TestGetProject(t *testing.T) {
	t.Run("empty project ID returns validation error", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetProject(context.Background(), "")
		require.Error(t, err)
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok, "expected ValidationError, got %T", err)
	})

	t.Run("success unwraps project envelope", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v2/api/projects/proj-123", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{
				"project": map[string]interface{}{"id": "proj-123", "display_name": "My Project"},
			})
		})
		client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetProject(context.Background(), "proj-123")
		require.NoError(t, err)
		assert.Equal(t, "proj-123", result.ID)
		assert.Equal(t, "My Project", result.DisplayName)
	})

	t.Run("not found", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			testhelpers.RespondError(w, http.StatusNotFound, "project not found", "NOT_FOUND")
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetProject(context.Background(), "missing-proj")
		require.Error(t, err)
		apiErr, ok := err.(*core.APIError)
		require.True(t, ok)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestGetIdentities(t *testing.T) {
	t.Run("usernames comma-joined with provision", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/api/identities", r.URL.Path)
			assert.Equal(t, "alice@x.org,bob@x.org", r.URL.Query().Get("usernames"))
			assert.Equal(t, "true", r.URL.Query().Get("provision"))
			testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{
				"identities": []map[string]interface{}{{"id": "id-1", "username": "alice@x.org"}},
			})
		})
		client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		ids, err := client.GetIdentities(context.Background(), &auth.GetIdentitiesOptions{
			Usernames: []string{"alice@x.org", "bob@x.org"},
			Provision: true,
		})
		require.NoError(t, err)
		require.Len(t, ids, 1)
		assert.Equal(t, "id-1", ids[0].ID)
	})

	t.Run("usernames and ids are mutually exclusive", func(t *testing.T) {
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
		client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetIdentities(context.Background(), &auth.GetIdentitiesOptions{
			Usernames: []string{"a"}, IDs: []string{"b"},
		})
		_, ok := err.(*core.ValidationError)
		assert.True(t, ok)
	})
}

func TestCreatePolicyEnvelope(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/api/policies", r.URL.Path)
		var body map[string]json.RawMessage
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_, wrapped := body["policy"]
		assert.True(t, wrapped, "create body must be wrapped in a \"policy\" key")
		testhelpers.RespondJSON(w, http.StatusCreated, map[string]interface{}{
			"policy": map[string]interface{}{"id": "pol-1", "display_name": "P"},
		})
	})
	client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	pol, err := client.CreatePolicy(context.Background(), &auth.PolicyCreate{
		ProjectID: "proj", DisplayName: "P", Description: "d",
	})
	require.NoError(t, err)
	assert.Equal(t, "pol-1", pol.ID)
}

func TestGetDependentTokens(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/oauth2/token", r.URL.Path)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "urn:globus:auth:grant_type:dependent_token", r.PostForm.Get("grant_type"))
		assert.Equal(t, "offline", r.PostForm.Get("access_type"))
		testhelpers.RespondJSON(w, http.StatusOK, []map[string]interface{}{
			{"access_token": "at1", "resource_server": "rs1", "expires_in": 3600},
		})
	})
	client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	toks, err := client.GetDependentTokens(context.Background(), "primary", &auth.DependentTokensOptions{RefreshTokens: true})
	require.NoError(t, err)
	require.Len(t, toks, 1)
	assert.Equal(t, "rs1", toks[0].ResourceServer)
}

func TestGetOpenIDConfigurationHitsRoot(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Must hit the host root, NOT under /v2.
		assert.Equal(t, "/.well-known/openid-configuration", r.URL.Path)
		testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{"issuer": "https://auth.globus.org"})
	})
	client, err := auth.NewClient(context.Background(), authTestConfig(server.URL))
	require.NoError(t, err)
	defer client.Close()

	doc, err := client.GetOpenIDConfiguration(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://auth.globus.org", doc["issuer"])
}

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close()) // idempotent
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

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

		_, err = client.IntrospectToken(context.Background(), "")
		require.Error(t, err)
		valErr, ok := err.(*core.ValidationError)
		require.True(t, ok, "expected ValidationError, got %T", err)
		assert.Equal(t, "token", valErr.Field)
	})

	t.Run("success", func(t *testing.T) {
		expected := &auth.TokenIntrospection{Active: true, Sub: "user-123"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/oauth2/token/introspect", r.URL.Path)
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.IntrospectToken(context.Background(), "some-token")
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

	t.Run("success", func(t *testing.T) {
		expected := &auth.Project{ID: "proj-123", DisplayName: "My Project"}
		server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "proj-123")
			testhelpers.RespondJSON(w, http.StatusOK, expected)
		})
		client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
		require.NoError(t, err)
		defer client.Close()

		result, err := client.GetProject(context.Background(), "proj-123")
		require.NoError(t, err)
		assert.Equal(t, "proj-123", result.ID)
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

func TestClose(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)

	assert.NoError(t, client.Close())
	assert.NoError(t, client.Close()) // idempotent
}

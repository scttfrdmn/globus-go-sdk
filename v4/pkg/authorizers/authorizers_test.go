// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package authorizers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessTokenAuthorizer(t *testing.T) {
	ctx := context.Background()

	t.Run("returns bearer header", func(t *testing.T) {
		a := authorizers.NewAccessTokenAuthorizer("my-token")
		header, err := a.GetAuthorizationHeader(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Bearer my-token", header)
	})

	t.Run("empty token returns error", func(t *testing.T) {
		a := authorizers.NewAccessTokenAuthorizer("")
		_, err := a.GetAuthorizationHeader(ctx)
		require.Error(t, err)
		var ve *core.ValidationError
		assert.ErrorAs(t, err, &ve)
	})

	t.Run("HandleMissingAuthorization always false", func(t *testing.T) {
		a := authorizers.NewAccessTokenAuthorizer("my-token")
		assert.False(t, a.HandleMissingAuthorization(ctx))
	})
}

func TestRefreshTokenAuthorizer(t *testing.T) {
	ctx := context.Background()

	t.Run("uses initial access token when not expired", func(t *testing.T) {
		a := authorizers.NewRefreshTokenAuthorizer("refresh-tok", "client-id", "",
			authorizers.WithInitialAccessToken("access-tok", time.Now().Add(time.Hour)),
		)
		header, err := a.GetAuthorizationHeader(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Bearer access-tok", header)
	})

	t.Run("refreshes when token is expired", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/oauth2/token", r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "refresh_token", r.FormValue("grant_type"))
			assert.Equal(t, "my-refresh-tok", r.FormValue("refresh_token"))
			assert.Equal(t, "my-client-id", r.FormValue("client_id"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "new-access-tok",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		var refreshed bool
		a := authorizers.NewRefreshTokenAuthorizer("my-refresh-tok", "my-client-id", "",
			authorizers.WithInitialAccessToken("old-tok", time.Now().Add(-time.Minute)), // already expired
			authorizers.WithAuthBaseURL(server.URL),
			authorizers.WithHTTPClient(server.Client()),
			authorizers.WithOnRefresh(func(_, _ string, _ time.Time) { refreshed = true }),
		)

		header, err := a.GetAuthorizationHeader(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Bearer new-access-tok", header)
		assert.True(t, refreshed)
	})
}

func TestClientCredentialsAuthorizer(t *testing.T) {
	ctx := context.Background()

	t.Run("fetches token via client credentials", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v2/oauth2/token", r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
			assert.Equal(t, "my-client-id", r.FormValue("client_id"))
			assert.Equal(t, "my-client-secret", r.FormValue("client_secret"))
			assert.Equal(t, "scope1 scope2", r.FormValue("scope"))

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "cc-access-tok",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		a := authorizers.NewClientCredentialsAuthorizer(
			"my-client-id", "my-client-secret",
			[]string{"scope1", "scope2"},
			authorizers.WithClientCredentialsAuthBaseURL(server.URL),
			authorizers.WithClientCredentialsHTTPClient(server.Client()),
		)

		header, err := a.GetAuthorizationHeader(ctx)
		require.NoError(t, err)
		assert.Equal(t, "Bearer cc-access-tok", header)
	})

	t.Run("HandleMissingAuthorization refetches token", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "cc-tok",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		a := authorizers.NewClientCredentialsAuthorizer(
			"id", "secret", nil,
			authorizers.WithClientCredentialsAuthBaseURL(server.URL),
			authorizers.WithClientCredentialsHTTPClient(server.Client()),
		)

		// First call fetches
		_, err := a.GetAuthorizationHeader(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		// HandleMissingAuthorization forces a re-fetch
		ok := a.HandleMissingAuthorization(ctx)
		assert.True(t, ok)
		assert.Equal(t, 2, callCount)
	})
}

func TestAuthorizerIntegrationWithCoreConfig(t *testing.T) {
	// Verify that core.Config accepts an Authorizer instead of AccessToken
	a := authorizers.NewAccessTokenAuthorizer("my-token")
	config := &core.Config{
		Authorizer: a,
		Scopes:     []string{"openid"},
	}
	err := config.Validate()
	assert.NoError(t, err)
}

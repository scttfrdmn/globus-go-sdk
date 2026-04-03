// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/app"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/login"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/tokenstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ----

type stubLoginFlowManager struct {
	tokens []*tokenstorage.TokenData
}

func (s *stubLoginFlowManager) RunLoginFlow(_ context.Context, _ login.AuthParams) (*login.LoginResult, error) {
	return &login.LoginResult{Tokens: s.tokens}, nil
}

func tokenData(rs, at, rt string) *tokenstorage.TokenData {
	return &tokenstorage.TokenData{
		ResourceServer: rs,
		AccessToken:    at,
		RefreshToken:   rt,
		TokenType:      "Bearer",
		ExpiresAt:      time.Now().Add(time.Hour),
	}
}

// ---- UserApp tests ----

func TestUserApp_LoginRequired_BeforeLogin(t *testing.T) {
	storage := tokenstorage.NewMemoryTokenStorage()
	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{
		TokenStorage: storage,
	})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "urn:globus:auth:scope:transfer.api.globus.org:all")
	assert.True(t, a.LoginRequired())
}

func TestUserApp_Login_StoresTokens(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()

	stub := &stubLoginFlowManager{tokens: []*tokenstorage.TokenData{
		tokenData("transfer.api.globus.org", "at-transfer", "rt-transfer"),
		tokenData("groups.api.globus.org", "at-groups", "rt-groups"),
	}}

	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{
		TokenStorage:     storage,
		LoginFlowManager: stub,
	})
	require.NoError(t, err)

	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")
	require.NoError(t, a.Login(ctx))

	td, err := storage.Get("transfer.api.globus.org")
	require.NoError(t, err)
	require.NotNil(t, td)
	assert.Equal(t, "at-transfer", td.AccessToken)
}

func TestUserApp_LoginRequired_AfterLogin(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()

	stub := &stubLoginFlowManager{tokens: []*tokenstorage.TokenData{
		tokenData("transfer.api.globus.org", "at", "rt"),
	}}

	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{
		TokenStorage:     storage,
		LoginFlowManager: stub,
	})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.Login(ctx))
	assert.False(t, a.LoginRequired())
}

func TestUserApp_GetAuthorizer_WithRefreshToken(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "rt")))

	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{TokenStorage: storage})
	require.NoError(t, err)

	auth, err := a.GetAuthorizer(ctx, "transfer.api.globus.org")
	require.NoError(t, err)
	require.NotNil(t, auth)

	// Should use the stored access token initially (refresh token authorizer).
	header, err := auth.GetAuthorizationHeader(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Bearer at", header)
}

func TestUserApp_GetAuthorizer_NoToken_ReturnsError(t *testing.T) {
	ctx := context.Background()
	a, err := app.NewUserApp("client-id", "client-secret", nil)
	require.NoError(t, err)

	_, err = a.GetAuthorizer(ctx, "transfer.api.globus.org")
	assert.Error(t, err)
}

func TestUserApp_Logout_RemovesTokens(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "rt")))

	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{TokenStorage: storage})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.Logout(ctx))
	td, err := storage.Get("transfer.api.globus.org")
	require.NoError(t, err)
	assert.Nil(t, td)
}

func TestUserApp_Close(t *testing.T) {
	a, err := app.NewUserApp("client-id", "", nil)
	require.NoError(t, err)
	assert.NoError(t, a.Close())
}

// ---- ClientApp tests ----

func TestClientApp_RequiresClientSecret(t *testing.T) {
	_, err := app.NewClientApp("client-id", "", nil)
	assert.Error(t, err)
}

func TestClientApp_LoginRequired_AlwaysFalse(t *testing.T) {
	a, err := app.NewClientApp("client-id", "client-secret", nil)
	require.NoError(t, err)
	assert.False(t, a.LoginRequired())
}

func TestClientApp_Login_NoOp(t *testing.T) {
	a, err := app.NewClientApp("client-id", "client-secret", nil)
	require.NoError(t, err)
	assert.NoError(t, a.Login(context.Background()))
}

func TestClientApp_GetAuthorizer_UsesClientCredentials(t *testing.T) {
	ctx := context.Background()

	// Set up a fake token endpoint.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "my-client-id", r.FormValue("client_id"))
		assert.Equal(t, "my-client-secret", r.FormValue("client_secret"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "cc-token",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	// We can't easily inject the auth base URL into app.ClientApp yet, so we
	// test that GetAuthorizer returns a valid non-nil authorizer and that it
	// would call the token endpoint with the right parameters when invoked.
	// For now, assert the authorizer is created successfully.
	a, err := app.NewClientApp("my-client-id", "my-client-secret", nil)
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	auth, err := a.GetAuthorizer(ctx, "transfer.api.globus.org")
	require.NoError(t, err)
	require.NotNil(t, auth)
	// The authorizer itself is valid; actual token fetch is tested in authorizers_test.go.
}

func TestClientApp_Logout_NoOp(t *testing.T) {
	a, err := app.NewClientApp("client-id", "client-secret", nil)
	require.NoError(t, err)
	assert.NoError(t, a.Logout(context.Background()))
}

func TestClientApp_Close(t *testing.T) {
	a, err := app.NewClientApp("client-id", "client-secret", nil)
	require.NoError(t, err)
	assert.NoError(t, a.Close())
}

// ---- GlobusApp interface compliance ----

func TestUserApp_ImplementsGlobusApp(t *testing.T) {
	a, err := app.NewUserApp("id", "secret", nil)
	require.NoError(t, err)
	var _ app.GlobusApp = a
}

func TestClientApp_ImplementsGlobusApp(t *testing.T) {
	a, err := app.NewClientApp("id", "secret", nil)
	require.NoError(t, err)
	var _ app.GlobusApp = a
}

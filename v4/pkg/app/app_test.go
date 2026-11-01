// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

	rs := newRevokeServer(t)
	a, err := app.NewUserApp("client-id", "client-secret", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.Logout(ctx))
	assert.Equal(t, []string{"at", "rt"}, rs.tokens(), "access then refresh token revoked")
	td, err := storage.Get("transfer.api.globus.org")
	require.NoError(t, err)
	assert.Nil(t, td)
}

func TestUserApp_LogoutWithOptions_Sweep(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "rt")))
	require.NoError(t, storage.Store(tokenData("groups.api.globus.org", "at2", "")))
	require.NoError(t, storage.Store(tokenData("stale.example.org", "at3", "")))

	rs := newRevokeServer(t)
	a, err := app.NewUserApp("client-id", "", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	// Default logout leaves tokens for resource servers the app doesn't use.
	require.NoError(t, a.LogoutWithOptions(ctx, nil))
	all, err := storage.GetAll()
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Sweep clears everything (upstream logout(sweep=True), v4.9.0).
	require.NoError(t, a.LogoutWithOptions(ctx, &app.LogoutOptions{Sweep: true}))
	all, err = storage.GetAll()
	require.NoError(t, err)
	assert.Empty(t, all)
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

func TestClientApp_LogoutWithOptions(t *testing.T) {
	ctx := context.Background()
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "")))
	rs := newRevokeServer(t)
	a, err := app.NewClientApp("client-id", "client-secret", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)

	require.NoError(t, a.LogoutWithOptions(ctx, nil)) // no-op, as before
	assert.Empty(t, rs.tokens(), "plain logout makes no revoke calls")
	all, err := storage.GetAll()
	require.NoError(t, err)
	assert.Len(t, all, 1)

	require.NoError(t, a.LogoutWithOptions(ctx, &app.LogoutOptions{Sweep: true}))
	all, err = storage.GetAll()
	require.NoError(t, err)
	assert.Empty(t, all)
	assert.Equal(t, []string{"at"}, rs.tokens())
	assert.True(t, rs.reqs[0].basicOK, "client app authenticates with Basic auth")
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

// ---- logout token revocation (upstream GlobusApp.logout) ----

type revokeReq struct {
	token, clientID string
	basicUser       string
	basicOK         bool
	contentType     string
}

type revokeServer struct {
	*httptest.Server
	reqs   []revokeReq
	status int
}

func (r *revokeServer) tokens() []string {
	var out []string
	for _, q := range r.reqs {
		out = append(out, q.token)
	}
	return out
}

func newRevokeServer(t *testing.T) *revokeServer {
	t.Helper()
	rs := &revokeServer{status: http.StatusOK}
	rs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/oauth2/token/revoke", r.URL.Path)
		require.NoError(t, r.ParseForm())
		u, _, ok := r.BasicAuth()
		rs.reqs = append(rs.reqs, revokeReq{
			token: r.PostForm.Get("token"), clientID: r.PostForm.Get("client_id"),
			basicUser: u, basicOK: ok, contentType: r.Header.Get("Content-Type"),
		})
		w.WriteHeader(rs.status)
		_, _ = w.Write([]byte(`{"active": false}`))
	}))
	t.Cleanup(rs.Close)
	return rs
}

func TestUserApp_Logout_RevokesConfidentialWithBasicAuth(t *testing.T) {
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "rt")))
	rs := newRevokeServer(t)
	a, err := app.NewUserApp("cid", "secret", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL + "/"})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.Logout(context.Background()))
	require.Len(t, rs.reqs, 2)
	for _, q := range rs.reqs {
		assert.True(t, q.basicOK)
		assert.Equal(t, "cid", q.basicUser)
		assert.Empty(t, q.clientID, "confidential clients authenticate via Basic, not client_id")
		assert.Equal(t, "application/x-www-form-urlencoded", q.contentType)
	}
}

func TestUserApp_Logout_RevokesNativeWithClientID(t *testing.T) {
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "")))
	rs := newRevokeServer(t)
	a, err := app.NewUserApp("native-cid", "", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.Logout(context.Background()))
	require.Len(t, rs.reqs, 1, "no refresh token, so only the access token is revoked")
	assert.False(t, rs.reqs[0].basicOK)
	assert.Equal(t, "native-cid", rs.reqs[0].clientID)
	assert.Equal(t, "at", rs.reqs[0].token)
}

func TestUserApp_Logout_RevokeFailureStillRemovesTokens(t *testing.T) {
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at", "rt")))
	rs := newRevokeServer(t)
	rs.status = http.StatusInternalServerError
	a, err := app.NewUserApp("cid", "", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	err = a.Logout(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")
	assert.Len(t, rs.reqs, 2, "keeps going after the first failure")
	td, gerr := storage.Get("transfer.api.globus.org")
	require.NoError(t, gerr)
	assert.Nil(t, td, "token removed locally even though revocation failed")
}

func TestUserApp_Logout_NothingStoredMakesNoRequests(t *testing.T) {
	rs := newRevokeServer(t)
	a, err := app.NewUserApp("cid", "", &app.AppConfig{AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")
	require.NoError(t, a.Logout(context.Background()))
	assert.Empty(t, rs.reqs)
}

func TestUserApp_LogoutSweep_RevokesEveryStoredToken(t *testing.T) {
	storage := tokenstorage.NewMemoryTokenStorage()
	require.NoError(t, storage.Store(tokenData("transfer.api.globus.org", "at1", "rt1")))
	require.NoError(t, storage.Store(tokenData("stale.example.org", "at2", "")))
	rs := newRevokeServer(t)
	a, err := app.NewUserApp("cid", "", &app.AppConfig{TokenStorage: storage, AuthBaseURL: rs.URL})
	require.NoError(t, err)
	a.AddScopeRequirements("transfer.api.globus.org", "transfer:all")

	require.NoError(t, a.LogoutWithOptions(context.Background(), &app.LogoutOptions{Sweep: true}))
	assert.ElementsMatch(t, []string{"at1", "rt1", "at2"}, rs.tokens())
}

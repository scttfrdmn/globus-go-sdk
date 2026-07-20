// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

// setupTestClient creates an auth client pointed at a test server
func setupTestClient(t *testing.T, server *httptest.Server) *auth.Client {
	t.Helper()
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithClientSecret("test-client-secret"),
		auth.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	return client
}

// ----- V2 method tests -----

func TestExchangeAuthorizationCodeV2_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResp := auth.TokenResponse{
			AccessToken:  "v2-access-token",
			RefreshToken: "v2-refresh-token",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			Scope:        "openid profile email",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp)
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	client.RedirectURL = "https://example.com/callback"

	resp, err := client.ExchangeAuthorizationCodeV2(context.Background(), "test-code")
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCodeV2() error = %v", err)
	}
	if resp == nil {
		t.Fatal("ExchangeAuthorizationCodeV2() returned nil response")
	}
	if resp.Data.AccessToken != "v2-access-token" {
		t.Errorf("Expected AccessToken=v2-access-token, got %s", resp.Data.AccessToken)
	}
}

func TestExchangeAuthorizationCodeV2_Error(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithClientSecret("test-client-secret"),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	// RedirectURL is empty, should get error
	_, err = client.ExchangeAuthorizationCodeV2(context.Background(), "test-code")
	if err == nil {
		t.Error("Expected error for missing redirect URL, got nil")
	}
}

func TestRefreshTokenV2_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResp := auth.TokenResponse{
			AccessToken: "refreshed-access-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp)
	}))
	defer server.Close()

	client := setupTestClient(t, server)

	resp, err := client.RefreshTokenV2(context.Background(), "old-refresh-token")
	if err != nil {
		t.Fatalf("RefreshTokenV2() error = %v", err)
	}
	if resp == nil {
		t.Fatal("RefreshTokenV2() returned nil response")
	}
	if resp.Data.AccessToken != "refreshed-access-token" {
		t.Errorf("Expected AccessToken=refreshed-access-token, got %s", resp.Data.AccessToken)
	}
}

func TestRefreshTokenV2_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	_, err := client.RefreshTokenV2(context.Background(), "bad-refresh-token")
	if err == nil {
		t.Error("Expected error for unauthorized, got nil")
	}
}

func TestIntrospectTokenV2_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := auth.TokenInfo{
			Active:   true,
			ClientID: "test-client-id",
			Username: "testuser",
			Exp:      time.Now().Add(time.Hour).Unix(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}))
	defer server.Close()

	client := setupTestClient(t, server)

	resp, err := client.IntrospectTokenV2(context.Background(), "some-token")
	if err != nil {
		t.Fatalf("IntrospectTokenV2() error = %v", err)
	}
	if resp == nil {
		t.Fatal("IntrospectTokenV2() returned nil response")
	}
	if !resp.Data.Active {
		t.Error("Expected Active=true")
	}
}

func TestIntrospectTokenV2_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	_, err := client.IntrospectTokenV2(context.Background(), "some-token")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetClientCredentialsTokenV2_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResp := auth.TokenResponse{
			AccessToken: "cc-access-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
			Scope:       "scope1",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp)
	}))
	defer server.Close()

	client := setupTestClient(t, server)

	resp, err := client.GetClientCredentialsTokenV2(context.Background(), "scope1")
	if err != nil {
		t.Fatalf("GetClientCredentialsTokenV2() error = %v", err)
	}
	if resp == nil {
		t.Fatal("GetClientCredentialsTokenV2() returned nil response")
	}
	if resp.Data.AccessToken != "cc-access-token" {
		t.Errorf("Expected AccessToken=cc-access-token, got %s", resp.Data.AccessToken)
	}
}

func TestGetClientCredentialsTokenV2_Error(t *testing.T) {
	// No client secret set — should fail before making HTTP call
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	_, err = client.GetClientCredentialsTokenV2(context.Background())
	if err == nil {
		t.Error("Expected error for missing client secret, got nil")
	}
}

// ----- Error case coverage for existing methods -----

func TestIntrospectToken_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_token","error_description":"Token is invalid"}`))
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	_, err := client.IntrospectToken(context.Background(), "bad-token")
	if err == nil {
		t.Error("Expected error for bad status response, got nil")
	}
}

func TestRevokeToken_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	err := client.RevokeToken(context.Background(), "bad-token")
	if err == nil {
		t.Error("Expected error for bad status response, got nil")
	}
}

func TestRevokeToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token/revoke" {
			t.Errorf("Expected path /oauth2/token/revoke, got %s", r.URL.Path)
		}
		r.ParseForm()
		if r.Form.Get("token") != "valid-token" {
			t.Errorf("Expected token=valid-token, got %s", r.Form.Get("token"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := setupTestClient(t, server)
	err := client.RevokeToken(context.Background(), "valid-token")
	if err != nil {
		t.Errorf("RevokeToken() unexpected error = %v", err)
	}
}

// ----- Options tests -----

func TestWithBaseURL(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithBaseURL("https://custom.auth.example.com/v2/"),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client with WithBaseURL: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestWithHTTPDebugging(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithHTTPDebugging(true),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client with WithHTTPDebugging: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestWithHTTPTracing(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client with WithHTTPTracing: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

func TestWithAuthorizer(t *testing.T) {
	// Create a simple mock authorizer
	mockAuth := &mockAuthorizer{token: "static-test-token"}

	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithAuthorizer(mockAuth),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client with WithAuthorizer: %v", err)
	}
	if client == nil {
		t.Error("Expected non-nil client")
	}
}

// mockAuthorizer implements interfaces.Authorizer
type mockAuthorizer struct {
	token string
}

func (m *mockAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	return "Bearer " + m.token, nil
}

func (m *mockAuthorizer) IsValid() bool {
	return m.token != ""
}

func (m *mockAuthorizer) GetToken() string {
	return m.token
}

var _ interfaces.Authorizer = (*mockAuthorizer)(nil)

// ----- Adapter tests -----

func TestAuthorizerAdapter(t *testing.T) {
	mockAuth := &mockAuthorizer{token: "adapter-test-token"}
	adapter := auth.NewAuthorizerAdapter(mockAuth)

	if adapter == nil {
		t.Fatal("NewAuthorizerAdapter returned nil")
	}

	// Test GetAuthorizationHeader with a context
	header, err := adapter.GetAuthorizationHeader(context.Background())
	if err != nil {
		t.Fatalf("GetAuthorizationHeader() error = %v", err)
	}
	if header != "Bearer adapter-test-token" {
		t.Errorf("Expected 'Bearer adapter-test-token', got %s", header)
	}

	// Test GetAuthorizationHeader with nil context (should use background)
	header, err = adapter.GetAuthorizationHeader()
	if err != nil {
		t.Fatalf("GetAuthorizationHeader() with no ctx error = %v", err)
	}
	if header != "Bearer adapter-test-token" {
		t.Errorf("Expected 'Bearer adapter-test-token', got %s", header)
	}

	// Test IsValid
	if !adapter.IsValid() {
		t.Error("Expected IsValid() = true")
	}

	// Test GetToken
	if adapter.GetToken() != "adapter-test-token" {
		t.Errorf("Expected GetToken() = adapter-test-token, got %s", adapter.GetToken())
	}
}

func TestAuthorizerAdapter_InvalidToken(t *testing.T) {
	mockAuth := &mockAuthorizer{token: ""}
	adapter := auth.NewAuthorizerAdapter(mockAuth)

	if adapter.IsValid() {
		t.Error("Expected IsValid() = false for empty token")
	}
}

// ----- Error helpers tests -----

func TestAuthError_Error(t *testing.T) {
	errWithDesc := &auth.AuthError{
		Code:        "invalid_grant",
		Description: "Token expired",
	}
	expected := "invalid_grant: Token expired"
	if errWithDesc.Error() != expected {
		t.Errorf("AuthError.Error() = %q, want %q", errWithDesc.Error(), expected)
	}

	errWithoutDesc := &auth.AuthError{
		Code: "server_error",
	}
	if errWithoutDesc.Error() != "server_error" {
		t.Errorf("AuthError.Error() without description = %q, want %q",
			errWithoutDesc.Error(), "server_error")
	}
}

func TestGlobusAuthRequirementsError_Error(t *testing.T) {
	errWithDesc := &auth.GlobusAuthRequirementsError{
		Code:        "consent_required",
		Description: "Consent needed",
	}
	if errWithDesc.Error() == "" {
		t.Error("GlobusAuthRequirementsError.Error() returned empty string")
	}

	errWithoutDesc := &auth.GlobusAuthRequirementsError{
		Code: "consent_required",
	}
	if errWithoutDesc.Error() == "" {
		t.Error("GlobusAuthRequirementsError.Error() without desc returned empty string")
	}
}

func TestIsInvalidClient(t *testing.T) {
	// Test with AuthError
	authErr := &auth.AuthError{Code: "invalid_client"}
	if !auth.IsInvalidClient(authErr) {
		t.Error("Expected IsInvalidClient(AuthError{invalid_client}) = true")
	}

	// Test with wrapped sentinel error
	if auth.IsInvalidClient(nil) {
		t.Error("Expected IsInvalidClient(nil) = false")
	}
}

func TestIsInvalidScope(t *testing.T) {
	authErr := &auth.AuthError{Code: "invalid_scope"}
	if !auth.IsInvalidScope(authErr) {
		t.Error("Expected IsInvalidScope(AuthError{invalid_scope}) = true")
	}

	if auth.IsInvalidScope(nil) {
		t.Error("Expected IsInvalidScope(nil) = false")
	}
}

func TestIsAccessDenied(t *testing.T) {
	authErr := &auth.AuthError{Code: "access_denied"}
	if !auth.IsAccessDenied(authErr) {
		t.Error("Expected IsAccessDenied(AuthError{access_denied}) = true")
	}

	if auth.IsAccessDenied(nil) {
		t.Error("Expected IsAccessDenied(nil) = false")
	}
}

func TestIsServerError(t *testing.T) {
	authErr := &auth.AuthError{Code: "server_error"}
	if !auth.IsServerError(authErr) {
		t.Error("Expected IsServerError(AuthError{server_error}) = true")
	}

	if auth.IsServerError(nil) {
		t.Error("Expected IsServerError(nil) = false")
	}
}

func TestIsUnauthorized(t *testing.T) {
	authErr := &auth.AuthError{
		Code:       "unauthorized",
		StatusCode: http.StatusUnauthorized,
	}
	if !auth.IsUnauthorized(authErr) {
		t.Error("Expected IsUnauthorized(AuthError{401}) = true")
	}

	if auth.IsUnauthorized(nil) {
		t.Error("Expected IsUnauthorized(nil) = false")
	}
}

func TestIsBadRequest(t *testing.T) {
	authErr := &auth.AuthError{
		Code:       "bad_request",
		StatusCode: http.StatusBadRequest,
	}
	if !auth.IsBadRequest(authErr) {
		t.Error("Expected IsBadRequest(AuthError{400}) = true")
	}

	if auth.IsBadRequest(nil) {
		t.Error("Expected IsBadRequest(nil) = false")
	}
}

func TestIsGlobusAuthRequirementsError(t *testing.T) {
	gareErr := &auth.GlobusAuthRequirementsError{Code: "consent_required"}
	if !auth.IsGlobusAuthRequirementsError(gareErr) {
		t.Error("Expected IsGlobusAuthRequirementsError = true")
	}

	if auth.IsGlobusAuthRequirementsError(nil) {
		t.Error("Expected IsGlobusAuthRequirementsError(nil) = false")
	}
}

func TestIsConsentRequired(t *testing.T) {
	gareErr := &auth.GlobusAuthRequirementsError{Code: "consent_required"}
	if !auth.IsConsentRequired(gareErr) {
		t.Error("Expected IsConsentRequired(GARE{consent_required}) = true")
	}

	authErr := &auth.AuthError{Code: "consent_required"}
	if !auth.IsConsentRequired(authErr) {
		t.Error("Expected IsConsentRequired(AuthError{consent_required}) = true")
	}

	if auth.IsConsentRequired(nil) {
		t.Error("Expected IsConsentRequired(nil) = false")
	}
}

func TestIsDependentConsentRequired(t *testing.T) {
	gareErr := &auth.GlobusAuthRequirementsError{Code: "dependent_consent_required"}
	if !auth.IsDependentConsentRequired(gareErr) {
		t.Error("Expected IsDependentConsentRequired(GARE{dependent_consent_required}) = true")
	}

	authErr := &auth.AuthError{Code: "dependent_consent_required"}
	if !auth.IsDependentConsentRequired(authErr) {
		t.Error("Expected IsDependentConsentRequired(AuthError{dependent_consent_required}) = true")
	}

	if auth.IsDependentConsentRequired(nil) {
		t.Error("Expected IsDependentConsentRequired(nil) = false")
	}
}

// ----- Model method tests -----

func TestTokenResponse_ExpiresAt(t *testing.T) {
	now := time.Now()
	expiry := now.Add(time.Hour)
	tr := &auth.TokenResponse{
		AccessToken: "test",
		ExpiryTime:  expiry,
	}
	got := tr.ExpiresAt()
	if !got.Equal(expiry) {
		t.Errorf("TokenResponse.ExpiresAt() = %v, want %v", got, expiry)
	}
}

// ----- GetUserInfo tests -----

func TestGetUserInfo_Success(t *testing.T) {
	userInfo := auth.UserInfo{
		Sub:               "user-123",
		Name:              "Test User",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
		EmailVerified:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/userinfo" {
			t.Errorf("Expected path /oauth2/userinfo, got %s", r.URL.Path)
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-access-token" {
			t.Errorf("Expected Authorization=Bearer test-access-token, got %s", authHeader)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer server.Close()

	// GetUserInfo uses http.DefaultClient directly, so we need to override the default transport
	// to route to our test server. We create the client with the test server base URL.
	client := setupTestClient(t, server)

	// GetUserInfo uses http.DefaultClient, which will make an actual request to the client's base URL.
	// We need to ensure it hits our test server. Since the function uses http.DefaultClient.Do,
	// we temporarily replace the default transport.
	origTransport := http.DefaultTransport
	http.DefaultTransport = &localRedirectTransport{
		serverURL: server.URL,
		base:      http.DefaultTransport,
	}
	defer func() { http.DefaultTransport = origTransport }()

	info, err := client.GetUserInfo(context.Background(), "test-access-token")
	if err != nil {
		t.Fatalf("GetUserInfo() error = %v", err)
	}
	if info.Sub != "user-123" {
		t.Errorf("Expected Sub=user-123, got %s", info.Sub)
	}
	if info.Name != "Test User" {
		t.Errorf("Expected Name=Test User, got %s", info.Name)
	}
	if info.Email != "test@example.com" {
		t.Errorf("Expected Email=test@example.com, got %s", info.Email)
	}
}

func TestGetUserInfo_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := setupTestClient(t, server)

	origTransport := http.DefaultTransport
	http.DefaultTransport = &localRedirectTransport{
		serverURL: server.URL,
		base:      http.DefaultTransport,
	}
	defer func() { http.DefaultTransport = origTransport }()

	_, err := client.GetUserInfo(context.Background(), "bad-token")
	if err == nil {
		t.Error("Expected error for 401 status, got nil")
	}
}

// localRedirectTransport intercepts HTTP requests and rewrites the host
// to point to the test server. GetUserInfo uses http.DefaultClient directly,
// so we replace the default transport temporarily.
type localRedirectTransport struct {
	serverURL string
	base      http.RoundTripper
}

func (t *localRedirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request and rewrite the host to point to our test server
	req2 := req.Clone(req.Context())
	newURL := *req.URL
	req2.URL = &newURL
	req2.URL.Scheme = "http"
	// serverURL is like "http://127.0.0.1:PORT", extract the host:port
	serverHost := t.serverURL
	if len(serverHost) > 7 && serverHost[:7] == "http://" {
		req2.URL.Host = serverHost[7:]
	}
	req2.Host = req2.URL.Host
	return t.base.RoundTrip(req2)
}

// ----- MFA method tests -----

func TestExchangeAuthorizationCodeWithMFA_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenResp := auth.TokenResponse{
				AccessToken:  "mfa-access-token",
				RefreshToken: "mfa-refresh-token",
				ExpiresIn:    3600,
				TokenType:    "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResp)
		}
	}))
	defer server.Close()

	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithClientSecret("test-client-secret"),
		auth.WithRedirectURL("https://example.com/callback"),
		auth.WithCoreOption(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tokenResp, err := client.ExchangeAuthorizationCodeWithMFA(
		context.Background(),
		"test-code",
		func(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
			return &auth.MFAResponse{
				ChallengeID: challenge.ChallengeID,
				Type:        "totp",
				Value:       "123456",
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCodeWithMFA() error = %v", err)
	}
	if tokenResp.AccessToken != "mfa-access-token" {
		t.Errorf("Expected AccessToken=mfa-access-token, got %s", tokenResp.AccessToken)
	}
}

func TestExchangeAuthorizationCodeWithMFA_NoRedirectURL(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("test-client-id"),
		auth.WithClientSecret("test-client-secret"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ExchangeAuthorizationCodeWithMFA(context.Background(), "code", nil)
	if err == nil {
		t.Error("Expected error for missing redirect URL, got nil")
	}
}

func TestRefreshTokenWithMFA_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			tokenResp := auth.TokenResponse{
				AccessToken: "refreshed-via-mfa",
				ExpiresIn:   3600,
				TokenType:   "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tokenResp)
		}
	}))
	defer server.Close()

	client := setupTestClient(t, server)

	tokenResp, err := client.RefreshTokenWithMFA(
		context.Background(),
		"test-refresh-token",
		func(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
			return &auth.MFAResponse{
				ChallengeID: "challenge-id",
				Type:        "totp",
				Value:       "654321",
			}, nil
		},
	)
	if err != nil {
		t.Fatalf("RefreshTokenWithMFA() error = %v", err)
	}
	if tokenResp.AccessToken != "refreshed-via-mfa" {
		t.Errorf("Expected AccessToken=refreshed-via-mfa, got %s", tokenResp.AccessToken)
	}
}
func TestGetMFAChallenge_Func_NilError(t *testing.T) {
	// auth.GetMFAChallenge (standalone func) should return nil for nil error
	challenge := auth.GetMFAChallenge(nil)
	if challenge != nil {
		t.Error("Expected GetMFAChallenge(nil) = nil")
	}
}

// ----- GetAuthorizationURL edge cases -----

func TestGetAuthorizationURL_MultipleScopes(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("client-123"),
		auth.WithRedirectURL("https://callback.example.com/"),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	url := client.GetAuthorizationURL("state-xyz", "openid", "profile", "email", "urn:globus:auth:scope:transfer.api.globus.org:all")
	if url == "" {
		t.Error("Expected non-empty URL")
	}
	// Should contain all scopes
	if !containsStr(url, "openid") {
		t.Errorf("URL missing 'openid' scope: %s", url)
	}
	if !containsStr(url, "client_id=client-123") {
		t.Errorf("URL missing client_id: %s", url)
	}
	if !containsStr(url, "state=state-xyz") {
		t.Errorf("URL missing state: %s", url)
	}
}

func TestGetAuthorizationURL_DefaultScopes(t *testing.T) {
	client, err := auth.NewClient(
		auth.WithClientID("client-abc"),
		auth.WithRedirectURL("https://callback.example.com/"),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

	// No scopes provided — should use default
	url := client.GetAuthorizationURL("state-abc")
	if url == "" {
		t.Error("Expected non-empty URL")
	}
	// Default scope should include openid
	if !containsStr(url, "openid") {
		t.Errorf("URL missing default 'openid' scope: %s", url)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())
}

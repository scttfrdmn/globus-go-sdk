// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Test mock server
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)

	// Create a client that uses the test server
	client := NewClient("test-client-id", "test-client-secret",
		core.WithBaseURL(server.URL+"/"),
	)

	return server, client
}

func TestGetAuthorizationURL(t *testing.T) {
	client := NewClient("test-client-id", "test-client-secret")
	client.SetRedirectURL("https://example.com/callback")

	// Test with default scope
	url := client.GetAuthorizationURL("test-state")
	expected := "https://auth.globus.org/v2/oauth2/authorize?client_id=test-client-id&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&response_type=code&scope=openid+profile+email&state=test-state"

	if url != expected {
		t.Errorf("GetAuthorizationURL() = %v, want %v", url, expected)
	}

	// Test with custom scopes
	url = client.GetAuthorizationURL("test-state", "custom-scope1", "custom-scope2")
	expected = "https://auth.globus.org/v2/oauth2/authorize?client_id=test-client-id&redirect_uri=https%3A%2F%2Fexample.com%2Fcallback&response_type=code&scope=custom-scope1+custom-scope2&state=test-state"

	if url != expected {
		t.Errorf("GetAuthorizationURL() with custom scopes = %v, want %v", url, expected)
	}
}

func TestExchangeAuthorizationCode(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/oauth2/token" {
			t.Errorf("Expected path /oauth2/token, got %s", r.URL.Path)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
		}

		// Parse form
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check form values
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("Expected grant_type=authorization_code, got %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("code") != "test-code" {
			t.Errorf("Expected code=test-code, got %s", r.Form.Get("code"))
		}
		if r.Form.Get("redirect_uri") != "https://example.com/callback" {
			t.Errorf("Expected redirect_uri=https://example.com/callback, got %s", r.Form.Get("redirect_uri"))
		}
		if r.Form.Get("client_id") != "test-client-id" {
			t.Errorf("Expected client_id=test-client-id, got %s", r.Form.Get("client_id"))
		}
		if r.Form.Get("client_secret") != "test-client-secret" {
			t.Errorf("Expected client_secret=test-client-secret, got %s", r.Form.Get("client_secret"))
		}

		// Return mock response
		response := TokenResponse{
			AccessToken:    "test-access-token",
			RefreshToken:   "test-refresh-token",
			ExpiresIn:      3600,
			ResourceServer: "test-resource-server",
			TokenType:      "Bearer",
			Scope:          "openid profile email",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	client.SetRedirectURL("https://example.com/callback")

	// Test exchange
	token, err := client.ExchangeAuthorizationCode(context.Background(), "test-code")
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error = %v", err)
	}

	// Check token response
	if token.AccessToken != "test-access-token" {
		t.Errorf("ExchangeAuthorizationCode() AccessToken = %v, want %v", token.AccessToken, "test-access-token")
	}
	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("ExchangeAuthorizationCode() RefreshToken = %v, want %v", token.RefreshToken, "test-refresh-token")
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("ExchangeAuthorizationCode() ExpiresIn = %v, want %v", token.ExpiresIn, 3600)
	}

	// Check expiry time calculation (allow small difference due to processing time)
	expectedExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	timeDiff := expectedExpiry.Sub(token.ExpiryTime).Seconds()
	if timeDiff < -1 || timeDiff > 1 {
		t.Errorf("ExchangeAuthorizationCode() ExpiryTime = %v, want close to %v", token.ExpiryTime, expectedExpiry)
	}

	// Test missing redirect URL
	client.RedirectURL = ""
	_, err = client.ExchangeAuthorizationCode(context.Background(), "test-code")
	if err == nil {
		t.Error("ExchangeAuthorizationCode() with no redirect URL should return error")
	}
}

func TestRefreshToken(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/oauth2/token" {
			t.Errorf("Expected path /oauth2/token, got %s", r.URL.Path)
		}

		// Parse form
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check form values
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("Expected grant_type=refresh_token, got %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("refresh_token") != "test-refresh-token" {
			t.Errorf("Expected refresh_token=test-refresh-token, got %s", r.Form.Get("refresh_token"))
		}

		// Return mock response
		response := TokenResponse{
			AccessToken:    "new-access-token",
			RefreshToken:   "new-refresh-token",
			ExpiresIn:      3600,
			ResourceServer: "test-resource-server",
			TokenType:      "Bearer",
			Scope:          "openid profile email",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test refresh
	token, err := client.RefreshToken(context.Background(), "test-refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	// Check token response
	if token.AccessToken != "new-access-token" {
		t.Errorf("RefreshToken() AccessToken = %v, want %v", token.AccessToken, "new-access-token")
	}
	if token.RefreshToken != "new-refresh-token" {
		t.Errorf("RefreshToken() RefreshToken = %v, want %v", token.RefreshToken, "new-refresh-token")
	}
}

func TestIntrospectToken(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost || r.URL.Path != "/oauth2/token/introspect" {
			t.Errorf("Expected POST /oauth2/token/introspect, got %s %s", r.Method, r.URL.Path)
		}

		// Parse form
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check form values
		if r.Form.Get("token") != "test-token" {
			t.Errorf("Expected token=test-token, got %s", r.Form.Get("token"))
		}

		// Return mock response
		response := TokenInfo{
			Active:      true,
			Scope:       "openid profile email",
			ClientID:    "test-client-id",
			Username:    "test-user",
			Exp:         time.Now().Add(time.Hour).Unix(),
			Subject:     "test-subject",
			SubjectType: "user",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test introspection
	info, err := client.IntrospectToken(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("IntrospectToken() error = %v", err)
	}

	// Check token info
	if !info.Active {
		t.Error("IntrospectToken() returned inactive token")
	}
	if info.ClientID != "test-client-id" {
		t.Errorf("IntrospectToken() ClientID = %v, want %v", info.ClientID, "test-client-id")
	}
	if info.Username != "test-user" {
		t.Errorf("IntrospectToken() Username = %v, want %v", info.Username, "test-user")
	}
}

func TestRevokeToken(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost || r.URL.Path != "/oauth2/token/revoke" {
			t.Errorf("Expected POST /oauth2/token/revoke, got %s %s", r.Method, r.URL.Path)
		}

		// Parse form
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check form values
		if r.Form.Get("token") != "test-token" {
			t.Errorf("Expected token=test-token, got %s", r.Form.Get("token"))
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test revocation
	err := client.RevokeToken(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("RevokeToken() error = %v", err)
	}
}

func TestGetClientCredentialsToken(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != http.MethodPost || r.URL.Path != "/oauth2/token" {
			t.Errorf("Expected POST /oauth2/token, got %s %s", r.Method, r.URL.Path)
		}

		// Parse form
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		// Check form values
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Errorf("Expected grant_type=client_credentials, got %s", r.Form.Get("grant_type"))
		}
		if r.Form.Get("scope") != "custom-scope" {
			t.Errorf("Expected scope=custom-scope, got %s", r.Form.Get("scope"))
		}

		// Return mock response
		response := TokenResponse{
			AccessToken:    "client-credentials-token",
			ExpiresIn:      3600,
			ResourceServer: "test-resource-server",
			TokenType:      "Bearer",
			Scope:          "custom-scope",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test client credentials flow
	token, err := client.GetClientCredentialsToken(context.Background(), "custom-scope")
	if err != nil {
		t.Fatalf("GetClientCredentialsToken() error = %v", err)
	}

	// Check token response
	if token.AccessToken != "client-credentials-token" {
		t.Errorf("GetClientCredentialsToken() AccessToken = %v, want %v", token.AccessToken, "client-credentials-token")
	}
	if token.RefreshToken != "" {
		t.Errorf("GetClientCredentialsToken() should not return a refresh token, got %v", token.RefreshToken)
	}
	if token.Scope != "custom-scope" {
		t.Errorf("GetClientCredentialsToken() Scope = %v, want %v", token.Scope, "custom-scope")
	}

	// Test with empty client secret
	emptyClient := NewClient("test-client-id", "")
	_, err = emptyClient.GetClientCredentialsToken(context.Background())
	if err == nil {
		t.Error("GetClientCredentialsToken() with empty client secret should return error")
	}
}

func TestCreateClientCredentialsAuthorizer(t *testing.T) {
	client := NewClient("test-client-id", "test-client-secret")
	authorizer := client.CreateClientCredentialsAuthorizer("scope1", "scope2")

	if authorizer.ClientID != "test-client-id" {
		t.Errorf("CreateClientCredentialsAuthorizer() ClientID = %v, want %v", authorizer.ClientID, "test-client-id")
	}
	if authorizer.ClientSecret != "test-client-secret" {
		t.Errorf("CreateClientCredentialsAuthorizer() ClientSecret = %v, want %v", authorizer.ClientSecret, "test-client-secret")
	}
	if len(authorizer.Scopes) != 2 || authorizer.Scopes[0] != "scope1" || authorizer.Scopes[1] != "scope2" {
		t.Errorf("CreateClientCredentialsAuthorizer() Scopes = %v, want %v", authorizer.Scopes, []string{"scope1", "scope2"})
	}
	if authorizer.AuthFunc == nil {
		t.Error("CreateClientCredentialsAuthorizer() AuthFunc is nil")
	}
}

func TestCreateRefreshableTokenAuthorizer(t *testing.T) {
	client := NewClient("test-client-id", "test-client-secret")
	authorizer := client.CreateRefreshableTokenAuthorizer("test-access-token", "test-refresh-token", 3600)

	if authorizer.AccessToken != "test-access-token" {
		t.Errorf("CreateRefreshableTokenAuthorizer() AccessToken = %v, want %v", authorizer.AccessToken, "test-access-token")
	}
	if authorizer.RefreshToken != "test-refresh-token" {
		t.Errorf("CreateRefreshableTokenAuthorizer() RefreshToken = %v, want %v", authorizer.RefreshToken, "test-refresh-token")
	}
	if authorizer.RefreshFunc == nil {
		t.Error("CreateRefreshableTokenAuthorizer() RefreshFunc is nil")
	}

	// Check expiry time calculation (allow small difference due to processing time)
	expectedExpiry := time.Now().Add(time.Hour)
	timeDiff := expectedExpiry.Sub(authorizer.ExpiresAt).Seconds()
	if timeDiff < -1 || timeDiff > 1 {
		t.Errorf("CreateRefreshableTokenAuthorizer() ExpiresAt = %v, want close to %v", authorizer.ExpiresAt, expectedExpiry)
	}
}

func TestCreateStaticTokenAuthorizer(t *testing.T) {
	client := NewClient("test-client-id", "test-client-secret")
	authorizer := client.CreateStaticTokenAuthorizer("test-access-token")

	if authorizer.Token != "test-access-token" {
		t.Errorf("CreateStaticTokenAuthorizer() Token = %v, want %v", authorizer.Token, "test-access-token")
	}
}

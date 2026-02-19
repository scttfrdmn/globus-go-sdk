// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pkg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
)

// ---------------------------------------------------------------------------
// NewConfig / SDKConfig builder methods
// ---------------------------------------------------------------------------

func TestNewConfig_NotNil(t *testing.T) {
	cfg := pkg.NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
}

func TestNewConfig_WithClientID(t *testing.T) {
	cfg := pkg.NewConfig().WithClientID("client-123")
	if cfg.ClientID != "client-123" {
		t.Errorf("expected ClientID 'client-123', got %q", cfg.ClientID)
	}
}

func TestNewConfig_WithClientSecret(t *testing.T) {
	cfg := pkg.NewConfig().WithClientSecret("secret-xyz")
	if cfg.ClientSecret != "secret-xyz" {
		t.Errorf("expected ClientSecret 'secret-xyz', got %q", cfg.ClientSecret)
	}
}

func TestNewConfig_ChainBuilders(t *testing.T) {
	cfg := pkg.NewConfig().
		WithClientID("id-1").
		WithClientSecret("sec-1")
	if cfg.ClientID != "id-1" {
		t.Errorf("ClientID = %q, want 'id-1'", cfg.ClientID)
	}
	if cfg.ClientSecret != "sec-1" {
		t.Errorf("ClientSecret = %q, want 'sec-1'", cfg.ClientSecret)
	}
}

// ---------------------------------------------------------------------------
// NewConfigFromEnvironment
// ---------------------------------------------------------------------------

func TestNewConfigFromEnvironment_NotNil(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := pkg.NewConfigFromEnvironment()
	if cfg == nil {
		t.Fatal("NewConfigFromEnvironment() returned nil")
	}
}

func TestNewConfigFromEnvironment_WithPoolEnabled(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "")

	cfg := pkg.NewConfigFromEnvironment()
	if cfg == nil {
		t.Fatal("NewConfigFromEnvironment() returned nil when pool enabled")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewAuthClient
// ---------------------------------------------------------------------------

func TestNewAuthClient_Success(t *testing.T) {
	cfg := pkg.NewConfig().
		WithClientID("test-client-id").
		WithClientSecret("test-secret")
	client, err := cfg.NewAuthClient()
	if err != nil {
		t.Fatalf("NewAuthClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewAuthClient() returned nil client")
	}
}

func TestNewAuthClient_WithConnectionPoolDisabled(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := pkg.NewConfig().WithClientID("test-id")
	client, err := cfg.NewAuthClient()
	if err != nil {
		t.Fatalf("NewAuthClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewAuthClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewGroupsClient
// ---------------------------------------------------------------------------

func TestNewGroupsClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewGroupsClient("test-token")
	if err != nil {
		t.Fatalf("NewGroupsClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewGroupsClient() returned nil")
	}
}

func TestNewGroupsClient_WithDebug(t *testing.T) {
	orig := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	defer os.Setenv("GLOBUS_SDK_HTTP_DEBUG", orig)
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewGroupsClient("tok")
	if err != nil {
		t.Fatalf("NewGroupsClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewGroupsClient() returned nil")
	}
}

func TestNewGroupsClient_WithTracing(t *testing.T) {
	orig := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer os.Setenv("GLOBUS_SDK_HTTP_TRACE", orig)
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewGroupsClient("tok")
	if err != nil {
		t.Fatalf("NewGroupsClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewGroupsClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewTransferClient
// ---------------------------------------------------------------------------

func TestNewTransferClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewTransferClient("test-token")
	if err != nil {
		t.Fatalf("NewTransferClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTransferClient() returned nil")
	}
}

func TestNewTransferClient_WithDebugAndTrace(t *testing.T) {
	origDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	origTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer func() {
		os.Setenv("GLOBUS_SDK_HTTP_DEBUG", origDebug)
		os.Setenv("GLOBUS_SDK_HTTP_TRACE", origTrace)
	}()
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewTransferClient("tok")
	if err != nil {
		t.Fatalf("NewTransferClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTransferClient() returned nil")
	}
}

func TestNewTransferClient_ConnectionPoolDisabled(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := pkg.NewConfig()
	client, err := cfg.NewTransferClient("tok")
	if err != nil {
		t.Fatalf("NewTransferClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTransferClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewSearchClient
// ---------------------------------------------------------------------------

func TestNewSearchClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewSearchClient("test-token")
	if err != nil {
		t.Fatalf("NewSearchClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewSearchClient() returned nil")
	}
}

func TestNewSearchClient_WithDebugAndTrace(t *testing.T) {
	origDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	origTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer func() {
		os.Setenv("GLOBUS_SDK_HTTP_DEBUG", origDebug)
		os.Setenv("GLOBUS_SDK_HTTP_TRACE", origTrace)
	}()
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewSearchClient("tok")
	if err != nil {
		t.Fatalf("NewSearchClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewSearchClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewFlowsClient
// ---------------------------------------------------------------------------

func TestNewFlowsClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewFlowsClient("test-token")
	if err != nil {
		t.Fatalf("NewFlowsClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewFlowsClient() returned nil")
	}
}

func TestNewFlowsClient_WithDebugAndTrace(t *testing.T) {
	origDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	origTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer func() {
		os.Setenv("GLOBUS_SDK_HTTP_DEBUG", origDebug)
		os.Setenv("GLOBUS_SDK_HTTP_TRACE", origTrace)
	}()
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewFlowsClient("tok")
	if err != nil {
		t.Fatalf("NewFlowsClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewFlowsClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewComputeClient
// ---------------------------------------------------------------------------

func TestNewComputeClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewComputeClient("test-token")
	if err != nil {
		t.Fatalf("NewComputeClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewComputeClient() returned nil")
	}
}

func TestNewComputeClient_WithDebugAndTrace(t *testing.T) {
	origDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	origTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer func() {
		os.Setenv("GLOBUS_SDK_HTTP_DEBUG", origDebug)
		os.Setenv("GLOBUS_SDK_HTTP_TRACE", origTrace)
	}()
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewComputeClient("tok")
	if err != nil {
		t.Fatalf("NewComputeClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewComputeClient() returned nil")
	}
}

func TestNewComputeClient_ConnectionPoolDisabled(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := pkg.NewConfig()
	client, err := cfg.NewComputeClient("tok")
	if err != nil {
		t.Fatalf("NewComputeClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewComputeClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewTimersClient
// ---------------------------------------------------------------------------

func TestNewTimersClient_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	client, err := cfg.NewTimersClient("test-token")
	if err != nil {
		t.Fatalf("NewTimersClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTimersClient() returned nil")
	}
}

func TestNewTimersClient_WithDebugAndTrace(t *testing.T) {
	origDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG")
	origTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE")
	defer func() {
		os.Setenv("GLOBUS_SDK_HTTP_DEBUG", origDebug)
		os.Setenv("GLOBUS_SDK_HTTP_TRACE", origTrace)
	}()
	os.Setenv("GLOBUS_SDK_HTTP_DEBUG", "1")
	os.Setenv("GLOBUS_SDK_HTTP_TRACE", "1")

	cfg := pkg.NewConfig()
	client, err := cfg.NewTimersClient("tok")
	if err != nil {
		t.Fatalf("NewTimersClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTimersClient() returned nil")
	}
}

func TestNewTimersClient_ConnectionPoolDisabled(t *testing.T) {
	orig := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", orig)
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := pkg.NewConfig()
	client, err := cfg.NewTimersClient("tok")
	if err != nil {
		t.Fatalf("NewTimersClient() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewTimersClient() returned nil")
	}
}

// ---------------------------------------------------------------------------
// SDKConfig.NewTokenManager
// ---------------------------------------------------------------------------

func TestNewTokenManager_Success(t *testing.T) {
	cfg := pkg.NewConfig()
	manager, err := cfg.NewTokenManager()
	if err != nil {
		t.Fatalf("NewTokenManager() error: %v", err)
	}
	if manager == nil {
		t.Fatal("NewTokenManager() returned nil")
	}
}

// ---------------------------------------------------------------------------
// GetScopesByService
// ---------------------------------------------------------------------------

func TestGetScopesByService_Known(t *testing.T) {
	tests := []struct {
		service string
		scope   string
	}{
		{"auth", pkg.AuthScope},
		{"groups", pkg.GroupsScope},
		{"transfer", pkg.TransferScope},
		{"search", pkg.SearchScope},
		{"flows", pkg.FlowsScope},
		{"compute", pkg.ComputeScope},
	}
	for _, tc := range tests {
		scopes := pkg.GetScopesByService(tc.service)
		if len(scopes) != 1 {
			t.Errorf("GetScopesByService(%q) returned %d scopes, want 1", tc.service, len(scopes))
			continue
		}
		if scopes[0] != tc.scope {
			t.Errorf("GetScopesByService(%q)[0] = %q, want %q", tc.service, scopes[0], tc.scope)
		}
	}
}

func TestGetScopesByService_Unknown(t *testing.T) {
	scopes := pkg.GetScopesByService("nonexistent")
	if len(scopes) != 0 {
		t.Errorf("expected 0 scopes for unknown service, got %d", len(scopes))
	}
}

func TestGetScopesByService_Multiple(t *testing.T) {
	scopes := pkg.GetScopesByService("auth", "transfer", "search")
	if len(scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(scopes))
	}
}

func TestGetScopesByService_Empty(t *testing.T) {
	scopes := pkg.GetScopesByService()
	if len(scopes) != 0 {
		t.Errorf("expected 0 scopes for empty call, got %d", len(scopes))
	}
}

// ---------------------------------------------------------------------------
// Scope constants are non-empty
// ---------------------------------------------------------------------------

func TestScopeConstants_NonEmpty(t *testing.T) {
	scopes := map[string]string{
		"AuthScope":     pkg.AuthScope,
		"GroupsScope":   pkg.GroupsScope,
		"TransferScope": pkg.TransferScope,
		"SearchScope":   pkg.SearchScope,
		"FlowsScope":    pkg.FlowsScope,
		"ComputeScope":  pkg.ComputeScope,
		"TimersScope":   pkg.TimersScope,
		"TokensScope":   pkg.TokensScope,
	}
	for name, scope := range scopes {
		if scope == "" {
			t.Errorf("%s is empty", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Version constant
// ---------------------------------------------------------------------------

func TestVersion_NonEmpty(t *testing.T) {
	if pkg.Version == "" {
		t.Error("pkg.Version is empty")
	}
}

// ---------------------------------------------------------------------------
// WithConfig
// ---------------------------------------------------------------------------

func TestWithConfig_SetsConfig(t *testing.T) {
	// Just verify it doesn't panic and returns a valid SDKConfig.
	cfg := pkg.NewConfig().WithConfig(nil)
	if cfg == nil {
		t.Fatal("WithConfig(nil) returned nil SDKConfig")
	}
}

// ---------------------------------------------------------------------------
// NewComputeClientV2 deprecated alias
// ---------------------------------------------------------------------------

func TestNewComputeClientV2_Works(t *testing.T) {
	client, err := pkg.NewComputeClientV2()
	if err != nil {
		t.Fatalf("NewComputeClientV2() error: %v", err)
	}
	if client == nil {
		t.Fatal("NewComputeClientV2() returned nil")
	}
}

// ---------------------------------------------------------------------------
// WithClientOption
// ---------------------------------------------------------------------------

func TestWithClientOption_NoError(t *testing.T) {
	// WithClientOption applies the option to a temporary client internally.
	// Verify it doesn't panic and returns the same config instance.
	cfg := pkg.NewConfig()
	result := cfg.WithClientOption(core.WithBaseURL("https://example.com/"))
	if result == nil {
		t.Fatal("WithClientOption() returned nil")
	}
}

func TestWithClientOption_WithNilConfig(t *testing.T) {
	// When Config is nil, WithClientOption should create a default one.
	cfg := &pkg.SDKConfig{}
	result := cfg.WithClientOption(core.WithBaseURL("https://example.com/"))
	if result == nil {
		t.Fatal("WithClientOption() on nil Config returned nil")
	}
}

// ---------------------------------------------------------------------------
// NewTokenManagerWithAuth
// ---------------------------------------------------------------------------

func TestNewTokenManagerWithAuth_NoStorageDir(t *testing.T) {
	cfg := pkg.NewConfig().
		WithClientID("test-client-id").
		WithClientSecret("test-client-secret")
	// NewTokenManagerWithAuth creates an auth client internally; it should
	// not error during construction.
	manager, err := cfg.NewTokenManagerWithAuth("")
	if err != nil {
		t.Fatalf("NewTokenManagerWithAuth(\"\") error: %v", err)
	}
	if manager == nil {
		t.Fatal("NewTokenManagerWithAuth returned nil manager")
	}
}

// TestSimpleAuthorizer_ViaGroupsClient exercises the simpleAuthorizer.GetAuthorizationHeader
// path by making a real HTTP request through a pkg-created groups client.
func TestSimpleAuthorizer_ViaGroupsClient(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		// Return a minimal group list
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"groups": []interface{}{},
			"total":  0,
		})
	}))
	defer server.Close()

	cfg := pkg.NewConfig()
	groupsClient, err := cfg.NewGroupsClient("test-bearer-token")
	if err != nil {
		t.Fatalf("NewGroupsClient() error: %v", err)
	}

	// Override the base URL to point to our test server.
	groupsClient.Client.BaseURL = server.URL + "/"

	// Make a real request which will trigger GetAuthorizationHeader on the simpleAuthorizer.
	_, _ = groupsClient.ListGroups(context.Background(), nil)

	if gotAuth != "Bearer test-bearer-token" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-bearer-token")
	}
}

// TestSimpleAuthorizer_EmptyToken exercises the empty-token branch of simpleAuthorizer.
func TestSimpleAuthorizer_EmptyToken(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"groups": []interface{}{}, "total": 0})
	}))
	defer server.Close()

	cfg := pkg.NewConfig()
	// Empty token - simpleAuthorizer should return empty string (no Authorization header).
	groupsClient, err := cfg.NewGroupsClient("")
	if err != nil {
		t.Fatalf("NewGroupsClient(\"\") error: %v", err)
	}
	groupsClient.Client.BaseURL = server.URL + "/"
	_, _ = groupsClient.ListGroups(context.Background(), nil)
	if gotAuth != "" {
		t.Errorf("expected empty Authorization header for empty token, got %q", gotAuth)
	}
}

func TestNewTokenManagerWithAuth_WithStorageDir(t *testing.T) {
	// Use a temp directory as the storage directory.
	dir := t.TempDir()
	cfg := pkg.NewConfig().
		WithClientID("test-client-id").
		WithClientSecret("test-client-secret")
	manager, err := cfg.NewTokenManagerWithAuth(dir)
	if err != nil {
		t.Fatalf("NewTokenManagerWithAuth(dir) error: %v", err)
	}
	if manager == nil {
		t.Fatal("NewTokenManagerWithAuth returned nil manager")
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockAuthorizer is a test implementation of Authorizer.
type mockAuthorizer struct {
	header string
	err    error
}

func (m *mockAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	return m.header, m.err
}

func (m *mockAuthorizer) HandleMissingAuthorization(_ context.Context) bool {
	return false
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid with access token",
			config:  &Config{AccessToken: "tok", Scopes: []string{"openid"}},
			wantErr: false,
		},
		{
			name:    "valid with authorizer",
			config:  &Config{Authorizer: &mockAuthorizer{header: "Bearer x"}, Scopes: []string{"openid"}},
			wantErr: false,
		},
		{
			name:    "invalid: neither token nor authorizer",
			config:  &Config{Scopes: []string{"openid"}},
			wantErr: true,
		},
		{
			name:    "invalid: no scopes",
			config:  &Config{AccessToken: "tok"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientUsesAuthorizer(t *testing.T) {
	// Set up a test server that records the Authorization header it receives.
	var gotAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	auth := &mockAuthorizer{header: "Bearer authorizer-token"}
	config := &Config{
		Authorizer: auth,
		Scopes:     []string{"openid"},
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	var result map[string]string
	if err := client.DoRequest(ctx, http.MethodGet, "/test", nil, nil, &result); err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}

	if gotAuthHeader != "Bearer authorizer-token" {
		t.Errorf("Authorization header = %q, want %q", gotAuthHeader, "Bearer authorizer-token")
	}
}

func TestClientFallsBackToAccessToken(t *testing.T) {
	var gotAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	config := &Config{
		AccessToken: "static-token",
		Scopes:      []string{"openid"},
		BaseURL:     server.URL,
		HTTPClient:  server.Client(),
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx := context.Background()
	var result map[string]string
	if err := client.DoRequest(ctx, http.MethodGet, "/test", nil, nil, &result); err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}

	if gotAuthHeader != "Bearer static-token" {
		t.Errorf("Authorization header = %q, want %q", gotAuthHeader, "Bearer static-token")
	}
}

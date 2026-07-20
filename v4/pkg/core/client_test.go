// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestClientClose tests the Close method
func TestClientClose(t *testing.T) {
	tests := []struct {
		name              string
		provideHTTPClient bool
		wantErr           bool
	}{
		{
			name:              "Close with internally created HTTP client",
			provideHTTPClient: false,
			wantErr:           false,
		},
		{
			name:              "Close with user-provided HTTP client",
			provideHTTPClient: true,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				AccessToken: "test-token",
				Scopes:      []string{"test-scope"},
			}

			if tt.provideHTTPClient {
				config.HTTPClient = &http.Client{}
			}

			client, err := NewClient(config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			// Close should not error
			err = client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Calling Close multiple times should be safe
			err = client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() second call error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBuildURLPreservesBasePath verifies that buildURL joins the endpoint onto
// the base URL's path instead of overwriting it, so a version prefix carried by
// the base URL (e.g. /v2, /v0.10) survives.
func TestBuildURLPreservesBasePath(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		want     string
	}{
		{
			name:     "auth base keeps /v2 prefix",
			baseURL:  "https://auth.globus.org/v2",
			endpoint: "/oauth2/token",
			want:     "https://auth.globus.org/v2/oauth2/token",
		},
		{
			name:     "transfer base keeps /v0.10 prefix",
			baseURL:  "https://transfer.api.globus.org/v0.10",
			endpoint: "/task_list",
			want:     "https://transfer.api.globus.org/v0.10/task_list",
		},
		{
			name:     "path-less base joins cleanly",
			baseURL:  "http://127.0.0.1:8080",
			endpoint: "/endpoint/abc",
			want:     "http://127.0.0.1:8080/endpoint/abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{config: &Config{BaseURL: tt.baseURL}}
			got := c.buildURL(tt.endpoint, nil)
			if got != tt.want {
				t.Errorf("buildURL(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

// TestDoRequestFormEncoding verifies that a url.Values body is sent as
// application/x-www-form-urlencoded with a flat form body (not JSON), while a
// struct body is still sent as JSON.
func TestDoRequestFormEncoding(t *testing.T) {
	var gotContentType, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		AccessToken: "test-token",
		Scopes:      []string{"test-scope"},
		BaseURL:     server.URL,
		RetryConfig: &RetryConfig{MaxRetries: 0},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Run("url.Values sends form encoding", func(t *testing.T) {
		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("code", "abc")
		if err := client.DoRequest(context.Background(), http.MethodPost, "/oauth2/token", nil, form, nil); err != nil {
			t.Fatalf("DoRequest() error = %v", err)
		}
		if gotContentType != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", gotContentType)
		}
		if gotBody != "code=abc&grant_type=authorization_code" {
			t.Errorf("body = %q, want form-encoded flat body", gotBody)
		}
	})

	t.Run("struct sends JSON encoding", func(t *testing.T) {
		if err := client.DoRequest(context.Background(), http.MethodPost, "/thing", nil, map[string]string{"a": "b"}, nil); err != nil {
			t.Fatalf("DoRequest() error = %v", err)
		}
		if gotContentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", gotContentType)
		}
		if gotBody != `{"a":"b"}` {
			t.Errorf("body = %q, want JSON body", gotBody)
		}
	})
}

// TestDoRequestNoAuth verifies that DoRequestNoAuth omits the Authorization
// header while DoRequest sets it.
func TestDoRequestNoAuth(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		AccessToken: "test-token",
		Scopes:      []string{"test-scope"},
		BaseURL:     server.URL,
		RetryConfig: &RetryConfig{MaxRetries: 0},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := client.DoRequestNoAuth(context.Background(), http.MethodGet, "/info", nil, nil, nil); err != nil {
		t.Fatalf("DoRequestNoAuth() error = %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header = %q, want empty", gotAuth)
	}

	if err := client.DoRequest(context.Background(), http.MethodGet, "/info", nil, nil, nil); err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want Bearer test-token", gotAuth)
	}
}

// TestClientCloseNilConfig tests Close with nil config edge case
func TestClientCloseNilConfig(t *testing.T) {
	// This test ensures Close handles edge cases gracefully
	client := &Client{
		config:            &Config{},
		httpClientCreated: true,
	}

	err := client.Close()
	if err != nil {
		t.Errorf("Close() with nil HTTPClient error = %v, want nil", err)
	}
}

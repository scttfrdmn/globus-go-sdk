// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// NewMockServer creates an httptest.Server with the given handler and registers
// t.Cleanup to close it when the test finishes.
func NewMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// NewTestConfig returns a *core.Config pointing at serverURL with retries
// disabled — suitable for unit tests that don't want retry delay.
func NewTestConfig(serverURL string) *core.Config {
	return &core.Config{
		AccessToken: "test-token",
		Scopes:      []string{"test-scope"},
		BaseURL:     serverURL,
		RetryConfig: &core.RetryConfig{MaxRetries: 0},
	}
}

// RespondJSON writes an HTTP response with Content-Type: application/json,
// the given status code, and body marshaled as JSON.
func RespondJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		json.NewEncoder(w).Encode(body) //nolint:errcheck
	}
}

// RespondError writes a structured Globus-style API error response.
func RespondError(w http.ResponseWriter, status int, message, code string) {
	RespondJSON(w, status, map[string]interface{}{
		"message": message,
		"code":    code,
	})
}

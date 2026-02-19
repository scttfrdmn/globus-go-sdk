// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/ratelimit"
)

// ---------------------------------------------------------------------------
// NewClient / Client options
// ---------------------------------------------------------------------------

func TestNewClient_Defaults(t *testing.T) {
	c := core.NewClient()
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.HTTPClient == nil {
		t.Fatal("expected non-nil HTTPClient")
	}
	if c.UserAgent == "" {
		t.Fatal("expected non-empty UserAgent")
	}
}

func TestNewClient_WithBaseURL(t *testing.T) {
	c := core.NewClient(core.WithBaseURL("https://example.com/"))
	if c.BaseURL != "https://example.com/" {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, "https://example.com/")
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := core.NewClient(core.WithHTTPClient(custom))
	if c.HTTPClient != custom {
		t.Error("expected custom HTTPClient to be set")
	}
}

func TestNewClient_WithHTTPDebugging(t *testing.T) {
	c := core.NewClient(core.WithHTTPDebugging(true))
	if !c.Debug {
		t.Error("expected Debug=true")
	}
}

func TestNewClient_WithHTTPTracing(t *testing.T) {
	c := core.NewClient(core.WithHTTPTracing(true))
	if !c.Trace {
		t.Error("expected Trace=true")
	}
	// Tracing implies Debug
	if !c.Debug {
		t.Error("expected Debug=true when Trace=true")
	}
}

func TestNewClient_WithVersionCheck(t *testing.T) {
	vc := core.NewVersionCheck()
	vc.DisableVersionCheck()
	c := core.NewClient(core.WithVersionCheck(vc))
	if c.VersionCheck == nil {
		t.Fatal("expected non-nil VersionCheck")
	}
	if c.VersionCheck.IsEnabled() {
		t.Error("expected VersionCheck to be disabled")
	}
}

// ---------------------------------------------------------------------------
// Client getters
// ---------------------------------------------------------------------------

func TestClient_GetBaseURL(t *testing.T) {
	c := core.NewClient(core.WithBaseURL("https://api.example.com/"))
	if got := c.GetBaseURL(); got != "https://api.example.com/" {
		t.Errorf("GetBaseURL() = %q", got)
	}
}

func TestClient_GetHTTPClient(t *testing.T) {
	c := core.NewClient()
	if c.GetHTTPClient() == nil {
		t.Fatal("GetHTTPClient returned nil")
	}
}

func TestClient_GetUserAgent(t *testing.T) {
	c := core.NewClient()
	if c.GetUserAgent() == "" {
		t.Fatal("GetUserAgent returned empty string")
	}
}

func TestClient_GetLogger(t *testing.T) {
	c := core.NewClient()
	if c.GetLogger() == nil {
		t.Fatal("GetLogger returned nil")
	}
}

// ---------------------------------------------------------------------------
// Client.Do
// ---------------------------------------------------------------------------

func TestClient_Do_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := core.NewClient(core.WithBaseURL(server.URL + "/"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/test", nil)
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestClient_Do_SetsUserAgent(t *testing.T) {
	var gotUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := core.NewClient(core.WithBaseURL(server.URL + "/"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	resp.Body.Close()
	if gotUA == "" {
		t.Error("User-Agent header was not set")
	}
}

func TestClient_Do_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"code": "AuthenticationFailed", "message": "No token"},
			},
		})
	}))
	defer server.Close()

	c := core.NewClient(core.WithBaseURL(server.URL + "/"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !core.IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized=true, got false; err=%v", err)
	}
}

func TestClient_Do_WithAuthorizer(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Use a simple static authorizer via the core authorizers package through a shim.
	c := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithAuthorizer(&staticAuth{header: "Bearer mytoken"}),
	)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	resp.Body.Close()
	if gotAuth != "Bearer mytoken" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer mytoken")
	}
}

// staticAuth is a minimal Authorizer implementation for tests.
type staticAuth struct{ header string }

func (a *staticAuth) GetAuthorizationHeader(_ ...context.Context) (string, error) {
	return a.header, nil
}

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

func TestNewAPIError_ParsedError(t *testing.T) {
	body := `{"errors":[{"code":"NotFound","message":"resource missing"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(body))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/")
	err := core.NewAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !core.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true; err=%v", err)
	}
}

func TestNewAPIError_UnparsedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/")
	err := core.NewAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestNewAPIError_EmptyErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[]}`))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/")
	err := core.NewAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestIsUnauthorized_False(t *testing.T) {
	if core.IsUnauthorized(nil) {
		t.Error("IsUnauthorized(nil) should be false")
	}
}

func TestIsNotFound_False(t *testing.T) {
	if core.IsNotFound(nil) {
		t.Error("IsNotFound(nil) should be false")
	}
}

func TestIsForbidden_True(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"code": "Forbidden", "message": "no access"},
			},
		})
	}))
	defer server.Close()

	c := core.NewClient(core.WithBaseURL(server.URL + "/"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !core.IsForbidden(err) {
		t.Errorf("expected IsForbidden=true; err=%v", err)
	}
}

func TestIsForbidden_False(t *testing.T) {
	if core.IsForbidden(nil) {
		t.Error("IsForbidden(nil) should be false")
	}
}

// ---------------------------------------------------------------------------
// Logger
// ---------------------------------------------------------------------------

func TestDefaultLogger_Levels(t *testing.T) {
	tests := []struct {
		level   core.LogLevel
		method  func(*core.DefaultLogger, string)
		wantOut bool
	}{
		{core.LogLevelDebug, func(l *core.DefaultLogger, msg string) { l.Debug(msg) }, true},
		{core.LogLevelInfo, func(l *core.DefaultLogger, msg string) { l.Info(msg) }, true},
		{core.LogLevelWarn, func(l *core.DefaultLogger, msg string) { l.Warn(msg) }, true},
		{core.LogLevelError, func(l *core.DefaultLogger, msg string) { l.Error(msg) }, true},
		// None level: nothing should be logged.
		{core.LogLevelNone, func(l *core.DefaultLogger, msg string) { l.Debug(msg) }, false},
	}

	for _, tc := range tests {
		buf := &bytes.Buffer{}
		logger := core.NewDefaultLogger(buf, tc.level)
		tc.method(logger, "test-message")
		if tc.wantOut && !strings.Contains(buf.String(), "test-message") {
			t.Errorf("level %d: expected output to contain 'test-message', got %q", tc.level, buf.String())
		}
		if !tc.wantOut && buf.Len() > 0 {
			t.Errorf("level %d: expected no output, got %q", tc.level, buf.String())
		}
	}
}

func TestDefaultLogger_Info_BelowLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	// Only errors should appear
	logger := core.NewDefaultLogger(buf, core.LogLevelError)
	logger.Info("should not appear")
	logger.Warn("should not appear")
	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Errorf("expected no output for below-level messages, got %q", buf.String())
	}
}

func TestDefaultLogger_Error_Appears(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := core.NewDefaultLogger(buf, core.LogLevelError)
	logger.Error("this is an error")
	if !strings.Contains(buf.String(), "this is an error") {
		t.Errorf("expected error message in output, got %q", buf.String())
	}
}

func TestWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	custom := core.NewDefaultLogger(buf, core.LogLevelDebug)
	c := core.NewClient(core.WithLogger(custom))
	if c.GetLogger() != custom {
		t.Error("expected custom logger to be set")
	}
}

func TestWithLogLevel_DefaultLogger(t *testing.T) {
	// When the client already has a DefaultLogger, WithLogLevel should update it.
	c := core.NewClient(core.WithLogLevel(core.LogLevelDebug))
	if c.Logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestWithLogLevel_NilLogger(t *testing.T) {
	// If Logger is not a *DefaultLogger the option should create a new one.
	// Start with a custom non-DefaultLogger, then apply WithLogLevel.
	buf := &bytes.Buffer{}
	customLogger := &customLog{buf: buf}
	c := core.NewClient(core.WithLogger(customLogger), core.WithLogLevel(core.LogLevelInfo))
	if c.Logger == nil {
		t.Fatal("expected non-nil logger after WithLogLevel")
	}
}

// customLog is a non-DefaultLogger Logger implementation for testing WithLogLevel fallback.
type customLog struct{ buf *bytes.Buffer }

func (l *customLog) Debug(format string, v ...interface{}) {}
func (l *customLog) Info(format string, v ...interface{})  {}
func (l *customLog) Warn(format string, v ...interface{})  {}
func (l *customLog) Error(format string, v ...interface{}) {}

// ---------------------------------------------------------------------------
// Version helpers
// ---------------------------------------------------------------------------

func TestGetInfo(t *testing.T) {
	info := core.GetInfo()
	if info.Version == "" {
		t.Error("GetInfo().Version is empty")
	}
}

func TestIsDevelopment(t *testing.T) {
	// Just ensure the function runs without panicking.
	_ = core.IsDevelopment()
}

func TestUserAgent(t *testing.T) {
	ua := core.UserAgent()
	if ua == "" {
		t.Error("UserAgent() returned empty string")
	}
	if !strings.Contains(ua, "Globus-Go-SDK") {
		t.Errorf("UserAgent() = %q, expected to contain 'Globus-Go-SDK'", ua)
	}
}

func TestParseVersion_Empty(t *testing.T) {
	_, err := core.ParseVersion("auth", "")
	if err == nil {
		t.Fatal("expected error for empty version string")
	}
}

func TestParseVersion_BetaWithMinor(t *testing.T) {
	v, err := core.ParseVersion("flows", "v1.2-beta")
	if err != nil {
		t.Fatalf("ParseVersion error: %v", err)
	}
	if !v.Beta {
		t.Error("expected Beta=true")
	}
	if v.Minor != 2 {
		t.Errorf("expected Minor=2, got %d", v.Minor)
	}
}

func TestAPIVersion_Compare(t *testing.T) {
	v1, _ := core.ParseVersion("auth", "v1")
	v2, _ := core.ParseVersion("auth", "v2")
	v1b, _ := core.ParseVersion("auth", "v1")

	if v1.Compare(v2) != -1 {
		t.Errorf("v1.Compare(v2) = %d, want -1", v1.Compare(v2))
	}
	if v2.Compare(v1) != 1 {
		t.Errorf("v2.Compare(v1) = %d, want 1", v2.Compare(v1))
	}
	if v1.Compare(v1b) != 0 {
		t.Errorf("v1.Compare(v1b) = %d, want 0", v1.Compare(v1b))
	}
}

func TestAPIVersion_Compare_Minor(t *testing.T) {
	v1, _ := core.ParseVersion("auth", "v1.0")
	v2, _ := core.ParseVersion("auth", "v1.1")

	if v1.Compare(v2) != -1 {
		t.Errorf("v1.Compare(v2) = %d, want -1 (minor)", v1.Compare(v2))
	}
}

func TestAPIVersion_Compare_Patch(t *testing.T) {
	v1, _ := core.ParseVersion("auth", "v1.0.0")
	v2, _ := core.ParseVersion("auth", "v1.0.1")

	if v1.Compare(v2) != -1 {
		t.Errorf("v1.Compare(v2) = %d, want -1 (patch)", v1.Compare(v2))
	}
}

func TestAPIVersion_Compare_Beta(t *testing.T) {
	vBeta, _ := core.ParseVersion("auth", "v1-beta")
	vStable, _ := core.ParseVersion("auth", "v1")

	// Beta < stable
	if vBeta.Compare(vStable) != -1 {
		t.Errorf("beta.Compare(stable) = %d, want -1", vBeta.Compare(vStable))
	}
}

func TestAPIVersion_IsCompatible_Value(t *testing.T) {
	v1, _ := core.ParseVersion("auth", "v2")

	// Pass by value instead of pointer
	v2, _ := core.ParseVersion("auth", "v2")
	if !v1.IsCompatible(*v2) {
		t.Error("expected compatible when passing by value")
	}
}

func TestAPIVersion_IsCompatible_WrongType(t *testing.T) {
	v1, _ := core.ParseVersion("auth", "v2")
	if v1.IsCompatible("not a version") {
		t.Error("expected not compatible for wrong type")
	}
}

func TestAPIVersion_GetEndpoint_AllServices(t *testing.T) {
	services := []string{"transfer", "auth", "search", "groups", "flows", "compute", "timers"}
	for _, svc := range services {
		v, _ := core.ParseVersion(svc, "v1")
		ep := v.GetEndpoint()
		if ep == "" {
			t.Errorf("GetEndpoint() for service %q returned empty string", svc)
		}
	}
}

func TestAPIVersion_Endpoint_Alias(t *testing.T) {
	v, _ := core.ParseVersion("auth", "v2")
	if v.Endpoint() != v.GetEndpoint() {
		t.Error("Endpoint() and GetEndpoint() should return the same value")
	}
}

func TestAPIVersion_String_AllFormats(t *testing.T) {
	tests := []struct {
		version  string
		service  string
		expected string
	}{
		{"v1", "auth", "v1"},
		{"v1.2", "auth", "v1.2"},
		{"v1.2.3", "auth", "v1.2.3"},
		{"v1-beta", "auth", "v1-beta"},
		{"v1.2-beta", "auth", "v1.2-beta"},
		{"v1.2.3-beta", "auth", "v1.2.3-beta"},
	}
	for _, tc := range tests {
		v, err := core.ParseVersion(tc.service, tc.version)
		if err != nil {
			t.Errorf("ParseVersion(%q) error: %v", tc.version, err)
			continue
		}
		got := v.String()
		if got != tc.expected {
			t.Errorf("String() for %q = %q, want %q", tc.version, got, tc.expected)
		}
	}
}

func TestExtractVersionFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://transfer.api.globus.org/v0.10/endpoint", "v0.10"},
		{"https://auth.globus.org/v2/token", "v2"},
		{"https://example.com/no-version/path", ""},
		{"https://example.com/v1-beta/path", "v1-beta"},
	}
	for _, tc := range tests {
		got := core.ExtractVersionFromURL(tc.url)
		if got != tc.expected {
			t.Errorf("ExtractVersionFromURL(%q) = %q, want %q", tc.url, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// VersionCheck
// ---------------------------------------------------------------------------

func TestVersionCheck_EnableDisable(t *testing.T) {
	vc := core.NewVersionCheck()
	if !vc.IsEnabled() {
		t.Fatal("expected enabled by default")
	}
	vc.DisableVersionCheck()
	if vc.IsEnabled() {
		t.Error("expected disabled after DisableVersionCheck")
	}
	vc.EnableVersionCheck()
	if !vc.IsEnabled() {
		t.Error("expected enabled after EnableVersionCheck")
	}
}

func TestVersionCheck_Enabled_Alias(t *testing.T) {
	vc := core.NewVersionCheck()
	if !vc.Enabled() {
		t.Error("Enabled() should be true by default")
	}
}

func TestVersionCheck_SetGetCustomVersion(t *testing.T) {
	vc := core.NewVersionCheck()
	err := vc.SetCustomVersion("auth", "v3")
	if err != nil {
		t.Fatalf("SetCustomVersion error: %v", err)
	}
	v, ok := vc.GetCustomVersion("auth")
	if !ok {
		t.Fatal("expected custom version to be set")
	}
	if v != "v3" {
		t.Errorf("expected 'v3', got %q", v)
	}
}

func TestVersionCheck_SetCustomVersion_Invalid(t *testing.T) {
	vc := core.NewVersionCheck()
	err := vc.SetCustomVersion("auth", "")
	if err == nil {
		t.Fatal("expected error for empty version string")
	}
}

func TestVersionCheck_GetCustomVersion_NotSet(t *testing.T) {
	vc := core.NewVersionCheck()
	_, ok := vc.GetCustomVersion("nonexistent")
	if ok {
		t.Error("expected not-ok for unset service")
	}
}

func TestVersionCheck_MarkIsServiceChecked(t *testing.T) {
	vc := core.NewVersionCheck()
	if vc.IsServiceChecked("auth") {
		t.Error("expected auth to not be checked yet")
	}
	vc.MarkServiceChecked("auth")
	if !vc.IsServiceChecked("auth") {
		t.Error("expected auth to be marked checked")
	}
}

func TestVersionCheck_CheckServiceVersion_InvalidParsed(t *testing.T) {
	vc := core.NewVersionCheck()
	err := vc.CheckServiceVersion("auth", "")
	if err == nil {
		t.Fatal("expected error for empty version")
	}
}

// ---------------------------------------------------------------------------
// Connection pool wrappers
// ---------------------------------------------------------------------------

func TestGetHTTPClientForService_NilManager(t *testing.T) {
	// Reset to nil manager to test the nil branch.
	core.SetConnectionPoolManager(nil)
	client := core.GetHTTPClientForService("auth")
	if client == nil {
		t.Fatal("expected non-nil http.Client even when manager is nil")
	}
	// Restore default state
	core.EnableDefaultConnectionPool()
}

func TestGetConnectionPool_NilManager(t *testing.T) {
	core.SetConnectionPoolManager(nil)
	pool := core.GetConnectionPool("auth", nil)
	if pool != nil {
		t.Error("expected nil pool when manager is nil")
	}
}

// ---------------------------------------------------------------------------
// RateLimiter integration (via Do)
// ---------------------------------------------------------------------------

func TestClient_Do_WithRateLimiter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Use the real NoopRateLimiter from the ratelimit package via NewClient.
	// We just verify the client works when a rate limiter is configured.
	rl := ratelimit.NewNoopRateLimiter()
	c := core.NewClient(
		core.WithBaseURL(server.URL+"/"),
		core.WithRateLimiter(rl),
	)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() with RateLimiter error: %v", err)
	}
	resp.Body.Close()
	// The NoopRateLimiter tracks total requests
	stats := rl.GetStats()
	if stats.TotalRequests == 0 {
		t.Error("expected RateLimiter.Wait to have been called (TotalRequests == 0)")
	}
}

// ---------------------------------------------------------------------------
// ParseVersion edge cases
// ---------------------------------------------------------------------------

func TestParseVersion_InvalidPatch(t *testing.T) {
	_, err := core.ParseVersion("auth", "v1.0.X")
	if err == nil {
		t.Fatal("expected error for invalid patch version")
	}
}

// ---------------------------------------------------------------------------
// Client.Do with no Authorizer (ensure empty header is not set)
// ---------------------------------------------------------------------------

func TestClient_Do_NoAuthorizer(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := core.NewClient(core.WithBaseURL(server.URL + "/"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	resp.Body.Close()
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}

// ---------------------------------------------------------------------------
// Ensure Error.Error() formats correctly
// ---------------------------------------------------------------------------

func TestError_ErrorString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"code":"NotFound","message":"item not found"}]}`))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL + "/")
	err := core.NewAPIError(resp)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "NotFound") {
		t.Errorf("Error() = %q, expected to contain 'NotFound'", errStr)
	}
	if !strings.Contains(errStr, "404") {
		t.Errorf("Error() = %q, expected to contain status code 404", errStr)
	}
}

// ---------------------------------------------------------------------------
// ParseAPIVersion alias
// ---------------------------------------------------------------------------

func TestParseAPIVersion_Alias(t *testing.T) {
	v1, err1 := core.ParseVersion("auth", "v2")
	v2, err2 := core.ParseAPIVersion("auth", "v2")
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if v1.Major != v2.Major || v1.Minor != v2.Minor {
		t.Error("ParseVersion and ParseAPIVersion returned different results")
	}
}

// ---------------------------------------------------------------------------
// WithLogger when logger is nil (should not panic)
// ---------------------------------------------------------------------------

func TestWithLogger_Nil(t *testing.T) {
	c := core.NewClient(core.WithLogger(nil))
	// After WithLogger(nil), c.Logger may be nil; ensure no panic on subsequent use.
	_ = c.GetLogger()
}

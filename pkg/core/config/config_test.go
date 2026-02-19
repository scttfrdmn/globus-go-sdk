// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package config_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/config"
)

// TestDefaultConfigNonNil verifies DefaultConfig returns a non-nil Config.
func TestDefaultConfigNonNil(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
}

// TestDefaultConfigUserAgent verifies that the UserAgent is set.
func TestDefaultConfigUserAgent(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.UserAgent == "" {
		t.Fatal("DefaultConfig() UserAgent should not be empty")
	}
}

// TestDefaultConfigTimeout verifies Timeout is 30 seconds.
func TestDefaultConfigTimeout(t *testing.T) {
	cfg := config.DefaultConfig()
	expected := 30 * time.Second
	if cfg.Timeout != expected {
		t.Fatalf("DefaultConfig() Timeout = %v, want %v", cfg.Timeout, expected)
	}
}

// TestDefaultConfigRetryMax verifies RetryMax is 3.
func TestDefaultConfigRetryMax(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.RetryMax != 3 {
		t.Fatalf("DefaultConfig() RetryMax = %d, want 3", cfg.RetryMax)
	}
}

// TestDefaultConfigLogLevel verifies LogLevel defaults to LogLevelNone.
func TestDefaultConfigLogLevel(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.LogLevel != core.LogLevelNone {
		t.Fatalf("DefaultConfig() LogLevel = %v, want LogLevelNone", cfg.LogLevel)
	}
}

// TestDefaultConfigHTTPClientNonNil verifies HTTPClient is non-nil.
func TestDefaultConfigHTTPClientNonNil(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.HTTPClient == nil {
		t.Fatal("DefaultConfig() HTTPClient should not be nil")
	}
}

// TestFromEnvironmentBaseURL verifies GLOBUS_SDK_BASE_URL is applied.
func TestFromEnvironmentBaseURL(t *testing.T) {
	const key = "GLOBUS_SDK_BASE_URL"
	const want = "https://custom.api.example.com"

	t.Setenv(key, want)

	cfg := config.FromEnvironment()
	if cfg.BaseURL != want {
		t.Fatalf("FromEnvironment() BaseURL = %q, want %q", cfg.BaseURL, want)
	}
}

// TestFromEnvironmentUserAgent verifies GLOBUS_SDK_USER_AGENT is applied.
func TestFromEnvironmentUserAgent(t *testing.T) {
	const key = "GLOBUS_SDK_USER_AGENT"
	const want = "my-app/2.0"

	t.Setenv(key, want)

	cfg := config.FromEnvironment()
	if cfg.UserAgent != want {
		t.Fatalf("FromEnvironment() UserAgent = %q, want %q", cfg.UserAgent, want)
	}
}

// TestFromEnvironmentFallsBackToDefaults verifies that with no relevant env
// vars set the result matches DefaultConfig on the fields that matter.
func TestFromEnvironmentFallsBackToDefaults(t *testing.T) {
	// Unset both env vars for this test.
	os.Unsetenv("GLOBUS_SDK_BASE_URL")
	os.Unsetenv("GLOBUS_SDK_USER_AGENT")

	cfg := config.FromEnvironment()
	defaults := config.DefaultConfig()

	if cfg.BaseURL != defaults.BaseURL {
		t.Errorf("FromEnvironment() BaseURL = %q, want %q", cfg.BaseURL, defaults.BaseURL)
	}
	if cfg.Timeout != defaults.Timeout {
		t.Errorf("FromEnvironment() Timeout = %v, want %v", cfg.Timeout, defaults.Timeout)
	}
	if cfg.RetryMax != defaults.RetryMax {
		t.Errorf("FromEnvironment() RetryMax = %d, want %d", cfg.RetryMax, defaults.RetryMax)
	}
}

// TestApplyToClientBaseURL verifies that a non-empty BaseURL is set on the client.
func TestApplyToClientBaseURL(t *testing.T) {
	cfg := &config.Config{
		BaseURL:    "https://example.com",
		HTTPClient: &http.Client{},
	}
	client := core.NewClient()
	cfg.ApplyToClient(client)
	if client.BaseURL != "https://example.com" {
		t.Fatalf("ApplyToClient() BaseURL = %q, want %q", client.BaseURL, "https://example.com")
	}
}

// TestApplyToClientUserAgent verifies that UserAgent is applied.
func TestApplyToClientUserAgent(t *testing.T) {
	cfg := &config.Config{
		UserAgent:  "test-agent/1.0",
		HTTPClient: &http.Client{},
	}
	client := core.NewClient()
	cfg.ApplyToClient(client)
	if client.UserAgent != "test-agent/1.0" {
		t.Fatalf("ApplyToClient() UserAgent = %q, want %q", client.UserAgent, "test-agent/1.0")
	}
}

// TestApplyToClientHTTPClient verifies that HTTPClient is applied.
func TestApplyToClientHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	cfg := &config.Config{HTTPClient: custom}
	client := core.NewClient()
	cfg.ApplyToClient(client)
	if client.HTTPClient != custom {
		t.Fatal("ApplyToClient() did not set the custom HTTPClient")
	}
}

// TestApplyToClientDebug verifies that Debug flag is applied.
func TestApplyToClientDebug(t *testing.T) {
	cfg := &config.Config{
		Debug:      true,
		HTTPClient: &http.Client{},
	}
	client := core.NewClient()
	cfg.ApplyToClient(client)
	if !client.Debug {
		t.Fatal("ApplyToClient() did not set Debug = true")
	}
}

// TestApplyToClientTrace verifies that Trace flag is applied.
func TestApplyToClientTrace(t *testing.T) {
	cfg := &config.Config{
		Trace:      true,
		HTTPClient: &http.Client{},
	}
	client := core.NewClient()
	cfg.ApplyToClient(client)
	if !client.Trace {
		t.Fatal("ApplyToClient() did not set Trace = true")
	}
}

// TestApplyToClientNilDoesNotPanic verifies calling ApplyToClient with a nil
// client does not panic.
func TestApplyToClientNilDoesNotPanic(t *testing.T) {
	cfg := config.DefaultConfig()
	// Should not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ApplyToClient(nil) panicked: %v", r)
		}
	}()
	cfg.ApplyToClient(nil)
}

// TestApplyToClientNilCustomTransportNoOverride verifies that when
// CustomTransport is nil the existing transport on the client is not replaced.
func TestApplyToClientNilCustomTransportNoOverride(t *testing.T) {
	cfg := &config.Config{
		HTTPClient:      &http.Client{},
		CustomTransport: nil,
	}
	client := core.NewClient()
	originalTransport := client.Transport
	cfg.ApplyToClient(client)
	// Transport should remain unchanged because CustomTransport was nil.
	if client.Transport != originalTransport {
		t.Fatal("ApplyToClient() with nil CustomTransport should not override the existing transport")
	}
}

// TestGetSetVersionCheckRoundTrip verifies that SetVersionCheck / GetVersionCheck
// form a round-trip.
func TestGetSetVersionCheckRoundTrip(t *testing.T) {
	cfg := config.DefaultConfig()
	vc := core.NewVersionCheck()
	cfg.SetVersionCheck(vc)
	if got := cfg.GetVersionCheck(); got != vc {
		t.Fatal("GetVersionCheck did not return the value set by SetVersionCheck")
	}
}

// TestGetVersionCheckInitiallyNonNil verifies that DefaultConfig returns a
// non-nil VersionCheck.
func TestGetVersionCheckInitiallyNonNil(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.GetVersionCheck() == nil {
		t.Fatal("DefaultConfig() VersionCheck should not be nil")
	}
}

// TestSetVersionCheckNil verifies that SetVersionCheck(nil) is accepted and
// GetVersionCheck subsequently returns nil.
func TestSetVersionCheckNil(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SetVersionCheck(nil)
	if cfg.GetVersionCheck() != nil {
		t.Fatal("After SetVersionCheck(nil), GetVersionCheck should return nil")
	}
}

// TestDisableConnectionPoolEnvVar verifies that when
// GLOBUS_DISABLE_CONNECTION_POOL=true is set, DefaultConfig still returns a
// non-nil HTTPClient (but a plain one rather than a pooled one).
func TestDisableConnectionPoolEnvVar(t *testing.T) {
	t.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "true")

	cfg := config.DefaultConfig()
	if cfg.HTTPClient == nil {
		t.Fatal("DefaultConfig() with GLOBUS_DISABLE_CONNECTION_POOL=true should still return a non-nil HTTPClient")
	}
}

// TestDisableConnectionPoolEnvVarFalse verifies behaviour when the env var is
// not "true" – pooled client should still be non-nil.
func TestDisableConnectionPoolEnvVarFalse(t *testing.T) {
	t.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "false")

	cfg := config.DefaultConfig()
	if cfg.HTTPClient == nil {
		t.Fatal("DefaultConfig() HTTPClient should not be nil")
	}
}

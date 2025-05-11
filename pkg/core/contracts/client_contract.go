// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyClientContract verifies that a ClientInterface implementation
// satisfies the behavioral contract of the interface.
func VerifyClientContract(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	t.Run("Do method", func(t *testing.T) {
		verifyDoMethod(t, client)
	})

	t.Run("GetHTTPClient method", func(t *testing.T) {
		verifyGetHTTPClientMethod(t, client)
	})

	t.Run("GetBaseURL method", func(t *testing.T) {
		verifyGetBaseURLMethod(t, client)
	})

	t.Run("GetUserAgent method", func(t *testing.T) {
		verifyGetUserAgentMethod(t, client)
	})

	t.Run("GetLogger method", func(t *testing.T) {
		verifyGetLoggerMethod(t, client)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		verifyContextCancellation(t, client)
	})
}

// verifyDoMethod tests the behavior of the Do method
func verifyDoMethod(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// Create a test request
	req, err := http.NewRequest("GET", client.GetBaseURL()+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test with valid request
	resp, err := client.Do(context.Background(), req)

	// We can't assume the request will succeed since we don't control the URL
	// Instead, we verify that the method doesn't panic and returns a response or error
	if err != nil {
		t.Logf("Request failed with error: %v (this may be expected)", err)
	} else if resp == nil {
		t.Error("Request succeeded but returned nil response")
	} else {
		// Clean up
		resp.Body.Close()
	}

	// Test with nil request (should return error)
	_, err = client.Do(context.Background(), nil)
	if err == nil {
		t.Error("Do with nil request should return error")
	}
}

// verifyGetHTTPClientMethod tests the behavior of the GetHTTPClient method
func verifyGetHTTPClientMethod(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// GetHTTPClient should return a non-nil HTTP client
	httpClient := client.GetHTTPClient()
	if httpClient == nil {
		t.Error("GetHTTPClient returned nil")
	}

	// Subsequent calls should return the same HTTP client instance
	httpClient2 := client.GetHTTPClient()
	if httpClient != httpClient2 {
		t.Error("GetHTTPClient returned different instances on consecutive calls")
	}
}

// verifyGetBaseURLMethod tests the behavior of the GetBaseURL method
func verifyGetBaseURLMethod(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// GetBaseURL should return a non-empty string
	baseURL := client.GetBaseURL()
	if baseURL == "" {
		t.Error("GetBaseURL returned empty string")
	}

	// Subsequent calls should return the same value
	baseURL2 := client.GetBaseURL()
	if baseURL != baseURL2 {
		t.Errorf("GetBaseURL returned different values: %q and %q", baseURL, baseURL2)
	}
}

// verifyGetUserAgentMethod tests the behavior of the GetUserAgent method
func verifyGetUserAgentMethod(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// GetUserAgent should return a non-empty string
	userAgent := client.GetUserAgent()
	if userAgent == "" {
		t.Error("GetUserAgent returned empty string")
	}

	// Subsequent calls should return the same value
	userAgent2 := client.GetUserAgent()
	if userAgent != userAgent2 {
		t.Errorf("GetUserAgent returned different values: %q and %q", userAgent, userAgent2)
	}
}

// verifyGetLoggerMethod tests the behavior of the GetLogger method
func verifyGetLoggerMethod(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// GetLogger should return a non-nil logger
	logger := client.GetLogger()
	if logger == nil {
		t.Error("GetLogger returned nil")
	}

	// Verify logger methods don't panic
	logger.Debug("Debug test message")
	logger.Info("Info test message")
	logger.Warn("Warning test message")
	logger.Error("Error test message")
}

// verifyContextCancellation tests that the client respects context cancellation
func verifyContextCancellation(t *testing.T, client interfaces.ClientInterface) {
	t.Helper()

	// Create a context that cancels after a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep to ensure the context deadline has passed
	time.Sleep(2 * time.Millisecond)

	// Create a test request
	req, err := http.NewRequest("GET", client.GetBaseURL()+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// The request should fail with a context deadline exceeded error
	_, err = client.Do(ctx, req)
	if err == nil {
		t.Error("Request with canceled context should fail")
	}
}

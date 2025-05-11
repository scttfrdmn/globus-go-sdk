// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyTransportContract verifies that a Transport implementation
// satisfies the behavioral contract of the interface.
func VerifyTransportContract(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	t.Run("Request method", func(t *testing.T) {
		verifyRequestMethod(t, transport)
	})

	t.Run("Get method", func(t *testing.T) {
		verifyGetMethod(t, transport)
	})

	t.Run("Post method", func(t *testing.T) {
		verifyPostMethod(t, transport)
	})

	t.Run("Put method", func(t *testing.T) {
		verifyPutMethod(t, transport)
	})

	t.Run("Delete method", func(t *testing.T) {
		verifyDeleteMethod(t, transport)
	})

	t.Run("Patch method", func(t *testing.T) {
		verifyPatchMethod(t, transport)
	})

	t.Run("RoundTrip method", func(t *testing.T) {
		verifyRoundTripMethod(t, transport)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		verifyTransportContextCancellation(t, transport)
	})
}

// verifyRequestMethod tests the behavior of the Request method
func verifyRequestMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Test with minimal valid parameters
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}

	// Attempt a request - we don't care about the response, just that it doesn't panic
	// and handles errors properly
	resp, err := transport.Request(ctx, "GET", "/test", nil, query, headers)
	if err != nil {
		// Error is expected in most cases since we're not using a real server
		t.Logf("Request failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up if we got a response
		resp.Body.Close()
	}

	// Test with an invalid method
	_, err = transport.Request(ctx, "INVALID", "/test", nil, query, headers)
	if err == nil {
		t.Error("Request with invalid method should return error")
	}

	// Test with an empty path
	_, err = transport.Request(ctx, "GET", "", nil, query, headers)
	if err == nil {
		t.Error("Request with empty path should return error")
	}
}

// verifyGetMethod tests the behavior of the Get method
func verifyGetMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Get should properly forward to Request with the GET method
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}

	resp, err := transport.Get(ctx, "/test", query, headers)
	if err != nil {
		// Error is expected in most cases
		t.Logf("Get failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}
}

// verifyPostMethod tests the behavior of the Post method
func verifyPostMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Post should properly forward to Request with the POST method
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}
	body := strings.NewReader("test body")

	resp, err := transport.Post(ctx, "/test", body, query, headers)
	if err != nil {
		// Error is expected in most cases
		t.Logf("Post failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}
}

// verifyPutMethod tests the behavior of the Put method
func verifyPutMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Put should properly forward to Request with the PUT method
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}
	body := strings.NewReader("test body")

	resp, err := transport.Put(ctx, "/test", body, query, headers)
	if err != nil {
		// Error is expected in most cases
		t.Logf("Put failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}
}

// verifyDeleteMethod tests the behavior of the Delete method
func verifyDeleteMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Delete should properly forward to Request with the DELETE method
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}

	resp, err := transport.Delete(ctx, "/test", query, headers)
	if err != nil {
		// Error is expected in most cases
		t.Logf("Delete failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}
}

// verifyPatchMethod tests the behavior of the Patch method
func verifyPatchMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Patch should properly forward to Request with the PATCH method
	ctx := context.Background()
	query := url.Values{}
	headers := http.Header{}
	body := strings.NewReader("test body")

	resp, err := transport.Patch(ctx, "/test", body, query, headers)
	if err != nil {
		// Error is expected in most cases
		t.Logf("Patch failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}
}

// verifyRoundTripMethod tests the behavior of the RoundTrip method
func verifyRoundTripMethod(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Create a test request
	req, err := http.NewRequest("GET", "https://example.com/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test RoundTrip
	resp, err := transport.RoundTrip(req)
	if err != nil {
		// Error is expected in most cases
		t.Logf("RoundTrip failed with error: %v (this may be expected)", err)
	} else if resp != nil {
		// Clean up
		resp.Body.Close()
	}

	// Test with nil request (should return error)
	_, err = transport.RoundTrip(nil)
	if err == nil {
		t.Error("RoundTrip with nil request should return error")
	}
}

// verifyTransportContextCancellation tests that transport respects context cancellation
func verifyTransportContextCancellation(t *testing.T, transport interfaces.Transport) {
	t.Helper()

	// Create a context that cancels after a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Sleep to ensure the context deadline has passed
	time.Sleep(2 * time.Millisecond)

	// Attempt a request with the canceled context
	query := url.Values{}
	headers := http.Header{}

	_, err := transport.Request(ctx, "GET", "/test", nil, query, headers)
	if err == nil {
		t.Error("Request with canceled context should fail")
	} else if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "deadline") {
		t.Errorf("Expected context deadline error, got: %v", err)
	}
}

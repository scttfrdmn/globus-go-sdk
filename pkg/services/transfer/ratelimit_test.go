// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/ratelimit"
)

// mockLimiter is a mock implementation of RateLimiter for testing
type mockLimiter struct {
	waitCalls      int
	updateCalls    int
	shouldWait     bool
	waitDuration   time.Duration
	rateLimitError bool
	mu             sync.Mutex
}

func (m *mockLimiter) Wait(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.waitCalls++

	if m.shouldWait {
		time.Sleep(m.waitDuration)
	}

	if m.rateLimitError {
		return &TransferError{
			Code:       ErrCodeRateLimitExceeded,
			Message:    "Rate limit exceeded",
			StatusCode: http.StatusTooManyRequests,
		}
	}

	return nil
}

func (m *mockLimiter) Reserve() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.waitCalls++
	return m.waitDuration
}

func (m *mockLimiter) UpdateLimit(limit, remaining, resetAt int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCalls++
	return nil
}

func (m *mockLimiter) SetOptions(options *ratelimit.RateLimiterOptions) {
	// No-op for the mock
}

func (m *mockLimiter) GetStats() ratelimit.RateLimiterStats {
	return ratelimit.RateLimiterStats{
		CurrentLimit:    10,
		RemainingTokens: 5,
		TotalRequests:   int64(m.waitCalls),
		TotalThrottled:  int64(m.updateCalls),
		LastUpdated:     time.Now(),
	}
}

func (m *mockLimiter) GetWaitCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.waitCalls
}

func (m *mockLimiter) GetUpdateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateCalls
}

// TestClientWithRateLimiter tests that the client respects rate limiting
func TestClientWithRateLimiter(t *testing.T) {
	// Setup test server that returns rate limit headers
	handlerCalls := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		handlerCalls++
		currentCall := handlerCalls
		mu.Unlock()

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "50")
		w.Header().Set("X-RateLimit-Reset", "1609459200") // 2021-01-01 00:00:00 UTC

		// Return a 429 for the second request to test error handling
		if currentCall == 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"code":    "RateLimitExceeded",
				"message": "Rate limit exceeded. Try again in 60 seconds.",
			})
			return
		}

		// For other requests, return a successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create a mock limiter
	limiter := &mockLimiter{
		shouldWait:   true,
		waitDuration: 10 * time.Millisecond,
	}

	// Create a client with the mock limiter
	client, err := NewClient(
		WithAuthorizer(mockAuthorizer("test-token")),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
		WithCoreOption(core.WithRateLimiter(limiter)),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request that should use the rate limiter
	var result map[string]string
	reqErr := client.doRequestLowLevel(context.Background(), http.MethodGet, "endpoint_search", nil, nil, &result)

	// Verify the limiter was used
	if limiter.GetWaitCalls() != 1 {
		t.Errorf("Expected Wait to be called once, got %d calls", limiter.GetWaitCalls())
	}

	// Verify the request succeeded
	if reqErr != nil {
		t.Errorf("Expected no error, got %v", reqErr)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status to be 'ok', got %v", result["status"])
	}

	// Make a second request that should hit a rate limit error
	err2 := client.doRequestLowLevel(context.Background(), http.MethodGet, "endpoint_search", nil, nil, &result)

	// Verify the limiter was used again
	if limiter.GetWaitCalls() != 2 {
		t.Errorf("Expected Wait to be called twice, got %d calls", limiter.GetWaitCalls())
	}

	// Verify we got a rate limit error
	if err2 == nil {
		t.Errorf("Expected rate limit error, got nil")
	}

	if !IsRateLimitExceeded(err2) {
		t.Errorf("Expected IsRateLimitExceeded to return true for %v", err2)
	}
}

// TestRateLimitRetry tests retrying requests that hit rate limits
func TestRateLimitRetry(t *testing.T) {
	// Setup test server that returns rate limit errors then succeeds
	handlerCalls := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		handlerCalls++
		currentCall := handlerCalls
		mu.Unlock()

		// Return a 429 for the first two requests
		if currentCall <= 2 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1") // Retry after 1 second

			// Return the error in our standard format
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"code":      ErrCodeRateLimitExceeded,
				"message":   "Rate limit exceeded. Try again in 1 second.",
				"requestID": "test-request-id",
			})
			return
		}

		// For the third request, return a successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create a client with a token bucket limiter
	client, err := NewClient(
		WithAuthorizer(mockAuthorizer("test-token")),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
		WithCoreOption(core.WithRateLimiter(ratelimit.NewTokenBucketLimiter(nil))),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Define a retryable function
	retryFn := func(ctx context.Context) error {
		var result map[string]string
		doReqErr := client.doRequestLowLevel(ctx, http.MethodGet, "endpoint_search", nil, nil, &result)
		if doReqErr != nil {
			return doReqErr
		}

		if result["status"] != "ok" {
			return fmt.Errorf("unexpected status: %v", result["status"])
		}

		return nil
	}

	// Retry the function with exponential backoff
	strategy := ratelimit.NewExponentialBackoff(
		100*time.Millisecond, // Initial delay
		2*time.Second,        // Max delay
		2.0,                  // Factor
		5,                    // Max attempts
	)

	retryErr := ratelimit.RetryWithBackoff(ctx, retryFn, strategy, IsRetryableTransferError)

	// Verify the retry succeeded
	if retryErr != nil {
		t.Errorf("Expected no error after retrying, got %v", retryErr)
	}

	// Verify the server was called the expected number of times
	mu.Lock()
	if handlerCalls != 3 {
		t.Errorf("Expected 3 calls to the server, got %d", handlerCalls)
	}
	mu.Unlock()
}

// TestClientWithRateLimiterResponseHandling tests handling of rate limit information from responses
func TestClientWithRateLimiterResponseHandling(t *testing.T) {
	// Setup test server that returns rate limit headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "50")
		w.Header().Set("X-RateLimit-Reset", "1609459200") // 2021-01-01 00:00:00 UTC

		// Return a successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create a mock limiter to track updates
	limiter := &mockLimiter{}

	// Create a client with the mock limiter
	client, err := NewClient(
		WithAuthorizer(mockAuthorizer("test-token")),
		WithCoreOption(core.WithBaseURL(server.URL+"/")),
		WithCoreOption(core.WithRateLimiter(limiter)),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request that should update the rate limiter
	var result map[string]string
	reqErr := client.doRequestLowLevel(context.Background(), http.MethodGet, "endpoint_search", nil, nil, &result)

	// Verify the limiter was used and updated
	if limiter.GetWaitCalls() != 1 {
		t.Errorf("Expected Wait to be called once, got %d calls", limiter.GetWaitCalls())
	}

	if limiter.GetUpdateCalls() != 1 {
		t.Errorf("Expected UpdateLimit to be called once, got %d calls", limiter.GetUpdateCalls())
	}

	// Verify the request succeeded
	if reqErr != nil {
		t.Errorf("Expected no error, got %v", reqErr)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status to be 'ok', got %v", result["status"])
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	// Create a backoff strategy with deterministic parameters
	backoff := NewExponentialBackoff(
		100*time.Millisecond, // Initial delay
		5*time.Second,        // Max delay
		2.0,                  // Factor
		5,                    // Max attempts
	)
	
	// Disable jitter for deterministic testing
	backoff.Jitter = false
	
	// Test initial backoff
	delay := backoff.NextBackoff(1)
	if delay != 100*time.Millisecond {
		t.Errorf("Expected initial delay of 100ms, got %v", delay)
	}
	
	// Test second attempt (should be doubled)
	delay = backoff.NextBackoff(2)
	if delay != 200*time.Millisecond {
		t.Errorf("Expected second delay of 200ms, got %v", delay)
	}
	
	// Test third attempt (should be doubled again)
	delay = backoff.NextBackoff(3)
	if delay != 400*time.Millisecond {
		t.Errorf("Expected third delay of 400ms, got %v", delay)
	}
	
	// Test that max delay is respected
	delay = backoff.NextBackoff(10)
	if delay != 5*time.Second {
		t.Errorf("Expected delay to be capped at 5s, got %v", delay)
	}
	
	// Test with jitter
	backoff.Jitter = true
	backoff.JitterFactor = 0.5
	
	delay = backoff.NextBackoff(1)
	if delay < 100*time.Millisecond || delay > 150*time.Millisecond {
		t.Errorf("Expected delay with jitter to be between 100-150ms, got %v", delay)
	}
	
	// Test Max Attempts
	if backoff.MaxAttempts() != 5 {
		t.Errorf("Expected max attempts to be 5, got %d", backoff.MaxAttempts())
	}
}

func TestRetryWithBackoff(t *testing.T) {
	// Create a backoff strategy with fast timings for testing
	backoff := NewExponentialBackoff(
		10*time.Millisecond, // Initial delay
		50*time.Millisecond, // Max delay
		2.0,                 // Factor
		3,                   // Max attempts
	)
	
	// Disable jitter for deterministic testing
	backoff.Jitter = false
	
	// Count attempts
	attempts := 0
	
	// Function that fails the first two times, then succeeds
	fn := func(ctx context.Context) error {
		attempts++
		if attempts <= 2 {
			return errors.New("temporary error")
		}
		return nil
	}
	
	// Always retry
	shouldRetry := func(err error) bool {
		return true
	}
	
	ctx := context.Background()
	err := RetryWithBackoff(ctx, fn, backoff, shouldRetry)
	
	if err != nil {
		t.Errorf("Expected successful retry, got error: %v", err)
	}
	
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
	
	// Test with max attempts exceeded
	attempts = 0
	fn = func(ctx context.Context) error {
		attempts++
		return errors.New("persistent error")
	}
	
	err = RetryWithBackoff(ctx, fn, backoff, shouldRetry)
	
	if err == nil {
		t.Error("Expected error after max retries, got nil")
	}
	
	if attempts != 4 { // Initial attempt + 3 retries
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
	
	// Test with context cancellation
	attempts = 0
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	
	fn = func(ctx context.Context) error {
		attempts++
		return errors.New("error with slow retry")
	}
	
	start := time.Now()
	err = RetryWithBackoff(ctx, fn, backoff, shouldRetry)
	elapsed := time.Since(start)
	
	if attempts > 2 {
		t.Errorf("Expected at most 2 attempts due to context timeout, got %d", attempts)
	}
	
	if elapsed < 20*time.Millisecond {
		t.Errorf("Expected to wait at least 20ms before cancellation, waited %v", elapsed)
	}
}

func TestIsRetryableError(t *testing.T) {
	// Test retryable errors
	retryableErrors := []string{
		"connection refused",
		"timeout exceeded",
		"too many requests",
		"please retry",
		"rate limit exceeded",
		"server error occurred",
		"bad gateway",
		"service temporarily unavailable",
		"internal server error occurred",
		"gateway timeout",
		"request timeout",
		"unexpected EOF",
		"connection reset by peer",
	}
	
	for _, errStr := range retryableErrors {
		err := errors.New(errStr)
		if !IsRetryableError(err) {
			t.Errorf("Expected error '%s' to be retryable", errStr)
		}
	}
	
	// Test non-retryable errors
	nonRetryableErrors := []string{
		"not found",
		"unauthorized",
		"bad request",
		"forbidden",
		"invalid input",
		"validation failed",
	}
	
	for _, errStr := range nonRetryableErrors {
		err := errors.New(errStr)
		if IsRetryableError(err) {
			t.Errorf("Expected error '%s' to NOT be retryable", errStr)
		}
	}
	
	// Test nil error
	if IsRetryableError(nil) {
		t.Error("Expected nil error to NOT be retryable")
	}
}
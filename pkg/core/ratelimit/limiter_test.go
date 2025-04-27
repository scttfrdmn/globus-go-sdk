// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	// Create a limiter with 10 requests per second and a burst of 5
	options := &RateLimiterOptions{
		RequestsPerSecond: 10,
		BurstSize:         5,
		UseAdaptive:       false,
	}
	
	limiter := NewTokenBucketLimiter(options)
	
	// Test initial state
	stats := limiter.GetStats()
	if stats.CurrentLimit != 10 {
		t.Errorf("Expected current limit to be 10, got %f", stats.CurrentLimit)
	}
	
	if stats.RemainingTokens != 5 {
		t.Errorf("Expected remaining tokens to be 5, got %f", stats.RemainingTokens)
	}
	
	// Test consuming tokens
	ctx := context.Background()
	
	// Should be able to make 5 requests immediately
	for i := 0; i < 5; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Unexpected error on request %d: %v", i+1, err)
		}
	}
	
	// The next request should be rate limited
	start := time.Now()
	err := limiter.Wait(ctx)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Errorf("Unexpected error on rate limited request: %v", err)
	}
	
	// Should have waited roughly 100ms (1/10th of a second)
	if elapsed < 80*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected to wait roughly 100ms, waited %v", elapsed)
	}
	
	// Check stats again
	stats = limiter.GetStats()
	if stats.TotalRequests != 6 {
		t.Errorf("Expected total requests to be 6, got %d", stats.TotalRequests)
	}
	
	if stats.TotalThrottled != 1 {
		t.Errorf("Expected total throttled to be 1, got %d", stats.TotalThrottled)
	}
}

func TestTokenBucketLimiterAdaptive(t *testing.T) {
	// Create a limiter with adaptive rate limiting
	options := &RateLimiterOptions{
		RequestsPerSecond: 20,
		BurstSize:         10,
		UseAdaptive:       true,
	}
	
	limiter := NewTokenBucketLimiter(options)
	
	// Update limit based on API response
	limit := 5      // 5 requests per period
	remaining := 3  // 3 requests remaining
	resetAt := int(time.Now().Add(10 * time.Second).Unix()) // Reset in 10 seconds
	
	err := limiter.UpdateLimit(limit, remaining, resetAt)
	if err != nil {
		t.Errorf("Failed to update limit: %v", err)
	}
	
	// Should allow the remaining requests immediately
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Unexpected error on request %d: %v", i+1, err)
		}
	}
	
	// The next request should be rate limited
	start := time.Now()
	err = limiter.Wait(ctx)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Errorf("Unexpected error on rate limited request: %v", err)
	}
	
	// Should have waited longer due to adaptive rate limiting
	if elapsed < 1*time.Second {
		t.Errorf("Expected to wait at least 1 second with adaptive limiting, waited %v", elapsed)
	}
}

func TestTokenBucketLimiterCancelContext(t *testing.T) {
	// Create a limiter with 1 request per second
	options := &RateLimiterOptions{
		RequestsPerSecond: 1,
		BurstSize:         1,
	}
	
	limiter := NewTokenBucketLimiter(options)
	
	// Consume the first token
	ctx := context.Background()
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Unexpected error on first request: %v", err)
	}
	
	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Try to make another request, which should be canceled
	start := time.Now()
	err = limiter.Wait(ctx)
	elapsed := time.Since(start)
	
	if err == nil {
		t.Error("Expected context timeout error, got nil")
	}
	
	if elapsed < 50*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Errorf("Expected to wait roughly 100ms before cancellation, waited %v", elapsed)
	}
}

func TestNoopRateLimiter(t *testing.T) {
	limiter := NewNoopRateLimiter()
	ctx := context.Background()
	
	// Should be able to make many requests immediately
	for i := 0; i < 100; i++ {
		start := time.Now()
		err := limiter.Wait(ctx)
		elapsed := time.Since(start)
		
		if err != nil {
			t.Errorf("Unexpected error on request %d: %v", i+1, err)
		}
		
		if elapsed > 5*time.Millisecond {
			t.Errorf("NoopRateLimiter should not block, waited %v", elapsed)
		}
	}
	
	// Check stats
	stats := limiter.GetStats()
	if stats.TotalRequests != 100 {
		t.Errorf("Expected total requests to be 100, got %d", stats.TotalRequests)
	}
	
	if stats.TotalThrottled != 0 {
		t.Errorf("Expected total throttled to be 0, got %d", stats.TotalThrottled)
	}
	
	if stats.RemainingTokens != -1 {
		t.Errorf("Expected remaining tokens to be -1 (unlimited), got %f", stats.RemainingTokens)
	}
}
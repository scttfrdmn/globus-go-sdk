// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"net/http"
	"testing"
	"time"
)

func TestExtractRateLimitInfo(t *testing.T) {
	// Create a test response with rate limit headers
	resp := &http.Response{
		Header: http.Header{},
	}
	
	// Standard rate limit headers
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "75")
	resp.Header.Set("X-RateLimit-Reset", "1609459200") // 2021-01-01 00:00:00 UTC
	
	// Extract info
	info, found := ExtractRateLimitInfo(resp)
	
	if !found {
		t.Error("Expected to find rate limit info, but none was found")
	}
	
	if info.Limit != 100 {
		t.Errorf("Expected limit to be 100, got %d", info.Limit)
	}
	
	if info.Remaining != 75 {
		t.Errorf("Expected remaining to be 75, got %d", info.Remaining)
	}
	
	if info.Reset != 1609459200 {
		t.Errorf("Expected reset to be 1609459200, got %d", info.Reset)
	}
	
	// Test with Retry-After header (seconds)
	resp.Header = http.Header{}
	resp.Header.Set("Retry-After", "30")
	
	info, found = ExtractRateLimitInfo(resp)
	
	if !found {
		t.Error("Expected to find rate limit info, but none was found")
	}
	
	if info.Retry != 30 {
		t.Errorf("Expected retry to be 30, got %d", info.Retry)
	}
	
	// Test with Retry-After header (HTTP date)
	futureTime := time.Now().Add(2 * time.Minute).Format(time.RFC1123)
	resp.Header = http.Header{}
	resp.Header.Set("Retry-After", futureTime)
	
	info, found = ExtractRateLimitInfo(resp)
	
	if !found {
		t.Error("Expected to find rate limit info, but none was found")
	}
	
	if info.Retry < 115 || info.Retry > 125 { // About 2 minutes (120 seconds)
		t.Errorf("Expected retry to be about 120 seconds, got %d", info.Retry)
	}
	
	// Test with Globus-specific headers
	resp.Header = http.Header{}
	resp.Header.Set("Server", "Globus API")
	resp.Header.Set("X-Globus-RateLimit-Limit", "50")
	resp.Header.Set("X-Globus-RateLimit-Remaining", "25")
	resp.Header.Set("X-Globus-RateLimit-Reset", "1609459200")
	
	info, found = ExtractRateLimitInfo(resp)
	
	if !found {
		t.Error("Expected to find rate limit info, but none was found")
	}
	
	if info.Limit != 50 {
		t.Errorf("Expected limit to be 50, got %d", info.Limit)
	}
	
	if info.Remaining != 25 {
		t.Errorf("Expected remaining to be 25, got %d", info.Remaining)
	}
	
	// Test with no rate limit headers
	resp.Header = http.Header{}
	
	info, found = ExtractRateLimitInfo(resp)
	
	if found {
		t.Error("Expected to not find rate limit info, but info was found")
	}
	
	// Test with nil response
	info, found = ExtractRateLimitInfo(nil)
	
	if found {
		t.Error("Expected to not find rate limit info with nil response, but info was found")
	}
}

func TestUpdateRateLimiterFromResponse(t *testing.T) {
	// Create a test limiter
	limiter := NewTokenBucketLimiter(&RateLimiterOptions{
		RequestsPerSecond: 10,
		BurstSize:         5,
		UseAdaptive:       true,
	})
	
	// Create a test response with rate limit headers
	resp := &http.Response{
		Header: http.Header{},
	}
	
	resp.Header.Set("X-RateLimit-Limit", "100")
	resp.Header.Set("X-RateLimit-Remaining", "75")
	resp.Header.Set("X-RateLimit-Reset", "1609459200")
	
	// Update limiter from response
	updated := UpdateRateLimiterFromResponse(limiter, resp)
	
	if !updated {
		t.Error("Expected limiter to be updated, but it was not")
	}
	
	// Test with nil limiter
	updated = UpdateRateLimiterFromResponse(nil, resp)
	
	if updated {
		t.Error("Expected update to fail with nil limiter, but it succeeded")
	}
	
	// Test with nil response
	updated = UpdateRateLimiterFromResponse(limiter, nil)
	
	if updated {
		t.Error("Expected update to fail with nil response, but it succeeded")
	}
	
	// Test with no rate limit headers
	resp.Header = http.Header{}
	
	updated = UpdateRateLimiterFromResponse(limiter, resp)
	
	if updated {
		t.Error("Expected update to fail with no rate limit headers, but it succeeded")
	}
}
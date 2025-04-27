// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ResponseRateLimitInfo contains rate limit information from HTTP response headers
type ResponseRateLimitInfo struct {
	// Limit is the maximum number of requests allowed in the period
	Limit int
	
	// Remaining is the number of requests remaining in the period
	Remaining int
	
	// Reset is the time when the rate limit will reset (Unix timestamp)
	Reset int
	
	// Window is the duration of the rate limit period in seconds
	Window int
	
	// Retry is the suggested retry time in seconds
	Retry int
}

// ExtractRateLimitInfo extracts rate limit information from HTTP response headers
func ExtractRateLimitInfo(resp *http.Response) (*ResponseRateLimitInfo, bool) {
	if resp == nil {
		return nil, false
	}
	
	// Initialize info
	info := &ResponseRateLimitInfo{}
	found := false
	
	// Extract standard rate limit headers
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			info.Limit = val
			found = true
		}
	}
	
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			info.Remaining = val
			found = true
		}
	}
	
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if val, err := strconv.Atoi(reset); err == nil {
			info.Reset = val
			found = true
		}
	}
	
	// Check for additional headers
	if window := resp.Header.Get("X-RateLimit-Window"); window != "" {
		if val, err := strconv.Atoi(window); err == nil {
			info.Window = val
		}
	}
	
	// Check for Retry-After header (used in 429 responses)
	if retry := resp.Header.Get("Retry-After"); retry != "" {
		// Retry-After can be either seconds or a HTTP date
		if val, err := strconv.Atoi(retry); err == nil {
			// It's a number of seconds
			info.Retry = val
			found = true
		} else {
			// Try to parse as HTTP date
			if t, err := http.ParseTime(retry); err == nil {
				info.Retry = int(time.Until(t).Seconds())
				if info.Retry < 0 {
					info.Retry = 0
				}
				found = true
			}
		}
	}
	
	// Check for Globus-specific headers
	if strings.Contains(resp.Header.Get("Server"), "Globus") {
		// Some Globus APIs use custom headers
		if limit := resp.Header.Get("X-Globus-RateLimit-Limit"); limit != "" {
			if val, err := strconv.Atoi(limit); err == nil {
				info.Limit = val
				found = true
			}
		}
		
		if remaining := resp.Header.Get("X-Globus-RateLimit-Remaining"); remaining != "" {
			if val, err := strconv.Atoi(remaining); err == nil {
				info.Remaining = val
				found = true
			}
		}
		
		if reset := resp.Header.Get("X-Globus-RateLimit-Reset"); reset != "" {
			if val, err := strconv.Atoi(reset); err == nil {
				info.Reset = val
				found = true
			}
		}
	}
	
	return info, found
}

// UpdateRateLimiterFromResponse updates a rate limiter based on response headers
func UpdateRateLimiterFromResponse(limiter RateLimiter, resp *http.Response) bool {
	if limiter == nil || resp == nil {
		return false
	}
	
	info, found := ExtractRateLimitInfo(resp)
	if !found {
		return false
	}
	
	// Update the rate limiter with the extracted information
	err := limiter.UpdateLimit(info.Limit, info.Remaining, info.Reset)
	
	return err == nil
}
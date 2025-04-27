// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides rate limiting capabilities for API requests
type RateLimiter interface {
	// Wait blocks until a request can be made according to the rate limit
	Wait(ctx context.Context) error
	
	// Reserve reserves a token without blocking and returns the time to wait
	Reserve() time.Duration
	
	// UpdateLimit updates the rate limit based on API response headers or other factors
	UpdateLimit(limit, remaining, resetAt int) error
	
	// SetOptions configures the rate limiter
	SetOptions(options *RateLimiterOptions)
	
	// GetStats returns current statistics about the rate limiter's usage
	GetStats() RateLimiterStats
}

// RateLimiterOptions contains configuration for rate limiters
type RateLimiterOptions struct {
	// RequestsPerSecond is the default rate limit (requests per second)
	RequestsPerSecond float64
	
	// BurstSize is the maximum number of requests that can be made at once
	BurstSize int
	
	// UseAdaptive enables adaptive rate limiting based on response headers
	UseAdaptive bool
	
	// MaxRetryCount is the maximum number of retries for a failed request
	MaxRetryCount int
	
	// MinRetryDelay is the minimum delay between retries
	MinRetryDelay time.Duration
	
	// MaxRetryDelay is the maximum delay between retries
	MaxRetryDelay time.Duration
	
	// UseJitter adds random jitter to retry delays to prevent thundering herd problems
	UseJitter bool
	
	// JitterFactor is the maximum percentage of jitter to add (0.0-1.0)
	JitterFactor float64
}

// DefaultRateLimiterOptions returns the default options for a rate limiter
func DefaultRateLimiterOptions() *RateLimiterOptions {
	return &RateLimiterOptions{
		RequestsPerSecond: 10.0,
		BurstSize:         20,
		UseAdaptive:       true,
		MaxRetryCount:     5,
		MinRetryDelay:     100 * time.Millisecond,
		MaxRetryDelay:     60 * time.Second,
		UseJitter:         true,
		JitterFactor:      0.2,
	}
}

// RateLimiterStats contains statistics about a rate limiter's usage
type RateLimiterStats struct {
	// CurrentLimit is the current rate limit (requests per second)
	CurrentLimit float64
	
	// RemainingTokens is the number of tokens currently available
	RemainingTokens float64
	
	// ResetAt is the time when the rate limit will reset
	ResetAt time.Time
	
	// TotalWaitTime is the total time spent waiting for rate limiting
	TotalWaitTime time.Duration
	
	// TotalRequests is the total number of requests processed
	TotalRequests int64
	
	// TotalThrottled is the number of requests that were throttled
	TotalThrottled int64
	
	// LastUpdated is the time when these stats were last updated
	LastUpdated time.Time
}

// TokenBucketLimiter implements a token bucket rate limiter
type TokenBucketLimiter struct {
	mu             sync.Mutex
	tokens         float64
	rate           float64
	burst          int
	lastCheck      time.Time
	options        *RateLimiterOptions
	stats          RateLimiterStats
	resetsAt       time.Time
	quotaRemaining int
	quotaLimit     int
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(options *RateLimiterOptions) *TokenBucketLimiter {
	if options == nil {
		options = DefaultRateLimiterOptions()
	}
	
	now := time.Now()
	
	return &TokenBucketLimiter{
		tokens:    float64(options.BurstSize),
		rate:      options.RequestsPerSecond,
		burst:     options.BurstSize,
		lastCheck: now,
		options:   options,
		stats: RateLimiterStats{
			CurrentLimit:   options.RequestsPerSecond,
			RemainingTokens: float64(options.BurstSize),
			LastUpdated:    now,
		},
	}
}

// Wait blocks until a token is available or the context is canceled
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	waitTime := l.reserve()
	
	// If no wait is needed, return immediately
	if waitTime <= 0 {
		return nil
	}
	
	// Record stats
	l.mu.Lock()
	l.stats.TotalWaitTime += waitTime
	l.stats.TotalThrottled++
	l.mu.Unlock()
	
	// Wait for the required time or until context is canceled
	timer := time.NewTimer(waitTime)
	defer timer.Stop()
	
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		// Return the token since we're not going to use it
		l.mu.Lock()
		l.tokens++
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
		l.mu.Unlock()
		
		return ctx.Err()
	}
}

// Reserve reserves a token and returns the time to wait
func (l *TokenBucketLimiter) Reserve() time.Duration {
	return l.reserve()
}

// reserve is the internal implementation of Reserve
func (l *TokenBucketLimiter) reserve() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	now := time.Now()
	
	// Add tokens based on time elapsed since last check
	elapsed := now.Sub(l.lastCheck).Seconds()
	l.lastCheck = now
	newTokens := elapsed * l.rate
	
	// Check if we need to respect API response limits
	if l.options.UseAdaptive && !l.resetsAt.IsZero() && now.Before(l.resetsAt) {
		// If we have quota info from the API, use that to adjust token generation
		if l.quotaLimit > 0 {
			// Calculate tokens based on remaining quota and time until reset
			timeUntilReset := l.resetsAt.Sub(now).Seconds()
			if timeUntilReset > 0 {
				// Allow using the remaining quota evenly until reset
				adjustedRate := float64(l.quotaRemaining) / timeUntilReset
				// Use the more conservative of our configured rate or the API-based rate
				if adjustedRate < l.rate {
					newTokens = elapsed * adjustedRate
				}
			}
		}
	}
	
	l.tokens += newTokens
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
	
	// Update stats
	l.stats.RemainingTokens = l.tokens
	l.stats.LastUpdated = now
	l.stats.TotalRequests++
	
	// If we have tokens, consume one and return
	if l.tokens >= 1 {
		l.tokens--
		return 0
	}
	
	// Otherwise, calculate wait time for the next token
	waitTime := (1 - l.tokens) / l.rate
	return time.Duration(waitTime * float64(time.Second))
}

// UpdateLimit updates the rate limit based on API response headers
func (l *TokenBucketLimiter) UpdateLimit(limit, remaining, resetAt int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Store the quota information
	l.quotaLimit = limit
	l.quotaRemaining = remaining
	
	// Calculate reset time
	l.resetsAt = time.Unix(int64(resetAt), 0)
	
	// Update stats
	l.stats.ResetAt = l.resetsAt
	
	return nil
}

// SetOptions updates the limiter's configuration
func (l *TokenBucketLimiter) SetOptions(options *RateLimiterOptions) {
	if options == nil {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Only update the rate if it's changing
	if l.rate != options.RequestsPerSecond {
		l.rate = options.RequestsPerSecond
		l.stats.CurrentLimit = options.RequestsPerSecond
	}
	
	// Update burst size if it's changing
	if l.burst != options.BurstSize {
		// If new burst size is larger, add tokens up to the new limit
		if options.BurstSize > l.burst {
			l.tokens += float64(options.BurstSize - l.burst)
			if l.tokens > float64(options.BurstSize) {
				l.tokens = float64(options.BurstSize)
			}
		}
		
		l.burst = options.BurstSize
	}
	
	l.options = options
}

// GetStats returns current statistics about the rate limiter
func (l *TokenBucketLimiter) GetStats() RateLimiterStats {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Make a copy of the stats to return
	stats := l.stats
	
	// Update the remaining tokens to the current value
	now := time.Now()
	if now.After(l.lastCheck) {
		elapsed := now.Sub(l.lastCheck).Seconds()
		newTokens := elapsed * l.rate
		
		currentTokens := l.tokens + newTokens
		if currentTokens > float64(l.burst) {
			currentTokens = float64(l.burst)
		}
		
		stats.RemainingTokens = currentTokens
		stats.LastUpdated = now
	}
	
	return stats
}

// NoopRateLimiter implements RateLimiter but performs no rate limiting
type NoopRateLimiter struct {
	stats RateLimiterStats
}

// NewNoopRateLimiter creates a new NoopRateLimiter that doesn't limit requests
func NewNoopRateLimiter() *NoopRateLimiter {
	return &NoopRateLimiter{
		stats: RateLimiterStats{
			CurrentLimit:    -1, // Unlimited
			RemainingTokens: -1, // Unlimited
			LastUpdated:     time.Now(),
		},
	}
}

// Wait returns immediately without blocking
func (l *NoopRateLimiter) Wait(ctx context.Context) error {
	l.stats.TotalRequests++
	return nil
}

// Reserve returns 0 indicating no wait time
func (l *NoopRateLimiter) Reserve() time.Duration {
	l.stats.TotalRequests++
	return 0
}

// UpdateLimit is a no-op
func (l *NoopRateLimiter) UpdateLimit(limit, remaining, resetAt int) error {
	return nil
}

// SetOptions is a no-op
func (l *NoopRateLimiter) SetOptions(options *RateLimiterOptions) {
	// No-op
}

// GetStats returns basic stats
func (l *NoopRateLimiter) GetStats() RateLimiterStats {
	l.stats.LastUpdated = time.Now()
	return l.stats
}
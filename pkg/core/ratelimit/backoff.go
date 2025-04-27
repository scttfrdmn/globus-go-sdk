// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy defines a strategy for calculating retry delays
type BackoffStrategy interface {
	// NextBackoff returns the next backoff duration
	NextBackoff(attempt int) time.Duration
	
	// Reset resets the backoff state
	Reset()
	
	// MaxAttempts returns the maximum number of retry attempts
	MaxAttempts() int
}

// ExponentialBackoff implements an exponential backoff strategy
type ExponentialBackoff struct {
	// InitialDelay is the delay for the first retry
	InitialDelay time.Duration
	
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	
	// Factor is the multiplier applied to the delay after each attempt
	Factor float64
	
	// Jitter is whether to add random jitter to the delay
	Jitter bool
	
	// JitterFactor is the maximum percentage of jitter to add (0.0-1.0)
	JitterFactor float64
	
	// MaxAttempt is the maximum number of retry attempts
	MaxAttempt int
	
	// rand is the random number generator used for jitter
	rand *rand.Rand
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, factor float64, maxAttempts int) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		Factor:        factor,
		Jitter:        true,
		JitterFactor:  0.2,
		MaxAttempt:    maxAttempts,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextBackoff calculates the next backoff duration
func (b *ExponentialBackoff) NextBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	
	// Calculate exponential backoff
	delayFloat := float64(b.InitialDelay) * math.Pow(b.Factor, float64(attempt-1))
	delay := time.Duration(delayFloat)
	
	// Cap delay at maximum
	if delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	
	// Add jitter if enabled
	if b.Jitter && b.JitterFactor > 0 {
		jitter := float64(delay) * b.JitterFactor * b.rand.Float64()
		delay = delay + time.Duration(jitter)
	}
	
	return delay
}

// Reset resets the backoff state
func (b *ExponentialBackoff) Reset() {
	// Reset the random number generator
	b.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// MaxAttempts returns the maximum number of retry attempts
func (b *ExponentialBackoff) MaxAttempts() int {
	return b.MaxAttempt
}

// DefaultBackoff creates a backoff strategy with sensible defaults
func DefaultBackoff() BackoffStrategy {
	return NewExponentialBackoff(
		100*time.Millisecond, // Initial delay
		60*time.Second,       // Max delay
		2.0,                  // Factor
		5,                    // Max attempts
	)
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryWithBackoff executes a function with retry using the given backoff strategy
func RetryWithBackoff(
	ctx context.Context,
	fn RetryableFunc,
	strategy BackoffStrategy,
	shouldRetry func(error) bool,
) error {
	var lastErr error
	
	// Reset backoff strategy
	strategy.Reset()
	
	for attempt := 0; attempt <= strategy.MaxAttempts(); attempt++ {
		// For the first attempt, don't wait
		if attempt > 0 {
			delay := strategy.NextBackoff(attempt)
			
			// Create a timer for the delay
			timer := time.NewTimer(delay)
			
			// Wait for either the timer or context cancellation
			select {
			case <-timer.C:
				// Continue with retry
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
		
		// Execute the function
		err := fn(ctx)
		if err == nil {
			// Success
			return nil
		}
		
		lastErr = err
		
		// Check if we should retry
		if shouldRetry != nil && !shouldRetry(err) {
			// Don't retry this error
			return err
		}
		
		// If we've reached the maximum attempts, return the last error
		if attempt == strategy.MaxAttempts() {
			return lastErr
		}
	}
	
	// This should not be reached
	return lastErr
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error) bool {
	// Network errors, timeouts, and certain HTTP status codes
	// (429, 500, 502, 503, 504) are retryable
	
	// This is a simplified check. In a real implementation,
	// you would check for specific error types and HTTP status codes.
	
	if err == nil {
		return false
	}
	
	// Sample check - in a real implementation, use proper error type assertions
	errStr := err.Error()
	
	// Check for common retryable error strings
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"too many requests",
		"retry",
		"rate limit",
		"server error",
		"gateway",
		"temporarily unavailable",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"request timeout",
		"EOF",
		"connection reset",
		"use of closed network connection",
		"i/o timeout",
	}
	
	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}
	
	return false
}

// contains checks if a string contains another string (case-insensitive)
func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) {
			match := true
			for j := 0; j < len(substr); j++ {
				if toLower(s[i+j]) != toLower(substr[j]) {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// toLower converts a character to lowercase
func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}
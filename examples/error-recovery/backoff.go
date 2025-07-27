// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package main

import (
	"math"
	"math/rand"
	"time"
)

// ExponentialBackoff implements an exponential backoff strategy with jitter
type ExponentialBackoff struct {
	initialBackoff time.Duration
	maxBackoff     time.Duration
	backoffFactor  float64
	jitter         float64
	attempt        int
}

// BackoffOption defines a configuration option for the backoff strategy
type BackoffOption func(*ExponentialBackoff)

// WithInitialBackoff sets the initial backoff delay
func WithInitialBackoff(duration time.Duration) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.initialBackoff = duration
	}
}

// WithMaxBackoff sets the maximum backoff delay
func WithMaxBackoff(duration time.Duration) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.maxBackoff = duration
	}
}

// WithBackoffFactor sets the multiplier for each retry attempt
func WithBackoffFactor(factor float64) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.backoffFactor = factor
	}
}

// WithJitter sets the jitter factor (0.0-1.0) to randomize delays
func WithJitter(jitter float64) BackoffOption {
	return func(b *ExponentialBackoff) {
		b.jitter = jitter
	}
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(options ...BackoffOption) *ExponentialBackoff {
	b := &ExponentialBackoff{
		initialBackoff: 100 * time.Millisecond, // Default: 100ms initial backoff
		maxBackoff:     60 * time.Second,       // Default: 60s max backoff
		backoffFactor:  2.0,                    // Default: Double the backoff each time
		jitter:         0.2,                    // Default: 20% jitter
		attempt:        0,
	}

	// Apply options
	for _, option := range options {
		option(b)
	}

	return b
}

// NextBackoff calculates the next backoff duration
func (b *ExponentialBackoff) NextBackoff() time.Duration {
	b.attempt++

	// Calculate base delay using exponential backoff formula
	// delay = initialBackoff * (backoffFactor ^ attempt)
	backoff := float64(b.initialBackoff) * math.Pow(b.backoffFactor, float64(b.attempt-1))

	// Apply jitter to prevent thundering herd problem
	if b.jitter > 0 {
		// Add randomness between -jitter/2 and +jitter/2
		jitterFactor := 1.0 + (rand.Float64()*b.jitter - b.jitter/2.0)
		backoff = backoff * jitterFactor
	}

	// Ensure backoff doesn't exceed max
	if backoff > float64(b.maxBackoff) {
		backoff = float64(b.maxBackoff)
	}

	return time.Duration(backoff)
}

// Reset resets the attempt counter back to zero
func (b *ExponentialBackoff) Reset() {
	b.attempt = 0
}

// GetAttempt returns the current attempt count
func (b *ExponentialBackoff) GetAttempt() int {
	return b.attempt
}

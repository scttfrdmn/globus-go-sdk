// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package ratelimit provides rate limiting and retry utilities for the Globus Go SDK.

This package implements token-bucket rate limiting, exponential backoff retry
strategies, and circuit breaking to improve the resilience of Globus API clients.
It is designed to work with API response headers (X-RateLimit-Limit,
X-RateLimit-Remaining, X-RateLimit-Reset) for adaptive throttling.

# STABILITY: BETA

This package is in beta. The core abstractions are stable, but some configuration
details and secondary APIs may undergo minor backward-incompatible changes in minor
releases. Changes will be documented in the CHANGELOG with migration guidance.

The following components are considered beta-stable:

  - RateLimiter interface (Wait, Reserve, UpdateLimit, SetOptions, GetStats)
  - RateLimiterOptions struct and all exported fields
  - RateLimiterStats struct and all exported fields
  - DefaultRateLimiterOptions constructor
  - TokenBucketLimiter type and constructor (NewTokenBucketLimiter)
  - TokenBucketLimiter methods (Wait, Reserve, UpdateLimit, SetOptions, GetStats)
  - NoopRateLimiter type and constructor (NewNoopRateLimiter)
  - BackoffStrategy interface (NextBackoff, Reset, MaxAttempts)
  - ExponentialBackoff type and constructor (NewExponentialBackoff)
  - ExponentialBackoff fields (InitialDelay, MaxDelay, Factor, Jitter, JitterFactor, MaxAttempt)
  - DefaultBackoff constructor
  - RetryableFunc type alias
  - RetryWithBackoff function
  - IsRetryableError function
  - CircuitBreaker type and constructor (NewCircuitBreaker)
  - CircuitBreaker methods (Execute, AllowRequest, RecordResult, State, Reset, SetOptions)
  - CircuitBreakerState constants (CircuitClosed, CircuitOpen, CircuitHalfOpen)
  - CircuitBreakerOptions struct and fields
  - DefaultCircuitBreakerOptions constructor
  - ErrCircuitOpen sentinel error

# Compatibility Guarantees

For beta components:
  - Minor backward-incompatible changes may still occur in minor releases
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Create a token bucket rate limiter:

	opts := ratelimit.DefaultRateLimiterOptions()
	opts.RequestsPerSecond = 5.0
	opts.BurstSize = 10

	limiter := ratelimit.NewTokenBucketLimiter(opts)

	// Wait for a token before making a request
	if err := limiter.Wait(ctx); err != nil {
		// Context was canceled
	}

Retry with exponential backoff:

	backoff := ratelimit.NewExponentialBackoff(
		100*time.Millisecond, // initial delay
		30*time.Second,       // max delay
		2.0,                  // factor
		5,                    // max attempts
	)

	err := ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
		return doSomeAPICall(ctx)
	}, backoff, ratelimit.IsRetryableError)

Use a circuit breaker to protect against cascading failures:

	cb := ratelimit.NewCircuitBreaker(ratelimit.DefaultCircuitBreakerOptions())

	err := cb.Execute(ctx, func(ctx context.Context) error {
		return callExternalService(ctx)
	})
	if errors.Is(err, ratelimit.ErrCircuitOpen) {
		// The circuit is open; the service is temporarily unavailable
	}
*/
package ratelimit

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	// CircuitClosed means the circuit is closed and requests are allowed
	CircuitClosed CircuitBreakerState = iota

	// CircuitOpen means the circuit is open and requests are blocked
	CircuitOpen

	// CircuitHalfOpen means the circuit is testing if service is healthy
	CircuitHalfOpen
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreakerOptions contains configuration for a circuit breaker
type CircuitBreakerOptions struct {
	// Threshold is the number of consecutive failures to open the circuit
	Threshold int

	// Timeout is how long to keep the circuit open before testing
	Timeout time.Duration

	// HalfOpenSuccesses is the number of successes needed in half-open state
	HalfOpenSuccesses int

	// OnStateChange is called when the circuit state changes
	OnStateChange func(from, to CircuitBreakerState)
}

// DefaultCircuitBreakerOptions returns the default options for a circuit breaker
func DefaultCircuitBreakerOptions() *CircuitBreakerOptions {
	return &CircuitBreakerOptions{
		Threshold:         5,
		Timeout:           30 * time.Second,
		HalfOpenSuccesses: 2,
	}
}

// CircuitBreaker implements the circuit breaker pattern for handling failures
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failures        int
	successes       int
	lastStateChange time.Time
	options         *CircuitBreakerOptions
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(options *CircuitBreakerOptions) *CircuitBreaker {
	if options == nil {
		options = DefaultCircuitBreakerOptions()
	}

	return &CircuitBreaker{
		state:           CircuitClosed,
		failures:        0,
		successes:       0,
		lastStateChange: time.Now(),
		options:         options,
	}
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if the circuit is open
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.RecordResult(err)

	return err
}

// AllowRequest checks if a request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if the timeout has elapsed
		if now.After(cb.lastStateChange.Add(cb.options.Timeout)) {
			// Transition to half-open state
			cb.transitionState(CircuitHalfOpen)
			return true
		}
		return false

	case CircuitHalfOpen:
		// Only allow one request at a time in half-open state
		return true

	default:
		return true
	}
}

// RecordResult records the result of a request
func (cb *CircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		if err != nil {
			cb.failures++

			// If we've reached the threshold, open the circuit
			if cb.failures >= cb.options.Threshold {
				cb.transitionState(CircuitOpen)
			}
		} else {
			// Reset failure count on success
			cb.failures = 0
		}

	case CircuitHalfOpen:
		if err != nil {
			// Any failure in half-open state opens the circuit again
			cb.transitionState(CircuitOpen)
		} else {
			cb.successes++

			// If we've reached the required success count, close the circuit
			if cb.successes >= cb.options.HalfOpenSuccesses {
				cb.transitionState(CircuitClosed)
			}
		}

	case CircuitOpen:
		// Shouldn't happen, but just in case
		if err != nil {
			cb.failures++
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.transitionState(CircuitClosed)
	cb.failures = 0
	cb.successes = 0
}

// SetOptions updates the circuit breaker's configuration
func (cb *CircuitBreaker) SetOptions(options *CircuitBreakerOptions) {
	if options == nil {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.options = options
}

// transitionState changes the state of the circuit breaker
func (cb *CircuitBreaker) transitionState(newState CircuitBreakerState) {
	if cb.state == newState {
		return
	}

	prevState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	switch newState {
	case CircuitClosed:
		cb.failures = 0
		cb.successes = 0
	case CircuitHalfOpen:
		cb.successes = 0
	case CircuitOpen:
		cb.failures = 0
	}

	// Notify of state change if callback is provided
	if cb.options.OnStateChange != nil {
		go cb.options.OnStateChange(prevState, newState)
	}
}

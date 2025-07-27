// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package main

import (
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed   CircuitState = iota // Circuit is closed (allowing requests)
	StateOpen                         // Circuit is open (blocking requests)
	StateHalfOpen                     // Circuit is half-open (testing if service is recovered)
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state            CircuitState
	failureThreshold int
	resetTimeout     time.Duration
	failureCount     int
	lastFailureTime  time.Time
	mutex            sync.RWMutex
}

// CircuitBreakerOption defines a configuration option for the circuit breaker
type CircuitBreakerOption func(*CircuitBreaker)

// WithFailureThreshold sets the number of failures before opening the circuit
func WithFailureThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = threshold
	}
}

// WithResetTimeout sets the timeout before testing if the service has recovered
func WithResetTimeout(timeout time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.resetTimeout = timeout
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given options
func NewCircuitBreaker(options ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: 5,                // Default: 5 failures to open circuit
		resetTimeout:     30 * time.Second, // Default: 30 second timeout
		failureCount:     0,
	}

	// Apply options
	for _, option := range options {
		option(cb)
	}

	return cb
}

// AllowRequest checks if a request should be allowed based on the circuit state
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if it's time to try again
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// Move to half-open state to test service recovery
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		// In half-open state, allow one request at a time to test the service
		return true
	default:
		return true
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// If in half-open state and success, close the circuit
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}

	// Reset failure count
	cb.failureCount = 0
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	// In half-open state, immediately open the circuit on failure
	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		return
	}

	// In closed state, open circuit if failure threshold reached
	if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Reset forcibly resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
}

// ForceOpen forcibly opens the circuit
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = StateOpen
	cb.lastFailureTime = time.Now()
}

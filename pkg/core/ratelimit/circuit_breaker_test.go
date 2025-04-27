// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	// Create a circuit breaker with custom options
	stateChanges := make([]CircuitBreakerState, 0)
	
	options := &CircuitBreakerOptions{
		Threshold:         3,                  // Open after 3 failures
		Timeout:           50 * time.Millisecond, // Half-open after 50ms
		HalfOpenSuccesses: 2,                  // Close after 2 successes in half-open
		OnStateChange: func(from, to CircuitBreakerState) {
			stateChanges = append(stateChanges, to)
		},
	}
	
	cb := NewCircuitBreaker(options)
	
	// Test initial state
	if cb.State() != CircuitClosed {
		t.Errorf("Expected initial state to be CircuitClosed, got %v", cb.State())
	}
	
	// Test successful request
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected successful execution, got error: %v", err)
	}
	
	if cb.State() != CircuitClosed {
		t.Errorf("Expected state to remain CircuitClosed after success, got %v", cb.State())
	}
	
	// Test failures that should open the circuit
	testErr := errors.New("test error")
	
	for i := 0; i < 3; i++ {
		err = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
		
		if err != testErr {
			t.Errorf("Expected test error, got: %v", err)
		}
	}
	
	// Circuit should now be open
	if cb.State() != CircuitOpen {
		t.Errorf("Expected state to be CircuitOpen after 3 failures, got %v", cb.State())
	}
	
	// Requests should fail with circuit open error
	err = cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected ErrCircuitOpen, got: %v", err)
	}
	
	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)
	
	// The next request should be allowed (half-open state)
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if cb.State() != CircuitHalfOpen {
		t.Errorf("Expected state to be CircuitHalfOpen after timeout, got %v", cb.State())
	}
	
	// Another successful request should close the circuit
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if cb.State() != CircuitClosed {
		t.Errorf("Expected state to be CircuitClosed after 2 successes in half-open, got %v", cb.State())
	}
	
	// Verify state transitions
	if len(stateChanges) != 3 { // Open, HalfOpen, Closed
		t.Errorf("Expected 3 state changes, got %d", len(stateChanges))
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	// Create a circuit breaker
	options := &CircuitBreakerOptions{
		Threshold:         2,                  // Open after 2 failures
		Timeout:           20 * time.Millisecond, // Half-open after 20ms
		HalfOpenSuccesses: 2,                  // Close after 2 successes in half-open
	}
	
	cb := NewCircuitBreaker(options)
	
	// Trigger circuit open
	testErr := errors.New("test error")
	
	for i := 0; i < 2; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	
	// Circuit should be open
	if cb.State() != CircuitOpen {
		t.Errorf("Expected state to be CircuitOpen, got %v", cb.State())
	}
	
	// Wait for timeout to transition to half-open
	time.Sleep(30 * time.Millisecond)
	
	// Fail the next request in half-open state
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return testErr
	})
	
	// Circuit should be open again
	if cb.State() != CircuitOpen {
		t.Errorf("Expected state to be CircuitOpen after failure in half-open, got %v", cb.State())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	// Create a circuit breaker
	cb := NewCircuitBreaker(nil) // Use default options
	
	// Trigger circuit open with failures
	testErr := errors.New("test error")
	
	for i := 0; i < 5; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	
	// Circuit should be open
	if cb.State() != CircuitOpen {
		t.Errorf("Expected state to be CircuitOpen, got %v", cb.State())
	}
	
	// Reset the circuit breaker
	cb.Reset()
	
	// Circuit should be closed
	if cb.State() != CircuitClosed {
		t.Errorf("Expected state to be CircuitClosed after reset, got %v", cb.State())
	}
	
	// Should allow requests
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected successful execution after reset, got error: %v", err)
	}
}
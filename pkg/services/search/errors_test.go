// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"fmt"
	"net/http"
	"testing"
)

func TestSearchError(t *testing.T) {
	// Create error
	err := &SearchError{
		Code:      ErrorCodeIndexNotFound,
		Message:   "Index not found",
		Status:    http.StatusNotFound,
		RequestID: "abc123",
	}

	// Check error message
	expectedMsg := "search error: Index not found (code=IndexNotFound, status=404, request_id=abc123)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Create error without request ID
	err = &SearchError{
		Code:    ErrorCodeIndexNotFound,
		Message: "Index not found",
		Status:  http.StatusNotFound,
	}

	// Check error message
	expectedMsg = "search error: Index not found (code=IndexNotFound, status=404)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Create error with cause
	cause := fmt.Errorf("original error")
	err = &SearchError{
		Code:    ErrorCodeIndexNotFound,
		Message: "Index not found",
		Status:  http.StatusNotFound,
		Cause:   cause,
	}

	// Check unwrap
	if err.Unwrap() != cause {
		t.Errorf("Expected unwrap to return original error")
	}
}

func TestIsSearchError(t *testing.T) {
	// Create error
	err := &SearchError{
		Code:    ErrorCodeIndexNotFound,
		Message: "Index not found",
		Status:  http.StatusNotFound,
	}

	// Check IsSearchError
	if !IsSearchError(err) {
		t.Errorf("Expected IsSearchError to return true")
	}

	// Create wrapped error
	wrapped := fmt.Errorf("wrapped: %w", err)

	// Check IsSearchError with wrapped error
	if !IsSearchError(wrapped) {
		t.Errorf("Expected IsSearchError to return true for wrapped error")
	}

	// Create regular error
	regularErr := fmt.Errorf("regular error")

	// Check IsSearchError with regular error
	if IsSearchError(regularErr) {
		t.Errorf("Expected IsSearchError to return false for regular error")
	}
}

func TestAsSearchError(t *testing.T) {
	// Create error
	err := &SearchError{
		Code:    ErrorCodeIndexNotFound,
		Message: "Index not found",
		Status:  http.StatusNotFound,
	}

	// Check AsSearchError
	searchErr, ok := AsSearchError(err)
	if !ok {
		t.Errorf("Expected AsSearchError to return true")
	}
	if searchErr.Code != ErrorCodeIndexNotFound {
		t.Errorf("Expected code = %s, got %s", ErrorCodeIndexNotFound, searchErr.Code)
	}

	// Create wrapped error
	wrapped := fmt.Errorf("wrapped: %w", err)

	// Check AsSearchError with wrapped error
	searchErr, ok = AsSearchError(wrapped)
	if !ok {
		t.Errorf("Expected AsSearchError to return true for wrapped error")
	}
	if searchErr.Code != ErrorCodeIndexNotFound {
		t.Errorf("Expected code = %s, got %s", ErrorCodeIndexNotFound, searchErr.Code)
	}

	// Create regular error
	regularErr := fmt.Errorf("regular error")

	// Check AsSearchError with regular error
	_, ok = AsSearchError(regularErr)
	if ok {
		t.Errorf("Expected AsSearchError to return false for regular error")
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{
			name: "IsIndexNotFoundError with IndexNotFound code",
			err: &SearchError{
				Code:   ErrorCodeIndexNotFound,
				Status: http.StatusNotFound,
			},
			checker:  IsIndexNotFoundError,
			expected: true,
		},
		{
			name: "IsIndexNotFoundError with 404 status",
			err: &SearchError{
				Code:   "SomeOtherCode",
				Status: http.StatusNotFound,
			},
			checker:  IsIndexNotFoundError,
			expected: true,
		},
		{
			name: "IsPermissionDeniedError with PermissionDenied code",
			err: &SearchError{
				Code:   ErrorCodePermissionDenied,
				Status: http.StatusForbidden,
			},
			checker:  IsPermissionDeniedError,
			expected: true,
		},
		{
			name: "IsPermissionDeniedError with 403 status",
			err: &SearchError{
				Code:   "SomeOtherCode",
				Status: http.StatusForbidden,
			},
			checker:  IsPermissionDeniedError,
			expected: true,
		},
		{
			name: "IsInvalidQueryError with InvalidQuerySyntax code",
			err: &SearchError{
				Code:   ErrorCodeInvalidQuery,
				Status: http.StatusBadRequest,
			},
			checker:  IsInvalidQueryError,
			expected: true,
		},
		{
			name: "IsInvalidQueryError with invalid query in message",
			err: &SearchError{
				Code:    "SomeOtherCode",
				Message: "Invalid query syntax at position 10",
				Status:  http.StatusBadRequest,
			},
			checker:  IsInvalidQueryError,
			expected: true,
		},
		{
			name: "IsTaskNotFoundError with TaskNotFound code",
			err: &SearchError{
				Code:   ErrorCodeTaskNotFound,
				Status: http.StatusNotFound,
			},
			checker:  IsTaskNotFoundError,
			expected: true,
		},
		{
			name: "IsIndexExistsError with IndexExists code",
			err: &SearchError{
				Code:   ErrorCodeIndexExists,
				Status: http.StatusConflict,
			},
			checker:  IsIndexExistsError,
			expected: true,
		},
		{
			name: "IsIndexExistsError with index already exists in message",
			err: &SearchError{
				Code:    "SomeOtherCode",
				Message: "Index already exists with that name",
				Status:  http.StatusConflict,
			},
			checker:  IsIndexExistsError,
			expected: true,
		},
		{
			name: "IsRateLimitError with RateLimit code",
			err: &SearchError{
				Code:   ErrorCodeRateLimit,
				Status: http.StatusTooManyRequests,
			},
			checker:  IsRateLimitError,
			expected: true,
		},
		{
			name: "IsRateLimitError with 429 status",
			err: &SearchError{
				Code:   "SomeOtherCode",
				Status: http.StatusTooManyRequests,
			},
			checker:  IsRateLimitError,
			expected: true,
		},
		{
			name:     "Non-SearchError",
			err:      fmt.Errorf("regular error"),
			checker:  IsIndexNotFoundError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

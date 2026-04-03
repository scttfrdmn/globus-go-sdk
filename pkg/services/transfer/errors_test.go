// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestParseTransferError(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		responseBody   []byte
		expectedError  error
		matchWithError error // Error to compare using errors.Is
	}{
		{
			name:           "Empty response with 404",
			statusCode:     http.StatusNotFound,
			responseBody:   []byte{},
			matchWithError: ErrResourceNotFound,
		},
		{
			name:           "Empty response with 403",
			statusCode:     http.StatusForbidden,
			responseBody:   []byte{},
			matchWithError: ErrPermissionDenied,
		},
		{
			name:           "Empty response with 401",
			statusCode:     http.StatusUnauthorized,
			responseBody:   []byte{},
			matchWithError: ErrAuthenticationRequired,
		},
		{
			name:           "Empty response with 429",
			statusCode:     http.StatusTooManyRequests,
			responseBody:   []byte{},
			matchWithError: ErrRateLimitExceeded,
		},
		{
			name:           "Empty response with 500",
			statusCode:     http.StatusInternalServerError,
			responseBody:   []byte{},
			matchWithError: ErrServerError,
		},
		{
			name:       "ResourceNotFound error response",
			statusCode: http.StatusNotFound,
			responseBody: []byte(`{
				"code": "ResourceNotFound",
				"message": "The requested resource was not found",
				"request_id": "abc123"
			}`),
			matchWithError: ErrResourceNotFound,
		},
		{
			name:       "EndpointNotFound error response",
			statusCode: http.StatusNotFound,
			responseBody: []byte(`{
				"code": "EndpointNotFound",
				"message": "The requested endpoint was not found",
				"request_id": "abc123"
			}`),
			matchWithError: ErrResourceNotFound,
		},
		{
			name:       "PermissionDenied error response",
			statusCode: http.StatusForbidden,
			responseBody: []byte(`{
				"code": "PermissionDenied",
				"message": "You do not have permission to perform this action",
				"request_id": "abc123"
			}`),
			matchWithError: ErrPermissionDenied,
		},
		{
			name:       "RateLimitExceeded error response",
			statusCode: http.StatusTooManyRequests,
			responseBody: []byte(`{
				"code": "RateLimitExceeded",
				"message": "Rate limit exceeded, retry after 60 seconds",
				"request_id": "abc123"
			}`),
			matchWithError: ErrRateLimitExceeded,
		},
		{
			name:       "EndpointNotActivated error response",
			statusCode: http.StatusBadRequest,
			responseBody: []byte(`{
				"code": "EndpointNotActivated",
				"message": "The endpoint is not activated",
				"request_id": "abc123"
			}`),
			matchWithError: ErrEndpointNotActivated,
		},
		{
			name:       "Invalid JSON response",
			statusCode: http.StatusBadRequest,
			responseBody: []byte(`{
				"code": "Bad
				Request",
				"message": "Invalid JSON"
			}`),
			expectedError: fmt.Errorf("request failed with status code 400: {\n\t\t\t\t\"code\": \"Bad\n\t\t\t\tRequest\",\n\t\t\t\t\"message\": \"Invalid JSON\"\n\t\t\t}"),
		},
		{
			name:          "Non-JSON response",
			statusCode:    http.StatusBadRequest,
			responseBody:  []byte(`Error occurred: The request was invalid`),
			expectedError: fmt.Errorf("request failed with status code 400: Error occurred: The request was invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseTransferError(tc.statusCode, tc.responseBody)

			// Ensure we got an error
			if err == nil {
				t.Fatalf("Expected an error, got nil")
			}

			// If we expect to match with a specific error via errors.Is
			if tc.matchWithError != nil {
				if !errors.Is(err, tc.matchWithError) {
					t.Errorf("Expected error to match %v, got %v", tc.matchWithError, err)
				}
			}

			// If we expect a specific error message
			if tc.expectedError != nil {
				if err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
			}
		})
	}
}

func TestIsResourceNotFound(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "ResourceNotFound sentinel error",
			err:      ErrResourceNotFound,
			expected: true,
		},
		{
			name:     "EndpointNotFound sentinel error",
			err:      ErrEndpointNotFound,
			expected: true,
		},
		{
			name:     "FileNotFound sentinel error",
			err:      ErrFileNotFound,
			expected: true,
		},
		{
			name:     "Wrapped ResourceNotFound error",
			err:      fmt.Errorf("operation failed: %w", ErrResourceNotFound),
			expected: true,
		},
		{
			name: "TransferError with ResourceNotFound code",
			err: &TransferError{
				Code:       ErrCodeResourceNotFound,
				Message:    "Resource not found",
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
		{
			name: "TransferError with EndpointNotFound code",
			err: &TransferError{
				Code:       ErrCodeEndpointNotFound,
				Message:    "Endpoint not found",
				StatusCode: http.StatusNotFound,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsResourceNotFound(tc.err)
			if result != tc.expected {
				t.Errorf("IsResourceNotFound(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestIsPermissionDenied(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "PermissionDenied sentinel error",
			err:      ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "Wrapped PermissionDenied error",
			err:      fmt.Errorf("operation failed: %w", ErrPermissionDenied),
			expected: true,
		},
		{
			name: "TransferError with PermissionDenied code",
			err: &TransferError{
				Code:       ErrCodePermissionDenied,
				Message:    "Permission denied",
				StatusCode: http.StatusForbidden,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsPermissionDenied(tc.err)
			if result != tc.expected {
				t.Errorf("IsPermissionDenied(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestIsRetryableTransferError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "RateLimitExceeded sentinel error",
			err:      ErrRateLimitExceeded,
			expected: true,
		},
		{
			name:     "ServerError sentinel error",
			err:      ErrServerError,
			expected: true,
		},
		{
			name:     "PermissionDenied sentinel error (not retryable)",
			err:      ErrPermissionDenied,
			expected: false,
		},
		{
			name: "TransferError with RateLimitExceeded code",
			err: &TransferError{
				Code:       ErrCodeRateLimitExceeded,
				Message:    "Rate limit exceeded",
				StatusCode: http.StatusTooManyRequests,
			},
			expected: true,
		},
		{
			name: "TransferError with ServiceUnavailable code",
			err: &TransferError{
				Code:       ErrCodeServiceUnavailable,
				Message:    "Service unavailable",
				StatusCode: http.StatusServiceUnavailable,
			},
			expected: true,
		},
		{
			name: "TransferError with ServerError code",
			err: &TransferError{
				Code:       ErrCodeServerError,
				Message:    "Internal server error",
				StatusCode: http.StatusInternalServerError,
			},
			expected: true,
		},
		{
			name: "TransferError with message containing 'temporarily'",
			err: &TransferError{
				Code:       "CustomError",
				Message:    "Service temporarily unavailable",
				StatusCode: http.StatusBadRequest,
			},
			expected: true,
		},
		{
			name: "TransferError with message containing 'retry'",
			err: &TransferError{
				Code:       "CustomError",
				Message:    "Please retry the request later",
				StatusCode: http.StatusBadRequest,
			},
			expected: true,
		},
		{
			name: "TransferError with 500 status code",
			err: &TransferError{
				Code:       "CustomError",
				Message:    "Something went wrong",
				StatusCode: http.StatusInternalServerError,
			},
			expected: true,
		},
		{
			name: "TransferError with 400 status code (not retryable)",
			err: &TransferError{
				Code:       "BadRequest",
				Message:    "Invalid request parameters",
				StatusCode: http.StatusBadRequest,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRetryableTransferError(tc.err)
			if result != tc.expected {
				t.Errorf("IsRetryableTransferError(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestGlobusError(t *testing.T) {
	t.Run("NewGlobusError", func(t *testing.T) {
		err := NewGlobusError("test", "TEST_CODE", "test message")

		if err.Service != "test" {
			t.Errorf("Expected service 'test', got %s", err.Service)
		}

		if err.Code != "TEST_CODE" {
			t.Errorf("Expected code 'TEST_CODE', got %s", err.Code)
		}

		if err.Message != "test message" {
			t.Errorf("Expected message 'test message', got %s", err.Message)
		}

		if err.Timestamp.IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})

	t.Run("Error method", func(t *testing.T) {
		err := NewGlobusError("test", "TEST_CODE", "test message")
		expected := "[test] TEST_CODE: test message"

		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("Error with detail", func(t *testing.T) {
		err := NewGlobusError("test", "TEST_CODE", "test message").WithDetail("additional info")
		expected := "[test] TEST_CODE: test message - additional info"

		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		err := NewGlobusError("test", "TEST_CODE", "test message").
			WithContext("key1", "value1").
			WithContext("key2", "value2")

		if err.Context["key1"] != "value1" {
			t.Errorf("Expected context key1 to be 'value1', got '%s'", err.Context["key1"])
		}

		if err.Context["key2"] != "value2" {
			t.Errorf("Expected context key2 to be 'value2', got '%s'", err.Context["key2"])
		}
	})

	t.Run("WithUnderlying", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewGlobusError("test", "TEST_CODE", "test message").WithUnderlying(underlying)

		if err.Unwrap() != underlying {
			t.Error("Expected underlying error to be set")
		}
	})

	t.Run("Is method", func(t *testing.T) {
		err1 := NewGlobusError("test", "TEST_CODE", "test message")
		err2 := NewGlobusError("test", "TEST_CODE", "different message")
		err3 := NewGlobusError("test", "OTHER_CODE", "test message")
		err4 := NewGlobusError("other", "TEST_CODE", "test message")

		if !errors.Is(err1, err2) {
			t.Error("Expected errors with same service and code to be equal")
		}

		if errors.Is(err1, err3) {
			t.Error("Expected errors with different codes to not be equal")
		}

		if errors.Is(err1, err4) {
			t.Error("Expected errors with different services to not be equal")
		}
	})
}

func TestGlobusErrorFromHTTPResponse(t *testing.T) {
	t.Run("Basic HTTP error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusNotFound,
			Header:     make(http.Header),
		}
		resp.Header.Set("X-Request-Id", "test-request-id")

		err := NewGlobusErrorFromHTTPResponse("test", resp)

		if err.Service != "test" {
			t.Errorf("Expected service 'test', got %s", err.Service)
		}

		if err.HTTPStatus != http.StatusNotFound {
			t.Errorf("Expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
		}

		if err.RequestID != "test-request-id" {
			t.Errorf("Expected request ID 'test-request-id', got %s", err.RequestID)
		}

		if err.Message != "Not Found" {
			t.Errorf("Expected message 'Not Found', got %s", err.Message)
		}

		if err.Code != "HTTP_404" {
			t.Errorf("Expected code 'HTTP_404', got %s", err.Code)
		}
	})

	t.Run("Retryable error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     make(http.Header),
		}

		err := NewGlobusErrorFromHTTPResponse("test", resp)

		if !err.IsRetryable() {
			t.Error("Expected 429 error to be retryable")
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
		}

		err := NewGlobusErrorFromHTTPResponse("test", resp)

		if err.IsRetryable() {
			t.Error("Expected 400 error to not be retryable")
		}
	})
}

func TestGlobusErrorClassification(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		code          string
		message       string
		isAuth        bool
		isAuthz       bool
		isNotFound    bool
		isRateLimit   bool
		isServerError bool
		isClientError bool
	}{
		{
			name:          "Authentication error",
			statusCode:    http.StatusUnauthorized,
			code:          "UNAUTHORIZED",
			message:       "Unauthorized",
			isAuth:        true,
			isClientError: true,
		},
		{
			name:          "Authorization error",
			statusCode:    http.StatusForbidden,
			code:          "FORBIDDEN",
			message:       "Forbidden",
			isAuthz:       true,
			isClientError: true,
		},
		{
			name:          "Not found error",
			statusCode:    http.StatusNotFound,
			code:          "NOT_FOUND",
			message:       "Not Found",
			isNotFound:    true,
			isClientError: true,
		},
		{
			name:          "Rate limit error",
			statusCode:    http.StatusTooManyRequests,
			code:          "RATE_LIMIT",
			message:       "Too Many Requests",
			isRateLimit:   true,
			isClientError: true,
		},
		{
			name:          "Server error",
			statusCode:    http.StatusInternalServerError,
			code:          "SERVER_ERROR",
			message:       "Internal Server Error",
			isServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewGlobusErrorWithStatus("test", tt.code, tt.message, tt.statusCode)

			if err.IsAuthenticationError() != tt.isAuth {
				t.Errorf("Expected IsAuthenticationError() = %v, got %v", tt.isAuth, err.IsAuthenticationError())
			}

			if err.IsAuthorizationError() != tt.isAuthz {
				t.Errorf("Expected IsAuthorizationError() = %v, got %v", tt.isAuthz, err.IsAuthorizationError())
			}

			if err.IsNotFoundError() != tt.isNotFound {
				t.Errorf("Expected IsNotFoundError() = %v, got %v", tt.isNotFound, err.IsNotFoundError())
			}

			if err.IsRateLimitError() != tt.isRateLimit {
				t.Errorf("Expected IsRateLimitError() = %v, got %v", tt.isRateLimit, err.IsRateLimitError())
			}

			if err.IsServerError() != tt.isServerError {
				t.Errorf("Expected IsServerError() = %v, got %v", tt.isServerError, err.IsServerError())
			}

			if err.IsClientError() != tt.isClientError {
				t.Errorf("Expected IsClientError() = %v, got %v", tt.isClientError, err.IsClientError())
			}
		})
	}
}

func TestServiceSpecificErrors(t *testing.T) {
	t.Run("Auth errors", func(t *testing.T) {
		err := NewAuthError(AuthInvalidToken, "Invalid token")

		if err.Service != "auth" {
			t.Errorf("Expected service 'auth', got %s", err.Service)
		}

		if err.Code != AuthInvalidToken {
			t.Errorf("Expected code '%s', got %s", AuthInvalidToken, err.Code)
		}
	})

	t.Run("Transfer errors", func(t *testing.T) {
		err := NewTransferError(TransferTaskNotFound, "Task not found")

		if err.Service != "transfer" {
			t.Errorf("Expected service 'transfer', got %s", err.Service)
		}

		if err.Code != TransferTaskNotFound {
			t.Errorf("Expected code '%s', got %s", TransferTaskNotFound, err.Code)
		}
	})

	t.Run("Groups errors", func(t *testing.T) {
		err := NewGroupsError(GroupsGroupNotFound, "Group not found")

		if err.Service != "groups" {
			t.Errorf("Expected service 'groups', got %s", err.Service)
		}

		if err.Code != GroupsGroupNotFound {
			t.Errorf("Expected code '%s', got %s", GroupsGroupNotFound, err.Code)
		}
	})
}

func TestParseGlobusErrorFromJSON(t *testing.T) {
	t.Run("Standard error format", func(t *testing.T) {
		jsonData := `{
			"error": {
				"code": "TEST_CODE",
				"message": "Test message",
				"detail": "Additional details"
			}
		}`

		err, parseErr := ParseGlobusErrorFromJSON("test", []byte(jsonData))
		if parseErr != nil {
			t.Fatalf("Failed to parse error: %v", parseErr)
		}

		if err.Code != "TEST_CODE" {
			t.Errorf("Expected code 'TEST_CODE', got %s", err.Code)
		}

		if err.Message != "Test message" {
			t.Errorf("Expected message 'Test message', got %s", err.Message)
		}

		if err.Detail != "Additional details" {
			t.Errorf("Expected detail 'Additional details', got %s", err.Detail)
		}
	})

	t.Run("Alternative error format", func(t *testing.T) {
		jsonData := `{
			"error_code": "ALT_CODE",
			"message": "Alternative message"
		}`

		err, parseErr := ParseGlobusErrorFromJSON("test", []byte(jsonData))
		if parseErr != nil {
			t.Fatalf("Failed to parse error: %v", parseErr)
		}

		if err.Code != "ALT_CODE" {
			t.Errorf("Expected code 'ALT_CODE', got %s", err.Code)
		}

		if err.Message != "Alternative message" {
			t.Errorf("Expected message 'Alternative message', got %s", err.Message)
		}
	})

	t.Run("Simple message format", func(t *testing.T) {
		jsonData := `{
			"message": "Simple error message"
		}`

		err, parseErr := ParseGlobusErrorFromJSON("test", []byte(jsonData))
		if parseErr != nil {
			t.Fatalf("Failed to parse error: %v", parseErr)
		}

		if err.Code != "UnknownError" {
			t.Errorf("Expected code 'UnknownError', got %s", err.Code)
		}

		if err.Message != "Simple error message" {
			t.Errorf("Expected message 'Simple error message', got %s", err.Message)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		jsonData := `{invalid json}`

		_, parseErr := ParseGlobusErrorFromJSON("test", []byte(jsonData))
		if parseErr == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("Empty error response", func(t *testing.T) {
		jsonData := `{}`

		_, parseErr := ParseGlobusErrorFromJSON("test", []byte(jsonData))
		if parseErr == nil {
			t.Error("Expected error for empty response")
		}
	})
}

func TestGlobusErrorString(t *testing.T) {
	err := NewGlobusError("test", "TEST_CODE", "test message").
		WithDetail("additional info").
		WithRequestID("req-123").
		WithContext("key", "value")
	err.HTTPStatus = http.StatusBadRequest

	result := err.String()

	expectedParts := []string{
		"Service: test",
		"Code: TEST_CODE",
		"Message: test message",
		"Detail: additional info",
		"RequestID: req-123",
		"HTTPStatus: 400",
		"Context: key=value",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected string to contain '%s', got '%s'", part, result)
		}
	}
}

func TestRetryableStatusCodes(t *testing.T) {
	retryableCodes := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}

	for _, code := range retryableCodes {
		t.Run(fmt.Sprintf("Status %d", code), func(t *testing.T) {
			if !isRetryableHTTPStatus(code) {
				t.Errorf("Expected status %d to be retryable", code)
			}
		})
	}

	nonRetryableCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
	}

	for _, code := range nonRetryableCodes {
		t.Run(fmt.Sprintf("Status %d", code), func(t *testing.T) {
			if isRetryableHTTPStatus(code) {
				t.Errorf("Expected status %d to not be retryable", code)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an enhanced error from the Globus API with structured details
// This matches Python SDK v4.x error handling improvements
type APIError struct {
	// StatusCode is the HTTP status code
	StatusCode int `json:"status_code"`

	// Message is the error message
	Message string `json:"message"`

	// Code is the Globus error code (e.g., "InvalidRequest", "AuthenticationFailed")
	Code string `json:"code,omitempty"`

	// RequestID is the Globus request ID for tracking
	RequestID string `json:"request_id,omitempty"`

	// Details contains additional error details from the API
	Details map[string]interface{} `json:"details,omitempty"`

	// Resource is the API resource that caused the error
	Resource string `json:"resource,omitempty"`

	// Notes contains additional context or suggestions
	Notes []string `json:"notes,omitempty"`

	// HTTPResponse is the original HTTP response
	HTTPResponse *http.Response `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Globus API error [%s] (HTTP %d): %s", e.Code, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Globus API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsAuthError returns true if this is an authentication error
func (e *APIError) IsAuthError() bool {
	return e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true if this is a not found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsConflict returns true if this is a conflict error
func (e *APIError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// IsRateLimited returns true if this is a rate limit error
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if this is a server error (5xx)
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// NewAPIError creates a new APIError from an HTTP response. The message
// argument is the already-read response body (the caller reads resp.Body to
// produce it); NewAPIError parses that string for structured fields rather than
// re-reading resp.Body, which by this point is drained. See issue #63: the old
// code decoded resp.Body a second time, always hit io.EOF, and left Details nil.
func NewAPIError(resp *http.Response, message string) *APIError {
	apiErr := &APIError{
		StatusCode:   resp.StatusCode,
		Message:      message,
		RequestID:    resp.Header.Get("X-Globus-Request-ID"),
		HTTPResponse: resp,
	}

	// Parse the body (passed in as message) for structured error fields.
	var errorBody map[string]interface{}
	if err := json.Unmarshal([]byte(message), &errorBody); err == nil {
		// Extract common top-level error fields.
		if code, ok := errorBody["code"].(string); ok {
			apiErr.Code = code
		}
		if msg, ok := errorBody["message"].(string); ok && msg != "" {
			apiErr.Message = msg
		}
		if resource, ok := errorBody["resource"].(string); ok {
			apiErr.Resource = resource
		}

		// Some Globus errors (e.g. Auth) nest the details under errors[0]
		// rather than at the top level; fall back to the first sub-error for
		// code/message when the top-level fields are absent.
		if sub, ok := firstSubError(errorBody); ok {
			if apiErr.Code == "" {
				if code, ok := sub["code"].(string); ok {
					apiErr.Code = code
				}
			}
			if msg, ok := sub["message"].(string); ok && msg != "" && apiErr.Message == message {
				apiErr.Message = msg
			}
		}

		// Store the full parsed body for consumers that need nested fields
		// (e.g. authorization_parameters.session_required_policies).
		apiErr.Details = errorBody
	}

	return apiErr
}

// firstSubError returns the first element of a top-level "errors" array as a
// map, if present. Globus Auth returns error details in this JSON:API-style
// shape (`{"errors":[{"code":...,"detail":...}]}`).
func firstSubError(body map[string]interface{}) (map[string]interface{}, bool) {
	errs, ok := body["errors"].([]interface{})
	if !ok || len(errs) == 0 {
		return nil, false
	}
	sub, ok := errs[0].(map[string]interface{})
	return sub, ok
}

// ValidationError represents a client-side validation error
// This occurs before making an API request
type ValidationError struct {
	// Field is the field that failed validation
	Field string `json:"field,omitempty"`

	// Message is the validation error message
	Message string `json:"message"`

	// Value is the invalid value (optional)
	Value interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NetworkError represents a network-level error
type NetworkError struct {
	// Operation is the operation that failed (e.g., "dial", "read", "write")
	Operation string `json:"operation"`

	// Message is the error message
	Message string `json:"message"`

	// Err is the underlying error
	Err error `json:"-"`
}

// Error implements the error interface
func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("network error during %s: %s (%v)", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("network error during %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying error
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Error represents an API error
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Resource   string `json:"resource,omitempty"`
	Field      string `json:"field,omitempty"`
	StatusCode int    `json:"-"`
	RawBody    []byte `json:"-"`
}

// Error returns the error message
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s (status: %d)", e.Code, e.Message, e.StatusCode)
}

// ErrorResponse represents the error response from the API
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// NewAPIError creates a new Error from an API response
func NewAPIError(resp *http.Response) error {
	// Read and capture the response body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	// Try to parse as ErrorResponse
	errorResponse := &ErrorResponse{}
	err = json.Unmarshal(body, errorResponse)
	if err != nil || len(errorResponse.Errors) == 0 {
		// Failed to parse or no errors, return a generic error
		return &Error{
			Code:       "unknown_error",
			Message:    fmt.Sprintf("Request failed with status code %d", resp.StatusCode),
			StatusCode: resp.StatusCode,
			RawBody:    body,
		}
	}

	// Return the first error
	apiError := errorResponse.Errors[0]
	apiError.StatusCode = resp.StatusCode
	apiError.RawBody = body
	return &apiError
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	if apiErr, ok := err.(*Error); ok {
		return apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

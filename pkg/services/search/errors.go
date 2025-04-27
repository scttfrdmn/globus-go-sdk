// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Common error codes for the Search service
const (
	ErrorCodeIndexNotFound          = "IndexNotFound"
	ErrorCodePermissionDenied       = "PermissionDenied"
	ErrorCodeInvalidQuery           = "InvalidQuerySyntax"
	ErrorCodeInvalidIndexDefinition = "InvalidIndexDefinition"
	ErrorCodeTaskNotFound           = "TaskNotFound"
	ErrorCodeIndexExists            = "IndexExists"
	ErrorCodeRateLimit              = "RateLimit"
)

// SearchError represents an error from the Search service
type SearchError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Status    int    `json:"status"`
	RequestID string `json:"request_id"`
	Cause     error  `json:"-"`
}

// Error implements the error interface
func (e *SearchError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("search error: %s (code=%s, status=%d, request_id=%s)",
			e.Message, e.Code, e.Status, e.RequestID)
	}
	return fmt.Sprintf("search error: %s (code=%s, status=%d)",
		e.Message, e.Code, e.Status)
}

// Unwrap returns the underlying error
func (e *SearchError) Unwrap() error {
	return e.Cause
}

// IsSearchError checks if an error is a search error
func IsSearchError(err error) bool {
	var searchErr *SearchError
	return errors.As(err, &searchErr)
}

// AsSearchError converts an error to a SearchError if possible
func AsSearchError(err error) (*SearchError, bool) {
	var searchErr *SearchError
	ok := errors.As(err, &searchErr)
	return searchErr, ok
}

// IsIndexNotFoundError checks if an error is an index not found error
func IsIndexNotFoundError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodeIndexNotFound ||
			searchErr.Status == http.StatusNotFound
	}
	return false
}

// IsPermissionDeniedError checks if an error is a permission denied error
func IsPermissionDeniedError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodePermissionDenied ||
			searchErr.Status == http.StatusForbidden
	}
	return false
}

// IsInvalidQueryError checks if an error is an invalid query error
func IsInvalidQueryError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodeInvalidQuery ||
			strings.Contains(strings.ToLower(searchErr.Message), "invalid query")
	}
	return false
}

// IsTaskNotFoundError checks if an error is a task not found error
func IsTaskNotFoundError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodeTaskNotFound
	}
	return false
}

// IsIndexExistsError checks if an error is an index exists error
func IsIndexExistsError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodeIndexExists ||
			strings.Contains(strings.ToLower(searchErr.Message), "index already exists")
	}
	return false
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	if searchErr, ok := AsSearchError(err); ok {
		return searchErr.Code == ErrorCodeRateLimit ||
			searchErr.Status == http.StatusTooManyRequests
	}
	return false
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package errors provides unified error handling for the Globus Go SDK.
//
// This package implements a consistent error handling system across all Globus services,
// following the patterns established by the Python SDK for compatibility and familiarity.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GlobusError represents a standardized error from any Globus service.
// This error type provides consistent error handling across all services
// and includes context information for debugging and user messaging.
type GlobusError struct {
	// Code is the error code from the Globus service
	Code string `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// Detail provides additional information about the error
	Detail string `json:"detail,omitempty"`

	// RequestID is the unique identifier for the request that caused the error
	RequestID string `json:"request_id,omitempty"`

	// Service is the name of the Globus service that returned the error
	Service string `json:"service"`

	// HTTPStatus is the HTTP status code returned by the service
	HTTPStatus int `json:"http_status"`

	// Context provides additional context about the error
	Context map[string]string `json:"context,omitempty"`

	// Timestamp is when the error occurred
	Timestamp time.Time `json:"timestamp"`

	// Retryable indicates whether the operation can be retried
	Retryable bool `json:"retryable"`

	// Underlying is the original error that caused this error (if any)
	Underlying error `json:"-"`
}

// Error implements the error interface
func (e *GlobusError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s - %s", e.Service, e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Service, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *GlobusError) Unwrap() error {
	return e.Underlying
}

// Is implements error comparison for errors.Is()
func (e *GlobusError) Is(target error) bool {
	if target == nil {
		return false
	}

	targetErr, ok := target.(*GlobusError)
	if !ok {
		return false
	}

	return e.Service == targetErr.Service && e.Code == targetErr.Code
}

// String returns a detailed string representation of the error
func (e *GlobusError) String() string {
	parts := []string{
		fmt.Sprintf("Service: %s", e.Service),
		fmt.Sprintf("Code: %s", e.Code),
		fmt.Sprintf("Message: %s", e.Message),
	}

	if e.Detail != "" {
		parts = append(parts, fmt.Sprintf("Detail: %s", e.Detail))
	}

	if e.RequestID != "" {
		parts = append(parts, fmt.Sprintf("RequestID: %s", e.RequestID))
	}

	if e.HTTPStatus != 0 {
		parts = append(parts, fmt.Sprintf("HTTPStatus: %d", e.HTTPStatus))
	}

	if len(e.Context) > 0 {
		contextParts := make([]string, 0, len(e.Context))
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%s", k, v))
		}
		parts = append(parts, fmt.Sprintf("Context: %s", strings.Join(contextParts, ", ")))
	}

	return strings.Join(parts, "; ")
}

// NewGlobusError creates a new GlobusError
func NewGlobusError(service, code, message string) *GlobusError {
	return &GlobusError{
		Service:   service,
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]string),
	}
}

// NewGlobusErrorWithStatus creates a new GlobusError with HTTP status
func NewGlobusErrorWithStatus(service, code, message string, httpStatus int) *GlobusError {
	err := NewGlobusError(service, code, message)
	err.HTTPStatus = httpStatus
	err.Retryable = isRetryableHTTPStatus(httpStatus)
	return err
}

// NewGlobusErrorFromHTTPResponse creates a GlobusError from an HTTP response
func NewGlobusErrorFromHTTPResponse(service string, resp *http.Response) *GlobusError {
	err := &GlobusError{
		Service:    service,
		HTTPStatus: resp.StatusCode,
		Timestamp:  time.Now(),
		Context:    make(map[string]string),
		Retryable:  isRetryableHTTPStatus(resp.StatusCode),
	}

	// Extract request ID from headers
	if requestID := resp.Header.Get("X-Request-Id"); requestID != "" {
		err.RequestID = requestID
	} else if requestID := resp.Header.Get("Request-Id"); requestID != "" {
		err.RequestID = requestID
	}

	// Set default message based on HTTP status
	err.Message = http.StatusText(resp.StatusCode)
	err.Code = fmt.Sprintf("HTTP_%d", resp.StatusCode)

	return err
}

// WithDetail adds detail information to the error
func (e *GlobusError) WithDetail(detail string) *GlobusError {
	e.Detail = detail
	return e
}

// WithRequestID adds a request ID to the error
func (e *GlobusError) WithRequestID(requestID string) *GlobusError {
	e.RequestID = requestID
	return e
}

// WithContext adds context information to the error
func (e *GlobusError) WithContext(key, value string) *GlobusError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// WithUnderlying adds an underlying error
func (e *GlobusError) WithUnderlying(err error) *GlobusError {
	e.Underlying = err
	return e
}

// WithRetryable sets whether the error is retryable
func (e *GlobusError) WithRetryable(retryable bool) *GlobusError {
	e.Retryable = retryable
	return e
}

// IsRetryable returns whether the error indicates a retryable condition
func (e *GlobusError) IsRetryable() bool {
	return e.Retryable
}

// IsAuthenticationError returns true if the error is related to authentication
func (e *GlobusError) IsAuthenticationError() bool {
	return e.HTTPStatus == http.StatusUnauthorized ||
		strings.Contains(strings.ToLower(e.Code), "auth") ||
		strings.Contains(strings.ToLower(e.Message), "unauthorized")
}

// IsAuthorizationError returns true if the error is related to authorization
func (e *GlobusError) IsAuthorizationError() bool {
	return e.HTTPStatus == http.StatusForbidden ||
		strings.Contains(strings.ToLower(e.Code), "forbidden") ||
		strings.Contains(strings.ToLower(e.Message), "forbidden")
}

// IsNotFoundError returns true if the error indicates a resource was not found
func (e *GlobusError) IsNotFoundError() bool {
	return e.HTTPStatus == http.StatusNotFound ||
		strings.Contains(strings.ToLower(e.Code), "not_found") ||
		strings.Contains(strings.ToLower(e.Message), "not found")
}

// IsRateLimitError returns true if the error indicates rate limiting
func (e *GlobusError) IsRateLimitError() bool {
	return e.HTTPStatus == http.StatusTooManyRequests ||
		strings.Contains(strings.ToLower(e.Code), "rate_limit") ||
		strings.Contains(strings.ToLower(e.Message), "rate limit")
}

// IsServerError returns true if the error is a server-side error
func (e *GlobusError) IsServerError() bool {
	return e.HTTPStatus >= 500
}

// IsClientError returns true if the error is a client-side error
func (e *GlobusError) IsClientError() bool {
	return e.HTTPStatus >= 400 && e.HTTPStatus < 500
}

// isRetryableHTTPStatus determines if an HTTP status code indicates a retryable error
func isRetryableHTTPStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// Common error codes for each service
const (
	// Auth Service Error Codes
	AuthInvalidGrant      = "invalid_grant"
	AuthInvalidClient     = "invalid_client"
	AuthInvalidScope      = "invalid_scope"
	AuthInvalidToken      = "invalid_token"
	AuthExpiredToken      = "expired_token"
	AuthInsufficientScope = "insufficient_scope"

	// Transfer Service Error Codes
	TransferTaskNotFound     = "TaskNotFound"
	TransferEndpointNotFound = "EndpointNotFound"
	TransferPermissionDenied = "PermissionDenied"
	TransferInvalidPath      = "InvalidPath"
	TransferTaskSubmitError  = "TaskSubmitError"

	// Groups Service Error Codes
	GroupsGroupNotFound    = "GroupNotFound"
	GroupsMemberNotFound   = "MemberNotFound"
	GroupsPermissionDenied = "PermissionDenied"
	GroupsInvalidRequest   = "InvalidRequest"

	// Search Service Error Codes
	SearchIndexNotFound    = "IndexNotFound"
	SearchQueryError       = "QueryError"
	SearchIngestError      = "IngestError"
	SearchPermissionDenied = "PermissionDenied"

	// Flows Service Error Codes
	FlowsFlowNotFound     = "FlowNotFound"
	FlowsRunNotFound      = "RunNotFound"
	FlowsExecutionError   = "ExecutionError"
	FlowsPermissionDenied = "PermissionDenied"

	// Compute Service Error Codes
	ComputeFunctionNotFound = "FunctionNotFound"
	ComputeTaskNotFound     = "TaskNotFound"
	ComputeExecutionError   = "ExecutionError"
	ComputeEndpointOffline  = "EndpointOffline"

	// Timers Service Error Codes
	TimersJobNotFound      = "JobNotFound"
	TimersScheduleError    = "ScheduleError"
	TimersPermissionDenied = "PermissionDenied"
)

// Service-specific error constructors

// NewAuthError creates a new authentication service error
func NewAuthError(code, message string) *GlobusError {
	return NewGlobusError("auth", code, message)
}

// NewTransferError creates a new transfer service error
func NewTransferError(code, message string) *GlobusError {
	return NewGlobusError("transfer", code, message)
}

// NewGroupsError creates a new groups service error
func NewGroupsError(code, message string) *GlobusError {
	return NewGlobusError("groups", code, message)
}

// NewSearchError creates a new search service error
func NewSearchError(code, message string) *GlobusError {
	return NewGlobusError("search", code, message)
}

// NewFlowsError creates a new flows service error
func NewFlowsError(code, message string) *GlobusError {
	return NewGlobusError("flows", code, message)
}

// NewComputeError creates a new compute service error
func NewComputeError(code, message string) *GlobusError {
	return NewGlobusError("compute", code, message)
}

// NewTimersError creates a new timers service error
func NewTimersError(code, message string) *GlobusError {
	return NewGlobusError("timers", code, message)
}

// ParseGlobusErrorFromJSON parses a GlobusError from JSON response body
func ParseGlobusErrorFromJSON(service string, data []byte) (*GlobusError, error) {
	// Try to parse as a structured error response
	var errorResponse struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Detail  string `json:"detail"`
		} `json:"error"`
		Message   string `json:"message"`    // Some services use this format
		ErrorCode string `json:"error_code"` // Alternative format
	}

	if err := json.Unmarshal(data, &errorResponse); err != nil {
		return nil, err
	}

	globusErr := NewGlobusError(service, "", "")

	// Extract error information from various formats
	if errorResponse.Error.Code != "" {
		globusErr.Code = errorResponse.Error.Code
		globusErr.Message = errorResponse.Error.Message
		globusErr.Detail = errorResponse.Error.Detail
	} else if errorResponse.ErrorCode != "" {
		globusErr.Code = errorResponse.ErrorCode
		globusErr.Message = errorResponse.Message
	} else if errorResponse.Message != "" {
		globusErr.Code = "UnknownError"
		globusErr.Message = errorResponse.Message
	} else {
		return nil, fmt.Errorf("unable to parse error from JSON")
	}

	return globusErr, nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

// This file defines error types specific to the Flows service, as well as
// methods for checking error types. The error handling is designed to work
// with both service-specific errors (like FlowNotFoundError) and generic
// core.Error instances that might be returned from the underlying HTTP client.
// This robust error handling enables consistent error checking regardless of
// whether errors are created by the service client or by the core HTTP layer.

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// ErrorResponse represents an error response from the Globus Flows API.
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Resource  string `json:"resource,omitempty"`
}

// Error implements the error interface for ErrorResponse.
func (e *ErrorResponse) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("flows error [%s] %s (request_id: %s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("flows error [%s] %s", e.Code, e.Message)
}

// FlowNotFoundError indicates that a requested flow was not found.
type FlowNotFoundError struct {
	FlowID string
	*ErrorResponse
}

// Error implements the error interface for FlowNotFoundError.
func (e *FlowNotFoundError) Error() string {
	return fmt.Sprintf("flow not found: %s", e.FlowID)
}

// RunNotFoundError indicates that a requested run was not found.
type RunNotFoundError struct {
	RunID string
	*ErrorResponse
}

// Error implements the error interface for RunNotFoundError.
func (e *RunNotFoundError) Error() string {
	return fmt.Sprintf("run not found: %s", e.RunID)
}

// ActionProviderNotFoundError indicates that a requested action provider was not found.
type ActionProviderNotFoundError struct {
	ProviderID string
	*ErrorResponse
}

// Error implements the error interface for ActionProviderNotFoundError.
func (e *ActionProviderNotFoundError) Error() string {
	return fmt.Sprintf("action provider not found: %s", e.ProviderID)
}

// ActionRoleNotFoundError indicates that a requested action role was not found.
type ActionRoleNotFoundError struct {
	ProviderID string
	RoleID     string
	*ErrorResponse
}

// Error implements the error interface for ActionRoleNotFoundError.
func (e *ActionRoleNotFoundError) Error() string {
	return fmt.Sprintf("action role not found: %s (provider: %s)", e.RoleID, e.ProviderID)
}

// ForbiddenError indicates that the user does not have permission to perform the requested action.
type ForbiddenError struct {
	*ErrorResponse
}

// Error implements the error interface for ForbiddenError.
func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Message)
}

// ValidationError indicates that the request failed validation.
type ValidationError struct {
	*ErrorResponse
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ParseErrorResponse attempts to parse an HTTP response body into an ErrorResponse.
func ParseErrorResponse(body []byte, statusCode int, resourceID string, resourceType string) error {
	var errResponse ErrorResponse
	err := json.Unmarshal(body, &errResponse)
	if err != nil {
		// If we can't parse the error response, create a generic one
		return &ErrorResponse{
			Code:    fmt.Sprintf("HTTP%d", statusCode),
			Message: fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
		}
	}

	// Create appropriate error based on status code and resource type
	switch statusCode {
	case http.StatusNotFound:
		switch resourceType {
		case "flow":
			return &FlowNotFoundError{
				FlowID:        resourceID,
				ErrorResponse: &errResponse,
			}
		case "run":
			return &RunNotFoundError{
				RunID:         resourceID,
				ErrorResponse: &errResponse,
			}
		case "action_provider":
			return &ActionProviderNotFoundError{
				ProviderID:    resourceID,
				ErrorResponse: &errResponse,
			}
		case "action_role":
			// For action roles, resourceID is in the format "providerID:roleID"
			providerID, roleID := parseResourceIDs(resourceID)
			return &ActionRoleNotFoundError{
				ProviderID:    providerID,
				RoleID:        roleID,
				ErrorResponse: &errResponse,
			}
		default:
			return &ErrorResponse{
				Code:    errResponse.Code,
				Message: errResponse.Message,
			}
		}
	case http.StatusForbidden:
		return &ForbiddenError{
			ErrorResponse: &errResponse,
		}
	case http.StatusBadRequest:
		return &ValidationError{
			ErrorResponse: &errResponse,
		}
	default:
		return &errResponse
	}
}

// parseResourceIDs parses a combined resource ID of the form "providerID:roleID".
func parseResourceIDs(resourceID string) (string, string) {
	for i, c := range resourceID {
		if c == ':' {
			return resourceID[:i], resourceID[i+1:]
		}
	}
	return resourceID, ""
}

// IsFlowNotFoundError checks if an error is a FlowNotFoundError.
func IsFlowNotFoundError(err error) bool {
	// Direct type check
	if _, ok := err.(*FlowNotFoundError); ok {
		return true
	}

	// Check if it's a core.Error with 404 status
	if core.IsNotFound(err) {
		return true
	}

	return false
}

// IsRunNotFoundError checks if an error is a RunNotFoundError.
func IsRunNotFoundError(err error) bool {
	// Direct type check
	if _, ok := err.(*RunNotFoundError); ok {
		return true
	}

	// Check if it's a core.Error with 404 status
	if core.IsNotFound(err) {
		return true
	}

	return false
}

// IsActionProviderNotFoundError checks if an error is an ActionProviderNotFoundError.
func IsActionProviderNotFoundError(err error) bool {
	// Direct type check
	if _, ok := err.(*ActionProviderNotFoundError); ok {
		return true
	}

	// Check if it's a core.Error with 404 status
	if core.IsNotFound(err) {
		return true
	}

	return false
}

// IsActionRoleNotFoundError checks if an error is an ActionRoleNotFoundError.
func IsActionRoleNotFoundError(err error) bool {
	// Direct type check
	if _, ok := err.(*ActionRoleNotFoundError); ok {
		return true
	}

	// Check if it's a core.Error with 404 status
	if core.IsNotFound(err) {
		return true
	}

	return false
}

// IsForbiddenError checks if an error is a ForbiddenError.
func IsForbiddenError(err error) bool {
	// Direct type check
	if _, ok := err.(*ForbiddenError); ok {
		return true
	}

	// Check if it's a core.Error with 403 status
	coreErr, ok := err.(*core.Error)
	if ok && coreErr.StatusCode == http.StatusForbidden {
		return true
	}

	return false
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	// Direct type check
	if _, ok := err.(*ValidationError); ok {
		return true
	}

	// Check if it's a core.Error with 400 status
	coreErr, ok := err.(*core.Error)
	if ok && coreErr.StatusCode == http.StatusBadRequest {
		return true
	}

	return false
}

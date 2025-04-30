// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Common error codes returned by the Globus Transfer API
const (
	// Error codes for basic operations
	ErrCodeResourceNotFound       = "ResourceNotFound"
	ErrCodePermissionDenied       = "PermissionDenied"
	ErrCodeBadRequest             = "BadRequest"
	ErrCodeServiceUnavailable     = "ServiceUnavailable"
	ErrCodeRateLimitExceeded      = "RateLimitExceeded"
	ErrCodeAuthenticationRequired = "AuthenticationRequired"
	ErrCodeServerError            = "ServerError"
	
	// Task-specific error codes
	ErrCodeTaskNotFound           = "TaskNotFound"
	ErrCodeTaskExpired            = "TaskExpired"
	ErrCodeTaskCanceled           = "TaskCanceled"
	ErrCodeTaskCompleted          = "TaskCompleted"
	
	// Endpoint-specific error codes
	ErrCodeEndpointNotFound       = "EndpointNotFound"
	// Kept for backward compatibility but activation is now automatic in v0.10+
	ErrCodeEndpointNotActivated   = "EndpointNotActivated"
	ErrCodeEndpointError          = "EndpointError"
	
	// File operation error codes
	ErrCodeFileNotFound           = "FileNotFound"
	ErrCodeDirectoryNotFound      = "DirectoryNotFound"
	ErrCodeFileExists             = "FileExists"
	ErrCodeNoSuchPath             = "NoSuchPath"
	ErrCodeNotADirectory          = "NotADirectory"
	ErrCodePathCreationFailed     = "PathCreationFailed"
)

// Common errors that can be directly checked
var (
	// General errors
	ErrResourceNotFound       = errors.New("resource not found")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrBadRequest             = errors.New("bad request")
	ErrServerError            = errors.New("server error")
	ErrAuthenticationRequired = errors.New("authentication required")
	
	// Task-specific errors
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskExpired            = errors.New("task expired")
	ErrTaskCanceled           = errors.New("task canceled")
	ErrTaskCompleted          = errors.New("task completed")
	
	// Endpoint-specific errors
	ErrEndpointNotFound       = errors.New("endpoint not found")
	// Kept for backward compatibility but activation is now automatic in v0.10+
	ErrEndpointNotActivated   = errors.New("endpoint not activated")
	
	// File operation errors
	ErrFileNotFound           = errors.New("file not found")
	ErrDirectoryNotFound      = errors.New("directory not found")
	ErrFileExists             = errors.New("file already exists")
	ErrNoSuchPath             = errors.New("no such path")
	ErrNotADirectory          = errors.New("not a directory")
)

// TransferError represents an error from the Globus Transfer API
type TransferError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Resource    string `json:"resource,omitempty"`
	RequestID   string `json:"request_id,omitempty"`
	StatusCode  int    `json:"-"`
}

// Error returns a string representation of the error
func (e *TransferError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("%s: %s (request_id: %s)", e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsResourceNotFound checks if the error indicates a resource not found condition
func IsResourceNotFound(err error) bool {
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodeResourceNotFound ||
			   transferErr.Code == ErrCodeEndpointNotFound ||
			   transferErr.Code == ErrCodeFileNotFound ||
			   transferErr.Code == ErrCodeDirectoryNotFound ||
			   transferErr.Code == ErrCodeTaskNotFound ||
			   transferErr.Code == ErrCodeNoSuchPath
	}
	return errors.Is(err, ErrResourceNotFound) ||
		   errors.Is(err, ErrEndpointNotFound) ||
		   errors.Is(err, ErrFileNotFound) ||
		   errors.Is(err, ErrDirectoryNotFound) ||
		   errors.Is(err, ErrTaskNotFound) ||
		   errors.Is(err, ErrNoSuchPath)
}

// IsPermissionDenied checks if the error indicates a permission denied condition
func IsPermissionDenied(err error) bool {
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodePermissionDenied
	}
	return errors.Is(err, ErrPermissionDenied)
}

// IsRateLimitExceeded checks if the error indicates rate limiting
func IsRateLimitExceeded(err error) bool {
	// Check for TransferError
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodeRateLimitExceeded ||
			   (transferErr.StatusCode == http.StatusTooManyRequests)
	}
	
	// Check for core.Error
	var coreErr *core.Error
	if errors.As(err, &coreErr) {
		return coreErr.StatusCode == http.StatusTooManyRequests
	}
	
	// Check for wrapped error
	return errors.Is(err, ErrRateLimitExceeded)
}

// IsAuthenticationRequired checks if the error indicates authentication is required
func IsAuthenticationRequired(err error) bool {
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodeAuthenticationRequired ||
			   transferErr.StatusCode == http.StatusUnauthorized
	}
	return errors.Is(err, ErrAuthenticationRequired)
}

// IsEndpointNotActivated checks if the error indicates an endpoint needs activation
// NOTE: This function is kept for backward compatibility but explicit activation
// is no longer supported as modern Globus endpoints (v0.10+) auto-activate with
// properly scoped tokens.
func IsEndpointNotActivated(err error) bool {
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodeEndpointNotActivated
	}
	return errors.Is(err, ErrEndpointNotActivated)
}

// IsTaskCompleted checks if the error indicates a task is already completed
func IsTaskCompleted(err error) bool {
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		return transferErr.Code == ErrCodeTaskCompleted
	}
	return errors.Is(err, ErrTaskCompleted)
}

// parseTransferError parses an error response from the Globus Transfer API
func parseTransferError(statusCode int, respBody []byte) error {
	// If the response body is empty, return a generic error based on status code
	if len(respBody) == 0 {
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAuthenticationRequired
		case http.StatusForbidden:
			return ErrPermissionDenied
		case http.StatusNotFound:
			return ErrResourceNotFound
		case http.StatusTooManyRequests:
			return ErrRateLimitExceeded
		case http.StatusBadRequest:
			return ErrBadRequest
		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			return ErrServerError
		default:
			return fmt.Errorf("request failed with status code %d", statusCode)
		}
	}

	// Try to parse the error as JSON
	var errorResp map[string]interface{}
	if err := json.Unmarshal(respBody, &errorResp); err != nil {
		// If parsing fails, return the body as a string
		return fmt.Errorf("request failed with status code %d: %s", statusCode, string(respBody))
	}

	// Check if it's an OperationResult with an error code
	code, hasCode := errorResp["code"].(string)
	message, hasMessage := errorResp["message"].(string)
	if hasCode && hasMessage {
		transferErr := &TransferError{
			Code:       code,
			Message:    message,
			StatusCode: statusCode,
		}

		// Extract optional fields if present
		if resource, ok := errorResp["resource"].(string); ok {
			transferErr.Resource = resource
		}
		if requestID, ok := errorResp["request_id"].(string); ok {
			transferErr.RequestID = requestID
		}

		// Map common error codes to standard errors for easier checking
		switch code {
		case ErrCodeResourceNotFound, ErrCodeEndpointNotFound, ErrCodeTaskNotFound, 
			 ErrCodeFileNotFound, ErrCodeDirectoryNotFound, ErrCodeNoSuchPath:
			return fmt.Errorf("%w: %s", ErrResourceNotFound, transferErr.Error())
		case ErrCodePermissionDenied:
			return fmt.Errorf("%w: %s", ErrPermissionDenied, transferErr.Error())
		case ErrCodeRateLimitExceeded:
			return fmt.Errorf("%w: %s", ErrRateLimitExceeded, transferErr.Error())
		case ErrCodeAuthenticationRequired:
			return fmt.Errorf("%w: %s", ErrAuthenticationRequired, transferErr.Error())
		case ErrCodeEndpointNotActivated:
			return fmt.Errorf("%w: %s", ErrEndpointNotActivated, transferErr.Error())
		case ErrCodeTaskCompleted:
			return fmt.Errorf("%w: %s", ErrTaskCompleted, transferErr.Error())
		case ErrCodeTaskCanceled:
			return fmt.Errorf("%w: %s", ErrTaskCanceled, transferErr.Error())
		case ErrCodeTaskExpired:
			return fmt.Errorf("%w: %s", ErrTaskExpired, transferErr.Error())
		case ErrCodeServerError, ErrCodeServiceUnavailable:
			return fmt.Errorf("%w: %s", ErrServerError, transferErr.Error())
		default:
			return transferErr
		}
	}

	// Handle other error formats
	return fmt.Errorf("request failed with status code %d: %s", statusCode, string(respBody))
}

// IsRetryableTransferError determines if a Globus Transfer API error should be retried
func IsRetryableTransferError(err error) bool {
	// Check for TransferError
	var transferErr *TransferError
	if errors.As(err, &transferErr) {
		// Rate limit errors are always retryable
		if transferErr.Code == ErrCodeRateLimitExceeded || 
		   transferErr.StatusCode == http.StatusTooManyRequests {
			return true
		}

		// Server errors are generally retryable
		if transferErr.Code == ErrCodeServerError ||
		   transferErr.Code == ErrCodeServiceUnavailable ||
		   transferErr.StatusCode >= 500 && transferErr.StatusCode < 600 {
			return true
		}

		// Some endpoint errors might be temporary
		if strings.Contains(transferErr.Message, "temporarily") ||
		   strings.Contains(transferErr.Message, "retry") {
			return true
		}

		return false
	}
	
	// Check for core.Error
	var coreErr *core.Error
	if errors.As(err, &coreErr) {
		// Rate limit errors are always retryable
		if coreErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
		
		// Server errors are generally retryable
		if coreErr.StatusCode >= 500 && coreErr.StatusCode < 600 {
			return true
		}
		
		// Check error message for retry hints
		if strings.Contains(strings.ToLower(coreErr.Message), "temporarily") ||
		   strings.Contains(strings.ToLower(coreErr.Message), "retry") ||
		   strings.Contains(strings.ToLower(coreErr.Message), "rate limit") {
			return true
		}
		
		return false
	}

	// For standard errors, check if they match known retryable errors
	return errors.Is(err, ErrRateLimitExceeded) ||
		   errors.Is(err, ErrServerError)
}
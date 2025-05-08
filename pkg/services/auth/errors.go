// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Common error codes returned by the Globus Auth API
const (
	// Error codes for token endpoint
	ErrCodeInvalidGrant         = "invalid_grant"
	ErrCodeInvalidRequest       = "invalid_request"
	ErrCodeInvalidClient        = "invalid_client"
	ErrCodeInvalidScope         = "invalid_scope"
	ErrCodeUnsupportedTokenType = "unsupported_token_type"
	ErrCodeAccessDenied         = "access_denied"

	// Error codes for authorization endpoint
	ErrCodeServerError            = "server_error"
	ErrCodeTemporarilyUnavailable = "temporarily_unavailable"

	// Error codes for device flow
	ErrCodeAuthorizationPending = "authorization_pending"
	ErrCodeSlowDown             = "slow_down"
	ErrCodeExpiredToken         = "expired_token"
)

// Common errors that can be directly checked
var (
	// ErrInvalidGrant is returned when the grant (authorization code or refresh token) is invalid
	ErrInvalidGrant = errors.New("invalid grant")

	// ErrInvalidClient is returned when the client credentials are invalid
	ErrInvalidClient = errors.New("invalid client")

	// ErrInvalidScope is returned when the requested scope is invalid
	ErrInvalidScope = errors.New("invalid scope")

	// ErrAccessDenied is returned when the user denies the authorization request
	ErrAccessDenied = errors.New("access denied")

	// ErrServerError is returned when the server encounters an error
	ErrServerError = errors.New("server error")

	// ErrUnauthorized is returned when the request is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrBadRequest is returned when the request is malformed
	ErrBadRequest = errors.New("bad request")

	// Device flow errors
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrSlowDown             = errors.New("slow down polling")
	ErrExpiredToken         = errors.New("expired token")
)

// AuthError represents an error from the Globus Auth API
type AuthError struct {
	Code        string `json:"error"`
	Description string `json:"error_description,omitempty"`
	StatusCode  int    `json:"-"`
}

// Error returns a string representation of the error
func (e *AuthError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	}
	return e.Code
}

// IsInvalidGrant checks if the error is an invalid grant error
func IsInvalidGrant(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == ErrCodeInvalidGrant
	}
	return errors.Is(err, ErrInvalidGrant)
}

// IsInvalidClient checks if the error is an invalid client error
func IsInvalidClient(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == ErrCodeInvalidClient
	}
	return errors.Is(err, ErrInvalidClient)
}

// IsInvalidScope checks if the error is an invalid scope error
func IsInvalidScope(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == ErrCodeInvalidScope
	}
	return errors.Is(err, ErrInvalidScope)
}

// IsAccessDenied checks if the error is an access denied error
func IsAccessDenied(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == ErrCodeAccessDenied
	}
	return errors.Is(err, ErrAccessDenied)
}

// IsServerError checks if the error is a server error
func IsServerError(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code == ErrCodeServerError
	}
	return errors.Is(err, ErrServerError)
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.StatusCode == http.StatusUnauthorized
	}
	return errors.Is(err, ErrUnauthorized)
}

// IsBadRequest checks if the error is a bad request error
func IsBadRequest(err error) bool {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.StatusCode == http.StatusBadRequest
	}
	return errors.Is(err, ErrBadRequest)
}

// DeviceAuthError represents an error specific to the device authorization flow
type DeviceAuthError struct {
	Code        string
	Description string
}

// Error returns a string representation of the error
func (e *DeviceAuthError) Error() string {
	return fmt.Sprintf("Device flow error: %s - %s", e.Code, e.Description)
}

// IsDeviceAuthError checks if an error is a DeviceAuthError with the specified code
// If code is empty, it checks for any DeviceAuthError
func IsDeviceAuthError(err error, code string) bool {
	var devErr *DeviceAuthError
	if errors.As(err, &devErr) {
		return code == "" || devErr.Code == code
	}
	return false
}

// IsAuthorizationPending checks if the error indicates that device authorization is pending
func IsAuthorizationPending(err error) bool {
	return IsDeviceAuthError(err, ErrCodeAuthorizationPending) || errors.Is(err, ErrAuthorizationPending)
}

// IsSlowDown checks if the error indicates that polling should slow down
func IsSlowDown(err error) bool {
	return IsDeviceAuthError(err, ErrCodeSlowDown) || errors.Is(err, ErrSlowDown)
}

// IsExpiredToken checks if the error indicates that the token has expired
func IsExpiredToken(err error) bool {
	return IsDeviceAuthError(err, ErrCodeExpiredToken) || errors.Is(err, ErrExpiredToken)
}

// parseAuthError parses an error response from the Globus Auth API
func parseAuthError(statusCode int, respBody []byte) error {
	// If the response body is empty, return a generic error based on status code
	if len(respBody) == 0 {
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrUnauthorized
		case http.StatusBadRequest:
			return ErrBadRequest
		default:
			return fmt.Errorf("request failed with status code %d", statusCode)
		}
	}

	// Try to parse the error as JSON
	var authErr AuthError
	if err := json.Unmarshal(respBody, &authErr); err != nil {
		// If parsing fails, return the body as a string
		return fmt.Errorf("request failed with status code %d: %s", statusCode, string(respBody))
	}

	// Set the status code for later checking
	authErr.StatusCode = statusCode

	// Map common error codes to standard errors or return the AuthError directly
	switch authErr.Code {
	case ErrCodeInvalidGrant:
		// Check specifically for refresh token expiration
		if strings.Contains(strings.ToLower(authErr.Description), "refresh token") &&
			strings.Contains(strings.ToLower(authErr.Description), "expired") {
			return fmt.Errorf("%w: %s", ErrTokenExpired, authErr.Description)
		}
		// Return the AuthError directly to maintain its type for other invalid_grant errors
		return &authErr
	case ErrCodeInvalidClient:
		return fmt.Errorf("%w: %s", ErrInvalidClient, authErr.Description)
	case ErrCodeInvalidScope:
		return fmt.Errorf("%w: %s", ErrInvalidScope, authErr.Description)
	case ErrCodeAccessDenied:
		return fmt.Errorf("%w: %s", ErrAccessDenied, authErr.Description)
	case ErrCodeServerError, ErrCodeTemporarilyUnavailable:
		return fmt.Errorf("%w: %s", ErrServerError, authErr.Description)
	default:
		return &authErr
	}
}

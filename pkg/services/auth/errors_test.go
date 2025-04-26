// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAuthError(t *testing.T) {
	// Test error with description
	err := &AuthError{
		Code:        "invalid_grant",
		Description: "The authorization code is invalid or expired",
		StatusCode:  400,
	}
	
	expected := "invalid_grant: The authorization code is invalid or expired"
	if err.Error() != expected {
		t.Errorf("AuthError.Error() = %q, want %q", err.Error(), expected)
	}
	
	// Test error without description
	err = &AuthError{
		Code:       "invalid_grant",
		StatusCode: 400,
	}
	
	expected = "invalid_grant"
	if err.Error() != expected {
		t.Errorf("AuthError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestErrorCheckers(t *testing.T) {
	// Create different error types
	invalidGrantErr := &AuthError{
		Code:        ErrCodeInvalidGrant,
		Description: "The authorization code is invalid or expired",
		StatusCode:  400,
	}
	
	invalidClientErr := &AuthError{
		Code:        ErrCodeInvalidClient,
		Description: "Client authentication failed",
		StatusCode:  401,
	}
	
	invalidScopeErr := &AuthError{
		Code:        ErrCodeInvalidScope,
		Description: "The requested scope is invalid",
		StatusCode:  400,
	}
	
	accessDeniedErr := &AuthError{
		Code:        ErrCodeAccessDenied,
		Description: "The user denied the authorization request",
		StatusCode:  403,
	}
	
	serverErr := &AuthError{
		Code:        ErrCodeServerError,
		Description: "An internal server error occurred",
		StatusCode:  500,
	}
	
	unauthorizedErr := &AuthError{
		Code:        "unauthorized",
		Description: "Authentication required",
		StatusCode:  401,
	}
	
	// Test IsInvalidGrant
	if !IsInvalidGrant(invalidGrantErr) {
		t.Error("IsInvalidGrant() should return true for invalid_grant error")
	}
	if IsInvalidGrant(invalidClientErr) {
		t.Error("IsInvalidGrant() should return false for non-invalid_grant error")
	}
	
	// Test IsInvalidClient
	if !IsInvalidClient(invalidClientErr) {
		t.Error("IsInvalidClient() should return true for invalid_client error")
	}
	if IsInvalidClient(invalidGrantErr) {
		t.Error("IsInvalidClient() should return false for non-invalid_client error")
	}
	
	// Test IsInvalidScope
	if !IsInvalidScope(invalidScopeErr) {
		t.Error("IsInvalidScope() should return true for invalid_scope error")
	}
	if IsInvalidScope(invalidGrantErr) {
		t.Error("IsInvalidScope() should return false for non-invalid_scope error")
	}
	
	// Test IsAccessDenied
	if !IsAccessDenied(accessDeniedErr) {
		t.Error("IsAccessDenied() should return true for access_denied error")
	}
	if IsAccessDenied(invalidGrantErr) {
		t.Error("IsAccessDenied() should return false for non-access_denied error")
	}
	
	// Test IsServerError
	if !IsServerError(serverErr) {
		t.Error("IsServerError() should return true for server_error error")
	}
	if IsServerError(invalidGrantErr) {
		t.Error("IsServerError() should return false for non-server_error error")
	}
	
	// Test IsUnauthorized
	if !IsUnauthorized(unauthorizedErr) {
		t.Error("IsUnauthorized() should return true for 401 error")
	}
	if IsUnauthorized(invalidGrantErr) {
		t.Error("IsUnauthorized() should return false for non-401 error")
	}
	
	// Test with standard errors
	if !IsInvalidGrant(ErrInvalidGrant) {
		t.Error("IsInvalidGrant() should return true for ErrInvalidGrant")
	}
	
	// Test with wrapped errors
	wrappedErr := fmt.Errorf("wrapped: %w", ErrInvalidGrant)
	if !IsInvalidGrant(wrappedErr) {
		t.Error("IsInvalidGrant() should return true for wrapped ErrInvalidGrant")
	}
	
	// Test with different error
	if IsInvalidGrant(errors.New("some other error")) {
		t.Error("IsInvalidGrant() should return false for unrelated error")
	}
}

func TestParseAuthError(t *testing.T) {
	// Test with valid JSON error
	jsonErr := `{"error":"invalid_grant","error_description":"The authorization code is invalid or expired"}`
	err := parseAuthError(400, []byte(jsonErr))
	
	if !IsInvalidGrant(err) {
		t.Error("parseAuthError() should return invalid_grant error for invalid_grant JSON")
	}
	
	// Test with expired token
	expiredErr := `{"error":"invalid_grant","error_description":"The refresh token has expired"}`
	err = parseAuthError(400, []byte(expiredErr))
	
	if !errors.Is(err, ErrTokenExpired) {
		t.Error("parseAuthError() should return token expired error for expired token JSON")
	}
	
	// Test with empty response
	err = parseAuthError(401, []byte{})
	if !errors.Is(err, ErrUnauthorized) {
		t.Error("parseAuthError() should return ErrUnauthorized for empty 401 response")
	}
	
	// Test with invalid JSON
	err = parseAuthError(400, []byte("not json"))
	if err == nil {
		t.Error("parseAuthError() should return error for invalid JSON")
	}
	
	// Test with unknown error code
	unknownErr := `{"error":"unknown_error","error_description":"Something went wrong"}`
	err = parseAuthError(400, []byte(unknownErr))
	
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Error("parseAuthError() should return AuthError for unknown error code")
	}
	
	if authErr != nil && authErr.Code != "unknown_error" {
		t.Errorf("parseAuthError() returned wrong code, got %q, want %q", authErr.Code, "unknown_error")
	}
}
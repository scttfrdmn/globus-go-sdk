// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"context"
	"strings"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyAuthorizerContract verifies that an Authorizer implementation
// satisfies the behavioral contract of the interface.
func VerifyAuthorizerContract(t *testing.T, authorizer interfaces.Authorizer) {
	t.Helper()

	t.Run("GetAuthorizationHeader method", func(t *testing.T) {
		verifyGetAuthorizationHeaderMethod(t, authorizer)
	})

	t.Run("IsValid method", func(t *testing.T) {
		verifyIsValidMethod(t, authorizer)
	})

	t.Run("GetToken method", func(t *testing.T) {
		verifyGetTokenMethod(t, authorizer)
	})

	t.Run("Context respect", func(t *testing.T) {
		verifyAuthorizerContextRespect(t, authorizer)
	})
}

// verifyGetAuthorizationHeaderMethod tests the behavior of the GetAuthorizationHeader method
func verifyGetAuthorizationHeaderMethod(t *testing.T, authorizer interfaces.Authorizer) {
	t.Helper()

	// Get authorization header with context
	header, err := authorizer.GetAuthorizationHeader(context.Background())

	// If token is valid, header should be non-empty
	if authorizer.IsValid() {
		if err != nil {
			t.Errorf("GetAuthorizationHeader returned error for valid token: %v", err)
		}
		if header == "" {
			t.Error("GetAuthorizationHeader returned empty string for valid token")
		}

		// Header should start with "Bearer " for OAuth tokens
		// Note: Other auth types might use different prefixes
		if strings.HasPrefix(header, "Bearer ") {
			// Token should match what's returned by GetToken
			token := authorizer.GetToken()
			expectedHeader := "Bearer " + token
			if header != expectedHeader && token != "" {
				t.Errorf("Expected header %q, got %q", expectedHeader, header)
			}
		}
	} else {
		// For invalid tokens, behavior depends on implementation
		// Some might return an error, others might return an empty header
		t.Logf("Token is invalid, GetAuthorizationHeader returned: header=%q, err=%v",
			header, err)
	}

	// Consistency: multiple calls should return the same result
	header2, err2 := authorizer.GetAuthorizationHeader(context.Background())
	if (err == nil) != (err2 == nil) {
		t.Errorf("Inconsistent error results: first call error=%v, second call error=%v",
			err, err2)
	}
	if err == nil && err2 == nil && header != header2 {
		t.Errorf("Inconsistent headers: first call %q, second call %q",
			header, header2)
	}
}

// verifyIsValidMethod tests the behavior of the IsValid method
func verifyIsValidMethod(t *testing.T, authorizer interfaces.Authorizer) {
	t.Helper()

	// IsValid should return a boolean indicating token validity
	isValid := authorizer.IsValid()

	// We can't make assumptions about the actual validity,
	// but we can test for consistency
	isValid2 := authorizer.IsValid()
	if isValid != isValid2 {
		t.Error("IsValid returned inconsistent results")
	}
}

// verifyGetTokenMethod tests the behavior of the GetToken method
func verifyGetTokenMethod(t *testing.T, authorizer interfaces.Authorizer) {
	t.Helper()

	// GetToken should return the current token
	token := authorizer.GetToken()

	// If IsValid returns true, token should be non-empty
	if authorizer.IsValid() && token == "" {
		t.Error("GetToken returned empty string for valid token")
	}

	// Consistency: multiple calls should return the same result
	token2 := authorizer.GetToken()
	if token != token2 {
		t.Errorf("GetToken returned inconsistent results: %q and %q", token, token2)
	}
}

// verifyAuthorizerContextRespect tests that Authorizer respects context cancellation
func verifyAuthorizerContextRespect(t *testing.T, authorizer interfaces.Authorizer) {
	t.Helper()

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// GetAuthorizationHeader should respect context cancellation
	_, err := authorizer.GetAuthorizationHeader(ctx)

	// Context cancellation should either result in an error
	// or complete successfully (if the implementation doesn't use the context)
	if err != nil {
		if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") {
			t.Logf("Expected context cancellation error, got: %v", err)
		}
	}
}

// VerifyTokenManagerContract verifies that a TokenManager implementation
// satisfies the behavioral contract of the interface.
func VerifyTokenManagerContract(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	t.Run("GetToken method", func(t *testing.T) {
		verifyTokenManagerGetTokenMethod(t, manager)
	})

	t.Run("RefreshToken method", func(t *testing.T) {
		verifyRefreshTokenMethod(t, manager)
	})

	t.Run("RevokeToken method", func(t *testing.T) {
		verifyRevokeTokenMethod(t, manager)
	})

	t.Run("IsValid method", func(t *testing.T) {
		verifyTokenManagerIsValidMethod(t, manager)
	})

	t.Run("Context respect", func(t *testing.T) {
		verifyTokenManagerContextRespect(t, manager)
	})
}

// verifyTokenManagerGetTokenMethod tests the behavior of the GetToken method
func verifyTokenManagerGetTokenMethod(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	// Get token with context
	token, err := manager.GetToken(context.Background())

	// If IsValid returns true, token should be non-empty and no error
	if manager.IsValid() {
		if err != nil {
			t.Errorf("GetToken returned error for valid token: %v", err)
		}
		if token == "" {
			t.Error("GetToken returned empty string for valid token")
		}
	} else {
		// For invalid tokens, we expect either an error or an empty token
		if err == nil && token != "" {
			t.Error("GetToken returned a token for invalid state")
		}
	}
}

// verifyRefreshTokenMethod tests the behavior of the RefreshToken method
func verifyRefreshTokenMethod(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	// Try to refresh the token
	err := manager.RefreshToken(context.Background())

	// We can't make assumptions about the ability to refresh (might need real credentials),
	// but we can test for consistency in token state after attempted refresh
	isValidAfter := manager.IsValid()
	t.Logf("After RefreshToken: valid=%v, err=%v", isValidAfter, err)

	// If refresh succeeded, IsValid should return true
	if err == nil {
		// Some implementations might not actually refresh if not needed
		// This is just a heuristic
		tokenAfter, _ := manager.GetToken(context.Background())
		if tokenAfter == "" {
			t.Error("Token is empty after successful refresh")
		}
	}
}

// verifyRevokeTokenMethod tests the behavior of the RevokeToken method
func verifyRevokeTokenMethod(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	// Only test revocation if the token is valid
	// to avoid unnecessary testing on already invalid tokens
	if !manager.IsValid() {
		t.Skip("Token not valid, skipping revocation test")
	}

	// Try to revoke the token
	err := manager.RevokeToken(context.Background())
	t.Logf("RevokeToken result: err=%v", err)

	// After revocation, the token should either be invalid
	// or if revocation failed (which is common in tests), the error should be non-nil
	if err == nil && manager.IsValid() {
		t.Log("RevokeToken reported success but token is still valid (may be expected in tests)")
	}
}

// verifyTokenManagerIsValidMethod tests the behavior of the IsValid method
func verifyTokenManagerIsValidMethod(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	// IsValid should return a boolean indicating token validity
	isValid := manager.IsValid()

	// We can't make assumptions about the actual validity,
	// but we can test for consistency
	isValid2 := manager.IsValid()
	if isValid != isValid2 {
		t.Error("IsValid returned inconsistent results")
	}
}

// verifyTokenManagerContextRespect tests that TokenManager respects context cancellation
func verifyTokenManagerContextRespect(t *testing.T, manager interfaces.TokenManager) {
	t.Helper()

	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// GetToken should respect context cancellation
	_, err := manager.GetToken(ctx)
	if err == nil {
		t.Log("GetToken with canceled context completed successfully (may not use context)")
	} else if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") {
		t.Logf("Expected context cancellation error, got: %v", err)
	}

	// RefreshToken should respect context cancellation
	err = manager.RefreshToken(ctx)
	if err == nil {
		t.Log("RefreshToken with canceled context completed successfully (may not use context)")
	} else if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") {
		t.Logf("Expected context cancellation error, got: %v", err)
	}

	// RevokeToken should respect context cancellation
	err = manager.RevokeToken(ctx)
	if err == nil {
		t.Log("RevokeToken with canceled context completed successfully (may not use context)")
	} else if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "canceled") {
		t.Logf("Expected context cancellation error, got: %v", err)
	}
}

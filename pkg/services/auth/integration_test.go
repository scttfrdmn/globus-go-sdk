//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
	
	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../../../.env.test")
	_ = godotenv.Load("../../.env.test")
	_ = godotenv.Load(".env.test")
}

func skipIfMissingCredentials(t *testing.T) (string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID environment variable")
	}
	
	if clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_SECRET environment variable")
	}

	return clientID, clientSecret
}

func TestIntegration_ClientCredentialsFlow(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with new initialization pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test getting token with client credentials with retry for rate limiting
	var tokenResp *TokenResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var tokenErr error
			tokenResp, tokenErr = client.GetClientCredentialsToken(ctx, AuthScope)
			return tokenErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			// Check for rate limiting or transient errors
			if err == nil {
				return false
			}
			return true // Retry all errors for this test
		},
	)
	
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Validate response
	if tokenResp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if tokenResp.TokenType != "Bearer" {
		t.Errorf("Expected token_type=Bearer, got %s", tokenResp.TokenType)
	}
	if tokenResp.ExpiresIn <= 0 {
		t.Errorf("Expected positive expires_in, got %d", tokenResp.ExpiresIn)
	}

	// Test token is valid with retry for rate limiting
	var tokenInfo *TokenInfo
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var introspectErr error
			tokenInfo, introspectErr = client.IntrospectToken(ctx, tokenResp.AccessToken)
			return introspectErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			// Check for rate limiting or transient errors
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("IntrospectToken failed: %v", err)
	}

	if !tokenInfo.Active {
		t.Error("Expected token to be active")
	}
	if tokenInfo.ClientID != clientID {
		t.Errorf("Expected client_id=%s, got %s", clientID, tokenInfo.ClientID)
	}
	
	t.Logf("Successfully validated client credentials flow with token access")
}

func TestIntegration_TokenUtils(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with new initialization pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get a token to validate with retry
	var tokenResp *TokenResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var tokenErr error
			tokenResp, tokenErr = client.GetClientCredentialsToken(ctx, AuthScope)
			return tokenErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Test token validity with retry
	var isValid bool
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var tokenErr error
			isValid, tokenErr = client.IsTokenValid(ctx, tokenResp.AccessToken)
			if tokenErr != nil {
				return tokenErr
			}
			if !isValid {
				return fmt.Errorf("token validity check failed")
			}
			return nil
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Errorf("Expected token to be valid: %v", err)
	}

	// Test GetTokenExpiry with retry
	var expiry time.Time
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var expiryErr error
			expiry, isValid, expiryErr = client.GetTokenExpiry(ctx, tokenResp.AccessToken)
			return expiryErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("GetTokenExpiry failed: %v", err)
	}
	
	if !isValid {
		t.Error("Expected token to be valid")
	}
	if !expiry.After(time.Now()) {
		t.Errorf("Expected expiry to be in the future, got %v", expiry)
	}

	// Test ShouldRefresh with minimum threshold with retry
	var shouldRefresh bool
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var refreshErr error
			shouldRefresh, refreshErr = client.ShouldRefresh(ctx, tokenResp.AccessToken, 5*time.Second)
			return refreshErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("ShouldRefresh failed: %v", err)
	}
	
	// Token should be fresh since we just got it
	if shouldRefresh {
		t.Error("Expected token to not need refresh with short threshold")
	}

	// Log token details for debugging
	t.Logf("Token lifetime: %d seconds", tokenResp.ExpiresIn)
	t.Logf("Token expires at: %v", tokenResp.ExpiryTime)
	t.Logf("Current time: %v", time.Now())
}

func TestIntegration_ClientCredentialsAuthorizer(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with new initialization pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client credentials authorizer
	authorizer := client.CreateClientCredentialsAuthorizer(AuthScope)

	// Test getting an authorization header with retry
	var authorizationHeader string
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var authErr error
			authorizationHeader, authErr = authorizer.GetAuthorizationHeader(ctx)
			return authErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("Authorizer.GetAuthorizationHeader failed: %v", err)
	}

	// Validate response
	if authorizationHeader == "" {
		t.Error("Expected non-empty authorization header")
	}

	// Verify header format
	if len(authorizationHeader) < 8 || authorizationHeader[:7] != "Bearer " {
		t.Errorf("Expected authorization header to start with 'Bearer ', got: %s", authorizationHeader)
	}
	
	t.Logf("Successfully validated client credentials authorizer")
}

func TestIntegration_StaticTokenAuthorizer(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with new initialization pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get a token to use with retry
	var tokenResp *TokenResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var tokenErr error
			tokenResp, tokenErr = client.GetClientCredentialsToken(ctx, AuthScope)
			return tokenErr
		},
		ratelimit.DefaultBackoff(),
		func(err error) bool {
			if err == nil {
				return false
			}
			return true
		},
	)
	
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Create static token authorizer
	authorizer := authorizers.NewStaticTokenAuthorizer(tokenResp.AccessToken)

	// Test getting an authorization header
	authorizationHeader, err := authorizer.GetAuthorizationHeader(ctx)
	if err != nil {
		t.Fatalf("Authorizer.GetAuthorizationHeader failed: %v", err)
	}

	// Validate response
	if authorizationHeader == "" {
		t.Error("Expected non-empty authorization header")
	}

	// Verify header format matches the token
	expectedHeader := "Bearer " + tokenResp.AccessToken
	if authorizationHeader != expectedHeader {
		t.Errorf("Expected header %s, got: %s", expectedHeader[:15]+"...", authorizationHeader[:15]+"...")
	}
	
	t.Logf("Successfully validated static token authorizer")
}
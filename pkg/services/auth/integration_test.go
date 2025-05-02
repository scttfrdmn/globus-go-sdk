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

	// Create client with the new pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test getting token with client credentials with retry for rate limiting
	var tokenResp *TokenResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var tokenErr error
			tokenResp, tokenErr = client.GetClientCredentialsToken(ctx, []string{AuthScope})
			return tokenErr
		},
		ratelimit.DefaultBackoff(),
		ratelimit.IsRetryableError,
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
	var tokenInfo *TokenIntrospectionResponse
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var introspectErr error
			tokenInfo, introspectErr = client.IntrospectToken(ctx, tokenResp.AccessToken)
			return introspectErr
		},
		ratelimit.DefaultBackoff(),
		ratelimit.IsRetryableError,
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
}

func TestIntegration_TokenUtils(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with the new pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get a token to validate
	tokenResp, err := client.GetClientCredentialsToken(ctx, []string{AuthScope})
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Test IsTokenValid
	// Test token validity with retry
	var valid bool
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			valid = client.IsTokenValid(ctx, tokenResp.AccessToken)
			if !valid {
				return fmt.Errorf("token validity check failed")
			}
			return nil
		},
		ratelimit.DefaultBackoff(),
		ratelimit.IsRetryableError,
	)
	
	if err != nil {
		t.Errorf("Expected token to be valid: %v", err)
	}

	// Test GetTokenExpiry with retry
	var expiry time.Time
	var expiryValid bool
	err = ratelimit.RetryWithBackoff(
		ctx,
		func(ctx context.Context) error {
			var expiryErr error
			expiry, expiryValid, expiryErr = client.GetTokenExpiry(ctx, tokenResp.AccessToken)
			return expiryErr
		},
		ratelimit.DefaultBackoff(),
		ratelimit.IsRetryableError,
	)
	
	if err != nil {
		t.Fatalf("GetTokenExpiry failed: %v", err)
	}
	valid = expiryValid // Update the valid flag with expiry result
	if !valid {
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
		ratelimit.IsRetryableError,
	)
	
	if err != nil {
		t.Fatalf("ShouldRefresh failed: %v", err)
	}
	
	// Token should be fresh since we just got it
	if shouldRefresh {
		t.Error("Expected token to not need refresh with short threshold")
	}

	// Test ShouldRefresh with maximum threshold 
	// (commented out since token will often have long lifetimes in testing environments)
	/*
	shouldRefresh, err = client.ShouldRefresh(ctx, tokenResp.AccessToken, 24*time.Hour)
	if err != nil {
		t.Fatalf("ShouldRefresh failed: %v", err)
	}
	// With a 24h threshold, token would need refreshing
	if !shouldRefresh {
		t.Error("Expected token to need refresh with long threshold")
	}
	*/
}

func TestIntegration_ClientCredentialsAuthorizer(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client with the new pattern
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create client credentials authorizer
	authorizer, err := client.CreateClientCredentialsAuthorizer(ctx, []string{AuthScope})
	if err != nil {
		t.Fatalf("Failed to create client credentials authorizer: %v", err)
	}

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
		ratelimit.IsRetryableError,
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
}

func TestIntegration_StaticTokenAuthorizer(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client and get a token to use
	client, err := NewClient(
		WithClientID(clientID),
		WithClientSecret(clientSecret),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	tokenResp, err := client.GetClientCredentialsToken(ctx, []string{AuthScope})
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Create static token authorizer
	authorizer := authorizers.NewStaticTokenAuthorizer(tokenResp.AccessToken)

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
		ratelimit.IsRetryableError,
	)
	
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
		t.Errorf("Expected header %s, got: %s", expectedHeader, authorizationHeader)
	}
}
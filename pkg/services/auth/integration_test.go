// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"os"
	"testing"
	"time"
)

func skipIfMissingCredentials(t *testing.T) (string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}

	return clientID, clientSecret
}

func TestIntegration_ClientCredentialsFlow(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create client
	client := NewClient(clientID, clientSecret)
	ctx := context.Background()

	// Test getting token with client credentials
	tokenResp, err := client.GetClientCredentialsToken(ctx, AuthScope)
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

	// Test token is valid
	tokenInfo, err := client.IntrospectToken(ctx, tokenResp.AccessToken)
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

	// Create client
	client := NewClient(clientID, clientSecret)
	ctx := context.Background()

	// Get a token to validate
	tokenResp, err := client.GetClientCredentialsToken(ctx, AuthScope)
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Test IsTokenValid
	valid := client.IsTokenValid(ctx, tokenResp.AccessToken)
	if !valid {
		t.Error("Expected token to be valid")
	}

	// Test GetTokenExpiry
	expiry, valid, err := client.GetTokenExpiry(ctx, tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("GetTokenExpiry failed: %v", err)
	}
	if !valid {
		t.Error("Expected token to be valid")
	}
	if !expiry.After(time.Now()) {
		t.Errorf("Expected expiry to be in the future, got %v", expiry)
	}

	// Test ShouldRefresh with minimum threshold
	shouldRefresh, err := client.ShouldRefresh(ctx, tokenResp.AccessToken, 5*time.Second)
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

	// Create client
	client := NewClient(clientID, clientSecret)
	ctx := context.Background()

	// Create client credentials authorizer
	authorizer := client.CreateClientCredentialsAuthorizer(AuthScope)

	// Test getting an authorization header
	authorizationHeader, err := authorizer.GetAuthorizationHeader(ctx)
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
	client := NewClient(clientID, clientSecret)
	ctx := context.Background()

	tokenResp, err := client.GetClientCredentialsToken(ctx, AuthScope)
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Create static token authorizer
	authorizer := client.CreateStaticTokenAuthorizer(tokenResp.AccessToken)

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
		t.Errorf("Expected header %s, got: %s", expectedHeader, authorizationHeader)
	}
}
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

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
	if tokenResp.ExpiresIn < 1 {
		t.Errorf("Expected positive expires_in, got %d", tokenResp.ExpiresIn)
	}
	
	// Verify we don't get a refresh token in client credentials flow
	if tokenResp.RefreshToken != "" {
		t.Errorf("Client credentials flow should not return refresh token, got %s", tokenResp.RefreshToken)
	}
	
	// Verify token introspection
	tokenInfo, err := client.IntrospectToken(ctx, tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("IntrospectToken failed: %v", err)
	}
	
	if !tokenInfo.Active {
		t.Error("Token should be active")
	}
	if tokenInfo.ClientID != clientID {
		t.Errorf("Expected client_id=%s, got %s", clientID, tokenInfo.ClientID)
	}
	
	// Verify token revocation
	err = client.RevokeToken(ctx, tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("RevokeToken failed: %v", err)
	}
	
	// Verify token is no longer active
	tokenInfo, err = client.IntrospectToken(ctx, tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("IntrospectToken after revocation failed: %v", err)
	}
	
	if tokenInfo.Active {
		t.Error("Token should be inactive after revocation")
	}
}

func TestIntegration_ClientCredentialsAuthorizer(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)
	
	// Create client
	client := NewClient(clientID, clientSecret)
	ctx := context.Background()
	
	// Create client credentials authorizer
	authorizer := client.CreateClientCredentialsAuthorizer(AuthScope)
	
	// Test getting a token
	token, expiresAt, err := authorizer.GetToken(ctx)
	if err != nil {
		t.Fatalf("Authorizer.GetToken failed: %v", err)
	}
	
	// Validate response
	if token == "" {
		t.Error("Expected non-empty token")
	}
	
	// Verify expiry time is in the future
	if !expiresAt.After(time.Now()) {
		t.Errorf("Expected expiry time to be in the future, got %v", expiresAt)
	}
	
	// Verify authorizer can add token to request
	req, err := client.Client.BuildRequest(ctx, "GET", "test", nil, nil)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	
	err = authorizer.AuthorizeRequest(ctx, req)
	if err != nil {
		t.Fatalf("AuthorizeRequest failed: %v", err)
	}
	
	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer "+token {
		t.Errorf("Expected Authorization=Bearer %s, got %s", token, authHeader)
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
	
	// Verify authorizer can add token to request
	req, err := client.Client.BuildRequest(ctx, "GET", "test", nil, nil)
	if err != nil {
		t.Fatalf("BuildRequest failed: %v", err)
	}
	
	err = authorizer.AuthorizeRequest(ctx, req)
	if err != nil {
		t.Fatalf("AuthorizeRequest failed: %v", err)
	}
	
	authHeader := req.Header.Get("Authorization")
	expectedHeader := "Bearer " + tokenResp.AccessToken
	if authHeader != expectedHeader {
		t.Errorf("Expected Authorization=%s, got %s", expectedHeader, authHeader)
	}
}
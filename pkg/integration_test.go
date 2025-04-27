//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pkg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func skipIfMissingCredentials(t *testing.T) (string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}

	return clientID, clientSecret
}

func TestIntegration_SDKConfig(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Verify config values
	if config.ClientID != clientID {
		t.Errorf("Config ClientID = %s, want %s", config.ClientID, clientID)
	}
	if config.ClientSecret != clientSecret {
		t.Errorf("Config ClientSecret = %s, want %s", config.ClientSecret, clientSecret)
	}

	// Create Auth client
	authClient := config.NewAuthClient()

	// Test client credentials flow to verify client works
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, auth.AuthScope)
	if err != nil {
		t.Fatalf("Auth client failed: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	// Create Groups client
	groupsClient := config.NewGroupsClient(tokenResp.AccessToken)

	// Test list groups to verify client works
	groups, err := groupsClient.ListGroups(ctx, nil)
	if err != nil {
		t.Fatalf("Groups client failed: %v", err)
	}

	t.Logf("Found %d groups", len(groups.Groups))

	// Create Transfer client
	transferClient := config.NewTransferClient(tokenResp.AccessToken)

	// Test list endpoints to verify client works
	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		t.Fatalf("Transfer client failed: %v", err)
	}

	t.Logf("Found %d endpoints", len(endpoints.DATA))

	// Create Search client
	searchClient := config.NewSearchClient(tokenResp.AccessToken)

	// Test list indexes to verify client works
	indexes, err := searchClient.ListIndexes(ctx, &search.ListIndexesOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search client failed: %v", err)
	}

	t.Logf("Found %d search indexes", len(indexes.Indexes))

	// Create Flows client
	flowsClient := config.NewFlowsClient(tokenResp.AccessToken)

	// Test list flows to verify client works
	flowsList, err := flowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Flows client failed: %v", err)
	}

	t.Logf("Found %d flows", len(flowsList.Flows))
}

func TestIntegration_GetScopesByService(t *testing.T) {
	// Test getting scopes for individual services
	authScopes := GetScopesByService("auth")
	if len(authScopes) != 1 || authScopes[0] != auth.AuthScope {
		t.Errorf("Auth scope = %v, want %v", authScopes, []string{auth.AuthScope})
	}

	groupsScopes := GetScopesByService("groups")
	if len(groupsScopes) != 1 || groupsScopes[0] != GroupsScope {
		t.Errorf("Groups scope = %v, want %v", groupsScopes, []string{GroupsScope})
	}

	transferScopes := GetScopesByService("transfer")
	if len(transferScopes) != 1 || transferScopes[0] != TransferScope {
		t.Errorf("Transfer scope = %v, want %v", transferScopes, []string{TransferScope})
	}

	searchScopes := GetScopesByService("search")
	if len(searchScopes) != 1 || searchScopes[0] != search.SearchScope {
		t.Errorf("Search scope = %v, want %v", searchScopes, []string{search.SearchScope})
	}

	flowsScopes := GetScopesByService("flows")
	if len(flowsScopes) != 1 || flowsScopes[0] != flows.FlowsScope {
		t.Errorf("Flows scope = %v, want %v", flowsScopes, []string{flows.FlowsScope})
	}

	// Test getting scopes for multiple services
	allScopes := GetScopesByService("auth", "groups", "transfer", "search", "flows")
	expectedScopes := []string{
		auth.AuthScope,
		GroupsScope,
		TransferScope,
		search.SearchScope,
		flows.FlowsScope,
	}

	if len(allScopes) != len(expectedScopes) {
		t.Errorf("All scopes length = %d, want %d", len(allScopes), len(expectedScopes))
	}

	// Check each scope is included
	for _, scope := range expectedScopes {
		found := false
		for _, s := range allScopes {
			if s == scope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Scope %s not found in all scopes", scope)
		}
	}
}

func TestIntegration_AuthClientFlow(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration and auth client
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	authClient := config.NewAuthClient()
	ctx := context.Background()

	// Get all scopes
	allScopes := GetScopesByService("auth", "groups", "transfer", "search", "flows")

	// Get client credentials token
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, allScopes...)
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Create a static token authorizer
	staticAuthorizer := authClient.CreateStaticTokenAuthorizer(tokenResp.AccessToken)

	// Get token from authorizer
	token, expiresAt, err := staticAuthorizer.GetToken(ctx)
	if err != nil {
		t.Fatalf("StaticTokenAuthorizer.GetToken failed: %v", err)
	}

	if token != tokenResp.AccessToken {
		t.Errorf("Authorizer token = %s, want %s", token, tokenResp.AccessToken)
	}

	// For static token authorizer, expiry time should be far in the future
	if !expiresAt.After(time.Now().Add(time.Hour * 24 * 365)) {
		t.Error("Static token expiry should be far in the future")
	}

	// Create client credentials authorizer
	credentialsAuthorizer := authClient.CreateClientCredentialsAuthorizer(allScopes...)

	// Get token from authorizer
	credToken, credExpiresAt, err := credentialsAuthorizer.GetToken(ctx)
	if err != nil {
		t.Fatalf("ClientCredentialsAuthorizer.GetToken failed: %v", err)
	}

	if credToken == "" {
		t.Error("Expected non-empty token from credentials authorizer")
	}

	if !credExpiresAt.After(time.Now()) {
		t.Error("Token expiry should be in the future")
	}
}

func TestIntegration_TokenManager(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration and auth client
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)

	// Test the token manager with memory storage
	tokenManager := config.NewTokenManager()
	ctx := context.Background()

	// Get a token using client credentials flow
	allScopes := GetScopesByService("auth", "groups", "transfer")
	scopesKey := "integration_test_scopes"

	// Store a token
	tokenResp, err := config.NewAuthClient().GetClientCredentialsToken(ctx, allScopes...)
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	err = tokenManager.StoreToken(ctx, scopesKey, tokenResp)
	if err != nil {
		t.Fatalf("StoreToken failed: %v", err)
	}

	// Retrieve the token
	retrievedToken, err := tokenManager.GetToken(ctx, scopesKey)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if retrievedToken.AccessToken != tokenResp.AccessToken {
		t.Errorf("Retrieved token = %s, want %s", retrievedToken.AccessToken, tokenResp.AccessToken)
	}

	// Create token authorizer from manager
	authorizer := tokenManager.GetAuthorizer(scopesKey)

	// Get token from authorizer
	authToken, authExpiresAt, err := authorizer.GetToken(ctx)
	if err != nil {
		t.Fatalf("Authorizer.GetToken failed: %v", err)
	}

	if authToken != tokenResp.AccessToken {
		t.Errorf("Authorizer token = %s, want %s", authToken, tokenResp.AccessToken)
	}

	if !authExpiresAt.After(time.Now()) {
		t.Error("Token expiry should be in the future")
	}
}

func TestIntegration_RateLimiting(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration with rate limiting enabled
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret).
		WithRateLimiting(true, 5) // 5 requests per second

	// Create Auth client
	authClient := config.NewAuthClient()
	ctx := context.Background()

	// Make a series of requests to test rate limiting
	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := authClient.IntrospectToken(ctx, "dummy-token")
		// We expect errors due to invalid token, but the requests should still be rate limited
		if err != nil && !IsUnauthorizedError(err) && !IsTokenInvalidError(err) {
			t.Fatalf("Unexpected error on request %d: %v", i, err)
		}
	}
	duration := time.Since(start)

	// With 5 requests per second and 10 requests, it should take at least 2 seconds
	minDuration := 2 * time.Second
	if duration < minDuration {
		t.Errorf("Rate limiting not effective: %d requests took %v, expected at least %v", 10, duration, minDuration)
	} else {
		t.Logf("Rate limiting working as expected: %d requests took %v", 10, duration)
	}
}

func TestIntegration_CircuitBreaker(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration with circuit breaker enabled
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret).
		WithCircuitBreaker(true, 5, 10*time.Second) // Trip after 5 failures, 10 second reset

	// Create Auth client
	authClient := config.NewAuthClient()
	ctx := context.Background()

	// Make a request to a valid endpoint with invalid token to trigger 401s, not circuit breaking
	for i := 0; i < 5; i++ {
		_, err := authClient.IntrospectToken(ctx, "invalid-token")
		// We expect Unauthorized errors, but these shouldn't trip the circuit breaker
		if err != nil && !IsUnauthorizedError(err) && !IsTokenInvalidError(err) {
			t.Logf("Unexpected error type on request %d: %v", i, err)
		}
	}

	// Make a request to a non-existent endpoint to trigger 5xx errors
	// This isn't ideal, but we need to find a way to force server errors
	customAuthClient := auth.NewClient(clientID, clientSecret).WithBaseURL("https://auth.globus.org/v2/nonexistent/")
	for i := 0; i < 10; i++ {
		_, err := customAuthClient.IntrospectToken(ctx, "test-token")
		t.Logf("Error on request %d to bad endpoint: %v", i, err)
		// After enough failures, we should see circuit breaker errors
		if i >= 5 && !IsCircuitBreakerOpenError(err) {
			t.Errorf("Expected circuit breaker to be open after %d failures", i)
		}
	}

	// Wait for circuit breaker to reset
	t.Logf("Waiting for circuit breaker to reset...")
	time.Sleep(10 * time.Second)

	// Try a valid request again
	_, err := authClient.IntrospectToken(ctx, "test-token")
	if IsCircuitBreakerOpenError(err) {
		t.Error("Circuit breaker should have reset but is still open")
	}
}

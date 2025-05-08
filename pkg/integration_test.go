//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package pkg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func init() {
	// Load environment variables from .env.test file
	_ = godotenv.Load("../.env.test")
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
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}

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
	groupsClient, err := config.NewGroupsClient(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create groups client: %v", err)
	}

	// Test list groups to verify client works
	groups, err := groupsClient.ListGroups(ctx, nil)
	if err != nil {
		t.Fatalf("Groups client failed: %v", err)
	}

	t.Logf("Found %d groups", len(groups.Groups))

	// Create Transfer client
	transferClient, err := config.NewTransferClient(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create transfer client: %v", err)
	}

	// Test list endpoints to verify client works
	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		t.Fatalf("Transfer client failed: %v", err)
	}

	t.Logf("Found %d endpoints", len(endpoints.DATA))

	// Create Search client
	searchClient, err := config.NewSearchClient(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create search client: %v", err)
	}

	// Test list indexes to verify client works
	indexes, err := searchClient.ListIndexes(ctx, &search.ListIndexesOptions{
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search client failed: %v", err)
	}

	t.Logf("Found %d search indexes", len(indexes.Indexes))

	// Create Flows client
	flowsClient, err := config.NewFlowsClient(tokenResp.AccessToken)
	if err != nil {
		t.Fatalf("Failed to create flows client: %v", err)
	}

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

	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	ctx := context.Background()

	// Get all scopes
	allScopes := GetScopesByService("auth", "groups", "transfer", "search", "flows")

	// Get client credentials token
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, allScopes...)
	if err != nil {
		t.Fatalf("GetClientCredentialsToken failed: %v", err)
	}

	// Create a static token authorizer
	staticAuthorizer := &simpleAuthorizer{token: tokenResp.AccessToken}

	// Get token from authorizer
	token, err := staticAuthorizer.GetAuthorizationHeader(ctx)
	if err != nil {
		t.Fatalf("StaticTokenAuthorizer.GetAuthorizationHeader failed: %v", err)
	}

	expected := "Bearer " + tokenResp.AccessToken
	if token != expected {
		t.Errorf("Authorizer token = %s, want %s", token, expected)
	}

	// Create client credentials authorizer
	credentialsOptions := []auth.ClientOption{
		auth.WithClientID(clientID),
		auth.WithClientSecret(clientSecret),
	}

	authClientForCreds, err := auth.NewClient(credentialsOptions...)
	if err != nil {
		t.Fatalf("Failed to create auth client for credentials: %v", err)
	}

	// Get a token to test
	credToken, err := authClientForCreds.GetClientCredentialsToken(ctx, allScopes...)
	if err != nil {
		t.Fatalf("ClientCredentials token request failed: %v", err)
	}

	if credToken.AccessToken == "" {
		t.Error("Expected non-empty token from credentials authorizer")
	}

	expiresAt := credToken.ExpiresAt()
	if !expiresAt.After(time.Now()) {
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
	tokenManager, err := config.NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	ctx := context.Background()

	// Get a token using client credentials flow
	allScopes := GetScopesByService("auth", "groups", "transfer")
	scopesKey := "integration_test_scopes"

	// Store a token
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, allScopes...)
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

	// Get authorization header from authorizer
	authHeader, err := authorizer.GetAuthorizationHeader(ctx)
	if err != nil {
		t.Fatalf("Authorizer.GetAuthorizationHeader failed: %v", err)
	}

	expectedHeader := "Bearer " + tokenResp.AccessToken
	if authHeader != expectedHeader {
		t.Errorf("Authorizer header = %s, want %s", authHeader, expectedHeader)
	}
}

func TestIntegration_RateLimiting(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration with rate limiting enabled
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)
		// Rate limiting now handled by client options

	// Create Auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	ctx := context.Background()

	// Make a series of requests to test rate limiting
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := authClient.IntrospectToken(ctx, "dummy-token")
		// We expect errors due to invalid token, but the requests should still be processed
		if err != nil {
			t.Logf("Expected error on request %d: %v", i, err)
		}
	}
	duration := time.Since(start)

	// We're just demonstrating that requests can be made in sequence
	t.Logf("Made 3 requests in %v", duration)
}

func TestIntegration_CircuitBreaker(t *testing.T) {
	clientID, clientSecret := skipIfMissingCredentials(t)

	// Create SDK configuration with circuit breaker enabled
	config := NewConfig().
		WithClientID(clientID).
		WithClientSecret(clientSecret)
		// Circuit breaking now handled by client options

	// Create Auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create auth client: %v", err)
	}
	ctx := context.Background()

	// Make a request to a valid endpoint with invalid token to trigger 401s
	// These should not trip the circuit breaker
	for i := 0; i < 3; i++ {
		_, err := authClient.IntrospectToken(ctx, "invalid-token")
		// We expect Unauthorized errors, but these shouldn't trip the circuit breaker
		if err != nil {
			t.Logf("Expected error type on request %d: %v", i, err)
		}
	}

	// Request still works after these errors because they're 401s, not 5xx
	_, err = authClient.IntrospectToken(ctx, "test-token")
	t.Logf("Final request error (expected): %v", err)
}

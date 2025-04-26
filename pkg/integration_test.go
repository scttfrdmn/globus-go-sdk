// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package pkg

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
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, AuthScope)
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
}

func TestIntegration_GetScopesByService(t *testing.T) {
	// Test getting scopes for individual services
	authScopes := GetScopesByService("auth")
	if len(authScopes) != 1 || authScopes[0] != AuthScope {
		t.Errorf("Auth scope = %v, want %v", authScopes, []string{AuthScope})
	}
	
	groupsScopes := GetScopesByService("groups")
	if len(groupsScopes) != 1 || groupsScopes[0] != GroupsScope {
		t.Errorf("Groups scope = %v, want %v", groupsScopes, []string{GroupsScope})
	}
	
	transferScopes := GetScopesByService("transfer")
	if len(transferScopes) != 1 || transferScopes[0] != TransferScope {
		t.Errorf("Transfer scope = %v, want %v", transferScopes, []string{TransferScope})
	}
	
	// Test getting scopes for multiple services
	allScopes := GetScopesByService("auth", "groups", "transfer")
	expectedScopes := []string{AuthScope, GroupsScope, TransferScope}
	
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
	allScopes := GetScopesByService("auth", "groups", "transfer")
	
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
}
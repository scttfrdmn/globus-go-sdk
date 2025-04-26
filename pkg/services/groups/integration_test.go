// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

//go:build integration
// +build integration

package groups

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/services/auth"
)

func getTestCredentials(t *testing.T) (string, string, string) {
	clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
	groupID := os.Getenv("GLOBUS_TEST_GROUP_ID")
	
	if clientID == "" || clientSecret == "" {
		t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
	}
	
	return clientID, clientSecret, groupID
}

func getAccessToken(t *testing.T, clientID, clientSecret string) string {
	authClient := auth.NewClient(clientID, clientSecret)
	
	tokenResp, err := authClient.GetClientCredentialsToken(context.Background(), GroupsScope)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}
	
	return tokenResp.AccessToken
}

func TestIntegration_ListGroups(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Groups client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// List groups
	options := &ListGroupsOptions{
		PageSize: 5,
	}
	
	groups, err := client.ListGroups(ctx, options)
	if err != nil {
		t.Fatalf("ListGroups failed: %v", err)
	}
	
	// Verify we got some data
	t.Logf("Found %d groups", len(groups.Groups))
	
	// The user might not be a member of any groups, so this isn't necessarily an error
	if len(groups.Groups) > 0 {
		// Check that the first group has expected fields
		firstGroup := groups.Groups[0]
		if firstGroup.ID == "" {
			t.Error("First group is missing ID")
		}
		if firstGroup.Name == "" {
			t.Error("First group is missing name")
		}
	}
}

func TestIntegration_GroupLifecycle(t *testing.T) {
	clientID, clientSecret, _ := getTestCredentials(t)
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Groups client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// 1. Create a new group
	timestamp := time.Now().Format("20060102_150405")
	groupName := fmt.Sprintf("Test Group %s", timestamp)
	groupDescription := "A test group created by integration tests"
	
	createRequest := &GroupCreate{
		Name:        groupName,
		Description: groupDescription,
		Visibility:  "private",
	}
	
	createdGroup, err := client.CreateGroup(ctx, createRequest)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	
	// Make sure the group gets deleted after the test
	defer func() {
		err := client.DeleteGroup(ctx, createdGroup.ID)
		if err != nil {
			t.Logf("Warning: Failed to delete test group (%s): %v", createdGroup.ID, err)
		} else {
			t.Logf("Successfully deleted test group (%s)", createdGroup.ID)
		}
	}()
	
	t.Logf("Created group: %s (%s)", createdGroup.Name, createdGroup.ID)
	
	// 2. Verify the group was created correctly
	if createdGroup.Name != groupName {
		t.Errorf("Created group name = %s, want %s", createdGroup.Name, groupName)
	}
	if createdGroup.Description != groupDescription {
		t.Errorf("Created group description = %s, want %s", createdGroup.Description, groupDescription)
	}
	
	// 3. Get the group
	fetchedGroup, err := client.GetGroup(ctx, createdGroup.ID)
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}
	
	if fetchedGroup.ID != createdGroup.ID {
		t.Errorf("Fetched group ID = %s, want %s", fetchedGroup.ID, createdGroup.ID)
	}
	
	// 4. Update the group
	updatedDescription := "Updated description for integration test"
	updateRequest := &GroupUpdate{
		Description: updatedDescription,
	}
	
	updatedGroup, err := client.UpdateGroup(ctx, createdGroup.ID, updateRequest)
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}
	
	if updatedGroup.Description != updatedDescription {
		t.Errorf("Updated group description = %s, want %s", updatedGroup.Description, updatedDescription)
	}
	
	// 5. List roles for the group
	roles, err := client.ListRoles(ctx, createdGroup.ID)
	if err != nil {
		t.Fatalf("Failed to list roles: %v", err)
	}
	
	t.Logf("Group has %d roles", len(roles.Roles))
	
	// Default groups should have admin and member roles
	if len(roles.Roles) < 2 {
		t.Errorf("Expected at least 2 roles, got %d", len(roles.Roles))
	}
	
	// Find the admin role
	var adminRoleID string
	for _, role := range roles.Roles {
		if role.Name == "admin" || role.Name == "administrator" {
			adminRoleID = role.ID
			break
		}
	}
	
	if adminRoleID == "" {
		t.Log("Could not find admin role, skipping role tests")
	} else {
		// 6. Get a specific role
		role, err := client.GetRole(ctx, createdGroup.ID, adminRoleID)
		if err != nil {
			t.Fatalf("Failed to get role: %v", err)
		}
		
		if role.ID != adminRoleID {
			t.Errorf("Got role ID = %s, want %s", role.ID, adminRoleID)
		}
	}
}

func TestIntegration_ExistingGroup(t *testing.T) {
	clientID, clientSecret, groupID := getTestCredentials(t)
	
	// Skip if no existing group ID is provided
	if groupID == "" {
		t.Skip("Integration test requires GLOBUS_TEST_GROUP_ID for existing group operations")
	}
	
	// Get access token
	accessToken := getAccessToken(t, clientID, clientSecret)
	
	// Create Groups client
	client := NewClient(accessToken)
	ctx := context.Background()
	
	// Verify we can get the group
	group, err := client.GetGroup(ctx, groupID)
	if err != nil {
		t.Fatalf("Failed to get group: %v", err)
	}
	
	t.Logf("Found group: %s (%s)", group.Name, group.ID)
	
	// List members
	members, err := client.ListMembers(ctx, groupID, nil)
	if err != nil {
		t.Fatalf("Failed to list members: %v", err)
	}
	
	t.Logf("Group has %d members", len(members.Members))
	
	// Check if we have members
	if len(members.Members) > 0 {
		// Check that the first member has expected fields
		firstMember := members.Members[0]
		if firstMember.ID == "" {
			t.Error("First member is missing ID")
		}
		if firstMember.Username == "" && firstMember.Email == "" {
			t.Error("First member is missing both username and email")
		}
	}
}
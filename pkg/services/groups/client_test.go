// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

// Test helper to set up a mock server and client
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)

	// Create an authorizer
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")

	// Create a client that uses the test server
	client, _ := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)

	return server, client
}

func TestBuildURLLowLevel(t *testing.T) {
	// Create an authorizer
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")

	// Create a client that uses the base URL
	client, _ := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOptions(core.WithBaseURL("https://example.com")),
	)

	// Test with no query parameters
	url := client.buildURLLowLevel("test/path", nil)
	if url != "https://example.com/test/path" {
		t.Errorf("buildURL() = %v, want %v", url, "https://example.com/test/path")
	}

	// Test with query parameters
	query := map[string][]string{
		"param1": {"value1"},
		"param2": {"value2"},
	}
	url = client.buildURLLowLevel("test/path", query)
	if url != "https://example.com/test/path?param1=value1&param2=value2" {
		t.Errorf("buildURL() with query = %v, want %v", url, "https://example.com/test/path?param1=value1&param2=value2")
	}

	// Test with trailing slash in base URL
	client, _ = NewClient(
		WithAuthorizer(authorizer),
		WithCoreOptions(core.WithBaseURL("https://example.com/")),
	)
	url = client.buildURLLowLevel("test/path", nil)
	if url != "https://example.com/test/path" {
		t.Errorf("buildURL() with trailing slash = %v, want %v", url, "https://example.com/test/path")
	}
}

func TestListGroups(t *testing.T) {
	// Setup test server
	now := time.Now()
	groups := []Group{
		{
			ID:          "group1",
			Name:        "Group 1",
			Description: "First group",
			Created:     now,
			LastUpdated: now,
		},
		{
			ID:          "group2",
			Name:        "Group 2",
			Description: "Second group",
			Created:     now,
			LastUpdated: now,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups" {
			t.Errorf("Expected path /groups, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("my_groups") != "true" {
			t.Errorf("Expected my_groups=true, got %s", query.Get("my_groups"))
		}
		if query.Get("include_group_membership") != "true" {
			t.Errorf("Expected include_group_membership=true, got %s", query.Get("include_group_membership"))
		}
		if query.Get("per_page") != "10" {
			t.Errorf("Expected per_page=10, got %s", query.Get("per_page"))
		}

		// Return mock response
		response := GroupList{
			Groups:        groups,
			HasNextPage:   true,
			NextPageToken: "next-page-token",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list groups
	options := &ListGroupsOptions{
		MyGroups:               true,
		IncludeGroupMembership: true,
		PageSize:               10,
	}

	groupList, err := client.ListGroups(context.Background(), options)
	if err != nil {
		t.Fatalf("ListGroups() error = %v", err)
	}

	// Check response
	if len(groupList.Groups) != 2 {
		t.Fatalf("ListGroups() returned %d groups, want 2", len(groupList.Groups))
	}
	if groupList.Groups[0].ID != "group1" {
		t.Errorf("ListGroups() group[0].ID = %v, want %v", groupList.Groups[0].ID, "group1")
	}
	if groupList.Groups[1].ID != "group2" {
		t.Errorf("ListGroups() group[1].ID = %v, want %v", groupList.Groups[1].ID, "group2")
	}
	if !groupList.HasNextPage {
		t.Errorf("ListGroups() HasNextPage = %v, want %v", groupList.HasNextPage, true)
	}
	if groupList.NextPageToken != "next-page-token" {
		t.Errorf("ListGroups() NextPageToken = %v, want %v", groupList.NextPageToken, "next-page-token")
	}

	// Test with nil options
	handler = func(w http.ResponseWriter, r *http.Request) {
		response := GroupList{
			Groups:      groups,
			HasNextPage: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	groupList, err = client.ListGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGroups() with nil options error = %v", err)
	}

	if len(groupList.Groups) != 2 {
		t.Fatalf("ListGroups() with nil options returned %d groups, want 2", len(groupList.Groups))
	}
}

func TestGetGroup(t *testing.T) {
	// Setup test server
	now := time.Now()
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1" {
			t.Errorf("Expected path /groups/group1, got %s", r.URL.Path)
		}

		// Return mock response
		group := Group{
			ID:          "group1",
			Name:        "Group 1",
			Description: "Test group",
			Created:     now,
			LastUpdated: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(group)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get group
	group, err := client.GetGroup(context.Background(), "group1")
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}

	// Check response
	if group.ID != "group1" {
		t.Errorf("GetGroup() ID = %v, want %v", group.ID, "group1")
	}
	if group.Name != "Group 1" {
		t.Errorf("GetGroup() Name = %v, want %v", group.Name, "Group 1")
	}
	if group.Description != "Test group" {
		t.Errorf("GetGroup() Description = %v, want %v", group.Description, "Test group")
	}

	// Test with empty ID
	_, err = client.GetGroup(context.Background(), "")
	if err == nil {
		t.Error("GetGroup() with empty ID should return error")
	}
}

func TestCreateGroup(t *testing.T) {
	// Setup test server
	now := time.Now()
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups" {
			t.Errorf("Expected path /groups, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody GroupCreate
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody.Name != "New Group" {
			t.Errorf("Expected Name=New Group, got %s", requestBody.Name)
		}
		if requestBody.Description != "A new group" {
			t.Errorf("Expected Description=A new group, got %s", requestBody.Description)
		}

		// Return mock response
		group := Group{
			ID:          "new-group-id",
			Name:        requestBody.Name,
			Description: requestBody.Description,
			Created:     now,
			LastUpdated: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(group)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test create group
	createRequest := &GroupCreate{
		Name:        "New Group",
		Description: "A new group",
		PublicGroup: true,
	}

	group, err := client.CreateGroup(context.Background(), createRequest)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}

	// Check response
	if group.ID != "new-group-id" {
		t.Errorf("CreateGroup() ID = %v, want %v", group.ID, "new-group-id")
	}
	if group.Name != "New Group" {
		t.Errorf("CreateGroup() Name = %v, want %v", group.Name, "New Group")
	}
	if group.Description != "A new group" {
		t.Errorf("CreateGroup() Description = %v, want %v", group.Description, "A new group")
	}

	// Test with nil request
	_, err = client.CreateGroup(context.Background(), nil)
	if err == nil {
		t.Error("CreateGroup() with nil request should return error")
	}

	// Test with empty name
	_, err = client.CreateGroup(context.Background(), &GroupCreate{})
	if err == nil {
		t.Error("CreateGroup() with empty name should return error")
	}
}

func TestUpdateGroup(t *testing.T) {
	// Setup test server
	now := time.Now()
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1" {
			t.Errorf("Expected path /groups/group1, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody GroupUpdate
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody.Name != "Updated Group" {
			t.Errorf("Expected Name=Updated Group, got %s", requestBody.Name)
		}
		if requestBody.Description != "An updated group" {
			t.Errorf("Expected Description=An updated group, got %s", requestBody.Description)
		}

		// Return mock response
		group := Group{
			ID:          "group1",
			Name:        requestBody.Name,
			Description: requestBody.Description,
			Created:     now,
			LastUpdated: now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(group)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test update group
	updateRequest := &GroupUpdate{
		Name:        "Updated Group",
		Description: "An updated group",
	}

	group, err := client.UpdateGroup(context.Background(), "group1", updateRequest)
	if err != nil {
		t.Fatalf("UpdateGroup() error = %v", err)
	}

	// Check response
	if group.ID != "group1" {
		t.Errorf("UpdateGroup() ID = %v, want %v", group.ID, "group1")
	}
	if group.Name != "Updated Group" {
		t.Errorf("UpdateGroup() Name = %v, want %v", group.Name, "Updated Group")
	}
	if group.Description != "An updated group" {
		t.Errorf("UpdateGroup() Description = %v, want %v", group.Description, "An updated group")
	}

	// Test with empty ID
	_, err = client.UpdateGroup(context.Background(), "", updateRequest)
	if err == nil {
		t.Error("UpdateGroup() with empty ID should return error")
	}

	// Test with nil request
	_, err = client.UpdateGroup(context.Background(), "group1", nil)
	if err == nil {
		t.Error("UpdateGroup() with nil request should return error")
	}
}

func TestDeleteGroup(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1" {
			t.Errorf("Expected path /groups/group1, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test delete group
	err := client.DeleteGroup(context.Background(), "group1")
	if err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}

	// Test with empty ID
	err = client.DeleteGroup(context.Background(), "")
	if err == nil {
		t.Error("DeleteGroup() with empty ID should return error")
	}
}

func TestListMembers(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/members" {
			t.Errorf("Expected path /groups/group1/members, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("role_id") != "admin" {
			t.Errorf("Expected role_id=admin, got %s", query.Get("role_id"))
		}
		if query.Get("per_page") != "10" {
			t.Errorf("Expected per_page=10, got %s", query.Get("per_page"))
		}

		// Return mock response
		members := []Member{
			{
				IdentityID: "member1",
				Username:   "user1",
				Email:      "user1@example.com",
				Status:     "active",
				RoleID:     "admin",
				Role: Role{
					ID:   "admin",
					Name: "Administrator",
				},
			},
			{
				IdentityID: "member2",
				Username:   "user2",
				Email:      "user2@example.com",
				Status:     "active",
				RoleID:     "admin",
				Role: Role{
					ID:   "admin",
					Name: "Administrator",
				},
			},
		}

		response := MemberList{
			Members:       members,
			HasNextPage:   true,
			NextPageToken: "next-page-token",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test list members
	options := &ListMembersOptions{
		RoleID:   "admin",
		PageSize: 10,
	}

	memberList, err := client.ListMembers(context.Background(), "group1", options)
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}

	// Check response
	if len(memberList.Members) != 2 {
		t.Fatalf("ListMembers() returned %d members, want 2", len(memberList.Members))
	}
	if memberList.Members[0].IdentityID != "member1" {
		t.Errorf("ListMembers() member[0].IdentityID = %v, want %v", memberList.Members[0].IdentityID, "member1")
	}
	if memberList.Members[1].IdentityID != "member2" {
		t.Errorf("ListMembers() member[1].IdentityID = %v, want %v", memberList.Members[1].IdentityID, "member2")
	}
	if !memberList.HasNextPage {
		t.Errorf("ListMembers() HasNextPage = %v, want %v", memberList.HasNextPage, true)
	}
	if memberList.NextPageToken != "next-page-token" {
		t.Errorf("ListMembers() NextPageToken = %v, want %v", memberList.NextPageToken, "next-page-token")
	}

	// Test with empty group ID
	_, err = client.ListMembers(context.Background(), "", options)
	if err == nil {
		t.Error("ListMembers() with empty group ID should return error")
	}
}

func TestAddMember(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/members" {
			t.Errorf("Expected path /groups/group1/members, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody["identity_id"] != "user1" {
			t.Errorf("Expected identity_id=user1, got %s", requestBody["identity_id"])
		}
		if requestBody["role_id"] != "member" {
			t.Errorf("Expected role_id=member, got %s", requestBody["role_id"])
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test add member
	err := client.AddMember(context.Background(), "group1", "user1", "member")
	if err != nil {
		t.Fatalf("AddMember() error = %v", err)
	}

	// Test with empty group ID
	err = client.AddMember(context.Background(), "", "user1", "member")
	if err == nil {
		t.Error("AddMember() with empty group ID should return error")
	}

	// Test with empty user ID
	err = client.AddMember(context.Background(), "group1", "", "member")
	if err == nil {
		t.Error("AddMember() with empty user ID should return error")
	}

	// Test with empty role ID
	err = client.AddMember(context.Background(), "group1", "user1", "")
	if err == nil {
		t.Error("AddMember() with empty role ID should return error")
	}
}

func TestRemoveMember(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/members/user1" {
			t.Errorf("Expected path /groups/group1/members/user1, got %s", r.URL.Path)
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test remove member
	err := client.RemoveMember(context.Background(), "group1", "user1")
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}

	// Test with empty group ID
	err = client.RemoveMember(context.Background(), "", "user1")
	if err == nil {
		t.Error("RemoveMember() with empty group ID should return error")
	}

	// Test with empty user ID
	err = client.RemoveMember(context.Background(), "group1", "")
	if err == nil {
		t.Error("RemoveMember() with empty user ID should return error")
	}
}

func TestUpdateMemberRole(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/members/user1" {
			t.Errorf("Expected path /groups/group1/members/user1, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody["role_id"] != "admin" {
			t.Errorf("Expected role_id=admin, got %s", requestBody["role_id"])
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test update member role
	err := client.UpdateMemberRole(context.Background(), "group1", "user1", "admin")
	if err != nil {
		t.Fatalf("UpdateMemberRole() error = %v", err)
	}

	// Test with empty group ID
	err = client.UpdateMemberRole(context.Background(), "", "user1", "admin")
	if err == nil {
		t.Error("UpdateMemberRole() with empty group ID should return error")
	}

	// Test with empty user ID
	err = client.UpdateMemberRole(context.Background(), "group1", "", "admin")
	if err == nil {
		t.Error("UpdateMemberRole() with empty user ID should return error")
	}

	// Test with empty role ID
	err = client.UpdateMemberRole(context.Background(), "group1", "user1", "")
	if err == nil {
		t.Error("UpdateMemberRole() with empty role ID should return error")
	}
}

// Tests for v3.62.0 subscription functionality

func TestSetSubscriptionAdminVerified(t *testing.T) {
	// Setup test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/subscription" {
			t.Errorf("Expected path /groups/group1/subscription, got %s", r.URL.Path)
		}

		// Check request body
		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if requestBody["subscription_id"] != "sub-12345" {
			t.Errorf("Expected subscription_id=sub-12345, got %s", requestBody["subscription_id"])
		}
		if requestBody["DATA_TYPE"] != "subscription_update" {
			t.Errorf("Expected DATA_TYPE=subscription_update, got %s", requestBody["DATA_TYPE"])
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test set subscription admin verified ID
	err := client.SetSubscriptionAdminVerified(context.Background(), "group1", "sub-12345")
	if err != nil {
		t.Fatalf("SetSubscriptionAdminVerified() error = %v", err)
	}

	// Test with empty group ID
	err = client.SetSubscriptionAdminVerified(context.Background(), "", "sub-12345")
	if err == nil {
		t.Error("SetSubscriptionAdminVerified() with empty group ID should return error")
	}

	// Test with empty subscription ID
	err = client.SetSubscriptionAdminVerified(context.Background(), "group1", "")
	if err == nil {
		t.Error("SetSubscriptionAdminVerified() with empty subscription ID should return error")
	}
}

func TestGetGroupSubscription(t *testing.T) {
	// Setup test server
	now := time.Now()
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check path
		if r.URL.Path != "/groups/group1/subscription" {
			t.Errorf("Expected path /groups/group1/subscription, got %s", r.URL.Path)
		}

		// Return mock response
		subscription := GroupSubscription{
			DATA_TYPE:        "group_subscription",
			SubscriptionID:   "sub-12345",
			GroupID:          "group1",
			IsActive:         true,
			Created:          now,
			LastUpdated:      now,
			SubscriptionType: "premium",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(subscription)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test get group subscription
	subscription, err := client.GetGroupSubscription(context.Background(), "group1")
	if err != nil {
		t.Fatalf("GetGroupSubscription() error = %v", err)
	}

	// Check response
	if subscription.SubscriptionID != "sub-12345" {
		t.Errorf("GetGroupSubscription() SubscriptionID = %v, want %v", subscription.SubscriptionID, "sub-12345")
	}
	if subscription.GroupID != "group1" {
		t.Errorf("GetGroupSubscription() GroupID = %v, want %v", subscription.GroupID, "group1")
	}
	if !subscription.IsActive {
		t.Errorf("GetGroupSubscription() IsActive = %v, want %v", subscription.IsActive, true)
	}
	if subscription.SubscriptionType != "premium" {
		t.Errorf("GetGroupSubscription() SubscriptionType = %v, want %v", subscription.SubscriptionType, "premium")
	}
	if subscription.DATA_TYPE != "group_subscription" {
		t.Errorf("GetGroupSubscription() DATA_TYPE = %v, want %v", subscription.DATA_TYPE, "group_subscription")
	}

	// Test with empty group ID
	_, err = client.GetGroupSubscription(context.Background(), "")
	if err == nil {
		t.Error("GetGroupSubscription() with empty group ID should return error")
	}

	// Test DATA_TYPE is set when missing from response
	handler = func(w http.ResponseWriter, r *http.Request) {
		// Return response without DATA_TYPE
		subscription := GroupSubscription{
			SubscriptionID:   "sub-12345",
			GroupID:          "group1",
			IsActive:         true,
			SubscriptionType: "premium",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(subscription)
	}

	server, client = setupMockServer(handler)
	defer server.Close()

	subscription, err = client.GetGroupSubscription(context.Background(), "group1")
	if err != nil {
		t.Fatalf("GetGroupSubscription() with missing DATA_TYPE error = %v", err)
	}

	// Check that DATA_TYPE was set by the client
	if subscription.DATA_TYPE != "group_subscription" {
		t.Errorf("GetGroupSubscription() should set DATA_TYPE to 'group_subscription', got %v", subscription.DATA_TYPE)
	}
}

// TestListGroupsWithStatuses tests the statuses parameter (added in v3.65.0)
func TestListGroupsWithStatuses(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify statuses query parameter
		statuses := r.URL.Query()["statuses"]
		if len(statuses) != 2 {
			t.Errorf("Expected 2 status values, got %d", len(statuses))
		}
		if statuses[0] != "active" || statuses[1] != "pending" {
			t.Errorf("Expected statuses ['active', 'pending'], got %v", statuses)
		}

		response := GroupList{
			Groups: []Group{
				{
					ID:   "group1",
					Name: "Test Group 1",
				},
				{
					ID:   "group2",
					Name: "Test Group 2",
				},
			},
			HasNextPage:   false,
			NextPageToken: "",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	// Test with multiple statuses
	options := &ListGroupsOptions{
		Statuses: []string{"active", "pending"},
	}

	groupList, err := client.ListGroups(context.Background(), options)
	if err != nil {
		t.Fatalf("ListGroups() with statuses error = %v", err)
	}

	if len(groupList.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groupList.Groups))
	}
}

// TestListGroupsWithSingleStatus tests a single status filter
func TestListGroupsWithSingleStatus(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		statuses := r.URL.Query()["statuses"]
		if len(statuses) != 1 {
			t.Errorf("Expected 1 status value, got %d", len(statuses))
		}
		if statuses[0] != "active" {
			t.Errorf("Expected status 'active', got %v", statuses[0])
		}

		response := GroupList{
			Groups: []Group{
				{
					ID:   "group1",
					Name: "Active Group",
				},
			},
			HasNextPage:   false,
			NextPageToken: "",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	options := &ListGroupsOptions{
		Statuses: []string{"active"},
	}

	groupList, err := client.ListGroups(context.Background(), options)
	if err != nil {
		t.Fatalf("ListGroups() with single status error = %v", err)
	}

	if len(groupList.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groupList.Groups))
	}
}

// TestListGroupsWithStatusesAndOtherFilters tests combining statuses with other filters
func TestListGroupsWithStatusesAndOtherFilters(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Verify both statuses and my_groups parameters
		statuses := r.URL.Query()["statuses"]
		myGroups := r.URL.Query().Get("my_groups")

		if len(statuses) == 0 {
			t.Error("Expected statuses parameter")
		}
		if myGroups != "true" {
			t.Errorf("Expected my_groups=true, got %s", myGroups)
		}

		response := GroupList{
			Groups: []Group{
				{
					ID:   "group1",
					Name: "My Active Group",
				},
			},
			HasNextPage:   false,
			NextPageToken: "",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	server, client := setupMockServer(handler)
	defer server.Close()

	options := &ListGroupsOptions{
		MyGroups: true,
		Statuses: []string{"active"},
	}

	groupList, err := client.ListGroups(context.Background(), options)
	if err != nil {
		t.Fatalf("ListGroups() with combined filters error = %v", err)
	}

	if len(groupList.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groupList.Groups))
	}
}

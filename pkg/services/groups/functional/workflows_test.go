// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package groups_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/testhelpers"
)

// TestCompleteGroupManagementWorkflow tests a complete user workflow following Python SDK patterns
func TestCompleteGroupManagementWorkflow(t *testing.T) {
	// Load workflow test scenario
	scenario := testhelpers.LoadTestScenario(t, "groups", "group_create_workflow")
	if scenario == nil {
		// Fallback if JSON not available
		scenario = &testhelpers.TestScenario{
			GroupID: "workflow-group-123",
			UserID:  "test-user-456",
			RoleID:  "admin",
			Subscription: map[string]interface{}{
				"subscription_id": "sub-workflow-789",
			},
			Expected: map[string]interface{}{
				"name":        "Test Workflow Group",
				"description": "A group for testing workflows",
			},
		}
	}

	now := time.Now()
	responseHandler := testhelpers.NewMockResponseHandler()

	// Step 1: Create Group Response
	responseHandler.RegisterResponse(http.MethodPost, "/groups", testhelpers.MockResponse{
		StatusCode: 201,
		Body: groups.Group{
			DATA_TYPE:   "group",
			ID:          scenario.GroupID,
			Name:        scenario.Expected["name"].(string),
			Description: scenario.Expected["description"].(string),
			Created:     now,
			LastUpdated: now,
		},
		Headers: map[string]string{
			"Location": "/groups/" + scenario.GroupID,
		},
	})

	// Step 2: Add Member Response
	responseHandler.RegisterResponse(http.MethodPost, "/groups/"+scenario.GroupID+"/members", testhelpers.MockResponse{
		StatusCode: 200,
		Body:       map[string]string{"status": "success"},
	})

	// Step 3: Set Subscription Response
	responseHandler.RegisterResponse(http.MethodPut, "/groups/"+scenario.GroupID+"/subscription", testhelpers.MockResponse{
		StatusCode: 200,
		Body:       map[string]string{"status": "success"},
	})

	// Step 4: Get Group Response (verification)
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+scenario.GroupID, testhelpers.MockResponse{
		StatusCode: 200,
		Body: groups.Group{
			DATA_TYPE:    "group",
			ID:           scenario.GroupID,
			Name:         scenario.Expected["name"].(string),
			Description:  scenario.Expected["description"].(string),
			MemberCount:  1,
			IsGroupAdmin: true,
			IsMember:     true,
			Created:      now,
			LastUpdated:  now,
		},
	})

	// Step 5: Delete Group Response (cleanup)
	responseHandler.RegisterResponse(http.MethodDelete, "/groups/"+scenario.GroupID, testhelpers.MockResponse{
		StatusCode: 200,
		Body:       map[string]string{"status": "deleted"},
	})

	// Create client with mock server
	server := httptest.NewServer(responseHandler)
	defer server.Close()

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-workflow-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// WORKFLOW EXECUTION
	t.Log("🚀 Starting complete group management workflow")

	// Step 1: Create Group
	t.Log("📝 Step 1: Creating group")
	createRequest := &groups.GroupCreate{
		Name:        scenario.Expected["name"].(string),
		Description: scenario.Expected["description"].(string),
		PublicGroup: false,
	}

	createdGroup, err := client.CreateGroup(ctx, createRequest)
	if err != nil {
		t.Fatalf("Step 1 failed - CreateGroup(): %v", err)
	}

	if createdGroup.ID != scenario.GroupID {
		t.Errorf("Step 1 validation failed - Expected group ID %s, got %s", scenario.GroupID, createdGroup.ID)
	}
	t.Logf("✅ Step 1 complete: Group created with ID %s", createdGroup.ID)

	// Step 2: Add Member
	t.Log("👥 Step 2: Adding member to group")
	err = client.AddMember(ctx, createdGroup.ID, scenario.UserID, scenario.RoleID)
	if err != nil {
		t.Fatalf("Step 2 failed - AddMember(): %v", err)
	}
	t.Logf("✅ Step 2 complete: Member %s added with role %s", scenario.UserID, scenario.RoleID)

	// Step 3: Set Subscription
	t.Log("💳 Step 3: Setting group subscription")
	subscriptionID := scenario.Subscription["subscription_id"].(string)
	err = client.SetSubscriptionAdminVerified(ctx, createdGroup.ID, subscriptionID)
	if err != nil {
		t.Fatalf("Step 3 failed - SetSubscriptionAdminVerified(): %v", err)
	}
	t.Logf("✅ Step 3 complete: Subscription %s set for group", subscriptionID)

	// Step 4: Verify Group State
	t.Log("🔍 Step 4: Verifying group state")
	verifiedGroup, err := client.GetGroup(ctx, createdGroup.ID)
	if err != nil {
		t.Fatalf("Step 4 failed - GetGroup(): %v", err)
	}

	if verifiedGroup.Name != scenario.Expected["name"].(string) {
		t.Errorf("Step 4 validation failed - Expected name %s, got %s",
			scenario.Expected["name"].(string), verifiedGroup.Name)
	}
	if verifiedGroup.MemberCount != 1 {
		t.Errorf("Step 4 validation failed - Expected member count 1, got %d", verifiedGroup.MemberCount)
	}
	t.Log("✅ Step 4 complete: Group state verified")

	// Step 5: Cleanup - Delete Group
	t.Log("🗑️ Step 5: Cleaning up - deleting group")
	err = client.DeleteGroup(ctx, createdGroup.ID)
	if err != nil {
		t.Fatalf("Step 5 failed - DeleteGroup(): %v", err)
	}
	t.Log("✅ Step 5 complete: Group deleted")

	t.Log("🎉 Workflow completed successfully!")
}

// TestSubscriptionManagementWorkflow tests subscription-specific workflows
func TestSubscriptionManagementWorkflow(t *testing.T) {
	groupID := "sub-test-group-456"
	subscriptionID := "sub-premium-789"

	responseHandler := testhelpers.NewMockResponseHandler()

	// Step 1: Set subscription
	responseHandler.RegisterResponse(http.MethodPut, "/groups/"+groupID+"/subscription", testhelpers.MockResponse{
		StatusCode: 200,
		Body: map[string]string{
			"status":          "success",
			"group_id":        groupID,
			"subscription_id": subscriptionID,
		},
	})

	// Step 2: Get subscription details
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+groupID+"/subscription", testhelpers.MockResponse{
		StatusCode: 200,
		Body: groups.GroupSubscription{
			DATA_TYPE:        "group_subscription",
			SubscriptionID:   subscriptionID,
			GroupID:          groupID,
			IsActive:         true,
			SubscriptionType: "premium",
			Created:          time.Now(),
			LastUpdated:      time.Now(),
		},
	})

	// Step 3: Get group by subscription ID (reverse lookup)
	responseHandler.RegisterResponse(http.MethodGet, "/groups", testhelpers.MockResponse{
		StatusCode: 200,
		Body: groups.Group{
			DATA_TYPE:   "group",
			ID:          groupID,
			Name:        "Premium Subscription Group",
			Description: "A group with premium subscription",
		},
	})

	// Create client
	server := httptest.NewServer(responseHandler)
	defer server.Close()

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-subscription-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Log("🚀 Starting subscription management workflow")

	// Step 1: Set Subscription
	t.Log("💳 Step 1: Setting subscription for group")
	err = client.SetSubscriptionAdminVerified(ctx, groupID, subscriptionID)
	if err != nil {
		t.Fatalf("Step 1 failed - SetSubscriptionAdminVerified(): %v", err)
	}
	t.Log("✅ Step 1 complete: Subscription set")

	// Step 2: Verify Subscription Details
	t.Log("🔍 Step 2: Getting subscription details")
	subscription, err := client.GetGroupSubscription(ctx, groupID)
	if err != nil {
		t.Fatalf("Step 2 failed - GetGroupSubscription(): %v", err)
	}

	if subscription.SubscriptionID != subscriptionID {
		t.Errorf("Step 2 validation failed - Expected subscription ID %s, got %s",
			subscriptionID, subscription.SubscriptionID)
	}
	if !subscription.IsActive {
		t.Error("Step 2 validation failed - Expected subscription to be active")
	}
	t.Log("✅ Step 2 complete: Subscription details verified")

	// Step 3: Test Reverse Lookup (Python SDK two-step pattern)
	t.Log("🔄 Step 3: Testing reverse lookup (get group by subscription)")
	foundGroup, err := client.GetGroupBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		t.Fatalf("Step 3 failed - GetGroupBySubscriptionID(): %v", err)
	}

	if foundGroup.ID != groupID {
		t.Errorf("Step 3 validation failed - Expected group ID %s, got %s",
			groupID, foundGroup.ID)
	}
	t.Log("✅ Step 3 complete: Reverse lookup successful")

	t.Log("🎉 Subscription workflow completed successfully!")
}

// TestPolicyConfigurationWorkflow tests policy management workflows
func TestPolicyConfigurationWorkflow(t *testing.T) {
	groupID := "policy-test-group-789"

	responseHandler := testhelpers.NewMockResponseHandler()

	// Original policies
	originalPolicies := groups.GroupPolicies{
		DATA_TYPE: "group_policies",
		GroupID:   groupID,
		Policies: map[string]interface{}{
			"join_requests": false,
			"visibility":    "private",
		},
		JoinRequests: false,
	}

	// Updated policies
	updatedPolicies := groups.GroupPolicies{
		DATA_TYPE: "group_policies",
		GroupID:   groupID,
		Policies: map[string]interface{}{
			"join_requests": true,
			"visibility":    "public",
		},
		JoinRequests: true,
		SignupFields: []string{"name", "email", "institution"},
	}

	// Step 1: Get current policies
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+groupID+"/policies", testhelpers.MockResponse{
		StatusCode: 200,
		Body:       originalPolicies,
	})

	// Step 2: Update policies
	responseHandler.RegisterResponse(http.MethodPut, "/groups/"+groupID+"/policies", testhelpers.MockResponse{
		StatusCode: 200,
		Body: map[string]string{
			"status":   "updated",
			"group_id": groupID,
		},
	})

	// Note: We'll register the updated policies response after step 2

	// Create client
	server := httptest.NewServer(responseHandler)
	defer server.Close()

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-policy-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Log("🚀 Starting policy configuration workflow")

	// Step 1: Get Current Policies
	t.Log("📋 Step 1: Getting current policies")
	currentPolicies, err := client.GetGroupPolicies(ctx, groupID)
	if err != nil {
		t.Fatalf("Step 1 failed - GetGroupPolicies(): %v", err)
	}

	if currentPolicies.JoinRequests != false {
		t.Error("Step 1 validation failed - Expected join requests to be disabled initially")
	}
	t.Log("✅ Step 1 complete: Current policies retrieved")

	// Step 2: Update Policies
	t.Log("⚙️ Step 2: Updating policies")
	newPolicies := &groups.GroupPolicies{
		Policies: map[string]interface{}{
			"join_requests": true,
			"visibility":    "public",
		},
		JoinRequests: true,
		SignupFields: []string{"name", "email", "institution"},
	}

	err = client.SetGroupPolicies(ctx, groupID, newPolicies)
	if err != nil {
		t.Fatalf("Step 2 failed - SetGroupPolicies(): %v", err)
	}
	t.Log("✅ Step 2 complete: Policies updated")

	// Step 3: Register updated response and verify changes
	t.Log("🔍 Step 3: Verifying policy changes")

	// Now register the updated policies response for verification
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+groupID+"/policies", testhelpers.MockResponse{
		StatusCode: 200,
		Body:       updatedPolicies,
	})

	verifiedPolicies, err := client.GetGroupPolicies(ctx, groupID)
	if err != nil {
		t.Fatalf("Step 3 failed - GetGroupPolicies(): %v", err)
	}

	if !verifiedPolicies.JoinRequests {
		t.Error("Step 3 validation failed - Expected join requests to be enabled")
	}
	if len(verifiedPolicies.SignupFields) != 3 {
		t.Errorf("Step 3 validation failed - Expected 3 signup fields, got %d",
			len(verifiedPolicies.SignupFields))
	}
	t.Log("✅ Step 3 complete: Policy changes verified")

	t.Log("🎉 Policy workflow completed successfully!")
}

// TestErrorRecoveryWorkflow tests workflow resilience with error conditions
func TestErrorRecoveryWorkflow(t *testing.T) {
	groupID := "error-test-group"

	responseHandler := testhelpers.NewMockResponseHandler()

	// First attempt fails
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+groupID, testhelpers.MockResponse{
		StatusCode: 500,
		Body: map[string]string{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		},
	})

	// Create client
	server := httptest.NewServer(responseHandler)
	defer server.Close()

	authorizer := authorizers.StaticTokenCoreAuthorizer("test-error-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Log("🚀 Starting error recovery workflow")

	// Attempt 1: Should fail
	t.Log("❌ Attempt 1: Expected to fail")
	_, err = client.GetGroup(ctx, groupID)
	if err == nil {
		t.Error("Expected error on first attempt, got nil")
	} else {
		t.Logf("✅ Expected error occurred: %v", err)
	}

	// Update handler for successful retry
	responseHandler.RegisterResponse(http.MethodGet, "/groups/"+groupID, testhelpers.MockResponse{
		StatusCode: 200,
		Body: groups.Group{
			DATA_TYPE:   "group",
			ID:          groupID,
			Name:        "Recovered Group",
			Description: "Group retrieved after error recovery",
		},
	})

	// Attempt 2: Should succeed
	t.Log("✅ Attempt 2: Should succeed after recovery")
	recoveredGroup, err := client.GetGroup(ctx, groupID)
	if err != nil {
		t.Fatalf("Expected success on retry, got error: %v", err)
	}

	if recoveredGroup.Name != "Recovered Group" {
		t.Errorf("Expected group name 'Recovered Group', got %s", recoveredGroup.Name)
	}

	t.Log("🎉 Error recovery workflow completed successfully!")
}

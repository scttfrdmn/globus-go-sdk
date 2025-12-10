// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/testhelpers"
)

// TestSetSubscriptionAdminVerified tests the SetSubscriptionAdminVerified method (v3.63.0 updated naming)
func TestSetSubscriptionAdminVerified(t *testing.T) {
	// Setup mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		expectedPath := "/groups/test-group-12345/subscription"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Check request body
		var requestBody map[string]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		expectedSubscriptionID := "sub-abcdef-67890"
		if requestBody["subscription_id"] != expectedSubscriptionID {
			t.Errorf("Expected subscription_id=%s, got %s", expectedSubscriptionID, requestBody["subscription_id"])
		}

		if requestBody["DATA_TYPE"] != "subscription_update" {
			t.Errorf("Expected DATA_TYPE=subscription_update, got %s", requestBody["DATA_TYPE"])
		}

		w.WriteHeader(http.StatusOK)
	}

	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	// Test SetSubscriptionAdminVerified
	err := client.SetSubscriptionAdminVerified(context.Background(), "test-group-12345", "sub-abcdef-67890")
	if err != nil {
		t.Fatalf("SetSubscriptionAdminVerified() error = %v", err)
	}

	// Test error cases
	err = client.SetSubscriptionAdminVerified(context.Background(), "", "sub-12345")
	if err == nil {
		t.Error("SetSubscriptionAdminVerified() with empty group ID should return error")
	}

	err = client.SetSubscriptionAdminVerified(context.Background(), "test-group", "")
	if err == nil {
		t.Error("SetSubscriptionAdminVerified() with empty subscription ID should return error")
	}

	t.Log("✅ SetSubscriptionAdminVerified method tested successfully")
}

// TestSetSubscriptionAdminVerifiedID_Deprecated tests the deprecated method still works
func TestSetSubscriptionAdminVerifiedID_Deprecated(t *testing.T) {
	// Setup mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}

	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	// Test that deprecated method still works (should delegate to new method)
	err := client.SetSubscriptionAdminVerifiedID(context.Background(), "test-group-12345", "sub-abcdef-67890")
	if err != nil {
		t.Fatalf("SetSubscriptionAdminVerifiedID() deprecated method error = %v", err)
	}

	t.Log("✅ Deprecated SetSubscriptionAdminVerifiedID method still works via delegation")
}

// TestGetGroupSubscription tests the GetGroupSubscription method (v3.62.0 feature)
func TestGetGroupSubscription(t *testing.T) {
	now := time.Now()

	// Setup mock server
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		expectedPath := "/groups/test-group-12345/subscription"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		subscription := groups.GroupSubscription{
			DATA_TYPE:        "group_subscription",
			SubscriptionID:   "sub-abcdef-67890",
			GroupID:          "test-group-12345",
			IsActive:         true,
			Created:          now,
			LastUpdated:      now,
			SubscriptionType: "premium",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(subscription)
	}

	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	// Test GetGroupSubscription
	subscription, err := client.GetGroupSubscription(context.Background(), "test-group-12345")
	if err != nil {
		t.Fatalf("GetGroupSubscription() error = %v", err)
	}

	// Validate response
	if subscription.DATA_TYPE != "group_subscription" {
		t.Errorf("Expected DATA_TYPE=group_subscription, got %s", subscription.DATA_TYPE)
	}
	if subscription.SubscriptionID != "sub-abcdef-67890" {
		t.Errorf("Expected SubscriptionID=sub-abcdef-67890, got %s", subscription.SubscriptionID)
	}
	if subscription.GroupID != "test-group-12345" {
		t.Errorf("Expected GroupID=test-group-12345, got %s", subscription.GroupID)
	}
	if !subscription.IsActive {
		t.Errorf("Expected IsActive=true, got %v", subscription.IsActive)
	}
	if subscription.SubscriptionType != "premium" {
		t.Errorf("Expected SubscriptionType=premium, got %s", subscription.SubscriptionType)
	}

	// Test with empty group ID
	_, err = client.GetGroupSubscription(context.Background(), "")
	if err == nil {
		t.Error("GetGroupSubscription() with empty group ID should return error")
	}

	t.Log("✅ GetGroupSubscription method tested successfully")
}

// TestSubscriptionWorkflow tests the complete subscription workflow
func TestSubscriptionWorkflow(t *testing.T) {
	// Load test scenario using metadata-driven approach
	scenario := testhelpers.LoadTestScenario(t, "groups_enhanced", "subscription_set_admin_verified")

	// Create mock response handler for workflow
	mockHandler := testhelpers.NewMockResponseHandler()

	// Step 1: Set subscription ID (admin operation)
	mockHandler.RegisterResponse("PUT", "/groups/test-group-12345/subscription", testhelpers.MockResponse{
		StatusCode: 200,
		Body:       map[string]interface{}{"status": "success"},
	})

	// Step 2: Get subscription information
	mockHandler.RegisterResponse("GET", "/groups/test-group-12345/subscription", testhelpers.MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"DATA_TYPE":         "group_subscription",
			"subscription_id":   "sub-abcdef-67890",
			"group_id":          "test-group-12345",
			"is_active":         true,
			"subscription_type": "premium",
		},
	})

	// Step 3: Get group by subscription ID
	mockHandler.RegisterResponse("GET", "/groups", testhelpers.MockResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"DATA_TYPE": "group",
			"id":        "test-group-12345",
			"name":      "Test Premium Group",
		},
	})

	client, server, cleanup := testhelpers.MockGroupsClient(t, mockHandler.ServeHTTP)
	defer cleanup()

	ctx := context.Background()

	// Execute workflow steps
	t.Log("🔄 Step 1: Setting subscription admin verified ID...")
	err := client.SetSubscriptionAdminVerified(ctx, scenario.GroupID, "sub-abcdef-67890")
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	t.Log("✅ Step 1: Subscription ID set successfully")

	t.Log("🔄 Step 2: Retrieving subscription information...")
	subscription, err := client.GetGroupSubscription(ctx, scenario.GroupID)
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	if subscription.SubscriptionID != "sub-abcdef-67890" {
		t.Errorf("Step 2: Expected subscription ID sub-abcdef-67890, got %s", subscription.SubscriptionID)
	}
	t.Log("✅ Step 2: Subscription information retrieved successfully")

	t.Log("🔄 Step 3: Getting group by subscription ID...")
	group, err := client.GetGroupBySubscriptionID(ctx, "sub-abcdef-67890")
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	if group.ID != scenario.GroupID {
		t.Errorf("Step 3: Expected group ID %s, got %s", scenario.GroupID, group.ID)
	}
	t.Log("✅ Step 3: Group retrieved by subscription ID successfully")

	t.Log("✅ Complete subscription workflow executed successfully")

	_ = server // Avoid unused variable warning
}

// TestSubscriptionMethodErrorHandling tests error handling for subscription methods
func TestSubscriptionMethodErrorHandling(t *testing.T) {
	scenarios := []struct {
		name          string
		setupHandler  func() http.HandlerFunc
		testOperation func(*groups.Client) error
		expectedError bool
		errorMsg      string
	}{
		{
			name: "SetSubscriptionAdminVerified - Group not found",
			setupHandler: func() http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Group not found",
						"code":  "GROUP_NOT_FOUND",
					})
				})
			},
			testOperation: func(client *groups.Client) error {
				return client.SetSubscriptionAdminVerified(context.Background(), "nonexistent-group", "sub-123")
			},
			expectedError: true,
			errorMsg:      "Group not found error should be handled",
		},
		{
			name: "GetGroupSubscription - No subscription",
			setupHandler: func() http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "No subscription found for this group",
						"code":  "SUBSCRIPTION_NOT_FOUND",
					})
				})
			},
			testOperation: func(client *groups.Client) error {
				_, err := client.GetGroupSubscription(context.Background(), "group-without-subscription")
				return err
			},
			expectedError: true,
			errorMsg:      "No subscription error should be handled",
		},
		{
			name: "SetSubscriptionAdminVerified - Permission denied",
			setupHandler: func() http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Insufficient permissions to set subscription ID",
						"code":  "FORBIDDEN",
					})
				})
			},
			testOperation: func(client *groups.Client) error {
				return client.SetSubscriptionAdminVerified(context.Background(), "restricted-group", "sub-123")
			},
			expectedError: true,
			errorMsg:      "Permission denied error should be handled",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			client, _, cleanup := testhelpers.MockGroupsClient(t, scenario.setupHandler())
			defer cleanup()

			err := scenario.testOperation(client)

			if scenario.expectedError && err == nil {
				t.Errorf("%s: expected error but got none", scenario.errorMsg)
			} else if !scenario.expectedError && err != nil {
				t.Errorf("%s: unexpected error: %v", scenario.errorMsg, err)
			}

			if err != nil {
				t.Logf("✅ Error correctly handled: %v", err)
			}
		})
	}
}

// TestSubscriptionModelValidation tests the GroupSubscription model
func TestSubscriptionModelValidation(t *testing.T) {
	now := time.Now()

	subscription := groups.GroupSubscription{
		DATA_TYPE:        "group_subscription",
		SubscriptionID:   "sub-test-12345",
		GroupID:          "group-test-67890",
		IsActive:         true,
		Created:          now,
		LastUpdated:      now,
		SubscriptionType: "enterprise",
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.Marshal(subscription)
	if err != nil {
		t.Fatalf("Failed to marshal GroupSubscription: %v", err)
	}

	var unmarshaled groups.GroupSubscription
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal GroupSubscription: %v", err)
	}

	// Validate fields
	if unmarshaled.DATA_TYPE != "group_subscription" {
		t.Errorf("Expected DATA_TYPE=group_subscription, got %s", unmarshaled.DATA_TYPE)
	}
	if unmarshaled.SubscriptionID != "sub-test-12345" {
		t.Errorf("Expected SubscriptionID=sub-test-12345, got %s", unmarshaled.SubscriptionID)
	}
	if unmarshaled.GroupID != "group-test-67890" {
		t.Errorf("Expected GroupID=group-test-67890, got %s", unmarshaled.GroupID)
	}
	if !unmarshaled.IsActive {
		t.Errorf("Expected IsActive=true, got %v", unmarshaled.IsActive)
	}
	if unmarshaled.SubscriptionType != "enterprise" {
		t.Errorf("Expected SubscriptionType=enterprise, got %s", unmarshaled.SubscriptionType)
	}

	t.Log("✅ GroupSubscription model validation successful")
}

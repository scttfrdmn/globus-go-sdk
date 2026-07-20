// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/testhelpers"
)

// TestPythonSDKParityIntegration exercises the corrected 3.65.0 groups surface
// (subscription info, policies, preferences, membership fields, batch action) in
// one integrated flow against a mock server.
func TestPythonSDKParityIntegration(t *testing.T) {
	groupID := "test-group-12345"
	subscriptionID := "sub-abcdef-67890"

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/groups/"+groupID+"/subscription_admin_verified":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/subscription_info/"+subscriptionID:
			_ = json.NewEncoder(w).Encode(groups.Group{ID: groupID})
		case r.Method == http.MethodGet && r.URL.Path == "/groups/"+groupID+"/policies":
			_ = json.NewEncoder(w).Encode(groups.GroupPolicies{IsHighAssurance: true, GroupVisibility: "private"})
		case r.Method == http.MethodPut && r.URL.Path == "/groups/"+groupID+"/policies":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/preferences":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"allow_add": true})
			} else {
				w.WriteHeader(http.StatusOK)
			}
		case r.URL.Path == "/groups/"+groupID+"/membership_fields":
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"field": "value"})
			} else {
				w.WriteHeader(http.StatusOK)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/groups/"+groupID:
			_ = json.NewEncoder(w).Encode(groups.Group{ID: groupID})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}

	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	ctx := context.Background()

	if err := client.SetSubscriptionAdminVerified(ctx, groupID, subscriptionID); err != nil {
		t.Fatalf("SetSubscriptionAdminVerified: %v", err)
	}
	if _, err := client.GetGroupBySubscriptionID(ctx, subscriptionID); err != nil {
		t.Fatalf("GetGroupBySubscriptionID: %v", err)
	}
	if _, err := client.GetGroupPolicies(ctx, groupID); err != nil {
		t.Fatalf("GetGroupPolicies: %v", err)
	}
	if err := client.SetGroupPolicies(ctx, groupID, &groups.GroupPolicies{GroupVisibility: "authenticated"}); err != nil {
		t.Fatalf("SetGroupPolicies: %v", err)
	}
	if _, err := client.GetIdentityPreferences(ctx); err != nil {
		t.Fatalf("GetIdentityPreferences: %v", err)
	}
	if err := client.SetIdentityPreferences(ctx, map[string]interface{}{"allow_add": false}); err != nil {
		t.Fatalf("SetIdentityPreferences: %v", err)
	}
	if _, err := client.GetMembershipFields(ctx, groupID); err != nil {
		t.Fatalf("GetMembershipFields: %v", err)
	}
	if _, err := client.BatchMembershipAction(ctx, groupID, &groups.BatchMembershipActions{
		Add: []groups.MemberWithRole{{IdentityID: "u1", Role: groups.RoleMember}},
	}); err != nil {
		t.Fatalf("BatchMembershipAction: %v", err)
	}
}

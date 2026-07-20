// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors

package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/testhelpers"
)

// TestSetSubscriptionAdminVerified exercises the corrected 3.65.0 route
// (PUT /groups/{id}/subscription_admin_verified with a
// {"subscription_admin_verified_id": ...} body).
func TestSetSubscriptionAdminVerified(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/groups/g1/subscription_admin_verified" {
			t.Errorf("path = %s", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["subscription_admin_verified_id"] != "sub1" {
			t.Errorf("subscription_admin_verified_id = %v, want sub1", body["subscription_admin_verified_id"])
		}
		w.WriteHeader(http.StatusOK)
	}
	client, server, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()
	_ = server

	if err := client.SetSubscriptionAdminVerified(context.Background(), "g1", "sub1"); err != nil {
		t.Fatalf("SetSubscriptionAdminVerified() error = %v", err)
	}
}

// TestGetGroupBySubscriptionID exercises GET /subscription_info/{id}.
func TestGetGroupBySubscriptionID(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscription_info/sub1" {
			t.Errorf("path = %s, want /subscription_info/sub1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(groups.Group{ID: "g1"})
	}
	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	g, err := client.GetGroupBySubscriptionID(context.Background(), "sub1")
	if err != nil {
		t.Fatalf("GetGroupBySubscriptionID() error = %v", err)
	}
	if g.ID != "g1" {
		t.Errorf("ID = %s", g.ID)
	}
}

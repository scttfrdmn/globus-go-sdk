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

// TestGetGroupPolicies exercises GET /groups/{id}/policies with the corrected
// GroupPolicies shape.
func TestGetGroupPolicies(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/g1/policies" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(groups.GroupPolicies{
			IsHighAssurance:        true,
			GroupVisibility:        "private",
			GroupMembersVisibility: "managers",
		})
	}
	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	p, err := client.GetGroupPolicies(context.Background(), "g1")
	if err != nil {
		t.Fatalf("GetGroupPolicies() error = %v", err)
	}
	if !p.IsHighAssurance || p.GroupVisibility != "private" || p.GroupMembersVisibility != "managers" {
		t.Errorf("policies = %+v", p)
	}
}

// TestSetGroupPolicies exercises PUT /groups/{id}/policies.
func TestSetGroupPolicies(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/groups/g1/policies" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["is_high_assurance"]; !ok {
			t.Error("body missing is_high_assurance")
		}
		w.WriteHeader(http.StatusOK)
	}
	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	err := client.SetGroupPolicies(context.Background(), "g1", &groups.GroupPolicies{GroupVisibility: "authenticated"})
	if err != nil {
		t.Fatalf("SetGroupPolicies() error = %v", err)
	}
}

// TestGetSetIdentityPreferences exercises GET/PUT /preferences.
func TestGetSetIdentityPreferences(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/preferences" {
			t.Errorf("path = %s, want /preferences", r.URL.Path)
		}
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"allow_add": true})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	client, _, cleanup := testhelpers.MockGroupsClient(t, handler)
	defer cleanup()

	prefs, err := client.GetIdentityPreferences(context.Background())
	if err != nil {
		t.Fatalf("GetIdentityPreferences() error = %v", err)
	}
	if prefs["allow_add"] != true {
		t.Errorf("allow_add = %v", prefs["allow_add"])
	}
	if err := client.SetIdentityPreferences(context.Background(), map[string]interface{}{"allow_add": false}); err != nil {
		t.Fatalf("SetIdentityPreferences() error = %v", err)
	}
}

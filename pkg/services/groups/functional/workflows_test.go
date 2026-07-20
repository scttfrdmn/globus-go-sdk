// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors

package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"
)

func newWorkflowClient(t *testing.T, handler http.HandlerFunc) (*groups.Client, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	if err != nil {
		server.Close()
		t.Fatalf("failed to create client: %v", err)
	}
	return client, server.Close
}

// TestCompleteGroupManagementWorkflow exercises create -> get -> batch add ->
// set policies -> delete against the corrected 3.65.0 surface.
func TestCompleteGroupManagementWorkflow(t *testing.T) {
	client, cleanup := newWorkflowClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/groups":
			_ = json.NewEncoder(w).Encode(groups.Group{ID: "g1", Name: "WF Group"})
		case r.Method == http.MethodGet && r.URL.Path == "/groups/g1":
			_ = json.NewEncoder(w).Encode(groups.Group{ID: "g1", Name: "WF Group"})
		case r.Method == http.MethodPost && r.URL.Path == "/groups/g1":
			_ = json.NewEncoder(w).Encode(groups.Group{ID: "g1"})
		case r.Method == http.MethodPut && r.URL.Path == "/groups/g1/policies":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && r.URL.Path == "/groups/g1":
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	ctx := context.Background()
	created, err := client.CreateGroup(ctx, &groups.GroupCreate{Name: "WF Group"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if _, err := client.GetGroup(ctx, created.ID, &groups.GetGroupOptions{Include: []string{"memberships"}}); err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if _, err := client.BatchMembershipAction(ctx, created.ID, &groups.BatchMembershipActions{
		Add: []groups.MemberWithRole{{IdentityID: "id1", Role: groups.RoleMember}},
	}); err != nil {
		t.Fatalf("BatchMembershipAction: %v", err)
	}
	if err := client.SetGroupPolicies(ctx, created.ID, &groups.GroupPolicies{GroupVisibility: "private"}); err != nil {
		t.Fatalf("SetGroupPolicies: %v", err)
	}
	if err := client.DeleteGroup(ctx, created.ID); err != nil {
		t.Fatalf("DeleteGroup: %v", err)
	}
}

// TestMyGroupsWorkflow exercises GetMyGroups.
func TestMyGroupsWorkflow(t *testing.T) {
	client, cleanup := newWorkflowClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/my_groups" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]groups.Group{{ID: "g1"}, {ID: "g2"}})
	})
	defer cleanup()

	list, err := client.GetMyGroups(context.Background(), []string{"active"})
	if err != nil {
		t.Fatalf("GetMyGroups: %v", err)
	}
	if len(list.Groups) != 2 {
		t.Errorf("got %d groups, want 2", len(list.Groups))
	}
}

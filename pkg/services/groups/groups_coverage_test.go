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

// setupTestClient creates a groups client pointed at the given test server URL.
func setupTestClient(t *testing.T, serverURL string) *groups.Client {
	t.Helper()
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithCoreOptions(core.WithBaseURL(serverURL+"/")),
	)
	if err != nil {
		t.Fatalf("failed to create groups client: %v", err)
	}
	return client
}

// withMockServer starts an httptest server using handler and returns the server
// plus a client configured to use it. Caller must defer server.Close().
func withMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *groups.Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := setupTestClient(t, server.URL)
	return server, client
}

func TestNewClientFunctionalOptions(t *testing.T) {
	authorizer := authorizers.StaticTokenCoreAuthorizer("tok")
	client, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithHTTPDebugging(true),
		groups.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if _, err := groups.NewClient(); err == nil {
		t.Error("expected error when authorizer is missing")
	}
}

func TestGetMyGroupsExternal(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/my_groups" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]groups.Group{{ID: "g1"}})
	})
	defer server.Close()

	list, err := client.GetMyGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyGroups() error = %v", err)
	}
	if len(list.Groups) != 1 {
		t.Errorf("got %d groups, want 1", len(list.Groups))
	}
}

func TestBatchMembershipActionExternal(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/groups/g1" {
			t.Errorf("%s %s, want POST /groups/g1", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(groups.Group{ID: "g1"})
	})
	defer server.Close()

	_, err := client.BatchMembershipAction(context.Background(), "g1", &groups.BatchMembershipActions{
		ChangeRole: []groups.MemberWithRole{{IdentityID: "id1", Role: groups.RoleManager}},
	})
	if err != nil {
		t.Fatalf("BatchMembershipAction() error = %v", err)
	}
}

func TestSetGroupPoliciesExternal(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/groups/g1/policies" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetGroupPolicies(context.Background(), "g1", &groups.GroupPolicies{GroupVisibility: "authenticated"})
	if err != nil {
		t.Fatalf("SetGroupPolicies() error = %v", err)
	}
}

func TestSetIdentityPreferencesExternal(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/preferences" {
			t.Errorf("%s %s, want PUT /preferences", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetIdentityPreferences(context.Background(), map[string]interface{}{"allow_add": false})
	if err != nil {
		t.Fatalf("SetIdentityPreferences() error = %v", err)
	}
}

func TestMembershipFieldsExternal(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/g1/membership_fields" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"field": "value"})
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	fields, err := client.GetMembershipFields(context.Background(), "g1")
	if err != nil {
		t.Fatalf("GetMembershipFields() error = %v", err)
	}
	if fields["field"] != "value" {
		t.Errorf("fields = %v", fields)
	}
	if err := client.SetMembershipFields(context.Background(), "g1", map[string]interface{}{"field": "v2"}); err != nil {
		t.Fatalf("SetMembershipFields() error = %v", err)
	}
}

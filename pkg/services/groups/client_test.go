// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

// setupMockServer sets up a mock server and a client pointed at it.
func setupMockServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")
	client, _ := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOptions(core.WithBaseURL(server.URL+"/")),
	)
	return server, client
}

func TestBuildURLLowLevel(t *testing.T) {
	authorizer := authorizers.StaticTokenCoreAuthorizer("test-access-token")
	client, _ := NewClient(
		WithAuthorizer(authorizer),
		WithCoreOptions(core.WithBaseURL("https://example.com")),
	)

	if url := client.buildURLLowLevel("test/path", nil); url != "https://example.com/test/path" {
		t.Errorf("buildURL() = %v, want https://example.com/test/path", url)
	}
	query := map[string][]string{"param1": {"value1"}}
	if url := client.buildURLLowLevel("test/path", query); url != "https://example.com/test/path?param1=value1" {
		t.Errorf("buildURL() with query = %v", url)
	}
}

func TestGetMyGroups(t *testing.T) {
	t.Run("hits my_groups with comma-joined statuses and decodes array", func(t *testing.T) {
		server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/groups/my_groups" {
				t.Errorf("path = %s, want /groups/my_groups", r.URL.Path)
			}
			if got := r.URL.Query().Get("statuses"); got != "active,invited" {
				t.Errorf("statuses = %q, want active,invited", got)
			}
			_ = json.NewEncoder(w).Encode([]Group{{ID: "g1"}, {ID: "g2"}})
		})
		defer server.Close()

		result, err := client.GetMyGroups(context.Background(), []string{"active", "invited"})
		if err != nil {
			t.Fatalf("GetMyGroups() error = %v", err)
		}
		if len(result.Groups) != 2 {
			t.Errorf("got %d groups, want 2", len(result.Groups))
		}
	})
}

func TestGetGroup(t *testing.T) {
	t.Run("empty ID returns error", func(t *testing.T) {
		server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {})
		defer server.Close()
		if _, err := client.GetGroup(context.Background(), "", nil); err == nil {
			t.Error("expected error for empty group ID")
		}
	})

	t.Run("include is comma-joined", func(t *testing.T) {
		server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/groups/g1" {
				t.Errorf("path = %s", r.URL.Path)
			}
			if got := r.URL.Query().Get("include"); got != "memberships,policies" {
				t.Errorf("include = %q, want memberships,policies", got)
			}
			_ = json.NewEncoder(w).Encode(Group{ID: "g1", Name: "G"})
		})
		defer server.Close()

		g, err := client.GetGroup(context.Background(), "g1", &GetGroupOptions{Include: []string{"memberships", "policies"}})
		if err != nil {
			t.Fatalf("GetGroup() error = %v", err)
		}
		if g.ID != "g1" {
			t.Errorf("ID = %s", g.ID)
		}
	})
}

func TestCreateGroup(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/groups" {
			t.Errorf("%s %s, want POST /groups", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Group{ID: "new", Name: "New"})
	})
	defer server.Close()

	g, err := client.CreateGroup(context.Background(), &GroupCreate{Name: "New"})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if g.ID != "new" {
		t.Errorf("ID = %s", g.ID)
	}
}

func TestUpdateGroup(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		_ = json.NewEncoder(w).Encode(Group{ID: "g1", Name: "Renamed"})
	})
	defer server.Close()

	g, err := client.UpdateGroup(context.Background(), "g1", &GroupUpdate{Name: "Renamed"})
	if err != nil {
		t.Fatalf("UpdateGroup() error = %v", err)
	}
	if g.Name != "Renamed" {
		t.Errorf("Name = %s", g.Name)
	}
}

func TestDeleteGroup(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	if err := client.DeleteGroup(context.Background(), "g1"); err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}
}

func TestBatchMembershipAction(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/groups/g1" {
			t.Errorf("%s %s, want POST /groups/g1", r.Method, r.URL.Path)
		}
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["add"]; !ok {
			t.Error("expected an add action key")
		}
		_ = json.NewEncoder(w).Encode(Group{ID: "g1"})
	})
	defer server.Close()

	_, err := client.BatchMembershipAction(context.Background(), "g1", &BatchMembershipActions{
		Add: []MemberWithRole{{IdentityID: "id1", Role: RoleMember}},
	})
	if err != nil {
		t.Fatalf("BatchMembershipAction() error = %v", err)
	}
}

func TestSetSubscriptionAdminVerified(t *testing.T) {
	t.Run("nil subscription sends JSON null", func(t *testing.T) {
		server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut || r.URL.Path != "/groups/g1/subscription_admin_verified" {
				t.Errorf("%s %s", r.Method, r.URL.Path)
			}
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			v, present := body["subscription_admin_verified_id"]
			if !present || v != nil {
				t.Errorf("subscription_admin_verified_id = %v (present=%v), want null", v, present)
			}
			w.WriteHeader(http.StatusOK)
		})
		defer server.Close()

		if err := client.SetSubscriptionAdminVerified(context.Background(), "g1", ""); err != nil {
			t.Fatalf("SetSubscriptionAdminVerified() error = %v", err)
		}
	})
}

func TestGetGroupBySubscriptionID(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/subscription_info/sub1" {
			t.Errorf("path = %s, want /subscription_info/sub1", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Group{ID: "g1"})
	})
	defer server.Close()

	g, err := client.GetGroupBySubscriptionID(context.Background(), "sub1")
	if err != nil {
		t.Fatalf("GetGroupBySubscriptionID() error = %v", err)
	}
	if g.ID != "g1" {
		t.Errorf("ID = %s", g.ID)
	}
}

func TestGroupPolicies(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/g1/policies" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(GroupPolicies{IsHighAssurance: true, GroupVisibility: "private"})
	})
	defer server.Close()

	p, err := client.GetGroupPolicies(context.Background(), "g1")
	if err != nil {
		t.Fatalf("GetGroupPolicies() error = %v", err)
	}
	if !p.IsHighAssurance || p.GroupVisibility != "private" {
		t.Errorf("policies = %+v", p)
	}
}

func TestIdentityPreferences(t *testing.T) {
	server, client := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/preferences" {
			t.Errorf("path = %s, want /preferences", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"allow_add": true})
	})
	defer server.Close()

	prefs, err := client.GetIdentityPreferences(context.Background())
	if err != nil {
		t.Fatalf("GetIdentityPreferences() error = %v", err)
	}
	if prefs["allow_add"] != true {
		t.Errorf("allow_add = %v", prefs["allow_add"])
	}
}

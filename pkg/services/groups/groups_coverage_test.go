// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// ---------------------------------------------------------------------------
// TestNewClient – functional options
// ---------------------------------------------------------------------------

func TestNewClientFunctionalOptions(t *testing.T) {
	authorizer := authorizers.StaticTokenCoreAuthorizer("tok")

	// WithHTTPDebugging
	_, err := groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithHTTPDebugging(true),
	)
	if err != nil {
		t.Fatalf("NewClient with WithHTTPDebugging: %v", err)
	}

	// WithHTTPTracing
	_, err = groups.NewClient(
		groups.WithAuthorizer(authorizer),
		groups.WithHTTPTracing(true),
	)
	if err != nil {
		t.Fatalf("NewClient with WithHTTPTracing: %v", err)
	}

	// No authorizer – must return error
	_, err = groups.NewClient()
	if err == nil {
		t.Fatal("NewClient without authorizer should return error")
	}
}

// ---------------------------------------------------------------------------
// TestListRoles
// ---------------------------------------------------------------------------

func TestListRoles(t *testing.T) {
	roles := []groups.Role{
		{ID: "admin", Name: "Administrator", Description: "Full access"},
		{ID: "member", Name: "Member", Description: "Read-only access"},
	}

	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/roles" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.RoleList{Roles: roles})
	})
	defer server.Close()

	list, err := client.ListRoles(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(list.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(list.Roles))
	}
	if list.Roles[0].ID != "admin" {
		t.Errorf("roles[0].ID = %q, want admin", list.Roles[0].ID)
	}

	// empty group ID
	_, err = client.ListRoles(context.Background(), "")
	if err == nil {
		t.Error("ListRoles with empty group ID should error")
	}
}

func TestListRolesDataTypeFilled(t *testing.T) {
	// Verify client fills DATA_TYPE when server omits it.
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Return roles without DATA_TYPE set
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": []map[string]interface{}{
				{"id": "admin", "name": "Administrator"},
			},
		})
	})
	defer server.Close()

	list, err := client.ListRoles(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if list.Roles[0].DATA_TYPE != "role" {
		t.Errorf("expected DATA_TYPE=role, got %q", list.Roles[0].DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestGetRole
// ---------------------------------------------------------------------------

func TestGetRole(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/roles/admin" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.Role{
			DATA_TYPE:   "role",
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full access",
		})
	})
	defer server.Close()

	role, err := client.GetRole(context.Background(), "grp-1", "admin")
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if role.ID != "admin" {
		t.Errorf("role.ID = %q, want admin", role.ID)
	}
	if role.DATA_TYPE != "role" {
		t.Errorf("role.DATA_TYPE = %q, want role", role.DATA_TYPE)
	}

	// empty group ID
	_, err = client.GetRole(context.Background(), "", "admin")
	if err == nil {
		t.Error("GetRole with empty group ID should error")
	}

	// empty role ID
	_, err = client.GetRole(context.Background(), "grp-1", "")
	if err == nil {
		t.Error("GetRole with empty role ID should error")
	}
}

func TestGetRoleDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Omit DATA_TYPE from response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "member",
			"name": "Member",
		})
	})
	defer server.Close()

	role, err := client.GetRole(context.Background(), "grp-1", "member")
	if err != nil {
		t.Fatalf("GetRole: %v", err)
	}
	if role.DATA_TYPE != "role" {
		t.Errorf("expected DATA_TYPE=role, got %q", role.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestCreateRole
// ---------------------------------------------------------------------------

func TestCreateRole(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/roles" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req groups.RoleCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.DATA_TYPE != "role_create" {
			t.Errorf("expected DATA_TYPE=role_create, got %q", req.DATA_TYPE)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.Role{
			DATA_TYPE:   "role",
			ID:          "new-role",
			Name:        req.Name,
			Description: req.Description,
		})
	})
	defer server.Close()

	role, err := client.CreateRole(context.Background(), "grp-1", &groups.RoleCreate{
		Name:        "Custom Role",
		Description: "A custom role",
	})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.ID != "new-role" {
		t.Errorf("role.ID = %q, want new-role", role.ID)
	}
	if role.Name != "Custom Role" {
		t.Errorf("role.Name = %q, want Custom Role", role.Name)
	}

	// nil role
	_, err = client.CreateRole(context.Background(), "grp-1", nil)
	if err == nil {
		t.Error("CreateRole with nil role should error")
	}

	// empty group ID
	_, err = client.CreateRole(context.Background(), "", &groups.RoleCreate{Name: "x"})
	if err == nil {
		t.Error("CreateRole with empty group ID should error")
	}

	// empty name
	_, err = client.CreateRole(context.Background(), "grp-1", &groups.RoleCreate{})
	if err == nil {
		t.Error("CreateRole with empty name should error")
	}
}

// ---------------------------------------------------------------------------
// TestUpdateRole
// ---------------------------------------------------------------------------

func TestUpdateRole(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/roles/admin" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req groups.RoleUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if req.DATA_TYPE != "role_update" {
			t.Errorf("expected DATA_TYPE=role_update, got %q", req.DATA_TYPE)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.Role{
			DATA_TYPE:   "role",
			ID:          "admin",
			Name:        req.Name,
			Description: req.Description,
		})
	})
	defer server.Close()

	role, err := client.UpdateRole(context.Background(), "grp-1", "admin", &groups.RoleUpdate{
		Name:        "Super Admin",
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if role.Name != "Super Admin" {
		t.Errorf("role.Name = %q, want Super Admin", role.Name)
	}

	// empty group ID
	_, err = client.UpdateRole(context.Background(), "", "admin", &groups.RoleUpdate{Name: "x"})
	if err == nil {
		t.Error("UpdateRole with empty group ID should error")
	}

	// empty role ID
	_, err = client.UpdateRole(context.Background(), "grp-1", "", &groups.RoleUpdate{Name: "x"})
	if err == nil {
		t.Error("UpdateRole with empty role ID should error")
	}

	// nil update
	_, err = client.UpdateRole(context.Background(), "grp-1", "admin", nil)
	if err == nil {
		t.Error("UpdateRole with nil update should error")
	}
}

// ---------------------------------------------------------------------------
// TestDeleteRole
// ---------------------------------------------------------------------------

func TestDeleteRole(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/roles/admin" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.DeleteRole(context.Background(), "grp-1", "admin")
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	// empty group ID
	err = client.DeleteRole(context.Background(), "", "admin")
	if err == nil {
		t.Error("DeleteRole with empty group ID should error")
	}

	// empty role ID
	err = client.DeleteRole(context.Background(), "grp-1", "")
	if err == nil {
		t.Error("DeleteRole with empty role ID should error")
	}
}

// ---------------------------------------------------------------------------
// TestGetMyGroups
// ---------------------------------------------------------------------------

func TestGetMyGroups(t *testing.T) {
	now := time.Now()
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/groups" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("my_groups") != "true" {
			t.Errorf("expected my_groups=true, got %q", q.Get("my_groups"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupList{
			Groups: []groups.Group{
				{ID: "g1", Name: "My Group", Created: now, LastUpdated: now},
			},
		})
	})
	defer server.Close()

	list, err := client.GetMyGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetMyGroups: %v", err)
	}
	if len(list.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(list.Groups))
	}
	if list.Groups[0].ID != "g1" {
		t.Errorf("groups[0].ID = %q, want g1", list.Groups[0].ID)
	}
}

func TestGetMyGroupsWithStatuses(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("my_groups") != "true" {
			t.Errorf("expected my_groups=true, got %q", q.Get("my_groups"))
		}
		statuses := q["statuses"]
		if len(statuses) != 2 {
			t.Errorf("expected 2 statuses, got %d", len(statuses))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupList{Groups: []groups.Group{}})
	})
	defer server.Close()

	_, err := client.GetMyGroups(context.Background(), []string{"active", "pending"})
	if err != nil {
		t.Fatalf("GetMyGroups with statuses: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestListGroupsV2
// ---------------------------------------------------------------------------

func TestListGroupsV2(t *testing.T) {
	now := time.Now()
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupList{
			Groups: []groups.Group{
				{ID: "g1", Name: "Group 1", Created: now, LastUpdated: now},
				{ID: "g2", Name: "Group 2", Created: now, LastUpdated: now},
			},
			HasNextPage: false,
		})
	})
	defer server.Close()

	resp, err := client.ListGroupsV2(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGroupsV2: %v", err)
	}
	if len(resp.Data.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(resp.Data.Groups))
	}
}

func TestListGroupsV2WithOptions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("my_groups") != "true" {
			t.Errorf("expected my_groups=true")
		}
		if q.Get("per_page") != "5" {
			t.Errorf("expected per_page=5, got %q", q.Get("per_page"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupList{Groups: []groups.Group{}})
	})
	defer server.Close()

	_, err := client.ListGroupsV2(context.Background(), &groups.ListGroupsOptions{
		MyGroups: true,
		PageSize: 5,
	})
	if err != nil {
		t.Fatalf("ListGroupsV2 with options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestListGroupsAllOptions – ensure all ListGroupsOptions fields are exercised
// ---------------------------------------------------------------------------

func TestListGroupsAllOptions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"include_group_membership": "true",
			"include_identity_set":     "true",
			"for_user_id":              "user-123",
			"my_groups":                "true",
			"per_page":                 "20",
			"marker":                   "tok-abc",
		}
		for k, v := range checks {
			if got := q.Get(k); got != v {
				t.Errorf("param %s: got %q, want %q", k, got, v)
			}
		}
		statuses := q["statuses"]
		if len(statuses) != 1 || statuses[0] != "active" {
			t.Errorf("expected statuses=[active], got %v", statuses)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupList{Groups: []groups.Group{}})
	})
	defer server.Close()

	_, err := client.ListGroups(context.Background(), &groups.ListGroupsOptions{
		IncludeGroupMembership: true,
		IncludeIdentitySet:     true,
		ForUserID:              "user-123",
		MyGroups:               true,
		Statuses:               []string{"active"},
		PageSize:               20,
		PageToken:              "tok-abc",
	})
	if err != nil {
		t.Fatalf("ListGroups with all options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestGetGroupBySubscriptionID
// ---------------------------------------------------------------------------

func TestGetGroupBySubscriptionIDCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("subscription_id") != "sub-999" {
			t.Errorf("expected subscription_id=sub-999")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.Group{
			ID:   "grp-sub",
			Name: "Subscription Group",
		})
	})
	defer server.Close()

	grp, err := client.GetGroupBySubscriptionID(context.Background(), "sub-999")
	if err != nil {
		t.Fatalf("GetGroupBySubscriptionID: %v", err)
	}
	if grp.ID != "grp-sub" {
		t.Errorf("grp.ID = %q, want grp-sub", grp.ID)
	}
	// DATA_TYPE filled
	if grp.DATA_TYPE != "group" {
		t.Errorf("grp.DATA_TYPE = %q, want group", grp.DATA_TYPE)
	}

	// empty subscription ID
	_, err = client.GetGroupBySubscriptionID(context.Background(), "")
	if err == nil {
		t.Error("GetGroupBySubscriptionID with empty ID should error")
	}
}

// ---------------------------------------------------------------------------
// TestGetGroupDataTypeFilled
// ---------------------------------------------------------------------------

func TestGetGroupDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Omit DATA_TYPE in response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "grp-1",
			"name": "Test",
		})
	})
	defer server.Close()

	grp, err := client.GetGroup(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("GetGroup: %v", err)
	}
	if grp.DATA_TYPE != "group" {
		t.Errorf("grp.DATA_TYPE = %q, want group", grp.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestCreateGroupDataTypeFilled
// ---------------------------------------------------------------------------

func TestCreateGroupDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req groups.GroupCreate
		json.NewDecoder(r.Body).Decode(&req)
		if req.DATA_TYPE != "group_create" {
			t.Errorf("expected DATA_TYPE=group_create, got %q", req.DATA_TYPE)
		}
		w.Header().Set("Content-Type", "application/json")
		// Omit DATA_TYPE in response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "grp-new",
			"name": req.Name,
		})
	})
	defer server.Close()

	grp, err := client.CreateGroup(context.Background(), &groups.GroupCreate{Name: "Fresh Group"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if grp.DATA_TYPE != "group" {
		t.Errorf("grp.DATA_TYPE = %q, want group", grp.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestUpdateGroupDataTypeFilled
// ---------------------------------------------------------------------------

func TestUpdateGroupDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req groups.GroupUpdate
		json.NewDecoder(r.Body).Decode(&req)
		if req.DATA_TYPE != "group_update" {
			t.Errorf("expected DATA_TYPE=group_update, got %q", req.DATA_TYPE)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "grp-1",
			"name": "Updated",
		})
	})
	defer server.Close()

	grp, err := client.UpdateGroup(context.Background(), "grp-1", &groups.GroupUpdate{Name: "Updated"})
	if err != nil {
		t.Fatalf("UpdateGroup: %v", err)
	}
	if grp.DATA_TYPE != "group" {
		t.Errorf("grp.DATA_TYPE = %q, want group", grp.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestListGroupsDataTypeFilled
// ---------------------------------------------------------------------------

func TestListGroupsDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Groups without DATA_TYPE set
		json.NewEncoder(w).Encode(map[string]interface{}{
			"groups": []map[string]interface{}{
				{"id": "g1", "name": "Group 1"},
				{"id": "g2", "name": "Group 2"},
			},
		})
	})
	defer server.Close()

	list, err := client.ListGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	for _, g := range list.Groups {
		if g.DATA_TYPE != "group" {
			t.Errorf("group DATA_TYPE = %q, want group", g.DATA_TYPE)
		}
	}
}

// ---------------------------------------------------------------------------
// TestListMembersAllOptions
// ---------------------------------------------------------------------------

func TestListMembersAllOptions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"role_id":  "admin",
			"status":   "active",
			"per_page": "15",
			"marker":   "page-token",
		}
		for k, v := range checks {
			if got := q.Get(k); got != v {
				t.Errorf("param %s: got %q, want %q", k, got, v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.MemberList{Members: []groups.Member{}})
	})
	defer server.Close()

	_, err := client.ListMembers(context.Background(), "grp-1", &groups.ListMembersOptions{
		RoleID:    "admin",
		Status:    "active",
		PageSize:  15,
		PageToken: "page-token",
	})
	if err != nil {
		t.Fatalf("ListMembers with all options: %v", err)
	}
}

func TestListMembersNilOptions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.MemberList{Members: []groups.Member{}})
	})
	defer server.Close()

	_, err := client.ListMembers(context.Background(), "grp-1", nil)
	if err != nil {
		t.Fatalf("ListMembers with nil options: %v", err)
	}
}

func TestListMembersDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"members": []map[string]interface{}{
				{
					"identity_id": "mem-1",
					"username":    "user1",
					"role":        map[string]interface{}{"id": "admin", "name": "Admin"},
				},
			},
		})
	})
	defer server.Close()

	list, err := client.ListMembers(context.Background(), "grp-1", nil)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(list.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(list.Members))
	}
	if list.Members[0].DATA_TYPE != "member" {
		t.Errorf("member DATA_TYPE = %q, want member", list.Members[0].DATA_TYPE)
	}
	if list.Members[0].Role.DATA_TYPE != "role" {
		t.Errorf("member role DATA_TYPE = %q, want role", list.Members[0].Role.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestSetSubscriptionAdminVerifiedID (deprecated wrapper)
// ---------------------------------------------------------------------------

func TestSetSubscriptionAdminVerifiedIDDeprecated(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/subscription" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetSubscriptionAdminVerifiedID(context.Background(), "grp-1", "sub-abc")
	if err != nil {
		t.Fatalf("SetSubscriptionAdminVerifiedID: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestGetGroupPoliciesCoverage
// ---------------------------------------------------------------------------

func TestGetGroupPoliciesCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/policies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.GroupPolicies{
			DATA_TYPE: "group_policies",
			GroupID:   "grp-1",
			Policies:  map[string]interface{}{"join_requests": true},
		})
	})
	defer server.Close()

	p, err := client.GetGroupPolicies(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("GetGroupPolicies: %v", err)
	}
	if p.GroupID != "grp-1" {
		t.Errorf("p.GroupID = %q, want grp-1", p.GroupID)
	}
	if p.DATA_TYPE != "group_policies" {
		t.Errorf("p.DATA_TYPE = %q, want group_policies", p.DATA_TYPE)
	}

	// empty group ID
	_, err = client.GetGroupPolicies(context.Background(), "")
	if err == nil {
		t.Error("GetGroupPolicies with empty group ID should error")
	}
}

func TestGetGroupPoliciesDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Omit DATA_TYPE
		json.NewEncoder(w).Encode(map[string]interface{}{
			"group_id": "grp-1",
			"policies": map[string]interface{}{},
		})
	})
	defer server.Close()

	p, err := client.GetGroupPolicies(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("GetGroupPolicies: %v", err)
	}
	if p.DATA_TYPE != "group_policies" {
		t.Errorf("expected DATA_TYPE=group_policies, got %q", p.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestSetGroupPoliciesCoverage
// ---------------------------------------------------------------------------

func TestSetGroupPoliciesCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/policies" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req groups.GroupPolicies
		json.NewDecoder(r.Body).Decode(&req)
		if req.DATA_TYPE != "group_policies_update" {
			t.Errorf("expected DATA_TYPE=group_policies_update, got %q", req.DATA_TYPE)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetGroupPolicies(context.Background(), "grp-1", &groups.GroupPolicies{
		Policies: map[string]interface{}{"join_requests": false},
	})
	if err != nil {
		t.Fatalf("SetGroupPolicies: %v", err)
	}

	// empty group ID
	err = client.SetGroupPolicies(context.Background(), "", &groups.GroupPolicies{})
	if err == nil {
		t.Error("SetGroupPolicies with empty group ID should error")
	}

	// nil policies
	err = client.SetGroupPolicies(context.Background(), "grp-1", nil)
	if err == nil {
		t.Error("SetGroupPolicies with nil policies should error")
	}
}

// ---------------------------------------------------------------------------
// TestGetIdentityPreferencesCoverage
// ---------------------------------------------------------------------------

func TestGetIdentityPreferencesCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/grp-1/identity_preferences/ident-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.IdentityPreferences{
			DATA_TYPE:   "identity_preferences",
			GroupID:     "grp-1",
			IdentityID:  "ident-1",
			Preferences: map[string]interface{}{"notify": true},
		})
	})
	defer server.Close()

	prefs, err := client.GetIdentityPreferences(context.Background(), "grp-1", "ident-1")
	if err != nil {
		t.Fatalf("GetIdentityPreferences: %v", err)
	}
	if prefs.IdentityID != "ident-1" {
		t.Errorf("prefs.IdentityID = %q, want ident-1", prefs.IdentityID)
	}

	// empty group ID
	_, err = client.GetIdentityPreferences(context.Background(), "", "ident-1")
	if err == nil {
		t.Error("GetIdentityPreferences with empty group ID should error")
	}

	// empty identity ID
	_, err = client.GetIdentityPreferences(context.Background(), "grp-1", "")
	if err == nil {
		t.Error("GetIdentityPreferences with empty identity ID should error")
	}
}

func TestGetIdentityPreferencesDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"group_id":    "grp-1",
			"identity_id": "ident-1",
			"preferences": map[string]interface{}{},
		})
	})
	defer server.Close()

	prefs, err := client.GetIdentityPreferences(context.Background(), "grp-1", "ident-1")
	if err != nil {
		t.Fatalf("GetIdentityPreferences: %v", err)
	}
	if prefs.DATA_TYPE != "identity_preferences" {
		t.Errorf("expected DATA_TYPE=identity_preferences, got %q", prefs.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestSetIdentityPreferencesCoverage
// ---------------------------------------------------------------------------

func TestSetIdentityPreferencesCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/identity_preferences/ident-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req groups.IdentityPreferences
		json.NewDecoder(r.Body).Decode(&req)
		if req.DATA_TYPE != "identity_preferences_update" {
			t.Errorf("expected DATA_TYPE=identity_preferences_update, got %q", req.DATA_TYPE)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetIdentityPreferences(context.Background(), "grp-1", "ident-1", &groups.IdentityPreferences{
		Preferences: map[string]interface{}{"notify": false},
	})
	if err != nil {
		t.Fatalf("SetIdentityPreferences: %v", err)
	}

	// empty group ID
	err = client.SetIdentityPreferences(context.Background(), "", "ident-1", &groups.IdentityPreferences{})
	if err == nil {
		t.Error("SetIdentityPreferences with empty group ID should error")
	}

	// empty identity ID
	err = client.SetIdentityPreferences(context.Background(), "grp-1", "", &groups.IdentityPreferences{})
	if err == nil {
		t.Error("SetIdentityPreferences with empty identity ID should error")
	}

	// nil preferences
	err = client.SetIdentityPreferences(context.Background(), "grp-1", "ident-1", nil)
	if err == nil {
		t.Error("SetIdentityPreferences with nil preferences should error")
	}
}

// ---------------------------------------------------------------------------
// TestGetMembershipFieldsCoverage
// ---------------------------------------------------------------------------

func TestGetMembershipFieldsCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/grp-1/membership_fields" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.MembershipFields{
			DATA_TYPE: "membership_fields",
			GroupID:   "grp-1",
			Fields:    map[string]interface{}{"department": "required"},
		})
	})
	defer server.Close()

	fields, err := client.GetMembershipFields(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("GetMembershipFields: %v", err)
	}
	if fields.GroupID != "grp-1" {
		t.Errorf("fields.GroupID = %q, want grp-1", fields.GroupID)
	}

	// empty group ID
	_, err = client.GetMembershipFields(context.Background(), "")
	if err == nil {
		t.Error("GetMembershipFields with empty group ID should error")
	}
}

func TestGetMembershipFieldsDataTypeFilled(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"group_id": "grp-1",
			"fields":   map[string]interface{}{},
		})
	})
	defer server.Close()

	fields, err := client.GetMembershipFields(context.Background(), "grp-1")
	if err != nil {
		t.Fatalf("GetMembershipFields: %v", err)
	}
	if fields.DATA_TYPE != "membership_fields" {
		t.Errorf("expected DATA_TYPE=membership_fields, got %q", fields.DATA_TYPE)
	}
}

// ---------------------------------------------------------------------------
// TestSetMembershipFieldsCoverage
// ---------------------------------------------------------------------------

func TestSetMembershipFieldsCoverage(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/membership_fields" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req groups.MembershipFields
		json.NewDecoder(r.Body).Decode(&req)
		if req.DATA_TYPE != "membership_fields_update" {
			t.Errorf("expected DATA_TYPE=membership_fields_update, got %q", req.DATA_TYPE)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.SetMembershipFields(context.Background(), "grp-1", &groups.MembershipFields{
		Fields: map[string]interface{}{"phone": "optional"},
	})
	if err != nil {
		t.Fatalf("SetMembershipFields: %v", err)
	}

	// empty group ID
	err = client.SetMembershipFields(context.Background(), "", &groups.MembershipFields{})
	if err == nil {
		t.Error("SetMembershipFields with empty group ID should error")
	}

	// nil fields
	err = client.SetMembershipFields(context.Background(), "grp-1", nil)
	if err == nil {
		t.Error("SetMembershipFields with nil fields should error")
	}
}

// ---------------------------------------------------------------------------
// TestChangeRole
// ---------------------------------------------------------------------------

func TestChangeRole(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/members/ident-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		if req["role_id"] != "admin" {
			t.Errorf("expected role_id=admin, got %v", req["role_id"])
		}
		// Return a simple map so the response unmarshal doesn't fail
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	})
	defer server.Close()

	err := client.ChangeRole(context.Background(), "grp-1", "ident-1", "admin")
	if err != nil {
		t.Fatalf("ChangeRole: %v", err)
	}

	// empty group ID
	err = client.ChangeRole(context.Background(), "", "ident-1", "admin")
	if err == nil {
		t.Error("ChangeRole with empty group ID should error")
	}

	// empty identity ID
	err = client.ChangeRole(context.Background(), "grp-1", "", "admin")
	if err == nil {
		t.Error("ChangeRole with empty identity ID should error")
	}

	// empty role ID
	err = client.ChangeRole(context.Background(), "grp-1", "ident-1", "")
	if err == nil {
		t.Error("ChangeRole with empty role ID should error")
	}
}

// ---------------------------------------------------------------------------
// TestChangeRoles
// ---------------------------------------------------------------------------

func TestChangeRoles(t *testing.T) {
	callCount := 0
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	})
	defer server.Close()

	changes := []groups.RoleChange{
		{GroupID: "grp-1", IdentityID: "ident-1", RoleID: "admin"},
		{GroupID: "grp-1", IdentityID: "ident-2", RoleID: "member"},
	}

	result, err := client.ChangeRoles(context.Background(), changes)
	if err != nil {
		t.Fatalf("ChangeRoles: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
	for i, r := range result.Results {
		if !r.Success {
			t.Errorf("result[%d] not successful", i)
		}
	}
	if result.DATA_TYPE != "batch_role_change_result" {
		t.Errorf("result.DATA_TYPE = %q", result.DATA_TYPE)
	}

	// empty changes
	_, err = client.ChangeRoles(context.Background(), nil)
	if err == nil {
		t.Error("ChangeRoles with nil changes should error")
	}

	_, err = client.ChangeRoles(context.Background(), []groups.RoleChange{})
	if err == nil {
		t.Error("ChangeRoles with empty changes should error")
	}
}

// ---------------------------------------------------------------------------
// TestBatchMembershipActions
// ---------------------------------------------------------------------------

func TestBatchMembershipActions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	})
	defer server.Close()

	batch := client.NewBatchMembershipActions()
	batch.ChangeRole("grp-1", "ident-1", "admin").
		ChangeRole("grp-1", "ident-2", "member")

	result, err := batch.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}
}

func TestBatchMembershipActionsEmpty(t *testing.T) {
	authorizer := authorizers.StaticTokenCoreAuthorizer("tok")
	client, err := groups.NewClient(groups.WithAuthorizer(authorizer))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	batch := client.NewBatchMembershipActions()
	result, err := batch.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute with empty batch: %v", err)
	}
	if len(result.Results) != 0 {
		t.Errorf("expected 0 results for empty batch, got %d", len(result.Results))
	}
}

// ---------------------------------------------------------------------------
// TestListMembersLowLevel
// ---------------------------------------------------------------------------

func TestListMembersLowLevel(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups/grp-1/members" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.MemberList{
			Members: []groups.Member{
				{IdentityID: "mem-1", Username: "alice"},
			},
		})
	})
	defer server.Close()

	list, err := client.ListMembersLowLevel(context.Background(), "grp-1", nil)
	if err != nil {
		t.Fatalf("ListMembersLowLevel: %v", err)
	}
	if len(list.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(list.Members))
	}
}

func TestListMembersLowLevelWithOptions(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("role_id") != "admin" {
			t.Errorf("expected role_id=admin, got %q", q.Get("role_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(groups.MemberList{Members: []groups.Member{}})
	})
	defer server.Close()

	_, err := client.ListMembersLowLevel(context.Background(), "grp-1", &groups.ListMembersOptions{
		RoleID:    "admin",
		Status:    "active",
		PageSize:  5,
		PageToken: "tok",
	})
	if err != nil {
		t.Fatalf("ListMembersLowLevel with options: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestAddMemberLowLevel
// ---------------------------------------------------------------------------

func TestAddMemberLowLevel(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.AddMemberLowLevel(context.Background(), "grp-1", "user-1", "admin")
	if err != nil {
		t.Fatalf("AddMemberLowLevel: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestRemoveMemberLowLevel
// ---------------------------------------------------------------------------

func TestRemoveMemberLowLevel(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/groups/grp-1/members/user-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.RemoveMemberLowLevel(context.Background(), "grp-1", "user-1")
	if err != nil {
		t.Fatalf("RemoveMemberLowLevel: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestUpdateMemberRoleLowLevel
// ---------------------------------------------------------------------------

func TestUpdateMemberRoleLowLevel(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.UpdateMemberRoleLowLevel(context.Background(), "grp-1", "user-1", "member")
	if err != nil {
		t.Fatalf("UpdateMemberRoleLowLevel: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestHTTPError – verify parseGroupsError paths
// ---------------------------------------------------------------------------

func TestHTTPErrorWithCode(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Access denied",
			"code":  "FORBIDDEN",
		})
	})
	defer server.Close()

	_, err := client.GetGroup(context.Background(), "grp-1")
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
}

func TestHTTPErrorWithoutCode(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Group not found",
		})
	})
	defer server.Close()

	_, err := client.GetGroup(context.Background(), "grp-1")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestHTTPErrorFallback(t *testing.T) {
	server, client := withMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		// Non-JSON body to test fallback path
		w.Write([]byte("internal server error"))
	})
	defer server.Close()

	_, err := client.GetGroup(context.Background(), "grp-1")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package groups

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	now := time.Now()
	g := Group{
		DATA_TYPE:   "group",
		ID:          "group-id",
		Name:        "Test Group",
		Description: "desc",
		Created:     now,
		LastUpdated: now,
	}
	if g.ID != "group-id" || g.Name != "Test Group" {
		t.Errorf("Group fields not set: %+v", g)
	}
}

func TestGroupCreate(t *testing.T) {
	gc := GroupCreate{DATA_TYPE: "group", Name: "New", Description: "d", PublicGroup: true}
	if gc.Name != "New" || !gc.PublicGroup {
		t.Errorf("GroupCreate fields: %+v", gc)
	}
}

func TestGroupList(t *testing.T) {
	gl := GroupList{Groups: []Group{{ID: "group1"}, {ID: "group2"}}}
	if len(gl.Groups) != 2 || gl.Groups[0].ID != "group1" {
		t.Errorf("GroupList: %+v", gl)
	}
}

func TestMember(t *testing.T) {
	m := Member{IdentityID: "id", Username: "u", Email: "e@x.org", Status: "active", Role: RoleManager}
	if m.IdentityID != "id" || m.Role != RoleManager {
		t.Errorf("Member: %+v", m)
	}
}

// TestGroupPoliciesShape verifies the policy payload uses the upstream keys.
func TestGroupPoliciesShape(t *testing.T) {
	timeout := 28800
	p := GroupPolicies{
		IsHighAssurance:                true,
		GroupVisibility:                "private",
		GroupMembersVisibility:         "managers",
		JoinRequests:                   false,
		SignupFields:                   []string{"institution"},
		AuthenticationAssuranceTimeout: &timeout,
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, k := range []string{"is_high_assurance", "group_visibility", "group_members_visibility", "join_requests", "signup_fields", "authentication_assurance_timeout"} {
		if _, ok := m[k]; !ok {
			t.Errorf("GroupPolicies JSON missing key %q; got %s", k, b)
		}
	}
	if _, ok := m["is_high_risk_group"]; ok {
		t.Error("GroupPolicies JSON has fabricated key is_high_risk_group")
	}
}

// TestBatchMembershipActionsShape verifies action keys serialize correctly.
func TestBatchMembershipActionsShape(t *testing.T) {
	a := BatchMembershipActions{
		Add:    []MemberWithRole{{IdentityID: "i1", Role: RoleMember}},
		Remove: []MemberID{{IdentityID: "i2"}},
	}
	b, _ := json.Marshal(a)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(b, &m)
	if _, ok := m["add"]; !ok {
		t.Error("missing add key")
	}
	if _, ok := m["remove"]; !ok {
		t.Error("missing remove key")
	}
	if _, ok := m["accept"]; ok {
		t.Error("empty accept should be omitted")
	}
}

func TestProvisionRule(t *testing.T) {
	pr := ProvisionRule{ID: "r", Label: "l", Expression: "x", MappedRoleID: "role-id"}
	if pr.MappedRoleID != "role-id" {
		t.Errorf("ProvisionRule: %+v", pr)
	}
}

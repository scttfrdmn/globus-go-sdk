// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
	// Test creation of a Group struct with all fields
	now := time.Now()
	group := Group{
		ID:                    "test-group-id",
		Name:                  "Test Group",
		Description:           "A test group",
		ParentID:              "parent-group-id",
		IdentityID:            "identity-id",
		MemberCount:           5,
		IsGroupAdmin:          true,
		IsMember:              true,
		Created:               now,
		LastUpdated:           now,
		PublicGroup:           true,
		RequiresSignAgreement: false,
		SignAgreementMessage:  "",
		Policies: map[string]interface{}{
			"policy1": "value1",
		},
		EnforceProvisionRules: true,
		ProvisionRules: []ProvisionRule{
			{
				ID:           "rule-id",
				Label:        "Test Rule",
				Expression:   "user.email.endsWith('@example.com')",
				MappedRoleID: "role-id",
			},
		},
	}

	// Check that fields are set correctly
	if group.ID != "test-group-id" {
		t.Errorf("Group.ID = %v, want %v", group.ID, "test-group-id")
	}
	if group.Name != "Test Group" {
		t.Errorf("Group.Name = %v, want %v", group.Name, "Test Group")
	}
	if group.MemberCount != 5 {
		t.Errorf("Group.MemberCount = %v, want %v", group.MemberCount, 5)
	}
	if !group.IsGroupAdmin {
		t.Errorf("Group.IsGroupAdmin = %v, want %v", group.IsGroupAdmin, true)
	}
	if !group.IsMember {
		t.Errorf("Group.IsMember = %v, want %v", group.IsMember, true)
	}
	if !group.Created.Equal(now) {
		t.Errorf("Group.Created = %v, want %v", group.Created, now)
	}
	if !group.LastUpdated.Equal(now) {
		t.Errorf("Group.LastUpdated = %v, want %v", group.LastUpdated, now)
	}
	if !group.PublicGroup {
		t.Errorf("Group.PublicGroup = %v, want %v", group.PublicGroup, true)
	}
	if group.RequiresSignAgreement {
		t.Errorf("Group.RequiresSignAgreement = %v, want %v", group.RequiresSignAgreement, false)
	}
	if value, ok := group.Policies["policy1"]; !ok || value.(string) != "value1" {
		t.Errorf("Group.Policies[\"policy1\"] = %v, want %v", value, "value1")
	}
	if !group.EnforceProvisionRules {
		t.Errorf("Group.EnforceProvisionRules = %v, want %v", group.EnforceProvisionRules, true)
	}
	if len(group.ProvisionRules) != 1 {
		t.Fatalf("len(Group.ProvisionRules) = %v, want %v", len(group.ProvisionRules), 1)
	}
	if group.ProvisionRules[0].ID != "rule-id" {
		t.Errorf("Group.ProvisionRules[0].ID = %v, want %v", group.ProvisionRules[0].ID, "rule-id")
	}
	if group.ProvisionRules[0].Expression != "user.email.endsWith('@example.com')" {
		t.Errorf("Group.ProvisionRules[0].Expression = %v, want %v", group.ProvisionRules[0].Expression, "user.email.endsWith('@example.com')")
	}
}

func TestGroupCreate(t *testing.T) {
	// Test creation of a GroupCreate struct
	groupCreate := GroupCreate{
		Name:                  "New Group",
		Description:           "A new group",
		ParentID:              "parent-group-id",
		PublicGroup:           true,
		RequiresSignAgreement: false,
		Policies: map[string]interface{}{
			"policy1": "value1",
		},
		EnforceProvisionRules: true,
	}

	// Check that fields are set correctly
	if groupCreate.Name != "New Group" {
		t.Errorf("GroupCreate.Name = %v, want %v", groupCreate.Name, "New Group")
	}
	if groupCreate.Description != "A new group" {
		t.Errorf("GroupCreate.Description = %v, want %v", groupCreate.Description, "A new group")
	}
	if groupCreate.ParentID != "parent-group-id" {
		t.Errorf("GroupCreate.ParentID = %v, want %v", groupCreate.ParentID, "parent-group-id")
	}
	if !groupCreate.PublicGroup {
		t.Errorf("GroupCreate.PublicGroup = %v, want %v", groupCreate.PublicGroup, true)
	}
	if groupCreate.RequiresSignAgreement {
		t.Errorf("GroupCreate.RequiresSignAgreement = %v, want %v", groupCreate.RequiresSignAgreement, false)
	}
	if value, ok := groupCreate.Policies["policy1"]; !ok || value.(string) != "value1" {
		t.Errorf("GroupCreate.Policies[\"policy1\"] = %v, want %v", value, "value1")
	}
	if !groupCreate.EnforceProvisionRules {
		t.Errorf("GroupCreate.EnforceProvisionRules = %v, want %v", groupCreate.EnforceProvisionRules, true)
	}
}

func TestGroupUpdate(t *testing.T) {
	// Test creation of a GroupUpdate struct with pointer fields
	publicGroupTrue := true
	requiresSignAgreementFalse := false
	enforceProvisionRulesTrue := true

	groupUpdate := GroupUpdate{
		Name:                  "Updated Group",
		Description:           "An updated group",
		ParentID:              "new-parent-id",
		PublicGroup:           &publicGroupTrue,
		RequiresSignAgreement: &requiresSignAgreementFalse,
		EnforceProvisionRules: &enforceProvisionRulesTrue,
		Policies: map[string]interface{}{
			"policy1": "new-value",
		},
	}

	// Check that fields are set correctly
	if groupUpdate.Name != "Updated Group" {
		t.Errorf("GroupUpdate.Name = %v, want %v", groupUpdate.Name, "Updated Group")
	}
	if groupUpdate.Description != "An updated group" {
		t.Errorf("GroupUpdate.Description = %v, want %v", groupUpdate.Description, "An updated group")
	}
	if groupUpdate.ParentID != "new-parent-id" {
		t.Errorf("GroupUpdate.ParentID = %v, want %v", groupUpdate.ParentID, "new-parent-id")
	}
	if *groupUpdate.PublicGroup != true {
		t.Errorf("*GroupUpdate.PublicGroup = %v, want %v", *groupUpdate.PublicGroup, true)
	}
	if *groupUpdate.RequiresSignAgreement != false {
		t.Errorf("*GroupUpdate.RequiresSignAgreement = %v, want %v", *groupUpdate.RequiresSignAgreement, false)
	}
	if *groupUpdate.EnforceProvisionRules != true {
		t.Errorf("*GroupUpdate.EnforceProvisionRules = %v, want %v", *groupUpdate.EnforceProvisionRules, true)
	}
	if value, ok := groupUpdate.Policies["policy1"]; !ok || value.(string) != "new-value" {
		t.Errorf("GroupUpdate.Policies[\"policy1\"] = %v, want %v", value, "new-value")
	}
}

func TestGroupList(t *testing.T) {
	// Test creation of a GroupList struct
	now := time.Now()

	groups := []Group{
		{
			ID:          "group1",
			Name:        "Group 1",
			Description: "First group",
			Created:     now,
			LastUpdated: now,
		},
		{
			ID:          "group2",
			Name:        "Group 2",
			Description: "Second group",
			Created:     now,
			LastUpdated: now,
		},
	}

	groupList := GroupList{
		Groups:        groups,
		HasNextPage:   true,
		NextPageToken: "next-page-token",
	}

	// Check that fields are set correctly
	if len(groupList.Groups) != 2 {
		t.Fatalf("len(GroupList.Groups) = %v, want %v", len(groupList.Groups), 2)
	}
	if groupList.Groups[0].ID != "group1" {
		t.Errorf("GroupList.Groups[0].ID = %v, want %v", groupList.Groups[0].ID, "group1")
	}
	if groupList.Groups[1].ID != "group2" {
		t.Errorf("GroupList.Groups[1].ID = %v, want %v", groupList.Groups[1].ID, "group2")
	}
	if !groupList.HasNextPage {
		t.Errorf("GroupList.HasNextPage = %v, want %v", groupList.HasNextPage, true)
	}
	if groupList.NextPageToken != "next-page-token" {
		t.Errorf("GroupList.NextPageToken = %v, want %v", groupList.NextPageToken, "next-page-token")
	}
}

func TestMember(t *testing.T) {
	// Test creation of a Member struct
	now := time.Now()

	member := Member{
		IdentityID:        "identity-id",
		Username:          "username",
		Email:             "user@example.com",
		Status:            "active",
		RoleID:            "role-id",
		Name:              "User Name",
		Organization:      "Example Org",
		JoinedDate:        now,
		LastUpdateDate:    now,
		ProvisionedByRule: "rule-id",
		Role: Role{
			ID:          "role-id",
			Name:        "admin",
			Description: "Administrator role",
		},
	}

	// Check that fields are set correctly
	if member.IdentityID != "identity-id" {
		t.Errorf("Member.IdentityID = %v, want %v", member.IdentityID, "identity-id")
	}
	if member.Username != "username" {
		t.Errorf("Member.Username = %v, want %v", member.Username, "username")
	}
	if member.Email != "user@example.com" {
		t.Errorf("Member.Email = %v, want %v", member.Email, "user@example.com")
	}
	if member.Status != "active" {
		t.Errorf("Member.Status = %v, want %v", member.Status, "active")
	}
	if member.RoleID != "role-id" {
		t.Errorf("Member.RoleID = %v, want %v", member.RoleID, "role-id")
	}
	if member.Name != "User Name" {
		t.Errorf("Member.Name = %v, want %v", member.Name, "User Name")
	}
	if member.Organization != "Example Org" {
		t.Errorf("Member.Organization = %v, want %v", member.Organization, "Example Org")
	}
	if !member.JoinedDate.Equal(now) {
		t.Errorf("Member.JoinedDate = %v, want %v", member.JoinedDate, now)
	}
	if !member.LastUpdateDate.Equal(now) {
		t.Errorf("Member.LastUpdateDate = %v, want %v", member.LastUpdateDate, now)
	}
	if member.ProvisionedByRule != "rule-id" {
		t.Errorf("Member.ProvisionedByRule = %v, want %v", member.ProvisionedByRule, "rule-id")
	}
	if member.Role.ID != "role-id" {
		t.Errorf("Member.Role.ID = %v, want %v", member.Role.ID, "role-id")
	}
	if member.Role.Name != "admin" {
		t.Errorf("Member.Role.Name = %v, want %v", member.Role.Name, "admin")
	}
}

func TestRole(t *testing.T) {
	// Test creation of a Role struct
	role := Role{
		ID:          "role-id",
		Name:        "admin",
		Description: "Administrator role",
	}

	// Check that fields are set correctly
	if role.ID != "role-id" {
		t.Errorf("Role.ID = %v, want %v", role.ID, "role-id")
	}
	if role.Name != "admin" {
		t.Errorf("Role.Name = %v, want %v", role.Name, "admin")
	}
	if role.Description != "Administrator role" {
		t.Errorf("Role.Description = %v, want %v", role.Description, "Administrator role")
	}
}

func TestMemberList(t *testing.T) {
	// Test creation of a MemberList struct
	members := []Member{
		{
			IdentityID: "member1",
			Username:   "user1",
			Email:      "user1@example.com",
		},
		{
			IdentityID: "member2",
			Username:   "user2",
			Email:      "user2@example.com",
		},
	}

	memberList := MemberList{
		Members:       members,
		HasNextPage:   true,
		NextPageToken: "next-page-token",
	}

	// Check that fields are set correctly
	if len(memberList.Members) != 2 {
		t.Fatalf("len(MemberList.Members) = %v, want %v", len(memberList.Members), 2)
	}
	if memberList.Members[0].IdentityID != "member1" {
		t.Errorf("MemberList.Members[0].IdentityID = %v, want %v", memberList.Members[0].IdentityID, "member1")
	}
	if memberList.Members[1].IdentityID != "member2" {
		t.Errorf("MemberList.Members[1].IdentityID = %v, want %v", memberList.Members[1].IdentityID, "member2")
	}
	if !memberList.HasNextPage {
		t.Errorf("MemberList.HasNextPage = %v, want %v", memberList.HasNextPage, true)
	}
	if memberList.NextPageToken != "next-page-token" {
		t.Errorf("MemberList.NextPageToken = %v, want %v", memberList.NextPageToken, "next-page-token")
	}
}

func TestMemberInvite(t *testing.T) {
	// Test creation of a MemberInvite struct
	invite := MemberInvite{
		Email:   "user@example.com",
		RoleID:  "role-id",
		Message: "Please join our group",
	}

	// Check that fields are set correctly
	if invite.Email != "user@example.com" {
		t.Errorf("MemberInvite.Email = %v, want %v", invite.Email, "user@example.com")
	}
	if invite.RoleID != "role-id" {
		t.Errorf("MemberInvite.RoleID = %v, want %v", invite.RoleID, "role-id")
	}
	if invite.Message != "Please join our group" {
		t.Errorf("MemberInvite.Message = %v, want %v", invite.Message, "Please join our group")
	}
}

func TestProvisionRule(t *testing.T) {
	// Test creation of a ProvisionRule struct
	rule := ProvisionRule{
		ID:           "rule-id",
		Label:        "Test Rule",
		Expression:   "user.email.endsWith('@example.com')",
		MappedRoleID: "role-id",
	}

	// Check that fields are set correctly
	if rule.ID != "rule-id" {
		t.Errorf("ProvisionRule.ID = %v, want %v", rule.ID, "rule-id")
	}
	if rule.Label != "Test Rule" {
		t.Errorf("ProvisionRule.Label = %v, want %v", rule.Label, "Test Rule")
	}
	if rule.Expression != "user.email.endsWith('@example.com')" {
		t.Errorf("ProvisionRule.Expression = %v, want %v", rule.Expression, "user.email.endsWith('@example.com')")
	}
	if rule.MappedRoleID != "role-id" {
		t.Errorf("ProvisionRule.MappedRoleID = %v, want %v", rule.MappedRoleID, "role-id")
	}
}

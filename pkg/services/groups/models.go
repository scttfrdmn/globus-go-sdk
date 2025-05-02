// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"time"
)

// Group represents a Globus group
type Group struct {
	DATA_TYPE             string    `json:"DATA_TYPE"`
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Description           string    `json:"description"`
	ParentID              string    `json:"parent_id,omitempty"`
	IdentityID            string    `json:"identity_id"`
	MemberCount           int       `json:"member_count"`
	IsGroupAdmin          bool      `json:"is_group_admin"`
	IsMember              bool      `json:"is_member"`
	Created               time.Time `json:"created"`
	LastUpdated           time.Time `json:"last_updated"`
	PublicGroup           bool      `json:"public_group"`
	RequiresSignAgreement bool      `json:"requires_sign_agreement"`
	SignAgreementMessage  string    `json:"sign_agreement_message,omitempty"`
	// Additional fields
	Policies              map[string]interface{} `json:"policies,omitempty"`
	EnforceProvisionRules bool              `json:"enforce_provision_rules,omitempty"`
	ProvisionRules        []ProvisionRule   `json:"provision_rules,omitempty"`
}

// ProvisionRule represents a rule for provisioning group membership
type ProvisionRule struct {
	DATA_TYPE    string `json:"DATA_TYPE"`
	ID           string `json:"id"`
	Label        string `json:"label"`
	Expression   string `json:"expression"`
	MappedRoleID string `json:"mapped_role_id"`
}

// GroupCreate represents the data needed to create a new group
type GroupCreate struct {
	DATA_TYPE             string            `json:"DATA_TYPE"`
	Name                  string            `json:"name"`
	Description           string            `json:"description,omitempty"`
	ParentID              string            `json:"parent_id,omitempty"`
	PublicGroup           bool              `json:"public_group,omitempty"`
	RequiresSignAgreement bool              `json:"requires_sign_agreement,omitempty"`
	SignAgreementMessage  string            `json:"sign_agreement_message,omitempty"`
	EnforceProvisionRules bool              `json:"enforce_provision_rules,omitempty"`
	Policies              map[string]interface{} `json:"policies,omitempty"`
}

// GroupUpdate represents the data to update in a group
type GroupUpdate struct {
	DATA_TYPE             string            `json:"DATA_TYPE"`
	Name                  string            `json:"name,omitempty"`
	Description           string            `json:"description,omitempty"`
	ParentID              string            `json:"parent_id,omitempty"`
	PublicGroup           *bool             `json:"public_group,omitempty"`
	RequiresSignAgreement *bool             `json:"requires_sign_agreement,omitempty"`
	SignAgreementMessage  string            `json:"sign_agreement_message,omitempty"`
	EnforceProvisionRules *bool             `json:"enforce_provision_rules,omitempty"`
	Policies              map[string]interface{} `json:"policies,omitempty"`
}

// GroupList represents a paginated list of groups
type GroupList struct {
	Groups        []Group `json:"groups"`
	HasNextPage   bool    `json:"has_next_page"`
	NextPageToken string  `json:"next_page_token,omitempty"`
}

// ListGroupsOptions contains options for filtering group listings
type ListGroupsOptions struct {
	IncludeGroupMembership bool
	IncludeIdentitySet     bool
	ForUserID              string
	MyGroups               bool
	PageSize               int
	PageToken              string
}

// Member represents a group member
type Member struct {
	DATA_TYPE         string    `json:"DATA_TYPE"`
	IdentityID        string    `json:"identity_id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Status            string    `json:"status"`
	RoleID            string    `json:"role_id"`
	Role              Role      `json:"role"`
	// Additional fields
	Name              string    `json:"name,omitempty"`
	Organization      string    `json:"organization,omitempty"`
	JoinedDate        time.Time `json:"joined_date,omitempty"`
	LastUpdateDate    time.Time `json:"last_update_date,omitempty"`
	ProvisionedByRule string    `json:"provisioned_by_rule,omitempty"`
}

// Role represents a member's role in a group
type Role struct {
	DATA_TYPE   string `json:"DATA_TYPE"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleCreate represents the data needed to create a new role
type RoleCreate struct {
	DATA_TYPE   string `json:"DATA_TYPE"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RoleUpdate represents the data to update in a role
type RoleUpdate struct {
	DATA_TYPE   string `json:"DATA_TYPE"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// RoleList represents a list of roles
type RoleList struct {
	Roles []Role `json:"roles"`
}

// MemberList represents a paginated list of group members
type MemberList struct {
	Members       []Member `json:"members"`
	HasNextPage   bool     `json:"has_next_page"`
	NextPageToken string   `json:"next_page_token,omitempty"`
}

// ListMembersOptions contains options for filtering member listings
type ListMembersOptions struct {
	RoleID    string
	Status    string
	PageSize  int
	PageToken string
}

// MemberInvite represents an invitation to join a group
type MemberInvite struct {
	Email   string `json:"email"`
	RoleID  string `json:"role_id"`
	Message string `json:"message,omitempty"`
}

// MemberInviteResponse represents the response from inviting a member
type MemberInviteResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message,omitempty"`
	Email    string `json:"email"`
}

// ProvisionRuleCreate represents the data needed to create a provision rule
type ProvisionRuleCreate struct {
	Label        string `json:"label"`
	Expression   string `json:"expression"`
	MappedRoleID string `json:"mapped_role_id"`
}

// ProvisionRuleUpdate represents the data to update in a provision rule
type ProvisionRuleUpdate struct {
	Label        string `json:"label,omitempty"`
	Expression   string `json:"expression,omitempty"`
	MappedRoleID string `json:"mapped_role_id,omitempty"`
}

// GroupMembership represents the membership status in a group
type GroupMembership struct {
	IsMember bool   `json:"is_member"`
	RoleID   string `json:"role_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

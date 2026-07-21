// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups

import "time"

// Group represents a Globus group
type Group struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	ParentID              string                 `json:"parent_id,omitempty"`
	IdentityID            string                 `json:"identity_id"`
	MemberCount           int                    `json:"member_count"`
	IsGroupAdmin          bool                   `json:"is_group_admin"`
	IsMember              bool                   `json:"is_member"`
	Created               time.Time              `json:"created"`
	LastUpdated           time.Time              `json:"last_updated"`
	PublicGroup           bool                   `json:"public_group"`
	RequiresSignAgreement bool                   `json:"requires_sign_agreement"`
	SignAgreementMessage  string                 `json:"sign_agreement_message,omitempty"`
	Policies              map[string]interface{} `json:"policies,omitempty"`
	EnforceProvisionRules bool                   `json:"enforce_provision_rules,omitempty"`
	// Memberships is populated only when a group is fetched with
	// include=memberships; the Groups API has no separate members endpoint.
	Memberships []Member `json:"memberships,omitempty"`
	// MyMemberships holds the caller's own membership(s) in the group, populated
	// only with include=my_memberships. Each entry carries the caller's Role
	// (member/manager/admin) — the only way to distinguish manager from member,
	// since the Group-level IsGroupAdmin/IsMember bools cannot.
	MyMemberships []Member `json:"my_memberships,omitempty"`
}

// GroupCreate represents the data needed to create a new group
type GroupCreate struct {
	Name                  string                 `json:"name"`
	Description           string                 `json:"description,omitempty"`
	ParentID              string                 `json:"parent_id,omitempty"`
	PublicGroup           bool                   `json:"public_group,omitempty"`
	RequiresSignAgreement bool                   `json:"requires_sign_agreement,omitempty"`
	SignAgreementMessage  string                 `json:"sign_agreement_message,omitempty"`
	EnforceProvisionRules bool                   `json:"enforce_provision_rules,omitempty"`
	Policies              map[string]interface{} `json:"policies,omitempty"`
}

// GroupUpdate represents the data to update in a group
type GroupUpdate struct {
	Name                  string                 `json:"name,omitempty"`
	Description           string                 `json:"description,omitempty"`
	ParentID              string                 `json:"parent_id,omitempty"`
	PublicGroup           *bool                  `json:"public_group,omitempty"`
	RequiresSignAgreement *bool                  `json:"requires_sign_agreement,omitempty"`
	SignAgreementMessage  string                 `json:"sign_agreement_message,omitempty"`
	EnforceProvisionRules *bool                  `json:"enforce_provision_rules,omitempty"`
	Policies              map[string]interface{} `json:"policies,omitempty"`
}

// GetGroupOptions controls GetGroup. Include is comma-joined into a single
// `include` query param (allowed: memberships, my_memberships, policies,
// allowed_actions, child_ids).
type GetGroupOptions struct {
	Include []string
}

// GetMyGroupsOptions controls GetMyGroupsWithOptions. Statuses filters by
// membership status; Include is comma-joined into the `include` query param.
// Use include=my_memberships to populate each group's MyMemberships (and thus
// the caller's role).
type GetMyGroupsOptions struct {
	Statuses []string
	Include  []string
}

// Member represents a group membership entry, as embedded in a group document
// fetched with include=memberships. The Groups API has no standalone member
// resource; role is a string (member/manager/admin), not a role ID.
type Member struct {
	IdentityID     string    `json:"identity_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Status         string    `json:"status"`
	Role           string    `json:"role"`
	Name           string    `json:"name,omitempty"`
	Organization   string    `json:"organization,omitempty"`
	JoinedDate     time.Time `json:"joined_date,omitempty"`
	LastUpdateDate time.Time `json:"last_update_date,omitempty"`
}

// Group role constants. Role is a string attribute on a membership.
const (
	RoleMember  = "member"
	RoleManager = "manager"
	RoleAdmin   = "admin"
)

// MemberID identifies a member by identity for batch actions that need no role.
type MemberID struct {
	IdentityID string `json:"identity_id"`
}

// MemberWithRole identifies a member and a role for add/invite/change_role
// batch actions.
type MemberWithRole struct {
	IdentityID string `json:"identity_id"`
	Role       string `json:"role"`
}

// BatchMembershipActions is the request body for BatchMembershipAction. Each
// non-empty key names a membership action to apply in a single POST.
type BatchMembershipActions struct {
	Accept      []MemberID       `json:"accept,omitempty"`
	Add         []MemberWithRole `json:"add,omitempty"`
	Approve     []MemberID       `json:"approve,omitempty"`
	ChangeRole  []MemberWithRole `json:"change_role,omitempty"`
	Decline     []MemberID       `json:"decline,omitempty"`
	Invite      []MemberWithRole `json:"invite,omitempty"`
	Join        []MemberID       `json:"join,omitempty"`
	Leave       []MemberID       `json:"leave,omitempty"`
	Reject      []MemberID       `json:"reject,omitempty"`
	Remove      []MemberID       `json:"remove,omitempty"`
	RequestJoin []MemberID       `json:"request_join,omitempty"`
}

// GroupPolicies represents the policy settings for a group (GET/PUT
// /groups/{id}/policies). Fields match the Groups API exactly.
type GroupPolicies struct {
	IsHighAssurance                bool     `json:"is_high_assurance"`
	GroupVisibility                string   `json:"group_visibility"`         // authenticated | private
	GroupMembersVisibility         string   `json:"group_members_visibility"` // members | managers
	JoinRequests                   bool     `json:"join_requests"`
	SignupFields                   []string `json:"signup_fields"`
	AuthenticationAssuranceTimeout *int     `json:"authentication_assurance_timeout,omitempty"`
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Groups service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Groups client
// In v4, config is required and must include explicit scopes
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	// Set default Groups service URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://groups.api.globus.org/v2"
	}

	// Create base client
	baseClient, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    config.BaseURL,
	}, nil
}

// GetGroup retrieves a specific group by ID
// v4: Context is always first parameter
func (c *Client) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	if groupID == "" {
		return nil, &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}

	var group Group
	endpoint := fmt.Sprintf("/groups/%s", groupID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListGroups lists groups with optional filtering
// v4: Context is always first parameter
func (c *Client) ListGroups(ctx context.Context, options *ListGroupsOptions) (*GroupList, error) {
	query := url.Values{}

	if options != nil {
		if options.IncludeGroupMembership {
			query.Set("include_group_membership", "true")
		}
		if options.IncludeIdentitySet {
			query.Set("include_identity_set", "true")
		}
		if options.ForUserID != "" {
			query.Set("for_user_id", options.ForUserID)
		}
		if options.MyGroups {
			query.Set("my_groups", "true")
		}
		if len(options.Statuses) > 0 {
			for _, status := range options.Statuses {
				query.Add("statuses", status)
			}
		}
		if options.PageSize > 0 {
			query.Set("per_page", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var groupList GroupList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/groups", query, nil, &groupList)
	if err != nil {
		return nil, err
	}

	return &groupList, nil
}

// CreateGroup creates a new group
// v4: Context is always first parameter
func (c *Client) CreateGroup(ctx context.Context, group *GroupCreate) (*Group, error) {
	if group == nil {
		return nil, &core.ValidationError{
			Field:   "group",
			Message: "group data is required",
		}
	}
	if group.Name == "" {
		return nil, &core.ValidationError{
			Field:   "Name",
			Message: "group name is required",
		}
	}

	var result Group
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/groups", nil, group, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroup updates an existing group
// v4: Context is always first parameter
func (c *Client) UpdateGroup(ctx context.Context, groupID string, update *GroupUpdate) (*Group, error) {
	if groupID == "" {
		return nil, &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}
	if update == nil {
		return nil, &core.ValidationError{
			Field:   "update",
			Message: "update data is required",
		}
	}

	var result Group
	endpoint := fmt.Sprintf("/groups/%s", groupID)
	err := c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroup deletes a group
// v4: Context is always first parameter
func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}

	endpoint := fmt.Sprintf("/groups/%s", groupID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// ListMembers lists members of a group
// v4: Context is always first parameter
func (c *Client) ListMembers(ctx context.Context, groupID string, options *ListMembersOptions) (*MemberList, error) {
	if groupID == "" {
		return nil, &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}

	query := url.Values{}
	if options != nil {
		if options.RoleID != "" {
			query.Set("role_id", options.RoleID)
		}
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.PageSize > 0 {
			query.Set("per_page", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var memberList MemberList
	endpoint := fmt.Sprintf("/groups/%s/members", groupID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, query, nil, &memberList)
	if err != nil {
		return nil, err
	}

	return &memberList, nil
}

// AddMember adds a member to a group
// v4: Context is always first parameter
func (c *Client) AddMember(ctx context.Context, groupID, identityID, roleID string) error {
	if groupID == "" {
		return &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}
	if identityID == "" {
		return &core.ValidationError{
			Field:   "identityID",
			Message: "identity ID is required",
		}
	}
	if roleID == "" {
		return &core.ValidationError{
			Field:   "roleID",
			Message: "role ID is required",
		}
	}

	body := map[string]interface{}{
		"identity_id": identityID,
		"role_id":     roleID,
	}

	endpoint := fmt.Sprintf("/groups/%s/members", groupID)
	return c.baseClient.DoRequest(ctx, http.MethodPost, endpoint, nil, body, nil)
}

// RemoveMember removes a member from a group
// v4: Context is always first parameter
func (c *Client) RemoveMember(ctx context.Context, groupID, identityID string) error {
	if groupID == "" {
		return &core.ValidationError{
			Field:   "groupID",
			Message: "group ID is required",
		}
	}
	if identityID == "" {
		return &core.ValidationError{
			Field:   "identityID",
			Message: "identity ID is required",
		}
	}

	endpoint := fmt.Sprintf("/groups/%s/members/%s", groupID, identityID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}
// UpdateMemberRole changes the role of an existing member.
func (c *Client) UpdateMemberRole(ctx context.Context, groupID, identityID, roleID string) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if identityID == "" {
		return &core.ValidationError{Field: "identityID", Message: "identity ID is required"}
	}
	if roleID == "" {
		return &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}

	body := map[string]interface{}{"role_id": roleID}
	endpoint := fmt.Sprintf("/groups/%s/members/%s", groupID, identityID)
	return c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, body, nil)
}

// GetGroupPolicies retrieves the policy settings for a group.
func (c *Client) GetGroupPolicies(ctx context.Context, groupID string) (map[string]interface{}, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}

	var policies map[string]interface{}
	endpoint := fmt.Sprintf("/groups/%s/policies", groupID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// SetGroupPolicies replaces the policy settings for a group.
func (c *Client) SetGroupPolicies(ctx context.Context, groupID string, policies map[string]interface{}) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if policies == nil {
		return &core.ValidationError{Field: "policies", Message: "policies are required"}
	}

	endpoint := fmt.Sprintf("/groups/%s/policies", groupID)
	return c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, policies, nil)
}

// GetMyGroups lists groups the current user is a member of.
func (c *Client) GetMyGroups(ctx context.Context, statuses []string) (*GroupList, error) {
	return c.ListGroups(ctx, &ListGroupsOptions{MyGroups: true, Statuses: statuses})
}

// ListRoles lists all roles for a group.
func (c *Client) ListRoles(ctx context.Context, groupID string) (*RoleList, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	var roleList RoleList
	endpoint := fmt.Sprintf("/groups/%s/roles", groupID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &roleList); err != nil {
		return nil, err
	}
	return &roleList, nil
}

// GetRole retrieves a specific role in a group.
func (c *Client) GetRole(ctx context.Context, groupID, roleID string) (*Role, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	var role Role
	endpoint := fmt.Sprintf("/groups/%s/roles/%s", groupID, roleID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole creates a new role in a group.
func (c *Client) CreateRole(ctx context.Context, groupID string, role *RoleCreate) (*Role, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if role == nil {
		return nil, &core.ValidationError{Field: "role", Message: "role data is required"}
	}
	var result Role
	endpoint := fmt.Sprintf("/groups/%s/roles", groupID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, endpoint, nil, role, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRole updates a role in a group.
func (c *Client) UpdateRole(ctx context.Context, groupID, roleID string, update *RoleUpdate) (*Role, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}
	var result Role
	endpoint := fmt.Sprintf("/groups/%s/roles/%s", groupID, roleID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRole deletes a role from a group.
func (c *Client) DeleteRole(ctx context.Context, groupID, roleID string) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if roleID == "" {
		return &core.ValidationError{Field: "roleID", Message: "role ID is required"}
	}
	endpoint := fmt.Sprintf("/groups/%s/roles/%s", groupID, roleID)
	return c.baseClient.DoRequest(ctx, http.MethodDelete, endpoint, nil, nil, nil)
}

// GetIdentityPreferences retrieves identity preferences for a member in a group.
func (c *Client) GetIdentityPreferences(ctx context.Context, groupID, identityID string) (*IdentityPreferences, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if identityID == "" {
		return nil, &core.ValidationError{Field: "identityID", Message: "identity ID is required"}
	}
	var prefs IdentityPreferences
	endpoint := fmt.Sprintf("/groups/%s/identity_preferences/%s", groupID, identityID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &prefs); err != nil {
		return nil, err
	}
	return &prefs, nil
}

// SetIdentityPreferences sets identity preferences for a member in a group.
func (c *Client) SetIdentityPreferences(ctx context.Context, groupID, identityID string, preferences *IdentityPreferences) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if identityID == "" {
		return &core.ValidationError{Field: "identityID", Message: "identity ID is required"}
	}
	if preferences == nil {
		return &core.ValidationError{Field: "preferences", Message: "preferences are required"}
	}
	endpoint := fmt.Sprintf("/groups/%s/identity_preferences/%s", groupID, identityID)
	return c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, preferences, nil)
}

// GetMembershipFields retrieves custom membership fields for a group.
func (c *Client) GetMembershipFields(ctx context.Context, groupID string) (*MembershipFields, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	var fields MembershipFields
	endpoint := fmt.Sprintf("/groups/%s/membership_fields", groupID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, nil, nil, &fields); err != nil {
		return nil, err
	}
	return &fields, nil
}

// SetMembershipFields sets custom membership fields for a group.
func (c *Client) SetMembershipFields(ctx context.Context, groupID string, fields *MembershipFields) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if fields == nil {
		return &core.ValidationError{Field: "fields", Message: "fields are required"}
	}
	endpoint := fmt.Sprintf("/groups/%s/membership_fields", groupID)
	return c.baseClient.DoRequest(ctx, http.MethodPut, endpoint, nil, fields, nil)
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}


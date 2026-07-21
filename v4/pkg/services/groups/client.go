// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

// GetGroup retrieves a specific group by ID (GET /groups/{id}). Pass opts.Include
// (e.g. "memberships", "policies") to expand the returned document; opts may be nil.
// v4: Context is always first parameter
func (c *Client) GetGroup(ctx context.Context, groupID string, opts *GetGroupOptions) (*Group, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}

	query := url.Values{}
	if opts != nil && len(opts.Include) > 0 {
		query.Set("include", strings.Join(opts.Include, ","))
	}

	var group Group
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/groups/%s", groupID), query, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// GetMyGroups lists the groups the caller belongs to (GET /groups/my_groups).
// statuses is comma-joined into a single query param (e.g. active, invited); pass
// nil for all. The response is a top-level JSON array.
func (c *Client) GetMyGroups(ctx context.Context, statuses []string) ([]Group, error) {
	query := url.Values{}
	if len(statuses) > 0 {
		query.Set("statuses", strings.Join(statuses, ","))
	}

	var out []Group
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/groups/my_groups", query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMyGroupsWithOptions lists the caller's groups (GET /groups/my_groups) with
// optional status filtering and includes. Pass Include=["my_memberships"] to
// populate each returned group's MyMemberships, which carries the caller's role
// (member/manager/admin) — not derivable from the Group-level bool fields.
func (c *Client) GetMyGroupsWithOptions(ctx context.Context, opts *GetMyGroupsOptions) ([]Group, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.Statuses) > 0 {
			query.Set("statuses", strings.Join(opts.Statuses, ","))
		}
		if len(opts.Include) > 0 {
			query.Set("include", strings.Join(opts.Include, ","))
		}
	}

	var out []Group
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/groups/my_groups", query, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupBySubscriptionID looks up the group associated with a subscription
// (GET /subscription_info/{subscription_id}).
func (c *Client) GetGroupBySubscriptionID(ctx context.Context, subscriptionID string) (*Group, error) {
	if subscriptionID == "" {
		return nil, &core.ValidationError{Field: "subscriptionID", Message: "subscription ID is required"}
	}
	var group Group
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/subscription_info/%s", subscriptionID), nil, nil, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// CreateGroup creates a new group (POST /groups).
// v4: Context is always first parameter
func (c *Client) CreateGroup(ctx context.Context, group *GroupCreate) (*Group, error) {
	if group == nil {
		return nil, &core.ValidationError{Field: "group", Message: "group data is required"}
	}
	if group.Name == "" {
		return nil, &core.ValidationError{Field: "Name", Message: "group name is required"}
	}

	var result Group
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/groups", nil, group, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateGroup updates an existing group (PUT /groups/{id}).
// v4: Context is always first parameter
func (c *Client) UpdateGroup(ctx context.Context, groupID string, update *GroupUpdate) (*Group, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	var result Group
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/groups/%s", groupID), nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteGroup deletes a group (DELETE /groups/{id}).
// v4: Context is always first parameter
func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/groups/%s", groupID), nil, nil, nil)
}

// BatchMembershipAction applies one or more membership actions in a single call
// (POST /groups/{id}). This is the sole membership-mutation endpoint: add,
// invite, accept, approve, change_role, decline, join, leave, reject, remove, and
// request_join are all expressed here.
func (c *Client) BatchMembershipAction(ctx context.Context, groupID string, actions *BatchMembershipActions) (*Group, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if actions == nil {
		return nil, &core.ValidationError{Field: "actions", Message: "actions are required"}
	}

	var result Group
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/groups/%s", groupID), nil, actions, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGroupPolicies retrieves the policy settings for a group
// (GET /groups/{id}/policies).
func (c *Client) GetGroupPolicies(ctx context.Context, groupID string) (*GroupPolicies, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	var policies GroupPolicies
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/groups/%s/policies", groupID), nil, nil, &policies); err != nil {
		return nil, err
	}
	return &policies, nil
}

// SetGroupPolicies replaces the policy settings for a group
// (PUT /groups/{id}/policies).
func (c *Client) SetGroupPolicies(ctx context.Context, groupID string, policies *GroupPolicies) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if policies == nil {
		return &core.ValidationError{Field: "policies", Message: "policies are required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/groups/%s/policies", groupID), nil, policies, nil)
}

// SetSubscriptionAdminVerified sets (or clears) the subscription that
// admin-verifies a group (PUT /groups/{id}/subscription_admin_verified). Pass a
// nil subscriptionID to disassociate (sends a JSON null).
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID string, subscriptionID *string) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	body := map[string]interface{}{"subscription_admin_verified_id": subscriptionID}
	return c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/groups/%s/subscription_admin_verified", groupID), nil, body, nil)
}

// GetIdentityPreferences retrieves the caller's Groups preferences
// (GET /preferences). Preferences are account-level, not group-scoped.
func (c *Client) GetIdentityPreferences(ctx context.Context) (map[string]interface{}, error) {
	var prefs map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/preferences", nil, nil, &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// SetIdentityPreferences updates the caller's Groups preferences
// (PUT /preferences), e.g. {"allow_add": false}.
func (c *Client) SetIdentityPreferences(ctx context.Context, prefs map[string]interface{}) error {
	if prefs == nil {
		return &core.ValidationError{Field: "prefs", Message: "preferences are required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodPut, "/preferences", nil, prefs, nil)
}

// GetMembershipFields retrieves the caller's membership field values for a group
// (GET /groups/{id}/membership_fields).
func (c *Client) GetMembershipFields(ctx context.Context, groupID string) (map[string]interface{}, error) {
	if groupID == "" {
		return nil, &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	var fields map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/groups/%s/membership_fields", groupID), nil, nil, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

// SetMembershipFields sets the caller's membership field values for a group
// (PUT /groups/{id}/membership_fields).
func (c *Client) SetMembershipFields(ctx context.Context, groupID string, fields map[string]interface{}) error {
	if groupID == "" {
		return &core.ValidationError{Field: "groupID", Message: "group ID is required"}
	}
	if fields == nil {
		return &core.ValidationError{Field: "fields", Message: "fields are required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/groups/%s/membership_fields", groupID), nil, fields, nil)
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListMembersLowLevel retrieves members of a group using low-level API
// This is an internal method.
// Most users should use ListMembers instead.
func (c *Client) ListMembersLowLevel(ctx context.Context, groupID string, options *ListMembersOptions) (*MemberList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.RoleID != "" {
			query.Set("role_id", options.RoleID)
		}
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.PageSize > 0 {
			query.Set("per_page", fmt.Sprintf("%d", options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	var memberList MemberList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "groups/"+groupID+"/members", query, nil, &memberList)
	if err != nil {
		return nil, err
	}

	return &memberList, nil
}

// AddMemberLowLevel adds a user to a group using low-level API
// This is an internal method.
// Most users should use AddMember instead.
func (c *Client) AddMemberLowLevel(ctx context.Context, groupID, userID string, roleID string) error {
	// Build the request body
	body := map[string]string{
		"identity_id": userID,
		"role_id":     roleID,
	}

	return c.doRequestLowLevel(ctx, http.MethodPost, "groups/"+groupID+"/members", nil, body, nil)
}

// RemoveMemberLowLevel removes a user from a group using low-level API
// This is an internal method.
// Most users should use RemoveMember instead.
func (c *Client) RemoveMemberLowLevel(ctx context.Context, groupID, userID string) error {
	return c.doRequestLowLevel(ctx, http.MethodDelete, "groups/"+groupID+"/members/"+userID, nil, nil, nil)
}

// UpdateMemberRoleLowLevel updates a member's role in a group using low-level API
// This is an internal method.
// Most users should use UpdateMemberRole instead.
func (c *Client) UpdateMemberRoleLowLevel(ctx context.Context, groupID, userID, roleID string) error {
	// Build the request body
	body := map[string]string{
		"role_id": roleID,
	}

	return c.doRequestLowLevel(ctx, http.MethodPatch, "groups/"+groupID+"/members/"+userID, nil, body, nil)
}

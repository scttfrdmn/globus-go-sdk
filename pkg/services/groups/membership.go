// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"net/url"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// ListMembers retrieves members of a group
func (c *Client) ListMembers(ctx context.Context, groupID string, options *ListMembersOptions) (*MemberList, error) {
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
			query.Set("per_page", string(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("marker", options.PageToken)
		}
	}

	// Make the request
	resp, err := c.Transport.Get(ctx, "groups/"+groupID+"/members", query, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var memberList MemberList
	if err := transport.DecodeResponse(resp, &memberList); err != nil {
		return nil, err
	}

	return &memberList, nil
}

// AddMember adds a user to a group
func (c *Client) AddMember(ctx context.Context, groupID, userID string, roleID string) error {
	// Build the request body
	body := map[string]string{
		"identity_id": userID,
		"role_id":     roleID,
	}

	// Make the request
	resp, err := c.Transport.Post(ctx, "groups/"+groupID+"/members", body, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// RemoveMember removes a user from a group
func (c *Client) RemoveMember(ctx context.Context, groupID, userID string) error {
	// Make the request
	resp, err := c.Transport.Delete(ctx, "groups/"+groupID+"/members/"+userID, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// UpdateMemberRole updates a member's role in a group
func (c *Client) UpdateMemberRole(ctx context.Context, groupID, userID, roleID string) error {
	// Build the request body
	body := map[string]string{
		"role_id": roleID,
	}

	// Make the request
	resp, err := c.Transport.Patch(ctx, "groups/"+groupID+"/members/"+userID, body, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

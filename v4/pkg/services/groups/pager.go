// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewGroupsPager returns a Paginator that iterates through all groups
// matching opts. Pass nil for default options.
func (c *Client) NewGroupsPager(opts *ListGroupsOptions) paging.Paginator[Group] {
	pageSize := 0
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}
	return paging.NewNextTokenPaginator(
		func(ctx context.Context, pageSize int, pageToken string) ([]Group, bool, string, error) {
			o := &ListGroupsOptions{PageSize: pageSize, PageToken: pageToken}
			if opts != nil {
				o.IncludeGroupMembership = opts.IncludeGroupMembership
				o.IncludeIdentitySet = opts.IncludeIdentitySet
				o.ForUserID = opts.ForUserID
				o.MyGroups = opts.MyGroups
				o.Statuses = opts.Statuses
			}
			result, err := c.ListGroups(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Groups, result.HasNextPage, result.NextPageToken, nil
		},
		pageSize,
	)
}

// NewMembersPager returns a Paginator that iterates through all members of
// the given group matching opts. Pass nil for default options.
func (c *Client) NewMembersPager(groupID string, opts *ListMembersOptions) paging.Paginator[Member] {
	pageSize := 0
	if opts != nil && opts.PageSize > 0 {
		pageSize = opts.PageSize
	}
	return paging.NewNextTokenPaginator(
		func(ctx context.Context, ps int, pageToken string) ([]Member, bool, string, error) {
			o := &ListMembersOptions{PageSize: ps, PageToken: pageToken}
			if opts != nil {
				o.RoleID = opts.RoleID
				o.Status = opts.Status
			}
			result, err := c.ListMembers(ctx, groupID, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Members, result.HasNextPage, result.NextPageToken, nil
		},
		pageSize,
	)
}

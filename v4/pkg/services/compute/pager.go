// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewEndpointsPager returns a Paginator that iterates through all compute
// endpoints matching opts. Pass nil for default options.
func (c *Client) NewEndpointsPager(opts *ListEndpointsOptions) paging.Paginator[Endpoint] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]Endpoint, int, error) {
			o := &ListEndpointsOptions{Limit: limit, Offset: offset}
			result, err := c.ListEndpoints(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Endpoints, result.Total, nil
		},
		pageSize,
	)
}

// NewTasksPager returns a Paginator that iterates through all compute tasks
// matching opts. Pass nil for default options.
func (c *Client) NewTasksPager(opts *ListTasksOptions) paging.Paginator[TaskStatus] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]TaskStatus, int, error) {
			o := &ListTasksOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.EndpointID = opts.EndpointID
			}
			result, err := c.ListTasks(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Tasks, result.Total, nil
		},
		pageSize,
	)
}

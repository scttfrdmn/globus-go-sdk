// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewEndpointsPager returns a Paginator that iterates through all endpoints
// matching opts. Pass nil for default options.
func (c *Client) NewEndpointsPager(opts *ListEndpointsOptions) paging.Paginator[Endpoint] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]Endpoint, int, error) {
			o := &ListEndpointsOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.Filter = opts.Filter
			}
			result, err := c.ListEndpoints(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Data, result.Total, nil
		},
		pageSize,
	)
}

// NewTasksPager returns a Paginator that iterates through all tasks
// matching opts. Pass nil for default options.
func (c *Client) NewTasksPager(opts *ListTasksOptions) paging.Paginator[Task] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]Task, int, error) {
			o := &ListTasksOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.Filter = opts.Filter
				o.FilterStatus = opts.FilterStatus
			}
			result, err := c.ListTasks(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Data, result.Total, nil
		},
		pageSize,
	)
}

// NewTunnelsPager returns a Paginator that iterates through all tunnels
// matching opts. Pass nil for default options.
func (c *Client) NewTunnelsPager(opts *ListTunnelsOptions) paging.Paginator[Tunnel] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]Tunnel, bool, string, error) {
			o := &ListTunnelsOptions{Limit: limit, Marker: marker}
			result, err := c.ListTunnels(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Tunnels, result.HasMore, result.Marker, nil
		},
		pageSize,
	)
}

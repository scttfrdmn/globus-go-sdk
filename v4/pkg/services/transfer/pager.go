// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package transfer

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// endpointSearchMaxResults matches the upstream endpoint_search cap.
const endpointSearchMaxResults = 1000

// NewEndpointSearchPager returns a Paginator over endpoint_search results,
// advancing offset until has_next_page is false or the 1000-result cap is hit.
func (c *Client) NewEndpointSearchPager(opts *EndpointSearchOptions) paging.Paginator[Endpoint] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	offset := 0
	if opts != nil {
		offset = opts.Offset
	}
	fetched := 0
	return paging.NewNextTokenPaginator(
		func(ctx context.Context, limit int, _ string) ([]Endpoint, bool, string, error) {
			o := EndpointSearchOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.FilterFulltext = opts.FilterFulltext
				o.FilterScope = opts.FilterScope
				o.FilterOwnerID = opts.FilterOwnerID
				o.FilterHostEndpoint = opts.FilterHostEndpoint
				o.FilterNonFunctional = opts.FilterNonFunctional
				o.FilterEntityType = opts.FilterEntityType
			}
			result, err := c.EndpointSearch(ctx, &o)
			if err != nil {
				return nil, false, "", err
			}
			offset += len(result.Data)
			fetched += len(result.Data)
			hasNext := result.HasNextPage && fetched < endpointSearchMaxResults
			return result.Data, hasNext, "", nil
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
				o.OrderBy = opts.OrderBy
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

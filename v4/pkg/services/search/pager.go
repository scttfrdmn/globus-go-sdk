// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewIndexesPager returns a Paginator that iterates through all search indexes
// matching opts. Pass nil for default options.
func (c *Client) NewIndexesPager(opts *ListIndexesOptions) paging.Paginator[Index] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]Index, int, error) {
			o := &ListIndexesOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.FilterRoles = opts.FilterRoles
			}
			result, err := c.IndexList(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Indexes, result.Total, nil
		},
		pageSize,
	)
}

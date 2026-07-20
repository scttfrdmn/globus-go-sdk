// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package search

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// searchMaxResults matches the upstream HasNextPaginator cap for search results.
const searchMaxResults = 10000

// NewSearchPager returns a Paginator over gmeta results for a POST search,
// advancing offset until has_next_page is false or the 10000-result cap is hit.
func (c *Client) NewSearchPager(indexID string, query *SearchQuery) paging.Paginator[GMetaResult] {
	pageSize := 100
	if query != nil && query.Limit > 0 {
		pageSize = query.Limit
	}
	offset := 0
	if query != nil {
		offset = query.Offset
	}
	fetched := 0
	return paging.NewNextTokenPaginator(
		func(ctx context.Context, ps int, _ string) ([]GMetaResult, bool, string, error) {
			q := &SearchQuery{}
			if query != nil {
				*q = *query
			}
			q.Limit = ps
			q.Offset = offset
			res, err := c.Search(ctx, indexID, q)
			if err != nil {
				return nil, false, "", err
			}
			offset += len(res.GMeta)
			fetched += len(res.GMeta)
			hasNext := res.HasNextPage && fetched < searchMaxResults
			return res.GMeta, hasNext, "", nil
		},
		pageSize,
	)
}

// NewScrollPager returns a marker-paginated Paginator over a scroll query's
// gmeta results, feeding each response marker into the next request.
func (c *Client) NewScrollPager(indexID string, query *ScrollQuery) paging.Paginator[GMetaResult] {
	pageSize := 0
	if query != nil && query.Limit > 0 {
		pageSize = query.Limit
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, _ int, marker string) ([]GMetaResult, bool, string, error) {
			q := &ScrollQuery{}
			if query != nil {
				*q = *query
			}
			if marker != "" {
				q.Marker = marker
			}
			res, err := c.Scroll(ctx, indexID, q)
			if err != nil {
				return nil, false, "", err
			}
			hasMore := res.HasNextPage && res.Marker != ""
			return res.GMeta, hasMore, res.Marker, nil
		},
		pageSize,
	)
}

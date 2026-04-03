// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewFlowsPager returns a Paginator that iterates through all flows
// matching opts. Pass nil for default options.
func (c *Client) NewFlowsPager(opts *ListFlowsOptions) paging.Paginator[Flow] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]Flow, int, error) {
			o := &ListFlowsOptions{Limit: limit, Offset: offset}
			result, err := c.ListFlows(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Flows, result.Total, nil
		},
		pageSize,
	)
}

// NewRunsPager returns a Paginator that iterates through all flow runs
// matching opts. Pass nil for default options.
func (c *Client) NewRunsPager(opts *ListRunsOptions) paging.Paginator[FlowRun] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewLimitOffsetPaginator(
		func(ctx context.Context, limit, offset int) ([]FlowRun, int, error) {
			o := &ListRunsOptions{Limit: limit, Offset: offset}
			if opts != nil {
				o.FlowID = opts.FlowID
			}
			result, err := c.ListRuns(ctx, o)
			if err != nil {
				return nil, 0, err
			}
			return result.Runs, result.Total, nil
		},
		pageSize,
	)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewFlowsPager returns a marker Paginator that iterates through all flows
// matching opts. Pass nil for default options.
func (c *Client) NewFlowsPager(opts *ListFlowsOptions) paging.Paginator[Flow] {
	return paging.NewMarkerPaginator(
		func(ctx context.Context, _ int, marker string) ([]Flow, bool, string, error) {
			o := &ListFlowsOptions{Marker: marker}
			if opts != nil {
				o.FilterRoles = opts.FilterRoles
				o.FilterFulltext = opts.FilterFulltext
				o.OrderBy = opts.OrderBy
			}
			result, err := c.ListFlows(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Flows, result.HasNextPage, result.Marker, nil
		},
		0,
	)
}

// NewRegisteredAPIsPager returns a Paginator that iterates through all
// registered APIs matching opts. Pass nil for default options.
// Uses marker pagination (upstream v4.6.0).
func (c *Client) NewRegisteredAPIsPager(opts *ListRegisteredAPIsOptions) paging.Paginator[RegisteredAPI] {
	pageSize := 0
	if opts != nil && opts.PerPage > 0 {
		pageSize = opts.PerPage
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]RegisteredAPI, bool, string, error) {
			o := &ListRegisteredAPIsOptions{Marker: marker, PerPage: limit}
			if opts != nil {
				o.FilterRoles = opts.FilterRoles
				o.OrderBy = opts.OrderBy
			}
			result, err := c.ListRegisteredAPIs(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.RegisteredAPIs, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}

// NewRunsPager returns a marker Paginator that iterates through all flow runs
// matching opts. Pass nil for default options.
func (c *Client) NewRunsPager(opts *ListRunsOptions) paging.Paginator[FlowRun] {
	return paging.NewMarkerPaginator(
		func(ctx context.Context, _ int, marker string) ([]FlowRun, bool, string, error) {
			o := &ListRunsOptions{Marker: marker}
			if opts != nil {
				o.FilterFlowID = opts.FilterFlowID
				o.FilterRoles = opts.FilterRoles
			}
			result, err := c.ListRuns(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Runs, result.HasNextPage, result.Marker, nil
		},
		0,
	)
}

// NewRunLogsPager returns a marker Paginator over a run's log entries.
func (c *Client) NewRunLogsPager(runID string, opts *ListRunLogsOptions) paging.Paginator[RunLog] {
	pageSize := 0
	if opts != nil && opts.Limit > 0 {
		pageSize = opts.Limit
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]RunLog, bool, string, error) {
			o := &ListRunLogsOptions{Marker: marker, Limit: limit}
			if opts != nil {
				o.ReverseOrder = opts.ReverseOrder
			}
			result, err := c.GetRunLogs(ctx, runID, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Entries, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}

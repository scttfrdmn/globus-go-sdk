// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Flows service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Flows client
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://flows.globus.org"
	}

	baseClient, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    config.BaseURL,
	}, nil
}

// GetFlow retrieves a flow by ID
func (c *Client) GetFlow(ctx context.Context, flowID string) (*Flow, error) {
	if flowID == "" {
		return nil, &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}

	var flow Flow
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/flows/%s", flowID), nil, nil, &flow)
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// ListFlows lists flows with optional filtering
func (c *Client) ListFlows(ctx context.Context, options *ListFlowsOptions) (*FlowList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var flowList FlowList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/flows", query, nil, &flowList)
	if err != nil {
		return nil, err
	}
	return &flowList, nil
}

// RunFlow starts a flow execution
func (c *Client) RunFlow(ctx context.Context, flowID string, input *FlowInput) (*FlowRun, error) {
	if flowID == "" {
		return nil, &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	if input == nil {
		return nil, &core.ValidationError{Field: "input", Message: "input is required"}
	}

	var run FlowRun
	path := fmt.Sprintf("/flows/%s/run", flowID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, input, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetRun retrieves a flow run by ID
func (c *Client) GetRun(ctx context.Context, runID string) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	var run FlowRun
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s", runID), nil, nil, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// CancelRun cancels a running flow
func (c *Client) CancelRun(ctx context.Context, runID string) error {
	if runID == "" {
		return &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/runs/%s/cancel", runID), nil, nil, nil)
}

// ListRuns lists flow runs
func (c *Client) ListRuns(ctx context.Context, options *ListRunsOptions) (*RunList, error) {
	query := url.Values{}
	if options != nil {
		if options.FlowID != "" {
			query.Set("flow_id", options.FlowID)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var runList RunList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/runs", query, nil, &runList)
	if err != nil {
		return nil, err
	}
	return &runList, nil
}
// CreateFlow deploys a new flow definition.
func (c *Client) CreateFlow(ctx context.Context, flow *FlowCreate) (*Flow, error) {
	if flow == nil {
		return nil, &core.ValidationError{Field: "flow", Message: "flow data is required"}
	}
	if flow.Title == "" {
		return nil, &core.ValidationError{Field: "Title", Message: "flow title is required"}
	}
	if flow.Definition == nil {
		return nil, &core.ValidationError{Field: "Definition", Message: "flow definition is required"}
	}

	var result Flow
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/flows", nil, flow, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFlow modifies an existing flow.
func (c *Client) UpdateFlow(ctx context.Context, flowID string, update *FlowUpdate) (*Flow, error) {
	if flowID == "" {
		return nil, &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	var result Flow
	if err := c.baseClient.DoRequest(ctx, http.MethodPatch, fmt.Sprintf("/flows/%s", flowID), nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFlow removes a flow definition.
func (c *Client) DeleteFlow(ctx context.Context, flowID string) error {
	if flowID == "" {
		return &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/flows/%s", flowID), nil, nil, nil)
}

// UpdateRun modifies metadata on an active or completed run.
func (c *Client) UpdateRun(ctx context.Context, runID string, update *RunUpdate) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	var result FlowRun
	if err := c.baseClient.DoRequest(ctx, http.MethodPatch, fmt.Sprintf("/runs/%s", runID), nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRunLogs retrieves log entries for a flow run.
func (c *Client) GetRunLogs(ctx context.Context, runID string, options *ListRunLogsOptions) (*RunLogList, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var logList RunLogList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s/log", runID), query, nil, &logList); err != nil {
		return nil, err
	}
	return &logList, nil
}

var terminalRunStatuses = map[string]bool{
	"SUCCEEDED": true,
	"FAILED":    true,
	"CANCELLED": true,
}

// WaitForRun polls GetRun until terminal state or ctx cancellation.
func (c *Client) WaitForRun(ctx context.Context, runID string, pollInterval time.Duration) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	for {
		run, err := c.GetRun(ctx, runID)
		if err != nil {
			return nil, err
		}
		if terminalRunStatuses[run.Status] {
			return run, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// ListActionProviders lists all Flows action providers.
func (c *Client) ListActionProviders(ctx context.Context, options *ListActionProvidersOptions) (*ActionProviderList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.OrderBy != "" {
			query.Set("orderby", options.OrderBy)
		}
		if options.Q != "" {
			query.Set("q", options.Q)
		}
		if options.FilterOwner != "" {
			query.Set("filter_owner", options.FilterOwner)
		}
		if options.FilterType != "" {
			query.Set("filter_type", options.FilterType)
		}
	}
	var list ActionProviderList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/action_providers", query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetActionProvider retrieves a specific action provider by ID.
func (c *Client) GetActionProvider(ctx context.Context, providerID string) (*ActionProvider, error) {
	if providerID == "" {
		return nil, &core.ValidationError{Field: "providerID", Message: "action provider ID is required"}
	}
	var provider ActionProvider
	path := fmt.Sprintf("/action_providers/%s", providerID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

// ListActionRoles lists all roles for an action provider.
func (c *Client) ListActionRoles(ctx context.Context, providerID string, limit, offset int) (*ActionRoleList, error) {
	if providerID == "" {
		return nil, &core.ValidationError{Field: "providerID", Message: "action provider ID is required"}
	}
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	var list ActionRoleList
	path := fmt.Sprintf("/action_providers/%s/roles", providerID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// GetActionRole retrieves a specific role for an action provider.
func (c *Client) GetActionRole(ctx context.Context, providerID, roleID string) (*ActionRole, error) {
	if providerID == "" {
		return nil, &core.ValidationError{Field: "providerID", Message: "action provider ID is required"}
	}
	if roleID == "" {
		return nil, &core.ValidationError{Field: "roleID", Message: "action role ID is required"}
	}
	var role ActionRole
	path := fmt.Sprintf("/action_providers/%s/roles/%s", providerID, roleID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &role); err != nil {
		return nil, err
	}
	return &role, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}


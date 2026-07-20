// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// ListFlows lists flows with optional filtering (GET /flows). Marker-paginated.
func (c *Client) ListFlows(ctx context.Context, options *ListFlowsOptions) (*FlowList, error) {
	query := url.Values{}
	if options != nil {
		if len(options.FilterRoles) > 0 {
			query.Set("filter_roles", strings.Join(options.FilterRoles, ","))
		}
		if options.FilterFulltext != "" {
			query.Set("filter_fulltext", options.FilterFulltext)
		}
		for _, o := range options.OrderBy {
			query.Add("orderby", o)
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}

	var flowList FlowList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/flows", query, nil, &flowList)
	if err != nil {
		return nil, err
	}
	return &flowList, nil
}

// RunFlow starts a flow execution (POST /flows/{id}/run).
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

// ValidateRun validates a run request without starting it
// (POST /flows/{id}/validate_run). The body envelope matches RunFlow.
func (c *Client) ValidateRun(ctx context.Context, flowID string, input *FlowInput) (map[string]interface{}, error) {
	if flowID == "" {
		return nil, &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	if input == nil {
		return nil, &core.ValidationError{Field: "input", Message: "input is required"}
	}

	var result map[string]interface{}
	path := fmt.Sprintf("/flows/%s/validate_run", flowID)
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, input, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetRun retrieves a flow run by ID (GET /runs/{id}). Pass opts to request the
// flow_description; opts may be nil.
func (c *Client) GetRun(ctx context.Context, runID string, options *GetRunOptions) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	query := url.Values{}
	if options != nil && options.IncludeFlowDescription {
		query.Set("include_flow_description", "true")
	}

	var run FlowRun
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s", runID), query, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// GetRunDefinition retrieves the flow definition a run was started with
// (GET /runs/{id}/definition).
func (c *Client) GetRunDefinition(ctx context.Context, runID string) (*RunDefinition, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	var def RunDefinition
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s/definition", runID), nil, nil, &def); err != nil {
		return nil, err
	}
	return &def, nil
}

// CancelRun cancels a running flow (POST /runs/{id}/cancel).
func (c *Client) CancelRun(ctx context.Context, runID string) error {
	if runID == "" {
		return &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/runs/%s/cancel", runID), nil, nil, nil)
}

// DeleteRun deletes (releases) a run (POST /runs/{id}/release). Note: this is a
// POST to /release, not an HTTP DELETE.
func (c *Client) DeleteRun(ctx context.Context, runID string) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	var run FlowRun
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/runs/%s/release", runID), nil, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// ResumeRun resumes a run that is awaiting resume (POST /runs/{id}/resume).
func (c *Client) ResumeRun(ctx context.Context, runID string) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	var run FlowRun
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/runs/%s/resume", runID), nil, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// ListRuns lists flow runs (GET /runs). Marker-paginated.
func (c *Client) ListRuns(ctx context.Context, options *ListRunsOptions) (*RunList, error) {
	query := url.Values{}
	if options != nil {
		if len(options.FilterFlowID) > 0 {
			query.Set("filter_flow_id", strings.Join(options.FilterFlowID, ","))
		}
		if len(options.FilterRoles) > 0 {
			query.Set("filter_roles", strings.Join(options.FilterRoles, ","))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}

	var runList RunList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/runs", query, nil, &runList)
	if err != nil {
		return nil, err
	}
	return &runList, nil
}

// CreateFlow deploys a new flow definition (POST /flows).
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
	if flow.InputSchema == nil {
		return nil, &core.ValidationError{Field: "InputSchema", Message: "flow input schema is required"}
	}

	var result Flow
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/flows", nil, flow, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ValidateFlow validates a flow definition without deploying it
// (POST /flows/validate).
func (c *Client) ValidateFlow(ctx context.Context, definition, inputSchema map[string]interface{}) (map[string]interface{}, error) {
	if definition == nil {
		return nil, &core.ValidationError{Field: "definition", Message: "flow definition is required"}
	}

	body := map[string]interface{}{"definition": definition}
	if inputSchema != nil {
		body["input_schema"] = inputSchema
	}

	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/flows/validate", nil, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateFlow modifies an existing flow (PUT /flows/{id}).
func (c *Client) UpdateFlow(ctx context.Context, flowID string, update *FlowUpdate) (*Flow, error) {
	if flowID == "" {
		return nil, &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	var result Flow
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/flows/%s", flowID), nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFlow removes a flow definition (DELETE /flows/{id}).
func (c *Client) DeleteFlow(ctx context.Context, flowID string) error {
	if flowID == "" {
		return &core.ValidationError{Field: "flowID", Message: "flow ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/flows/%s", flowID), nil, nil, nil)
}

// UpdateRun modifies metadata on an active or completed run (PUT /runs/{id}).
func (c *Client) UpdateRun(ctx context.Context, runID string, update *RunUpdate) (*FlowRun, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update data is required"}
	}

	var result FlowRun
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/runs/%s", runID), nil, update, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetRunLogs retrieves log entries for a flow run (GET /runs/{id}/log).
// Marker-paginated.
func (c *Client) GetRunLogs(ctx context.Context, runID string, options *ListRunLogsOptions) (*RunLogList, error) {
	if runID == "" {
		return nil, &core.ValidationError{Field: "runID", Message: "run ID is required"}
	}

	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.ReverseOrder != nil {
			query.Set("reverse_order", strconv.FormatBool(*options.ReverseOrder))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
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
		run, err := c.GetRun(ctx, runID, nil)
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

// GetRegisteredAPI retrieves a registered API by ID.
// Added in Python SDK v4.6.0 (GET /registered_apis/{id}).
func (c *Client) GetRegisteredAPI(ctx context.Context, registeredAPIID string) (*RegisteredAPI, error) {
	if registeredAPIID == "" {
		return nil, &core.ValidationError{Field: "registeredAPIID", Message: "registered API ID is required"}
	}

	var api RegisteredAPI
	path := fmt.Sprintf("/registered_apis/%s", registeredAPIID)
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &api); err != nil {
		return nil, err
	}
	return &api, nil
}

// ListRegisteredAPIs lists registered APIs with optional filtering.
// Added in Python SDK v4.6.0; per_page added in v4.7.0
// (GET /registered_apis, marker pagination).
func (c *Client) ListRegisteredAPIs(ctx context.Context, options *ListRegisteredAPIsOptions) (*RegisteredAPIList, error) {
	query := url.Values{}
	if options != nil {
		if len(options.FilterRoles) > 0 {
			query.Set("filter_roles", strings.Join(options.FilterRoles, ","))
		}
		for _, o := range options.OrderBy {
			query.Add("orderby", o)
		}
		if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
	}

	var list RegisteredAPIList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/registered_apis", query, nil, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

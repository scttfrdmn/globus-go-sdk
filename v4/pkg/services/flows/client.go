// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}


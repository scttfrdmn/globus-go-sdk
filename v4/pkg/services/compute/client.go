// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Compute service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Compute client
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://compute.api.globus.org/v2"
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

// GetEndpoint retrieves a compute endpoint by ID
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}

	var endpoint Endpoint
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/endpoints/%s", endpointID), nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

// ListEndpoints lists compute endpoints
func (c *Client) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var endpointList EndpointList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/endpoints", query, nil, &endpointList)
	if err != nil {
		return nil, err
	}
	return &endpointList, nil
}

// SubmitFunction submits a function for execution
func (c *Client) SubmitFunction(ctx context.Context, endpointID string, submission *FunctionSubmission) (*FunctionRun, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if submission == nil {
		return nil, &core.ValidationError{Field: "submission", Message: "submission is required"}
	}

	var run FunctionRun
	path := fmt.Sprintf("/endpoints/%s/functions", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, submission, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetFunction retrieves a function run by ID
func (c *Client) GetFunction(ctx context.Context, functionID string) (*FunctionRun, error) {
	if functionID == "" {
		return nil, &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}

	var run FunctionRun
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/functions/%s", functionID), nil, nil, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// CancelFunction cancels a running function
func (c *Client) CancelFunction(ctx context.Context, functionID string) error {
	if functionID == "" {
		return &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/functions/%s/cancel", functionID), nil, nil, nil)
}

// ListFunctions lists function runs
func (c *Client) ListFunctions(ctx context.Context, options *ListFunctionsOptions) (*FunctionList, error) {
	query := url.Values{}
	if options != nil {
		if options.EndpointID != "" {
			query.Set("endpoint_id", options.EndpointID)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var functionList FunctionList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/functions", query, nil, &functionList)
	if err != nil {
		return nil, err
	}
	return &functionList, nil
}

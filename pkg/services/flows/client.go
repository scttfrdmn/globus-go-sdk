// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package flows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/yourusername/globus-go-sdk/pkg/core"
	"github.com/yourusername/globus-go-sdk/pkg/core/authorizers"
)

// Constants for Globus Flows
const (
	DefaultBaseURL = "https://flows.globus.org/v1/"
	FlowsScope     = "https://auth.globus.org/scopes/eec9b274-0c81-4334-bdc2-54e90e689b9a/manage_flows"
)

// Client provides methods for interacting with Globus Flows
type Client struct {
	Client *core.Client
}

// NewClient creates a new Flows client
func NewClient(accessToken string, options ...core.ClientOption) *Client {
	// Create the authorizer with the access token
	authorizer := authorizers.NewStaticTokenAuthorizer(accessToken)
	
	// Apply default options specific to Flows
	defaultOptions := []core.ClientOption{
		core.WithBaseURL(DefaultBaseURL),
		core.WithAuthorizer(authorizer),
	}
	
	// Merge with user options
	options = append(defaultOptions, options...)
	
	// Create the base client
	baseClient := core.NewClient(options...)
	
	return &Client{
		Client: baseClient,
	}
}

// buildURL builds a URL for the flows API
func (c *Client) buildURL(path string, query url.Values) string {
	baseURL := c.Client.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	
	url := baseURL + path
	if query != nil && len(query) > 0 {
		url += "?" + query.Encode()
	}
	
	return url
}

// doRequest performs an HTTP request and decodes the JSON response
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	url := c.buildURL(path, query)
	
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// For non-GET requests with no response body, just check status
	if method != http.MethodGet && response == nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	
	// Read and decode response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	if len(respBody) == 0 {
		return nil
	}
	
	if err := json.Unmarshal(respBody, response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return nil
}

// ListFlows lists all flows the user has access to
func (c *Client) ListFlows(ctx context.Context, options *ListFlowsOptions) (*FlowList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		} else if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
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
		if options.FilterRoles != "" {
			query.Set("filter_roles", options.FilterRoles)
		}
		if options.FilterOwner != "" {
			query.Set("filter_owner", options.FilterOwner)
		}
		if options.FilterPublic {
			query.Set("filter_public", "true")
		}
		if options.RolesOnly {
			query.Set("roles_only", "true")
		}
	}
	
	var flowList FlowList
	err := c.doRequest(ctx, http.MethodGet, "flows", query, nil, &flowList)
	if err != nil {
		return nil, err
	}
	
	return &flowList, nil
}

// GetFlow retrieves a specific flow by ID
func (c *Client) GetFlow(ctx context.Context, flowID string) (*Flow, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow ID is required")
	}
	
	var flow Flow
	err := c.doRequest(ctx, http.MethodGet, "flows/"+flowID, nil, nil, &flow)
	if err != nil {
		return nil, err
	}
	
	return &flow, nil
}

// CreateFlow creates a new flow
func (c *Client) CreateFlow(ctx context.Context, request *FlowCreateRequest) (*Flow, error) {
	if request == nil {
		return nil, fmt.Errorf("flow create request is required")
	}
	
	if request.Title == "" {
		return nil, fmt.Errorf("flow title is required")
	}
	
	if request.Definition == nil {
		return nil, fmt.Errorf("flow definition is required")
	}
	
	var flow Flow
	err := c.doRequest(ctx, http.MethodPost, "flows", nil, request, &flow)
	if err != nil {
		return nil, err
	}
	
	return &flow, nil
}

// UpdateFlow updates an existing flow
func (c *Client) UpdateFlow(ctx context.Context, flowID string, request *FlowUpdateRequest) (*Flow, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow ID is required")
	}
	
	if request == nil {
		return nil, fmt.Errorf("flow update request is required")
	}
	
	var flow Flow
	err := c.doRequest(ctx, http.MethodPut, "flows/"+flowID, nil, request, &flow)
	if err != nil {
		return nil, err
	}
	
	return &flow, nil
}

// DeleteFlow deletes a flow
func (c *Client) DeleteFlow(ctx context.Context, flowID string) error {
	if flowID == "" {
		return fmt.Errorf("flow ID is required")
	}
	
	return c.doRequest(ctx, http.MethodDelete, "flows/"+flowID, nil, nil, nil)
}

// RunFlow starts a new flow run
func (c *Client) RunFlow(ctx context.Context, request *RunRequest) (*RunResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("run request is required")
	}
	
	if request.FlowID == "" {
		return nil, fmt.Errorf("flow ID is required")
	}
	
	if request.Input == nil {
		return nil, fmt.Errorf("input is required")
	}
	
	var run RunResponse
	err := c.doRequest(ctx, http.MethodPost, "runs", nil, request, &run)
	if err != nil {
		return nil, err
	}
	
	return &run, nil
}

// ListRuns lists all flow runs the user has access to
func (c *Client) ListRuns(ctx context.Context, options *ListRunsOptions) (*RunList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		} else if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
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
		if options.FlowID != "" {
			query.Set("flow_id", options.FlowID)
		}
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.RoleType != "" {
			query.Set("role_type", options.RoleType)
		}
		if options.Label != "" {
			query.Set("label", options.Label)
		}
	}
	
	var runList RunList
	err := c.doRequest(ctx, http.MethodGet, "runs", query, nil, &runList)
	if err != nil {
		return nil, err
	}
	
	return &runList, nil
}

// GetRun retrieves a specific flow run by ID
func (c *Client) GetRun(ctx context.Context, runID string) (*RunResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}
	
	var run RunResponse
	err := c.doRequest(ctx, http.MethodGet, "runs/"+runID, nil, nil, &run)
	if err != nil {
		return nil, err
	}
	
	return &run, nil
}

// CancelRun cancels a flow run
func (c *Client) CancelRun(ctx context.Context, runID string) error {
	if runID == "" {
		return fmt.Errorf("run ID is required")
	}
	
	return c.doRequest(ctx, http.MethodPost, "runs/"+runID+"/cancel", nil, nil, nil)
}

// UpdateRun updates a flow run's metadata
func (c *Client) UpdateRun(ctx context.Context, runID string, request *RunUpdateRequest) (*RunResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}
	
	if request == nil {
		return nil, fmt.Errorf("run update request is required")
	}
	
	var run RunResponse
	err := c.doRequest(ctx, http.MethodPatch, "runs/"+runID, nil, request, &run)
	if err != nil {
		return nil, err
	}
	
	return &run, nil
}

// GetRunLogs retrieves logs for a specific run
func (c *Client) GetRunLogs(ctx context.Context, runID string, limit, offset int) (*RunLogList, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}
	
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	
	var logs RunLogList
	err := c.doRequest(ctx, http.MethodGet, "runs/"+runID+"/log", query, nil, &logs)
	if err != nil {
		return nil, err
	}
	
	return &logs, nil
}

// ListActionProviders lists all action providers
func (c *Client) ListActionProviders(ctx context.Context, options *ListActionProvidersOptions) (*ActionProviderList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		} else if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
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
		if options.FilterGlobus {
			query.Set("filter_globus", "true")
		}
	}
	
	var providerList ActionProviderList
	err := c.doRequest(ctx, http.MethodGet, "action_providers", query, nil, &providerList)
	if err != nil {
		return nil, err
	}
	
	return &providerList, nil
}

// GetActionProvider retrieves a specific action provider by ID
func (c *Client) GetActionProvider(ctx context.Context, providerID string) (*ActionProvider, error) {
	if providerID == "" {
		return nil, fmt.Errorf("action provider ID is required")
	}
	
	var provider ActionProvider
	err := c.doRequest(ctx, http.MethodGet, "action_providers/"+providerID, nil, nil, &provider)
	if err != nil {
		return nil, err
	}
	
	return &provider, nil
}

// ListActionRoles lists all action roles for a provider
func (c *Client) ListActionRoles(ctx context.Context, providerID string, limit, offset int) (*ActionRoleList, error) {
	if providerID == "" {
		return nil, fmt.Errorf("action provider ID is required")
	}
	
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	
	var roleList ActionRoleList
	err := c.doRequest(ctx, http.MethodGet, "action_providers/"+providerID+"/roles", query, nil, &roleList)
	if err != nil {
		return nil, err
	}
	
	return &roleList, nil
}

// GetActionRole retrieves a specific action role by ID
func (c *Client) GetActionRole(ctx context.Context, providerID, roleID string) (*ActionRole, error) {
	if providerID == "" {
		return nil, fmt.Errorf("action provider ID is required")
	}
	
	if roleID == "" {
		return nil, fmt.Errorf("action role ID is required")
	}
	
	var role ActionRole
	err := c.doRequest(ctx, http.MethodGet, "action_providers/"+providerID+"/roles/"+roleID, nil, nil, &role)
	if err != nil {
		return nil, err
	}
	
	return &role, nil
}
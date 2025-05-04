// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
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
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
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
func NewClient(opts ...ClientOption) (*Client, error) {
	// Apply default options
	options := defaultOptions()
	
	// Apply user options
	for _, opt := range opts {
		opt(options)
	}
	
	// If an access token was provided, create a static token authorizer
	if options.accessToken != "" {
		authorizer := authorizers.StaticTokenCoreAuthorizer(options.accessToken)
		options.coreOptions = append(options.coreOptions, core.WithAuthorizer(authorizer))
	}
	
	// Create the base client
	baseClient := core.NewClient(options.coreOptions...)
	
	return &Client{
		Client: baseClient,
	}, nil
}

// buildURLLowLevel builds a URL for the flows API
// This is an internal method used by the client.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
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

// doRequestLowLevel performs an HTTP request and decodes the JSON response
// This is an internal method used by higher-level API methods.
func (c *Client) doRequestLowLevel(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	url := c.buildURLLowLevel(path, query)

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

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error responses
	if resp.StatusCode >= 400 {
		// Extract resource ID and type for better error messages
		resourceID := ""
		resourceType := ""

		// Parse path to determine resource type and ID
		if path == "flows" && method == http.MethodPost {
			resourceType = "flow"
		} else if len(path) > 6 && path[:6] == "flows/" {
			resourceType = "flow"
			resourceID = path[6:]
			// Remove trailing components if any
			for i, c := range resourceID {
				if c == '/' {
					resourceID = resourceID[:i]
					break
				}
			}
		} else if path == "runs" && method == http.MethodPost {
			resourceType = "run"
		} else if len(path) > 5 && path[:5] == "runs/" {
			resourceType = "run"
			resourceID = path[5:]
			// Remove trailing components if any
			for i, c := range resourceID {
				if c == '/' {
					resourceID = resourceID[:i]
					break
				}
			}
		} else if path == "action_providers" {
			resourceType = "action_provider"
		} else if len(path) > 17 && path[:17] == "action_providers/" {
			resourceType = "action_provider"
			resourceID = path[17:]
			// Check if this is an action role request
			for i, c := range resourceID {
				if c == '/' {
					if len(resourceID) > i+7 && resourceID[i+1:i+7] == "roles/" {
						resourceType = "action_role"
						roleID := resourceID[i+7:]
						providerID := resourceID[:i]
						resourceID = providerID + ":" + roleID
					}
					break
				}
			}
		}

		// Print debug information during tests
		fmt.Printf("Error response status: %d, body: %s, resourceID: %s, resourceType: %s\n", 
			resp.StatusCode, string(respBody), resourceID, resourceType)

		return ParseErrorResponse(respBody, resp.StatusCode, resourceID, resourceType)
	}

	// For non-GET requests with no response body, just return nil
	if method != http.MethodGet && response == nil {
		return nil
	}

	// If there's no content, just return nil
	if len(respBody) == 0 {
		return nil
	}

	// Unmarshal response
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "flows", query, nil, &flowList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "flows/"+flowID, nil, nil, &flow)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "flows", nil, request, &flow)
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
	err := c.doRequestLowLevel(ctx, http.MethodPut, "flows/"+flowID, nil, request, &flow)
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

	return c.doRequestLowLevel(ctx, http.MethodDelete, "flows/"+flowID, nil, nil, nil)
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
	err := c.doRequestLowLevel(ctx, http.MethodPost, "runs", nil, request, &run)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "runs", query, nil, &runList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "runs/"+runID, nil, nil, &run)
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

	return c.doRequestLowLevel(ctx, http.MethodPost, "runs/"+runID+"/cancel", nil, nil, nil)
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
	err := c.doRequestLowLevel(ctx, http.MethodPatch, "runs/"+runID, nil, request, &run)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "runs/"+runID+"/log", query, nil, &logs)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "action_providers", query, nil, &providerList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "action_providers/"+providerID, nil, nil, &provider)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "action_providers/"+providerID+"/roles", query, nil, &roleList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "action_providers/"+providerID+"/roles/"+roleID, nil, nil, &role)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// Iterators for pagination

// GetFlowsIterator returns an iterator for listing flows with pagination.
func (c *Client) GetFlowsIterator(options *ListFlowsOptions) *FlowIterator {
	return NewFlowIterator(c, options)
}

// GetRunsIterator returns an iterator for listing flow runs with pagination.
func (c *Client) GetRunsIterator(options *ListRunsOptions) *RunIterator {
	return NewRunIterator(c, options)
}

// GetActionProvidersIterator returns an iterator for listing action providers with pagination.
func (c *Client) GetActionProvidersIterator(options *ListActionProvidersOptions) *ActionProviderIterator {
	return NewActionProviderIterator(c, options)
}

// GetActionRolesIterator returns an iterator for listing action roles with pagination.
func (c *Client) GetActionRolesIterator(providerID string, limit int) *ActionRoleIterator {
	return NewActionRoleIterator(c, providerID, limit)
}

// GetRunLogsIterator returns an iterator for listing run logs with pagination.
func (c *Client) GetRunLogsIterator(runID string, limit int) *RunLogIterator {
	return NewRunLogIterator(c, runID, limit)
}

// Batch operations

// WaitForRun waits for a flow run to complete or reach a terminal state.
// It returns the final run state or an error if the context is canceled or the polling fails.
func (c *Client) WaitForRun(ctx context.Context, runID string, pollInterval time.Duration) (*RunResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	if pollInterval <= 0 {
		pollInterval = time.Second * 3
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			run, err := c.GetRun(ctx, runID)
			if err != nil {
				return nil, err
			}

			// Check if run is in a terminal state
			switch run.Status {
			case "SUCCEEDED", "FAILED", "CANCELLED":
				return run, nil
			}
		}
	}
}

// ListAllFlows lists all flows using pagination, collecting all results.
// This is a convenience method that uses the FlowIterator internally.
func (c *Client) ListAllFlows(ctx context.Context, options *ListFlowsOptions) ([]Flow, error) {
	iterator := c.GetFlowsIterator(options)
	var flows []Flow

	for iterator.Next(ctx) {
		flows = append(flows, *iterator.Flow())
	}

	if err := iterator.Err(); err != nil {
		return nil, err
	}

	return flows, nil
}

// ListAllRuns lists all runs using pagination, collecting all results.
// This is a convenience method that uses the RunIterator internally.
func (c *Client) ListAllRuns(ctx context.Context, options *ListRunsOptions) ([]RunResponse, error) {
	iterator := c.GetRunsIterator(options)
	var runs []RunResponse

	for iterator.Next(ctx) {
		runs = append(runs, *iterator.Run())
	}

	if err := iterator.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

// ListAllActionProviders lists all action providers using pagination, collecting all results.
// This is a convenience method that uses the ActionProviderIterator internally.
func (c *Client) ListAllActionProviders(ctx context.Context, options *ListActionProvidersOptions) ([]ActionProvider, error) {
	iterator := c.GetActionProvidersIterator(options)
	var providers []ActionProvider

	for iterator.Next(ctx) {
		providers = append(providers, *iterator.ActionProvider())
	}

	if err := iterator.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

// ListAllActionRoles lists all action roles for a provider using pagination, collecting all results.
// This is a convenience method that uses the ActionRoleIterator internally.
func (c *Client) ListAllActionRoles(ctx context.Context, providerID string) ([]ActionRole, error) {
	iterator := c.GetActionRolesIterator(providerID, 100)
	var roles []ActionRole

	for iterator.Next(ctx) {
		roles = append(roles, *iterator.ActionRole())
	}

	if err := iterator.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// ListAllRunLogs lists all logs for a run using pagination, collecting all results.
// This is a convenience method that uses the RunLogIterator internally.
func (c *Client) ListAllRunLogs(ctx context.Context, runID string) ([]RunLogEntry, error) {
	iterator := c.GetRunLogsIterator(runID, 100)
	var entries []RunLogEntry

	for iterator.Next(ctx) {
		entries = append(entries, *iterator.LogEntry())
	}

	if err := iterator.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

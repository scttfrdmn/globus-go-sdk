// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
)

// Constants for Globus Compute
const (
	DefaultBaseURL = "https://compute.api.globus.org/v2/"
	ComputeScope   = "https://auth.globus.org/scopes/facd7ccc-c5f4-42aa-916b-a0e270e2c2a9/all"
)

// Client provides methods for interacting with Globus Compute
type Client struct {
	Client *core.Client
}

// NewClient creates a new Compute client
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

// buildURL builds a URL for the compute API
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

// ListEndpoints lists all compute endpoints the user has access to
func (c *Client) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*ComputeEndpointList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.OrderBy != "" {
			query.Set("orderby", options.OrderBy)
		}
		if options.Search != "" {
			query.Set("search", options.Search)
		}
		if options.FilterScope != "" {
			query.Set("filter_scope", options.FilterScope)
		}
		if options.FilterStatus != "" {
			query.Set("filter_status", options.FilterStatus)
		}
		if options.IncludeInfo {
			query.Set("include_info", "true")
		}
	}

	var endpointList ComputeEndpointList
	err := c.doRequest(ctx, http.MethodGet, "endpoints", query, nil, &endpointList)
	if err != nil {
		return nil, err
	}

	return &endpointList, nil
}

// GetEndpoint retrieves a specific compute endpoint by ID
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*ComputeEndpoint, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var endpoint ComputeEndpoint
	err := c.doRequest(ctx, http.MethodGet, "endpoints/"+endpointID, nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// RegisterFunction registers a new function with Globus Compute
func (c *Client) RegisterFunction(ctx context.Context, request *FunctionRegisterRequest) (*FunctionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("function register request is required")
	}

	if request.Function == "" {
		return nil, fmt.Errorf("function code is required")
	}

	var function FunctionResponse
	err := c.doRequest(ctx, http.MethodPost, "functions", nil, request, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// GetFunction retrieves a specific function by ID
func (c *Client) GetFunction(ctx context.Context, functionID string) (*FunctionResponse, error) {
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}

	var function FunctionResponse
	err := c.doRequest(ctx, http.MethodGet, "functions/"+functionID, nil, nil, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// ListFunctions lists all functions the user has access to
func (c *Client) ListFunctions(ctx context.Context, options *ListFunctionsOptions) (*FunctionList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.OrderBy != "" {
			query.Set("orderby", options.OrderBy)
		}
		if options.Search != "" {
			query.Set("search", options.Search)
		}
		if options.FilterScope != "" {
			query.Set("filter_scope", options.FilterScope)
		}
	}

	var functionList FunctionList
	err := c.doRequest(ctx, http.MethodGet, "functions", query, nil, &functionList)
	if err != nil {
		return nil, err
	}

	return &functionList, nil
}

// UpdateFunction updates an existing function
func (c *Client) UpdateFunction(ctx context.Context, functionID string, request *FunctionUpdateRequest) (*FunctionResponse, error) {
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}

	if request == nil {
		return nil, fmt.Errorf("function update request is required")
	}

	var function FunctionResponse
	err := c.doRequest(ctx, http.MethodPut, "functions/"+functionID, nil, request, &function)
	if err != nil {
		return nil, err
	}

	return &function, nil
}

// DeleteFunction deletes a function
func (c *Client) DeleteFunction(ctx context.Context, functionID string) error {
	if functionID == "" {
		return fmt.Errorf("function ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "functions/"+functionID, nil, nil, nil)
}

// RunFunction runs a function on a specific endpoint
func (c *Client) RunFunction(ctx context.Context, request *TaskRequest) (*TaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("task request is required")
	}

	if request.FunctionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}

	if request.EndpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var response TaskResponse
	err := c.doRequest(ctx, http.MethodPost, "run", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// RunBatch runs multiple functions in a batch
func (c *Client) RunBatch(ctx context.Context, request *BatchTaskRequest) (*BatchTaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("batch task request is required")
	}

	if len(request.Tasks) == 0 {
		return nil, fmt.Errorf("at least one task is required")
	}

	// Validate each task
	for i, task := range request.Tasks {
		if task.FunctionID == "" {
			return nil, fmt.Errorf("function ID is required for task %d", i)
		}
		if task.EndpointID == "" {
			return nil, fmt.Errorf("endpoint ID is required for task %d", i)
		}
	}

	var response BatchTaskResponse
	err := c.doRequest(ctx, http.MethodPost, "batch", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTaskStatus gets the status of a task
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	var status TaskStatus
	err := c.doRequest(ctx, http.MethodGet, "status/"+taskID, nil, nil, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// GetBatchStatus gets the status of multiple tasks
func (c *Client) GetBatchStatus(ctx context.Context, taskIDs []string) (*BatchTaskStatus, error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("at least one task ID is required")
	}

	// Build the request body as a map with a single "task_ids" key
	requestBody := map[string][]string{
		"task_ids": taskIDs,
	}

	var status BatchTaskStatus
	err := c.doRequest(ctx, http.MethodPost, "batch_status", nil, requestBody, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// ListTasks lists all tasks the user has submitted
func (c *Client) ListTasks(ctx context.Context, options *TaskListOptions) (*TaskList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(options.PerPage))
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.Status != "" {
			query.Set("status", options.Status)
		}
		if options.EndpointID != "" {
			query.Set("endpoint_id", options.EndpointID)
		}
		if options.FunctionID != "" {
			query.Set("function_id", options.FunctionID)
		}
	}

	var taskList TaskList
	err := c.doRequest(ctx, http.MethodGet, "tasks", query, nil, &taskList)
	if err != nil {
		return nil, err
	}

	return &taskList, nil
}

// CancelTask cancels a running task
func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	// The cancel endpoint might expect a specific format
	// This is just a placeholder; adjust according to the actual API
	return c.doRequest(ctx, http.MethodPost, "tasks/"+taskID+"/cancel", nil, nil, nil)
}

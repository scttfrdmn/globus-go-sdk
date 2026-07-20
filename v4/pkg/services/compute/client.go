// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

// RegisterFunction registers a serialized function with Globus Compute.
func (c *Client) RegisterFunction(ctx context.Context, fn *FunctionDefinition) (*FunctionRegistration, error) {
	if fn == nil {
		return nil, &core.ValidationError{Field: "fn", Message: "function definition is required"}
	}
	if fn.Name == "" {
		return nil, &core.ValidationError{Field: "Name", Message: "function name is required"}
	}
	if fn.Serialized == "" {
		return nil, &core.ValidationError{Field: "Serialized", Message: "serialized function code is required"}
	}

	var result FunctionRegistration
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/functions", nil, fn, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFunction updates metadata on a registered function.
func (c *Client) UpdateFunction(ctx context.Context, functionID string, update *FunctionUpdate) error {
	if functionID == "" {
		return &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}
	if update == nil {
		return &core.ValidationError{Field: "update", Message: "update data is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodPatch, fmt.Sprintf("/functions/%s", functionID), nil, update, nil)
}

// DeleteFunction removes a registered function.
func (c *Client) DeleteFunction(ctx context.Context, functionID string) error {
	if functionID == "" {
		return &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/functions/%s", functionID), nil, nil, nil)
}

// GetTaskStatus retrieves the status of a specific function execution task.
func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}

	var status TaskStatus
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/tasks/%s", taskID), nil, nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// ListTasks lists function execution tasks for an endpoint.
func (c *Client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
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

	var taskList TaskList
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/tasks", query, nil, &taskList); err != nil {
		return nil, err
	}
	return &taskList, nil
}

// CancelTask cancels a running compute task.
func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	path := fmt.Sprintf("/tasks/%s/cancel", taskID)
	return c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, nil, nil)
}

// RunBatch submits multiple function calls in a single request.
func (c *Client) RunBatch(ctx context.Context, request *BatchTaskRequest) (*BatchTaskResponse, error) {
	if request == nil {
		return nil, &core.ValidationError{Field: "request", Message: "batch task request is required"}
	}
	if len(request.Tasks) == 0 {
		return nil, &core.ValidationError{Field: "Tasks", Message: "at least one task is required"}
	}
	var response BatchTaskResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/batch", nil, request, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetBatchStatus retrieves the status of multiple tasks by ID.
func (c *Client) GetBatchStatus(ctx context.Context, taskIDs []string) (*BatchTaskStatus, error) {
	if len(taskIDs) == 0 {
		return nil, &core.ValidationError{Field: "taskIDs", Message: "at least one task ID is required"}
	}
	body := map[string][]string{"task_ids": taskIDs}
	var status BatchTaskStatus
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/batch_status", nil, body, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

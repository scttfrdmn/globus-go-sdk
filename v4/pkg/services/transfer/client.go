// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Transfer service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Transfer client
// In v4, config is required and must include explicit scopes
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	// Set default Transfer service URL if not provided
	if config.BaseURL == "" {
		config.BaseURL = "https://transfer.api.globus.org/v0.10"
	}

	// Create base client
	baseClient, err := core.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseClient: baseClient,
		baseURL:    config.BaseURL,
	}, nil
}

// GetEndpoint retrieves information about a specific endpoint
// v4: Context is always first parameter
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}

	var endpoint Endpoint
	path := fmt.Sprintf("/endpoint/%s", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}
	return &endpoint, nil
}

// ListEndpoints lists endpoints accessible to the user
// v4: Context is always first parameter
func (c *Client) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error) {
	query := url.Values{}

	if options != nil {
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var endpointList EndpointList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/endpoint_list", query, nil, &endpointList)
	if err != nil {
		return nil, err
	}

	return &endpointList, nil
}

// SubmitTransfer submits a transfer task
// v4: Context is always first parameter
func (c *Client) SubmitTransfer(ctx context.Context, transfer *Transfer) (*TaskSubmitResponse, error) {
	if transfer == nil {
		return nil, &core.ValidationError{
			Field:   "transfer",
			Message: "transfer data is required",
		}
	}

	// Validate required fields
	if transfer.SourceEndpoint == "" {
		return nil, &core.ValidationError{
			Field:   "SourceEndpoint",
			Message: "source endpoint is required",
		}
	}
	if transfer.DestinationEndpoint == "" {
		return nil, &core.ValidationError{
			Field:   "DestinationEndpoint",
			Message: "destination endpoint is required",
		}
	}
	if len(transfer.Items) == 0 {
		return nil, &core.ValidationError{
			Field:   "Items",
			Message: "at least one transfer item is required",
		}
	}

	// Set DATA_TYPE if not set
	if transfer.DATA_TYPE == "" {
		transfer.DATA_TYPE = "transfer"
	}

	var response TaskSubmitResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/transfer", nil, transfer, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SubmitDelete submits a delete task
// v4: Context is always first parameter
func (c *Client) SubmitDelete(ctx context.Context, delete *Delete) (*TaskSubmitResponse, error) {
	if delete == nil {
		return nil, &core.ValidationError{
			Field:   "delete",
			Message: "delete data is required",
		}
	}

	// Validate required fields
	if delete.Endpoint == "" {
		return nil, &core.ValidationError{
			Field:   "Endpoint",
			Message: "endpoint is required",
		}
	}
	if len(delete.Items) == 0 {
		return nil, &core.ValidationError{
			Field:   "Items",
			Message: "at least one delete item is required",
		}
	}

	// Set DATA_TYPE if not set
	if delete.DATA_TYPE == "" {
		delete.DATA_TYPE = "delete"
	}

	var response TaskSubmitResponse
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/delete", nil, delete, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTask retrieves information about a specific task
// v4: Context is always first parameter
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, &core.ValidationError{
			Field:   "taskID",
			Message: "task ID is required",
		}
	}

	var task Task
	path := fmt.Sprintf("/task/%s", taskID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, path, nil, nil, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// CancelTask cancels a running task
// v4: Context is always first parameter
func (c *Client) CancelTask(ctx context.Context, taskID string) (*TaskCancelResponse, error) {
	if taskID == "" {
		return nil, &core.ValidationError{
			Field:   "taskID",
			Message: "task ID is required",
		}
	}

	var response TaskCancelResponse
	path := fmt.Sprintf("/task/%s/cancel", taskID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// ListTasks lists tasks with optional filtering
// v4: Context is always first parameter
func (c *Client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
	query := url.Values{}

	if options != nil {
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if len(options.FilterStatus) > 0 {
			for _, status := range options.FilterStatus {
				query.Add("filter_status", status)
			}
		}
	}

	var taskList TaskList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/task_list", query, nil, &taskList)
	if err != nil {
		return nil, err
	}

	return &taskList, nil
}

// ListDirectory lists directory contents on an endpoint
// v4: Context is always first parameter
func (c *Client) ListDirectory(ctx context.Context, endpointID, path string, options *ListDirectoryOptions) (*DirectoryListing, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}

	query := url.Values{}
	if path != "" {
		query.Set("path", path)
	}

	if options != nil {
		if options.ShowHidden {
			query.Set("show_hidden", "1")
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var listing DirectoryListing
	apiPath := fmt.Sprintf("/operation/endpoint/%s/ls", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodGet, apiPath, query, nil, &listing)
	if err != nil {
		return nil, err
	}

	return &listing, nil
}

// MakeDirectory creates a directory on an endpoint
// v4: Context is always first parameter
func (c *Client) MakeDirectory(ctx context.Context, endpointID, path string) (*OperationResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}
	if path == "" {
		return nil, &core.ValidationError{
			Field:   "path",
			Message: "path is required",
		}
	}

	body := map[string]interface{}{
		"DATA_TYPE": "mkdir",
		"path":      path,
	}

	var response OperationResponse
	apiPath := fmt.Sprintf("/operation/endpoint/%s/mkdir", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, apiPath, nil, body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Rename renames a file or directory on an endpoint
// v4: Context is always first parameter
func (c *Client) Rename(ctx context.Context, endpointID, oldPath, newPath string) (*OperationResponse, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{
			Field:   "endpointID",
			Message: "endpoint ID is required",
		}
	}
	if oldPath == "" {
		return nil, &core.ValidationError{
			Field:   "oldPath",
			Message: "old path is required",
		}
	}
	if newPath == "" {
		return nil, &core.ValidationError{
			Field:   "newPath",
			Message: "new path is required",
		}
	}

	body := map[string]interface{}{
		"DATA_TYPE": "rename",
		"old_path":  oldPath,
		"new_path":  newPath,
	}

	var response OperationResponse
	apiPath := fmt.Sprintf("/operation/endpoint/%s/rename", endpointID)
	err := c.baseClient.DoRequest(ctx, http.MethodPost, apiPath, nil, body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}


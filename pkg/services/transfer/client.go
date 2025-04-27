// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

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

// Constants for Globus Transfer
const (
	DefaultBaseURL = "https://transfer.api.globus.org/v0.10/"
	TransferScope  = "urn:globus:auth:scope:transfer.api.globus.org:all"
)

// Client provides methods for interacting with Globus Transfer
type Client struct {
	Client *core.Client
}

// NewClient creates a new Transfer client
func NewClient(accessToken string, options ...core.ClientOption) *Client {
	// Create the authorizer with the access token
	authorizer := authorizers.StaticTokenCoreAuthorizer(accessToken)

	// Apply default options specific to Transfer
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

// buildURL builds a URL for the transfer API
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

// ListEndpoints retrieves endpoints the user has access to
func (c *Client) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.FilterFullText != "" {
			query.Set("filter_fulltext", options.FilterFullText)
		}
		if options.FilterOwnerID != "" {
			query.Set("filter_owner_id", options.FilterOwnerID)
		}
		if options.FilterHostEndpoint != "" {
			query.Set("filter_host_endpoint", options.FilterHostEndpoint)
		}
		if options.FilterScope != "" {
			query.Set("filter_scope", options.FilterScope)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("page_token", options.PageToken)
		}
	}

	var endpointList EndpointList
	err := c.doRequest(ctx, http.MethodGet, "endpoint_search", query, nil, &endpointList)
	if err != nil {
		return nil, err
	}

	return &endpointList, nil
}

// GetEndpoint retrieves a specific endpoint by ID
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var endpoint Endpoint
	err := c.doRequest(ctx, http.MethodGet, "endpoint/"+endpointID, nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// ActivateEndpoint activates an endpoint
func (c *Client) ActivateEndpoint(ctx context.Context, endpointID string) error {
	if endpointID == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	body := map[string]bool{
		"auto_activate": true,
	}

	var result OperationResult
	err := c.doRequest(ctx, http.MethodPost, "endpoint/"+endpointID+"/autoactivate", nil, body, &result)
	if err != nil {
		return err
	}

	// Check for activation error (code is not "AutoActivated" or "AlreadyActivated")
	if result.Code != "AutoActivated" && result.Code != "AlreadyActivated" {
		return fmt.Errorf("activation failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// GetActivationRequirements gets the requirements for activating an endpoint
func (c *Client) GetActivationRequirements(ctx context.Context, endpointID string) (*ActivationRequirements, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var requirements ActivationRequirements
	err := c.doRequest(ctx, http.MethodGet, "endpoint/"+endpointID+"/activation_requirements", nil, nil, &requirements)
	if err != nil {
		return nil, err
	}

	return &requirements, nil
}

// ListFiles lists the files and directories in a path on an endpoint
func (c *Client) ListFiles(ctx context.Context, endpointID, path string, options *ListFileOptions) (*FileList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	// Convert options to query parameters
	query := url.Values{}
	query.Set("path", path)

	if options != nil {
		if options.OrderBy != "" {
			query.Set("orderby", options.OrderBy)
		}
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.ShowHidden {
			query.Set("show_hidden", "1")
		}
		if options.ContinueFrom != "" {
			query.Set("continue_from", options.ContinueFrom)
		}
		if options.Marker != "" {
			query.Set("marker", options.Marker)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.ExcludedTypes != "" {
			query.Set("excluded_types", options.ExcludedTypes)
		}
	}

	var fileList FileList
	err := c.doRequest(ctx, http.MethodGet, "operation/endpoint/"+endpointID+"/ls", query, nil, &fileList)
	if err != nil {
		return nil, err
	}

	return &fileList, nil
}

// CreateTransferTask creates a new transfer task
func (c *Client) CreateTransferTask(ctx context.Context, request *TransferTaskRequest) (*TaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("transfer task request is required")
	}

	if request.SourceEndpointID == "" {
		return nil, fmt.Errorf("source endpoint is required")
	}

	if request.DestinationEndpointID == "" {
		return nil, fmt.Errorf("destination endpoint is required")
	}

	if len(request.Items) == 0 {
		return nil, fmt.Errorf("at least one transfer item is required")
	}

	// Set data type if not already set
	if request.DataType == "" {
		request.DataType = "transfer"
	}

	var response TaskResponse
	err := c.doRequest(ctx, http.MethodPost, "transfer", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateDeleteTask creates a new delete task
func (c *Client) CreateDeleteTask(ctx context.Context, request *DeleteTaskRequest) (*TaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("delete task request is required")
	}

	if request.EndpointID == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	if len(request.Items) == 0 {
		return nil, fmt.Errorf("at least one delete item is required")
	}

	// Set data type if not already set
	if request.DataType == "" {
		request.DataType = "delete"
	}

	var response TaskResponse
	err := c.doRequest(ctx, http.MethodPost, "delete", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ListTasks retrieves tasks the user has submitted
func (c *Client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
	// Convert options to query parameters
	query := url.Values{}
	if options != nil {
		if options.FilterTaskID != "" {
			query.Set("filter_task_id", options.FilterTaskID)
		}
		if options.FilterType != "" {
			query.Set("filter_type", options.FilterType)
		}
		if options.FilterStatus != "" {
			query.Set("filter_status", options.FilterStatus)
		}
		if !options.FilterCompletedSince.IsZero() {
			query.Set("filter_completion_time.min", options.FilterCompletedSince.Format(time.RFC3339))
		}
		if !options.FilterCompletedUntil.IsZero() {
			query.Set("filter_completion_time.max", options.FilterCompletedUntil.Format(time.RFC3339))
		}
		if !options.FilterRequestedSince.IsZero() {
			query.Set("filter_request_time.min", options.FilterRequestedSince.Format(time.RFC3339))
		}
		if !options.FilterRequestedUntil.IsZero() {
			query.Set("filter_request_time.max", options.FilterRequestedUntil.Format(time.RFC3339))
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if options.PageSize > 0 {
			query.Set("page_size", strconv.Itoa(options.PageSize))
		}
		if options.PageToken != "" {
			query.Set("page_token", options.PageToken)
		}
	}

	var taskList TaskList
	err := c.doRequest(ctx, http.MethodGet, "task_list", query, nil, &taskList)
	if err != nil {
		return nil, err
	}

	return &taskList, nil
}

// GetTask retrieves a specific task by ID
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	var task Task
	err := c.doRequest(ctx, http.MethodGet, "task/"+taskID, nil, nil, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CancelTask cancels a task
func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	var result OperationResult
	err := c.doRequest(ctx, http.MethodPost, "task/"+taskID+"/cancel", nil, nil, &result)
	if err != nil {
		return err
	}

	// Check for cancellation error
	if result.Code != "Canceled" {
		return fmt.Errorf("cancellation failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// Mkdir creates a directory on an endpoint
func (c *Client) Mkdir(ctx context.Context, endpointID, path string) error {
	if endpointID == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	if path == "" {
		return fmt.Errorf("path is required")
	}

	body := map[string]string{
		"path":      path,
		"DATA_TYPE": "mkdir",
	}

	var result OperationResult
	err := c.doRequest(ctx, http.MethodPost, "operation/endpoint/"+endpointID+"/mkdir", nil, body, &result)
	if err != nil {
		return err
	}

	// Check for mkdir error
	if result.Code != "DirectoryCreated" {
		return fmt.Errorf("mkdir failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// Rename renames a file or directory on an endpoint
func (c *Client) Rename(ctx context.Context, endpointID, oldPath, newPath string) error {
	if endpointID == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	if oldPath == "" || newPath == "" {
		return fmt.Errorf("old path and new path are required")
	}

	body := map[string]string{
		"old_path":  oldPath,
		"new_path":  newPath,
		"DATA_TYPE": "rename",
	}

	var result OperationResult
	err := c.doRequest(ctx, http.MethodPost, "operation/endpoint/"+endpointID+"/rename", nil, body, &result)
	if err != nil {
		return err
	}

	// Check for rename error
	if result.Code != "FileRenamed" {
		return fmt.Errorf("rename failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// SubmitTransfer is a helper function to create and submit a simple transfer task
func (c *Client) SubmitTransfer(
	ctx context.Context,
	sourceEndpointID, sourcePath string,
	destinationEndpointID, destinationPath string,
	label string,
	options map[string]interface{},
) (*TaskResponse, error) {
	// Create transfer item
	item := TransferItem{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
	}

	// Create transfer request
	request := &TransferTaskRequest{
		DataType:              "transfer",
		Label:                 label,
		SourceEndpointID:      sourceEndpointID,
		DestinationEndpointID: destinationEndpointID,
		Items:                 []TransferItem{item},
	}

	// Apply options if provided
	if options != nil {
		if v, ok := options["recursive"].(bool); ok {
			item.Recursive = v
			request.Items[0] = item
		}
		if v, ok := options["verify_checksum"].(bool); ok {
			request.VerifyChecksum = v
		}
		if v, ok := options["encrypt"].(bool); ok {
			request.Encrypt = v
		}
		if v, ok := options["sync_level"].(int); ok {
			request.SyncLevel = v
		}
		if v, ok := options["delete_destination_extra"].(bool); ok {
			request.DeleteDestinationExtra = v
		}
		if v, ok := options["deadline"].(*time.Time); ok {
			request.Deadline = v
		}
		if v, ok := options["notify_on_succeeded"].(bool); ok {
			request.NotifyOnSucceeded = v
		}
		if v, ok := options["notify_on_failed"].(bool); ok {
			request.NotifyOnFailed = v
		}
		if v, ok := options["notify_on_inactive"].(bool); ok {
			request.NotifyOnInactive = v
		}
		if v, ok := options["preserve_mtime"].(bool); ok {
			request.PreserveMtime = v
		}
	}

	// Submit the transfer task
	return c.CreateTransferTask(ctx, request)
}

// SubmitResumableTransfer creates and starts a resumable transfer
func (c *Client) SubmitResumableTransfer(
	ctx context.Context,
	sourceEndpointID, sourcePath string,
	destinationEndpointID, destinationPath string,
	options *ResumableTransferOptions,
) (string, error) {
	return c.CreateResumableTransfer(ctx, sourceEndpointID, sourcePath, destinationEndpointID, destinationPath, options)
}

// GetResumableTransferStatus gets the status of a resumable transfer
func (c *Client) GetResumableTransferStatus(
	ctx context.Context,
	checkpointID string,
) (*CheckpointState, error) {
	return c.GetTransferCheckpoint(ctx, checkpointID)
}

// ResumeResumableTransfer resumes a previously started resumable transfer
func (c *Client) ResumeResumableTransfer(
	ctx context.Context, 
	checkpointID string,
	options *ResumableTransferOptions,
) (*ResumableTransferResult, error) {
	return c.ResumeTransfer(ctx, checkpointID, options)
}

// CancelResumableTransfer cancels a resumable transfer by deleting its checkpoint
func (c *Client) CancelResumableTransfer(
	ctx context.Context,
	checkpointID string,
) error {
	return c.DeleteTransferCheckpoint(ctx, checkpointID)
}

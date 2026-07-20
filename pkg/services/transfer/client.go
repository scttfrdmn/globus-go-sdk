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
	"os"
	"strconv"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/ratelimit"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/response"
)

// Constants for Globus Transfer
const (
	DefaultBaseURL    = "https://transfer.api.globus.org/v0.10/"
	TransferScope     = "urn:globus:auth:scope:transfer.api.globus.org:all"
	MinimumAPIVersion = "v0.10" // Minimum supported API version
	ServiceName       = "transfer"
	APIVersion        = "v0.10"
)

// Client provides methods for interacting with Globus Transfer
type Client struct {
	Client *core.Client
}

// NewClient creates a new Transfer client
func NewClient(options ...Option) (*Client, error) {
	// Apply the options to create the client configuration
	cfg := &ClientConfig{}
	for _, option := range options {
		option(cfg)
	}

	// Validate configuration
	if cfg.authorizer == nil {
		return nil, fmt.Errorf("authorizer is required")
	}

	// Apply default options specific to Transfer
	defaultOptions := []core.ClientOption{
		core.WithBaseURL(DefaultBaseURL),
		core.WithAuthorizer(cfg.authorizer),
		// Default to a token bucket rate limiter
		core.WithRateLimiter(ratelimit.NewTokenBucketLimiter(nil)),
	}

	// Apply debug options if enabled
	if cfg.debug {
		defaultOptions = append(defaultOptions, core.WithHTTPDebugging(true))
	}
	if cfg.trace {
		defaultOptions = append(defaultOptions, core.WithHTTPTracing(true))
	}
	if cfg.logger != nil {
		defaultOptions = append(defaultOptions, core.WithLogger(cfg.logger))
	}

	// Apply any additional core options
	if cfg.coreOptions != nil {
		defaultOptions = append(defaultOptions, cfg.coreOptions...)
	}

	// Create the base client
	baseClient := core.NewClient(defaultOptions...)

	return &Client{
		Client: baseClient,
	}, nil
}

// buildURLLowLevel builds a URL for the transfer API
// This is an internal method used by the client.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
	baseURL := c.Client.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	url := baseURL + path
	if len(query) > 0 {
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

		// Debug output for request
		if os.Getenv("HTTP_DEBUG") != "" {
			fmt.Printf("DEBUG REQUEST URL: %s\n", url)
			fmt.Printf("DEBUG REQUEST BODY: %s\n", string(bodyJSON))
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
	defer func() { _ = resp.Body.Close() }()

	// Check for non-success status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return parseTransferError(resp.StatusCode, respBody)
	}

	// Process rate limit headers if present
	if limiter := c.Client.RateLimiter; limiter != nil {
		limit := parseIntHeader(resp.Header, "X-RateLimit-Limit", -1)
		remaining := parseIntHeader(resp.Header, "X-RateLimit-Remaining", -1)
		reset := parseIntHeader(resp.Header, "X-RateLimit-Reset", -1)

		if limit > 0 && remaining >= 0 && reset > 0 {
			_ = limiter.UpdateLimit(limit, remaining, reset)
		}
	}

	// Process 204 No Content or empty responses
	if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
		if response == nil {
			return nil
		}
		// If caller expects a response but we got none, set an empty response
		// This can happen with PATCH/PUT operations that don't return content
		return nil
	}

	// Read and decode response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug output for response
	if os.Getenv("HTTP_DEBUG") != "" {
		fmt.Printf("DEBUG RESPONSE STATUS: %d\n", resp.StatusCode)
		fmt.Printf("DEBUG RESPONSE BODY: %s\n", string(respBody))
	}

	if len(respBody) == 0 {
		return nil
	}

	// Parse the response body
	if response != nil {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// endpointSearchQuery builds the endpoint_search wire query from options.
// PageSize/PageToken are divergent Go aliases and are intentionally not sent.
func endpointSearchQuery(options *ListEndpointsOptions) url.Values {
	query := url.Values{}
	if options == nil {
		return query
	}
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
	if options.FilterNonFunctional {
		query.Set("filter_non_functional", "1")
	}
	if options.FilterEntityType != "" {
		query.Set("filter_entity_type", options.FilterEntityType)
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Offset > 0 {
		query.Set("offset", strconv.Itoa(options.Offset))
	}
	return query
}

// ListEndpoints retrieves endpoints the user has access to
func (c *Client) ListEndpoints(ctx context.Context, options *ListEndpointsOptions) (*EndpointList, error) {
	var endpointList EndpointList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_search", endpointSearchQuery(options), nil, &endpointList)
	if err != nil {
		return nil, err
	}

	return &endpointList, nil
}

// ListEndpointsV2 retrieves endpoints with unified response system
func (c *Client) ListEndpointsV2(ctx context.Context, options *ListEndpointsOptions) (*response.TransferResponse[EndpointList], error) {
	var endpointList EndpointList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint_search", endpointSearchQuery(options), nil, &endpointList)
	if err != nil {
		// Convert to GlobusError if it's not already
		if _, ok := err.(*errors.GlobusError); !ok {
			return nil, errors.NewTransferError("EndpointListError", err.Error()).WithUnderlying(err)
		}
		return nil, err
	}

	transferResp := response.NewTransferResponse(endpointList)
	transferResp.WithRequestID("transfer-endpoints-" + strconv.FormatInt(time.Now().UnixNano(), 10))

	return transferResp, nil
}

// GetEndpoint retrieves a specific endpoint by ID
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (*Endpoint, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	var endpoint Endpoint
	err := c.doRequestLowLevel(ctx, http.MethodGet, "endpoint/"+endpointID, nil, nil, &endpoint)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// NOTE: ActivateEndpoint and GetActivationRequirements have been removed.
// Modern Globus endpoints supporting the minimum API version (v0.10+) use
// auto-activation with properly scoped tokens. Explicit activation is no longer
// needed or supported by this SDK.

// ListDirectoryOptions contains options for listing directories.
//
// ContinueFrom/Marker/ExcludedTypes are divergent Go aliases retained for
// source compatibility; they are not operation_ls wire params.
type ListDirectoryOptions struct {
	EndpointID    string
	Path          string
	OrderBy       string
	Filter        string
	ShowHidden    bool
	Limit         int
	Offset        int
	LocalUser     string
	ContinueFrom  string // divergent alias; not sent
	Marker        string // divergent alias; not sent
	ExcludedTypes string // divergent alias; not sent
}

// ListDirectory lists files and directories at a path - helper method with structured options
func (c *Client) ListDirectory(ctx context.Context, options *ListDirectoryOptions) (*FileList, error) {
	if options == nil {
		return nil, fmt.Errorf("options are required")
	}

	if options.EndpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	// Convert to ListFileOptions for the underlying implementation
	fileOptions := &ListFileOptions{
		OrderBy:    options.OrderBy,
		Filter:     options.Filter,
		ShowHidden: options.ShowHidden,
		Limit:      options.Limit,
		Offset:     options.Offset,
		LocalUser:  options.LocalUser,
	}

	return c.ListFiles(ctx, options.EndpointID, options.Path, fileOptions)
}

// ListFiles lists the files and directories in a path on an endpoint (operation_ls).
func (c *Client) ListFiles(ctx context.Context, endpointID, path string, options *ListFileOptions) (*FileList, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}

	// Convert options to query parameters. show_hidden encodes as 1/0.
	// ExcludedTypes/ContinueFrom/Marker are not operation_ls wire params.
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
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
		if options.LocalUser != "" {
			query.Set("local_user", options.LocalUser)
		}
	}

	var fileList FileList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "operation/endpoint/"+endpointID+"/ls", query, nil, &fileList)
	if err != nil {
		return nil, err
	}

	return &fileList, nil
}

// GetSubmissionID obtains a submission ID for transfer operations
func (c *Client) GetSubmissionID(ctx context.Context) (string, error) {
	// Return a simulated submission ID only for unit tests, not integration tests
	if os.Getenv("MOCK_SUBMISSION_ID") == "true" {
		return "mock-submission-id-for-testing", nil
	}

	var response struct {
		Value        string `json:"value"`
		SubmissionID string `json:"submission_id"`
	}

	// The API endpoint is a GET request, not POST
	err := c.doRequestLowLevel(ctx, http.MethodGet, "submission_id", nil, nil, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get submission ID: %w", err)
	}

	// Depending on the API response format, one of these will be populated
	if response.SubmissionID != "" {
		return response.SubmissionID, nil
	}
	return response.Value, nil
}

// GetSubmissionIDV2 obtains a submission ID for transfer operations with unified response system
func (c *Client) GetSubmissionIDV2(ctx context.Context) (*response.TransferResponse[string], error) {
	// Return a simulated submission ID only for unit tests, not integration tests
	if os.Getenv("MOCK_SUBMISSION_ID") == "true" {
		mockResponse := response.NewTransferResponse("mock-submission-id-for-testing")
		mockResponse.WithRequestID("mock-request-id")
		return mockResponse, nil
	}

	var resp struct {
		Value        string `json:"value"`
		SubmissionID string `json:"submission_id"`
	}

	// The API endpoint is a GET request, not POST
	err := c.doRequestLowLevel(ctx, http.MethodGet, "submission_id", nil, nil, &resp)
	if err != nil {
		// Convert to GlobusError if it's not already
		if _, ok := err.(*errors.GlobusError); !ok {
			return nil, errors.NewTransferError("SubmissionIDError", err.Error()).WithUnderlying(err)
		}
		return nil, err
	}

	// Depending on the API response format, one of these will be populated
	submissionID := resp.SubmissionID
	if submissionID == "" {
		submissionID = resp.Value
	}

	transferResp := response.NewTransferResponse(submissionID)
	transferResp.WithRequestID("transfer-" + submissionID)

	return transferResp, nil
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

	// Ensure each transfer item has the DATA_TYPE field set
	for i := range request.Items {
		if request.Items[i].DataType == "" {
			request.Items[i].DataType = "transfer_item"
		}
	}

	// Get a submission ID if not provided
	if request.SubmissionID == "" {
		var err error
		request.SubmissionID, err = c.GetSubmissionID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get submission ID: %w", err)
		}
	}

	var response TaskResponse
	err := c.doRequestLowLevel(ctx, http.MethodPost, "transfer", nil, request, &response)
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

	// Ensure each delete item has the DATA_TYPE field set
	for i := range request.Items {
		if request.Items[i].DataType == "" {
			request.Items[i].DataType = "delete_item"
		}
	}

	// Note: The API does not support a "recursive" field for delete_item as of API v0.10
	// Instead, all deletions in Globus Transfer appear to be recursive by default

	// Get a submission ID if not provided
	if request.SubmissionID == "" {
		var err error
		request.SubmissionID, err = c.GetSubmissionID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get submission ID: %w", err)
		}
	}

	var response TaskResponse
	err := c.doRequestLowLevel(ctx, http.MethodPost, "delete", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ListTasks retrieves tasks the user has submitted (task_list).
//
// The 3.65.0 wire form is limit, offset, orderby (comma-joined) and a single
// combined filter param (key:v1,v2/key2:v3). The individual Filter*/PageSize/
// PageToken option fields are divergent Go aliases and are not sent.
func (c *Client) ListTasks(ctx context.Context, options *ListTasksOptions) (*TaskList, error) {
	query := url.Values{}
	if options != nil {
		if options.Filter != "" {
			query.Set("filter", options.Filter)
		}
		if options.OrderBy != "" {
			query.Set("orderby", options.OrderBy)
		}
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var taskList TaskList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "task_list", query, nil, &taskList)
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
	err := c.doRequestLowLevel(ctx, http.MethodGet, "task/"+taskID, nil, nil, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CancelTask cancels a task
func (c *Client) CancelTask(ctx context.Context, taskID string) (*OperationResult, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	var result OperationResult
	err := c.doRequestLowLevel(ctx, http.MethodPost, "task/"+taskID+"/cancel", nil, nil, &result)
	if err != nil {
		return nil, err
	}

	// Add the task ID to the result for convenience
	result.TaskID = taskID

	return &result, nil
}

// CreateDirectoryOptions contains options for the CreateDirectory method
type CreateDirectoryOptions struct {
	EndpointID string
	Path       string
}

// CreateDirectory creates a directory on an endpoint - helper method with structured options
func (c *Client) CreateDirectory(ctx context.Context, options *CreateDirectoryOptions) error {
	if options == nil {
		return fmt.Errorf("options are required")
	}

	if options.EndpointID == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	if options.Path == "" {
		return fmt.Errorf("path is required")
	}

	return c.Mkdir(ctx, options.EndpointID, options.Path, nil)
}

// MkdirOptions carries optional fields for Mkdir.
type MkdirOptions struct {
	LocalUser string // maps to the optional local_user body field
}

// Mkdir creates a directory on an endpoint (operation_mkdir).
func (c *Client) Mkdir(ctx context.Context, endpointID, path string, opts *MkdirOptions) error {
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
	if opts != nil && opts.LocalUser != "" {
		body["local_user"] = opts.LocalUser
	}

	var result OperationResult
	err := c.doRequestLowLevel(ctx, http.MethodPost, "operation/endpoint/"+endpointID+"/mkdir", nil, body, &result)
	if err != nil {
		return err
	}

	// Check for mkdir error
	if result.Code != "DirectoryCreated" {
		return fmt.Errorf("mkdir failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// RenameOptions carries optional fields for Rename.
type RenameOptions struct {
	LocalUser string // maps to the optional local_user body field
}

// Rename renames a file or directory on an endpoint (operation_rename).
func (c *Client) Rename(ctx context.Context, endpointID, oldPath, newPath string, opts *RenameOptions) error {
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
	if opts != nil && opts.LocalUser != "" {
		body["local_user"] = opts.LocalUser
	}

	var result OperationResult
	err := c.doRequestLowLevel(ctx, http.MethodPost, "operation/endpoint/"+endpointID+"/rename", nil, body, &result)
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
		DataType:        "transfer_item",
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

	// Get a submission ID
	submissionID, err := c.GetSubmissionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission ID: %w", err)
	}
	request.SubmissionID = submissionID

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

// parseIntHeader parses an integer header value with a default fallback
func parseIntHeader(header http.Header, key string, defaultValue int) int {
	if header == nil {
		return defaultValue
	}

	value := header.Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// SetSubscriptionID sets the subscription ID for a collection/endpoint
// (PUT endpoint/{id}/subscription). subscriptionID may be a subscription UUID,
// "DEFAULT", or "null" to clear. The body carries no DATA_TYPE key.
func (c *Client) SetSubscriptionID(ctx context.Context, collectionID, subscriptionID string) error {
	if collectionID == "" {
		return fmt.Errorf("collection ID is required")
	}

	body := map[string]interface{}{
		"subscription_id": subscriptionID,
	}

	return c.doRequestLowLevel(ctx, http.MethodPut, "endpoint/"+collectionID+"/subscription", nil, body, nil)
}

// SetSubscriptionAdminVerified marks a collection/endpoint's subscription as
// admin-verified (PUT endpoint/{id}/subscription_admin_verified). Admin-only.
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, collectionID string, verified bool) error {
	if collectionID == "" {
		return fmt.Errorf("collection ID is required")
	}

	body := map[string]interface{}{
		"subscription_admin_verified": verified,
	}

	return c.doRequestLowLevel(ctx, http.MethodPut, "endpoint/"+collectionID+"/subscription_admin_verified", nil, body, nil)
}

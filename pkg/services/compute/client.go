// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
)

// Constants for Globus Compute. The base URL is the bare host; each endpoint
// carries its own /v2 or /v3 prefix (upstream globus-sdk 3.65.0 defines no
// request/response models and no pagination, so bodies and results are
// passthrough documents).
const (
	DefaultBaseURL = "https://compute.api.globus.org/"
	ComputeScope   = "https://auth.globus.org/scopes/facd7ccc-c5f4-42aa-916b-a0e270e2c2a9/all"
)

// Client provides methods for interacting with Globus Compute. It folds the
// upstream ComputeClientV2 and ComputeClientV3 into one client; V3 methods carry
// a "V3" suffix.
type Client struct {
	Client *core.Client
}

// NewClient creates a new Compute client
func NewClient(opts ...ClientOption) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	if options.accessToken != "" {
		authorizer := authorizers.StaticTokenCoreAuthorizer(options.accessToken)
		options.coreOptions = append(options.coreOptions, core.WithAuthorizer(authorizer))
	}
	baseClient := core.NewClient(options.coreOptions...)
	return &Client{Client: baseClient}, nil
}

func (c *Client) buildURL(path string, query url.Values) string {
	baseURL := c.Client.BaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}
	u := baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	u := c.buildURL(path, query)
	var bodyReader io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.doRequest(ctx, http.MethodGet, path, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) send(ctx context.Context, method, path string, body interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.doRequest(ctx, method, path, nil, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// getInto issues a GET and decodes the response into dest, which may be any JSON
// shape (object, array, or scalar). Used by endpoints whose responses are not
// JSON objects.
func (c *Client) getInto(ctx context.Context, path string, query url.Values, dest interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, query, nil, dest)
}

// --- Service-level (V2) ---

// GetVersion returns the compute service version (GET /v2/version). service is
// an optional query param; pass "" to omit.
//
// The response is polymorphic: with no service it is a bare JSON string (the API
// version), and with a service it is a JSON object. The result is returned as an
// untyped value (string or map[string]interface{}); type-assert as needed.
func (c *Client) GetVersion(ctx context.Context, service string) (interface{}, error) {
	query := url.Values{}
	if service != "" {
		query.Set("service", service)
	}
	var result interface{}
	if err := c.getInto(ctx, "v2/version", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetResultAMQPURL returns a connection URL for the result AMQP queue
// (GET /v2/get_amqp_result_connection_url).
func (c *Client) GetResultAMQPURL(ctx context.Context) (map[string]interface{}, error) {
	return c.get(ctx, "v2/get_amqp_result_connection_url", nil)
}

// --- Endpoints (V2) ---

// GetEndpointsOptions carries the optional "role" query param for GetEndpoints.
type GetEndpointsOptions struct {
	Role string
}

// RegisterEndpoint registers a compute endpoint (POST /v2/endpoints).
func (c *Client) RegisterEndpoint(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("endpoint document is required")
	}
	return c.send(ctx, http.MethodPost, "v2/endpoints", data)
}

// GetEndpoint retrieves a compute endpoint (GET /v2/endpoints/{id}).
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.get(ctx, "v2/endpoints/"+endpointID, nil)
}

// GetEndpoints lists compute endpoints (GET /v2/endpoints). Pass opts.Role to
// filter; opts may be nil.
//
// The response is a top-level JSON array of endpoint documents, so the result is
// returned as a slice of passthrough maps.
func (c *Client) GetEndpoints(ctx context.Context, opts *GetEndpointsOptions) ([]map[string]interface{}, error) {
	query := url.Values{}
	if opts != nil && opts.Role != "" {
		query.Set("role", opts.Role)
	}
	var result []map[string]interface{}
	if err := c.getInto(ctx, "v2/endpoints", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetEndpointStatus retrieves an endpoint's status (GET /v2/endpoints/{id}/status).
func (c *Client) GetEndpointStatus(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.get(ctx, "v2/endpoints/"+endpointID+"/status", nil)
}

// DeleteEndpoint deletes a compute endpoint (DELETE /v2/endpoints/{id}).
func (c *Client) DeleteEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.send(ctx, http.MethodDelete, "v2/endpoints/"+endpointID, nil)
}

// LockEndpoint locks a compute endpoint (POST /v2/endpoints/{id}/lock).
func (c *Client) LockEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.send(ctx, http.MethodPost, "v2/endpoints/"+endpointID+"/lock", nil)
}

// --- Functions (V2) ---

// RegisterFunction registers a function (POST /v2/functions).
func (c *Client) RegisterFunction(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("function document is required")
	}
	return c.send(ctx, http.MethodPost, "v2/functions", data)
}

// GetFunction retrieves a registered function's metadata (GET /v2/functions/{id}).
func (c *Client) GetFunction(ctx context.Context, functionID string) (map[string]interface{}, error) {
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}
	return c.get(ctx, "v2/functions/"+functionID, nil)
}

// DeleteFunction deletes a registered function (DELETE /v2/functions/{id}).
func (c *Client) DeleteFunction(ctx context.Context, functionID string) (map[string]interface{}, error) {
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}
	return c.send(ctx, http.MethodDelete, "v2/functions/"+functionID, nil)
}

// --- Tasks (V2) ---

// GetTask retrieves a task's status and result (GET /v2/tasks/{id}).
func (c *Client) GetTask(ctx context.Context, taskID string) (map[string]interface{}, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	return c.get(ctx, "v2/tasks/"+taskID, nil)
}

// GetBatchStatus retrieves the status of multiple tasks (POST /v2/batch_status).
func (c *Client) GetBatchStatus(ctx context.Context, taskIDs []string) (map[string]interface{}, error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("at least one task ID is required")
	}
	return c.send(ctx, http.MethodPost, "v2/batch_status", map[string]interface{}{"task_ids": taskIDs})
}

// GetTaskGroup lists the task IDs for a task group (GET /v2/taskgroup/{id}).
func (c *Client) GetTaskGroup(ctx context.Context, taskGroupID string) (map[string]interface{}, error) {
	if taskGroupID == "" {
		return nil, fmt.Errorf("task group ID is required")
	}
	return c.get(ctx, "v2/taskgroup/"+taskGroupID, nil)
}

// Submit submits a task batch (POST /v2/submit).
func (c *Client) Submit(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("submit document is required")
	}
	return c.send(ctx, http.MethodPost, "v2/submit", data)
}

// --- V3 ---

// RegisterEndpointV3 registers an endpoint via the v3 API (POST /v3/endpoints).
func (c *Client) RegisterEndpointV3(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("endpoint document is required")
	}
	return c.send(ctx, http.MethodPost, "v3/endpoints", data)
}

// UpdateEndpointV3 updates an endpoint via the v3 API (PUT /v3/endpoints/{id}).
func (c *Client) UpdateEndpointV3(ctx context.Context, endpointID string, data map[string]interface{}) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	if data == nil {
		return nil, fmt.Errorf("endpoint document is required")
	}
	return c.send(ctx, http.MethodPut, "v3/endpoints/"+endpointID, data)
}

// LockEndpointV3 locks an endpoint via the v3 API (POST /v3/endpoints/{id}/lock).
func (c *Client) LockEndpointV3(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.send(ctx, http.MethodPost, "v3/endpoints/"+endpointID+"/lock", nil)
}

// GetEndpointAllowlistV3 lists an endpoint's allowed function IDs
// (GET /v3/endpoints/{id}/allowed_functions).
func (c *Client) GetEndpointAllowlistV3(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	return c.get(ctx, "v3/endpoints/"+endpointID+"/allowed_functions", nil)
}

// RegisterFunctionV3 registers a function via the v3 API (POST /v3/functions).
func (c *Client) RegisterFunctionV3(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("function document is required")
	}
	return c.send(ctx, http.MethodPost, "v3/functions", data)
}

// SubmitV3 submits a task batch to an endpoint via the v3 API
// (POST /v3/endpoints/{id}/submit).
func (c *Client) SubmitV3(ctx context.Context, endpointID string, data map[string]interface{}) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, fmt.Errorf("endpoint ID is required")
	}
	if data == nil {
		return nil, fmt.Errorf("submit document is required")
	}
	return c.send(ctx, http.MethodPost, "v3/endpoints/"+endpointID+"/submit", data)
}

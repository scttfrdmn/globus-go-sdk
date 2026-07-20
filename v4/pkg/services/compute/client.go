// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Compute service client with context-first design. It folds
// the upstream ComputeClientV2 and ComputeClientV3 into a single client; V3
// methods carry a "V3" suffix. Upstream compute defines no request/response
// models, so bodies and results are passthrough documents.
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Compute client
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	if config.BaseURL == "" {
		// Host root, not /v2: each endpoint carries its own /v2 or /v3 prefix so
		// both API surfaces are reachable.
		config.BaseURL = "https://compute.api.globus.org"
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

// --- Service-level (V2) ---

// GetVersion returns the compute service version (GET /v2/version). Pass opts to
// scope to a particular service component; opts may be nil.
func (c *Client) GetVersion(ctx context.Context, opts *GetVersionOptions) (map[string]interface{}, error) {
	query := url.Values{}
	if opts != nil && opts.Service != "" {
		query.Set("service", opts.Service)
	}
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/version", query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetResultAMQPURL returns a connection URL for the result AMQP queue
// (GET /v2/get_amqp_result_connection_url).
func (c *Client) GetResultAMQPURL(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/v2/get_amqp_result_connection_url", nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Endpoints (V2) ---

// RegisterEndpoint registers a compute endpoint (POST /v2/endpoints).
func (c *Client) RegisterEndpoint(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "endpoint document is required"}
	}
	return c.post(ctx, "/v2/endpoints", data)
}

// GetEndpoint retrieves a compute endpoint (GET /v2/endpoints/{id}).
func (c *Client) GetEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v2/endpoints/%s", endpointID), nil)
}

// GetEndpoints lists compute endpoints (GET /v2/endpoints). Pass opts.Role to
// filter (e.g. "owner", "any"); opts may be nil.
func (c *Client) GetEndpoints(ctx context.Context, opts *GetEndpointsOptions) (map[string]interface{}, error) {
	query := url.Values{}
	if opts != nil && opts.Role != "" {
		query.Set("role", opts.Role)
	}
	return c.get(ctx, "/v2/endpoints", query)
}

// GetEndpointStatus retrieves an endpoint's status (GET /v2/endpoints/{id}/status).
func (c *Client) GetEndpointStatus(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v2/endpoints/%s/status", endpointID), nil)
}

// DeleteEndpoint deletes a compute endpoint (DELETE /v2/endpoints/{id}).
func (c *Client) DeleteEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/v2/endpoints/%s", endpointID), nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LockEndpoint locks a compute endpoint (POST /v2/endpoints/{id}/lock).
func (c *Client) LockEndpoint(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.post(ctx, fmt.Sprintf("/v2/endpoints/%s/lock", endpointID), nil)
}

// --- Functions (V2) ---

// RegisterFunction registers a function (POST /v2/functions).
func (c *Client) RegisterFunction(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "function document is required"}
	}
	return c.post(ctx, "/v2/functions", data)
}

// GetFunction retrieves a registered function's metadata (GET /v2/functions/{id}).
func (c *Client) GetFunction(ctx context.Context, functionID string) (map[string]interface{}, error) {
	if functionID == "" {
		return nil, &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v2/functions/%s", functionID), nil)
}

// DeleteFunction deletes a registered function (DELETE /v2/functions/{id}).
func (c *Client) DeleteFunction(ctx context.Context, functionID string) (map[string]interface{}, error) {
	if functionID == "" {
		return nil, &core.ValidationError{Field: "functionID", Message: "function ID is required"}
	}
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/v2/functions/%s", functionID), nil, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// --- Tasks (V2) ---

// GetTask retrieves a task's status and result (GET /v2/tasks/{id}).
func (c *Client) GetTask(ctx context.Context, taskID string) (map[string]interface{}, error) {
	if taskID == "" {
		return nil, &core.ValidationError{Field: "taskID", Message: "task ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v2/tasks/%s", taskID), nil)
}

// GetTaskBatch retrieves the status of multiple tasks (POST /v2/batch_status).
func (c *Client) GetTaskBatch(ctx context.Context, taskIDs []string) (map[string]interface{}, error) {
	if len(taskIDs) == 0 {
		return nil, &core.ValidationError{Field: "taskIDs", Message: "at least one task ID is required"}
	}
	return c.post(ctx, "/v2/batch_status", map[string]interface{}{"task_ids": taskIDs})
}

// GetTaskGroup lists the task IDs for a task group (GET /v2/taskgroup/{id}).
func (c *Client) GetTaskGroup(ctx context.Context, taskGroupID string) (map[string]interface{}, error) {
	if taskGroupID == "" {
		return nil, &core.ValidationError{Field: "taskGroupID", Message: "task group ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v2/taskgroup/%s", taskGroupID), nil)
}

// Submit submits a task batch (POST /v2/submit).
func (c *Client) Submit(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "submit document is required"}
	}
	return c.post(ctx, "/v2/submit", data)
}

// --- V3 ---

// RegisterEndpointV3 registers an endpoint via the v3 API (POST /v3/endpoints).
func (c *Client) RegisterEndpointV3(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "endpoint document is required"}
	}
	return c.post(ctx, "/v3/endpoints", data)
}

// UpdateEndpointV3 updates an endpoint via the v3 API (PUT /v3/endpoints/{id}).
func (c *Client) UpdateEndpointV3(ctx context.Context, endpointID string, data map[string]interface{}) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "endpoint document is required"}
	}
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/v3/endpoints/%s", endpointID), nil, data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LockEndpointV3 locks an endpoint via the v3 API (POST /v3/endpoints/{id}/lock).
func (c *Client) LockEndpointV3(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.post(ctx, fmt.Sprintf("/v3/endpoints/%s/lock", endpointID), nil)
}

// GetEndpointAllowlistV3 lists an endpoint's allowed function IDs
// (GET /v3/endpoints/{id}/allowed_functions).
func (c *Client) GetEndpointAllowlistV3(ctx context.Context, endpointID string) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	return c.get(ctx, fmt.Sprintf("/v3/endpoints/%s/allowed_functions", endpointID), nil)
}

// RegisterFunctionV3 registers a function via the v3 API (POST /v3/functions).
func (c *Client) RegisterFunctionV3(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "function document is required"}
	}
	return c.post(ctx, "/v3/functions", data)
}

// SubmitV3 submits a task batch to an endpoint via the v3 API
// (POST /v3/endpoints/{id}/submit).
func (c *Client) SubmitV3(ctx context.Context, endpointID string, data map[string]interface{}) (map[string]interface{}, error) {
	if endpointID == "" {
		return nil, &core.ValidationError{Field: "endpointID", Message: "endpoint ID is required"}
	}
	if data == nil {
		return nil, &core.ValidationError{Field: "data", Message: "submit document is required"}
	}
	return c.post(ctx, fmt.Sprintf("/v3/endpoints/%s/submit", endpointID), data)
}

// --- helpers ---

func (c *Client) get(ctx context.Context, path string, query url.Values) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, path, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, path, nil, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

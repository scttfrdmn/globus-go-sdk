// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package timers

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

// DefaultBaseURL is the default base URL for the Timers service. It is the bare
// host; classic routes are under /jobs/ and timer creation is POST /v2/timer.
const DefaultBaseURL = "https://timer.automate.globus.org/"

// TimersScope is the required scope for accessing the Timers service
const TimersScope = "https://auth.globus.org/scopes/a1a171d5-48fb-4c77-a7ba-b8c628c20fd5/timers.api"

// Client provides methods for interacting with the Globus Timers service
type Client struct {
	Client *core.Client
}

// NewClient creates a new Timers client
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

// buildURLLowLevel builds a URL for the Timers API.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
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

// doRequestLowLevel performs an HTTP request and decodes the JSON response.
func (c *Client) doRequestLowLevel(ctx context.Context, method, path string, query url.Values, body, response interface{}) error {
	u := c.buildURLLowLevel(path, query)

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

	if method != http.MethodGet && response == nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

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

// CreateTimer creates a new timer (POST /v2/timer). The document is wrapped in a
// {"timer": ...} envelope. Pass a *TransferTimer or *FlowTimer (or any value that
// marshals to a valid timer document).
func (c *Client) CreateTimer(ctx context.Context, timer interface{}) (*Timer, error) {
	if timer == nil {
		return nil, fmt.Errorf("timer is required")
	}
	body := map[string]interface{}{"timer": timer}
	var result Timer
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "v2/timer", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateJob creates a timer using the legacy job document (POST /jobs/).
func (c *Client) CreateJob(ctx context.Context, job *TimerJob) (*Timer, error) {
	if job == nil {
		return nil, fmt.Errorf("job is required")
	}
	var result Timer
	if err := c.doRequestLowLevel(ctx, http.MethodPost, "jobs/", nil, job, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTimer retrieves a timer ("job") by ID (GET /jobs/{id}).
func (c *Client) GetTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}
	var timer Timer
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "jobs/"+timerID, nil, nil, &timer); err != nil {
		return nil, err
	}
	return &timer, nil
}

// UpdateTimer updates an existing timer (PATCH /jobs/{id}).
func (c *Client) UpdateTimer(ctx context.Context, timerID string, update interface{}) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update document is required")
	}
	var timer Timer
	if err := c.doRequestLowLevel(ctx, http.MethodPatch, "jobs/"+timerID, nil, update, &timer); err != nil {
		return nil, err
	}
	return &timer, nil
}

// DeleteTimer deletes a timer (DELETE /jobs/{id}).
func (c *Client) DeleteTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return fmt.Errorf("timer ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodDelete, "jobs/"+timerID, nil, nil, nil)
}

// ListTimers lists timers (GET /jobs/).
func (c *Client) ListTimers(ctx context.Context, options *ListTimersOptions) (*TimerList, error) {
	query := url.Values{}
	if options != nil {
		for k, v := range options.QueryParams {
			query.Set(k, v)
		}
	}
	var timerList TimerList
	if err := c.doRequestLowLevel(ctx, http.MethodGet, "jobs/", query, nil, &timerList); err != nil {
		return nil, err
	}
	return &timerList, nil
}

// PauseTimer pauses a timer (POST /jobs/{id}/pause).
func (c *Client) PauseTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return fmt.Errorf("timer ID is required")
	}
	return c.doRequestLowLevel(ctx, http.MethodPost, "jobs/"+timerID+"/pause", nil, nil, nil)
}

// ResumeTimer resumes a paused timer (POST /jobs/{id}/resume). When
// updateCredentials is non-nil, its value controls whether the resuming caller's
// credentials replace the timer's stored credentials.
func (c *Client) ResumeTimer(ctx context.Context, timerID string, updateCredentials *bool) error {
	if timerID == "" {
		return fmt.Errorf("timer ID is required")
	}
	var body interface{}
	if updateCredentials != nil {
		body = map[string]interface{}{"update_credentials": *updateCredentials}
	}
	return c.doRequestLowLevel(ctx, http.MethodPost, "jobs/"+timerID+"/resume", nil, body, nil)
}

// Close releases any resources held by the client, such as idle HTTP connections.
func (c *Client) Close() {
	if c.Client != nil && c.Client.HTTPClient != nil {
		if transport, ok := c.Client.HTTPClient.Transport.(interface{ CloseIdleConnections() }); ok {
			transport.CloseIdleConnections()
		}
	}
}

// FlowUserScope returns the scope string required for a TimersClient to execute a
// specific flow. Add the returned scope to your authorization request.
func FlowUserScope(flowID string) string {
	return "https://auth.globus.org/scopes/" + flowID + "/flow_" + flowID + "_user"
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Client is the v4 Timers service client with context-first design
type Client struct {
	baseClient *core.Client
	baseURL    string
}

// NewClient creates a new v4 Timers client
func NewClient(ctx context.Context, config *core.Config) (*Client, error) {
	if config.BaseURL == "" {
		config.BaseURL = "https://timer.automate.globus.org"
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

// GetTimer retrieves a timer ("job") by ID (GET /jobs/{id}).
func (c *Client) GetTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	var timer Timer
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/jobs/%s", timerID), nil, nil, &timer)
	if err != nil {
		return nil, err
	}
	return &timer, nil
}

// CreateTimer creates a new timer (POST /v2/timer). The timer document is wrapped
// in a {"timer": ...} envelope by the service contract. Pass a *TransferTimer or
// *FlowTimer (or any value that marshals to a valid timer document).
func (c *Client) CreateTimer(ctx context.Context, timer interface{}) (*Timer, error) {
	if timer == nil {
		return nil, &core.ValidationError{Field: "timer", Message: "timer is required"}
	}

	body := map[string]interface{}{"timer": timer}
	var result Timer
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/v2/timer", nil, body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateJob creates a timer using the legacy job document (POST /jobs/). Upstream
// marks this deprecated for transfer use-cases in favor of CreateTimer, but it is
// still supported for non-transfer callbacks.
func (c *Client) CreateJob(ctx context.Context, job *TimerJob) (*Timer, error) {
	if job == nil {
		return nil, &core.ValidationError{Field: "job", Message: "job is required"}
	}

	var result Timer
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/jobs/", nil, job, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTimer updates an existing timer (PATCH /jobs/{id}).
func (c *Client) UpdateTimer(ctx context.Context, timerID string, update interface{}) (*Timer, error) {
	if timerID == "" {
		return nil, &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update document is required"}
	}

	var result Timer
	err := c.baseClient.DoRequest(ctx, http.MethodPatch, fmt.Sprintf("/jobs/%s", timerID), nil, update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTimer deletes a timer (DELETE /jobs/{id}).
func (c *Client) DeleteTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/jobs/%s", timerID), nil, nil, nil)
}

// ListTimers lists timers (GET /jobs/).
func (c *Client) ListTimers(ctx context.Context, options *ListTimersOptions) (*TimerList, error) {
	query := options.toQuery()

	var timerList TimerList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/jobs/", query, nil, &timerList)
	if err != nil {
		return nil, err
	}
	return &timerList, nil
}

// PauseTimer pauses a timer (POST /jobs/{id}/pause).
func (c *Client) PauseTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/jobs/%s/pause", timerID), nil, nil, nil)
}

// ResumeTimer resumes a paused timer (POST /jobs/{id}/resume). When
// updateCredentials is non-nil, its value is sent in the request body to control
// whether the resuming caller's credentials replace the timer's stored credentials.
func (c *Client) ResumeTimer(ctx context.Context, timerID string, updateCredentials *bool) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	var body interface{}
	if updateCredentials != nil {
		body = map[string]interface{}{"update_credentials": *updateCredentials}
	}
	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/jobs/%s/resume", timerID), nil, body, nil)
}

// CreateFlowTimer is a convenience helper that builds a flow timer create document
// and submits it via CreateTimer (POST /v2/timer).
func (c *Client) CreateFlowTimer(ctx context.Context, name, flowID string, schedule *Schedule, flowInput map[string]interface{}) (*Timer, error) {
	if name == "" {
		return nil, &core.ValidationError{Field: "name", Message: "timer name is required"}
	}
	if schedule == nil {
		return nil, &core.ValidationError{Field: "schedule", Message: "schedule is required"}
	}
	return c.CreateTimer(ctx, NewFlowTimer(name, flowID, schedule, flowInput))
}

// CreateTransferTimer is a convenience helper that builds a transfer timer create
// document and submits it via CreateTimer (POST /v2/timer). transferBody is a
// TransferData document.
func (c *Client) CreateTransferTimer(ctx context.Context, name string, schedule *Schedule, transferBody map[string]interface{}) (*Timer, error) {
	if name == "" {
		return nil, &core.ValidationError{Field: "name", Message: "timer name is required"}
	}
	if schedule == nil {
		return nil, &core.ValidationError{Field: "schedule", Message: "schedule is required"}
	}
	if transferBody == nil {
		return nil, &core.ValidationError{Field: "transferBody", Message: "transfer body is required"}
	}
	return c.CreateTimer(ctx, NewTransferTimer(name, schedule, transferBody))
}

// Close closes the client and releases resources
func (c *Client) Close() error {
	return c.baseClient.Close()
}

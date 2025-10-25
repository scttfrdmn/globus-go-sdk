// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

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
		config.BaseURL = "https://timer.automate.globus.org/api/v1"
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

// GetTimer retrieves a timer by ID
func (c *Client) GetTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	var timer Timer
	err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/timers/%s", timerID), nil, nil, &timer)
	if err != nil {
		return nil, err
	}
	return &timer, nil
}

// CreateTimer creates a new timer
func (c *Client) CreateTimer(ctx context.Context, timer *Timer) (*Timer, error) {
	if timer == nil {
		return nil, &core.ValidationError{Field: "timer", Message: "timer is required"}
	}

	var result Timer
	err := c.baseClient.DoRequest(ctx, http.MethodPost, "/timers", nil, timer, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateTimer updates an existing timer
func (c *Client) UpdateTimer(ctx context.Context, timerID string, timer *Timer) (*Timer, error) {
	if timerID == "" {
		return nil, &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}
	if timer == nil {
		return nil, &core.ValidationError{Field: "timer", Message: "timer is required"}
	}

	var result Timer
	err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/timers/%s", timerID), nil, timer, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTimer deletes a timer
func (c *Client) DeleteTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/timers/%s", timerID), nil, nil, nil)
}

// ListTimers lists timers
func (c *Client) ListTimers(ctx context.Context, options *ListTimersOptions) (*TimerList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit > 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset > 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}

	var timerList TimerList
	err := c.baseClient.DoRequest(ctx, http.MethodGet, "/timers", query, nil, &timerList)
	if err != nil {
		return nil, err
	}
	return &timerList, nil
}

// PauseTimer pauses a timer
func (c *Client) PauseTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/timers/%s/pause", timerID), nil, nil, nil)
}

// ResumeTimer resumes a paused timer
func (c *Client) ResumeTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return &core.ValidationError{Field: "timerID", Message: "timer ID is required"}
	}

	return c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/timers/%s/resume", timerID), nil, nil, nil)
}

// CreateFlowTimer creates a timer that runs a Globus Flow (v4 helper matching v3.65.0)
func (c *Client) CreateFlowTimer(ctx context.Context, name string, schedule *Schedule, flowID, flowScope string, flowInput map[string]interface{}) (*Timer, error) {
	timer := &Timer{
		Name:     name,
		Schedule: schedule,
		Callback: &Callback{
			Type: "flow",
			URL:  fmt.Sprintf("https://flows.globus.org/flows/%s", flowID),
			Body: map[string]interface{}{
				"input": flowInput,
			},
			Scope: flowScope,
		},
	}

	return c.CreateTimer(ctx, timer)
}

// CreateOnceTimer creates a one-time timer (helper)
func (c *Client) CreateOnceTimer(ctx context.Context, name string, startTime time.Time, callback *Callback) (*Timer, error) {
	timer := &Timer{
		Name: name,
		Schedule: &Schedule{
			Type:      "once",
			StartTime: startTime,
		},
		Callback: callback,
	}

	return c.CreateTimer(ctx, timer)
}

// CreateRecurringTimer creates a recurring timer (helper)
func (c *Client) CreateRecurringTimer(ctx context.Context, name string, startTime time.Time, interval string, callback *Callback) (*Timer, error) {
	timer := &Timer{
		Name: name,
		Schedule: &Schedule{
			Type:      "recurring",
			StartTime: startTime,
			Interval:  interval,
		},
		Callback: callback,
	}

	return c.CreateTimer(ctx, timer)
}

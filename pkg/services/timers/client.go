// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package timers

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

// DefaultBaseURL is the default base URL for the Timers service
const DefaultBaseURL = "https://timer.automate.globus.org/api/v1/"

// TimersScope is the required scope for accessing the Timers service
const TimersScope = "https://auth.globus.org/scopes/a1a171d5-48fb-4c77-a7ba-b8c628c20fd5/timers.api"

// Client provides methods for interacting with the Globus Timers service
type Client struct {
	Client *core.Client
}

// NewClient creates a new Timers client
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

// buildURLLowLevel builds a URL for the Timers API
// This is an internal method used by the client.
func (c *Client) buildURLLowLevel(path string, query url.Values) string {
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

// CreateTimer creates a new timer
func (c *Client) CreateTimer(ctx context.Context, request *CreateTimerRequest) (*Timer, error) {
	if request == nil {
		return nil, fmt.Errorf("request is required")
	}

	var timer Timer
	err := c.doRequestLowLevel(ctx, http.MethodPost, "timers", nil, request, &timer)
	if err != nil {
		return nil, err
	}

	return &timer, nil
}

// GetTimer retrieves a timer by ID
func (c *Client) GetTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}

	var timer Timer
	err := c.doRequestLowLevel(ctx, http.MethodGet, fmt.Sprintf("timers/%s", timerID), nil, nil, &timer)
	if err != nil {
		return nil, err
	}

	return &timer, nil
}

// UpdateTimer updates an existing timer
func (c *Client) UpdateTimer(ctx context.Context, timerID string, request *UpdateTimerRequest) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}
	if request == nil {
		return nil, fmt.Errorf("request is required")
	}

	var timer Timer
	err := c.doRequestLowLevel(ctx, http.MethodPatch, fmt.Sprintf("timers/%s", timerID), nil, request, &timer)
	if err != nil {
		return nil, err
	}

	return &timer, nil
}

// DeleteTimer deletes a timer
func (c *Client) DeleteTimer(ctx context.Context, timerID string) error {
	if timerID == "" {
		return fmt.Errorf("timer ID is required")
	}

	err := c.doRequestLowLevel(ctx, http.MethodDelete, fmt.Sprintf("timers/%s", timerID), nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// ListTimers retrieves a list of timers
func (c *Client) ListTimers(ctx context.Context, options *ListTimersOptions) (*TimerList, error) {
	query := url.Values{}
	if options != nil {
		if options.Limit != nil {
			query.Set("limit", strconv.Itoa(*options.Limit))
		}
		if options.Marker != nil {
			query.Set("marker", *options.Marker)
		}
		if options.Status != nil {
			query.Set("status", *options.Status)
		}
		if options.ScheduleType != nil {
			query.Set("schedule_type", *options.ScheduleType)
		}
		if options.CallbackType != nil {
			query.Set("callback_type", *options.CallbackType)
		}
	}

	var timerList TimerList
	err := c.doRequestLowLevel(ctx, http.MethodGet, "timers", query, nil, &timerList)
	if err != nil {
		return nil, err
	}

	return &timerList, nil
}

// PauseTimer pauses a timer
func (c *Client) PauseTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}

	var timer Timer
	err := c.doRequestLowLevel(ctx, http.MethodPost, fmt.Sprintf("timers/%s/pause", timerID), nil, nil, &timer)
	if err != nil {
		return nil, err
	}

	return &timer, nil
}

// ResumeTimer resumes a paused timer
func (c *Client) ResumeTimer(ctx context.Context, timerID string) (*Timer, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}

	var timer Timer
	err := c.doRequestLowLevel(ctx, http.MethodPost, fmt.Sprintf("timers/%s/resume", timerID), nil, nil, &timer)
	if err != nil {
		return nil, err
	}

	return &timer, nil
}

// RunTimer manually triggers a timer run
func (c *Client) RunTimer(ctx context.Context, timerID string) (*TimerRun, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}

	var run TimerRun
	err := c.doRequestLowLevel(ctx, http.MethodPost, fmt.Sprintf("timers/%s/run", timerID), nil, nil, &run)
	if err != nil {
		return nil, err
	}

	return &run, nil
}

// ListRuns retrieves a list of runs for a timer
func (c *Client) ListRuns(ctx context.Context, timerID string, options *ListRunsOptions) (*TimerRunList, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}

	query := url.Values{}
	if options != nil {
		if options.Limit != nil {
			query.Set("limit", strconv.Itoa(*options.Limit))
		}
		if options.Marker != nil {
			query.Set("marker", *options.Marker)
		}
		if options.Status != nil {
			query.Set("status", *options.Status)
		}
		if options.StartAfter != nil {
			query.Set("start_after", options.StartAfter.Format(http.TimeFormat))
		}
		if options.StartBefore != nil {
			query.Set("start_before", options.StartBefore.Format(http.TimeFormat))
		}
	}

	var runList TimerRunList
	err := c.doRequestLowLevel(ctx, http.MethodGet, fmt.Sprintf("timers/%s/runs", timerID), query, nil, &runList)
	if err != nil {
		return nil, err
	}

	return &runList, nil
}

// GetRun retrieves a specific run
func (c *Client) GetRun(ctx context.Context, timerID, runID string) (*TimerRun, error) {
	if timerID == "" {
		return nil, fmt.Errorf("timer ID is required")
	}
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	var run TimerRun
	err := c.doRequestLowLevel(ctx, http.MethodGet, fmt.Sprintf("timers/%s/runs/%s", timerID, runID), nil, nil, &run)
	if err != nil {
		return nil, err
	}

	return &run, nil
}

// GetCurrentUser retrieves information about the current user
func (c *Client) GetCurrentUser(ctx context.Context) (*CurrentUserInfo, error) {
	var user CurrentUserInfo
	err := c.doRequestLowLevel(ctx, http.MethodGet, "user", nil, nil, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Helper functions for creating common timer types

// CreateOnceTimer creates a timer that runs once at a specific time
func (c *Client) CreateOnceTimer(
	ctx context.Context,
	name string,
	startTime time.Time,
	callback Callback,
	data map[string]interface{},
) (*Timer, error) {
	schedule := Schedule{
		Type:      string(ScheduleTypeOnce),
		StartTime: &startTime,
	}

	request := &CreateTimerRequest{
		Name:     name,
		Schedule: schedule,
		Callback: callback,
		Data:     data,
	}

	return c.CreateTimer(ctx, request)
}

// CreateRecurringTimer creates a timer that runs at a regular interval
func (c *Client) CreateRecurringTimer(
	ctx context.Context,
	name string,
	startTime time.Time,
	interval string,
	endTime *time.Time,
	callback Callback,
	data map[string]interface{},
) (*Timer, error) {
	schedule := Schedule{
		Type:      string(ScheduleTypeRecurring),
		StartTime: &startTime,
		EndTime:   endTime,
		Interval:  &interval,
	}

	request := &CreateTimerRequest{
		Name:     name,
		Schedule: schedule,
		Callback: callback,
		Data:     data,
	}

	return c.CreateTimer(ctx, request)
}

// CreateCronTimer creates a timer that runs on a cron schedule
func (c *Client) CreateCronTimer(
	ctx context.Context,
	name string,
	cronExpression string,
	timezone string,
	endTime *time.Time,
	callback Callback,
	data map[string]interface{},
) (*Timer, error) {
	schedule := Schedule{
		Type:           string(ScheduleTypeCron),
		CronExpression: &cronExpression,
		Timezone:       &timezone,
		EndTime:        endTime,
	}

	request := &CreateTimerRequest{
		Name:     name,
		Schedule: schedule,
		Callback: callback,
		Data:     data,
	}

	return c.CreateTimer(ctx, request)
}

// CreateFlowCallback creates a callback configuration for triggering a flow
func CreateFlowCallback(flowID, flowLabel string, flowInput map[string]interface{}) Callback {
	return Callback{
		Type:      string(CallbackTypeFlow),
		FlowID:    &flowID,
		FlowLabel: &flowLabel,
		FlowInput: flowInput,
	}
}

// CreateWebCallback creates a callback configuration for making a web request
func CreateWebCallback(url, method string, headers map[string]string, body *string) Callback {
	return Callback{
		Type:    string(CallbackTypeWeb),
		URL:     &url,
		Method:  &method,
		Headers: headers,
		Body:    body,
	}
}

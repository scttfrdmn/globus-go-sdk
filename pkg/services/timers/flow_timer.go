// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package timers

import (
	"context"
	"fmt"
	"time"
)

// FlowTimer represents a timer configuration for running Globus Flows
// Added in v3.65.0 to match Python SDK v3.65.0
//
// This is a convenience helper that simplifies creating timers that run flows,
// similar to the FlowTimer payload class in the Python SDK.
type FlowTimer struct {
	// FlowID is the unique identifier of the flow to run
	FlowID string `json:"flow_id"`

	// FlowScope is the scope required to run the flow
	FlowScope string `json:"flow_scope"`

	// FlowInput contains the input data to pass to the flow
	FlowInput map[string]interface{} `json:"flow_input,omitempty"`

	// FlowLabel is an optional label for the flow run
	FlowLabel string `json:"flow_label,omitempty"`
}

// CreateFlowTimerOnce creates a one-time timer that runs a Globus Flow
//
// This is a convenience method added in v3.65.0 that combines FlowTimer
// with a once schedule, matching the Python SDK's FlowTimer functionality.
//
// Parameters:
//   - ctx: Context for the request
//   - name: User-provided name for the timer
//   - startTime: When the timer should run
//   - flowTimer: FlowTimer configuration specifying which flow to run
//   - data: Optional additional user data for the timer
//
// Returns the created Timer or an error
func (c *Client) CreateFlowTimerOnce(
	ctx context.Context,
	name string,
	startTime time.Time,
	flowTimer *FlowTimer,
	data map[string]interface{},
) (*Timer, error) {
	if flowTimer == nil {
		return nil, fmt.Errorf("flowTimer is required")
	}
	if flowTimer.FlowID == "" {
		return nil, fmt.Errorf("flowTimer.FlowID is required")
	}
	if flowTimer.FlowScope == "" {
		return nil, fmt.Errorf("flowTimer.FlowScope is required")
	}

	callback := Callback{
		Type:      string(CallbackTypeFlow),
		FlowID:    &flowTimer.FlowID,
		FlowInput: flowTimer.FlowInput,
	}
	if flowTimer.FlowLabel != "" {
		callback.FlowLabel = &flowTimer.FlowLabel
	}

	// Add flow scope to data if not already present
	if data == nil {
		data = make(map[string]interface{})
	}
	if _, exists := data["flow_scope"]; !exists {
		data["flow_scope"] = flowTimer.FlowScope
	}

	return c.CreateOnceTimer(ctx, name, startTime, callback, data)
}

// CreateFlowTimerRecurring creates a recurring timer that runs a Globus Flow
//
// This is a convenience method added in v3.65.0 that combines FlowTimer
// with a recurring schedule, matching the Python SDK's FlowTimer functionality.
//
// Parameters:
//   - ctx: Context for the request
//   - name: User-provided name for the timer
//   - startTime: When the timer should start running
//   - interval: ISO 8601 duration string for the interval (e.g., "P1D" for daily)
//   - endTime: Optional time when the timer should stop (nil for no end)
//   - flowTimer: FlowTimer configuration specifying which flow to run
//   - data: Optional additional user data for the timer
//
// Returns the created Timer or an error
func (c *Client) CreateFlowTimerRecurring(
	ctx context.Context,
	name string,
	startTime time.Time,
	interval string,
	endTime *time.Time,
	flowTimer *FlowTimer,
	data map[string]interface{},
) (*Timer, error) {
	if flowTimer == nil {
		return nil, fmt.Errorf("flowTimer is required")
	}
	if flowTimer.FlowID == "" {
		return nil, fmt.Errorf("flowTimer.FlowID is required")
	}
	if flowTimer.FlowScope == "" {
		return nil, fmt.Errorf("flowTimer.FlowScope is required")
	}

	callback := Callback{
		Type:      string(CallbackTypeFlow),
		FlowID:    &flowTimer.FlowID,
		FlowInput: flowTimer.FlowInput,
	}
	if flowTimer.FlowLabel != "" {
		callback.FlowLabel = &flowTimer.FlowLabel
	}

	// Add flow scope to data if not already present
	if data == nil {
		data = make(map[string]interface{})
	}
	if _, exists := data["flow_scope"]; !exists {
		data["flow_scope"] = flowTimer.FlowScope
	}

	return c.CreateRecurringTimer(ctx, name, startTime, interval, endTime, callback, data)
}

// CreateFlowTimerCron creates a cron-scheduled timer that runs a Globus Flow
//
// This is a convenience method added in v3.65.0 that combines FlowTimer
// with a cron schedule, matching the Python SDK's FlowTimer functionality.
//
// Parameters:
//   - ctx: Context for the request
//   - name: User-provided name for the timer
//   - cronExpression: Cron expression defining when the timer runs (e.g., "0 0 * * *" for daily at midnight)
//   - timezone: Timezone for the cron schedule (e.g., "America/New_York")
//   - endTime: Optional time when the timer should stop (nil for no end)
//   - flowTimer: FlowTimer configuration specifying which flow to run
//   - data: Optional additional user data for the timer
//
// Returns the created Timer or an error
func (c *Client) CreateFlowTimerCron(
	ctx context.Context,
	name string,
	cronExpression string,
	timezone string,
	endTime *time.Time,
	flowTimer *FlowTimer,
	data map[string]interface{},
) (*Timer, error) {
	if flowTimer == nil {
		return nil, fmt.Errorf("flowTimer is required")
	}
	if flowTimer.FlowID == "" {
		return nil, fmt.Errorf("flowTimer.FlowID is required")
	}
	if flowTimer.FlowScope == "" {
		return nil, fmt.Errorf("flowTimer.FlowScope is required")
	}

	callback := Callback{
		Type:      string(CallbackTypeFlow),
		FlowID:    &flowTimer.FlowID,
		FlowInput: flowTimer.FlowInput,
	}
	if flowTimer.FlowLabel != "" {
		callback.FlowLabel = &flowTimer.FlowLabel
	}

	// Add flow scope to data if not already present
	if data == nil {
		data = make(map[string]interface{})
	}
	if _, exists := data["flow_scope"]; !exists {
		data["flow_scope"] = flowTimer.FlowScope
	}

	return c.CreateCronTimer(ctx, name, cronExpression, timezone, endTime, callback, data)
}

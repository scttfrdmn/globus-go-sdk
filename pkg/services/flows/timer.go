// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package flows

import (
	"time"
)

// FlowTimer represents a timer payload for scheduled flow execution
// v3.65.0: Added FlowTimer payload class for integration with Globus Timers
//
// FlowTimer provides a structured way to define scheduled flow executions
// using the Globus Timers service. This allows flows to be triggered
// automatically based on time-based schedules.
type FlowTimer struct {
	// Name is a human-readable name for the timer
	Name string `json:"name"`

	// FlowID is the UUID of the flow to execute
	FlowID string `json:"flow_id"`

	// FlowInput is the input data to pass to the flow when it executes
	FlowInput map[string]interface{} `json:"flow_input"`

	// Schedule defines when the flow should be triggered
	Schedule TimerSchedule `json:"schedule"`

	// CallbackURL is an optional URL to call when the flow completes
	// If provided, the Globus Flows service will POST the flow result to this URL
	CallbackURL string `json:"callback_url,omitempty"`

	// FlowScope is the scope required to run the flow
	// If not specified, defaults to the flow's required scope
	FlowScope string `json:"flow_scope,omitempty"`

	// RunManagers are optional identity IDs that can manage flow runs
	RunManagers []string `json:"run_managers,omitempty"`

	// RunMonitors are optional identity IDs that can monitor flow runs
	RunMonitors []string `json:"run_monitors,omitempty"`
}

// TimerSchedule defines when a flow should be executed
// v3.65.0: Schedule definition for FlowTimer
type TimerSchedule struct {
	// Type is the schedule type: "cron", "interval", or "once"
	Type string `json:"type"`

	// Value is the schedule value
	// - For "cron": a cron expression (e.g., "0 0 * * *" for daily at midnight)
	// - For "interval": an ISO 8601 duration (e.g., "PT1H" for every hour)
	// - For "once": an ISO 8601 timestamp (e.g., "2025-01-01T00:00:00Z")
	Value string `json:"value"`

	// Timezone is the IANA timezone name (e.g., "America/New_York")
	// Only applicable for "cron" schedules
	// Defaults to UTC if not specified
	Timezone string `json:"timezone,omitempty"`

	// StartTime is an optional start time for the schedule (ISO 8601 timestamp)
	// The timer will not trigger before this time
	StartTime string `json:"start_time,omitempty"`

	// EndTime is an optional end time for the schedule (ISO 8601 timestamp)
	// The timer will not trigger after this time
	EndTime string `json:"end_time,omitempty"`
}

// NewFlowTimer creates a new FlowTimer with the specified parameters
// v3.65.0: Convenience constructor for FlowTimer
func NewFlowTimer(name, flowID string, flowInput map[string]interface{}, schedule TimerSchedule) *FlowTimer {
	return &FlowTimer{
		Name:      name,
		FlowID:    flowID,
		FlowInput: flowInput,
		Schedule:  schedule,
	}
}

// WithCallbackURL sets the callback URL for the FlowTimer
func (ft *FlowTimer) WithCallbackURL(url string) *FlowTimer {
	ft.CallbackURL = url
	return ft
}

// WithFlowScope sets the flow scope for the FlowTimer
func (ft *FlowTimer) WithFlowScope(scope string) *FlowTimer {
	ft.FlowScope = scope
	return ft
}

// WithRunManagers sets the run managers for the FlowTimer
func (ft *FlowTimer) WithRunManagers(managers []string) *FlowTimer {
	ft.RunManagers = managers
	return ft
}

// WithRunMonitors sets the run monitors for the FlowTimer
func (ft *FlowTimer) WithRunMonitors(monitors []string) *FlowTimer {
	ft.RunMonitors = monitors
	return ft
}

// NewCronSchedule creates a cron-based schedule
// v3.65.0: Convenience constructor for cron schedules
//
// Example: NewCronSchedule("0 0 * * *", "America/New_York") // Daily at midnight EST
func NewCronSchedule(cronExpression, timezone string) TimerSchedule {
	return TimerSchedule{
		Type:     "cron",
		Value:    cronExpression,
		Timezone: timezone,
	}
}

// NewIntervalSchedule creates an interval-based schedule
// v3.65.0: Convenience constructor for interval schedules
//
// Example: NewIntervalSchedule("PT1H") // Every hour
func NewIntervalSchedule(duration string) TimerSchedule {
	return TimerSchedule{
		Type:  "interval",
		Value: duration,
	}
}

// NewOnceSchedule creates a one-time schedule
// v3.65.0: Convenience constructor for one-time schedules
//
// Example: NewOnceSchedule(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
func NewOnceSchedule(when time.Time) TimerSchedule {
	return TimerSchedule{
		Type:  "once",
		Value: when.Format(time.RFC3339),
	}
}

// WithStartTime sets the start time for the schedule
func (ts *TimerSchedule) WithStartTime(t time.Time) *TimerSchedule {
	ts.StartTime = t.Format(time.RFC3339)
	return ts
}

// WithEndTime sets the end time for the schedule
func (ts *TimerSchedule) WithEndTime(t time.Time) *TimerSchedule {
	ts.EndTime = t.Format(time.RFC3339)
	return ts
}

// Validate checks if the FlowTimer is valid
func (ft *FlowTimer) Validate() error {
	if ft.Name == "" {
		return ErrInvalidFlowTimer{Field: "name", Message: "name is required"}
	}
	if ft.FlowID == "" {
		return ErrInvalidFlowTimer{Field: "flow_id", Message: "flow_id is required"}
	}
	if ft.Schedule.Type == "" {
		return ErrInvalidFlowTimer{Field: "schedule.type", Message: "schedule type is required"}
	}
	if ft.Schedule.Value == "" {
		return ErrInvalidFlowTimer{Field: "schedule.value", Message: "schedule value is required"}
	}

	// Validate schedule type
	switch ft.Schedule.Type {
	case "cron", "interval", "once":
		// Valid types
	default:
		return ErrInvalidFlowTimer{
			Field:   "schedule.type",
			Message: "schedule type must be 'cron', 'interval', or 'once'",
		}
	}

	return nil
}

// ErrInvalidFlowTimer is returned when a FlowTimer is invalid
type ErrInvalidFlowTimer struct {
	Field   string
	Message string
}

func (e ErrInvalidFlowTimer) Error() string {
	return "invalid flow timer: " + e.Field + ": " + e.Message
}

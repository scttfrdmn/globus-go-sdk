// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package timers

import (
	"time"
)

// Timer-related constants are defined in client.go

// Timer represents a Globus timer
type Timer struct {
	// ID is the unique identifier for the timer
	ID string `json:"id"`

	// Name is the user-provided name for the timer
	Name string `json:"name"`

	// Owner is the ID of the user who created the timer
	Owner string `json:"owner"`

	// Schedule defines when the timer should run
	Schedule *Schedule `json:"schedule"`

	// Callback defines what the timer should do when triggered
	Callback *Callback `json:"callback"`

	// Status indicates the current status of the timer
	Status string `json:"status"`

	// LastUpdate is when the timer was last updated
	LastUpdate time.Time `json:"last_update"`

	// NextDue is when the timer is next scheduled to run
	NextDue *time.Time `json:"next_due,omitempty"`

	// LastRun is when the timer was last run
	LastRun *time.Time `json:"last_run,omitempty"`

	// LastRunStatus indicates the status of the last run
	LastRunStatus string `json:"last_run_status,omitempty"`

	// CreateTime is when the timer was created
	CreateTime time.Time `json:"create_time"`

	// Data contains additional user-provided data for the timer
	Data map[string]interface{} `json:"data,omitempty"`
}

// Schedule defines when a timer should run
type Schedule struct {
	// Type is the type of schedule (once, recurring, cron)
	Type string `json:"type"`

	// StartTime is when the timer should start
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime is when the timer should stop
	EndTime *time.Time `json:"end_time,omitempty"`

	// Interval is the interval for recurring timers
	Interval *string `json:"interval,omitempty"`

	// CronExpression is the cron expression for cron-based timers
	CronExpression *string `json:"cron_expression,omitempty"`

	// Timezone is the timezone for the schedule
	Timezone *string `json:"timezone,omitempty"`
}

// Callback defines what a timer should do when triggered
type Callback struct {
	// Type is the type of callback (flow, web)
	Type string `json:"type"`

	// URL is the URL to call for web callbacks
	URL *string `json:"url,omitempty"`

	// Method is the HTTP method to use for web callbacks
	Method *string `json:"method,omitempty"`

	// FlowID is the ID of the flow to run for flow callbacks
	FlowID *string `json:"flow_id,omitempty"`

	// FlowLabel is the label to use for flow callbacks
	FlowLabel *string `json:"flow_label,omitempty"`

	// FlowInput is the input data for flow callbacks
	FlowInput map[string]interface{} `json:"flow_input,omitempty"`

	// Headers are the HTTP headers to use for web callbacks
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the HTTP body to use for web callbacks
	Body *string `json:"body,omitempty"`
}

// TimerList represents a list of timers
type TimerList struct {
	// Timers is the list of timers
	Timers []Timer `json:"timers"`

	// Total is the total number of timers available
	Total int `json:"total"`

	// HasNextPage indicates whether there are more timers to retrieve
	HasNextPage bool `json:"has_next_page"`

	// NextPage is the marker for retrieving the next page of timers
	NextPage *string `json:"next_page,omitempty"`
}

// TimerRunList represents a list of timer runs
type TimerRunList struct {
	// Runs is the list of timer runs
	Runs []TimerRun `json:"runs"`

	// Total is the total number of runs available
	Total int `json:"total"`

	// HasNextPage indicates whether there are more runs to retrieve
	HasNextPage bool `json:"has_next_page"`

	// NextPage is the marker for retrieving the next page of runs
	NextPage *string `json:"next_page,omitempty"`
}

// TimerRun represents a single run of a timer
type TimerRun struct {
	// ID is the unique identifier for the run
	ID string `json:"id"`

	// TimerID is the ID of the timer that was run
	TimerID string `json:"timer_id"`

	// Status is the status of the run
	Status string `json:"status"`

	// StartTime is when the run started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the run ended
	EndTime *time.Time `json:"end_time,omitempty"`

	// Result contains the result of the run
	Result *RunResult `json:"result,omitempty"`
}

// RunResult represents the result of a timer run
type RunResult struct {
	// Status is the status of the callback execution
	Status string `json:"status"`

	// StatusCode is the HTTP status code for web callbacks
	StatusCode *int `json:"status_code,omitempty"`

	// RunID is the ID of the flow run for flow callbacks
	RunID *string `json:"run_id,omitempty"`

	// Error contains error information if the run failed
	Error *RunError `json:"error,omitempty"`
}

// RunError represents an error that occurred during a timer run
type RunError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Detail contains additional error details
	Detail map[string]interface{} `json:"detail,omitempty"`
}

// CreateTimerRequest represents a request to create a new timer
type CreateTimerRequest struct {
	// Name is the user-provided name for the timer
	Name string `json:"name"`

	// Schedule defines when the timer should run
	Schedule Schedule `json:"schedule"`

	// Callback defines what the timer should do when triggered
	Callback Callback `json:"callback"`

	// Data contains additional user-provided data for the timer
	Data map[string]interface{} `json:"data,omitempty"`
}

// UpdateTimerRequest represents a request to update an existing timer
type UpdateTimerRequest struct {
	// Name is the user-provided name for the timer
	Name *string `json:"name,omitempty"`

	// Schedule defines when the timer should run
	Schedule *Schedule `json:"schedule,omitempty"`

	// Callback defines what the timer should do when triggered
	Callback *Callback `json:"callback,omitempty"`

	// Data contains additional user-provided data for the timer
	Data map[string]interface{} `json:"data,omitempty"`
}

// ListTimersOptions represents options for listing timers
type ListTimersOptions struct {
	// Limit is the maximum number of timers to return
	Limit *int `url:"limit,omitempty"`

	// Marker is the marker for pagination
	Marker *string `url:"marker,omitempty"`

	// Status filters timers by status
	Status *string `url:"status,omitempty"`

	// ScheduleType filters timers by schedule type
	ScheduleType *string `url:"schedule_type,omitempty"`

	// CallbackType filters timers by callback type
	CallbackType *string `url:"callback_type,omitempty"`
}

// ListRunsOptions represents options for listing timer runs
type ListRunsOptions struct {
	// Limit is the maximum number of runs to return
	Limit *int `url:"limit,omitempty"`

	// Marker is the marker for pagination
	Marker *string `url:"marker,omitempty"`

	// Status filters runs by status
	Status *string `url:"status,omitempty"`

	// StartAfter filters runs that started after this time
	StartAfter *time.Time `url:"start_after,omitempty"`

	// StartBefore filters runs that started before this time
	StartBefore *time.Time `url:"start_before,omitempty"`
}

// CurrentUserInfo represents information about the current user
type CurrentUserInfo struct {
	// ID is the user's ID
	ID string `json:"id"`

	// Username is the user's username
	Username string `json:"username"`

	// Email is the user's email address
	Email string `json:"email"`

	// Name is the user's display name
	Name *string `json:"name,omitempty"`
}

// TimerStatus represents the possible statuses of a timer
type TimerStatus string

const (
	// TimerStatusActive indicates the timer is active
	TimerStatusActive TimerStatus = "active"

	// TimerStatusPaused indicates the timer is paused
	TimerStatusPaused TimerStatus = "paused"

	// TimerStatusExpired indicates the timer has expired
	TimerStatusExpired TimerStatus = "expired"

	// TimerStatusFailed indicates the timer has failed
	TimerStatusFailed TimerStatus = "failed"

	// TimerStatusComplete indicates the timer has completed
	TimerStatusComplete TimerStatus = "complete"
)

// ScheduleType represents the possible types of timer schedules
type ScheduleType string

const (
	// ScheduleTypeOnce indicates the timer should run once
	ScheduleTypeOnce ScheduleType = "once"

	// ScheduleTypeRecurring indicates the timer should run recurringly
	ScheduleTypeRecurring ScheduleType = "recurring"

	// ScheduleTypeCron indicates the timer should run on a cron schedule
	ScheduleTypeCron ScheduleType = "cron"
)

// CallbackType represents the possible types of timer callbacks
type CallbackType string

const (
	// CallbackTypeFlow indicates the timer should run a flow
	CallbackTypeFlow CallbackType = "flow"

	// CallbackTypeWeb indicates the timer should make a web request
	CallbackTypeWeb CallbackType = "web"
)

// RunStatus represents the possible statuses of a timer run
type RunStatus string

const (
	// RunStatusPending indicates the run is pending
	RunStatusPending RunStatus = "pending"

	// RunStatusInProgress indicates the run is in progress
	RunStatusInProgress RunStatus = "in_progress"

	// RunStatusSuccess indicates the run succeeded
	RunStatusSuccess RunStatus = "success"

	// RunStatusFailure indicates the run failed
	RunStatusFailure RunStatus = "failure"
)
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package timers

import "time"

// Timer represents a Globus Timers "job" as returned by the service. The Timers
// API returns an untyped document; commonly used fields are typed here and
// unrecognized fields are ignored on decode.
type Timer struct {
	JobID    string    `json:"job_id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Status   string    `json:"status,omitempty"`
	Schedule *Schedule `json:"schedule,omitempty"`
	Created  time.Time `json:"submitted_at,omitempty"`
	LastRun  time.Time `json:"last_ran_at,omitempty"`
	NextRun  time.Time `json:"next_run,omitempty"`
}

// Schedule describes when a timer runs. It serializes to either a "once" or a
// "recurring" schedule per the upstream OnceTimerSchedule / RecurringTimerSchedule.
type Schedule struct {
	Type string `json:"type"` // "once" | "recurring"

	// Once schedules.
	Datetime string `json:"datetime,omitempty"`

	// Recurring schedules.
	IntervalSeconds int          `json:"interval_seconds,omitempty"`
	Start           string       `json:"start,omitempty"`
	End             *ScheduleEnd `json:"end,omitempty"`
}

// ScheduleEnd is the optional end condition for a recurring schedule. Exactly one
// of Iterations (condition "iterations") or Datetime (condition "time") applies.
type ScheduleEnd struct {
	Condition  string `json:"condition"` // "iterations" | "time"
	Iterations int    `json:"iterations,omitempty"`
	Datetime   string `json:"datetime,omitempty"`
}

// NewOnceSchedule builds a schedule that runs once at datetime (ISO 8601, or ""
// to run as soon as possible).
func NewOnceSchedule(datetime string) *Schedule {
	return &Schedule{Type: "once", Datetime: datetime}
}

// NewRecurringSchedule builds a schedule that runs every intervalSeconds.
func NewRecurringSchedule(intervalSeconds int, start string, end *ScheduleEnd) *Schedule {
	return &Schedule{Type: "recurring", IntervalSeconds: intervalSeconds, Start: start, End: end}
}

// TransferTimer is the create document for a data-transfer timer, posted to
// POST /v2/timer via CreateTimer. Body is a preprocessed TransferData document.
type TransferTimer struct {
	TimerType string                 `json:"timer_type"` // always "transfer"
	Name      string                 `json:"name,omitempty"`
	Schedule  *Schedule              `json:"schedule"`
	Body      map[string]interface{} `json:"body"`
}

// NewTransferTimer builds a transfer timer create document.
func NewTransferTimer(name string, schedule *Schedule, body map[string]interface{}) *TransferTimer {
	return &TransferTimer{TimerType: "transfer", Name: name, Schedule: schedule, Body: body}
}

// FlowTimer is the create document for a flow timer, posted to POST /v2/timer via
// CreateTimer. Body is the flow run input document.
type FlowTimer struct {
	TimerType string                 `json:"timer_type"` // always "flow"
	FlowID    string                 `json:"flow_id"`
	Name      string                 `json:"name,omitempty"`
	Schedule  *Schedule              `json:"schedule"`
	Body      map[string]interface{} `json:"body"`
}

// NewFlowTimer builds a flow timer create document.
func NewFlowTimer(name, flowID string, schedule *Schedule, body map[string]interface{}) *FlowTimer {
	return &FlowTimer{TimerType: "flow", FlowID: flowID, Name: name, Schedule: schedule, Body: body}
}

// TimerJob is the legacy create document for CreateJob (POST /jobs/).
type TimerJob struct {
	CallbackURL  string                 `json:"callback_url"`
	CallbackBody map[string]interface{} `json:"callback_body"`
	Start        string                 `json:"start"`
	Interval     *int                   `json:"interval"` // seconds; nil when the job runs once
	Name         string                 `json:"name,omitempty"`
	StopAfter    string                 `json:"stop_after,omitempty"`
	StopAfterN   int                    `json:"stop_after_n,omitempty"`
	Scope        string                 `json:"scope,omitempty"`
}

// TimerList represents a list of timers as returned by GET /jobs/.
type TimerList struct {
	Timers []Timer `json:"jobs"`
}

// ListTimersOptions carries extra query parameters for ListTimers. The upstream
// list_jobs endpoint accepts arbitrary query params and is not paginated.
type ListTimersOptions struct {
	QueryParams map[string]string
}

// TimerStatus represents the possible statuses of a timer.
type TimerStatus string

const (
	TimerStatusActive   TimerStatus = "active"
	TimerStatusPaused   TimerStatus = "paused"
	TimerStatusExpired  TimerStatus = "expired"
	TimerStatusFailed   TimerStatus = "failed"
	TimerStatusComplete TimerStatus = "complete"
)

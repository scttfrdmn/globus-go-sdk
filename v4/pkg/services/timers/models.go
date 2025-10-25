// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package timers

import "time"

// Timer represents a Globus timer
type Timer struct {
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schedule    *Schedule              `json:"schedule"`
	Callback    *Callback              `json:"callback"`
	Status      string                 `json:"status,omitempty"`
	Created     time.Time              `json:"created_at,omitempty"`
	Updated     time.Time              `json:"updated_at,omitempty"`
	LastRun     time.Time              `json:"last_run_at,omitempty"`
	NextRun     time.Time              `json:"next_run_at,omitempty"`
}

// Schedule represents a timer schedule
type Schedule struct {
	Type           string    `json:"type"` // "once", "recurring", "cron"
	StartTime      time.Time `json:"start,omitempty"`
	EndTime        *time.Time `json:"end,omitempty"`
	Interval       string    `json:"interval,omitempty"`       // ISO 8601 duration for recurring
	CronExpression string    `json:"cron,omitempty"`           // Cron expression
	Timezone       string    `json:"timezone,omitempty"`       // Timezone for cron
}

// Callback represents a timer callback action
type Callback struct {
	Type  string                 `json:"type"`            // "flow", "https"
	URL   string                 `json:"url"`
	Body  map[string]interface{} `json:"body,omitempty"`
	Scope string                 `json:"scope,omitempty"` // Required scope for flow callbacks
}

// TimerList represents a list of timers
type TimerList struct {
	Timers []Timer `json:"timers"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
	Total  int     `json:"total"`
}

// ListTimersOptions contains options for listing timers
type ListTimersOptions struct {
	Limit  int
	Offset int
}

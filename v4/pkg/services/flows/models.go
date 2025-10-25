// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import "time"

// Flow represents a Globus flow
type Flow struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Definition  map[string]interface{} `json:"definition"`
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
	Created     time.Time              `json:"created_at"`
	Updated     time.Time              `json:"updated_at"`
	OwnerID     string                 `json:"owner_id"`
	Visible     bool                   `json:"visible_to,omitempty"`
}

// FlowList represents a list of flows
type FlowList struct {
	Flows  []Flow `json:"flows"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
}

// ListFlowsOptions contains options for listing flows
type ListFlowsOptions struct {
	Limit  int
	Offset int
}

// FlowInput represents input for running a flow
type FlowInput struct {
	Input map[string]interface{} `json:"input"`
	Label string                 `json:"label,omitempty"`
	Tags  []string               `json:"tags,omitempty"`
}

// FlowRun represents a flow execution
type FlowRun struct {
	RunID     string                 `json:"run_id"`
	FlowID    string                 `json:"flow_id"`
	Status    string                 `json:"status"`
	Label     string                 `json:"label,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"completion_time,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// RunList represents a list of flow runs
type RunList struct {
	Runs   []FlowRun `json:"runs"`
	Offset int       `json:"offset"`
	Limit  int       `json:"limit"`
	Total  int       `json:"total"`
}

// ListRunsOptions contains options for listing runs
type ListRunsOptions struct {
	FlowID string
	Limit  int
	Offset int
}

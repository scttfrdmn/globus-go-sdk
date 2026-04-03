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

// FlowAuthenticationPolicy specifies authentication requirements for running a flow.
// Added in Python SDK v4.1.0.
type FlowAuthenticationPolicy struct {
	HighAssurance   *bool    `json:"high_assurance,omitempty"`
	RequiredMFA     *bool    `json:"required_mfa,omitempty"`
	SessionPolicies []string `json:"session_policies,omitempty"`
}

// FlowCreate represents the payload for creating a flow
type FlowCreate struct {
	Title                string                     `json:"title"`
	Definition           map[string]interface{}     `json:"definition"`
	InputSchema          map[string]interface{}     `json:"input_schema,omitempty"`
	Description          string                     `json:"description,omitempty"`
	// AuthenticationPolicy specifies auth requirements for running this flow.
	// Added in Python SDK v4.1.0; service support may be pending.
	AuthenticationPolicy *FlowAuthenticationPolicy  `json:"authentication_policy,omitempty"`
}

// FlowUpdate represents the payload for updating a flow
type FlowUpdate struct {
	Title                string                     `json:"title,omitempty"`
	Definition           map[string]interface{}     `json:"definition,omitempty"`
	InputSchema          map[string]interface{}     `json:"input_schema,omitempty"`
	Description          string                     `json:"description,omitempty"`
	// AuthenticationPolicy specifies auth requirements for running this flow.
	// Added in Python SDK v4.1.0; service support may be pending.
	AuthenticationPolicy *FlowAuthenticationPolicy  `json:"authentication_policy,omitempty"`
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

// RunUpdate contains fields that can be changed on a running flow execution.
type RunUpdate struct {
	Label string   `json:"label,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

// RunLog is a single log entry for a flow run.
type RunLog struct {
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
	Time    time.Time              `json:"time"`
}

// RunLogList is a paginated list of log entries for a flow run.
type RunLogList struct {
	Entries []RunLog `json:"entries"`
	Total   int      `json:"total"`
	Offset  int      `json:"offset"`
	Limit   int      `json:"limit"`
}

// ListRunLogsOptions controls which run log entries are returned.
type ListRunLogsOptions struct {
	Limit  int
	Offset int
}

// ActionProvider represents a Flows action provider
type ActionProvider struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Type        string    `json:"type"`
	Globus      bool      `json:"globus"`
	Visible     bool      `json:"visible"`
}

// ActionProviderList represents a list of Flows action providers
type ActionProviderList struct {
	ActionProviders []ActionProvider `json:"action_providers"`
	Total           int              `json:"total"`
	HadMore         bool             `json:"had_more"`
	Offset          int              `json:"offset"`
	Limit           int              `json:"limit"`
}

// ActionRole represents a role in a Flows action provider
type ActionRole struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ActionFields map[string]interface{} `json:"action_fields,omitempty"`
	InputSchema  map[string]interface{} `json:"input_schema,omitempty"`
	Visible      bool                   `json:"visible"`
}

// ActionRoleList represents a list of Flows action roles
type ActionRoleList struct {
	ActionRoles []ActionRole `json:"action_roles"`
	Total       int          `json:"total"`
	HadMore     bool         `json:"had_more"`
	Offset      int          `json:"offset"`
	Limit       int          `json:"limit"`
}

// ListActionProvidersOptions controls which action providers are returned
type ListActionProvidersOptions struct {
	Limit       int
	Offset      int
	Marker      string
	OrderBy     string
	Q           string
	FilterOwner string
	FilterType  string
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package flows

import (
	"time"
)

// Flow represents a Globus Flow definition
type Flow struct {
	ID            string                 `json:"id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	FlowOwner     string                 `json:"flow_owner,omitempty"`
	SubsID        string                 `json:"subscription_id,omitempty"`
	CreatedAt     time.Time              `json:"created_at,omitempty"`
	UpdatedAt     time.Time              `json:"updated_at,omitempty"`
	Definition    map[string]interface{} `json:"definition"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	Keywords      []string               `json:"keywords,omitempty"`
	RunCount      int                    `json:"run_count,omitempty"`
	Public        bool                   `json:"public,omitempty"`
	Managed       bool                   `json:"managed,omitempty"`
	AdminOnly     bool                   `json:"admin_only,omitempty"`
	RunsRequired  bool                   `json:"runs_required,omitempty"`
	RunAsApprover bool                   `json:"run_as_approver,omitempty"`
}

// FlowList represents a list of Flows
type FlowList struct {
	Flows   []Flow `json:"flows"`
	Total   int    `json:"total"`
	HadMore bool   `json:"had_more"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// FlowAuthenticationPolicy represents authentication policy parameters for a Flow.
// Added in Python SDK v4.1.0.
type FlowAuthenticationPolicy struct {
	// HighAssurance requires high-assurance authentication for flow runs
	HighAssurance *bool `json:"high_assurance,omitempty"`
	// RequiredMFA requires multi-factor authentication for flow runs
	RequiredMFA *bool `json:"required_mfa,omitempty"`
	// SessionPolicies specifies named authentication policies required for flow runs
	SessionPolicies []string `json:"session_policies,omitempty"`
}

// FlowCreateRequest represents a request to create a new Flow
type FlowCreateRequest struct {
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Definition    map[string]interface{} `json:"definition"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	Keywords      []string               `json:"keywords,omitempty"`
	Public        bool                   `json:"public,omitempty"`
	Managed       bool                   `json:"managed,omitempty"`
	AdminOnly     bool                   `json:"admin_only,omitempty"`
	RunsRequired  bool                   `json:"runs_required,omitempty"`
	RunAsApprover bool                   `json:"run_as_approver,omitempty"`
	// AuthenticationPolicy specifies authentication requirements for running this flow.
	// Added in Python SDK v4.1.0; service support may be pending.
	AuthenticationPolicy *FlowAuthenticationPolicy `json:"authentication_policy,omitempty"`
}

// FlowUpdateRequest represents a request to update a Flow
type FlowUpdateRequest struct {
	Title         string                 `json:"title,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Definition    map[string]interface{} `json:"definition,omitempty"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	Keywords      []string               `json:"keywords,omitempty"`
	Public        *bool                  `json:"public,omitempty"`
	Managed       *bool                  `json:"managed,omitempty"`
	AdminOnly     *bool                  `json:"admin_only,omitempty"`
	RunsRequired  *bool                  `json:"runs_required,omitempty"`
	RunAsApprover *bool                  `json:"run_as_approver,omitempty"`
	// AuthenticationPolicy specifies authentication requirements for running this flow.
	// Added in Python SDK v4.1.0; service support may be pending.
	AuthenticationPolicy *FlowAuthenticationPolicy `json:"authentication_policy,omitempty"`
}

// ListFlowsOptions represents options for listing Flows
type ListFlowsOptions struct {
	Limit        int    `url:"limit,omitempty"`
	Offset       int    `url:"offset,omitempty"`
	Marker       string `url:"marker,omitempty"`
	PerPage      int    `url:"per_page,omitempty"` // Alias for Limit
	OrderBy      string `url:"orderby,omitempty"`
	Q            string `url:"q,omitempty"`
	FilterRoles  string `url:"filter_roles,omitempty"`
	FilterOwner  string `url:"filter_owner,omitempty"`
	FilterPublic bool   `url:"filter_public,omitempty"`
	RolesOnly    bool   `url:"roles_only,omitempty"`
}

// ListRunsOptions represents options for listing Flow runs
type ListRunsOptions struct {
	Limit    int    `url:"limit,omitempty"`
	Offset   int    `url:"offset,omitempty"`
	Marker   string `url:"marker,omitempty"`
	PerPage  int    `url:"per_page,omitempty"` // Alias for Limit
	OrderBy  string `url:"orderby,omitempty"`
	Q        string `url:"q,omitempty"`
	FlowID   string `url:"flow_id,omitempty"`
	Status   string `url:"status,omitempty"`
	RoleType string `url:"role_type,omitempty"`
	Label    string `url:"label,omitempty"`
}

// RunRequest represents a request to run a Flow. FlowID selects the target flow
// (used in the URL, not sent in the body). The flow's first-state input goes
// under Body per the upstream run_flow contract.
type RunRequest struct {
	FlowID      string                 `json:"-"`
	Body        map[string]interface{} `json:"body"`
	Label       string                 `json:"label,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	RunManagers []string               `json:"run_managers,omitempty"`
	RunMonitors []string               `json:"run_monitors,omitempty"`
}

// RunResponse represents a Flow run
type RunResponse struct {
	RunID       string                 `json:"run_id"`
	FlowID      string                 `json:"flow_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"start_time,omitempty"`
	CompletedAt time.Time              `json:"completion_time,omitempty"`
	Label       string                 `json:"label,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	UserID      string                 `json:"user_id"`
	RunOwner    string                 `json:"run_owner"`
	RunManagers []string               `json:"run_managers,omitempty"`
	RunMonitors []string               `json:"run_monitors,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	FlowTitle   string                 `json:"flow_title,omitempty"`
	FlowScope   string                 `json:"flow_scope,omitempty"`
}

// RunList represents a list of Flow runs
type RunList struct {
	Runs    []RunResponse `json:"runs"`
	Total   int           `json:"total"`
	HadMore bool          `json:"had_more"`
	Offset  int           `json:"offset"`
	Limit   int           `json:"limit"`
}

// RunUpdateRequest represents a request to update a Flow run
type RunUpdateRequest struct {
	Label       string   `json:"label,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RunManagers []string `json:"run_managers,omitempty"`
	RunMonitors []string `json:"run_monitors,omitempty"`
}

// RunLogEntry represents an entry in a Flow run log. The timestamp key is "time".
type RunLogEntry struct {
	Code        string                 `json:"code"`
	Time        time.Time              `json:"time"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Description string                 `json:"description"`
}

// RunLogList represents a list of Flow run logs
type RunLogList struct {
	Entries []RunLogEntry `json:"entries"`
	Total   int           `json:"total"`
	HadMore bool          `json:"had_more"`
	Offset  int           `json:"offset"`
	Limit   int           `json:"limit"`
}

// RunMutableFields contains the fields that can be modified on a Flow run
type RunMutableFields struct {
	Label       string   `json:"label,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RunManagers []string `json:"run_managers,omitempty"`
	RunMonitors []string `json:"run_monitors,omitempty"`
}

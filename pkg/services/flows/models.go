// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
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

// RunRequest represents a request to run a Flow
type RunRequest struct {
	FlowID        string                 `json:"flow_id"`
	FlowTitle     string                 `json:"flow_title,omitempty"`
	FlowScope     string                 `json:"flow_scope,omitempty"`
	Label         string                 `json:"label,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	RunManagers   []string               `json:"run_managers,omitempty"`
	RunMonitors   []string               `json:"run_monitors,omitempty"`
	RunManagersRW bool                   `json:"run_managers_rw,omitempty"`
	Input         map[string]interface{} `json:"input"`
	ManageBy      string                 `json:"manage_by,omitempty"`
	MonitorBy     string                 `json:"monitor_by,omitempty"`
}

// RunResponse represents a Flow run
type RunResponse struct {
	RunID       string                 `json:"run_id"`
	FlowID      string                 `json:"flow_id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
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

// RunLogEntry represents an entry in a Flow run log
type RunLogEntry struct {
	Code        string                 `json:"code"`
	RunID       string                 `json:"run_id"`
	CreatedAt   time.Time              `json:"created_at"`
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

// ActionProvider represents a Flow action provider
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

// ActionProviderList represents a list of Flow action providers
type ActionProviderList struct {
	ActionProviders []ActionProvider `json:"action_providers"`
	Total           int              `json:"total"`
	HadMore         bool             `json:"had_more"`
	Offset          int              `json:"offset"`
	Limit           int              `json:"limit"`
}

// ActionRole represents a role in a Flow action
type ActionRole struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ActionFields map[string]interface{} `json:"action_fields,omitempty"`
	InputSchema  map[string]interface{} `json:"input_schema,omitempty"`
	Visible      bool                   `json:"visible"`
}

// ActionRoleList represents a list of Flow action roles
type ActionRoleList struct {
	ActionRoles []ActionRole `json:"action_roles"`
	Total       int          `json:"total"`
	HadMore     bool         `json:"had_more"`
	Offset      int          `json:"offset"`
	Limit       int          `json:"limit"`
}

// ListActionProvidersOptions represents options for listing action providers
type ListActionProvidersOptions struct {
	Limit        int    `url:"limit,omitempty"`
	Offset       int    `url:"offset,omitempty"`
	Marker       string `url:"marker,omitempty"`
	PerPage      int    `url:"per_page,omitempty"` // Alias for Limit
	OrderBy      string `url:"orderby,omitempty"`
	Q            string `url:"q,omitempty"`
	FilterOwner  string `url:"filter_owner,omitempty"`
	FilterType   string `url:"filter_type,omitempty"`
	FilterGlobus bool   `url:"filter_globus,omitempty"`
}

// RunMutableFields contains the fields that can be modified on a Flow run
type RunMutableFields struct {
	Label       string   `json:"label,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RunManagers []string `json:"run_managers,omitempty"`
	RunMonitors []string `json:"run_monitors,omitempty"`
}

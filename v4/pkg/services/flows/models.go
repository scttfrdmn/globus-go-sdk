// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
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

// FlowList is a marker-paginated list of flows.
type FlowList struct {
	Flows       []Flow `json:"flows"`
	Marker      string `json:"marker,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
}

// ListFlowsOptions contains options for listing flows. FilterRoles and
// FilterFulltext filter results; OrderBy is sent as repeated params.
type ListFlowsOptions struct {
	FilterRoles    []string
	FilterFulltext string
	OrderBy        []string
	Marker         string
}

// FlowCreate represents the payload for creating a flow. Title, Definition, and
// InputSchema are required upstream.
type FlowCreate struct {
	Title                  string                 `json:"title"`
	Definition             map[string]interface{} `json:"definition"`
	InputSchema            map[string]interface{} `json:"input_schema"`
	Subtitle               string                 `json:"subtitle,omitempty"`
	Description            string                 `json:"description,omitempty"`
	FlowViewers            []string               `json:"flow_viewers,omitempty"`
	FlowStarters           []string               `json:"flow_starters,omitempty"`
	FlowAdministrators     []string               `json:"flow_administrators,omitempty"`
	RunManagers            []string               `json:"run_managers,omitempty"`
	RunMonitors            []string               `json:"run_monitors,omitempty"`
	Keywords               []string               `json:"keywords,omitempty"`
	SubscriptionID         string                 `json:"subscription_id,omitempty"`
	AuthenticationPolicyID string                 `json:"authentication_policy_id,omitempty"`
}

// FlowUpdate represents the payload for updating a flow. Only set fields are sent.
type FlowUpdate struct {
	Title                  string                 `json:"title,omitempty"`
	Definition             map[string]interface{} `json:"definition,omitempty"`
	InputSchema            map[string]interface{} `json:"input_schema,omitempty"`
	Subtitle               string                 `json:"subtitle,omitempty"`
	Description            string                 `json:"description,omitempty"`
	FlowOwner              string                 `json:"flow_owner,omitempty"`
	FlowViewers            []string               `json:"flow_viewers,omitempty"`
	FlowStarters           []string               `json:"flow_starters,omitempty"`
	FlowAdministrators     []string               `json:"flow_administrators,omitempty"`
	RunManagers            []string               `json:"run_managers,omitempty"`
	RunMonitors            []string               `json:"run_monitors,omitempty"`
	Keywords               []string               `json:"keywords,omitempty"`
	SubscriptionID         string                 `json:"subscription_id,omitempty"`
	AuthenticationPolicyID string                 `json:"authentication_policy_id,omitempty"`
}

// RunActivityNotificationPolicy configures which run states trigger a
// notification. Status values: INACTIVE, SUCCEEDED, FAILED.
type RunActivityNotificationPolicy struct {
	Status []string `json:"status,omitempty"`
}

// FlowInput represents the body for running (or validating a run of) a flow. The
// flow's first-state input goes under Body.
type FlowInput struct {
	Body                       map[string]interface{}         `json:"body"`
	Label                      string                         `json:"label,omitempty"`
	Tags                       []string                       `json:"tags,omitempty"`
	RunMonitors                []string                       `json:"run_monitors,omitempty"`
	RunManagers                []string                       `json:"run_managers,omitempty"`
	ActivityNotificationPolicy *RunActivityNotificationPolicy `json:"activity_notification_policy,omitempty"`
}

// FlowRun represents a flow execution
type FlowRun struct {
	RunID     string                 `json:"run_id"`
	FlowID    string                 `json:"flow_id"`
	FlowTitle string                 `json:"flow_title,omitempty"`
	RunOwner  string                 `json:"run_owner,omitempty"`
	Status    string                 `json:"status"`
	Label     string                 `json:"label,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"completion_time,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// RunList is a marker-paginated list of flow runs.
type RunList struct {
	Runs        []FlowRun `json:"runs"`
	Marker      string    `json:"marker,omitempty"`
	HasNextPage bool      `json:"has_next_page"`
}

// ListRunsOptions contains options for listing runs.
type ListRunsOptions struct {
	FilterFlowID []string
	FilterRoles  []string
	Marker       string
}

// GetRunOptions carries optional parameters for GetRun.
type GetRunOptions struct {
	IncludeFlowDescription bool
}

// RunUpdate contains fields that can be changed on a running flow execution.
type RunUpdate struct {
	Label       string   `json:"label,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	RunMonitors []string `json:"run_monitors,omitempty"`
	RunManagers []string `json:"run_managers,omitempty"`
}

// RunDefinition is the response of GetRunDefinition.
type RunDefinition struct {
	FlowID      string                 `json:"flow_id"`
	Definition  map[string]interface{} `json:"definition"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// RunLog is a single log entry for a flow run.
type RunLog struct {
	Code        string                 `json:"code"`
	Description string                 `json:"description,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Time        time.Time              `json:"time"`
}

// RunLogList is a marker-paginated list of log entries for a flow run.
type RunLogList struct {
	Entries     []RunLog `json:"entries"`
	Marker      string   `json:"marker,omitempty"`
	HasNextPage bool     `json:"has_next_page"`
}

// ListRunLogsOptions controls which run log entries are returned. ReverseOrder is
// a pointer so an unset value omits the param.
type ListRunLogsOptions struct {
	Limit        int
	ReverseOrder *bool
	Marker       string
}

// RegisteredAPIRoles describes the principals holding each role on a
// registered API. Added in Python SDK v4.6.0.
type RegisteredAPIRoles struct {
	Owners         []string `json:"owners,omitempty"`
	Administrators []string `json:"administrators,omitempty"`
	Viewers        []string `json:"viewers,omitempty"`
}

// RegisteredAPI represents a Flows registered API.
// Added in Python SDK v4.6.0 (GET /registered_apis/{id}).
//
// The target and data_templates payloads are open-ended service documents, so
// they are modeled as map[string]interface{} rather than fixed structs.
type RegisteredAPI struct {
	ID                         string                 `json:"id"`
	Name                       string                 `json:"name"`
	Description                string                 `json:"description,omitempty"`
	Roles                      *RegisteredAPIRoles    `json:"roles,omitempty"`
	Target                     map[string]interface{} `json:"target,omitempty"`
	DataTemplates              map[string]interface{} `json:"data_templates,omitempty"`
	StateInputSchema           map[string]interface{} `json:"state_input_schema,omitempty"`
	Status                     string                 `json:"status,omitempty"`
	SubscriptionID             string                 `json:"subscription_id,omitempty"`
	CreatedTimestamp           time.Time              `json:"created_timestamp,omitempty"`
	EditedTimestamp            time.Time              `json:"edited_timestamp,omitempty"`
	UpdatedTimestamp           time.Time              `json:"updated_timestamp,omitempty"`
	ScheduledDeletionTimestamp *time.Time             `json:"scheduled_deletion_timestamp,omitempty"`
}

// RegisteredAPIList is a page of registered APIs.
// Uses marker pagination (keys registered_apis, marker, has_next_page).
type RegisteredAPIList struct {
	RegisteredAPIs []RegisteredAPI `json:"registered_apis"`
	Limit          int             `json:"limit,omitempty"`
	HasNextPage    bool            `json:"has_next_page"`
	Marker         string          `json:"marker,omitempty"`
}

// ListRegisteredAPIsOptions controls which registered APIs are returned.
// FilterRoles is comma-joined into the filter_roles query parameter; OrderBy is
// sent as repeated orderby params.
type ListRegisteredAPIsOptions struct {
	FilterRoles []string
	OrderBy     []string
	PerPage     int
	Marker      string
}

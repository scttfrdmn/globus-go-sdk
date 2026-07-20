// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package transfer

// This file adds the response/request model types required for 3.65.0 wire
// parity that were missing from the original transfer surface. All list
// envelopes decode items under the uppercase "DATA" key (per upstream
// iterable.py default_iter_key=DATA), except get_shared_endpoint_list which
// uses "shared_endpoints".

// Bookmark represents a Globus Transfer bookmark.
type Bookmark struct {
	DataType   string `json:"DATA_TYPE,omitempty"` // "bookmark"
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	EndpointID string `json:"endpoint_id,omitempty"`
	Path       string `json:"path,omitempty"`
}

// BookmarkList is a list of bookmarks (bookmark_list).
type BookmarkList struct {
	Data []Bookmark `json:"DATA"`
}

// ACLRule represents an endpoint access-control rule.
type ACLRule struct {
	DataType      string `json:"DATA_TYPE,omitempty"` // "access"
	ID            string `json:"id,omitempty"`
	AccessID      string `json:"access_id,omitempty"`
	PrincipalType string `json:"principal_type,omitempty"`
	Principal     string `json:"principal,omitempty"`
	Path          string `json:"path,omitempty"`
	Permissions   string `json:"permissions,omitempty"`
	RoleID        string `json:"role_id,omitempty"`
}

// ACLList is a list of ACL rules (access_list).
type ACLList struct {
	Data []ACLRule `json:"DATA"`
}

// Role represents an endpoint role assignment.
type Role struct {
	DataType      string `json:"DATA_TYPE,omitempty"` // "role"
	ID            string `json:"id,omitempty"`
	PrincipalType string `json:"principal_type,omitempty"`
	Principal     string `json:"principal,omitempty"`
	Role          string `json:"role,omitempty"`
}

// RoleList is a list of endpoint roles (role_list).
type RoleList struct {
	Data []Role `json:"DATA"`
}

// Server represents a GridFTP server backing an endpoint.
type Server struct {
	DataType string `json:"DATA_TYPE,omitempty"` // "server"
	ID       int    `json:"id,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Port     int    `json:"port,omitempty"`
	Scheme   string `json:"scheme,omitempty"`
	URI      string `json:"uri,omitempty"`
	Subject  string `json:"subject,omitempty"`
}

// ServerList is a list of servers (server_list).
type ServerList struct {
	Data []Server `json:"DATA"`
}

// PauseRule represents an endpoint-manager pause rule.
type PauseRule struct {
	DataType               string `json:"DATA_TYPE,omitempty"` // "pause_rule"
	ID                     string `json:"id,omitempty"`
	Message                string `json:"message,omitempty"`
	EndpointID             string `json:"endpoint_id,omitempty"`
	IdentityID             string `json:"identity_id,omitempty"`
	StartTime              string `json:"start_time,omitempty"`
	PauseLs                bool   `json:"pause_ls,omitempty"`
	PauseTaskTransferRead  bool   `json:"pause_task_transfer_read,omitempty"`
	PauseTaskTransferWrite bool   `json:"pause_task_transfer_write,omitempty"`
	Editable               bool   `json:"editable,omitempty"`
}

// PauseRuleList is a list of pause rules (pause_rule_list).
type PauseRuleList struct {
	Data []PauseRule `json:"DATA"`
}

// PauseInfo is the pause_info response for a task.
type PauseInfo struct {
	PauseRules              []PauseRule `json:"pause_rules"`
	SourcePauseMessage      string      `json:"source_pause_message,omitempty"`
	DestinationPauseMessage string      `json:"destination_pause_message,omitempty"`
}

// TaskEventListItem represents a single task event.
type TaskEventListItem struct {
	DataType    string `json:"DATA_TYPE,omitempty"` // "event"
	Time        string `json:"time,omitempty"`
	Description string `json:"description,omitempty"`
	Code        string `json:"code,omitempty"`
	IsError     bool   `json:"is_error,omitempty"`
	Details     string `json:"details,omitempty"`
}

// TaskEventListResponse is the event_list response (limit/offset/total paged).
type TaskEventListResponse struct {
	Data   []TaskEventListItem `json:"DATA"`
	Total  int                 `json:"total"`
	Offset int                 `json:"offset"`
	Limit  int                 `json:"limit"`
}

// SuccessfulTransfer represents one entry in a task's successful_transfers list.
type SuccessfulTransfer struct {
	DataType        string `json:"DATA_TYPE,omitempty"` // "successful_transfer"
	SourcePath      string `json:"source_path,omitempty"`
	DestinationPath string `json:"destination_path,omitempty"`
}

// SuccessfulTransfersList is the successful_transfers response (marker paged).
type SuccessfulTransfersList struct {
	Data       []SuccessfulTransfer `json:"DATA"`
	NextMarker string               `json:"next_marker"`
}

// SkippedError represents one entry in a task's skipped_errors list.
type SkippedError struct {
	DataType          string `json:"DATA_TYPE,omitempty"` // "skipped_error"
	ErrorCode         string `json:"error_code,omitempty"`
	SourcePath        string `json:"source_path,omitempty"`
	DestinationPath   string `json:"destination_path,omitempty"`
	ExternalChecksum  string `json:"external_checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`
}

// SkippedErrorsList is the skipped_errors response (marker paged).
type SkippedErrorsList struct {
	Data       []SkippedError `json:"DATA"`
	NextMarker string         `json:"next_marker"`
}

// SharedEndpointList is the get_shared_endpoint_list response. Items are keyed
// under "shared_endpoints" (NOT "DATA") and paginated by next_token.
type SharedEndpointList struct {
	SharedEndpoints []Endpoint `json:"shared_endpoints"`
	NextToken       string     `json:"next_token,omitempty"`
}

// AdminCancelStatus is the response for admin_cancel and admin_cancel status.
type AdminCancelStatus struct {
	ID     interface{} `json:"id,omitempty"`
	Status string      `json:"status,omitempty"`
	Done   bool        `json:"done,omitempty"`
	Code   string      `json:"code,omitempty"`
}

// SharedEndpointListOptions carries pagination for get_shared_endpoint_list.
type SharedEndpointListOptions struct {
	MaxResults int
	NextToken  string
}

// StatOptions carries options for operation_stat.
type StatOptions struct {
	LocalUser string
}

// MarkerPageOptions carries the marker cursor for marker-paginated lists.
type MarkerPageOptions struct {
	Marker string
}

// TaskEventListOptions carries limit/offset for task event_list.
type TaskEventListOptions struct {
	Limit  int
	Offset int
}

// EMTaskEventListOptions carries options for endpoint_manager task event_list.
type EMTaskEventListOptions struct {
	Limit         int
	Offset        int
	FilterIsError *bool
}

// EndpointManagerTaskListOptions carries options for endpoint_manager task_list.
// Slice filters are comma-joined; filter_completion_time is a preformatted
// string or "start,end"; last_key drives LastKeyPaginator.
type EndpointManagerTaskListOptions struct {
	FilterStatus         []string
	FilterTaskID         []string
	FilterOwnerID        string
	FilterEndpoint       string
	FilterEndpointUse    string // source | destination
	FilterIsPaused       *bool
	FilterCompletionTime string
	FilterMinFaults      int
	FilterLocalUser      string
	LastKey              string
}

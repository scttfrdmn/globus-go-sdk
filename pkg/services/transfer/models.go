// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"encoding/json"
	"time"
)

const (
	// SyncLevel constants define how to synchronize files
	SyncLevelExists    = 0 // Only transfer if destination doesn't exist
	SyncLevelSize      = 1 // Transfer if size differs
	SyncLevelModified  = 2 // Transfer if size or modification time differs
	SyncLevelChecksum  = 3 // Transfer if size, modification time, or checksum differs
)

// Endpoint represents a Globus endpoint (source or destination for transfers)
type Endpoint struct {
	ID                     string                 `json:"id"`
	DisplayName            string                 `json:"display_name"`
	CanonicalName          string                 `json:"canonical_name,omitempty"`
	Description            string                 `json:"description,omitempty"`
	ActivationRequirements map[string]interface{} `json:"activation_requirements,omitempty"`
	ActivationProfileID    string                 `json:"activation_profile_id,omitempty"`
	OwnerString            string                 `json:"owner_string"`
	OwnerID                string                 `json:"owner_id"`
	Organization           string                 `json:"organization,omitempty"`
	Department             string                 `json:"department,omitempty"`
	Keywords               []string               `json:"keywords,omitempty"`
	ContactEmail           string                 `json:"contact_email,omitempty"`
	ContactInfo            string                 `json:"contact_info,omitempty"`
	Public                 bool                   `json:"public"`
	Subscription           interface{}            `json:"subscription_id,omitempty"`
	NetworkUse             string                 `json:"network_use,omitempty"`
	DefaultDirectory       string                 `json:"default_directory,omitempty"`
	Force                  bool                   `json:"force_encryption,omitempty"`
	OAuth                  bool                   `json:"oauth_server,omitempty"`
	DeactivationTime       *time.Time             `json:"deactivation_time,omitempty"`
	Activated              bool                   `json:"activated"`
	GCPPaused              bool                   `json:"gcp_paused,omitempty"`
	GCPConnected           bool                   `json:"gcp_connected,omitempty"`
	HibernationState       string                 `json:"hibernation_state,omitempty"`
	Collections            []Collection           `json:"collections,omitempty"`
	HostEndpointID         string                 `json:"host_endpoint_id,omitempty"`
	LocalUserInfo          *LocalUserInfo         `json:"local_user_info,omitempty"`
	LocalUserInfoAvailable bool                   `json:"local_user_info_available,omitempty"`
}

// LocalUserInfo represents local user information for an endpoint
type LocalUserInfo struct {
	Username string `json:"username"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	HomeDir  string `json:"home_dir,omitempty"`
}

// Collection represents a Globus collection (subset of an endpoint)
type Collection struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	BasePath         string `json:"base_path"`
	AdminReadACL     bool   `json:"admin_read_acl,omitempty"`
	AdminWriteACL    bool   `json:"admin_write_acl,omitempty"`
	IdentitiesRead   bool   `json:"identities_read,omitempty"`
	IdentitiesWrite  bool   `json:"identities_write,omitempty"`
	UserManageShares bool   `json:"user_manage_shares,omitempty"`
	UserMessageOnly  bool   `json:"user_message_only,omitempty"`
}

// EndpointList represents a paginated list of endpoints
type EndpointList struct {
	Data          []Endpoint `json:"data"`
	NextPageToken string     `json:"next_page_token,omitempty"`
	HasNextPage   bool       `json:"has_next_page"`
}

// ListEndpointsOptions contains options for filtering endpoint listings
type ListEndpointsOptions struct {
	FilterFullText     string `url:"filter_fulltext,omitempty"`
	FilterOwnerID      string `url:"filter_owner_id,omitempty"`
	FilterHostEndpoint string `url:"filter_host_endpoint,omitempty"`
	FilterScope        string `url:"filter_scope,omitempty"` // all, recently-used, in-use, my-endpoints, shared-with-me
	Limit              int    `url:"limit,omitempty"`
	Offset             int    `url:"offset,omitempty"`
	PageSize           int    `url:"page_size,omitempty"`
	PageToken          string `url:"page_token,omitempty"`
}

// Task represents a transfer or delete task
type Task struct {
	DataType               string                 `json:"data_type"`
	TaskID                 string                 `json:"task_id"`
	Type                   string                 `json:"type"`   // TRANSFER or DELETE
	Status                 string                 `json:"status"` // ACTIVE, INACTIVE, FAILED, SUCCEEDED, CANCELLED
	Label                  string                 `json:"label,omitempty"`
	SourceEndpointID       string                 `json:"source_endpoint_id,omitempty"`
	SourceEndpointDisplay  string                 `json:"source_endpoint_display_name,omitempty"`
	DestinationEndpointID  string                 `json:"destination_endpoint_id,omitempty"`
	DestEndpointDisplay    string                 `json:"destination_endpoint_display_name,omitempty"`
	RequestTime            time.Time              `json:"request_time"`
	CompletionTime         *time.Time             `json:"completion_time,omitempty"`
	Deadline               *time.Time             `json:"deadline,omitempty"`
	CancelTime             *time.Time             `json:"cancel_time,omitempty"`
	CreatorID              string                 `json:"creator_id"`
	OwnerID                string                 `json:"owner_id"`
	FilesTransferred       int                    `json:"files_transferred"`
	FilesSkipped           int                    `json:"files_skipped"`
	BytesTransferred       int64                  `json:"bytes_transferred"`
	BytesSkipped           int64                  `json:"bytes_skipped"`
	Subtasks               int                    `json:"subtasks_total"`
	SubtasksSucceeded      int                    `json:"subtasks_succeeded"`
	SubtasksFailed         int                    `json:"subtasks_failed"`
	SubtasksPending        int                    `json:"subtasks_pending"`
	SubtasksCanceled       int                    `json:"subtasks_canceled"`
	SubtasksExpired        int                    `json:"subtasks_expired"`
	SyncLevel              int                    `json:"sync_level"`
	EncryptData            bool                   `json:"encrypt_data"`
	VerifyChecksum         bool                   `json:"verify_checksum"`
	DeleteDestinationExtra bool                   `json:"delete_destination_extra"`
	Recursive              bool                   `json:"recursive"`
	PerfOerationID         string                 `json:"perf_operation_id,omitempty"`
	UseSharing             bool                   `json:"use_sharing"`
	HistoryDeleted         bool                   `json:"history_deleted"`
	PublicationID          string                 `json:"publication_id,omitempty"`
	SkipSourceErrors       bool                   `json:"skip_source_errors"`
	FailOnQuotaErrors      bool                   `json:"fail_on_quota_errors"`
	FilesFixedPerms        int                    `json:"files_fixed_perms,omitempty"`
	FilesWithErrorPerms    int                    `json:"files_with_error_perms,omitempty"`
	UserMessage            string                 `json:"user_message,omitempty"`
	SymlinkDepth           int                    `json:"symlink_depth,omitempty"`
	PreserveMtime          bool                   `json:"preserve_mtime"`
	FatalErrorDetails      map[string]interface{} `json:"fatal_error,omitempty"`
}

// TaskList represents a paginated list of tasks
type TaskList struct {
	Data          []Task `json:"data"`
	NextPageToken string `json:"next_page_token,omitempty"`
	NextMarker    string `json:"next_marker,omitempty"` // Alternative name for NextPageToken
	HasNextPage   bool   `json:"has_next_page"`
}

// ListTasksOptions contains options for filtering task listings
type ListTasksOptions struct {
	FilterTaskID         string    `url:"filter_task_id,omitempty"`
	FilterType           string    `url:"filter_type,omitempty"`   // TRANSFER or DELETE
	FilterStatus         string    `url:"filter_status,omitempty"` // ACTIVE, INACTIVE, FAILED, SUCCEEDED, CANCELLED
	TaskType             string    `url:"task_type,omitempty"`     // Alias for FilterType
	Status               string    `url:"status,omitempty"`        // Alias for FilterStatus
	FilterCompletedSince time.Time `url:"filter_completion_time.min,omitempty"`
	FilterCompletedUntil time.Time `url:"filter_completion_time.max,omitempty"`
	FilterRequestedSince time.Time `url:"filter_request_time.min,omitempty"`
	FilterRequestedUntil time.Time `url:"filter_request_time.max,omitempty"`
	Limit                int       `url:"limit,omitempty"`
	Offset               int       `url:"offset,omitempty"`
	PageSize             int       `url:"page_size,omitempty"`
	PageToken            string    `url:"page_token,omitempty"`
}

// TransferItem represents a single file or directory to transfer
type TransferItem struct {
	DataType        string `json:"DATA_TYPE,omitempty"` // Should be "transfer_item"
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
	Recursive       bool   `json:"recursive,omitempty"`
	Checksum        string `json:"checksum,omitempty"`
}

// DeleteItem represents a single file or directory to delete
type DeleteItem struct {
	DataType  string `json:"DATA_TYPE"` // Must be "delete_item" - Important: This field is required for the API
	Path      string `json:"path"`
	// Note: The API does not support a "recursive" field for delete_item as of API v0.10
}

// TransferTaskRequest represents a request to create a transfer task
type TransferTaskRequest struct {
	DataType               string         `json:"DATA_TYPE,omitempty"`
	Label                  string         `json:"label,omitempty"`
	SourceEndpointID       string         `json:"source_endpoint"`
	DestinationEndpointID  string         `json:"destination_endpoint"`
	VerifyChecksum         bool           `json:"verify_checksum,omitempty"`
	Encrypt                bool           `json:"encrypt_data,omitempty"`
	SyncLevel              int            `json:"sync_level,omitempty"`
	DeleteDestinationExtra bool           `json:"delete_destination_extra,omitempty"`
	Deadline               *time.Time     `json:"deadline,omitempty"`
	NotifyOnSucceeded      bool           `json:"notify_on_succeeded,omitempty"`
	NotifyOnFailed         bool           `json:"notify_on_failed,omitempty"`
	NotifyOnInactive       bool           `json:"notify_on_inactive,omitempty"`
	SkipSourceErrors       bool           `json:"skip_source_errors,omitempty"`
	FailOnQuotaErrors      bool           `json:"fail_on_quota_errors,omitempty"`
	SubmissionID           string         `json:"submission_id,omitempty"`
	UseSharing             bool           `json:"use_sharing,omitempty"`
	SymlinkDepth           int            `json:"symlink_depth,omitempty"`
	PreserveMtime          bool           `json:"preserve_mtime,omitempty"`
	Items                  []TransferItem `json:"DATA"`
}

// DeleteTaskRequest represents a request to create a delete task
type DeleteTaskRequest struct {
	DataType          string       `json:"DATA_TYPE,omitempty"`
	Label             string       `json:"label,omitempty"`
	EndpointID        string       `json:"endpoint"`
	Deadline          *time.Time   `json:"deadline,omitempty"`
	NotifyOnSucceeded bool         `json:"notify_on_succeeded,omitempty"`
	NotifyOnFailed    bool         `json:"notify_on_failed,omitempty"`
	NotifyOnInactive  bool         `json:"notify_on_inactive,omitempty"`
	SubmissionID      string       `json:"submission_id,omitempty"`
	Items             []DeleteItem `json:"DATA"` // Important: "DATA" field must be uppercase
}

// TaskResponse represents the response from creating a task
type TaskResponse struct {
	TaskID          string      `json:"task_id"`
	SubmissionID    string      `json:"submission_id,omitempty"`
	Code            string      `json:"code,omitempty"`
	Message         string      `json:"message,omitempty"`
	OAuthURL        string      `json:"oauth_url,omitempty"`
	TransferDetails interface{} `json:"transfer_details,omitempty"`
}

// OperationResult represents the result of an operation
type OperationResult struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Resource  string      `json:"resource,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	TaskID    string      `json:"task_id,omitempty"` // Added for convenience in some operations
}

// NOTE: ActivationRequirements struct has been removed as activation is now handled
// automatically with properly scoped tokens in modern Globus endpoints (v0.10+).

// FileListItem represents an item in a file listing
type FileListItem struct {
	DataType     string `json:"data_type"`
	Name         string `json:"name"`
	Type         string `json:"type"` // file or dir
	Size         int64  `json:"size,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	Permissions  string `json:"permissions,omitempty"`
	User         string `json:"user,omitempty"`
	Group        string `json:"group,omitempty"`
	Link         string `json:"link_target,omitempty"`
}

// FileList represents a paginated list of files and directories
type FileList struct {
	Data          []FileListItem `json:"data"`
	EndpointID    string         `json:"endpoint_id"`
	Path          string         `json:"path"`
	MaybeSharing  bool           `json:"maybe_sharing,omitempty"`
	ConstantTasks []string       `json:"constant_tasks,omitempty"`
	AbsolutePath  string         `json:"absolute_path,omitempty"`
	Marker        string         `json:"marker,omitempty"`
	ContinueFrom  string         `json:"continue_from,omitempty"`
	HasNextPage   bool           `json:"has_next_page"`
}

// ListFileOptions contains options for listing files
type ListFileOptions struct {
	OrderBy       string `url:"orderby,omitempty"` // name, type, date, size
	Filter        string `url:"filter,omitempty"`
	ShowHidden    bool   `url:"show_hidden,omitempty"`
	ContinueFrom  string `url:"continue_from,omitempty"`
	Marker        string `url:"marker,omitempty"`
	Limit         int    `url:"limit,omitempty"`
	ExcludedTypes string `url:"excluded_types,omitempty"` // Comma-separated list of types to exclude: file, dir, symlink
}

// MarshalJSON for time.Time types properly formats them for the API
func MarshalJSONTime(t time.Time) ([]byte, error) {
	return json.Marshal(t.Format(time.RFC3339))
}

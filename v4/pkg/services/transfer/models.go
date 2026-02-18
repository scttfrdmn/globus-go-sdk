// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transfer

import "time"

// Endpoint represents a Globus Transfer endpoint
type Endpoint struct {
	ID                   string    `json:"id"`
	DisplayName          string    `json:"display_name"`
	Owner                string    `json:"owner_string"`
	Activated            bool      `json:"activated"`
	Public               bool      `json:"public"`
	Description          string    `json:"description,omitempty"`
	Organization         string    `json:"organization,omitempty"`
	Department           string    `json:"department,omitempty"`
	Keywords             []string  `json:"keywords,omitempty"`
	MyEffectiveRoles     []string  `json:"my_effective_roles,omitempty"`
	SubscriptionID       string    `json:"subscription_id,omitempty"`
	NetworkUse           string    `json:"network_use,omitempty"`
	MaxConcurrency       int       `json:"max_concurrency,omitempty"`
	PreferredConcurrency int       `json:"preferred_concurrency,omitempty"`
	Created              time.Time `json:"creation_time"`
	LastModified         time.Time `json:"last_modified_time,omitempty"`
}

// EndpointList represents a paginated list of endpoints
type EndpointList struct {
	DATA_TYPE string     `json:"DATA_TYPE"`
	Data      []Endpoint `json:"DATA"`
	Offset    int        `json:"offset"`
	Limit     int        `json:"limit"`
	Total     int        `json:"total"`
}

// ListEndpointsOptions contains options for listing endpoints
type ListEndpointsOptions struct {
	Filter string
	Limit  int
	Offset int
}

// Transfer represents a transfer task submission
type Transfer struct {
	DATA_TYPE           string         `json:"DATA_TYPE"`
	SourceEndpoint      string         `json:"source_endpoint"`
	DestinationEndpoint string         `json:"destination_endpoint"`
	Label               string         `json:"label,omitempty"`
	SyncLevel           int            `json:"sync_level,omitempty"`
	VerifyChecksum      bool           `json:"verify_checksum,omitempty"`
	PreserveTimestamp   bool           `json:"preserve_timestamp,omitempty"`
	EncryptData         bool           `json:"encrypt_data,omitempty"`
	DeleteDestinationExtra bool        `json:"delete_destination_extra,omitempty"`
	SkipSourceErrors    bool           `json:"skip_source_errors,omitempty"`
	FailOnQuotaErrors   bool           `json:"fail_on_quota_errors,omitempty"`
	Items               []TransferItem `json:"DATA"`
	NotifyOnSucceeded   bool           `json:"notify_on_succeeded,omitempty"`
	NotifyOnFailed      bool           `json:"notify_on_failed,omitempty"`
	NotifyOnInactive    bool           `json:"notify_on_inactive,omitempty"`
	Deadline            string         `json:"deadline,omitempty"`
}

// TransferItem represents a single file/directory to transfer
type TransferItem struct {
	DATA_TYPE         string `json:"DATA_TYPE"`
	SourcePath        string `json:"source_path"`
	DestinationPath   string `json:"destination_path"`
	Recursive         bool   `json:"recursive,omitempty"`
	ExternalChecksum  string `json:"external_checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`
}

// Delete represents a delete task submission
type Delete struct {
	DATA_TYPE          string       `json:"DATA_TYPE"`
	Endpoint           string       `json:"endpoint"`
	Label              string       `json:"label,omitempty"`
	Recursive          bool         `json:"recursive,omitempty"`
	IgnoreMissing      bool         `json:"ignore_missing,omitempty"`
	InterpretGlob      bool         `json:"interpret_globs,omitempty"`
	Items              []DeleteItem `json:"DATA"`
	NotifyOnSucceeded  bool         `json:"notify_on_succeeded,omitempty"`
	NotifyOnFailed     bool         `json:"notify_on_failed,omitempty"`
	NotifyOnInactive   bool         `json:"notify_on_inactive,omitempty"`
	Deadline           string       `json:"deadline,omitempty"`
}

// DeleteItem represents a single file/directory to delete
type DeleteItem struct {
	DATA_TYPE string `json:"DATA_TYPE"`
	Path      string `json:"path"`
}

// Task represents a transfer or delete task
type Task struct {
	DATA_TYPE           string    `json:"DATA_TYPE"`
	TaskID              string    `json:"task_id"`
	Type                string    `json:"type"`
	Status              string    `json:"status"`
	Label               string    `json:"label,omitempty"`
	SourceEndpoint      string    `json:"source_endpoint,omitempty"`
	DestinationEndpoint string    `json:"destination_endpoint,omitempty"`
	Endpoint            string    `json:"endpoint,omitempty"`
	RequestTime         time.Time `json:"request_time"`
	CompletionTime      time.Time `json:"completion_time,omitempty"`
	DeadlineTime        time.Time `json:"deadline,omitempty"`
	BytesTransferred    int64     `json:"bytes_transferred"`
	BytesChecksummed    int64     `json:"bytes_checksummed"`
	FilesTransferred    int       `json:"files_transferred"`
	FilesSkipped        int       `json:"files_skipped"`
	Directories         int       `json:"directories"`
	FatalError          *TaskError `json:"fatal_error,omitempty"`
	IsOk                bool      `json:"is_ok"`
	IsPaused            bool      `json:"is_paused"`
	Owner               string    `json:"owner_string,omitempty"`
	NiceStatus          string    `json:"nice_status,omitempty"`
	NiceStatusShortDescription string `json:"nice_status_short_description,omitempty"`
	NiceStatusDetails   string    `json:"nice_status_details,omitempty"`
}

// TaskError represents an error that occurred during a task
type TaskError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
}

// TaskSubmitResponse represents the response from submitting a task
type TaskSubmitResponse struct {
	DATA_TYPE        string `json:"DATA_TYPE"`
	Code             string `json:"code"`
	Message          string `json:"message"`
	RequestID        string `json:"request_id"`
	Resource         string `json:"resource"`
	TaskID           string `json:"task_id"`
	SubmissionID     string `json:"submission_id"`
}

// TaskCancelResponse represents the response from canceling a task
type TaskCancelResponse struct {
	DATA_TYPE string `json:"DATA_TYPE"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Resource  string `json:"resource"`
}

// TaskList represents a paginated list of tasks
type TaskList struct {
	DATA_TYPE string `json:"DATA_TYPE"`
	Data      []Task `json:"DATA"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
	Total     int    `json:"total"`
}

// ListTasksOptions contains options for listing tasks
type ListTasksOptions struct {
	Filter       string
	FilterStatus []string
	Limit        int
	Offset       int
}

// DirectoryEntry represents a file or directory in a listing
type DirectoryEntry struct {
	DATA_TYPE    string    `json:"DATA_TYPE"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "file" or "dir"
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	Permissions  string    `json:"permissions,omitempty"`
	User         string    `json:"user,omitempty"`
	Group        string    `json:"group,omitempty"`
	LinkTarget   string    `json:"link_target,omitempty"`
}

// DirectoryListing represents the contents of a directory
type DirectoryListing struct {
	DATA_TYPE     string           `json:"DATA_TYPE"`
	Path          string           `json:"path"`
	Endpoint      string           `json:"endpoint"`
	Data          []DirectoryEntry `json:"DATA"`
	AbsolutePath  string           `json:"absolute_path,omitempty"`
	Offset        int              `json:"offset"`
	Limit         int              `json:"limit"`
	Total         int              `json:"total"`
}

// ListDirectoryOptions contains options for listing directory contents
type ListDirectoryOptions struct {
	ShowHidden bool
	Limit      int
	Offset     int
}

// OperationResponse represents the response from an operation (mkdir, rename, etc.)
type OperationResponse struct {
	DATA_TYPE string `json:"DATA_TYPE"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Resource  string `json:"resource"`
}

// Tunnel represents a Globus Streams tunnel.
// Added in Python SDK v4.3.0.
type Tunnel struct {
	ID               string                 `json:"id"`
	DisplayName      string                 `json:"display_name,omitempty"`
	Owner            string                 `json:"owner,omitempty"`
	SourceEndpointID string                 `json:"source_endpoint_id,omitempty"`
	SourcePath       string                 `json:"source_path,omitempty"`
	Status           string                 `json:"status,omitempty"`
	CreatedAt        *time.Time             `json:"created_at,omitempty"`
	UpdatedAt        *time.Time             `json:"updated_at,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelList represents a paginated list of tunnels
type TunnelList struct {
	Tunnels []Tunnel `json:"DATA"`
	Total   int      `json:"total,omitempty"`
	HasMore bool     `json:"has_next_page,omitempty"`
	Marker  string   `json:"next_marker,omitempty"`
}

// TunnelCreate is the payload for creating a new tunnel
type TunnelCreate struct {
	DisplayName      string                 `json:"display_name"`
	SourceEndpointID string                 `json:"source_endpoint_id"`
	SourcePath       string                 `json:"source_path"`
	ExpiresIn        *int                   `json:"expires_in,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelUpdate is the payload for updating an existing tunnel
type TunnelUpdate struct {
	DisplayName string                 `json:"display_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListTunnelsOptions contains options for listing tunnels
type ListTunnelsOptions struct {
	Limit  int
	Marker string
}

// StreamAccessPoint represents a Globus Stream Access Point.
// Added in Python SDK v4.3.0.
type StreamAccessPoint struct {
	ID         string                 `json:"id"`
	TunnelID   string                 `json:"tunnel_id,omitempty"`
	EndpointID string                 `json:"endpoint_id,omitempty"`
	Path       string                 `json:"path,omitempty"`
	AccessURL  string                 `json:"access_url,omitempty"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TunnelEvent represents an event associated with a tunnel.
// Added in Python SDK v4.4.0.
type TunnelEvent struct {
	ID          string                 `json:"id"`
	TunnelID    string                 `json:"tunnel_id"`
	Code        string                 `json:"code"`
	Description string                 `json:"description,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	OccurredAt  *time.Time             `json:"occurred_at,omitempty"`
}

// TunnelEventList represents a paginated list of tunnel events
type TunnelEventList struct {
	Events  []TunnelEvent `json:"DATA"`
	Total   int           `json:"total,omitempty"`
	HasMore bool          `json:"has_next_page,omitempty"`
	Marker  string        `json:"next_marker,omitempty"`
}

// ListTunnelEventsOptions contains options for listing tunnel events
type ListTunnelEventsOptions struct {
	Limit  int
	Marker string
}

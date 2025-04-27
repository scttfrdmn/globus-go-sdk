<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK Data Schemas

This document describes the key data structures and schemas used in the Globus Go SDK.

## Table of Contents

- [Authentication Schemas](#authentication-schemas)
- [Transfer Schemas](#transfer-schemas)
- [Groups Schemas](#groups-schemas)
- [Common Schemas](#common-schemas)

## Authentication Schemas

### TokenInfo

The `TokenInfo` struct represents an OAuth2 token and its associated metadata.

```go
type TokenInfo struct {
    // The OAuth2 access token.
    AccessToken string `json:"access_token"`

    // The OAuth2 refresh token, may be empty for some grant types.
    RefreshToken string `json:"refresh_token,omitempty"`

    // The token type, typically "Bearer".
    TokenType string `json:"token_type,omitempty"`

    // The time when this token expires.
    ExpiresAt time.Time `json:"expires_at,omitempty"`

    // The scope string associated with this token.
    Scope string `json:"scope,omitempty"`

    // The resource this token is valid for.
    Resource string `json:"resource,omitempty"`

    // The client ID associated with this token.
    ClientID string `json:"client_id,omitempty"`
}
```

### ClientConfig

The `ClientConfig` struct contains configuration for the OAuth client.

```go
type ClientConfig struct {
    // The registered client ID.
    ClientID string

    // The client secret. Only required for confidential clients.
    ClientSecret string

    // The redirect URL for the OAuth flow.
    RedirectURL string

    // The base URL for the Globus Auth service.
    BaseURL string

    // Optional HTTP client to use for requests.
    HTTPClient *http.Client
}
```

### AuthCodeOptions

The `AuthCodeOptions` struct contains options for the authorization code flow.

```go
type AuthCodeOptions struct {
    // Additional scopes to request.
    Scopes []string

    // State parameter for security.
    State string

    // Preferred authentication method.
    AuthMethod string

    // Session required parameter.
    SessionRequired bool

    // Additional URL parameters.
    AdditionalParams map[string]string
}
```

## Transfer Schemas

### TransferItem

The `TransferItem` struct represents a single file or directory to be transferred.

```go
type TransferItem struct {
    // The source path for the transfer.
    Source string `json:"source_path"`

    // The destination path for the transfer.
    Destination string `json:"destination_path"`

    // Whether this is a recursive directory transfer.
    Recursive bool `json:"recursive,omitempty"`

    // Additional options for this specific item.
    Options map[string]interface{} `json:"options,omitempty"`
}
```

### TransferData

The `TransferData` struct represents a transfer task submission.

```go
type TransferData struct {
    // A human-readable label for the transfer.
    Label string `json:"label"`

    // The list of items to transfer.
    Items []TransferItem `json:"DATA"`

    // The sync level for the transfer.
    SyncLevel string `json:"sync_level,omitempty"`

    // Whether to verify file integrity after transfer.
    Verify bool `json:"verify_checksums,omitempty"`

    // Whether to preserve source file timestamps.
    PreserveTimestamp bool `json:"preserve_timestamp,omitempty"`

    // Whether to encrypt data in transit.
    EncryptData bool `json:"encrypt_data,omitempty"`

    // When the transfer task should expire.
    Deadline *time.Time `json:"deadline,omitempty"`

    // Whether to notify on success.
    NotifyOnSucceeded bool `json:"notify_on_succeeded,omitempty"`

    // Whether to notify on failure.
    NotifyOnFailed bool `json:"notify_on_failed,omitempty"`

    // Whether to notify on inactivity.
    NotifyOnInactive bool `json:"notify_on_inactive,omitempty"`
}
```

### EndpointItem

The `EndpointItem` struct represents a file or directory on an endpoint.

```go
type EndpointItem struct {
    // The type of item (file or directory).
    Type string `json:"type"`

    // The name of the item.
    Name string `json:"name"`

    // The size of the item in bytes.
    Size int64 `json:"size"`

    // The last modified time.
    LastModified time.Time `json:"last_modified"`

    // The permissions string.
    Permissions string `json:"permissions,omitempty"`

    // The user ID of the owner.
    User string `json:"user,omitempty"`

    // The group ID of the owner.
    Group string `json:"group,omitempty"`
}
```

### TaskStatus

The `TaskStatus` struct represents the status of a transfer task.

```go
type TaskStatus struct {
    // The task ID.
    TaskID string `json:"task_id"`

    // The current status (ACTIVE, SUCCEEDED, FAILED, etc.).
    Status string `json:"status"`

    // The number of files transferred.
    FilesTransferred int `json:"files_transferred"`

    // The total number of files to transfer.
    FilesTotal int `json:"files_total"`

    // The number of bytes transferred.
    BytesTransferred int64 `json:"bytes_transferred"`

    // The total number of bytes to transfer.
    BytesTotal int64 `json:"bytes_total"`

    // The time the task was created.
    CreatedAt time.Time `json:"created_at"`

    // The time the task was completed (if applicable).
    CompletedAt *time.Time `json:"completed_at,omitempty"`
}
```

### RecursiveTransferOptions

The `RecursiveTransferOptions` struct contains options for recursive directory transfers.

```go
type RecursiveTransferOptions struct {
    // A human-readable label for the transfer.
    Label string

    // The sync level for the transfer (exists, size, mtime, or checksum).
    SyncLevel string

    // Whether to verify file integrity after transfer.
    Verify bool

    // Whether to preserve source file timestamps.
    PreserveTimestamp bool

    // Whether to encrypt data in transit.
    EncryptData bool

    // When the transfer task should expire.
    Deadline *time.Time

    // Whether to notify on success.
    NotifyOnSucceeded bool

    // Whether to notify on failure.
    NotifyOnFailed bool

    // Whether to notify on inactivity.
    NotifyOnInactive bool

    // The number of items per batch.
    BatchSize int

    // The number of concurrent batches to submit.
    ConcurrentBatches int

    // Whether to skip the size limit check.
    SkipSizeLimitCheck bool
}
```

### RecursiveTransferResult

The `RecursiveTransferResult` struct contains the results of a recursive directory transfer.

```go
type RecursiveTransferResult struct {
    // The list of task IDs created for this transfer.
    TaskIDs []string

    // The primary task ID (the first one).
    TaskID string

    // The total number of items transferred.
    ItemsTransferred int

    // The total number of bytes transferred.
    BytesTransferred int64

    // The time the transfer started.
    StartTime time.Time

    // The time the transfer completed.
    EndTime time.Time

    // Whether the transfer was successful.
    Success bool

    // Any errors encountered during the transfer.
    Errors []error
}
```

## Groups Schemas

### Group

The `Group` struct represents a Globus group.

```go
type Group struct {
    // The group ID.
    ID string `json:"id"`

    // The group name.
    Name string `json:"name"`

    // The group description.
    Description string `json:"description,omitempty"`

    // Whether the group is a high assurance group.
    HighAssurance bool `json:"high_assurance,omitempty"`

    // The group's parent ID, if any.
    ParentID string `json:"parent_id,omitempty"`

    // The group's policies.
    Policies GroupPolicies `json:"policies,omitempty"`

    // The time the group was created.
    CreatedAt time.Time `json:"created_at,omitempty"`

    // The time the group was last updated.
    LastUpdated time.Time `json:"last_updated,omitempty"`
}
```

### Member

The `Member` struct represents a member of a Globus group.

```go
type Member struct {
    // The member's ID.
    ID string `json:"id"`

    // The member's username.
    Username string `json:"username,omitempty"`

    // The member's email.
    Email string `json:"email,omitempty"`

    // The member's name.
    Name string `json:"name,omitempty"`

    // The member's role in the group.
    Role string `json:"role"`

    // The time the member was added to the group.
    JoinedAt time.Time `json:"joined_at,omitempty"`
}
```

### GroupPolicies

The `GroupPolicies` struct represents a group's policies.

```go
type GroupPolicies struct {
    // Whether membership is managed by a membership service provider.
    ManagedByMSP bool `json:"managed_by_msp,omitempty"`

    // Whether the group is visible to users outside the group.
    IsVisible bool `json:"is_visible,omitempty"`

    // Whether users can discover the group.
    IsDiscoverable bool `json:"is_discoverable,omitempty"`

    // Whether the admin can manage members.
    AdminsManageMembers bool `json:"admins_manage_members,omitempty"`

    // Whether the admin can manage admins.
    AdminsManageAdmins bool `json:"admins_manage_admins,omitempty"`
}
```

## Common Schemas

### Pagination

The `Pagination` struct is used for paginated requests and responses.

```go
type Pagination struct {
    // The maximum number of items per page.
    Limit int `json:"limit,omitempty"`

    // The offset for pagination.
    Offset int `json:"offset,omitempty"`

    // The total number of items available.
    Total int `json:"total,omitempty"`

    // Whether there are more items available.
    HasNext bool `json:"has_next,omitempty"`
}
```

### Options

The `Options` interface is a common interface for request options.

```go
type Options interface {
    // ToQueryParams converts the options to a URL query parameter map.
    ToQueryParams() map[string]string

    // Validate validates the options.
    Validate() error
}
```

### Error

The SDK uses structured errors to provide detailed information about failures.

```go
type Error struct {
    // The HTTP status code.
    StatusCode int `json:"status_code,omitempty"`

    // The error code from the API.
    Code string `json:"code,omitempty"`

    // The error message.
    Message string `json:"message,omitempty"`

    // The request ID for support.
    RequestID string `json:"request_id,omitempty"`

    // The original error, if any.
    Cause error `json:"-"`
}
```

## Usage Examples

### Creating and Using a TokenInfo

```go
// Create a new token
token := &auth.TokenInfo{
    AccessToken:  "eyJhbGciOiJSUzI1NiIsImtpZCI6...",
    RefreshToken: "eyJhbGciOiJSUzI1NiIsImtpZCI6...",
    TokenType:    "Bearer",
    ExpiresAt:    time.Now().Add(1 * time.Hour),
    Scope:        "openid profile email urn:globus:auth:scope:transfer.api.globus.org:all",
}

// Check if the token is expired
if token.IsExpired() {
    fmt.Println("Token is expired and needs refreshing")
}

// Get the token's remaining lifetime
lifetime := token.Lifetime()
fmt.Printf("Token expires in %v\n", lifetime)
```

### Creating a Transfer Request

```go
// Create a transfer request
transfer := &transfer.TransferData{
    Label: "Important Data Transfer",
    Items: []transfer.TransferItem{
        {
            Source:      "/source/path/file1.txt",
            Destination: "/destination/path/file1.txt",
        },
        {
            Source:      "/source/path/file2.txt",
            Destination: "/destination/path/file2.txt",
        },
    },
    SyncLevel:         transfer.SyncChecksum,
    Verify:            true,
    PreserveTimestamp: true,
}

// Submit the transfer
task, err := client.SubmitTransfer(ctx, "source-endpoint", "destination-endpoint", transfer)
```

### Working with Groups

```go
// Create a new group
newGroup := &groups.Group{
    Name:        "My Research Team",
    Description: "A group for my research collaborators",
    Policies: groups.GroupPolicies{
        IsVisible:      true,
        IsDiscoverable: true,
    },
}

// Create the group
createdGroup, err := client.CreateGroup(ctx, newGroup)

// Add a member to the group
member := &groups.Member{
    ID:   "user@example.com",
    Role: "member",
}

err = client.AddMember(ctx, createdGroup.ID, member)
```
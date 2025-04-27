# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Transfer Service Guide

_Last Updated: April 27, 2025_
_Compatible with SDK versions: v0.1.0 and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

This guide covers the Transfer service client in the Globus Go SDK, which enables file transfer operations between Globus endpoints.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Basic Operations](#basic-operations)
  - [Creating a Client](#creating-a-client)
  - [Listing Endpoints](#listing-endpoints)
  - [File Operations](#file-operations)
  - [Transfer Operations](#transfer-operations)
  - [Task Management](#task-management)
- [Advanced Features](#advanced-features)
  - [Recursive Transfers](#recursive-transfers)
  - [Resumable Transfers](#resumable-transfers)
  - [Memory-Optimized Transfers](#memory-optimized-transfers)
  - [Transfer Options](#transfer-options)
- [Error Handling](#error-handling)
- [Examples](#examples)
  - [Basic Transfer](#basic-transfer)
  - [Directory Transfer](#directory-transfer)
  - [Task Monitoring](#task-monitoring)
- [Troubleshooting](#troubleshooting)
- [Related Topics](#related-topics)

## Overview

The Transfer service allows you to move files and directories between Globus endpoints. The Globus Go SDK provides a client for interacting with the Transfer service API, enabling:

- File and directory operations
- Transfer task submission and management
- Endpoint management
- Optimized transfers for various scenarios

## Authentication

The Transfer client requires an access token with the `urn:globus:auth:scope:transfer.api.globus.org:all` scope. You can obtain this token using the Auth service:

```go
// Create an Auth client
authClient := pkg.NewAuthClient(clientID, clientSecret)

// Get an access token with transfer scope
tokenResp, err := authClient.GetClientCredentialsToken(
    context.Background(),
    []string{"urn:globus:auth:scope:transfer.api.globus.org:all"},
)
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

// Use the access token to create a Transfer client
accessToken := tokenResp.AccessToken
```

## Basic Operations

### Creating a Client

To use the Transfer service, create a client instance:

```go
// Create a client using the SDK configuration
config := pkg.NewConfigFromEnvironment()
transferClient := config.NewTransferClient(accessToken)

// Or create a client directly
transferClient := pkg.services.transfer.NewClient(accessToken)
```

### Listing Endpoints

To list endpoints you have access to:

```go
// List endpoints with default options
endpoints, err := transferClient.ListEndpoints(ctx, nil)
if err != nil {
    log.Fatalf("Failed to list endpoints: %v", err)
}

// Display endpoint information
for _, endpoint := range endpoints.DATA {
    fmt.Printf("ID: %s, Name: %s\n", endpoint.ID, endpoint.DisplayName)
}

// List with filtering options
filteredEndpoints, err := transferClient.ListEndpoints(ctx, &pkg.EndpointListOptions{
    FilterScope:     "my-endpoints",
    FilterOwnerType: "all",
    Limit:           100,
    Offset:          0,
})
```

### File Operations

List files in a directory:

```go
listing, err := transferClient.ListDirectory(ctx, endpointID, path, &pkg.ListOptions{
    ShowHidden: true,
    Limit:      1000,
})
if err != nil {
    log.Fatalf("Failed to list directory: %v", err)
}

for _, entry := range listing.DATA {
    if entry.Type == "file" {
        fmt.Printf("File: %s (Size: %d bytes)\n", entry.Name, entry.Size)
    } else if entry.Type == "dir" {
        fmt.Printf("Directory: %s\n", entry.Name)
    }
}
```

Create a directory:

```go
err := transferClient.MakeDirectory(ctx, endpointID, "/path/to/new/directory")
if err != nil {
    log.Fatalf("Failed to create directory: %v", err)
}
```

### Transfer Operations

Submit a basic transfer:

```go
transferData := &pkg.TransferData{
    Label: "SDK Example Transfer",
    Items: []pkg.TransferItem{
        {
            Source:      "/source/file.txt",
            Destination: "/destination/file.txt",
        },
        {
            Source:      "/source/data.csv",
            Destination: "/destination/data.csv",
        },
    },
    SyncLevel:        pkg.SyncChecksum,
    VerifyChecksum:   true,
    PreserveTimestamp: true,
}

task, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpointID,
    destinationEndpointID,
    transferData,
)
if err != nil {
    log.Fatalf("Transfer submission failed: %v", err)
}

fmt.Printf("Transfer submitted successfully. Task ID: %s\n", task.TaskID)
```

### Task Management

Check task status:

```go
status, err := transferClient.GetTaskStatus(ctx, taskID)
if err != nil {
    log.Fatalf("Failed to get task status: %v", err)
}

fmt.Printf("Task status: %s\n", status.Status)
fmt.Printf("Files transferred: %d/%d\n", status.FilesTransferred, status.FilesTotal)
```

List tasks:

```go
tasks, err := transferClient.ListTasks(ctx, &pkg.TaskListOptions{
    Limit:  100,
    Status: "ACTIVE",
})
if err != nil {
    log.Fatalf("Failed to list tasks: %v", err)
}

for _, task := range tasks.DATA {
    fmt.Printf("Task ID: %s, Label: %s, Status: %s\n", 
        task.TaskID, task.Label, task.Status)
}
```

Cancel a task:

```go
err := transferClient.CancelTask(ctx, taskID)
if err != nil {
    log.Fatalf("Failed to cancel task: %v", err)
}
```

## Advanced Features

### Recursive Transfers

For transferring entire directory structures:

```go
result, err := transferClient.SubmitRecursiveTransfer(
    ctx,
    sourceEndpointID, "/source/directory",
    destinationEndpointID, "/destination/directory",
    &pkg.RecursiveTransferOptions{
        Label:     "Recursive Directory Transfer",
        SyncLevel: pkg.SyncChecksum,
        BatchSize: 100,  // Process 100 files per batch
        MaxDepth:  -1,   // No limit on directory depth
    },
)
if err != nil {
    log.Fatalf("Recursive transfer failed: %v", err)
}

fmt.Printf("Submitted %d tasks\n", len(result.TaskIDs))
fmt.Printf("Transferred %d items\n", result.ItemsTransferred)
```

See [Recursive Transfers](../advanced/recursive-transfers.md) for more details.

### Resumable Transfers

For large transfers that may need to be resumed:

```go
// Start a resumable transfer
resumable, err := transferClient.StartResumableTransfer(
    ctx,
    sourceEndpointID, "/source/directory",
    destinationEndpointID, "/destination/directory",
    &pkg.ResumableTransferOptions{
        Label:            "Resumable Transfer",
        CheckpointInterval: 10 * time.Minute,
        SaveCheckpoint:   true,
    },
)
if err != nil {
    log.Fatalf("Failed to start resumable transfer: %v", err)
}

// Resume a previously started transfer
resumedTransfer, err := transferClient.ResumeTransfer(
    ctx,
    resumable.CheckpointID,
)
```

See [Resumable Transfers](../advanced/resumable-transfers.md) for more details.

### Memory-Optimized Transfers

For transfers of directories with many files:

```go
result, err := transferClient.SubmitMemoryOptimizedTransfer(
    ctx,
    sourceEndpointID, "/source/directory",
    destinationEndpointID, "/destination/directory",
    &pkg.MemoryOptimizedOptions{
        BatchSize:         100,
        MaxConcurrentTasks: 4,
        Label:             "Memory-Efficient Transfer",
        ProgressCallback: func(processed, total int, bytes int64, message string) {
            fmt.Printf("Progress: %d/%d files, %.2f MB\n", 
                processed, total, float64(bytes)/(1024*1024))
        },
    },
)
```

See [Performance Optimization](../topics/performance.md) for more details.

### Transfer Options

The SDK supports many transfer options:

```go
options := &pkg.TransferOptions{
    Label:             "Custom Transfer",
    SyncLevel:         pkg.SyncMtime,     // Compare modification times
    VerifyChecksum:    true,              // Verify checksums after transfer
    PreserveTimestamp: true,              // Preserve file timestamps
    EncryptData:       true,              // Request data encryption
    Deadline:          time.Now().Add(24 * time.Hour), // Set deadline
    NotificationTarget: "user@example.com", // Email notification
    FailOnQuotaErrors: true,              // Fail if quota exceeded
    SkipSourceErrors:  false,             // Don't skip source errors
}
```

## Error Handling

The Transfer client defines specific error types:

```go
// Check for specific error types
if err != nil {
    switch {
    case pkg.IsEndpointNotFoundError(err):
        fmt.Println("Endpoint not found")
    case pkg.IsPermissionDeniedError(err):
        fmt.Println("Permission denied")
    case pkg.IsRateLimitExceededError(err):
        fmt.Println("Rate limit exceeded, retry later")
    case pkg.IsTaskNotFoundError(err):
        fmt.Println("Task not found")
    case pkg.IsServiceUnavailableError(err):
        fmt.Println("Service temporarily unavailable")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

Implementing retries:

```go
// Retry with backoff
err := pkg.RetryWithBackoff(ctx, func(ctx context.Context) error {
    _, err := transferClient.GetTaskStatus(ctx, taskID)
    return err
}, &pkg.ExponentialBackoff{
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     30 * time.Second,
    MaxAttempts:  5,
}, pkg.IsRetryableError)
```

## Examples

### Basic Transfer

```go
func BasicTransfer(ctx context.Context, accessToken string) error {
    // Create the transfer client
    client := pkg.NewTransferClient(accessToken)
    
    // Define source and destination endpoints
    sourceEndpoint := "ddb59aef-6d04-11e5-ba46-22000b92c6ec"
    destEndpoint := "ddb59af0-6d04-11e5-ba46-22000b92c6ec"
    
    // Define the transfer
    transfer := &pkg.TransferData{
        Label: "Basic File Transfer",
        Items: []pkg.TransferItem{
            {
                Source:      "/source/file1.txt",
                Destination: "/destination/file1.txt",
            },
        },
        SyncLevel: pkg.SyncChecksum,
    }
    
    // Submit the transfer
    task, err := client.SubmitTransfer(
        ctx,
        sourceEndpoint,
        destEndpoint,
        transfer,
    )
    if err != nil {
        return fmt.Errorf("transfer submission failed: %w", err)
    }
    
    fmt.Printf("Transfer submitted. Task ID: %s\n", task.TaskID)
    return nil
}
```

### Directory Transfer

```go
func DirectoryTransfer(ctx context.Context, accessToken string) error {
    // Create the transfer client
    client := pkg.NewTransferClient(accessToken)
    
    // Define source and destination endpoints
    sourceEndpoint := "ddb59aef-6d04-11e5-ba46-22000b92c6ec"
    destEndpoint := "ddb59af0-6d04-11e5-ba46-22000b92c6ec"
    
    // Submit a recursive directory transfer
    result, err := client.SubmitRecursiveTransfer(
        ctx,
        sourceEndpoint, "/source/directory",
        destEndpoint, "/destination/directory",
        &pkg.RecursiveTransferOptions{
            Label:     "Directory Transfer Example",
            SyncLevel: pkg.SyncExistence,
            BatchSize: 100,
            DeleteSourceFiles: false,
        },
    )
    if err != nil {
        return fmt.Errorf("recursive transfer failed: %w", err)
    }
    
    fmt.Printf("Transfer submitted with %d tasks\n", len(result.TaskIDs))
    return nil
}
```

### Task Monitoring

```go
func MonitorTask(ctx context.Context, accessToken, taskID string) error {
    // Create the transfer client
    client := pkg.NewTransferClient(accessToken)
    
    // Set up a ticker to check status every 5 seconds
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            status, err := client.GetTaskStatus(ctx, taskID)
            if err != nil {
                return fmt.Errorf("failed to get task status: %w", err)
            }
            
            fmt.Printf("Status: %s, Progress: %d/%d files\n",
                status.Status, status.FilesTransferred, status.FilesTotal)
            
            // Check if task is complete
            if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
                if status.Status == "SUCCEEDED" {
                    fmt.Println("Transfer completed successfully!")
                } else {
                    fmt.Printf("Transfer failed: %s\n", status.NiceStatusShortDescription)
                }
                return nil
            }
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Endpoint Not Activated**

   ```
   Error: The endpoint could not be accessed. Either it is not activated or the network connection to the server is not available.
   ```

   **Solution**: Ensure the endpoint is activated. Check the endpoint's activation status via the web interface or use:

   ```go
   endpoint, err := client.GetEndpoint(ctx, endpointID)
   if !endpoint.Activated {
       fmt.Println("Endpoint needs activation")
   }
   ```

2. **Permission Denied**

   ```
   Error: Permission denied
   ```

   **Solution**: Verify that the user associated with the access token has appropriate permissions on both endpoints.

3. **Task Submission Failures**

   **Solution**: Verify that:
   - The access token is valid and has not expired
   - Both endpoints exist and are active
   - Source and destination paths are valid
   - The user has permission to access the specified paths

4. **Rate Limit Exceeded**

   ```
   Error: Rate limit exceeded
   ```

   **Solution**: Implement rate limiting and retry logic:

   ```go
   // Create a client with rate limiting
   config := pkg.NewConfigFromEnvironment().
       WithRateLimit(5.0).  // 5 requests per second
       WithRetryCount(3)    // Retry up to 3 times
   client := config.NewTransferClient(accessToken)
   ```

## Related Topics

- [Recursive Transfers](../advanced/recursive-transfers.md)
- [Resumable Transfers](../advanced/resumable-transfers.md)
- [Performance Optimization](../topics/performance.md)
- [Authentication Guide](authentication.md)
- [Error Handling](../topics/error-handling.md)
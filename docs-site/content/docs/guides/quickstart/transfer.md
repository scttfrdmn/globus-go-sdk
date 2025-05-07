---
title: "Transfer Service Quick Start"
weight: 20
---

# Transfer Service Quick Start

This guide will help you get started with the Globus Transfer service using the Go SDK. The Transfer service enables file transfers between Globus endpoints, endpoint management, and tasks for data movement.

## Setup

First, import the required packages and create a context:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()
    
    // Continue with the examples below...
}
```

## Creating a Transfer Client

There are two main ways to create a Transfer client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Transfer client
transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create transfer client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Transfer client with options
transferClient, err := transfer.NewClient(
    transfer.WithAccessToken(os.Getenv("GLOBUS_ACCESS_TOKEN")),
    transfer.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create transfer client: %v", err)
}
```

## Working with Endpoints

Endpoints are locations where your files are stored and can be accessed via Globus.

### Listing Your Endpoints

```go
// List all endpoints you have access to
endpoints, err := transferClient.ListEndpoints(ctx, nil)
if err != nil {
    log.Fatalf("Failed to list endpoints: %v", err)
}

fmt.Printf("Found %d endpoints:\n", len(endpoints.DATA))
for i, endpoint := range endpoints.DATA {
    fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
}
```

### Filtering Endpoints

```go
// List only your endpoints with specific filtering
endpoints, err := transferClient.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    Filter:       "my-endpoints", // Only show endpoints you own
    Limit:        25,             // Limit the number of results
    SearchString: "cluster",      // Search for endpoints containing "cluster"
})
if err != nil {
    log.Fatalf("Failed to list endpoints: %v", err)
}

fmt.Printf("Found %d matching endpoints\n", len(endpoints.DATA))
```

### Getting Endpoint Details

```go
// Get details about a specific endpoint
endpointID := "your-endpoint-id"
endpoint, err := transferClient.GetEndpoint(ctx, endpointID)
if err != nil {
    log.Fatalf("Failed to get endpoint: %v", err)
}

fmt.Printf("Endpoint: %s\n", endpoint.DisplayName)
fmt.Printf("Description: %s\n", endpoint.Description)
fmt.Printf("Owner: %s\n", endpoint.OwnerString)
fmt.Printf("Activated: %t\n", endpoint.Activated)
```

## Working with Files and Directories

### Listing Files in a Directory

```go
// List files in a directory on an endpoint
endpointID := "your-endpoint-id"
path := "/path/to/directory"

files, err := transferClient.ListDirectory(ctx, endpointID, path, nil)
if err != nil {
    log.Fatalf("Failed to list directory: %v", err)
}

fmt.Printf("Files in %s:\n", path)
for _, file := range files.DATA {
    if file.Type == "file" {
        fmt.Printf("File: %s (Size: %d bytes)\n", file.Name, file.Size)
    } else {
        fmt.Printf("Directory: %s/\n", file.Name)
    }
}
```

### Listing Files with Options

```go
// List files with additional options
files, err := transferClient.ListDirectory(ctx, endpointID, path, &transfer.ListDirectoryOptions{
    ShowHidden:    true,        // Show hidden files
    OrderBy:       "name",      // Sort by name
    SortDirection: "asc",       // Sort in ascending order
    Limit:         100,         // Limit results to 100 files
})
if err != nil {
    log.Fatalf("Failed to list directory: %v", err)
}
```

### Creating a Directory

```go
// Create a new directory on an endpoint
newPath := "/path/to/new/directory"
result, err := transferClient.CreateDirectory(ctx, endpointID, newPath)
if err != nil {
    log.Fatalf("Failed to create directory: %v", err)
}

if result.Code == "DirectoryCreated" {
    fmt.Printf("Directory created: %s\n", newPath)
} else {
    fmt.Printf("Response: %s - %s\n", result.Code, result.Message)
}
```

### Renaming a File or Directory

```go
// Rename a file or directory
oldPath := "/path/to/old/filename"
newPath := "/path/to/new/filename"

result, err := transferClient.Rename(ctx, endpointID, oldPath, newPath)
if err != nil {
    log.Fatalf("Failed to rename: %v", err)
}

if result.Code == "FileRenamed" {
    fmt.Printf("Renamed %s to %s\n", oldPath, newPath)
}
```

## Basic File Transfers

### Simple File Transfer

```go
// Transfer a single file between two endpoints
sourceEndpointID := "source-endpoint-id"
destinationEndpointID := "destination-endpoint-id"
sourcePath := "/path/on/source/file.txt"
destPath := "/path/on/destination/file.txt"

// Submit the transfer
task, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpointID,
    destinationEndpointID,
    &transfer.TransferData{
        Label: "Simple File Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      sourcePath,
                Destination: destPath,
            },
        },
    },
)
if err != nil {
    log.Fatalf("Failed to submit transfer: %v", err)
}

fmt.Printf("Transfer task submitted with ID: %s\n", task.TaskID)
```

### Transferring Multiple Files

```go
// Transfer multiple files in a single task
task, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpointID,
    destinationEndpointID,
    &transfer.TransferData{
        Label: "Multiple File Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/on/source/file1.txt",
                Destination: "/path/on/destination/file1.txt",
            },
            {
                Source:      "/path/on/source/file2.txt",
                Destination: "/path/on/destination/file2.txt",
            },
            {
                Source:      "/path/on/source/file3.txt",
                Destination: "/path/on/destination/file3.txt",
            },
        },
    },
)
if err != nil {
    log.Fatalf("Failed to submit transfer: %v", err)
}
```

### Directory Transfer

```go
// Transfer a directory recursively
task, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpointID,
    destinationEndpointID,
    &transfer.TransferData{
        Label: "Directory Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/directory",
                Destination: "/path/to/destination/directory",
                Recursive:   true, // Enable recursive transfer for directories
            },
        },
    },
)
if err != nil {
    log.Fatalf("Failed to submit transfer: %v", err)
}
```

### Transfer with Sync Levels

Sync levels control when files should be transferred:

```go
// Transfer with checksum verification
task, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpointID,
    destinationEndpointID,
    &transfer.TransferData{
        Label: "Transfer with Sync Level",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        // Options for SyncLevel:
        // - SyncExists: Only transfer if destination doesn't exist
        // - SyncSize: Transfer if file size differs
        // - SyncMtime: Transfer if modification time differs
        // - SyncChecksum: Transfer if checksums differ (most accurate, but slowest)
        SyncLevel: transfer.SyncChecksum,
        Verify:    true, // Verify transfer after completion
    },
)
if err != nil {
    log.Fatalf("Failed to submit transfer: %v", err)
}
```

## Monitoring Transfer Tasks

### Getting Task Status

```go
// Get the status of a transfer task
taskID := task.TaskID
taskStatus, err := transferClient.GetTask(ctx, taskID)
if err != nil {
    log.Fatalf("Failed to get task status: %v", err)
}

fmt.Printf("Task: %s\n", taskStatus.Label)
fmt.Printf("Status: %s\n", taskStatus.Status)
fmt.Printf("Files: %d/%d\n", taskStatus.FilesTransferred, taskStatus.FilesTotal)
fmt.Printf("Bytes: %d/%d\n", taskStatus.BytesTransferred, taskStatus.BytesTotal)
```

### Waiting for Task Completion

```go
// Wait for a task to complete with progress updates
taskStatus, err := transferClient.WaitForTaskCompletion(ctx, taskID, &transfer.WaitOptions{
    Timeout:      20 * time.Minute,  // Maximum time to wait
    PollInterval: 5 * time.Second,   // How often to check status
    ProgressFunc: func(status *transfer.Task) {
        fmt.Printf("Progress: %d/%d files (%d%%)\n", 
            status.FilesTransferred, 
            status.FilesTotal,
            calculatePercentage(status.BytesTransferred, status.BytesTotal),
        )
    },
})
if err != nil {
    log.Fatalf("Task monitoring failed: %v", err)
}

if taskStatus.Status == "SUCCEEDED" {
    fmt.Println("Transfer completed successfully!")
} else {
    fmt.Printf("Transfer ended with status: %s\n", taskStatus.Status)
}

// Helper function to calculate percentage
func calculatePercentage(current, total int64) int {
    if total == 0 {
        return 0
    }
    return int((float64(current) / float64(total)) * 100)
}
```

### Listing Task Events

```go
// Get events for a task to debug or monitor
events, err := transferClient.GetTaskEventList(ctx, taskID, nil)
if err != nil {
    log.Fatalf("Failed to get task events: %v", err)
}

fmt.Println("Task events:")
for _, event := range events.DATA {
    fmt.Printf("%s: %s (%s)\n", event.Time, event.Description, event.Code)
}
```

## Deleting Files

```go
// Get a submission ID
submissionID, err := transferClient.GetSubmissionID(ctx)
if err != nil {
    log.Fatalf("Failed to get submission ID: %v", err)
}

// Create a delete task
request := &transfer.DeleteTaskRequest{
    SubmissionID: submissionID,
    EndpointID:   endpointID,
    Label:        "Delete Files",
    DATA: []transfer.DeleteItem{
        {
            DataType: "delete_item",
            Path:     "/path/to/file/to/delete.txt",
        },
        {
            DataType: "delete_item",
            Path:     "/path/to/directory/to/delete", // Directories are deleted recursively by default
        },
    },
}

deleteTask, err := transferClient.CreateDeleteTask(ctx, request)
if err != nil {
    log.Fatalf("Failed to create delete task: %v", err)
}

fmt.Printf("Delete task submitted with ID: %s\n", deleteTask.TaskID)
```

## Canceling a Task

```go
// Cancel a running task
cancelResult, err := transferClient.CancelTask(ctx, taskID)
if err != nil {
    log.Fatalf("Failed to cancel task: %v", err)
}

if cancelResult.Code == "Canceled" {
    fmt.Println("Task canceled successfully")
} else {
    fmt.Printf("Cancel result: %s - %s\n", cancelResult.Code, cancelResult.Message)
}
```

## Error Handling

The Transfer service provides specific error types for better error handling:

```go
// Try an operation that might fail
_, err := transferClient.GetEndpoint(ctx, "non-existent-endpoint")
if err != nil {
    switch {
    case transfer.IsResourceNotFound(err):
        fmt.Println("Endpoint not found - check the endpoint ID")
    case transfer.IsPermissionDenied(err):
        fmt.Println("Permission denied - check your access token and permissions")
    case transfer.IsServiceUnavailable(err):
        fmt.Println("Service unavailable - try again later")
    case transfer.IsRateLimited(err):
        fmt.Println("Rate limited - slow down your requests")
    case transfer.IsRetryableTransferError(err):
        fmt.Println("Retryable error - you can retry this operation")
    default:
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Complete Example

Here's a complete example that creates a directory on a destination endpoint and transfers a file to it:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Get required environment variables
    sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
    destEndpointID := os.Getenv("DEST_ENDPOINT_ID")
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    if sourceEndpointID == "" || destEndpointID == "" || accessToken == "" {
        log.Fatalf("Missing required environment variables")
    }
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create transfer client
    transferClient, err := config.NewTransferClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // 1. Create a directory on the destination endpoint
    destPath := "/~/transfer-example"
    _, err = transferClient.CreateDirectory(ctx, destEndpointID, destPath)
    if err != nil {
        log.Fatalf("Failed to create directory: %v", err)
    }
    fmt.Printf("Created directory: %s\n", destPath)
    
    // 2. Submit a transfer task
    sourcePath := "/~/example-data.txt"
    task, err := transferClient.SubmitTransfer(
        ctx,
        sourceEndpointID,
        destEndpointID,
        &transfer.TransferData{
            Label: "Example Transfer",
            Items: []transfer.TransferItem{
                {
                    Source:      sourcePath,
                    Destination: destPath + "/example-data.txt",
                },
            },
            SyncLevel: transfer.SyncChecksum,
            Verify:    true,
        },
    )
    if err != nil {
        log.Fatalf("Failed to submit transfer: %v", err)
    }
    fmt.Printf("Transfer submitted with task ID: %s\n", task.TaskID)
    
    // 3. Wait for the task to complete with progress updates
    taskID := task.TaskID
    fmt.Println("Waiting for transfer to complete...")
    
    completedTask, err := transferClient.WaitForTaskCompletion(ctx, taskID, &transfer.WaitOptions{
        Timeout:      10 * time.Minute,
        PollInterval: 3 * time.Second,
    })
    if err != nil {
        log.Fatalf("Error waiting for task: %v", err)
    }
    
    // 4. Display final status
    if completedTask.Status == "SUCCEEDED" {
        fmt.Println("Transfer completed successfully!")
        fmt.Printf("Transferred %d files (%d bytes)\n", 
            completedTask.FilesTransferred, 
            completedTask.BytesTransferred)
    } else {
        fmt.Printf("Transfer ended with status: %s\n", completedTask.Status)
        
        // 5. Get task events to see what happened if not successful
        events, err := transferClient.GetTaskEventList(ctx, taskID, &transfer.GetTaskEventListOptions{
            Limit: 5, // Get the last 5 events
        })
        if err != nil {
            log.Printf("Failed to get events: %v", err)
        } else {
            fmt.Println("Last events:")
            for _, event := range events.DATA {
                fmt.Printf("- %s: %s\n", event.Time, event.Description)
            }
        }
    }
}
```

## Next Steps

Now that you understand the basics of the Transfer service, you can:

1. **Implement recursive transfers**: Use the recursive transfer methods for efficiently transferring entire directory structures
2. **Add progress monitoring**: Implement a progress bar or other UI elements to track transfers
3. **Handle errors gracefully**: Add retry logic for transient errors
4. **Optimize for large transfers**: Use appropriate sync levels and batch operations

For more details, check out the [Transfer Service API Reference](/docs/reference/transfer/) and the [Recursive Transfers](/docs/reference/transfer/recursive) and [Resumable Transfers](/docs/reference/transfer/resumable) guides.
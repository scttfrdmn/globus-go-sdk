# Transfer Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Transfer client provides access to the Globus Transfer API, which handles file transfers between Globus endpoints, endpoint management, and task operations.

## Client Structure

```go
type Client struct {
    client    *core.Client
    transport *core.HTTPTransport
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*core.Client` | Core client for making HTTP requests |
| `transport` | `*core.HTTPTransport` | HTTP transport for request/response handling |

## Creating a Transfer Client

```go
// Create a transfer client with options
client, err := transfer.NewClient(
    transfer.WithAccessToken("access-token"),
    transfer.WithHTTPDebugging(),
)
if err != nil {
    // Handle error
}
```

### Options

| Option | Description |
|--------|-------------|
| `WithAccessToken(token string)` | Sets the access token for authorization |
| `WithAuthorizer(auth core.Authorizer)` | Sets a custom authorizer (alternative to access token) |
| `WithBaseURL(url string)` | Sets a custom base URL (default: "https://transfer.api.globus.org/v0.10/") |
| `WithHTTPDebugging()` | Enables HTTP debugging |
| `WithHTTPTracing()` | Enables HTTP tracing |
| `WithRateLimiter(limiter RateLimiter)` | Sets a custom rate limiter |

## Endpoint Operations

### Listing Endpoints

```go
// List all endpoints
endpoints, err := client.ListEndpoints(ctx, nil)
if err != nil {
    // Handle error
}

// List endpoints with filtering options
endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    Filter:       "my-endpoints",
    Limit:        100,
    OwnerID:      "owner-id",
    SearchString: "cluster",
})
if err != nil {
    // Handle error
}

// Iterate through endpoints
for _, endpoint := range endpoints.DATA {
    fmt.Printf("Endpoint: %s (%s)\n", endpoint.DisplayName, endpoint.ID)
}
```

### Getting an Endpoint

```go
// Get a specific endpoint by ID
endpoint, err := client.GetEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

fmt.Printf("Endpoint Name: %s\n", endpoint.DisplayName)
fmt.Printf("Description: %s\n", endpoint.Description)
fmt.Printf("Owner: %s\n", endpoint.OwnerString)
```

## File Operations

### Listing Files

```go
// List files in a directory
files, err := client.ListDirectory(ctx, "endpoint-id", "/path/to/directory", nil)
if err != nil {
    // Handle error
}

// List files with options
files, err := client.ListDirectory(ctx, "endpoint-id", "/path/to/directory", &transfer.ListDirectoryOptions{
    Limit:           100,
    ShowHidden:      true,
    OrderBy:         "name",
    SortDirection:   "asc",
    IncludeMetadata: true,
})
if err != nil {
    // Handle error
}

// Iterate through files
for _, file := range files.DATA {
    if file.Type == "file" {
        fmt.Printf("File: %s (Size: %d)\n", file.Name, file.Size)
    } else {
        fmt.Printf("Directory: %s\n", file.Name)
    }
}
```

### Creating a Directory

```go
// Create a directory
result, err := client.CreateDirectory(ctx, "endpoint-id", "/path/to/new/directory")
if err != nil {
    // Handle error
}

if result.Code == "DirectoryCreated" {
    fmt.Println("Directory created successfully")
}
```

### Renaming a File or Directory

```go
// Rename a file or directory
result, err := client.Rename(ctx, "endpoint-id", "/path/to/old", "/path/to/new")
if err != nil {
    // Handle error
}

if result.Code == "FileRenamed" {
    fmt.Println("File renamed successfully")
}
```

## Transfer Operations

### Basic Transfer

```go
// Get a submission ID
submissionID, err := client.GetSubmissionID(ctx)
if err != nil {
    // Handle error
}

// Create a transfer task
request := &transfer.TransferTaskRequest{
    SubmissionID:    submissionID,
    SourceEndpointID: "source-endpoint-id",
    DestinationEndpointID: "destination-endpoint-id",
    Label:           "My Transfer",
    SyncLevel:       transfer.SyncChecksum,
    DATA: []transfer.TransferItem{
        {
            Source:      "/path/to/source/file.txt",
            Destination: "/path/to/destination/file.txt",
        },
        {
            Source:      "/path/to/source/directory",
            Destination: "/path/to/destination/directory",
            Recursive:   true,
        },
    },
}

taskResponse, err := client.CreateTransferTask(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Transfer task created: %s\n", taskResponse.TaskID)
```

### Simplified Transfer

```go
// Submit a transfer using the simplified helper method
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Simple Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        SyncLevel: transfer.SyncChecksum,
        Verify:    true,
    },
)
if err != nil {
    // Handle error
}

fmt.Printf("Transfer task created: %s\n", taskResponse.TaskID)
```

### Delete Operation

```go
// Get a submission ID
submissionID, err := client.GetSubmissionID(ctx)
if err != nil {
    // Handle error
}

// Create a delete task
request := &transfer.DeleteTaskRequest{
    SubmissionID: submissionID,
    EndpointID:   "endpoint-id",
    Label:        "My Deletion",
    DATA: []transfer.DeleteItem{
        {
            Path: "/path/to/file/to/delete.txt",
        },
        {
            Path:      "/path/to/directory/to/delete",
            Recursive: true,
        },
    },
}

taskResponse, err := client.CreateDeleteTask(ctx, request)
if err != nil {
    // Handle error
}

fmt.Printf("Delete task created: %s\n", taskResponse.TaskID)
```

## Task Management

### Listing Tasks

```go
// List all tasks
tasks, err := client.ListTasks(ctx, nil)
if err != nil {
    // Handle error
}

// List tasks with filtering options
tasks, err := client.ListTasks(ctx, &transfer.ListTasksOptions{
    Limit:  100,
    Fields: []string{"task_id", "status", "label"},
    Filter: "status:ACTIVE,type:TRANSFER",
})
if err != nil {
    // Handle error
}

// Iterate through tasks
for _, task := range tasks.DATA {
    fmt.Printf("Task: %s (%s) - Status: %s\n", task.Label, task.TaskID, task.Status)
}
```

### Getting Task Details

```go
// Get details for a specific task
task, err := client.GetTask(ctx, "task-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task: %s\n", task.Label)
fmt.Printf("Status: %s\n", task.Status)
fmt.Printf("Files: %d/%d\n", task.FilesTransferred, task.FilesTotal)
fmt.Printf("Bytes: %d/%d\n", task.BytesTransferred, task.BytesTotal)
```

### Getting Task Event List

```go
// Get task events
events, err := client.GetTaskEventList(ctx, "task-id", nil)
if err != nil {
    // Handle error
}

// Get task events with options
events, err := client.GetTaskEventList(ctx, "task-id", &transfer.GetTaskEventListOptions{
    Limit:  100,
    Fields: []string{"time", "description", "code"},
})
if err != nil {
    // Handle error
}

// Iterate through events
for _, event := range events.DATA {
    fmt.Printf("%s: %s (%s)\n", event.Time, event.Description, event.Code)
}
```

### Canceling a Task

```go
// Cancel a task
cancelResult, err := client.CancelTask(ctx, "task-id")
if err != nil {
    // Handle error
}

if cancelResult.Code == "Canceled" {
    fmt.Println("Task canceled successfully")
}
```

## Waiting for Task Completion

```go
// Wait for a task to complete with default options
task, err := client.WaitForTaskCompletion(ctx, "task-id", nil)
if err != nil {
    // Handle error
}

// Wait with custom options
task, err := client.WaitForTaskCompletion(ctx, "task-id", &transfer.WaitOptions{
    Timeout:      5 * time.Minute,
    PollInterval: 5 * time.Second,
})
if err != nil {
    // Handle error
}

fmt.Printf("Task completed with status: %s\n", task.Status)
```

## Error Handling

The transfer package provides specific error types and helper functions for common error conditions:

```go
// Try an operation
_, err := client.GetEndpoint(ctx, "non-existent-endpoint")
if err != nil {
    switch {
    case transfer.IsResourceNotFound(err):
        fmt.Println("Endpoint not found")
    case transfer.IsPermissionDenied(err):
        fmt.Println("Permission denied")
    case transfer.IsServiceUnavailable(err):
        fmt.Println("Service unavailable - try again later")
    case transfer.IsRetryableTransferError(err):
        fmt.Println("Retryable error - can be retried")
    default:
        fmt.Println("Other error:", err)
    }
}
```

### Error Details

```go
// Get more details from a transfer error
if transferErr, ok := err.(*transfer.TransferError); ok {
    fmt.Println("Error code:", transferErr.Code)
    fmt.Println("Error message:", transferErr.Message)
    fmt.Println("Request ID:", transferErr.RequestID)
}
```

## Common Patterns

### Transfer with Sync Level

```go
// Transfer with different sync levels
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Transfer with Sync Level",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        // Choose a sync level:
        // SyncExists - Only transfer if destination doesn't exist
        // SyncSize - Transfer if size differs (faster, less accurate)
        // SyncMtime - Transfer if modification time differs
        // SyncChecksum - Transfer if checksum differs (slower, most accurate)
        SyncLevel: transfer.SyncChecksum,
    },
)
```

### Directory Creation and Transfer

```go
// Create a directory and then transfer to it
_, err := client.CreateDirectory(ctx, "endpoint-id", "/path/to/new/directory")
if err != nil {
    // Handle error
}

// Transfer to the new directory
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "endpoint-id",
    &transfer.TransferData{
        Label: "Transfer to New Directory",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/new/directory/file.txt",
            },
        },
    },
)
```

### Transfer with Verification

```go
// Transfer with verification
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Transfer with Verification",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        SyncLevel: transfer.SyncChecksum,
        Verify:    true, // Enable verification
    },
)
```

### Transfer with Recursive Directory

```go
// Transfer a directory recursively
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Recursive Directory Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/directory",
                Destination: "/path/to/destination/directory",
                Recursive:   true, // Enable recursive transfer
            },
        },
    },
)
```

### Transfer with Deadline

```go
// Create a context with deadline
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
defer cancel()

// Submit a transfer with the deadline context
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Transfer with Deadline",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
    },
)
if err != nil {
    // Check if deadline exceeded
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Transfer request timed out")
    } else {
        fmt.Println("Other error:", err)
    }
}
```

### Listing All Files in a Directory Recursively

```go
// Helper function to recursively list files
func listAllFiles(ctx context.Context, client *transfer.Client, endpointID, path string) ([]transfer.FileListItem, error) {
    var allFiles []transfer.FileListItem
    
    files, err := client.ListDirectory(ctx, endpointID, path, nil)
    if err != nil {
        return nil, err
    }
    
    for _, file := range files.DATA {
        if file.Type == "dir" {
            // Recursively list files in directory
            subFiles, err := listAllFiles(ctx, client, endpointID, path+"/"+file.Name)
            if err != nil {
                return nil, err
            }
            allFiles = append(allFiles, subFiles...)
        } else {
            allFiles = append(allFiles, file)
        }
    }
    
    return allFiles, nil
}

// Use the helper function
allFiles, err := listAllFiles(ctx, client, "endpoint-id", "/path/to/directory")
if err != nil {
    // Handle error
}

fmt.Printf("Found %d files recursively\n", len(allFiles))
```

## Best Practices

1. Always use context for cancellation and deadlines
2. Handle rate limiting by respecting retry-after headers
3. Use appropriate sync levels to balance speed and accuracy
4. Use error type checking for specific error handling
5. Consider using the recursive transfer methods for directory transfers
6. Use the wait methods for waiting on task completion
7. Implement retry logic for retryable errors
8. Use the simplified helper methods for common operations
9. Check task status regularly for long-running transfers
10. Use transfer events to track progress and diagnose issues
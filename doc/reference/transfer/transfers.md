# Transfer Service: Transfer Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Transfer operations allow you to move files and directories between Globus endpoints. The Transfer client provides methods for initiating, monitoring, and managing these operations.

## Transfer Types

The Globus Transfer service supports several types of transfer operations:

- **Basic Transfers**: Simple file and directory transfers
- **Recursive Transfers**: Transfers of entire directory structures
- **Resumable Transfers**: Transfers that can be paused and resumed
- **Delete Operations**: Removal of files and directories from endpoints

## Basic Transfer Structure

```go
type TransferTaskRequest struct {
    SubmissionID           string          `json:"submission_id"`
    SourceEndpointID       string          `json:"source_endpoint"`
    DestinationEndpointID  string          `json:"destination_endpoint"`
    Label                  string          `json:"label,omitempty"`
    SyncLevel              int             `json:"sync_level,omitempty"`
    Verify                 bool            `json:"verify_checksum,omitempty"`
    PreserveTimestamp      bool            `json:"preserve_timestamp,omitempty"`
    EncryptData            bool            `json:"encrypt_data,omitempty"`
    DeadlineSeconds        int             `json:"deadline,omitempty"`
    SkipSourceErrors       bool            `json:"skip_source_errors,omitempty"`
    FailOnQuotaErrors      bool            `json:"fail_on_quota_errors,omitempty"`
    DATA                   []TransferItem  `json:"DATA"`
}

type TransferItem struct {
    Source      string `json:"source_path"`
    Destination string `json:"destination_path"`
    Recursive   bool   `json:"recursive,omitempty"`
}
```

## Simplified Transfer Helper

For common transfer operations, the SDK provides a simplified helper method:

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

### TransferData Structure

The `TransferData` struct is used with the simplified helper method:

```go
type TransferData struct {
    Label              string         `json:"label,omitempty"`
    Items              []TransferItem `json:"items"`
    SyncLevel          int            `json:"sync_level,omitempty"`
    Verify             bool           `json:"verify_checksum,omitempty"`
    PreserveTimestamp  bool           `json:"preserve_timestamp,omitempty"`
    EncryptData        bool           `json:"encrypt_data,omitempty"`
    DeadlineSeconds    int            `json:"deadline,omitempty"`
    SkipSourceErrors   bool           `json:"skip_source_errors,omitempty"`
    FailOnQuotaErrors  bool           `json:"fail_on_quota_errors,omitempty"`
}
```

## Manual Transfer Process

For more control, you can use the lower-level API:

```go
// Step 1: Get a submission ID
submissionID, err := client.GetSubmissionID(ctx)
if err != nil {
    // Handle error
}

// Step 2: Create a transfer task
request := &transfer.TransferTaskRequest{
    SubmissionID:          submissionID,
    SourceEndpointID:      "source-endpoint-id",
    DestinationEndpointID: "destination-endpoint-id",
    Label:                 "My Transfer",
    SyncLevel:             transfer.SyncChecksum,
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

## Synchronization Levels

The Transfer service supports different synchronization levels that determine when files are transferred:

```go
const (
    SyncExists   = 0 // Only transfer if destination doesn't exist
    SyncSize     = 1 // Transfer if size differs (faster, less accurate)
    SyncMtime    = 2 // Transfer if modification time differs
    SyncChecksum = 3 // Transfer if checksum differs (slower, most accurate)
)
```

```go
// Transfer with SyncChecksum (most accurate)
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Checksum Sync Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        SyncLevel: transfer.SyncChecksum,
    },
)
```

## Transferring Directories

To transfer directories, you can use the `Recursive` flag:

```go
// Transfer a directory recursively
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Directory Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/directory",
                Destination: "/path/to/destination/directory",
                Recursive:   true,
            },
        },
    },
)
```

## Delete Operations

You can also submit tasks to delete files and directories:

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

## Monitoring Transfer Tasks

Once a transfer task is created, you can monitor its progress:

```go
// Get task status
task, err := client.GetTask(ctx, "task-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task: %s\n", task.Label)
fmt.Printf("Status: %s\n", task.Status)
fmt.Printf("Files: %d/%d\n", task.FilesTransferred, task.FilesTotal)
fmt.Printf("Bytes: %d/%d\n", task.BytesTransferred, task.BytesTotal)
```

### Waiting for Task Completion

For convenience, the SDK provides a method to wait for task completion:

```go
// Wait for a task to complete
task, err := client.WaitForTaskCompletion(ctx, "task-id", nil)
if err != nil {
    // Handle error
}

if task.Status == "SUCCEEDED" {
    fmt.Println("Transfer completed successfully")
} else {
    fmt.Printf("Transfer failed with status: %s\n", task.Status)
}
```

### Custom Wait Options

You can customize the wait behavior:

```go
// Wait with custom options
task, err := client.WaitForTaskCompletion(ctx, "task-id", &transfer.WaitOptions{
    Timeout:      5 * time.Minute,
    PollInterval: 5 * time.Second,
})
```

## Task Status Codes

Transfer tasks can have the following status codes:

| Status | Description |
|--------|-------------|
| `"ACTIVE"` | Task is queued or executing |
| `"INACTIVE"` | Task is waiting for an activation |
| `"SUCCEEDED"` | Task completed successfully |
| `"FAILED"` | Task failed |
| `"CANCELED"` | Task was canceled |

## Task Event Tracking

To get detailed information about a task's progress and any issues:

```go
// Get task events
events, err := client.GetTaskEventList(ctx, "task-id", nil)
if err != nil {
    // Handle error
}

for _, event := range events.DATA {
    fmt.Printf("%s: %s (%s)\n", event.Time, event.Description, event.Code)
}
```

## Transfer Options

### Verification

Enable checksum verification after transfer:

```go
// Transfer with verification
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Verified Transfer",
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

### Timestamp Preservation

Preserve modification times of transferred files:

```go
// Transfer with timestamp preservation
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Timestamp Preserved Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        PreserveTimestamp: true, // Preserve modification times
    },
)
```

### Encryption

Enable encryption for data in transit:

```go
// Transfer with encryption
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Encrypted Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        EncryptData: true, // Enable encryption
    },
)
```

### Deadlines

Set a deadline for transfer completion:

```go
// Transfer with deadline
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Deadline Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/file.txt",
                Destination: "/path/to/destination/file.txt",
            },
        },
        DeadlineSeconds: 3600, // 1 hour deadline
    },
)
```

### Error Handling Options

Control how errors are handled during transfer:

```go
// Transfer with error handling options
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Error Handling Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/directory",
                Destination: "/path/to/destination/directory",
                Recursive:   true,
            },
        },
        SkipSourceErrors:  true,  // Skip errors reading source files
        FailOnQuotaErrors: false, // Continue if quota errors occur
    },
)
```

## Transfer Performance Tips

### Batch Transfers

For better performance with many small files, batch them into a single transfer request:

```go
// Create a batch of transfer items
var items []transfer.TransferItem
for i := 1; i <= 100; i++ {
    items = append(items, transfer.TransferItem{
        Source:      fmt.Sprintf("/path/to/source/file%d.txt", i),
        Destination: fmt.Sprintf("/path/to/destination/file%d.txt", i),
    })
}

// Submit as a single transfer
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label: "Batch Transfer",
        Items: items,
    },
)
```

### Choose Appropriate Sync Level

Select the appropriate sync level based on your needs:

- **SyncExists (0)**: Fastest, but only checks if the destination file exists
- **SyncSize (1)**: Checks file size, good balance of speed and accuracy
- **SyncMtime (2)**: Checks modification time, more accurate but slower
- **SyncChecksum (3)**: Most accurate, but slowest

### Concurrent Transfers

For transferring many files, consider using multiple transfer tasks in parallel:

```go
// Split large transfers into manageable chunks
func splitTransfer(items []transfer.TransferItem, chunkSize int) [][]transfer.TransferItem {
    var chunks [][]transfer.TransferItem
    for i := 0; i < len(items); i += chunkSize {
        end := i + chunkSize
        if end > len(items) {
            end = len(items)
        }
        chunks = append(chunks, items[i:end])
    }
    return chunks
}

// Submit concurrent transfer tasks
var taskIDs []string
chunks := splitTransfer(allItems, 100)
for i, chunk := range chunks {
    taskResponse, err := client.SubmitTransfer(
        ctx,
        "source-endpoint-id",
        "destination-endpoint-id",
        &transfer.TransferData{
            Label: fmt.Sprintf("Concurrent Transfer %d", i+1),
            Items: chunk,
        },
    )
    if err != nil {
        // Handle error
        continue
    }
    taskIDs = append(taskIDs, taskResponse.TaskID)
}
```

## Common Patterns

### Transfer with Verification and Error Handling

```go
// Comprehensive transfer with all options
taskResponse, err := client.SubmitTransfer(
    ctx,
    "source-endpoint-id",
    "destination-endpoint-id",
    &transfer.TransferData{
        Label:             "Comprehensive Transfer",
        Items: []transfer.TransferItem{
            {
                Source:      "/path/to/source/directory",
                Destination: "/path/to/destination/directory",
                Recursive:   true,
            },
        },
        SyncLevel:          transfer.SyncChecksum,
        Verify:             true,
        PreserveTimestamp:  true,
        EncryptData:        true,
        DeadlineSeconds:    7200,  // 2 hours
        SkipSourceErrors:   false, // Fail on any source error
        FailOnQuotaErrors:  true,  // Fail if quota exceeded
    },
)
if err != nil {
    // Handle error
    return
}

// Monitor task progress with events
for {
    task, err := client.GetTask(ctx, taskResponse.TaskID)
    if err != nil {
        // Handle error
        break
    }
    
    // Print progress
    fmt.Printf("\rProgress: %d/%d files, %d/%d bytes",
        task.FilesTransferred, task.FilesTotal,
        task.BytesTransferred, task.BytesTotal)
    
    if task.Status != "ACTIVE" {
        fmt.Printf("\nTask completed with status: %s\n", task.Status)
        break
    }
    
    time.Sleep(3 * time.Second)
}
```

### Transfer Directory with Filtering

For more complex transfers, you can combine directory listing with transfers:

```go
// List files in a directory
files, err := client.ListDirectory(ctx, "source-endpoint-id", "/path/to/source", nil)
if err != nil {
    // Handle error
}

// Filter files (e.g., only .txt files)
var items []transfer.TransferItem
for _, file := range files.DATA {
    if file.Type == "file" && strings.HasSuffix(file.Name, ".txt") {
        items = append(items, transfer.TransferItem{
            Source:      "/path/to/source/" + file.Name,
            Destination: "/path/to/destination/" + file.Name,
        })
    }
}

// Submit transfer with filtered files
if len(items) > 0 {
    taskResponse, err := client.SubmitTransfer(
        ctx,
        "source-endpoint-id",
        "destination-endpoint-id",
        &transfer.TransferData{
            Label: "Filtered Transfer",
            Items: items,
        },
    )
    if err != nil {
        // Handle error
    }
    
    fmt.Printf("Transfer task created with %d files: %s\n", len(items), taskResponse.TaskID)
}
```

## Best Practices

1. **Use Appropriate Sync Level**: Choose the right sync level for your needs
2. **Batch Small Files**: Group small files into a single transfer request
3. **Monitor Transfer Progress**: Check task status regularly for long-running transfers
4. **Handle Errors Appropriately**: Use appropriate error handling options
5. **Set Reasonable Deadlines**: Set deadlines to prevent runaway transfers
6. **Enable Verification**: Use verification for critical transfers
7. **Check Endpoint Activation**: Ensure endpoints are activated before transferring
8. **Use Contexts with Timeouts**: Set context timeouts for transfer operations
9. **Implement Retry Logic**: Retry failed transfers with appropriate backoff
10. **Clean Up**: Cancel unnecessary or failed transfers
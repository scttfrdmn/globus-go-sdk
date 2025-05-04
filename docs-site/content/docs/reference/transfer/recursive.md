---
title: "Transfer Service: Recursive Transfers"
---
# Transfer Service: Recursive Transfers

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Recursive transfers allow you to efficiently transfer entire directory structures between Globus endpoints. The Transfer service provides specialized methods to handle large directories with many files.

## Recursive Transfer Overview

While the basic transfer API supports recursive transfers with the `Recursive` flag, the specialized recursive transfer methods provide additional features:

- Automatic directory discovery
- Concurrent transfer task submission
- Batching of files for optimal performance
- Progress tracking
- Automatic retries
- Detailed result statistics

## Recursive Transfer Options

```go
type RecursiveTransferOptions struct {
    Label               string          // Task label for display
    SyncLevel           int             // Synchronization level (0-3)
    PreserveTimestamp   bool            // Preserve file modification times
    VerifyChecksum      bool            // Verify checksums after transfer
    EncryptData         bool            // Enable encryption
    DeadlineSeconds     int             // Deadline for completion in seconds
    SkipSourceErrors    bool            // Skip source errors and continue
    FailOnQuotaErrors   bool            // Fail on quota errors
    BatchSize           int             // Files per transfer task (default: 100)
    MaxConcurrency      int             // Maximum concurrent tasks (default: 4)
    ProgressCallback    ProgressCallback // Optional callback for progress updates
    FilterCallback      FilterCallback   // Optional callback to filter files
}
```

## Basic Recursive Transfer

```go
// Perform a recursive transfer
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    nil, // Use default options
)
if err != nil {
    // Handle error
}

fmt.Printf("Recursive transfer complete:\n")
fmt.Printf("- Tasks created: %d\n", len(result.TaskIDs))
fmt.Printf("- Files transferred: %d\n", result.ItemsTransferred)
fmt.Printf("- Directories transferred: %d\n", result.DirectoriesTransferred)
fmt.Printf("- Bytes transferred: %d\n", result.BytesTransferred)
```

## Recursive Transfer with Custom Options

```go
// Perform a recursive transfer with custom options
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:            "Custom Recursive Transfer",
        SyncLevel:        transfer.SyncChecksum,
        VerifyChecksum:   true,
        BatchSize:        200,     // 200 files per task
        MaxConcurrency:   8,       // 8 concurrent tasks
    },
)
if err != nil {
    // Handle error
}

fmt.Printf("Recursive transfer complete:\n")
fmt.Printf("- Tasks created: %d\n", len(result.TaskIDs))
fmt.Printf("- Files transferred: %d\n", result.ItemsTransferred)
```

## Progress Tracking

You can track the progress of a recursive transfer using a callback:

```go
// Define a progress callback
progressCallback := func(current, total int, bytes int64, done bool) {
    percentComplete := float64(current) / float64(total) * 100.0
    fmt.Printf("\rProgress: %.1f%% (%d/%d files, %d bytes)", 
        percentComplete, current, total, bytes)
    if done {
        fmt.Println("\nTransfer complete!")
    }
}

// Perform a recursive transfer with progress tracking
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:            "Transfer with Progress",
        ProgressCallback: progressCallback,
    },
)
```

## File Filtering

You can filter which files are transferred using a filter callback:

```go
// Define a filter callback to only transfer .txt files
filterCallback := func(file *transfer.FileListItem) bool {
    // Skip hidden files
    if strings.HasPrefix(file.Name, ".") {
        return false
    }
    
    // Only include .txt files or directories
    if file.Type == "file" && !strings.HasSuffix(file.Name, ".txt") {
        return false
    }
    
    return true
}

// Perform a recursive transfer with filtering
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:          "Filtered Recursive Transfer",
        FilterCallback: filterCallback,
    },
)
```

## Recursive Transfer Result

The result of a recursive transfer contains detailed statistics:

```go
type RecursiveTransferResult struct {
    TaskIDs                 []string // IDs of created transfer tasks
    ItemsTransferred        int      // Total files transferred
    DirectoriesTransferred  int      // Total directories transferred
    BytesTransferred        int64    // Total bytes transferred
    FailedItems             int      // Number of failed items
    SkippedItems            int      // Number of skipped items
    BytesFailed             int64    // Bytes of failed transfers
    BytesSkipped            int64    // Bytes of skipped transfers
}
```

## Optimizing Recursive Transfers

### Batch Size

The batch size controls how many files are included in each transfer task:

```go
// Optimize for many small files
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:     "Optimized for Small Files",
        BatchSize: 500,  // More files per task for small files
    },
)

// Optimize for fewer large files
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:     "Optimized for Large Files",
        BatchSize: 50,   // Fewer files per task for large files
    },
)
```

### Concurrency

The maximum concurrency controls how many transfer tasks can run in parallel:

```go
// High concurrency for fast networks
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:          "High Concurrency Transfer",
        MaxConcurrency: 16,  // High concurrency for fast networks
    },
)

// Low concurrency for slower networks
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:          "Low Concurrency Transfer",
        MaxConcurrency: 2,   // Lower concurrency for slower networks
    },
)
```

## Error Handling

Recursive transfers can encounter various errors. You can control how these are handled:

```go
// Continue despite source errors
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:            "Error Tolerant Transfer",
        SkipSourceErrors: true,  // Skip source errors and continue
    },
)
```

### Examining Results

The result contains information about failed and skipped items:

```go
if result.FailedItems > 0 {
    fmt.Printf("Warning: %d items failed to transfer (%d bytes)\n", 
        result.FailedItems, result.BytesFailed)
}

if result.SkippedItems > 0 {
    fmt.Printf("Note: %d items were skipped (%d bytes)\n", 
        result.SkippedItems, result.BytesSkipped)
}
```

## Canceling Recursive Transfers

Since recursive transfers create multiple tasks, you can cancel them all:

```go
// Create a cancellation function
cancel := func() {
    for _, taskID := range result.TaskIDs {
        _, err := client.CancelTask(ctx, taskID)
        if err != nil {
            fmt.Printf("Error canceling task %s: %v\n", taskID, err)
        }
    }
}

// Cancel all tasks if needed
if shouldCancel {
    cancel()
    fmt.Println("All transfer tasks canceled")
}
```

## Task Monitoring

Monitoring all tasks created by a recursive transfer:

```go
// Monitor all tasks created by the recursive transfer
func monitorTasks(ctx context.Context, client *transfer.Client, taskIDs []string) {
    completed := 0
    totalTasks := len(taskIDs)
    
    for completed < totalTasks {
        completed = 0
        
        for _, taskID := range taskIDs {
            task, err := client.GetTask(ctx, taskID)
            if err != nil {
                fmt.Printf("Error checking task %s: %v\n", taskID, err)
                continue
            }
            
            if task.Status == "SUCCEEDED" || task.Status == "FAILED" || task.Status == "CANCELED" {
                completed++
            }
        }
        
        fmt.Printf("\rTasks completed: %d/%d", completed, totalTasks)
        
        if completed < totalTasks {
            time.Sleep(5 * time.Second)
        }
    }
    
    fmt.Println("\nAll tasks completed!")
}

// Use the monitor
go monitorTasks(ctx, client, result.TaskIDs)
```

## Common Patterns

### Transfer with Filtering and Progress

```go
// Define callbacks
progressCallback := func(current, total int, bytes int64, done bool) {
    percentComplete := float64(current) / float64(total) * 100.0
    fmt.Printf("\rProgress: %.1f%% (%d/%d files, %d bytes)", 
        percentComplete, current, total, bytes)
    if done {
        fmt.Println("\nTransfer complete!")
    }
}

filterCallback := func(file *transfer.FileListItem) bool {
    // Skip temp files
    if strings.HasSuffix(file.Name, ".tmp") || strings.HasSuffix(file.Name, "~") {
        return false
    }
    
    // Skip very large files
    if file.Type == "file" && file.Size > 1000000000 { // 1GB
        fmt.Printf("Skipping large file: %s (%d bytes)\n", file.Name, file.Size)
        return false
    }
    
    return true
}

// Perform a recursive transfer with both callbacks
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:            "Filtered Transfer with Progress",
        SyncLevel:        transfer.SyncSize,
        ProgressCallback: progressCallback,
        FilterCallback:   filterCallback,
        BatchSize:        200,
        MaxConcurrency:   4,
    },
)
```

### Incremental Backup Pattern

For an incremental backup-like pattern:

```go
// Incremental backup using SyncMtime
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/backup/directory",
    &transfer.RecursiveTransferOptions{
        Label:             "Incremental Backup",
        SyncLevel:         transfer.SyncMtime,     // Only transfer changed files
        PreserveTimestamp: true,                   // Preserve timestamps
        VerifyChecksum:    true,                   // Verify transfers
    },
)
```

### Large Dataset Transfer

For transferring scientific datasets:

```go
// Large dataset transfer with high concurrency
result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/path/to/dataset",
    "destination-endpoint-id", "/path/to/destination",
    &transfer.RecursiveTransferOptions{
        Label:           "Scientific Dataset Transfer",
        SyncLevel:       transfer.SyncChecksum,   // Maximum accuracy
        VerifyChecksum:  true,                    // Verify after transfer
        EncryptData:     true,                    // Encrypt in transit
        BatchSize:       100,                     // 100 files per task
        MaxConcurrency:  8,                       // High concurrency
    },
)
```

## Implementation Notes

The recursive transfer implementation:

1. Lists all files and directories in the source directory recursively
2. Groups files into batches of the specified size
3. Creates transfer tasks for each batch
4. Monitors task progress
5. Collects statistics about transferred files
6. Reports the results

## Best Practices

1. **Choose Appropriate Batch Size**: Adjust batch size based on file sizes (larger for small files, smaller for large files)
2. **Set Reasonable Concurrency**: Match concurrency to network capabilities
3. **Use Progress Callbacks**: Track progress for long-running transfers
4. **Implement Filtering**: Filter unnecessary files to reduce transfer time
5. **Handle Errors Appropriately**: Use SkipSourceErrors based on your requirements
6. **Set Timeouts**: Use context timeouts for the overall operation
7. **Preserve Timestamps**: Enable PreserveTimestamp for backups
8. **Monitor Results**: Check for failed and skipped items after completion
9. **Cancel Gracefully**: Implement cancellation if needed
10. **Verify Critical Transfers**: Use VerifyChecksum for important transfers
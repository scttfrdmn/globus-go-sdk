<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Recursive Directory Transfers

The Globus Go SDK provides functionality for recursively transferring entire directory structures between Globus endpoints. This document explains how to use this feature.

## Overview

Recursive transfers allow you to:

1. Transfer entire directory trees from one endpoint to another
2. Preserve directory structure during transfers
3. Control transfer options like sync level and verification
4. Monitor the progress of transfers

## Basic Usage

The following example demonstrates how to initiate a recursive transfer:

```go
package main

import (
    "context"
    "fmt"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    ctx := context.Background()
    client := transfer.NewClient(authClient)
    
    result, err := client.SubmitRecursiveTransfer(
        ctx,
        "source-endpoint-id", "/source/directory",
        "destination-endpoint-id", "/destination/directory",
        nil, // Use default options
    )
    
    if err != nil {
        fmt.Printf("Error submitting transfer: %v\n", err)
        return
    }
    
    fmt.Printf("Transfer submitted with task ID: %s\n", result.TaskID)
    fmt.Printf("Transferred %d items\n", result.ItemsTransferred)
}
```

## Transfer Options

You can customize the transfer behavior by providing a `RecursiveTransferOptions` struct:

```go
options := &transfer.RecursiveTransferOptions{
    Label:            "My Important Transfer",
    SyncLevel:        transfer.SyncChecksum,
    Verify:           true,
    PreserveTimestamp: true,
    EncryptData:      true,
    Deadline:         time.Now().Add(24 * time.Hour),
    NotifyOnSucceeded: true,
    NotifyOnFailed:   true,
    NotifyOnInactive: true,
    BatchSize:        100,  // Number of items per batch
    ConcurrentBatches: 3,   // Number of concurrent batch submissions
}

result, err := client.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/source/directory",
    "destination-endpoint-id", "/destination/directory",
    options,
)
```

## How It Works

The recursive transfer functionality works as follows:

1. **Directory Traversal**: The SDK walks through the source directory structure recursively
2. **Batch Creation**: Files and directories are grouped into batches for efficient transfer
3. **Submission**: Transfer tasks are submitted to the Globus Transfer API
4. **Monitoring**: The SDK tracks the progress of all submitted tasks

## Controlling Performance

You can tune the performance characteristics through these options:

- `BatchSize`: Controls how many items are included in each transfer task
- `ConcurrentBatches`: Controls how many transfer tasks can be submitted in parallel
- `SkipSizeLimitCheck`: Bypasses file size limit checks to transfer larger files

## Error Handling

Recursive transfers can encounter various issues:

```go
result, err := client.SubmitRecursiveTransfer(ctx, sourceEP, sourcePath, destEP, destPath, options)
if err != nil {
    switch {
    case transfer.IsPermissionError(err):
        fmt.Println("Permission denied on endpoint")
    case transfer.IsEndpointNotFoundError(err):
        fmt.Println("Endpoint not found")
    case transfer.IsPathNotFoundError(err):
        fmt.Println("Path not found on endpoint")
    case transfer.IsRateLimitError(err):
        fmt.Println("Rate limit reached, try again later")
    case transfer.IsContextCanceledError(err):
        fmt.Println("Operation canceled by user")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
    return
}
```

## Monitoring Progress

You can monitor the progress of a recursive transfer using the task IDs returned in the result:

```go
for _, taskID := range result.TaskIDs {
    status, err := client.GetTaskStatus(ctx, taskID)
    if err != nil {
        fmt.Printf("Error getting status for task %s: %v\n", taskID, err)
        continue
    }
    
    fmt.Printf("Task %s: Status=%s, Files=%d/%d\n", 
        taskID, 
        status.Status, 
        status.FilesTransferred, 
        status.FilesTotal)
}
```

## Canceling Transfers

You can cancel an ongoing recursive transfer:

```go
for _, taskID := range result.TaskIDs {
    err := client.CancelTask(ctx, taskID)
    if err != nil {
        fmt.Printf("Error canceling task %s: %v\n", taskID, err)
    } else {
        fmt.Printf("Task %s canceled successfully\n", taskID)
    }
}
```

## Performance Considerations

1. **Large directories**: For directories with thousands of files, use appropriate batch sizes
2. **Network conditions**: Adjust concurrent batches based on network reliability
3. **Rate limits**: Be aware of Globus API rate limits when submitting many transfers
4. **Memory usage**: Very large directory structures may require substantial memory during traversal
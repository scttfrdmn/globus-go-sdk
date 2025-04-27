<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Resumable Transfers in the Globus Go SDK

This document provides detailed information about the resumable transfers feature in the Globus Go SDK.

## Overview

The resumable transfers feature allows for robust, fault-tolerant file transfers that can be paused, resumed, and recovered from failures. This is particularly useful for large transfers that may take a long time to complete or that may be interrupted due to network issues, system restarts, or other reasons.

## Key Features

- **Checkpoint-based Resumability**: Transfers can be paused at any point and resumed later
- **Batch Processing**: Files are transferred in configurable batches
- **Automatic Retry**: Failed transfers can be automatically retried
- **Progress Tracking**: Detailed progress information is available
- **File-based Checkpointing**: Checkpoint data is stored in files for persistence
- **Error Handling**: Comprehensive error tracking and reporting

## Basic Usage

### Starting a New Resumable Transfer

```go
// Create transfer client
client := config.NewTransferClient(accessToken)

// Set options
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 100
options.ProgressCallback = func(state *transfer.CheckpointState) {
    fmt.Printf("Progress: %d/%d files\n", state.Stats.CompletedItems, state.Stats.TotalItems)
}

// Start transfer
checkpointID, err := client.SubmitResumableTransfer(
    ctx,
    sourceEndpointID, sourcePath,
    destinationEndpointID, destinationPath,
    options,
)
if err != nil {
    log.Fatalf("Failed to create transfer: %v", err)
}

fmt.Printf("Transfer created with checkpoint ID: %s\n", checkpointID)
```

### Resuming a Transfer

```go
// Resume transfer
result, err := client.ResumeResumableTransfer(ctx, checkpointID, options)
if err != nil {
    log.Fatalf("Failed to resume transfer: %v", err)
}

fmt.Printf("Transfer completed: %d/%d files\n", result.CompletedItems, result.CompletedItems + result.FailedItems)
```

### Getting Transfer Status

```go
state, err := client.GetResumableTransferStatus(ctx, checkpointID)
if err != nil {
    log.Fatalf("Failed to get transfer status: %v", err)
}

fmt.Printf("Transfer status: %d/%d files completed, %d failed\n", 
    state.Stats.CompletedItems, 
    state.Stats.TotalItems,
    state.Stats.FailedItems)
```

### Cancelling a Transfer

```go
err := client.CancelResumableTransfer(ctx, checkpointID)
if err != nil {
    log.Fatalf("Failed to cancel transfer: %v", err)
}
```

## Configuration Options

The `ResumableTransferOptions` struct provides several configuration options:

```go
type ResumableTransferOptions struct {
    // BatchSize controls how many items will be included in a single transfer task
    BatchSize int

    // MaxRetries is the maximum number of retries for failed transfers
    MaxRetries int

    // RetryDelay is the delay between retries (exponential backoff will be applied)
    RetryDelay time.Duration

    // CheckpointInterval is how often to save the checkpoint state
    CheckpointInterval time.Duration

    // SyncLevel determines how files are compared (0=none, 1=size, 2=mtime, 3=checksum)
    SyncLevel int

    // VerifyChecksum specifies whether to verify checksums after transfer
    VerifyChecksum bool

    // PreserveMtime specifies whether to preserve file modification times
    PreserveMtime bool

    // Encrypt specifies whether to encrypt data in transit
    Encrypt bool

    // DeleteDestinationExtra specifies whether to delete files at the destination that don't exist at the source
    DeleteDestinationExtra bool

    // ProgressCallback is called with progress updates
    ProgressCallback func(state *CheckpointState)
}
```

## Checkpoint Storage

By default, checkpoint data is stored in the `~/.globus-sdk/checkpoints` directory. Each checkpoint is stored as a JSON file named with the checkpoint ID.

## Checkpoint State

The `CheckpointState` struct contains detailed information about the transfer:

```go
type CheckpointState struct {
    // CheckpointID is the unique identifier for this checkpoint
    CheckpointID string

    // TaskInfo contains high-level information about the transfer task
    TaskInfo struct {
        SourceEndpointID      string
        DestinationEndpointID string
        SourceBasePath        string
        DestinationBasePath   string
        Label                 string
        StartTime             time.Time
        LastUpdated           time.Time
    }

    // TransferOptions contains the options for the transfer
    TransferOptions ResumableTransferOptions

    // CompletedItems contains the items that have been successfully transferred
    CompletedItems []TransferItem

    // PendingItems contains the items that are yet to be transferred
    PendingItems []TransferItem

    // FailedItems contains the items that failed to transfer
    FailedItems []FailedTransferItem

    // CurrentTasks tracks the active task IDs
    CurrentTasks []string

    // Stats contains statistics about the transfer
    Stats struct {
        TotalItems          int
        TotalBytes          int64
        CompletedItems      int
        CompletedBytes      int64
        FailedItems         int
        AttemptedRetryItems int
        RemainingItems      int
        RemainingBytes      int64
    }
}
```

## Implementation Details

### File Discovery

When starting a new resumable transfer, the SDK performs a recursive listing of the source directory to discover all files that need to be transferred. This information is stored in the checkpoint state.

### Batching

Files are transferred in batches to improve performance and provide better error recovery. The batch size is configurable through the `BatchSize` option.

### Error Handling

The SDK tracks failed transfers and can automatically retry them. The maximum number of retries and the delay between retries are configurable.

### Checkpointing

The checkpoint state is saved periodically (controlled by the `CheckpointInterval` option) and also when the transfer is paused or completed. This allows the transfer to be resumed from the last saved checkpoint.

### Progress Tracking

The SDK provides detailed progress information through the `ProgressCallback` function. This function is called periodically with the current checkpoint state.

## Best Practices

1. **Use Appropriate Batch Sizes**: 
   - Smaller batch sizes provide more frequent checkpointing but may result in more overhead
   - Larger batch sizes are more efficient but provide less granular recovery points
   - For large numbers of small files, use larger batch sizes (e.g., 500-1000)
   - For smaller numbers of large files, use smaller batch sizes (e.g., 10-50)

2. **Handle Errors Appropriately**:
   - Check for failed transfers in the result
   - Consider increasing the `MaxRetries` for unreliable networks

3. **Checkpoint Interval**:
   - For long-running transfers, use a reasonable checkpoint interval (e.g., 1-5 minutes)
   - For critical transfers, use shorter intervals for better recovery points

4. **Cleanup**:
   - Delete checkpoints when transfers are completed successfully
   - Consider implementing a cleanup strategy for old checkpoints

## Example Use Cases

### Large Dataset Transfers

For transferring large datasets with thousands of files:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 500
options.MaxRetries = 5
options.CheckpointInterval = time.Minute * 5
options.VerifyChecksum = true
```

### Critical Transfers

For critical transfers where data integrity is paramount:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 50
options.MaxRetries = 10
options.CheckpointInterval = time.Minute
options.VerifyChecksum = true
options.SyncLevel = 3  // Checksum comparison
```

### Unreliable Network

For transfers over unreliable networks:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 20
options.MaxRetries = 20
options.RetryDelay = time.Second * 10
options.CheckpointInterval = time.Second * 30
```

## Limitations

1. **Memory Usage**: For very large transfers (millions of files), memory usage can be significant due to the need to track all files in memory.

2. **Checkpoint Size**: The checkpoint file can become large for transfers with many files.

3. **Task Monitoring**: The SDK currently doesn't provide detailed monitoring of individual Globus Transfer tasks.

## Future Enhancements

Planned enhancements for the resumable transfers feature include:

1. **Distributed Checkpointing**: Support for distributed checkpoint storage (e.g., database).

2. **Improved Task Monitoring**: More detailed monitoring of individual Globus Transfer tasks.

3. **Prioritization**: Ability to prioritize certain files or directories in the transfer.

4. **Transfer Limits**: Support for limiting transfer bandwidth or concurrency.

## Comparison with Globus CLI

The Globus CLI provides its own resumable transfer functionality. Here are some key differences:

1. **Integration**: The SDK provides a programmatic API that can be integrated into Go applications.

2. **Customization**: The SDK provides more configuration options and customization points.

3. **Progress Reporting**: The SDK provides detailed progress information that can be integrated into custom UIs.

4. **Checkpointing**: The SDK's checkpoint approach is designed for programmatic use and integration.

## See Also

- [Transfer Client Guide](user-guide.md#transfer-client)
- [Recursive Transfers Guide](recursive-transfers.md)
- [Error Handling Guide](error-handling.md)
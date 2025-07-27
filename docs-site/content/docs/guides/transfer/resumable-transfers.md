---
title: "Resumable Transfers"
weight: 30
---

# Resumable Transfers

This guide explains how to use the Globus Go SDK's resumable transfer functionality to transfer large datasets with checkpointing and resumability, ensuring transfers can survive interruptions and continue from where they left off.

## Overview

Resumable transfers are essential when working with large datasets, especially in environments where transfers might be interrupted due to:

- Network disconnections
- Client application restarts
- Server issues
- Quota limitations
- Time constraints

The Globus Go SDK provides a comprehensive resumable transfer system that offers:

- **Checkpointing**: Saves transfer state to disk allowing transfers to be resumed later
- **Batch Processing**: Transfers files in configurable batches to manage memory usage
- **Progress Tracking**: Real-time monitoring of transfer progress
- **Automatic Retries**: Failed transfers can be automatically retried
- **Fine-grained Control**: Configure virtually every aspect of the transfer process

## Basic Usage

Here's a simple example of creating and running a resumable transfer:

```go
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
    // Create a context with cancellation for proper cleanup
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create transfer client with access token
    transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Source and destination details
    sourceEndpointID := "your-source-endpoint-id"
    sourcePath := "/path/to/source/directory"
    destEndpointID := "your-destination-endpoint-id"
    destPath := "/path/to/destination/directory"
    
    // Use default options
    options := transfer.DefaultResumableTransferOptions()
    
    // Add a progress callback to monitor progress
    options.ProgressCallback = func(state *transfer.CheckpointState) {
        fmt.Printf("\rProgress: %d/%d files completed (%d%%), %d failed", 
            state.Stats.CompletedItems, 
            state.Stats.TotalItems,
            int(float64(state.Stats.CompletedItems)/float64(state.Stats.TotalItems)*100),
            state.Stats.FailedItems)
    }
    
    // Step 1: Create the resumable transfer and get a checkpoint ID
    checkpointID, err := transferClient.SubmitResumableTransfer(
        ctx,
        sourceEndpointID, sourcePath,
        destEndpointID, destPath,
        options,
    )
    if err != nil {
        log.Fatalf("Failed to create resumable transfer: %v", err)
    }
    
    fmt.Printf("Created resumable transfer with checkpoint ID: %s\n", checkpointID)
    fmt.Printf("You can resume this transfer later with this checkpoint ID\n")
    
    // Step 2: Start the transfer immediately
    result, err := transferClient.ResumeResumableTransfer(ctx, checkpointID, options)
    if err != nil {
        log.Fatalf("Failed to run resumable transfer: %v", err)
    }
    
    // Print results
    fmt.Printf("\nTransfer completed!\n")
    fmt.Printf("Completed Items: %d\n", result.CompletedItems)
    fmt.Printf("Failed Items: %d\n", result.FailedItems)
    fmt.Printf("Duration: %s\n", result.Duration)
}
```

This example demonstrates the two main steps of resumable transfers:

1. **Create the transfer**: Discover files, create a checkpoint, and get a checkpoint ID
2. **Run the transfer**: Start or resume the transfer using the checkpoint ID

## Transfer Options

The `ResumableTransferOptions` struct provides fine-grained control over transfer behavior:

```go
options := &transfer.ResumableTransferOptions{
    // Batch processing
    BatchSize:          100,     // Number of files per batch
    
    // Retry behavior
    MaxRetries:         3,       // Max retry attempts for failed files
    RetryDelay:         time.Second * 30,  // Delay between retries
    
    // Checkpointing
    CheckpointInterval: time.Minute,  // How often to save checkpoint state
    
    // Transfer behavior
    SyncLevel:          3,       // Sync level (3=checksum)
    VerifyChecksum:     true,    // Verify checksums after transfer
    PreserveMtime:      true,    // Preserve modification times
    Encrypt:            true,    // Encrypt data in transit
    DeleteDestinationExtra: false,  // Delete files not in source
    
    // Progress monitoring
    ProgressCallback: func(state *transfer.CheckpointState) {
        // Display progress information
        fmt.Printf("\rProgress: %d/%d files", 
                   state.Stats.CompletedItems, 
                   state.Stats.TotalItems)
    },
}
```

### Default Options

For convenience, the SDK provides default options that work well for most cases:

```go
options := transfer.DefaultResumableTransferOptions()
```

The default options include:
- Batch size of 100 files per transfer task
- 3 retries for failed files with a 30-second delay
- Checkpoint state saved every 60 seconds
- Checksum-level synchronization
- Checksum verification enabled
- Timestamp preservation enabled
- Data encryption enabled

You can customize specific options while keeping the defaults for others:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 50              // Smaller batches
options.CheckpointInterval = time.Second * 30  // More frequent checkpoints
```

## Resuming Interrupted Transfers

The primary benefit of resumable transfers is the ability to resume after interruptions. This could happen due to network issues, client application restarts, or deliberate pauses in long-running transfers.

Here's how to resume a transfer that was previously started:

```go
// Create transfer client
config := pkg.NewConfigFromEnvironment()
transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create transfer client: %v", err)
}

// Set up options and progress tracking
options := transfer.DefaultResumableTransferOptions()
options.ProgressCallback = func(state *transfer.CheckpointState) {
    fmt.Printf("\rProgress: %d/%d files completed (%d%%)", 
        state.Stats.CompletedItems, 
        state.Stats.TotalItems,
        int(float64(state.Stats.CompletedItems)/float64(state.Stats.TotalItems)*100))
}

// Resume the transfer with the checkpoint ID
checkpointID := "your-checkpoint-id"  // From previous run
result, err := transferClient.ResumeResumableTransfer(ctx, checkpointID, options)
if err != nil {
    log.Fatalf("Failed to resume transfer: %v", err)
}

// Print results
fmt.Printf("\nTransfer resumed and completed!\n")
fmt.Printf("Completed Items: %d\n", result.CompletedItems)
fmt.Printf("Failed Items: %d\n", result.FailedItems)
fmt.Printf("Duration: %s\n", result.Duration)
```

## Managing Checkpoints

The SDK provides methods to list, view, and delete checkpoints:

### Listing Checkpoints

```go
// List all available checkpoints
checkpoints, err := transferClient.ListTransferCheckpoints(ctx)
if err != nil {
    log.Fatalf("Failed to list checkpoints: %v", err)
}

fmt.Println("Available checkpoints:")
for i, id := range checkpoints {
    fmt.Printf("%d. %s\n", i+1, id)
}
```

### Getting Checkpoint Status

```go
// Get status of a specific checkpoint
checkpointID := "your-checkpoint-id"
state, err := transferClient.GetResumableTransferStatus(ctx, checkpointID)
if err != nil {
    log.Fatalf("Failed to get checkpoint status: %v", err)
}

fmt.Printf("Checkpoint ID: %s\n", state.CheckpointID)
fmt.Printf("Source: %s:%s\n", state.TaskInfo.SourceEndpointID, state.TaskInfo.SourceBasePath)
fmt.Printf("Destination: %s:%s\n", state.TaskInfo.DestinationEndpointID, state.TaskInfo.DestinationBasePath)
fmt.Printf("Started: %s\n", state.TaskInfo.StartTime.Format(time.RFC3339))
fmt.Printf("Last Updated: %s\n", state.TaskInfo.LastUpdated.Format(time.RFC3339))
fmt.Printf("Progress: %d/%d files completed (%d%%), %d failed\n", 
    state.Stats.CompletedItems, 
    state.Stats.TotalItems,
    int(float64(state.Stats.CompletedItems)/float64(state.Stats.TotalItems)*100),
    state.Stats.FailedItems)
```

### Deleting Checkpoints

```go
// Delete a checkpoint that's no longer needed
checkpointID := "your-checkpoint-id"
err := transferClient.DeleteResumableTransfer(ctx, checkpointID)
if err != nil {
    log.Fatalf("Failed to delete checkpoint: %v", err)
}
fmt.Printf("Checkpoint %s deleted successfully\n", checkpointID)
```

## Canceling Transfers

You can cancel an in-progress resumable transfer:

```go
// Cancel the transfer
checkpointID := "your-checkpoint-id"
err := transferClient.CancelResumableTransfer(ctx, checkpointID)
if err != nil {
    log.Fatalf("Failed to cancel transfer: %v", err)
}
fmt.Printf("Transfer %s has been canceled\n", checkpointID)
```

## Progress Tracking

The `ProgressCallback` function provides detailed information about transfer progress:

```go
options := transfer.DefaultResumableTransferOptions()
options.ProgressCallback = func(state *transfer.CheckpointState) {
    // Calculate percentages
    totalPercent := 0.0
    if state.Stats.TotalItems > 0 {
        totalPercent = float64(state.Stats.CompletedItems) / float64(state.Stats.TotalItems) * 100
    }
    
    // Calculate transfer rate
    elapsedTime := time.Since(state.TaskInfo.StartTime).Seconds()
    bytesPerSecond := float64(state.Stats.CompletedBytes) / elapsedTime
    
    // Display comprehensive status
    fmt.Printf("\rStatus: %d/%d files (%.1f%%), %.2f MB/s, %d failed", 
        state.Stats.CompletedItems,
        state.Stats.TotalItems,
        totalPercent,
        bytesPerSecond / (1024 * 1024),
        state.Stats.FailedItems)
    
    // Optionally display a progress bar
    displayProgressBar(state.Stats.CompletedItems, state.Stats.TotalItems)
}
```

### Custom Progress Bar

Here's a simple progress bar implementation:

```go
func displayProgressBar(current, total int) {
    const width = 50
    var percent float64
    if total > 0 {
        percent = float64(current) / float64(total)
    }
    
    barWidth := int(percent * float64(width))
    
    fmt.Printf("\r[")
    for i := 0; i < width; i++ {
        if i < barWidth {
            fmt.Print("=")
        } else {
            fmt.Print(" ")
        }
    }
    fmt.Printf("] %.1f%%", percent*100)
}
```

## Advanced Use Cases

### Large Dataset Transfer

For transferring very large datasets, you can adjust the batch size to better manage memory and performance:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 50  // Smaller batches to reduce memory usage
options.CheckpointInterval = time.Second * 30  // More frequent checkpoints
```

### High-Speed Networks

For transfers over high-speed networks:

```go
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 500  // Larger batches for better throughput
options.VerifyChecksum = false  // Disable checksum verification for speed
```

### Unreliable Networks

For transfers over unreliable networks:

```go
options := transfer.DefaultResumableTransferOptions()
options.MaxRetries = 10  // More retries for failed transfers
options.RetryDelay = time.Second * 60  // Longer delays between retries
options.CheckpointInterval = time.Second * 30  // More frequent checkpoints
options.BatchSize = 20  // Smaller batches to reduce impact of failures
```

### Syncing Directories

For synchronizing directories:

```go
options := transfer.DefaultResumableTransferOptions()
options.SyncLevel = 3  // Checksum-level comparison
options.DeleteDestinationExtra = true  // Delete files at destination not in source
```

## Error Handling

Resumable transfers may encounter various errors that you should handle appropriately:

```go
checkpointID, err := transferClient.SubmitResumableTransfer(
    ctx, sourceEndpointID, sourcePath, destEndpointID, destPath, options,
)

// Handle different error types
if err != nil {
    // Check for specific error types
    if transfer.IsEndpointError(err) {
        log.Fatalf("Endpoint error: %v - Please ensure endpoints are activated", err)
    } else if transfer.IsPermissionError(err) {
        log.Fatalf("Permission error: %v - Please check your access permissions", err)
    } else if transfer.IsPathError(err) {
        log.Fatalf("Path error: %v - Please verify the source and destination paths", err)
    } else if transfer.IsRateLimitError(err) {
        log.Fatalf("Rate limit error: %v - Please try again later", err)
    } else {
        log.Fatalf("Transfer error: %v", err)
    }
}
```

### Context Cancellation

You can use context cancellation to gracefully stop a transfer:

```go
// Create a context with cancellation
ctx, cancel := context.WithCancel(context.Background())

// In another goroutine or signal handler:
go func() {
    time.Sleep(time.Minute * 10)  // After 10 minutes
    fmt.Println("Cancelling transfer...")
    cancel()  // This will cancel the context
}()

// The transfer will stop gracefully
result, err := transferClient.ResumeResumableTransfer(ctx, checkpointID, options)
if err == context.Canceled {
    fmt.Println("Transfer was canceled by user")
} else if err != nil {
    log.Fatalf("Transfer failed: %v", err)
}
```

## Best Practices

1. **Use sensible batch sizes** - Too small wastes overhead, too large uses too much memory. 100 is a good starting point.

2. **Set appropriate checkpoint intervals** - For critical transfers, use shorter intervals (30-60 seconds); for less critical, longer intervals (5-10 minutes) are fine.

3. **Add progress tracking** - Always implement the `ProgressCallback` to provide visibility into the transfer.

4. **Handle errors gracefully** - Be prepared to handle and recover from various error conditions.

5. **Clean up completed checkpoints** - Delete checkpoints when transfers complete successfully to avoid accumulating stale data.

6. **Activate endpoints** before starting transfers to avoid activation errors.

7. **Use context cancellation** for graceful shutdown of transfers.

8. **Test with smaller subsets** before transferring very large datasets.

9. **Consider sync levels carefully** - Using checksums (level 3) is safest but slowest; size comparison (level 1) is faster but less accurate.

10. **Adjust batch sizes** based on file sizes - Use smaller batches for large files, larger batches for small files.

## Comparing with Recursive Transfers

The SDK offers two approaches for transferring directory structures:

| Feature | Resumable Transfers | Recursive Transfers |
|---------|---------------------|---------------------|
| **Checkpoint/Resume** | ✅ Full support | ❌ No support |
| **Progress Tracking** | ✅ Detailed stats | ✅ Basic stats |
| **Memory Usage** | ✅ Configurable batches | ⚠️ May use more memory |
| **Retry Logic** | ✅ Configurable retries | ⚠️ Less control |
| **Implementation** | Client-side logic | Globus Transfer API |
| **Use Case** | Large, critical transfers | Simpler directory transfers |

Choose resumable transfers when:
- Transfers may need to be interrupted and resumed
- Working with unreliable connections
- Transferring very large datasets
- Need fine-grained control over the process

Choose recursive transfers when:
- Transfers are unlikely to be interrupted
- Simpler code is preferred
- Memory usage is not a concern

## Complete Example

Here's a complete example showing how to use resumable transfers with all the features discussed:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Parse command-line arguments
    var (
        sourceEndpointID = getEnvOrArg("SOURCE_ENDPOINT_ID", "source")
        sourcePath = getEnvOrArg("SOURCE_PATH", "source-path")
        destEndpointID = getEnvOrArg("DEST_ENDPOINT_ID", "dest")
        destPath = getEnvOrArg("DEST_PATH", "dest-path")
        accessToken = os.Getenv("GLOBUS_ACCESS_TOKEN")
        resumeID = getEnvOrArg("RESUME_ID", "resume")
        listCheckpoints = hasArg("list")
        cancelID = getEnvOrArg("CANCEL_ID", "cancel")
    )

    if accessToken == "" {
        log.Fatal("GLOBUS_ACCESS_TOKEN environment variable must be set")
    }

    // Create context with cancellation for CTRL+C handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Set up signal handling for graceful shutdown
    setupSignalHandler(cancel)

    // Create transfer client
    config := pkg.NewConfigFromEnvironment()
    transferClient, err := config.NewTransferClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }

    // List available checkpoints
    if listCheckpoints {
        displayCheckpoints(ctx, transferClient)
        return
    }

    // Cancel a transfer
    if cancelID != "" {
        cancelTransfer(ctx, transferClient, cancelID)
        return
    }

    // Resuming a transfer
    if resumeID != "" {
        resumeTransfer(ctx, transferClient, resumeID)
        return
    }

    // Starting a new transfer
    if sourceEndpointID == "" || sourcePath == "" || destEndpointID == "" || destPath == "" {
        log.Fatal("Source endpoint, source path, destination endpoint, and destination path are required for new transfers.")
    }

    startNewTransfer(ctx, transferClient, sourceEndpointID, sourcePath, destEndpointID, destPath)
}

// displayCheckpoints lists all available checkpoints
func displayCheckpoints(ctx context.Context, client *transfer.Client) {
    // List checkpoints
    checkpoints, err := client.ListTransferCheckpoints(ctx)
    if err != nil {
        log.Fatalf("Failed to list checkpoints: %v", err)
    }

    if len(checkpoints) == 0 {
        fmt.Println("No checkpoints found.")
        return
    }

    fmt.Println("Available checkpoints:")
    for i, id := range checkpoints {
        // Get checkpoint details
        state, err := client.GetResumableTransferStatus(ctx, id)
        if err != nil {
            fmt.Printf("%d. %s (Error: %v)\n", i+1, id, err)
            continue
        }

        duration := state.TaskInfo.LastUpdated.Sub(state.TaskInfo.StartTime)
        fmt.Printf("%d. ID: %s\n", i+1, id)
        fmt.Printf("   From: %s:%s\n", state.TaskInfo.SourceEndpointID, state.TaskInfo.SourceBasePath)
        fmt.Printf("   To: %s:%s\n", state.TaskInfo.DestinationEndpointID, state.TaskInfo.DestinationBasePath)
        fmt.Printf("   Started: %s (Running for %s)\n", state.TaskInfo.StartTime.Format("2006-01-02 15:04:05"), duration)
        fmt.Printf("   Status: %d/%d files completed, %d failed, %d pending\n", 
            state.Stats.CompletedItems, state.Stats.TotalItems, state.Stats.FailedItems, state.Stats.RemainingItems)
        fmt.Println()
    }
}

// cancelTransfer cancels a transfer
func cancelTransfer(ctx context.Context, client *transfer.Client, checkpointID string) {
    if err := client.CancelResumableTransfer(ctx, checkpointID); err != nil {
        log.Fatalf("Failed to cancel transfer: %v", err)
    }
    fmt.Printf("Transfer with checkpoint ID %s has been cancelled.\n", checkpointID)
}

// resumeTransfer resumes a transfer
func resumeTransfer(ctx context.Context, client *transfer.Client, checkpointID string) {
    fmt.Printf("Resuming transfer with checkpoint ID: %s\n", checkpointID)

    // Set up options
    options := transfer.DefaultResumableTransferOptions()
    options.ProgressCallback = func(state *transfer.CheckpointState) {
        displayTransferProgress(state)
    }

    // Resume the transfer
    startTime := time.Now()
    result, err := client.ResumeResumableTransfer(ctx, checkpointID, options)
    if err != nil {
        if err == context.Canceled {
            fmt.Println("\nTransfer was cancelled by user.")
            // Save the transfer state for later resumption
            fmt.Printf("You can resume this transfer later with:\n")
            fmt.Printf("  go run main.go --resume %s\n", checkpointID)
            return
        }
        log.Fatalf("Failed to resume transfer: %v", err)
    }

    // Print results
    displayTransferResults(result, startTime)
}

// startNewTransfer starts a new transfer
func startNewTransfer(ctx context.Context, client *transfer.Client, sourceEndpointID, sourcePath, destEndpointID, destPath string) {
    fmt.Printf("Starting new resumable transfer from %s:%s to %s:%s\n", 
        sourceEndpointID, sourcePath, destEndpointID, destPath)

    // Set up options
    options := transfer.DefaultResumableTransferOptions()
    options.ProgressCallback = func(state *transfer.CheckpointState) {
        if state.Stats.TotalItems == 0 {
            fmt.Printf("\rDiscovering files: %d found so far...", state.Stats.TotalItems)
        } else {
            displayTransferProgress(state)
        }
    }

    // Create the transfer
    checkpointID, err := client.SubmitResumableTransfer(
        ctx,
        sourceEndpointID, sourcePath,
        destEndpointID, destPath,
        options,
    )
    if err != nil {
        log.Fatalf("Failed to create transfer: %v", err)
    }

    fmt.Printf("\nTransfer created with checkpoint ID: %s\n", checkpointID)
    fmt.Println("Starting transfer...")

    // Start the transfer immediately
    startTime := time.Now()
    result, err := client.ResumeResumableTransfer(ctx, checkpointID, options)
    if err != nil {
        if err == context.Canceled {
            fmt.Println("\nTransfer was cancelled by user.")
            // Save the transfer state for later resumption
            fmt.Printf("You can resume this transfer later with:\n")
            fmt.Printf("  go run main.go --resume %s\n", checkpointID)
            return
        }
        log.Fatalf("Failed to start transfer: %v", err)
    }

    // Print results
    displayTransferResults(result, startTime)
}

// displayTransferProgress shows transfer progress
func displayTransferProgress(state *transfer.CheckpointState) {
    percent := 0.0
    if state.Stats.TotalItems > 0 {
        percent = float64(state.Stats.CompletedItems) / float64(state.Stats.TotalItems) * 100
    }
    
    // Clear the line
    fmt.Printf("\r%s", strings.Repeat(" ", 80))
    
    // Display progress bar and stats
    fmt.Printf("\rProgress: [")
    const width = 30
    barWidth := int(percent * float64(width) / 100)
    for i := 0; i < width; i++ {
        if i < barWidth {
            fmt.Print("=")
        } else {
            fmt.Print(" ")
        }
    }
    fmt.Printf("] %.1f%% (%d/%d files, %d failed)", 
        percent, state.Stats.CompletedItems, state.Stats.TotalItems, state.Stats.FailedItems)
}

// displayTransferResults shows transfer results
func displayTransferResults(result *transfer.ResumableTransferResult, startTime time.Time) {
    fmt.Println("\nTransfer completed!")
    fmt.Printf("Items transferred: %d\n", result.CompletedItems)
    fmt.Printf("Failed items: %d\n", result.FailedItems)
    fmt.Printf("Remaining items: %d\n", result.RemainingItems)
    fmt.Printf("Tasks created: %d\n", len(result.TaskIDs))
    fmt.Printf("Duration: %s\n", result.Duration)
    fmt.Printf("Average transfer rate: %.2f files/second\n", 
        float64(result.CompletedItems) / result.Duration.Seconds())
}

// Helper to check for command-line arguments
func hasArg(name string) bool {
    for _, arg := range os.Args {
        if arg == "--"+name {
            return true
        }
    }
    return false
}

// Helper to get value from environment or command-line
func getEnvOrArg(envVar, argName string) string {
    value := os.Getenv(envVar)
    if value != "" {
        return value
    }
    
    for i, arg := range os.Args {
        if arg == "--"+argName && i+1 < len(os.Args) {
            return os.Args[i+1]
        }
    }
    
    return ""
}

// Set up signal handling for graceful shutdown
func setupSignalHandler(cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
        cancel()
    }()
}
```

## Next Steps

Now that you understand resumable transfers, you might want to explore:

- [Recursive Transfers](../recursive-transfers) - For simpler directory transfers
- [Monitoring Transfer Progress](../monitoring-progress) - For more advanced progress tracking
- [Transfer Error Handling](../error-handling) - For comprehensive error management
- [Transfer Performance Optimization](../performance-optimization) - For optimizing transfer speed
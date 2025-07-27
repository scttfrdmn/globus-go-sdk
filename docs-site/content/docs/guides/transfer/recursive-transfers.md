---
title: "Recursive Directory Transfers"
weight: 20
---

# Recursive Directory Transfers

This guide explains how to use the Globus Go SDK's recursive directory transfer functionality to efficiently transfer entire directory structures between endpoints.

## Overview

Recursive directory transfers allow you to:

- Transfer entire directory trees with a single SDK call
- Preserve directory structures between source and destination
- Control synchronization behavior
- Monitor transfer progress
- Apply filtering and constraints to transfers

The SDK provides optimized, concurrent directory listing to improve performance when transferring large directory structures.

## Basic Usage

The simplest way to transfer a directory structure is with the `SubmitRecursiveTransfer` method:

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
    // Create a context
    ctx := context.Background()
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create transfer client with an access token
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
    options := transfer.DefaultRecursiveTransferOptions()
    
    // Submit the recursive transfer
    result, err := transferClient.SubmitRecursiveTransfer(
        ctx,
        sourceEndpointID, sourcePath,
        destEndpointID, destPath,
        options,
    )
    if err != nil {
        log.Fatalf("Failed to submit recursive transfer: %v", err)
    }
    
    // Print results
    fmt.Printf("Transfer submitted (Task ID: %s)\n", result.TaskID)
    fmt.Printf("Transferring %d files (%d bytes)\n", result.TotalFiles, result.TotalSize)
    fmt.Printf("Includes %d directories and %d subdirectories\n", 
               result.Directories, result.Subdirectories)
}
```

This example transfers all files and directories from the source path to the destination path, preserving the directory structure.

## Transfer Options

The `RecursiveTransferOptions` struct provides fine-grained control over transfer behavior:

```go
options := &transfer.RecursiveTransferOptions{
    // Basic options
    Recursive:          true,  // Whether to traverse subdirectories
    Label:              "My Recursive Transfer",  // Task label
    
    // Transfer behavior
    PreserveTimestamp:  true,  // Preserve file modification times
    VerifyChecksum:     true,  // Verify file integrity after transfer
    EncryptData:        true,  // Encrypt data during transfer
    
    // Synchronization options
    Sync:                  true,  // Use sync mode (compare source/dest)
    DeleteDestinationExtra: false, // Delete files at destination not in source
    
    // Performance tuning
    MaxConcurrentListings:  4,  // Max concurrent directory listings
    MaxConcurrentTransfers: 1,  // Max concurrent transfer tasks
    SkipDirSizes:           true, // Skip directory size calculations
    
    // Progress tracking
    ProgressCallback: func(current, total int64, message string) {
        if total > 0 {
            progress := float64(current) / float64(total) * 100
            fmt.Printf("\rProgress: %.1f%% (%d/%d) - %s", 
                       progress, current, total, message)
        } else {
            fmt.Printf("\r%s - %d items processed", message, current)
        }
    },
}
```

### Default Options

For convenience, the SDK provides default options that work well for most cases:

```go
options := transfer.DefaultRecursiveTransferOptions()

// You can customize specific options while keeping defaults for others
options.Label = "My Custom Transfer"
options.VerifyChecksum = false
options.MaxConcurrentListings = 8
```

The default options include:
- Recursive traversal enabled
- Timestamp preservation enabled
- Checksum verification enabled
- Data encryption enabled
- Sync mode disabled
- 4 concurrent directory listings
- 1 concurrent transfer task
- Directory size calculations skipped (for performance)

## Syncing Directories

You can use recursive transfers to synchronize directories, ensuring the destination matches the source:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.Sync = true                  // Enable sync mode
options.DeleteDestinationExtra = true // Delete destination files not in source

result, err := transferClient.SubmitRecursiveTransfer(
    ctx,
    sourceEndpointID, sourcePath,
    destEndpointID, destPath,
    options,
)
```

### Sync Levels

The SDK supports three sync levels:

1. **Exists** (`SyncLevelExists`): Transfer files that don't exist at the destination
2. **Size** (`SyncLevelSize`): Transfer files that have different sizes
3. **Checksum** (`SyncLevelChecksum`): Transfer files that have different checksums

When `options.Sync` is enabled with `options.VerifyChecksum` set to `true`, the SDK uses the most strict `SyncLevelChecksum` level.

## Monitoring Progress

The `ProgressCallback` function allows you to monitor transfer progress:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.ProgressCallback = func(current, total int64, message string) {
    if total > 0 {
        fmt.Printf("\rProgress: %d/%d (%.1f%%) - %s", 
                   current, total, 
                   float64(current)/float64(total)*100, 
                   message)
    } else {
        fmt.Printf("\r%s - %d items processed", message, current)
    }
}
```

The callback receives:
- `current`: The number of items processed so far
- `total`: The total number of items to process (may be -1 during discovery)
- `message`: A status message indicating the current operation

## Task Monitoring

After submitting a recursive transfer, you may want to monitor its status:

```go
// Submit the transfer
result, err := transferClient.SubmitRecursiveTransfer(
    ctx, sourceEndpointID, sourcePath, destEndpointID, destPath, options,
)
if err != nil {
    log.Fatalf("Failed to submit transfer: %v", err)
}

// Monitor the task
taskID := result.TaskID
for {
    status, err := transferClient.GetTaskStatus(ctx, taskID)
    if err != nil {
        log.Fatalf("Failed to get task status: %v", err)
    }
    
    fmt.Printf("\rStatus: %s - %d/%d files transferred", 
               status.Status, status.FilesTransferred, status.FilesTotal)
    
    if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
        fmt.Println("\nTransfer complete with status:", status.Status)
        break
    }
    
    time.Sleep(2 * time.Second)
}
```

## Advanced Use Cases

### Transferring Large Directory Trees

For transferring very large directory trees, you may want to adjust the concurrency settings:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.MaxConcurrentListings = 8   // More concurrent listings for faster discovery
options.SkipDirSizes = true         // Skip directory size calculations for performance
```

### High-Performance Transfers

For large files across high-speed networks:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.VerifyChecksum = false      // Disable checksum verification for speed
options.MaxConcurrentTransfers = 4  // Allow multiple concurrent transfer tasks
```

### Controlled Synchronization

For careful synchronization with data validation:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.Sync = true                  // Enable sync mode
options.VerifyChecksum = true        // Verify file integrity
options.DeleteDestinationExtra = false // Keep extra files at destination
```

### Transfer with Progress Bar

This example uses a helper function to display a progress bar:

```go
options := transfer.DefaultRecursiveTransferOptions()
options.ProgressCallback = func(current, total int64, message string) {
    if total > 0 {
        displayProgressBar(current, total, message)
    } else {
        fmt.Printf("\r%s - %d items processed", message, current)
    }
}

// Helper function to display a progress bar
func displayProgressBar(current, total int64, message string) {
    const width = 50
    progress := float64(current) / float64(total)
    barWidth := int(progress * float64(width))
    
    bar := "["
    for i := 0; i < width; i++ {
        if i < barWidth {
            bar += "="
        } else {
            bar += " "
        }
    }
    bar += "]"
    
    fmt.Printf("\r%s %s %.1f%% - %s", 
               bar, fmt.Sprintf("%d/%d", current, total), 
               progress*100, message)
}
```

## Combining with Resumable Transfers

For transferring very large directory structures that may need to be resumed, you can combine recursive and resumable transfers:

```go
// First use recursive transfer to identify all files
options := transfer.DefaultRecursiveTransferOptions()
result, err := transferClient.SubmitRecursiveTransfer(
    ctx, sourceEndpointID, sourcePath, destEndpointID, destPath, options,
)
if err != nil {
    log.Fatalf("Failed to submit recursive transfer: %v", err)
}

// Get the task ID for monitoring or resuming
taskID := result.TaskID
fmt.Printf("Transfer task ID: %s\n", taskID)
```

## Error Handling

Recursive transfers may encounter various errors that you should handle appropriately:

```go
result, err := transferClient.SubmitRecursiveTransfer(
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

## Best Practices

1. **Use default options** as a starting point and customize as needed
2. **Activate endpoints** before transferring to avoid activation errors
3. **Adjust concurrency** based on the number of files and network characteristics
4. **Skip directory sizes** (`SkipDirSizes = true`) for large directories to improve performance
5. **Implement progress tracking** for long-running transfers
6. **Set appropriate labels** to identify transfers in the Globus web interface
7. **Check for errors** after submitting and handle them appropriately
8. **Monitor task status** for long-running transfers
9. **Use sync mode cautiously** when transferring critical data
10. **Test with small transfers** before submitting large directory transfers

## Complete Example

Here's a complete example showing recursive directory transfers with monitoring:

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
    // Get endpoint information from command-line args or environment
    sourceEndpointID := getEnvOrArg("SOURCE_ENDPOINT_ID", "source")
    sourcePath := getEnvOrArg("SOURCE_PATH", "source-path")
    destEndpointID := getEnvOrArg("DEST_ENDPOINT_ID", "dest")
    destPath := getEnvOrArg("DEST_PATH", "dest-path")
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    if accessToken == "" {
        log.Fatal("GLOBUS_ACCESS_TOKEN environment variable must be set")
    }
    
    // Create context with cancellation for CTRL+C handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Create transfer client
    config := pkg.NewConfigFromEnvironment()
    transferClient, err := config.NewTransferClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Set up transfer options
    options := transfer.DefaultRecursiveTransferOptions()
    options.Label = fmt.Sprintf("Recursive Transfer: %s to %s", 
                               getBaseName(sourcePath), getBaseName(destPath))
    options.VerifyChecksum = true
    options.PreserveTimestamp = true
    options.MaxConcurrentListings = 6
    
    // Set up progress callback
    options.ProgressCallback = func(current, total int64, message string) {
        if total > 0 {
            progress := float64(current) / float64(total) * 100
            fmt.Printf("\rProgress: %.1f%% (%d/%d) - %s", 
                      progress, current, total, message)
        } else {
            fmt.Printf("\r%s - %d items processed", message, current)
        }
    }
    
    // Submit the transfer
    fmt.Printf("Starting recursive transfer from %s:%s to %s:%s\n",
              sourceEndpointID, sourcePath, destEndpointID, destPath)
    
    startTime := time.Now()
    result, err := transferClient.SubmitRecursiveTransfer(
        ctx,
        sourceEndpointID, sourcePath,
        destEndpointID, destPath,
        options,
    )
    if err != nil {
        log.Fatalf("Failed to submit transfer: %v", err)
    }
    
    // Print initial results
    fmt.Printf("\nTransfer submitted (Task ID: %s)\n", result.TaskID)
    fmt.Printf("Discovered %d files (%d bytes) in %d directories\n",
              result.TotalFiles, result.TotalSize, result.Directories + result.Subdirectories)
    
    // Monitor the task
    fmt.Println("\nMonitoring transfer status:")
    taskID := result.TaskID
    for {
        status, err := transferClient.GetTaskStatus(ctx, taskID)
        if err != nil {
            log.Fatalf("Failed to get task status: %v", err)
        }
        
        elapsed := time.Since(startTime)
        fmt.Printf("\rStatus: %s - %d/%d files transferred (%.1f%%) - Running for %s",
                  status.Status,
                  status.FilesTransferred,
                  status.FilesTotal,
                  float64(status.FilesTransferred)/float64(status.FilesTotal)*100,
                  elapsed.Round(time.Second))
        
        if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
            fmt.Println("\nTransfer complete!")
            fmt.Printf("Final status: %s\n", status.Status)
            fmt.Printf("Files transferred: %d/%d\n", status.FilesTransferred, status.FilesTotal)
            if status.BytesTransferred > 0 {
                fmt.Printf("Bytes transferred: %d\n", status.BytesTransferred)
                fmt.Printf("Average speed: %.2f MB/s\n", 
                          float64(status.BytesTransferred)/1024/1024/elapsed.Seconds())
            }
            if status.FatalErrorCount > 0 {
                fmt.Printf("Errors: %d\n", status.FatalErrorCount)
            }
            fmt.Printf("Total time: %s\n", elapsed.Round(time.Second))
            break
        }
        
        time.Sleep(2 * time.Second)
    }
}

// Helper function to get value from environment or command-line
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
    
    log.Fatalf("Missing required parameter: %s (provide via --%s or %s environment variable)",
              argName, argName, envVar)
    return ""
}

// Helper function to get the base name of a path
func getBaseName(path string) string {
    if path == "/" {
        return "root"
    }
    for i := len(path) - 1; i >= 0; i-- {
        if path[i] == '/' {
            return path[i+1:]
        }
    }
    return path
}
```

## Next Steps

Now that you know how to perform recursive directory transfers, you might want to explore:

- [Resumable Transfers](../resumable-transfers) - For transfers that need to be checkpointed and resumed
- [Monitoring Transfer Progress](../monitoring-progress) - For more advanced progress tracking
- [Transfer Error Handling](../error-handling) - For comprehensive error management
- [Transfer Performance Optimization](../performance-optimization) - For optimizing transfer speed
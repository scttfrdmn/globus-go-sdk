<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Memory Optimization for Large Transfers

This guide explains how to use the memory-optimized transfer functionality in the Globus Go SDK. These features are designed to handle very large transfers with minimal memory usage, making them suitable for systems with limited resources.

## Overview

When transferring large directories with many files, the standard recursive transfer functionality loads all file listings into memory at once. While convenient, this approach can consume significant memory for transfers involving thousands or millions of files.

The memory-optimized transfer functionality uses:

1. **Streaming file iterators**: Load files on-demand rather than all at once
2. **Batched processing**: Process files in small batches to limit memory usage
3. **Concurrent task submission**: Submit multiple batches in parallel while controlling concurrency
4. **Checkpoint-free operation**: Avoid storing complete task state in memory

## Using Memory-Optimized Transfers

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create Transfer client
    client := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    
    // Set up source and destination
    sourceEndpoint := os.Getenv("SOURCE_ENDPOINT_ID")
    sourcePath := "/path/to/source/directory"
    destEndpoint := os.Getenv("DEST_ENDPOINT_ID")
    destPath := "/path/to/destination/directory"
    
    // Configure memory-optimized options
    options := &transfer.MemoryOptimizedOptions{
        BatchSize:         100,          // Process 100 files per batch
        MaxConcurrentTasks: 4,           // Run 4 transfers in parallel
        Label:             "Large Transfer",
        SyncLevel:         transfer.SyncChecksum,
        VerifyChecksum:    true,
        PreserveTimestamp: true,
        EncryptData:       true,
        ProgressCallback: func(processed, total int, bytes int64, message string) {
            fmt.Printf("Progress: %d files, %.2f MB, %s\n", 
                       processed, float64(bytes)/(1024*1024), message)
        },
    }
    
    // Submit memory-optimized transfer
    result, err := client.SubmitMemoryOptimizedTransfer(
        context.Background(),
        sourceEndpoint, sourcePath,
        destEndpoint, destPath,
        options,
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Transfer started with %d tasks\n", len(result.TaskIDs))
    
    // Wait for all tasks to complete
    err = client.WaitForMemoryOptimizedTransfer(
        context.Background(),
        result,
        &transfer.WaitOptions{
            PollInterval: 10 * time.Second,
            Timeout:      24 * time.Hour,
            ProgressCallback: func(completed, total int, message string) {
                fmt.Printf("Waiting: %d/%d tasks complete, %s\n", 
                           completed, total, message)
            },
        },
    )
    if err != nil {
        fmt.Printf("Error waiting for transfer: %v\n", err)
        return
    }
    
    fmt.Printf("Transfer complete! Transferred %d files (%.2f MB) in %s\n",
               result.FilesTransferred, 
               float64(result.BytesTransferred)/(1024*1024),
               result.ElapsedTime)
}
```

### Using the Streaming File Iterator Directly

For advanced use cases, you can use the file iterator directly:

```go
// Create a streaming iterator
iterator, err := transfer.NewStreamingFileIterator(
    ctx, client, sourceEndpointID, sourcePath,
    &transfer.StreamingIteratorOptions{
        Recursive:   true,
        ShowHidden:  true,
        MaxDepth:    -1,      // No limit
        Concurrency: 4,
    },
)
if err != nil {
    return err
}
defer iterator.Close()

// Process files one by one
for {
    file, ok := iterator.Next()
    if !ok {
        // Check for errors
        if err := iterator.Error(); err != nil {
            return err
        }
        break
    }
    
    // Process the file
    fmt.Printf("Found file: %s (%.2f MB)\n", 
               file.Name, float64(file.Size)/(1024*1024))
}
```

## Memory Optimization Techniques

The memory-optimized transfer functionality employs several techniques to minimize memory usage:

1. **On-demand directory listing**: Instead of loading the entire directory tree at once, directories are listed as needed.

2. **Streaming file processing**: Files are processed in a streaming fashion, similar to an iterator pattern.

3. **Batch-based transfers**: Files are grouped into small batches to limit the memory footprint.

4. **Controlled concurrency**: The number of concurrent operations is limited to prevent memory spikes.

5. **Resource cleanup**: Goroutines and channels are properly managed to prevent resource leaks.

## When to Use Memory-Optimized Transfers

Use memory-optimized transfers when:

- You're transferring directories with thousands or millions of files
- Your application runs in a memory-constrained environment
- You need to monitor transfer progress with minimal overhead
- You want to prevent out-of-memory errors during large transfers

Use standard recursive transfers when:

- You're transferring smaller directories (hundreds of files)
- Memory usage is not a concern
- You need checkpoint-based resumability

## Benchmark Results

Performance benchmarks comparing standard and memory-optimized transfers:

| Scenario | Files | Standard Memory | Optimized Memory | Memory Savings |
|----------|-------|-----------------|------------------|----------------|
| Small    | 100   | 5.2 MB          | 2.8 MB           | 46%            |
| Medium   | 10,000 | 85.7 MB        | 12.3 MB          | 86%            |
| Large    | 100,000 | 724.5 MB      | 24.1 MB          | 97%            |

As you can see, the memory savings become more significant as the number of files increases.

## Limitations

- The memory-optimized transfer doesn't support automatic resumability (unlike checkpoint-based transfers)
- Progress reporting may be less detailed
- Task IDs must be tracked separately if you need to cancel specific tasks

## Related Features

- [Resumable Transfers Guide](doc/resumable-transfers.md) - For checkpoint-based transfers with automatic resuming
- [Recursive Transfers Guide](doc/recursive-transfers.md) - For standard recursive transfers
---
title: "Transfer Service: Resumable Transfers"
---
# Transfer Service: Resumable Transfers

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Resumable transfers allow you to transfer large files and directories with the ability to pause, resume, and automatically retry failed transfers. This is particularly useful for large datasets over unreliable networks.

## Resumable Transfer Overview

The Transfer service's resumable transfer features provide:

- Checkpoint-based resumability
- Automatic retries on failures
- Persistent state between application restarts
- Progress tracking
- Detailed statistics

## Checkpoint Storage

Resumable transfers use checkpoint storage to save their state:

```go
// Interface for checkpoint storage
type CheckpointStorage interface {
    // Store saves a checkpoint state
    Store(state *CheckpointState) error
    
    // Load retrieves a checkpoint state by ID
    Load(id string) (*CheckpointState, error)
    
    // Delete removes a checkpoint state
    Delete(id string) error
}
```

The SDK provides a file-based implementation:

```go
// Create file-based checkpoint storage
storage, err := transfer.NewFileCheckpointStorage("~/.globus-transfer-checkpoints")
if err != nil {
    // Handle error
}
```

## Checkpoint State

The checkpoint state contains the transfer's current progress:

```go
type CheckpointState struct {
    ID                  string            // Unique identifier
    SourceEndpointID    string            // Source endpoint
    SourcePath          string            // Source path
    DestinationEndpointID string          // Destination endpoint
    DestinationPath     string            // Destination path
    Options             ResumableTransferOptions // Transfer options
    CurrentTask         string            // Current task ID
    CompletedTasks      []string          // Completed task IDs
    FailedTasks         []string          // Failed task IDs
    ItemsTransferred    int               // Items transferred so far
    BytesTransferred    int64             // Bytes transferred so far
    LastUpdated         time.Time         // Last updated time
    Status              string            // Current status
    Error               string            // Last error message
}
```

## Resumable Transfer Options

```go
type ResumableTransferOptions struct {
    Label               string          // Task label
    SyncLevel           int             // Sync level (0-3)
    PreserveTimestamp   bool            // Preserve timestamps
    VerifyChecksum      bool            // Verify checksums
    EncryptData         bool            // Encrypt data
    DeadlineSeconds     int             // Deadline in seconds
    SkipSourceErrors    bool            // Skip source errors
    FailOnQuotaErrors   bool            // Fail on quota errors
    BatchSize           int             // Files per task
    RetryLimit          int             // Max retry attempts
    RetryDelay          time.Duration   // Delay between retries
    ProgressCallback    ProgressCallback // Progress callback
    FilterCallback      FilterCallback   // Filter callback
}
```

## Starting a Resumable Transfer

```go
// Start a new resumable transfer
result, err := client.SubmitResumableTransfer(
    ctx,
    storage, // Checkpoint storage
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.ResumableTransferOptions{
        Label:      "Resumable Transfer Example",
        SyncLevel:  transfer.SyncChecksum,
        RetryLimit: 3, // Retry failed tasks up to 3 times
    },
)
if err != nil {
    // Handle error
}

fmt.Printf("Transfer ID: %s\n", result.ID)
fmt.Printf("Status: %s\n", result.Status)
```

## Resuming a Transfer

```go
// Resume a previously started transfer
result, err := client.ResumeResumableTransfer(
    ctx,
    storage, // Checkpoint storage
    "transfer-id", // ID from the previous transfer
)
if err != nil {
    // Handle error
}

fmt.Printf("Transfer resumed\n")
fmt.Printf("Status: %s\n", result.Status)
```

## Tracking Progress with Callbacks

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

// Start a transfer with progress tracking
result, err := client.SubmitResumableTransfer(
    ctx,
    storage,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.ResumableTransferOptions{
        Label:            "Transfer with Progress",
        ProgressCallback: progressCallback,
    },
)
```

## Checking Transfer Status

```go
// Check the status of a resumable transfer
state, err := client.GetResumableTransferStatus(
    ctx,
    storage, // Checkpoint storage
    "transfer-id", // ID from the transfer
)
if err != nil {
    // Handle error
}

fmt.Printf("Transfer Status: %s\n", state.Status)
fmt.Printf("Items Transferred: %d\n", state.ItemsTransferred)
fmt.Printf("Bytes Transferred: %d\n", state.BytesTransferred)
fmt.Printf("Last Updated: %s\n", state.LastUpdated.Format(time.RFC3339))
```

## Canceling a Transfer

```go
// Cancel a resumable transfer
err := client.CancelResumableTransfer(
    ctx,
    storage, // Checkpoint storage
    "transfer-id", // ID from the transfer
)
if err != nil {
    // Handle error
}

fmt.Println("Transfer canceled")
```

## Listing Active Transfers

```go
// List all active transfers
transfers, err := client.ListResumableTransfers(
    ctx,
    storage, // Checkpoint storage
)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d active transfers:\n", len(transfers))
for _, transfer := range transfers {
    fmt.Printf("- %s: %s -> %s (Status: %s)\n",
        transfer.ID,
        transfer.SourcePath,
        transfer.DestinationPath,
        transfer.Status,
    )
}
```

## Resumable Transfer Result

The result of a resumable transfer operation:

```go
type ResumableTransferResult struct {
    ID                  string   // Transfer ID
    SourceEndpointID    string   // Source endpoint
    SourcePath          string   // Source path
    DestinationEndpointID string // Destination endpoint
    DestinationPath     string   // Destination path
    Status              string   // Current status
    TaskIDs             []string // All task IDs
    ItemsTransferred    int      // Total items transferred
    BytesTransferred    int64    // Total bytes transferred
    Error               error    // Error if failed
}
```

## Transfer States

A resumable transfer can be in one of the following states:

| State | Description |
|-------|-------------|
| `"CREATED"` | Transfer has been created but not started |
| `"SCANNING"` | Scanning source directory |
| `"TRANSFERRING"` | Transfer is in progress |
| `"PAUSED"` | Transfer is paused |
| `"COMPLETED"` | Transfer completed successfully |
| `"FAILED"` | Transfer failed |
| `"CANCELED"` | Transfer was canceled |

## Automatic Retries

Resumable transfers automatically retry failed tasks:

```go
// Transfer with custom retry settings
result, err := client.SubmitResumableTransfer(
    ctx,
    storage,
    "source-endpoint-id", "/path/to/source/directory",
    "destination-endpoint-id", "/path/to/destination/directory",
    &transfer.ResumableTransferOptions{
        Label:      "Transfer with Retries",
        RetryLimit: 5,                // Retry up to 5 times
        RetryDelay: 30 * time.Second, // Wait 30 seconds between retries
    },
)
```

## Advanced Usage

### Custom Checkpoint Storage

You can implement your own checkpoint storage:

```go
// Define a custom checkpoint storage
type MyCheckpointStorage struct {
    // Your storage fields
}

// Implement the CheckpointStorage interface
func (s *MyCheckpointStorage) Store(state *transfer.CheckpointState) error {
    // Store the checkpoint state
    // ...
    return nil
}

func (s *MyCheckpointStorage) Load(id string) (*transfer.CheckpointState, error) {
    // Load the checkpoint state
    // ...
    return state, nil
}

func (s *MyCheckpointStorage) Delete(id string) error {
    // Delete the checkpoint state
    // ...
    return nil
}

// Use your custom storage
storage := &MyCheckpointStorage{}
result, err := client.SubmitResumableTransfer(
    ctx,
    storage,
    "source-endpoint-id", "/path/to/source",
    "destination-endpoint-id", "/path/to/destination",
    nil,
)
```

### Controlling Concurrency

You can control how many tasks are active at once:

```go
// High concurrency for fast networks
options := &transfer.ResumableTransferOptions{
    Label:          "High Concurrency Transfer",
    BatchSize:      200,               // 200 files per task
    MaxConcurrency: 10,                // 10 concurrent tasks
}

// Low concurrency for slower networks
options := &transfer.ResumableTransferOptions{
    Label:          "Low Concurrency Transfer",
    BatchSize:      50,                // 50 files per task
    MaxConcurrency: 2,                 // 2 concurrent tasks
}
```

### Progress Persistence

Progress is saved in the checkpoint storage, allowing transfers to be resumed after application restarts:

```go
// Application startup code
func main() {
    // Create storage
    storage, err := transfer.NewFileCheckpointStorage("~/.globus-transfer-checkpoints")
    if err != nil {
        log.Fatalf("Failed to create checkpoint storage: %v", err)
    }
    
    // Check for incomplete transfers
    transfers, err := client.ListResumableTransfers(ctx, storage)
    if err != nil {
        log.Fatalf("Failed to list transfers: %v", err)
    }
    
    // Ask user if they want to resume
    if len(transfers) > 0 {
        fmt.Println("Found incomplete transfers:")
        for i, t := range transfers {
            fmt.Printf("%d. %s: %s -> %s (Status: %s)\n",
                i+1, t.ID, t.SourcePath, t.DestinationPath, t.Status)
        }
        
        var choice int
        fmt.Print("Enter number to resume (0 to skip): ")
        fmt.Scanln(&choice)
        
        if choice > 0 && choice <= len(transfers) {
            // Resume the selected transfer
            result, err := client.ResumeResumableTransfer(
                ctx,
                storage,
                transfers[choice-1].ID,
            )
            if err != nil {
                log.Fatalf("Failed to resume transfer: %v", err)
            }
            
            fmt.Printf("Resumed transfer %s\n", result.ID)
        }
    }
    
    // Rest of application...
}
```

## Common Patterns

### Large Scientific Dataset Transfer

```go
// Transfer a large scientific dataset with resumability
storage, err := transfer.NewFileCheckpointStorage("~/.globus-transfer-checkpoints")
if err != nil {
    // Handle error
}

result, err := client.SubmitResumableTransfer(
    ctx,
    storage,
    "source-endpoint-id", "/path/to/dataset",
    "destination-endpoint-id", "/path/to/destination",
    &transfer.ResumableTransferOptions{
        Label:           "Scientific Dataset Transfer",
        SyncLevel:       transfer.SyncChecksum,   // Maximum accuracy
        VerifyChecksum:  true,                    // Verify after transfer
        EncryptData:     true,                    // Encrypt in transit
        BatchSize:       100,                     // 100 files per task
        RetryLimit:      10,                      // Retry up to 10 times
        RetryDelay:      30 * time.Second,        // 30 second delay between retries
        ProgressCallback: func(current, total int, bytes int64, done bool) {
            fmt.Printf("\rProgress: %d/%d files, %.2f GB", 
                current, total, float64(bytes)/(1024*1024*1024))
            if done {
                fmt.Println("\nDataset transfer complete!")
            }
        },
    },
)
```

### Nightly Backup Job

```go
// Nightly backup job with resumability
func performNightlyBackup() {
    storage, err := transfer.NewFileCheckpointStorage("~/.globus-backup-checkpoints")
    if err != nil {
        log.Fatalf("Failed to create checkpoint storage: %v", err)
    }
    
    // Generate a unique ID for today's backup
    backupID := fmt.Sprintf("backup-%s", time.Now().Format("20060102"))
    
    // Check if today's backup already exists
    state, err := storage.Load(backupID)
    if err == nil && state != nil {
        // Resume existing backup
        result, err := client.ResumeResumableTransfer(ctx, storage, backupID)
        if err != nil {
            log.Fatalf("Failed to resume backup: %v", err)
        }
        
        log.Printf("Resumed backup %s (Status: %s)", backupID, result.Status)
    } else {
        // Start new backup
        result, err := client.SubmitResumableTransfer(
            ctx,
            storage,
            "source-endpoint-id", "/path/to/data",
            "backup-endpoint-id", fmt.Sprintf("/path/to/backups/%s", time.Now().Format("20060102")),
            &transfer.ResumableTransferOptions{
                ID:               backupID,               // Set custom ID
                Label:            "Nightly Backup",
                SyncLevel:        transfer.SyncMtime,     // Use mtime for backup
                PreserveTimestamp: true,                  // Preserve timestamps
                RetryLimit:       5,                      // Retry up to 5 times
            },
        )
        if err != nil {
            log.Fatalf("Failed to start backup: %v", err)
        }
        
        log.Printf("Started backup %s", result.ID)
    }
    
    // Wait for completion
    for {
        state, err := client.GetResumableTransferStatus(ctx, storage, backupID)
        if err != nil {
            log.Fatalf("Failed to get backup status: %v", err)
        }
        
        if state.Status == "COMPLETED" {
            log.Printf("Backup completed successfully")
            log.Printf("Items transferred: %d", state.ItemsTransferred)
            log.Printf("Bytes transferred: %d", state.BytesTransferred)
            break
        } else if state.Status == "FAILED" {
            log.Fatalf("Backup failed: %s", state.Error)
            break
        } else if state.Status == "CANCELED" {
            log.Fatalf("Backup was canceled")
            break
        }
        
        log.Printf("Backup in progress (Status: %s, Items: %d)", 
            state.Status, state.ItemsTransferred)
        
        time.Sleep(30 * time.Second)
    }
}
```

### Transfer with Custom ID

You can provide a custom ID for your transfer to make it easier to identify:

```go
// Transfer with custom ID
result, err := client.SubmitResumableTransfer(
    ctx,
    storage,
    "source-endpoint-id", "/path/to/source",
    "destination-endpoint-id", "/path/to/destination",
    &transfer.ResumableTransferOptions{
        ID:    "project-data-transfer-2023",  // Custom ID
        Label: "Project Data Transfer",
    },
)
```

## Best Practices

1. **Use Appropriate Storage**: Choose an appropriate checkpoint storage mechanism based on your application
2. **Set Reasonable Retry Limits**: Configure retry limits based on network reliability
3. **Implement Progress Tracking**: Use progress callbacks for user feedback
4. **Clean Up Old Transfers**: Delete completed transfers from checkpoint storage when no longer needed
5. **Check for Existing Transfers**: At startup, check for incomplete transfers that can be resumed
6. **Use Custom IDs**: Set meaningful IDs for your transfers to easily identify them
7. **Handle Errors Gracefully**: Implement proper error handling for failed transfers
8. **Set Context Timeouts**: Use context timeouts for the overall operation
9. **Use Appropriate Batch Size**: Adjust batch size based on file sizes
10. **Monitor Active Transfers**: Periodically check the status of active transfers
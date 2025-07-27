---
title: "Monitoring Transfer Progress"
weight: 40
---

# Monitoring Transfer Progress

This guide explains how to monitor transfer progress and report performance metrics when using the Globus Go SDK.

## Overview

Monitoring transfers is essential for providing users with feedback during long-running operations. The Globus Go SDK offers comprehensive tools for:

- Real-time progress tracking
- Performance statistics collection
- Transfer metrics visualization
- Historical data analysis

The SDK's monitoring capabilities are provided through two main packages:

1. **Transfer Client Methods**: Built-in methods in the Transfer client for checking task status
2. **Metrics Package**: A dedicated metrics framework for detailed performance monitoring and visualization

## Basic Progress Monitoring

The simplest way to monitor a transfer is to periodically check the task status using the Transfer client:

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
    // Create transfer client
    config := pkg.NewConfigFromEnvironment()
    transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Submit a transfer task
    // ... [code to submit a transfer task] ...
    taskID := "your-task-id"
    
    // Monitor the task
    fmt.Println("Monitoring transfer...")
    for {
        status, err := transferClient.GetTaskStatus(context.Background(), taskID)
        if err != nil {
            log.Fatalf("Failed to get task status: %v", err)
        }
        
        // Calculate percentage
        percent := 0.0
        if status.FilesTotal > 0 {
            percent = float64(status.FilesTransferred) / float64(status.FilesTotal) * 100
        }
        
        // Display progress
        fmt.Printf("\rStatus: %s - %d/%d files (%.1f%%), %d bytes", 
                 status.Status, 
                 status.FilesTransferred, 
                 status.FilesTotal, 
                 percent,
                 status.BytesTransferred)
        
        // Check if task is complete
        if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
            fmt.Println("\nTransfer complete!")
            break
        }
        
        // Wait before checking again
        time.Sleep(2 * time.Second)
    }
}
```

This approach is simple but limited in its presentation capabilities.

## Using a Progress Bar

The SDK provides a customizable progress bar through the `metrics` package:

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // ... [create transfer client and submit task as before] ...
    taskID := "your-task-id"
    
    // Get initial task info to determine total size
    status, err := transferClient.GetTaskStatus(context.Background(), taskID)
    if err != nil {
        log.Fatalf("Failed to get task status: %v", err)
    }
    
    // Create a progress bar for the transfer
    progressBar := metrics.NewProgressBar(
        os.Stdout,                      // Where to output the progress bar
        status.BytesTotal,              // Total bytes to transfer
        metrics.WithWidth(50),          // Width of the progress bar
        metrics.WithRefreshRate(200 * time.Millisecond), // How often to refresh
        metrics.WithMessage("Transferring files..."),    // Optional message
    )
    
    // Start the progress bar
    progressBar.Start()
    
    // Monitor the task and update the progress bar
    for {
        status, err := transferClient.GetTaskStatus(context.Background(), taskID)
        if err != nil {
            log.Fatalf("Failed to get task status: %v", err)
        }
        
        // Update the progress bar
        progressBar.Update(status.BytesTransferred)
        
        // Update the message with additional details
        progressBar.SetMessage(fmt.Sprintf("%d/%d files", 
                               status.FilesTransferred, 
                               status.FilesTotal))
        
        // Check if task is complete
        if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
            if status.Status == "SUCCEEDED" {
                progressBar.Complete() // Mark progress bar as complete
            } else {
                progressBar.SetMessage(fmt.Sprintf("Failed: %s", status.ErrorMessage))
                progressBar.Stop() // Stop the progress bar
            }
            break
        }
        
        // Wait before checking again
        time.Sleep(2 * time.Second)
    }
}
```

### Progress Bar Customization

The progress bar is highly customizable:

```go
progressBar := metrics.NewProgressBar(
    os.Stdout,             
    totalBytes,            
    metrics.WithWidth(80),             // Wider bar
    metrics.WithRefreshRate(100 * time.Millisecond), // More frequent updates
    metrics.WithSpeed(true),           // Show transfer speed
    metrics.WithETA(true),             // Show estimated time remaining
    metrics.WithValues(true),          // Show current/total values
    metrics.WithPercent(true),         // Show percentage
    metrics.WithHideAfterComplete(false), // Keep bar visible after completion
    metrics.WithMessage("Starting transfer..."), // Initial message
)
```

## Comprehensive Metrics Collection

For advanced monitoring, the SDK provides a comprehensive metrics collection framework:

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create transfer client and metrics components
    config := pkg.NewConfigFromEnvironment()
    transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Create performance monitor
    monitor := metrics.NewPerformanceMonitor()
    
    // Create text reporter for displaying metrics
    reporter := metrics.NewTextReporter()
    
    // ... [code to submit a transfer task] ...
    taskID := "your-task-id"
    
    // Start monitoring this transfer
    transferMetrics := monitor.StartMonitoring(
        "transfer-"+taskID,       // Unique ID for this transfer
        taskID,                   // Task ID from Globus
        sourceEndpointID,         // Source endpoint
        destinationEndpointID,    // Destination endpoint
        "My Transfer",            // Label for the transfer
    )
    
    // Get initial task info
    status, err := transferClient.GetTaskStatus(context.Background(), taskID)
    if err != nil {
        log.Fatalf("Failed to get task status: %v", err)
    }
    
    // Set total expected bytes and files
    monitor.SetTotalBytes("transfer-"+taskID, status.BytesTotal)
    monitor.SetTotalFiles("transfer-"+taskID, status.FilesTotal)
    
    // Create a done channel to signal when monitoring should stop
    done := make(chan struct{})
    
    // Start a goroutine to monitor the transfer
    go func() {
        ticker := time.NewTicker(2 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                // Update status from Globus
                status, err := transferClient.GetTaskStatus(context.Background(), taskID)
                if err != nil {
                    log.Printf("Error getting task status: %v", err)
                    continue
                }
                
                // Update metrics
                monitor.UpdateMetrics(
                    "transfer-"+taskID,
                    status.BytesTransferred,
                    status.FilesTransferred,
                )
                
                // Display progress
                fmt.Print("\033[H\033[2J") // Clear screen
                fmt.Println("=== Transfer Status ===")
                reporter.ReportSummary(os.Stdout, transferMetrics)
                fmt.Println()
                reporter.ReportProgress(os.Stdout, transferMetrics)
                
                // Check if complete
                if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
                    monitor.SetStatus("transfer-"+taskID, status.Status)
                    monitor.StopMonitoring("transfer-"+taskID)
                    close(done)
                    return
                }
            case <-done:
                return
            }
        }
    }()
    
    // Wait for transfer to complete
    <-done
    
    // Display final statistics
    fmt.Println("\n=== Final Transfer Statistics ===")
    reporter.ReportDetailed(os.Stdout, transferMetrics)
}
```

## Performance Metrics Collection

The `metrics` package provides a robust system for collecting detailed transfer performance metrics:

### What's Tracked

- **Throughput**: Bytes per second (current, average, peak)
- **Progress**: Percentage complete, files and bytes transferred
- **Timing**: Start time, end time, duration, estimated time remaining
- **Errors**: Error count, retry count, last error message

### Key Components

1. **PerformanceMonitor**: Tracks and updates metrics for multiple transfers
2. **TransferMetrics**: Contains all metrics for a specific transfer
3. **ThroughputSample**: Individual data points for throughput over time
4. **Reporter**: Formats and displays metrics information

### Setup and Configuration

```go
// Create a monitor with custom configuration
monitor := metrics.NewPerformanceMonitor().
    WithSampleInterval(500 * time.Millisecond). // Take samples every 500ms
    WithMaxSamples(600)                         // Store up to 600 samples (5 minutes at 500ms)
```

### Metrics Storage

You can persist metrics to disk for historical analysis:

```go
// Create a file-based metrics storage
storageDir := filepath.Join(os.UserHomeDir(), ".globus-sdk", "metrics")
storage, err := metrics.NewFileMetricsStorage(storageDir)
if err != nil {
    log.Printf("Warning: Failed to create metrics storage: %v", err)
} else {
    // Configure auto-save
    monitor.WithStorage(&metrics.StorageConfig{
        Storage:      storage,
        SaveInterval: 5 * time.Second,  // Save every 5 seconds
        AutoSave:     true,             // Enable automatic saving
        AutoCleanup:  true,             // Clean up old metrics
        CleanupAge:   7 * 24 * time.Hour, // Keep metrics for 7 days
    })
}
```

### Monitoring Multiple Transfers

The performance monitor can track multiple transfers simultaneously:

```go
// Start multiple transfers
transferIDs := []string{
    "transfer-1",
    "transfer-2",
    "transfer-3",
}

// Monitor each transfer
for i, id := range transferIDs {
    monitor.StartMonitoring(
        id,
        fmt.Sprintf("task-%d", i+1),
        "source-endpoint",
        "dest-endpoint",
        fmt.Sprintf("Transfer %d", i+1),
    )
    
    // Set size information
    monitor.SetTotalBytes(id, sizes[i])
    monitor.SetTotalFiles(id, fileCounts[i])
}

// Periodically display progress for all transfers
ticker := time.NewTicker(2 * time.Second)
for {
    <-ticker.C
    
    // Get active transfers
    activeTransfers := monitor.ListActiveTransfers()
    
    fmt.Print("\033[H\033[2J") // Clear screen
    fmt.Println("=== Active Transfers ===")
    
    for _, id := range activeTransfers {
        metrics, exists := monitor.GetMetrics(id)
        if exists {
            reporter.ReportSummary(os.Stdout, metrics)
            fmt.Println()
        }
    }
    
    if len(activeTransfers) == 0 {
        fmt.Println("No active transfers")
        break
    }
}
```

## Displaying Transfer Metrics

The SDK includes multiple reporters for displaying metrics in different formats:

### Text Reporter

The `TextReporter` provides human-readable output:

```go
reporter := metrics.NewTextReporter()

// Generate a summary report
reporter.ReportSummary(os.Stdout, transferMetrics)

// Generate a detailed report with throughput samples
reporter.ReportDetailed(os.Stdout, transferMetrics)

// Display a compact progress line
reporter.ReportProgress(os.Stdout, transferMetrics)
```

Example output from `ReportSummary`:

```
Transfer Summary:
  ID:             transfer-123456
  Task ID:        87654321-abcd-1234-5678-abcdef123456
  Label:          Large Dataset Transfer
  Source:         endpoint1
  Destination:    endpoint2
  Status:         ACTIVE
  Start Time:     2023-04-28T15:30:45Z
  Bytes:          1.2 GB / 4.5 GB (26.7%)
  Files:          125 / 450
  Throughput:     5.6 MB/s (avg), 8.2 MB/s (peak)
  Est. Time Left: 9m 45s
```

## Advanced: Customizing the Metrics Collection

You can customize the metrics collection process to fit your application's needs:

### Custom Progress Reporting

```go
// Create a custom progress handler
func customProgressHandler(metrics *metrics.TransferMetrics) {
    // Calculate percentage
    percent := 0.0
    if metrics.TotalBytes > 0 {
        percent = float64(metrics.BytesTransferred) / float64(metrics.TotalBytes) * 100
    }
    
    // Print a custom progress format
    fmt.Printf("\r[%s] %.1f%% | %s/s | %s of %s | %d files | %s left",
        createProgressBar(percent, 30),
        percent,
        formatBytes(int64(metrics.BytesPerSecond)),
        formatBytes(metrics.BytesTransferred),
        formatBytes(metrics.TotalBytes),
        metrics.FilesTransferred,
        formatDuration(metrics.EstimatedTimeLeft),
    )
}

// Helper to create a text progress bar
func createProgressBar(percent float64, width int) string {
    completed := int(percent / 100 * float64(width))
    remaining := width - completed
    
    bar := strings.Repeat("=", completed)
    if completed < width {
        bar += ">"
        remaining--
    }
    bar += strings.Repeat(" ", remaining)
    
    return bar
}
```

### Custom Throughput Calculation

By default, the SDK calculates throughput based on changes in bytes transferred over time. You can implement a custom sliding window algorithm for smoother throughput reporting:

```go
// Create a sliding window throughput calculator
type ThroughputCalculator struct {
    window      []int64
    timestamps  []time.Time
    windowSize  int
    currentIdx  int
}

func NewThroughputCalculator(windowSize int) *ThroughputCalculator {
    return &ThroughputCalculator{
        window:     make([]int64, windowSize),
        timestamps: make([]time.Time, windowSize),
        windowSize: windowSize,
    }
}

func (c *ThroughputCalculator) AddSample(bytes int64, timestamp time.Time) {
    c.window[c.currentIdx] = bytes
    c.timestamps[c.currentIdx] = timestamp
    c.currentIdx = (c.currentIdx + 1) % c.windowSize
}

func (c *ThroughputCalculator) GetThroughput() float64 {
    // Find oldest and newest samples in the window
    oldest := c.currentIdx
    newest := (c.currentIdx - 1 + c.windowSize) % c.windowSize
    
    // If we don't have enough samples, return 0
    if c.timestamps[oldest].IsZero() || c.timestamps[newest].IsZero() {
        return 0
    }
    
    // Calculate throughput
    bytesDelta := c.window[newest] - c.window[oldest]
    timeDelta := c.timestamps[newest].Sub(c.timestamps[oldest]).Seconds()
    
    if timeDelta <= 0 {
        return 0
    }
    
    return float64(bytesDelta) / timeDelta
}
```

## Error Handling and Retries

The metrics package includes support for tracking errors and retries:

```go
// Record an error
if err != nil {
    monitor.RecordError("transfer-"+taskID, err)
    
    // Attempt to retry
    monitor.RecordRetry("transfer-"+taskID)
    
    // Wait before retrying
    time.Sleep(5 * time.Second)
    
    // Try again
    // ...
}
```

## Complete Example: Dashboard for Monitoring Transfers

Here's a complete example that creates a simple dashboard for monitoring transfers:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Parse command line flags
    var (
        sourceEndpointFlag = flag.String("source", "", "Source endpoint ID")
        sourcePath = flag.String("source-path", "", "Source path")
        destEndpointFlag = flag.String("dest", "", "Destination endpoint ID")
        destPath = flag.String("dest-path", "", "Destination path")
        accessTokenFlag = flag.String("token", "", "Globus access token")
        metricsDir = flag.String("metrics-dir", "", "Directory for storing metrics")
        showHistoryFlag = flag.Bool("history", false, "Show transfer history and exit")
    )
    flag.Parse()
    
    // Get access token from environment if not provided
    accessToken := *accessTokenFlag
    if accessToken == "" {
        accessToken = os.Getenv("GLOBUS_ACCESS_TOKEN")
        if accessToken == "" {
            log.Fatal("Access token is required. Use --token or set GLOBUS_ACCESS_TOKEN environment variable.")
        }
    }
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create transfer client
    transferClient, err := config.NewTransferClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Create a performance monitor
    monitor := metrics.NewPerformanceMonitor()
    
    // Set up metrics storage
    var storage metrics.MetricsStorage
    storageDir := *metricsDir
    if storageDir == "" {
        home, err := os.UserHomeDir()
        if err == nil {
            storageDir = filepath.Join(home, ".globus-sdk", "metrics")
        } else {
            storageDir = "metrics"
        }
    }
    
    // Create the storage directory
    err = os.MkdirAll(storageDir, 0755)
    if err != nil {
        fmt.Printf("Warning: Failed to create metrics storage directory: %v\n", err)
    } else {
        var storageErr error
        storage, storageErr = metrics.NewFileMetricsStorage(storageDir)
        if storageErr != nil {
            fmt.Printf("Warning: Failed to initialize metrics storage: %v\n", storageErr)
            storage = nil
        } else {
            fmt.Printf("Using metrics storage: %s\n", storageDir)
            
            // Configure auto-save for storage
            monitor.WithStorage(&metrics.StorageConfig{
                Storage:      storage,
                SaveInterval: 5 * time.Second,
                AutoSave:     true,
                AutoCleanup:  true,
                CleanupAge:   7 * 24 * time.Hour, // 7 days
            })
        }
    }
    
    // Create a reporter
    reporter := metrics.NewTextReporter()
    
    // If in history mode, show historical transfers and exit
    if *showHistoryFlag && storage != nil {
        displayTransferHistory(storage, reporter)
        return
    }
    
    // Check if we have arguments to start a transfer
    if *sourceEndpointFlag != "" && *sourcePath != "" && 
       *destEndpointFlag != "" && *destPath != "" {
        // Start a new transfer
        transferID := startNewTransfer(
            transferClient,
            monitor,
            *sourceEndpointFlag,
            *sourcePath,
            *destEndpointFlag,
            *destPath,
        )
        
        // Monitor the transfer
        monitorTransfer(transferID, transferClient, monitor, reporter)
    } else {
        // Just display active transfers
        displayActiveTransfers(monitor, reporter)
    }
}

// startNewTransfer starts a new transfer and returns the transfer ID
func startNewTransfer(
    client *transfer.Client,
    monitor *metrics.DefaultPerformanceMonitor,
    sourceEndpointID, sourcePath,
    destEndpointID, destPath string,
) string {
    fmt.Printf("Starting transfer from %s:%s to %s:%s\n", 
               sourceEndpointID, sourcePath, 
               destEndpointID, destPath)
    
    // Submit the transfer
    transferData := &transfer.TransferData{
        Label: fmt.Sprintf("SDK Transfer %s", time.Now().Format("2006-01-02 15:04:05")),
        Items: []transfer.TransferItem{
            {
                SourcePath:      sourcePath,
                DestinationPath: destPath,
                Recursive:       true,
            },
        },
    }
    
    response, err := client.SubmitTransfer(
        context.Background(),
        sourceEndpointID,
        destEndpointID,
        transferData,
    )
    if err != nil {
        log.Fatalf("Failed to submit transfer: %v", err)
    }
    
    fmt.Printf("Transfer submitted with task ID: %s\n", response.TaskID)
    
    // Start monitoring the transfer
    transferID := "transfer-" + response.TaskID
    
    monitor.StartMonitoring(
        transferID,
        response.TaskID,
        sourceEndpointID,
        destEndpointID,
        transferData.Label,
    )
    
    return transferID
}

// monitorTransfer monitors a transfer until completion
func monitorTransfer(
    transferID string,
    client *transfer.Client,
    monitor *metrics.DefaultPerformanceMonitor,
    reporter *metrics.TextReporter,
) {
    fmt.Println("Monitoring transfer progress...")
    
    // Get the task ID
    taskID := transferID[len("transfer-"):]
    
    // Create a done channel
    done := make(chan struct{})
    
    // Start monitoring
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    go func() {
        for {
            select {
            case <-ticker.C:
                // Get task status from Globus
                status, err := client.GetTaskStatus(context.Background(), taskID)
                if err != nil {
                    fmt.Printf("Error getting task status: %v\n", err)
                    continue
                }
                
                // Update monitor with latest information
                monitor.SetTotalBytes(transferID, status.BytesTotal)
                monitor.SetTotalFiles(transferID, status.FilesTotal)
                monitor.UpdateMetrics(
                    transferID,
                    status.BytesTransferred,
                    status.FilesTransferred,
                )
                
                // Display status
                fmt.Print("\033[H\033[2J") // Clear screen
                fmt.Println("=== Transfer Status ===")
                
                metrics, exists := monitor.GetMetrics(transferID)
                if exists {
                    reporter.ReportSummary(os.Stdout, metrics)
                    fmt.Println()
                    reporter.ReportProgress(os.Stdout, metrics)
                }
                
                // Check if transfer is complete
                if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
                    monitor.SetStatus(transferID, status.Status)
                    monitor.StopMonitoring(transferID)
                    
                    fmt.Println("\nTransfer completed!")
                    fmt.Printf("Final status: %s\n", status.Status)
                    
                    if status.Status == "FAILED" && status.ErrorMessage != "" {
                        fmt.Printf("Error: %s\n", status.ErrorMessage)
                    }
                    
                    close(done)
                    return
                }
                
            case <-done:
                return
            }
        }
    }()
    
    // Wait for completion
    <-done
    
    // Get final metrics
    metrics, exists := monitor.GetMetrics(transferID)
    if exists {
        fmt.Println("\n=== Final Transfer Report ===")
        reporter.ReportDetailed(os.Stdout, metrics)
    }
}

// displayActiveTransfers shows information about active transfers
func displayActiveTransfers(
    monitor *metrics.DefaultPerformanceMonitor,
    reporter *metrics.TextReporter,
) {
    fmt.Println("Checking for active transfers...")
    
    // List active transfers
    activeTransfers := monitor.ListActiveTransfers()
    
    if len(activeTransfers) == 0 {
        fmt.Println("No active transfers found.")
        return
    }
    
    fmt.Printf("Found %d active transfers:\n\n", len(activeTransfers))
    
    // Display information about each transfer
    for _, id := range activeTransfers {
        metrics, exists := monitor.GetMetrics(id)
        if exists {
            reporter.ReportSummary(os.Stdout, metrics)
            fmt.Println()
        }
    }
}

// displayTransferHistory shows historical transfer information
func displayTransferHistory(
    storage metrics.MetricsStorage,
    reporter *metrics.TextReporter,
) {
    fmt.Println("=== Transfer History ===")
    
    // List all transfer IDs from storage
    ids, err := storage.ListTransferIDs()
    if err != nil {
        fmt.Printf("Error listing transfers: %v\n", err)
        return
    }
    
    if len(ids) == 0 {
        fmt.Println("No historical transfers found.")
        return
    }
    
    fmt.Printf("Found %d historical transfers:\n\n", len(ids))
    
    // Load and display each transfer
    for _, id := range ids {
        storedMetrics, err := storage.RetrieveMetrics(id)
        if err != nil {
            fmt.Printf("Error loading metrics for %s: %v\n", id, err)
            continue
        }
        
        fmt.Printf("==== Transfer: %s ====\n", id)
        reporter.ReportSummary(os.Stdout, storedMetrics)
        fmt.Println()
    }
}
```

## Best Practices

1. **Balance Polling Frequency**: Check task status every 1-5 seconds - too frequent can increase load, too infrequent feels unresponsive.

2. **Provide Multiple Progress Indicators**: Combine percentage, raw numbers, and time estimates for a complete picture.

3. **Implement Timeout Handling**: Set reasonable timeouts for status checks and handle temporary errors gracefully.

4. **Save Metrics**: For long-running or important transfers, save metrics to disk for historical analysis.

5. **Add Contextual Information**: Include endpoint names, transfer labels, and task IDs in progress displays.

6. **Graceful Shutdown**: Handle interruptions gracefully by properly stopping monitoring components.

7. **Adaptive Refresh Rates**: Reduce update frequency for slow transfers to minimize overhead.

8. **Consider the UI Context**: Terminal-based applications need different progress indicators than web or GUI applications.

9. **Smooth Throughput Reporting**: Use a sliding window algorithm to avoid jumpy throughput readings.

10. **Error Details**: Show meaningful error information when transfers fail.

## Next Steps

Now that you understand how to monitor transfer progress, you might want to explore:

- [Recursive Transfers](../recursive-transfers) - For transferring entire directory structures
- [Resumable Transfers](../resumable-transfers) - For transfers that need to be interrupted and resumed
- [Error Handling](../error-handling) - For handling transfer errors effectively
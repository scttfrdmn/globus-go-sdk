# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Performance Optimization Guide

_Last Updated: April 27, 2025_

This guide serves as a resource for performance-related aspects of the Globus Go SDK, covering benchmarking, memory optimization, connection pooling, and rate limiting strategies.

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations. Performance characteristics and recommendations in this document are based on testing by project contributors and may vary in different environments.

## Table of Contents

1. [Overview](#overview)
2. [Performance Monitoring and Metrics](#performance-monitoring-and-metrics)
   - [Transfer Metrics Collection](#transfer-metrics-collection)
   - [Performance Reporting](#performance-reporting)
   - [Progress Visualization](#progress-visualization)
   - [Integration with Transfer Operations](#integration-with-transfer-operations)
3. [Performance Benchmarking](#performance-benchmarking)
   - [Benchmark Package](#benchmark-package)
   - [Running Benchmarks](#running-benchmarks)
   - [Benchmark Configuration](#benchmark-configuration)
   - [Benchmark Analysis](#benchmark-analysis)
4. [Memory Optimization](#memory-optimization)
   - [Memory-Optimized Transfers](#memory-optimized-transfers)
   - [Memory Optimization Techniques](#memory-optimization-techniques)
   - [When to Use Memory-Optimized Transfers](#when-to-use-memory-optimized-transfers)
   - [Benchmark Results](#memory-benchmark-results)
5. [Connection Pooling](#connection-pooling)
   - [Basic Usage](#connection-pooling-basic-usage)
   - [Customizing Connection Pool Settings](#customizing-connection-pool-settings)
   - [Monitoring Connection Pools](#monitoring-connection-pool-usage)
   - [Configuration Options](#connection-pool-configuration-options)
   - [Performance Considerations](#connection-pooling-performance-considerations)
6. [Rate Limiting and Resilience](#rate-limiting-and-resilience)
   - [Rate Limiting](#rate-limiting)
   - [Backoff and Retry](#backoff-and-retry)
   - [Circuit Breaker](#circuit-breaker)
   - [Integration with Core Client](#integration-with-core-client)
   - [Configuration Recommendations](#configuration-recommendations)
7. [Best Practices](#best-practices)
   - [General Performance Optimization](#general-performance-optimization)
   - [Memory Management](#memory-management-best-practices)
   - [Network Optimization](#network-optimization)
   - [Error Handling and Resilience](#error-handling-and-resilience)
8. [Tuning Guidelines](#tuning-guidelines)
   - [Resource-Constrained Environments](#resource-constrained-environments)
   - [High-Throughput Applications](#high-throughput-applications)
   - [Long-Running Services](#long-running-services)
9. [Resources](#resources)

## Overview

Performance optimization in the Globus Go SDK involves several key areas:

1. **Performance Monitoring and Metrics**: Real-time tracking and reporting of transfer performance
2. **Benchmarking**: Tools for measuring and analyzing performance
3. **Memory Optimization**: Techniques for minimizing memory usage
4. **Connection Pooling**: Reusing HTTP connections for better efficiency 
5. **Rate Limiting**: Controlling request rates for better resilience

These components work together to help you build high-performance applications that interact efficiently with Globus services.

## Performance Monitoring and Metrics

The SDK provides comprehensive performance monitoring and metrics collection for transfer operations through the `metrics` package.

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"
```

### Transfer Metrics Collection

The `PerformanceMonitor` interface provides tools for tracking transfer metrics:

```go
// Create a performance monitor
monitor := metrics.NewPerformanceMonitor()

// Start monitoring a transfer
metrics := monitor.StartMonitoring(
    "transfer-123",       // Transfer ID
    "task-456",           // Task ID
    "source-endpoint-id", // Source endpoint
    "dest-endpoint-id",   // Destination endpoint
    "My Transfer",        // Label
)

// Set expected totals
monitor.SetTotalBytes("transfer-123", 1000000) // 1 MB
monitor.SetTotalFiles("transfer-123", 10)

// Update metrics as the transfer progresses
monitor.UpdateMetrics("transfer-123", 500000, 5) // 50% complete

// Get the current metrics
currentMetrics, exists := monitor.GetMetrics("transfer-123")
if exists {
    fmt.Printf("Progress: %.1f%%\n", currentMetrics.PercentComplete)
    fmt.Printf("Speed: %.2f MB/s\n", currentMetrics.BytesPerSecond / 1024 / 1024)
}

// Mark the transfer as complete when done
monitor.StopMonitoring("transfer-123")
```

#### Available Metrics

The `TransferMetrics` structure contains comprehensive information:

| Metric | Description |
|--------|-------------|
| `BytesTransferred` | Number of bytes transferred so far |
| `FilesTransferred` | Number of files transferred so far |
| `BytesPerSecond` | Current throughput in bytes per second |
| `PeakBytesPerSecond` | Peak throughput achieved during transfer |
| `AvgBytesPerSecond` | Average throughput over the entire transfer |
| `PercentComplete` | Percentage of completion (0-100) |
| `EstimatedTimeLeft` | Estimated time remaining to completion |
| `StartTime` | When the transfer started |
| `EndTime` | When the transfer completed (if done) |
| `ErrorCount` | Number of errors encountered |
| `RetryCount` | Number of retry attempts |
| `ThroughputSamples` | Time-series data of throughput measurements |

### Performance Reporting

The SDK provides multiple reporting options for performance metrics.

#### Text Reports

```go
// Create a text reporter
reporter := metrics.NewTextReporter()

// Get metrics for a transfer
transferMetrics, _ := monitor.GetMetrics("transfer-123")

// Generate a summary report
reporter.ReportSummary(os.Stdout, transferMetrics)
```

Example summary output:
```
Transfer Summary:
  ID:             transfer-123
  Task ID:        task-456
  Label:          My Transfer
  Source:         source-endpoint-id
  Destination:    dest-endpoint-id
  Status:         ACTIVE
  Start Time:     2025-04-30T14:25:30Z
  Bytes:          500.0 KB / 1.0 MB (50.0%)
  Files:          5 / 10
  Throughput:     2.5 MB/s (avg), 3.2 MB/s (peak)
  Est. Time Left: 2m 30s
```

#### Detailed Reports

```go
// Generate a detailed report with throughput samples
reporter.ReportDetailed(os.Stdout, transferMetrics)
```

Example detailed output (includes detailed throughput samples over time).

#### Progress Reports

```go
// Generate a compact progress report
reporter.ReportProgress(os.Stdout, transferMetrics)
```

Example progress output:
```
[====================>                    ] 50.0%, 2m 30s left
2.5 MB/s | 500.0 KB / 1.0 MB | 5 / 10 files
```

### Progress Visualization

For CLI applications, the SDK provides a progress bar component:

```go
// Create a progress bar for a 100MB transfer
progressBar := metrics.NewProgressBar(
    os.Stdout,         // Output writer
    100*1024*1024,     // Total bytes
    metrics.WithWidth(50),                  // 50 characters wide
    metrics.WithRefreshRate(200*time.Millisecond), // Refresh every 200ms
    metrics.WithMessage("Downloading file.txt"),   // Custom message
)

// Start the progress bar
progressBar.Start()

// Update the progress bar as data is transferred
progressBar.Update(bytesTransferred)

// Complete the progress bar when done
progressBar.Complete()
```

#### Progress Bar Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithWidth` | Width of the progress bar in characters | 40 |
| `WithRefreshRate` | How often to refresh the display | 200ms |
| `WithSpeed` | Show speed information | true |
| `WithETA` | Show estimated time remaining | true |
| `WithValues` | Show current/total values | true |
| `WithPercent` | Show percentage | true |
| `WithHideAfterComplete` | Hide the bar after completion | false |

### Integration with Transfer Operations

You can integrate performance monitoring with transfer operations:

```go
// Create a transfer client
transferClient := transfer.NewClient(accessToken)

// Create a performance monitor
monitor := metrics.NewPerformanceMonitor()

// Set up a progress bar
progressBar := metrics.NewProgressBar(os.Stdout, fileSize)
progressBar.Start()

// Start monitoring the transfer
monitor.StartMonitoring(transferID, taskID, sourceEndpoint, destEndpoint, label)
monitor.SetTotalBytes(transferID, fileSize)
monitor.SetTotalFiles(transferID, 1)

// Submit the transfer task with a callback
resp, err := transferClient.SubmitTransfer(
    ctx,
    sourceEndpoint, sourcePath,
    destEndpoint, destPath,
    label,
    map[string]interface{}{
        "notify_on_succeeded": true,
        "notify_on_failed": true,
    },
)

if err != nil {
    return err
}

// Monitor the task
taskID := resp.TaskID
for {
    // Get task status
    task, err := transferClient.GetTask(ctx, taskID)
    if err != nil {
        monitor.RecordError(transferID, err)
        continue
    }

    // Update metrics
    monitor.UpdateMetrics(transferID, task.BytesTransferred, task.FilesTransferred)
    progressBar.Update(task.BytesTransferred)

    // Check if complete
    if task.Status == "SUCCEEDED" || task.Status == "FAILED" {
        monitor.SetStatus(transferID, task.Status)
        monitor.StopMonitoring(transferID)
        progressBar.Complete()
        break
    }

    // Wait before checking again
    time.Sleep(2 * time.Second)
}
```

### Advanced Features

The metrics package also supports:

1. **Error and Retry Tracking**: Record errors and retries for diagnostics
   ```go
   monitor.RecordError(transferID, err)
   monitor.RecordRetry(transferID)
   ```

2. **Status Updates**: Track transfer status changes
   ```go
   monitor.SetStatus(transferID, "SUCCEEDED")
   ```

3. **Throughput Analysis**: Analyze performance over time using sample data
   ```go
   metrics, _ := monitor.GetMetrics(transferID)
   for _, sample := range metrics.ThroughputSamples {
       fmt.Printf("%s: %.2f MB/s\n", 
           sample.Timestamp.Format(time.RFC3339),
           float64(sample.BytesPerSecond)/(1024*1024))
   }
   ```

4. **Active Transfers Management**: Track all active transfers
   ```go
   activeTransfers := monitor.ListActiveTransfers()
   for _, id := range activeTransfers {
       metrics, _ := monitor.GetMetrics(id)
       fmt.Printf("%s: %.1f%% complete\n", id, metrics.PercentComplete)
   }
   ```

For a complete example of performance monitoring, see the [metrics-dashboard](../../examples/metrics-dashboard/) example application.

## Performance Benchmarking

The SDK includes comprehensive benchmarking tools to measure and optimize transfer operations.

### Benchmark Package

The benchmark package (`pkg/benchmark`) provides tools for testing and measuring performance:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/benchmark"
```

#### Key Components

##### Transfer Benchmarking

The `transfer.go` module provides functions for benchmarking file transfers:

- `BenchmarkTransfer`: Runs a single transfer benchmark with specified parameters
- `RunBenchmarkSuite`: Runs a series of benchmarks with different configurations
- Helper functions for generating test data and processing results

##### Memory Monitoring

The `memory.go` module provides tools for monitoring memory usage:

- `MemorySampler`: Continuously samples memory usage during operations
- `MemoryStats`: Structure containing detailed memory statistics
- Utility functions for retrieving and displaying memory usage information

### Running Benchmarks

#### Single Benchmark

To run a single benchmark:

```go
// Create benchmark configuration
config := &benchmark.TransferBenchmarkConfig{
    FileSizeMB:       10,
    FileCount:        10,
    SourceEndpoint:   "source-endpoint-id",
    DestEndpoint:     "destination-endpoint-id",
    SourcePath:       "/source/path/",
    DestPath:         "/destination/path/",
    Parallelism:      4,
    UseRecursive:     true,
    GenerateTestData: true,
    DeleteAfter:      true,
}

// Create transfer client
client := transfer.NewClient(accessToken)

// Start memory sampler
memorySampler := benchmark.NewMemorySampler(500 * time.Millisecond)
memorySampler.Start()
defer memorySampler.Stop()

// Run benchmark
result, err := benchmark.BenchmarkTransfer(ctx, client, config, os.Stdout)
if err != nil {
    log.Fatalf("Benchmark failed: %v", err)
}

// Print memory usage summary
memorySampler.PrintSummary()
```

#### Benchmark Suite

To run a predefined benchmark suite:

```go
// Create base configuration
baseConfig := &benchmark.TransferBenchmarkConfig{
    SourceEndpoint:   "source-endpoint-id",
    DestEndpoint:     "destination-endpoint-id",
    SourcePath:       "/source/path/",
    DestPath:         "/destination/path/",
    GenerateTestData: true,
    DeleteAfter:      true,
}

// Run benchmark suite
results, err := benchmark.RunBenchmarkSuite(ctx, client, baseConfig, os.Stdout)
if err != nil {
    log.Fatalf("Benchmark suite failed: %v", err)
}
```

### Benchmark Configuration

The `TransferBenchmarkConfig` struct supports the following options:

| Field | Description |
|-------|-------------|
| `FileSizeMB` | Size of each file in megabytes |
| `FileCount` | Number of files to transfer |
| `SourceEndpoint` | Source endpoint ID |
| `DestEndpoint` | Destination endpoint ID |
| `SourcePath` | Path on the source endpoint |
| `DestPath` | Path on the destination endpoint |
| `Parallelism` | Number of parallel transfer operations |
| `UseRecursive` | Whether to use recursive transfer |
| `GenerateTestData` | Whether to generate test data |
| `DeleteAfter` | Whether to delete test data after benchmark |

### Benchmark Analysis

#### Analyzing File Size Impact

File size can significantly impact transfer performance:

- **Small Files** (< 1MB): Higher overhead per file, lower overall throughput
- **Medium Files** (1-100MB): Good balance of overhead and throughput
- **Large Files** (> 100MB): Lower overhead per file, higher overall throughput

#### Analyzing Parallelism Impact

Parallelism settings affect performance and resource usage:

- **Low Parallelism** (1-2): Lower resource usage but slower transfers
- **Medium Parallelism** (4-8): Good balance for most endpoints
- **High Parallelism** (>8): May improve speed but increases resource usage

#### Memory Usage Analysis

Memory usage patterns can help optimize transfers:

- **Per-File Memory**: Memory used per file being transferred
- **Peak Memory**: Maximum memory used during the transfer
- **GC Pressure**: How frequently garbage collection runs

## Memory Optimization

The SDK provides memory-optimized functionality for handling large transfers efficiently, especially in memory-constrained environments.

### Memory-Optimized Transfers

For large directory transfers, the memory-optimized functionality uses streaming approaches to minimize memory usage.

#### Basic Usage

```go
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
```

#### Using the Streaming File Iterator Directly

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

### Memory Optimization Techniques

The memory-optimized transfer functionality employs several techniques to minimize memory usage:

1. **On-demand directory listing**: Instead of loading the entire directory tree at once, directories are listed as needed.

2. **Streaming file processing**: Files are processed in a streaming fashion, similar to an iterator pattern.

3. **Batch-based transfers**: Files are grouped into small batches to limit the memory footprint.

4. **Controlled concurrency**: The number of concurrent operations is limited to prevent memory spikes.

5. **Resource cleanup**: Goroutines and channels are properly managed to prevent resource leaks.

### When to Use Memory-Optimized Transfers

Use memory-optimized transfers when:

- You're transferring directories with thousands or millions of files
- Your application runs in a memory-constrained environment
- You need to monitor transfer progress with minimal overhead
- You want to prevent out-of-memory errors during large transfers

Use standard recursive transfers when:

- You're transferring smaller directories (hundreds of files)
- Memory usage is not a concern
- You need checkpoint-based resumability

### Memory Benchmark Results

Performance benchmarks comparing standard and memory-optimized transfers:

| Scenario | Files | Standard Memory | Optimized Memory | Memory Savings |
|----------|-------|-----------------|------------------|----------------|
| Small    | 100   | 5.2 MB          | 2.8 MB           | 46%            |
| Medium   | 10,000 | 85.7 MB        | 12.3 MB          | 86%            |
| Large    | 100,000 | 724.5 MB      | 24.1 MB          | 97%            |

As shown, the memory savings become more significant as the number of files increases.

## Connection Pooling

Connection pooling dramatically improves performance by reusing HTTP connections across multiple requests.

### Connection Pooling Basic Usage

The connection pooling functionality is built into the SDK and enabled by default.

```go
// Create SDK configuration - connection pooling is enabled by default
config := pkg.NewConfigFromEnvironment()

// Create clients for multiple services
authClient := config.NewAuthClient()
transferClient := config.NewTransferClient("your-access-token")
searchClient := config.NewSearchClient("your-access-token")

// Make requests - connections are automatically pooled and reused
```

### Customizing Connection Pool Settings

For advanced use cases, you can customize the connection pool settings:

```go
// Create custom connection pool configuration
poolConfig := &transport.ConnectionPoolConfig{
    MaxIdleConnsPerHost:   20,      // Max idle connections per host
    MaxIdleConns:          100,     // Max idle connections total
    MaxConnsPerHost:       50,      // Max connections per host
    IdleConnTimeout:       120 * time.Second,
    DisableKeepAlives:     false,
    ResponseHeaderTimeout: 45 * time.Second,
}

// Get a connection pool for the transfer service
pool := transport.GetServicePool("transfer", poolConfig)

// Create an HTTP client with the connection pool
httpClient := pool.GetClient()

// Create an SDK configuration with the custom HTTP client
config := pkg.NewConfigFromEnvironment().
    WithHTTPClient(httpClient)

// Create a transfer client that uses the connection pool
transferClient := config.NewTransferClient("your-access-token")
```

#### Connection Pool Manager

For applications that need to manage multiple connection pools across different services:

```go
// Create a connection pool manager
manager := transport.NewConnectionPoolManager(nil)

// Get connection pools for different services
transferPool := manager.GetPool("transfer", nil)
authPool := manager.GetPool("auth", nil)

// Get clients using the pools
transferClient := transferPool.GetClient()
authClient := authPool.GetClient()

// Get stats on all pools
stats := manager.GetAllStats()
fmt.Printf("Transfer pool stats: %+v\n", stats["transfer"])
fmt.Printf("Auth pool stats: %+v\n", stats["auth"])

// Close idle connections across all pools
manager.CloseAllIdleConnections()
```

### Monitoring Connection Pool Usage

You can monitor the usage of the connection pool to understand its behavior and fine-tune the configuration:

```go
// Get a connection pool
pool := transport.GetServicePool("transfer", nil)

// Get and print statistics
stats := pool.GetStats()
fmt.Printf("Connection pool stats:\n")
fmt.Printf("  Active hosts:         %d\n", stats.ActiveHosts)
fmt.Printf("  Total active:         %d\n", stats.TotalActive)
fmt.Printf("  Max idle per host:    %d\n", stats.Config.MaxIdleConnsPerHost)
fmt.Printf("  Max connections/host: %d\n", stats.Config.MaxConnsPerHost)
fmt.Printf("  Active connections by host:\n")
for host, count := range stats.ActiveByHost {
    fmt.Printf("    %s: %d\n", host, count)
}
```

### Connection Pool Configuration Options

Here's a detailed explanation of the available configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `MaxIdleConnsPerHost` | Maximum number of idle connections to keep per host | CPU count × 2 |
| `MaxIdleConns` | Maximum number of idle connections across all hosts | 100 |
| `MaxConnsPerHost` | Limit on the total number of connections per host | CPU count × 4 |
| `IdleConnTimeout` | How long an idle connection will remain idle before being closed | 90 seconds |
| `DisableKeepAlives` | Disables HTTP keep-alives (not recommended) | false |
| `ResponseHeaderTimeout` | Time to wait for a server's response headers | 30 seconds |
| `ExpectContinueTimeout` | Time to wait for a 100-continue response | 5 seconds |
| `TLSHandshakeTimeout` | Maximum time for a TLS handshake | 10 seconds |
| `TLSClientConfig` | Custom TLS configuration | nil |

### Connection Pooling Performance Considerations

Connection pooling can significantly improve performance:

- **Reduced latency**: Eliminates TCP and TLS handshake overhead for subsequent requests
- **Improved throughput**: Allows more efficient use of connections with HTTP/2 multiplexing
- **Reduced resource usage**: Manages the number of connections to prevent resource exhaustion
- **Better error handling**: Detects and removes stale connections automatically

## Rate Limiting and Resilience

The SDK provides robust mechanisms for handling API rate limits and transient failures.

### Rate Limiting

The `RateLimiter` interface provides rate limiting functionality:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
```

#### Creating a Rate Limiter

```go
// Create a token bucket rate limiter
options := &ratelimit.RateLimiterOptions{
    RequestsPerSecond: 10.0,
    BurstSize:         20,
    UseAdaptive:       true,
    MaxRetryCount:     5,
    MinRetryDelay:     100 * time.Millisecond,
    MaxRetryDelay:     60 * time.Second,
    UseJitter:         true,
    JitterFactor:      0.2,
}

limiter := ratelimit.NewTokenBucketLimiter(options)
```

#### Using the Rate Limiter

```go
// Wait before making a request
ctx := context.Background()
err := limiter.Wait(ctx)
if err != nil {
    return err // Context canceled or other error
}

// Make your API request
response, err := client.MakeAPICall()

// Update rate limiter from response
ratelimit.UpdateRateLimiterFromResponse(limiter, response)
```

#### Adaptive Rate Limiting

The limiter can adapt to server-provided rate limits:

```go
// Extract rate limit information from response headers
info, found := ratelimit.ExtractRateLimitInfo(response)
if found {
    fmt.Printf("Rate limit: %d, Remaining: %d, Reset: %d\n",
        info.Limit, info.Remaining, info.Reset)
}

// Update limiter based on response headers (automatic)
ratelimit.UpdateRateLimiterFromResponse(limiter, response)
```

### Backoff and Retry

The `BackoffStrategy` interface provides retry logic with exponential backoff:

```go
// Create a backoff strategy
backoff := ratelimit.NewExponentialBackoff(
    100*time.Millisecond, // Initial delay
    60*time.Second,       // Max delay
    2.0,                  // Factor (doubles each retry)
    5,                    // Max attempts
)

// Use retry with backoff
err := ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
    // Make your API call
    return client.MakeAPICall(ctx)
}, backoff, ratelimit.IsRetryableError)
```

#### Customizing Retry Logic

You can customize retry behavior:

```go
// Custom retry decision function
shouldRetry := func(err error) bool {
    // Custom logic to determine if the error is retryable
    if err == nil {
        return false
    }
    
    // Example: retry rate limit errors
    if strings.Contains(err.Error(), "rate limit") {
        return true
    }
    
    return false
}

// Use custom retry logic
err := ratelimit.RetryWithBackoff(ctx, apiCall, backoff, shouldRetry)
```

### Circuit Breaker

The circuit breaker pattern prevents cascading failures:

```go
// Create a circuit breaker
options := &ratelimit.CircuitBreakerOptions{
    Threshold:         5,                // Open after 5 failures
    Timeout:           30 * time.Second, // Half-open after 30 seconds
    HalfOpenSuccesses: 2,                // Close after 2 successes
    OnStateChange: func(from, to ratelimit.CircuitBreakerState) {
        fmt.Printf("Circuit state changed: %v -> %v\n", from, to)
    },
}

cb := ratelimit.NewCircuitBreaker(options)

// Use the circuit breaker
err := cb.Execute(ctx, func(ctx context.Context) error {
    // Make your API call
    return client.MakeAPICall(ctx)
})

if errors.Is(err, ratelimit.ErrCircuitOpen) {
    // Circuit is open, handle accordingly
    return fallbackOperation()
}
```

### Integration with Core Client

The SDK integrates these resilience mechanisms into the core client:

```go
// Create a client with rate limiting
client := core.NewClient(
    core.WithRateLimit(10.0),          // 10 requests per second
    core.WithRetryCount(3),            // Retry up to 3 times
    core.WithCircuitBreaker(true),     // Enable circuit breaker
)
```

### Configuration Recommendations

| Service | Recommended Rate | Burst Size | Max Retries |
|---------|-----------------|------------|------------|
| Auth    | 10 req/s        | 20         | 3          |
| Transfer| 5 req/s         | 10         | 5          |
| Search  | 3 req/s         | 6          | 3          |
| Flows   | 5 req/s         | 10         | 3          |

## Best Practices

### General Performance Optimization

1. **Use Appropriate Benchmarks**: Run benchmarks that match your expected workload.
2. **Monitor Performance Metrics**: Keep track of key metrics like transfer speed, memory usage, and request latency.
3. **Optimize for Your Use Case**: Different applications have different performance requirements.
4. **Profile Before Optimizing**: Use profiling tools to identify bottlenecks before making changes.

### Memory Management Best Practices

1. **Use Memory-Optimized Transfers**: For large directory transfers, especially when memory is constrained.
2. **Control Concurrency**: Limit the number of concurrent operations to prevent memory spikes.
3. **Implement Streaming Processing**: Process data in chunks rather than loading everything into memory.
4. **Monitor Memory Usage**: Keep track of memory usage patterns to detect leaks or inefficiencies.
5. **Release Resources Promptly**: Close file handles, readers, and other resources when done with them.

### Network Optimization

1. **Use Connection Pooling**: Enable and configure connection pooling for efficient network usage.
2. **Implement Batching**: Batch API requests when possible to reduce network overhead.
3. **Optimize for Latency**: Minimize round trips for latency-sensitive operations.
4. **Use Appropriate Timeouts**: Set reasonable timeouts for network operations.
5. **Consider Network Conditions**: Adjust settings based on network quality and constraints.

### Error Handling and Resilience

1. **Use Appropriate Limits**: Start with conservative rate limits and adjust based on experience.
2. **Enable Adaptive Limits**: Allow rate limiters to adjust based on server responses.
3. **Add Jitter to Retries**: Prevent thundering herd problems with randomized delays.
4. **Monitor Circuit Breakers**: Log state changes to understand service health.
5. **Implement Fallbacks**: Have fallback mechanisms when services are unavailable.

## Tuning Guidelines

### Resource-Constrained Environments

For applications running in resource-constrained environments:

1. **Minimize Memory Usage**: Use memory-optimized transfers and streaming processing.
2. **Limit Concurrency**: Reduce parallelism settings to prevent resource exhaustion.
3. **Optimize Connection Pool Settings**: Reduce idle connections and timeouts.
4. **Implement Aggressive Resource Cleanup**: Close and release resources promptly.
5. **Use Conservative Rate Limits**: Lower request rates to reduce resource contention.

### High-Throughput Applications

For applications requiring maximum throughput:

1. **Increase Parallelism**: Use higher parallelism settings for transfers and API requests.
2. **Optimize Connection Pool Size**: Increase maximum connections per host.
3. **Use Batch Operations**: Batch API requests when possible to reduce overhead.
4. **Tune Rate Limits**: Increase request rates and burst sizes for higher throughput.
5. **Optimize for Specific File Sizes**: Adjust settings based on the file size distribution.

### Long-Running Services

For long-running services and applications:

1. **Implement Adaptive Rate Limiting**: Adjust request rates based on server responses.
2. **Use Robust Circuit Breakers**: Prevent cascading failures during service degradation.
3. **Implement Health Checks**: Monitor the health of dependent services.
4. **Manage Connection Lifecycle**: Close idle connections periodically to prevent leaks.
5. **Implement Graceful Degradation**: Degrade functionality gracefully during partial outages.

## Resources

- [Metrics Dashboard Example](../../examples/metrics-dashboard/)
- [Benchmark Example Application](../../examples/benchmark/)
- [Rate Limiting Example Application](../../examples/ratelimit/)
- [Web Application Example](../../examples/webapp/)
- [SDK Code Repository](https://github.com/scttfrdmn/globus-go-sdk)
- [Globus API Documentation](https://docs.globus.org/api/)

## Cross-References

- For more details on recursive transfers, see [Recursive Transfers Guide](../advanced/recursive-transfers.md)
- For information about resumable transfers, see [Resumable Transfers Guide](../advanced/resumable-transfers.md)
- For error handling information, see [Error Handling Guide](../error-handling.md)
- For logging and tracing functionality, see [Logging Guide](../topics/logging.md)
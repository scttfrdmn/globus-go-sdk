# Performance Benchmarking

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This guide explains how to use the performance benchmarking capabilities in the Globus Go SDK to optimize and test transfer operations.

## Overview

The Globus Go SDK includes tools to benchmark and monitor transfer performance. These tools help you:

1. Measure transfer speeds for different file sizes and counts
2. Evaluate the impact of parallelism settings on performance
3. Monitor memory usage during transfers
4. Compare different transfer strategies

## Benchmark Package

The benchmark package (`pkg/benchmark`) provides tools for testing and measuring performance:

```go
import "github.com/yourusername/globus-go-sdk/pkg/benchmark"
```

### Key Components

#### Transfer Benchmarking

The `transfer.go` module provides functions for benchmarking file transfers:

- `BenchmarkTransfer`: Runs a single transfer benchmark with specified parameters
- `RunBenchmarkSuite`: Runs a series of benchmarks with different configurations
- Helper functions for generating test data and processing results

#### Memory Monitoring

The `memory.go` module provides tools for monitoring memory usage:

- `MemorySampler`: Continuously samples memory usage during operations
- `MemoryStats`: Structure containing detailed memory statistics
- Utility functions for retrieving and displaying memory usage information

## Running Benchmarks

### Single Benchmark

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

### Benchmark Suite

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

## Benchmark Configuration

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

## Memory Monitoring

To track memory usage during operations:

```go
// Create memory sampler
sampler := benchmark.NewMemorySampler(100 * time.Millisecond)

// Start sampling
sampler.Start()

// Perform operations...

// Stop sampling
sampler.Stop()

// Get peak memory usage
peakMemoryMB := sampler.GetPeakMemory()
fmt.Printf("Peak memory usage: %.2f MB\n", peakMemoryMB)

// Print detailed summary
sampler.PrintSummary()
```

## Benchmark Results

The `BenchmarkResult` struct contains the results of a benchmark:

| Field | Description |
|-------|-------------|
| `FileSizeMB` | Size of each file in MB |
| `FileCount` | Number of files transferred |
| `TotalSizeMB` | Total size transferred in MB |
| `ElapsedTime` | Total time taken for the transfer |
| `TransferSpeedMBs` | Transfer speed in MB/s |
| `SuccessRate` | Ratio of bytes transferred to bytes expected |
| `TaskID` | Globus task ID for the transfer |
| `MemoryPeakMB` | Peak memory usage during the transfer |
| `CPUUsagePercent` | CPU usage percentage during the transfer |

## Benchmark Tool

The SDK includes a command-line benchmark tool in `examples/benchmark/`:

```bash
go run examples/benchmark/main.go \
  --src=source-endpoint-id \
  --dest=destination-endpoint-id \
  --file-size=10 \
  --file-count=10
```

The tool supports various flags for customizing the benchmark:

- `--src`: Source endpoint ID (required)
- `--dest`: Destination endpoint ID (required)
- `--src-path`: Source path on the endpoint
- `--dest-path`: Destination path on the endpoint
- `--file-size`: Size of each file in MB
- `--file-count`: Number of files to transfer
- `--parallel`: Transfer parallelism
- `--recursive`: Whether to use recursive transfer
- `--size`: Run file size benchmark suite
- `--parallelism-test`: Run parallelism benchmark suite

## Benchmark Analysis

### Analyzing File Size Impact

File size can significantly impact transfer performance:

- **Small Files** (< 1MB): Higher overhead per file, lower overall throughput
- **Medium Files** (1-100MB): Good balance of overhead and throughput
- **Large Files** (> 100MB): Lower overhead per file, higher overall throughput

### Analyzing Parallelism Impact

Parallelism settings affect performance and resource usage:

- **Low Parallelism** (1-2): Lower resource usage but slower transfers
- **Medium Parallelism** (4-8): Good balance for most endpoints
- **High Parallelism** (>8): May improve speed but increases resource usage

### Memory Usage Analysis

Memory usage patterns can help optimize transfers:

- **Per-File Memory**: Memory used per file being transferred
- **Peak Memory**: Maximum memory used during the transfer
- **GC Pressure**: How frequently garbage collection runs

## Best Practices

Based on benchmark results, consider these best practices:

1. **Adjust Parallelism**: Find the optimal parallelism setting for your use case
2. **Use Recursive Transfers**: For many small files, recursive transfers are often more efficient
3. **Balance File Sizes**: When possible, bundle small files or split large files
4. **Monitor Memory**: Be aware of memory usage for large transfers
5. **Consider Endpoint Performance**: Different endpoints may have different optimal settings

## Next Steps

After benchmarking, consider:

1. Integrating optimal settings into your application
2. Creating custom benchmark suites for your specific use cases
3. Setting up regular benchmark runs to track performance over time
4. Using benchmark results to inform resource allocation and planning
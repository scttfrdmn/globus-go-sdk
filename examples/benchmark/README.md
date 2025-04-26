# Globus Transfer Benchmarking Tool

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This tool measures performance characteristics of Globus transfers using the Globus Go SDK. It helps users understand and optimize transfers by measuring:

- Transfer speed for different file sizes
- Performance impact of parallelism settings
- Memory usage during transfers
- Impact of recursive vs. individual file transfers

## Features

- Customizable benchmarks with configurable file sizes and counts
- Memory usage monitoring during transfers
- Automatic test data generation and cleanup
- Pre-defined benchmark suites for file size and parallelism testing
- Detailed performance reporting

## Usage

### Prerequisites

Before running the benchmark tool, you need:

1. Two Globus endpoints that you have access to
2. A Globus account with permission to transfer between these endpoints
3. A Globus API access token or client credentials

### Running a Basic Benchmark

```bash
go run main.go \
  --src=source-endpoint-id \
  --dest=destination-endpoint-id \
  --src-path=/source/path/ \
  --dest-path=/destination/path/ \
  --file-size=10 \
  --file-count=10
```

### Running Benchmark Suites

To run the file size benchmark suite:

```bash
go run main.go \
  --src=source-endpoint-id \
  --dest=destination-endpoint-id \
  --src-path=/source/path/ \
  --dest-path=/destination/path/ \
  --size
```

To run the parallelism benchmark suite:

```bash
go run main.go \
  --src=source-endpoint-id \
  --dest=destination-endpoint-id \
  --src-path=/source/path/ \
  --dest-path=/destination/path/ \
  --parallelism-test
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--src` | (required) | Source endpoint ID |
| `--dest` | (required) | Destination endpoint ID |
| `--src-path` | "~/" | Source path on the endpoint |
| `--dest-path` | "~/" | Destination path on the endpoint |
| `--file-size` | 10.0 | Size of each file in MB |
| `--file-count` | 10 | Number of files to transfer |
| `--parallel` | 4 | Transfer parallelism |
| `--recursive` | true | Use recursive transfer |
| `--generate` | true | Generate test data |
| `--delete` | true | Delete test data after benchmark |
| `--token` | "" | Globus access token (if not provided, will use auth flow) |
| `--size` | false | Run file size benchmark suite |
| `--parallelism-test` | false | Run parallelism benchmark suite |

## Understanding the Results

The benchmark tool provides detailed information about transfer performance:

### Basic Metrics

- **Transfer Speed**: Rate of data transfer in MB/s
- **Elapsed Time**: Total time taken to complete the transfer
- **Success Rate**: Percentage of bytes successfully transferred

### Memory Usage

- **Peak Allocated**: Maximum memory allocated during the transfer
- **Total Allocated**: Cumulative memory allocations
- **GC Cycles**: Number of garbage collection cycles

### Benchmark Suite Results

When running a benchmark suite, you'll get a comparison table showing how different configurations affect performance.

## Optimizing Transfers

Based on benchmark results, you can optimize your transfers by:

1. **Adjusting Parallelism**: Finding the optimal parallelism setting for your endpoints
2. **File Packaging**: Deciding whether to transfer many small files or fewer large files
3. **Memory Management**: Understanding memory requirements for different transfer types
4. **Transfer Method**: Choosing between recursive and individual file transfers

## Examples

### File Size Benchmark Example

```
====== File Size Benchmark Summary ======

| Benchmark      | Size/File | Files     | Total Size     | Time           | Speed (MB/s)   |
|---------------|----------|----------|---------------|---------------|---------------|
| Small Files    | 1.0      | 20       | 20.0          | 12.345s       | 1.62          |
| Medium Files   | 10.0     | 10       | 100.0         | 24.567s       | 4.07          |
| Large Files    | 100.0    | 2        | 200.0         | 35.789s       | 5.59          |
| Very Large File| 500.0    | 1        | 500.0         | 67.890s       | 7.36          |
```

### Parallelism Benchmark Example

```
====== Parallelism Benchmark Summary ======

| Benchmark           | Parallelism  | Time           | Speed (MB/s)   | Memory (MB)    |
|--------------------|------------|---------------|---------------|---------------|
| Sequential          | 1          | 45.678s       | 2.19          | 15.23         |
| Low Parallelism     | 2          | 25.432s       | 3.93          | 17.45         |
| Medium Parallelism  | 4          | 15.654s       | 6.39          | 21.67         |
| High Parallelism    | 8          | 11.234s       | 8.90          | 29.12         |
| Very High Parallelism| 16         | 10.987s       | 9.10          | 42.56         |
```

## Extending the Benchmarks

You can customize the benchmark by modifying the source code:

- Add new benchmark types in `main.go`
- Modify memory sampling interval in `memory.go`
- Adjust test file generation strategy in `transfer.go`

## Performance Considerations

When running benchmarks, keep in mind:

- Network conditions between endpoints will significantly affect results
- Endpoint hardware and configuration play a major role in performance
- Large benchmarks may consume significant memory and storage space
- The Globus service may impose rate limits on transfers

## Troubleshooting

If you encounter issues:

- Ensure you have proper permissions on both endpoints
- Check that your access token has the necessary scopes
- Verify that the endpoint paths exist and are accessible
- Make sure you have sufficient storage space on both endpoints
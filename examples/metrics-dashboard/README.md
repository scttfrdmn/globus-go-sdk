# Metrics Dashboard Example

This example demonstrates the performance monitoring and metrics capabilities of the Globus Go SDK.

## Features

- Performance monitoring for multiple simultaneous transfers
- Visual progress bars for tracking transfer progress
- Real-time throughput and progress statistics
- Persistent storage of transfer metrics
- Historical transfer viewing

## Usage

### Basic Usage

Run the example with default settings:

```bash
go run main.go
```

This will simulate three transfers of different sizes and display real-time metrics and progress bars.

### With Metrics Storage

To save metrics to persistent storage:

```bash
go run main.go --storage-dir ./metrics
```

This will save transfer metrics to the specified directory. If you don't specify a directory, it will use `~/.globus-sdk/metrics`.

### View Historical Transfers

To view previously stored transfers:

```bash
go run main.go --storage-dir ./metrics --history
```

This will display a summary of all previously recorded transfers.

### Load and Continue Transfers

To load previously stored metrics and continue tracking:

```bash
go run main.go --storage-dir ./metrics --load-existing
```

## Command Line Options

- `--storage-dir DIR`: Directory for metrics storage
- `--load-existing`: Load existing metrics from storage
- `--history`: Show historical transfers and exit

## How It Works

This example simulates transfers by periodically updating metrics and progress bars. In a real application, you would connect this to actual transfer operations from the Globus Transfer service.

The key components demonstrated are:

1. `PerformanceMonitor`: Collects and tracks transfer metrics
2. `ProgressBar`: Displays visual progress for transfers
3. `TextReporter`: Formats and displays metrics information
4. `FileMetricsStorage`: Provides persistent storage for metrics

## Integration with Real Transfers

To use these metrics capabilities with real Globus transfers, you would:

1. Start monitoring before beginning a transfer
2. Update metrics when checking task status
3. Store metrics for historical analysis
4. Display progress and statistics to users

See the [performance documentation](../../doc/topics/performance.md) for more detailed information.
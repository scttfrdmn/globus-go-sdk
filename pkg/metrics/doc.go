// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package metrics provides transfer performance monitoring for the Globus Go SDK.

This package implements real-time monitoring of Globus transfer operations,
including throughput measurement, progress tracking, error and retry counting,
and optional persistence of metrics to disk. It is designed for applications
that need to display transfer progress, generate performance reports, or
analyze historical transfer data.

# STABILITY: BETA

This package is in beta. The core monitoring abstractions are stable, but
storage backends and some metric fields may evolve in minor releases as
new use cases are identified. Changes will be documented in the CHANGELOG
with migration guidance.

The following components are considered beta-stable:

  - TransferMetrics struct and all exported fields (TransferID, TaskID,
    SourceEndpoint, DestEndpoint, Label, StartTime, EndTime, TotalBytes,
    BytesTransferred, FilesTotal, FilesTransferred, BytesPerSecond,
    PeakBytesPerSecond, AvgBytesPerSecond, EstimatedTimeLeft, PercentComplete,
    ThroughputSamples, ErrorCount, RetryCount, LastError, Status, LastUpdated)
  - ThroughputSample struct and fields (Timestamp, BytesPerSecond,
    BytesTransferred, FilesTransferred)
  - PerformanceMonitor interface (StartMonitoring, StopMonitoring, UpdateMetrics,
    GetMetrics, ListActiveTransfers, SetTotalBytes, SetTotalFiles, RecordError,
    RecordRetry, SetStatus)
  - DefaultPerformanceMonitor type and constructor (NewPerformanceMonitor)
  - DefaultPerformanceMonitor configuration methods (WithSampleInterval, WithMaxSamples)
  - DefaultPerformanceMonitor storage methods (WithStorage, SaveMetrics,
    LoadMetrics, LoadAllMetrics)
  - MetricsStorage interface (StoreMetrics, RetrieveMetrics, ListTransferIDs,
    DeleteMetrics, Cleanup)
  - FileMetricsStorage type and constructor (NewFileMetricsStorage)
  - FileMetricsStorage methods (StoreMetrics, RetrieveMetrics, ListTransferIDs,
    DeleteMetrics, Cleanup)
  - StorageConfig struct and fields (Storage, SaveInterval, AutoSave,
    AutoCleanup, CleanupAge)

# Compatibility Guarantees

For beta components:
  - Minor backward-incompatible changes may still occur in minor releases
  - New fields may be added to TransferMetrics in minor releases
  - Storage backend interfaces may be extended with new optional methods
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Create a performance monitor and start tracking a transfer:

	monitor := metrics.NewPerformanceMonitor()

	m := monitor.StartMonitoring(
		"my-transfer-id",
		"task-uuid",
		"source-endpoint-uuid",
		"dest-endpoint-uuid",
		"My Transfer Label",
	)

	monitor.SetTotalBytes("my-transfer-id", totalBytes)
	monitor.SetTotalFiles("my-transfer-id", totalFiles)

Update metrics as the transfer progresses:

	monitor.UpdateMetrics("my-transfer-id", bytesTransferred, filesTransferred)

	current, ok := monitor.GetMetrics("my-transfer-id")
	if ok {
		fmt.Printf("Progress: %.1f%% (%.2f MB/s)\n",
			current.PercentComplete,
			current.BytesPerSecond/1e6)
	}

Stop monitoring when the transfer completes:

	monitor.StopMonitoring("my-transfer-id")

Persist metrics to disk with automatic periodic saving:

	storage, err := metrics.NewFileMetricsStorage("/var/lib/myapp/metrics")
	if err != nil {
		// Handle error
	}

	monitor.WithStorage(&metrics.StorageConfig{
		Storage:      storage,
		SaveInterval: 30 * time.Second,
		AutoSave:     true,
		AutoCleanup:  true,
		CleanupAge:   7 * 24 * time.Hour, // Keep 7 days of history
	})
*/
package metrics

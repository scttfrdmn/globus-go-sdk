// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"bytes"
	"testing"
	"time"
)

func TestPerformanceMonitor(t *testing.T) {
	// Create a performance monitor
	monitor := NewPerformanceMonitor()

	// Test starting monitoring
	metrics := monitor.StartMonitoring("transfer-123", "task-123", "source-ep", "dest-ep", "Test Transfer")
	if metrics == nil {
		t.Fatal("Expected metrics to be returned, got nil")
	}

	if metrics.TransferID != "transfer-123" {
		t.Errorf("Expected TransferID to be 'transfer-123', got '%s'", metrics.TransferID)
	}

	if metrics.Status != "ACTIVE" {
		t.Errorf("Expected Status to be 'ACTIVE', got '%s'", metrics.Status)
	}

	// Test setting total bytes and files
	monitor.SetTotalBytes("transfer-123", 1000000)
	monitor.SetTotalFiles("transfer-123", 10)

	// Test updating metrics
	monitor.UpdateMetrics("transfer-123", 100000, 1)

	// Test getting metrics
	retrievedMetrics, exists := monitor.GetMetrics("transfer-123")
	if !exists {
		t.Fatal("Expected metrics to exist")
	}

	if retrievedMetrics.BytesTransferred != 100000 {
		t.Errorf("Expected BytesTransferred to be 100000, got %d", retrievedMetrics.BytesTransferred)
	}

	if retrievedMetrics.FilesTransferred != 1 {
		t.Errorf("Expected FilesTransferred to be 1, got %d", retrievedMetrics.FilesTransferred)
	}

	// Test percent complete calculation
	if retrievedMetrics.PercentComplete != 10.0 {
		t.Errorf("Expected PercentComplete to be 10.0, got %.1f", retrievedMetrics.PercentComplete)
	}

	// Test updating metrics again
	monitor.UpdateMetrics("transfer-123", 200000, 2)

	// Test listing active transfers
	activeTransfers := monitor.ListActiveTransfers()
	if len(activeTransfers) != 1 {
		t.Errorf("Expected 1 active transfer, got %d", len(activeTransfers))
	}

	if activeTransfers[0] != "transfer-123" {
		t.Errorf("Expected active transfer ID to be 'transfer-123', got '%s'", activeTransfers[0])
	}

	// Test throughput samples
	retrievedMetrics, _ = monitor.GetMetrics("transfer-123")
	if len(retrievedMetrics.ThroughputSamples) != 2 {
		t.Errorf("Expected 2 throughput samples, got %d", len(retrievedMetrics.ThroughputSamples))
	}

	// Test setting status
	monitor.SetStatus("transfer-123", "SUCCEEDED")
	retrievedMetrics, _ = monitor.GetMetrics("transfer-123")
	if retrievedMetrics.Status != "SUCCEEDED" {
		t.Errorf("Expected Status to be 'SUCCEEDED', got '%s'", retrievedMetrics.Status)
	}

	// Test recording error
	monitor.RecordError("transfer-123", nil)
	monitor.RecordError("transfer-123", nil)
	retrievedMetrics, _ = monitor.GetMetrics("transfer-123")
	if retrievedMetrics.ErrorCount != 2 {
		t.Errorf("Expected ErrorCount to be 2, got %d", retrievedMetrics.ErrorCount)
	}

	// Test recording retry
	monitor.RecordRetry("transfer-123")
	retrievedMetrics, _ = monitor.GetMetrics("transfer-123")
	if retrievedMetrics.RetryCount != 1 {
		t.Errorf("Expected RetryCount to be 1, got %d", retrievedMetrics.RetryCount)
	}

	// Test stopping monitoring
	monitor.StopMonitoring("transfer-123")
	retrievedMetrics, _ = monitor.GetMetrics("transfer-123")
	if retrievedMetrics.Status != "COMPLETED" {
		t.Errorf("Expected Status to be 'COMPLETED', got '%s'", retrievedMetrics.Status)
	}
	if retrievedMetrics.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}

	// Test listing active transfers after stopping
	activeTransfers = monitor.ListActiveTransfers()
	if len(activeTransfers) != 0 {
		t.Errorf("Expected 0 active transfers, got %d", len(activeTransfers))
	}
}

func TestTextReporter(t *testing.T) {
	// Create a transfer metrics object
	metrics := &TransferMetrics{
		TransferID:         "transfer-123",
		TaskID:             "task-123",
		SourceEndpoint:     "source-ep",
		DestEndpoint:       "dest-ep",
		Label:              "Test Transfer",
		StartTime:          time.Now().Add(-5 * time.Minute),
		EndTime:            time.Now(),
		TotalBytes:         1000000,
		BytesTransferred:   750000,
		FilesTotal:         10,
		FilesTransferred:   7,
		BytesPerSecond:     2500,
		PeakBytesPerSecond: 5000,
		AvgBytesPerSecond:  2500,
		PercentComplete:    75.0,
		EstimatedTimeLeft:  100 * time.Second,
		Status:             "ACTIVE",
		ErrorCount:         2,
		RetryCount:         1,
		LastError:          "Connection reset by peer",
		ThroughputSamples: []ThroughputSample{
			{
				Timestamp:        time.Now().Add(-4 * time.Minute),
				BytesPerSecond:   1000,
				BytesTransferred: 100000,
				FilesTransferred: 1,
			},
			{
				Timestamp:        time.Now().Add(-3 * time.Minute),
				BytesPerSecond:   2000,
				BytesTransferred: 300000,
				FilesTransferred: 3,
			},
			{
				Timestamp:        time.Now().Add(-2 * time.Minute),
				BytesPerSecond:   5000,
				BytesTransferred: 500000,
				FilesTransferred: 5,
			},
			{
				Timestamp:        time.Now().Add(-1 * time.Minute),
				BytesPerSecond:   2500,
				BytesTransferred: 750000,
				FilesTransferred: 7,
			},
		},
	}

	// Create a reporter
	reporter := NewTextReporter()

	// Test summary report
	var buf bytes.Buffer
	err := reporter.ReportSummary(&buf, metrics)
	if err != nil {
		t.Fatalf("ReportSummary returned error: %v", err)
	}

	summary := buf.String()
	if len(summary) == 0 {
		t.Error("Expected summary to be non-empty")
	}

	// Test detailed report
	buf.Reset()
	err = reporter.ReportDetailed(&buf, metrics)
	if err != nil {
		t.Fatalf("ReportDetailed returned error: %v", err)
	}

	detailed := buf.String()
	if len(detailed) == 0 {
		t.Error("Expected detailed report to be non-empty")
	}

	// Test progress report
	buf.Reset()
	err = reporter.ReportProgress(&buf, metrics)
	if err != nil {
		t.Fatalf("ReportProgress returned error: %v", err)
	}

	progress := buf.String()
	if len(progress) == 0 {
		t.Error("Expected progress report to be non-empty")
	}
}

func TestFormatting(t *testing.T) {
	// Test formatBytes
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range testCases {
		result := formatBytes(tc.bytes)
		if result != tc.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", tc.bytes, result, tc.expected)
		}
	}

	// Test formatDuration
	durationTests := []struct {
		duration time.Duration
		expected string
	}{
		{5 * time.Second, "5s"},
		{65 * time.Second, "1m 5s"},
		{3665 * time.Second, "1h 1m 5s"},
	}

	for _, tc := range durationTests {
		result := formatDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("formatDuration(%v) = %s, want %s", tc.duration, result, tc.expected)
		}
	}
}

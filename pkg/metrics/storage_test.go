// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileMetricsStorage(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "metrics-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a storage instance
	storage, err := NewFileMetricsStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create test metrics
	metrics := &TransferMetrics{
		TransferID:      "test-transfer",
		TaskID:          "test-task",
		SourceEndpoint:  "source-endpoint",
		DestEndpoint:    "dest-endpoint",
		Label:           "Test Transfer",
		StartTime:       time.Now().Add(-10 * time.Minute),
		EndTime:         time.Now(),
		TotalBytes:      1000000,
		BytesTransferred: 750000,
		FilesTotal:      10,
		FilesTransferred: 7,
		BytesPerSecond:   1250,
		PeakBytesPerSecond: 2500,
		AvgBytesPerSecond:  1250,
		PercentComplete:    75.0,
		EstimatedTimeLeft:  180 * time.Second,
		Status:             "ACTIVE",
		ErrorCount:         2,
		RetryCount:         1,
		LastError:          "Temporary network error",
		LastUpdated:        time.Now(),
		ThroughputSamples: []ThroughputSample{
			{
				Timestamp:       time.Now().Add(-5 * time.Minute),
				BytesPerSecond:  1000,
				BytesTransferred: 200000,
				FilesTransferred: 2,
			},
			{
				Timestamp:       time.Now().Add(-2 * time.Minute),
				BytesPerSecond:  1500,
				BytesTransferred: 600000,
				FilesTransferred: 5,
			},
		},
	}

	// Test StoreMetrics
	err = storage.StoreMetrics("test-transfer", metrics)
	if err != nil {
		t.Fatalf("StoreMetrics failed: %v", err)
	}

	// Verify the file exists
	filePath := filepath.Join(tmpDir, "test-transfer.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Metrics file was not created at %s", filePath)
	}

	// Test RetrieveMetrics
	retrievedMetrics, err := storage.RetrieveMetrics("test-transfer")
	if err != nil {
		t.Fatalf("RetrieveMetrics failed: %v", err)
	}

	// Verify the metrics were retrieved correctly
	if retrievedMetrics.TransferID != metrics.TransferID {
		t.Errorf("Retrieved TransferID = %s, want %s", retrievedMetrics.TransferID, metrics.TransferID)
	}
	if retrievedMetrics.TaskID != metrics.TaskID {
		t.Errorf("Retrieved TaskID = %s, want %s", retrievedMetrics.TaskID, metrics.TaskID)
	}
	if retrievedMetrics.TotalBytes != metrics.TotalBytes {
		t.Errorf("Retrieved TotalBytes = %d, want %d", retrievedMetrics.TotalBytes, metrics.TotalBytes)
	}
	if retrievedMetrics.BytesTransferred != metrics.BytesTransferred {
		t.Errorf("Retrieved BytesTransferred = %d, want %d", retrievedMetrics.BytesTransferred, metrics.BytesTransferred)
	}
	if retrievedMetrics.Status != metrics.Status {
		t.Errorf("Retrieved Status = %s, want %s", retrievedMetrics.Status, metrics.Status)
	}

	// Check that we have the expected number of throughput samples
	if len(retrievedMetrics.ThroughputSamples) != len(metrics.ThroughputSamples) {
		t.Errorf("Retrieved %d throughput samples, want %d", len(retrievedMetrics.ThroughputSamples), len(metrics.ThroughputSamples))
	}

	// Test ListTransferIDs
	ids, err := storage.ListTransferIDs()
	if err != nil {
		t.Fatalf("ListTransferIDs failed: %v", err)
	}
	if len(ids) != 1 || ids[0] != "test-transfer" {
		t.Errorf("ListTransferIDs = %v, want [test-transfer]", ids)
	}

	// Store a second transfer
	secondMetrics := &TransferMetrics{
		TransferID: "second-transfer",
		TaskID:     "second-task",
		Status:     "COMPLETED",
	}
	if err := storage.StoreMetrics("second-transfer", secondMetrics); err != nil {
		t.Fatalf("Failed to store second metrics: %v", err)
	}

	// Test ListTransferIDs with multiple transfers
	ids, err = storage.ListTransferIDs()
	if err != nil {
		t.Fatalf("ListTransferIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 transfer IDs, got %d: %v", len(ids), ids)
	}

	// Test DeleteMetrics
	err = storage.DeleteMetrics("second-transfer")
	if err != nil {
		t.Fatalf("DeleteMetrics failed: %v", err)
	}

	// Verify the file was deleted
	filePath = filepath.Join(tmpDir, "second-transfer.json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Second metrics file was not deleted")
	}

	// Test Cleanup
	// Create a file with old modification time
	oldFilePath := filepath.Join(tmpDir, "old-transfer.json")
	if err := os.WriteFile(oldFilePath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create old metrics file: %v", err)
	}
	
	// Set the modification time to be in the past
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldFilePath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file modification time: %v", err)
	}

	// Run cleanup for files older than 24 hours
	err = storage.Cleanup(24 * time.Hour)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify the old file was deleted
	if _, err := os.Stat(oldFilePath); !os.IsNotExist(err) {
		t.Errorf("Old metrics file was not deleted during cleanup")
	}

	// Verify the active file was not deleted
	filePath = filepath.Join(tmpDir, "test-transfer.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Active metrics file was incorrectly deleted during cleanup")
	}
}

func TestPerformanceMonitorWithStorage(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a storage instance
	storage, err := NewFileMetricsStorage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create a performance monitor
	monitor := NewPerformanceMonitor()

	// Start monitoring a transfer
	monitor.StartMonitoring(
		"test-transfer",
		"test-task",
		"source-endpoint",
		"dest-endpoint",
		"Test Transfer",
	)

	// Set some metrics
	monitor.SetTotalBytes("test-transfer", 1000000)
	monitor.SetTotalFiles("test-transfer", 10)
	monitor.UpdateMetrics("test-transfer", 300000, 3)

	// Save the metrics manually
	err = monitor.SaveMetrics(storage, "test-transfer")
	if err != nil {
		t.Fatalf("SaveMetrics failed: %v", err)
	}

	// Create a new monitor to test loading
	newMonitor := NewPerformanceMonitor()

	// Load the metrics
	err = newMonitor.LoadMetrics(storage, "test-transfer")
	if err != nil {
		t.Fatalf("LoadMetrics failed: %v", err)
	}

	// Verify the metrics were loaded correctly
	metrics, exists := newMonitor.GetMetrics("test-transfer")
	if !exists {
		t.Fatal("Metrics not found after loading")
	}

	if metrics.TransferID != "test-transfer" {
		t.Errorf("Loaded TransferID = %s, want test-transfer", metrics.TransferID)
	}
	if metrics.TotalBytes != 1000000 {
		t.Errorf("Loaded TotalBytes = %d, want 1000000", metrics.TotalBytes)
	}
	if metrics.BytesTransferred != 300000 {
		t.Errorf("Loaded BytesTransferred = %d, want 300000", metrics.BytesTransferred)
	}
	if metrics.FilesTransferred != 3 {
		t.Errorf("Loaded FilesTransferred = %d, want 3", metrics.FilesTransferred)
	}

	// Test LoadAllMetrics by adding another transfer
	monitor.StartMonitoring(
		"second-transfer",
		"second-task",
		"source-endpoint",
		"dest-endpoint",
		"Second Transfer",
	)
	monitor.SetTotalBytes("second-transfer", 500000)
	monitor.UpdateMetrics("second-transfer", 250000, 5)
	
	// Save both metrics
	err = monitor.SaveMetrics(storage, "second-transfer")
	if err != nil {
		t.Fatalf("SaveMetrics failed for second transfer: %v", err)
	}

	// Create a new monitor and load all metrics
	finalMonitor := NewPerformanceMonitor()
	err = finalMonitor.LoadAllMetrics(storage)
	if err != nil {
		t.Fatalf("LoadAllMetrics failed: %v", err)
	}

	// Verify both transfers were loaded
	metrics1, exists1 := finalMonitor.GetMetrics("test-transfer")
	metrics2, exists2 := finalMonitor.GetMetrics("second-transfer")
	
	if !exists1 || !exists2 {
		t.Fatal("Not all metrics were loaded")
	}

	if metrics1.BytesTransferred != 300000 || metrics2.BytesTransferred != 250000 {
		t.Errorf("Incorrect metrics loaded: %d and %d", metrics1.BytesTransferred, metrics2.BytesTransferred)
	}
}
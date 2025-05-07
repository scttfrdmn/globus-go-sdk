// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MetricsStorage defines the interface for storing and retrieving metrics data
type MetricsStorage interface {
	// StoreMetrics stores transfer metrics data
	StoreMetrics(transferID string, metrics *TransferMetrics) error

	// RetrieveMetrics retrieves transfer metrics data
	RetrieveMetrics(transferID string) (*TransferMetrics, error)

	// ListTransferIDs lists all transfer IDs in storage
	ListTransferIDs() ([]string, error)

	// DeleteMetrics deletes metrics for a transfer
	DeleteMetrics(transferID string) error

	// Cleanup removes old metrics data
	Cleanup(olderThan time.Duration) error
}

// FileMetricsStorage implements the MetricsStorage interface using the filesystem
type FileMetricsStorage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFileMetricsStorage creates a new file-based metrics storage
func NewFileMetricsStorage(baseDir string) (*FileMetricsStorage, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create metrics directory: %w", err)
	}

	return &FileMetricsStorage{
		baseDir: baseDir,
	}, nil
}

// getFilePath returns the file path for a transfer ID
func (s *FileMetricsStorage) getFilePath(transferID string) string {
	// Sanitize the transfer ID to be safe for use in filenames
	safeID := filepath.Clean(transferID)
	return filepath.Join(s.baseDir, safeID+".json")
}

// StoreMetrics stores transfer metrics data
func (s *FileMetricsStorage) StoreMetrics(transferID string, metrics *TransferMetrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get a read lock on the metrics to ensure thread safety
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	// Create a storable copy of the metrics without the mutex
	storableMetrics := &storableTransferMetrics{
		TransferID:         metrics.TransferID,
		TaskID:             metrics.TaskID,
		SourceEndpoint:     metrics.SourceEndpoint,
		DestEndpoint:       metrics.DestEndpoint,
		Label:              metrics.Label,
		StartTime:          metrics.StartTime,
		EndTime:            metrics.EndTime,
		TotalBytes:         metrics.TotalBytes,
		BytesTransferred:   metrics.BytesTransferred,
		FilesTotal:         metrics.FilesTotal,
		FilesTransferred:   metrics.FilesTransferred,
		BytesPerSecond:     metrics.BytesPerSecond,
		PeakBytesPerSecond: metrics.PeakBytesPerSecond,
		AvgBytesPerSecond:  metrics.AvgBytesPerSecond,
		EstimatedTimeLeft:  metrics.EstimatedTimeLeft,
		PercentComplete:    metrics.PercentComplete,
		ThroughputSamples:  metrics.ThroughputSamples,
		ErrorCount:         metrics.ErrorCount,
		RetryCount:         metrics.RetryCount,
		LastError:          metrics.LastError,
		Status:             metrics.Status,
		LastUpdated:        metrics.LastUpdated,
	}

	// Marshal the metrics to JSON
	data, err := json.MarshalIndent(storableMetrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Write the metrics to the file
	if err := os.WriteFile(s.getFilePath(transferID), data, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// RetrieveMetrics retrieves transfer metrics data
func (s *FileMetricsStorage) RetrieveMetrics(transferID string) (*TransferMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read the metrics file
	data, err := os.ReadFile(s.getFilePath(transferID))
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics file: %w", err)
	}

	// Unmarshal the metrics
	var storableMetrics storableTransferMetrics
	if err := json.Unmarshal(data, &storableMetrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	// Convert to TransferMetrics with proper mutex
	metrics := &TransferMetrics{
		TransferID:         storableMetrics.TransferID,
		TaskID:             storableMetrics.TaskID,
		SourceEndpoint:     storableMetrics.SourceEndpoint,
		DestEndpoint:       storableMetrics.DestEndpoint,
		Label:              storableMetrics.Label,
		StartTime:          storableMetrics.StartTime,
		EndTime:            storableMetrics.EndTime,
		TotalBytes:         storableMetrics.TotalBytes,
		BytesTransferred:   storableMetrics.BytesTransferred,
		FilesTotal:         storableMetrics.FilesTotal,
		FilesTransferred:   storableMetrics.FilesTransferred,
		BytesPerSecond:     storableMetrics.BytesPerSecond,
		PeakBytesPerSecond: storableMetrics.PeakBytesPerSecond,
		AvgBytesPerSecond:  storableMetrics.AvgBytesPerSecond,
		EstimatedTimeLeft:  storableMetrics.EstimatedTimeLeft,
		PercentComplete:    storableMetrics.PercentComplete,
		ThroughputSamples:  storableMetrics.ThroughputSamples,
		ErrorCount:         storableMetrics.ErrorCount,
		RetryCount:         storableMetrics.RetryCount,
		LastError:          storableMetrics.LastError,
		Status:             storableMetrics.Status,
		LastUpdated:        storableMetrics.LastUpdated,
	}

	return metrics, nil
}

// ListTransferIDs lists all transfer IDs in storage
func (s *FileMetricsStorage) ListTransferIDs() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read the base directory
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read metrics directory: %w", err)
	}

	// Extract transfer IDs from filenames
	var transferIDs []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if the file is a JSON file
		name := file.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}

		// Remove the extension to get the transfer ID
		transferID := name[:len(name)-5] // Remove .json
		transferIDs = append(transferIDs, transferID)
	}

	return transferIDs, nil
}

// DeleteMetrics deletes metrics for a transfer
func (s *FileMetricsStorage) DeleteMetrics(transferID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete the metrics file
	if err := os.Remove(s.getFilePath(transferID)); err != nil {
		return fmt.Errorf("failed to delete metrics file: %w", err)
	}

	return nil
}

// Cleanup removes old metrics data
func (s *FileMetricsStorage) Cleanup(olderThan time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read the base directory
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read metrics directory: %w", err)
	}

	// Calculate the cutoff time
	cutoff := time.Now().Add(-olderThan)

	// Delete files older than the cutoff
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Get file info to check modification time
		info, err := file.Info()
		if err != nil {
			// Skip files we can't get info for
			continue
		}

		// Check if the file is older than the cutoff
		if info.ModTime().Before(cutoff) {
			// Delete the file
			filePath := filepath.Join(s.baseDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				// Log the error but continue with other files
				fmt.Printf("Failed to delete old metrics file %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// storableTransferMetrics is a version of TransferMetrics that can be marshaled to JSON
// It omits the mutex which cannot be marshaled
type storableTransferMetrics struct {
	// Transfer identification
	TransferID     string
	TaskID         string
	SourceEndpoint string
	DestEndpoint   string
	Label          string

	// Overall metrics
	StartTime        time.Time
	EndTime          time.Time
	TotalBytes       int64
	BytesTransferred int64
	FilesTotal       int64
	FilesTransferred int64

	// Performance metrics
	BytesPerSecond     float64
	PeakBytesPerSecond float64
	AvgBytesPerSecond  float64
	EstimatedTimeLeft  time.Duration
	PercentComplete    float64

	// Time-series data for throughput
	ThroughputSamples []ThroughputSample

	// Error tracking
	ErrorCount int
	RetryCount int
	LastError  string

	// State
	Status      string
	LastUpdated time.Time
}

// Add storage capabilities to DefaultPerformanceMonitor
type StorageConfig struct {
	// The storage backend to use
	Storage MetricsStorage

	// How often to save metrics
	SaveInterval time.Duration

	// Whether to save metrics automatically
	AutoSave bool

	// Whether to automatically cleanup old metrics
	AutoCleanup bool

	// How old metrics should be before cleanup
	CleanupAge time.Duration
}

// WithStorage adds storage capabilities to a performance monitor
func (m *DefaultPerformanceMonitor) WithStorage(config *StorageConfig) *DefaultPerformanceMonitor {
	// Start a goroutine to periodically save metrics
	if config.AutoSave && config.Storage != nil && config.SaveInterval > 0 {
		go m.autoSaveLoop(config.Storage, config.SaveInterval)
	}

	// Start a goroutine to periodically clean up old metrics
	if config.AutoCleanup && config.Storage != nil && config.CleanupAge > 0 {
		go m.autoCleanupLoop(config.Storage, config.CleanupAge)
	}

	return m
}

// autoSaveLoop periodically saves metrics to storage
func (m *DefaultPerformanceMonitor) autoSaveLoop(storage MetricsStorage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		m.saveAllMetrics(storage)
	}
}

// autoCleanupLoop periodically cleans up old metrics
func (m *DefaultPerformanceMonitor) autoCleanupLoop(storage MetricsStorage, age time.Duration) {
	ticker := time.NewTicker(24 * time.Hour) // Clean up once a day
	defer ticker.Stop()

	for range ticker.C {
		if err := storage.Cleanup(age); err != nil {
			fmt.Printf("Failed to clean up old metrics: %v\n", err)
		}
	}
}

// SaveMetrics saves metrics for a transfer to storage
func (m *DefaultPerformanceMonitor) SaveMetrics(storage MetricsStorage, transferID string) error {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("metrics not found for transfer ID: %s", transferID)
	}

	return storage.StoreMetrics(transferID, metrics)
}

// saveAllMetrics saves all active metrics to storage
func (m *DefaultPerformanceMonitor) saveAllMetrics(storage MetricsStorage) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for id, metrics := range m.metrics {
		if err := storage.StoreMetrics(id, metrics); err != nil {
			fmt.Printf("Failed to save metrics for transfer %s: %v\n", id, err)
		}
	}
}

// LoadMetrics loads metrics for a transfer from storage
func (m *DefaultPerformanceMonitor) LoadMetrics(storage MetricsStorage, transferID string) error {
	metrics, err := storage.RetrieveMetrics(transferID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.metrics[transferID] = metrics
	m.mu.Unlock()

	return nil
}

// LoadAllMetrics loads all metrics from storage
func (m *DefaultPerformanceMonitor) LoadAllMetrics(storage MetricsStorage) error {
	// Get all transfer IDs
	ids, err := storage.ListTransferIDs()
	if err != nil {
		return err
	}

	// Load each set of metrics
	for _, id := range ids {
		if err := m.LoadMetrics(storage, id); err != nil {
			fmt.Printf("Failed to load metrics for transfer %s: %v\n", id, err)
		}
	}

	return nil
}

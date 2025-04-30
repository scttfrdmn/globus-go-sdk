// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package metrics

import (
	"sync"
	"time"
)

// TransferMetrics contains performance metrics for a transfer operation
type TransferMetrics struct {
	// Transfer identification
	TransferID    string
	TaskID        string
	SourceEndpoint string
	DestEndpoint   string
	Label         string

	// Overall metrics
	StartTime        time.Time
	EndTime          time.Time
	TotalBytes       int64
	BytesTransferred int64
	FilesTotal       int64
	FilesTransferred int64
	
	// Performance metrics
	BytesPerSecond       float64
	PeakBytesPerSecond   float64
	AvgBytesPerSecond    float64
	EstimatedTimeLeft    time.Duration
	PercentComplete      float64
	
	// Time-series data for throughput
	ThroughputSamples    []ThroughputSample
	
	// Error tracking
	ErrorCount          int
	RetryCount          int
	LastError           string
	
	// State
	Status              string
	LastUpdated         time.Time
	
	// Mutex for thread-safe updates
	mu                  sync.RWMutex
}

// ThroughputSample represents a single throughput measurement
type ThroughputSample struct {
	Timestamp       time.Time
	BytesPerSecond  float64
	BytesTransferred int64
	FilesTransferred int64
}

// PerformanceMonitor provides an interface for monitoring transfer performance
type PerformanceMonitor interface {
	// StartMonitoring begins monitoring a transfer
	StartMonitoring(transferID, taskID, sourceEndpoint, destEndpoint, label string) *TransferMetrics
	
	// StopMonitoring stops monitoring a transfer
	StopMonitoring(transferID string)
	
	// UpdateMetrics updates metrics for a transfer
	UpdateMetrics(transferID string, bytesTransferred, filesTransferred int64)
	
	// GetMetrics gets the current metrics for a transfer
	GetMetrics(transferID string) (*TransferMetrics, bool)
	
	// ListActiveTransfers lists all active transfers being monitored
	ListActiveTransfers() []string
}

// DefaultPerformanceMonitor implements the PerformanceMonitor interface
type DefaultPerformanceMonitor struct {
	metrics map[string]*TransferMetrics
	mu      sync.RWMutex
	
	// Configuration
	sampleInterval time.Duration
	maxSamples     int
}

// NewPerformanceMonitor creates a new performance monitor with default settings
func NewPerformanceMonitor() *DefaultPerformanceMonitor {
	return &DefaultPerformanceMonitor{
		metrics:        make(map[string]*TransferMetrics),
		sampleInterval: 1 * time.Second,
		maxSamples:     300, // 5 minutes of samples at 1 per second
	}
}

// WithSampleInterval sets the sample interval for the monitor
func (m *DefaultPerformanceMonitor) WithSampleInterval(interval time.Duration) *DefaultPerformanceMonitor {
	m.sampleInterval = interval
	return m
}

// WithMaxSamples sets the maximum number of samples to store
func (m *DefaultPerformanceMonitor) WithMaxSamples(maxSamples int) *DefaultPerformanceMonitor {
	m.maxSamples = maxSamples
	return m
}

// StartMonitoring begins monitoring a transfer
func (m *DefaultPerformanceMonitor) StartMonitoring(transferID, taskID, sourceEndpoint, destEndpoint, label string) *TransferMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	metrics := &TransferMetrics{
		TransferID:     transferID,
		TaskID:         taskID,
		SourceEndpoint: sourceEndpoint,
		DestEndpoint:   destEndpoint,
		Label:          label,
		StartTime:      time.Now(),
		Status:         "ACTIVE",
		LastUpdated:    time.Now(),
		ThroughputSamples: make([]ThroughputSample, 0, m.maxSamples),
	}
	
	m.metrics[transferID] = metrics
	return metrics
}

// StopMonitoring stops monitoring a transfer
func (m *DefaultPerformanceMonitor) StopMonitoring(transferID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if metrics, exists := m.metrics[transferID]; exists {
		metrics.mu.Lock()
		metrics.EndTime = time.Now()
		metrics.Status = "COMPLETED"
		metrics.mu.Unlock()
		
		// Keep completed transfers in the map for retrieval
		// In a production system, you might want to clean these up or move them to storage
	}
}

// UpdateMetrics updates metrics for a transfer
func (m *DefaultPerformanceMonitor) UpdateMetrics(transferID string, bytesTransferred, filesTransferred int64) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	now := time.Now()
	
	// Calculate bytes per second based on the time since the last update
	var bytesPerSecond float64
	if len(metrics.ThroughputSamples) > 0 {
		lastSample := metrics.ThroughputSamples[len(metrics.ThroughputSamples)-1]
		bytesDelta := bytesTransferred - lastSample.BytesTransferred
		timeDelta := now.Sub(lastSample.Timestamp).Seconds()
		
		if timeDelta > 0 {
			bytesPerSecond = float64(bytesDelta) / timeDelta
		}
	} else if now.Sub(metrics.StartTime).Seconds() > 0 {
		// First sample, calculate from the start
		bytesPerSecond = float64(bytesTransferred) / now.Sub(metrics.StartTime).Seconds()
	}
	
	// Update current metrics
	metrics.BytesTransferred = bytesTransferred
	metrics.FilesTransferred = filesTransferred
	metrics.BytesPerSecond = bytesPerSecond
	metrics.LastUpdated = now
	
	// Update peak bytes per second
	if bytesPerSecond > metrics.PeakBytesPerSecond {
		metrics.PeakBytesPerSecond = bytesPerSecond
	}
	
	// Calculate percent complete if total bytes is known
	if metrics.TotalBytes > 0 {
		metrics.PercentComplete = float64(bytesTransferred) / float64(metrics.TotalBytes) * 100
	}
	
	// Calculate average bytes per second
	totalTime := now.Sub(metrics.StartTime).Seconds()
	if totalTime > 0 {
		metrics.AvgBytesPerSecond = float64(bytesTransferred) / totalTime
	}
	
	// Calculate estimated time left
	if metrics.BytesPerSecond > 0 && metrics.TotalBytes > 0 {
		remainingBytes := metrics.TotalBytes - metrics.BytesTransferred
		if remainingBytes > 0 {
			secondsLeft := float64(remainingBytes) / metrics.BytesPerSecond
			metrics.EstimatedTimeLeft = time.Duration(secondsLeft * float64(time.Second))
		}
	}
	
	// Add throughput sample
	sample := ThroughputSample{
		Timestamp:       now,
		BytesPerSecond:  bytesPerSecond,
		BytesTransferred: bytesTransferred,
		FilesTransferred: filesTransferred,
	}
	
	// Add the sample
	metrics.ThroughputSamples = append(metrics.ThroughputSamples, sample)
	
	// Limit the number of samples
	if len(metrics.ThroughputSamples) > m.maxSamples {
		// Remove the oldest samples
		excess := len(metrics.ThroughputSamples) - m.maxSamples
		metrics.ThroughputSamples = metrics.ThroughputSamples[excess:]
	}
}

// GetMetrics gets the current metrics for a transfer
func (m *DefaultPerformanceMonitor) GetMetrics(transferID string) (*TransferMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics, exists := m.metrics[transferID]
	return metrics, exists
}

// ListActiveTransfers lists all active transfers being monitored
func (m *DefaultPerformanceMonitor) ListActiveTransfers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var activeTransfers []string
	for id, metrics := range m.metrics {
		metrics.mu.RLock()
		if metrics.Status == "ACTIVE" {
			activeTransfers = append(activeTransfers, id)
		}
		metrics.mu.RUnlock()
	}
	
	return activeTransfers
}

// SetTotalBytes sets the total bytes expected for a transfer
func (m *DefaultPerformanceMonitor) SetTotalBytes(transferID string, totalBytes int64) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	metrics.TotalBytes = totalBytes
}

// SetTotalFiles sets the total files expected for a transfer
func (m *DefaultPerformanceMonitor) SetTotalFiles(transferID string, totalFiles int64) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	metrics.FilesTotal = totalFiles
}

// RecordError records an error for a transfer
func (m *DefaultPerformanceMonitor) RecordError(transferID string, err error) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	metrics.ErrorCount++
	if err != nil {
		metrics.LastError = err.Error()
	}
}

// RecordRetry records a retry for a transfer
func (m *DefaultPerformanceMonitor) RecordRetry(transferID string) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	metrics.RetryCount++
}

// SetStatus sets the status for a transfer
func (m *DefaultPerformanceMonitor) SetStatus(transferID string, status string) {
	m.mu.RLock()
	metrics, exists := m.metrics[transferID]
	m.mu.RUnlock()
	
	if !exists {
		return
	}
	
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	
	metrics.Status = status
	metrics.LastUpdated = time.Now()
}
// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package benchmark

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// BenchmarkResult contains the results of a transfer benchmark
type BenchmarkResult struct {
	FileSizeMB       float64       `json:"file_size_mb"`
	FileCount        int           `json:"file_count"`
	TotalSizeMB      float64       `json:"total_size_mb"`
	ElapsedTime      time.Duration `json:"elapsed_time"`
	TransferSpeedMBs float64       `json:"transfer_speed_mbs"`
	SuccessRate      float64       `json:"success_rate"`
	TaskID           string        `json:"task_id"`
	MemoryPeakMB     float64       `json:"memory_peak_mb"`
	CPUUsagePercent  float64       `json:"cpu_usage_percent"`
}

// TransferBenchmarkConfig holds configuration for a transfer benchmark
type TransferBenchmarkConfig struct {
	FileSizeMB       float64 `json:"file_size_mb"`
	FileCount        int     `json:"file_count"`
	SourceEndpoint   string  `json:"source_endpoint"`
	DestEndpoint     string  `json:"dest_endpoint"`
	SourcePath       string  `json:"source_path"`
	DestPath         string  `json:"dest_path"`
	Parallelism      int     `json:"parallelism"`
	UseRecursive     bool    `json:"use_recursive"`
	GenerateTestData bool    `json:"generate_test_data"`
	DeleteAfter      bool    `json:"delete_after"`
}

// DefaultTransferBenchmarkConfig returns a default configuration for transfer benchmarks
func DefaultTransferBenchmarkConfig() *TransferBenchmarkConfig {
	return &TransferBenchmarkConfig{
		FileSizeMB:       100,  // 100MB per file
		FileCount:        10,   // 10 files (1GB total)
		Parallelism:      4,    // 4 parallel operations
		UseRecursive:     true, // Use recursive transfers
		GenerateTestData: true, // Generate test data
		DeleteAfter:      true, // Delete test data after benchmark
	}
}

// BenchmarkTransfer runs a transfer benchmark with the given configuration
func BenchmarkTransfer(
	ctx context.Context,
	client *transfer.Client,
	config *TransferBenchmarkConfig,
	output io.Writer,
) (*BenchmarkResult, error) {
	if output == nil {
		output = os.Stdout
	}

	fmt.Fprintf(output, "Starting transfer benchmark with %d files of %.2f MB each (%.2f MB total)\n",
		config.FileCount, config.FileSizeMB, float64(config.FileCount)*config.FileSizeMB)

	// Create a result to be populated
	result := &BenchmarkResult{
		FileSizeMB:  config.FileSizeMB,
		FileCount:   config.FileCount,
		TotalSizeMB: float64(config.FileCount) * config.FileSizeMB,
	}

	// Generate test data if needed
	if config.GenerateTestData {
		fmt.Fprintf(output, "Generating test data on source endpoint...\n")
		if err := generateTestData(ctx, client, config, output); err != nil {
			return nil, fmt.Errorf("failed to generate test data: %w", err)
		}
	}

	// Start timer
	startTime := time.Now()

	// Get a submission ID for the transfer
	submissionID, err := client.GetSubmissionID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission ID: %w", err)
	}

	// Set up options for the transfer
	options := map[string]interface{}{
		"submission_id":      submissionID,
		"label":              "Benchmark Transfer",
		"deadline":           time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"sync_level":         "checksum",
		"verify_checksum":    true,
		"preserve_timestamp": true,
		"encrypt_data":       true,
		"notify_on_success":  false,
		"notify_on_fail":     false,
		"notify_on_inactive": false,
		"parallelism":        config.Parallelism,
	}

	// Submit the transfer
	fmt.Fprintf(output, "Submitting transfer task...\n")

	var submitResult *transfer.TaskResponse

	if config.UseRecursive {
		// Submit recursive transfer
		sourcePath := filepath.Join(config.SourcePath, "benchmark")
		destPath := filepath.Join(config.DestPath, "benchmark")
		label := "Benchmark Recursive Transfer"

		submitResult, err = client.SubmitTransfer(
			ctx,
			config.SourceEndpoint, sourcePath,
			config.DestEndpoint, destPath,
			label, options,
		)
	} else {
		// Submit individual transfers for each file
		// For simplicity, we'll just do the first file for now
		sourcePath := filepath.Join(config.SourcePath, "benchmark", "file_0.dat")
		destPath := filepath.Join(config.DestPath, "benchmark", "file_0.dat")
		label := "Benchmark Single File Transfer"

		submitResult, err = client.SubmitTransfer(
			ctx,
			config.SourceEndpoint, sourcePath,
			config.DestEndpoint, destPath,
			label, options,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to submit transfer task: %w", err)
	}

	result.TaskID = submitResult.TaskID
	fmt.Fprintf(output, "Transfer task submitted with ID: %s\n", result.TaskID)

	// Wait for the transfer to complete
	fmt.Fprintf(output, "Waiting for transfer to complete...\n")
	_, err = waitForTaskCompletion(ctx, client, submitResult.TaskID, output)
	if err != nil {
		return nil, fmt.Errorf("error waiting for task completion: %w", err)
	}

	// Stop timer
	elapsedTime := time.Since(startTime)
	result.ElapsedTime = elapsedTime

	// Get task details
	taskInfo, err := client.GetTask(ctx, submitResult.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task details: %w", err)
	}

	// Calculate statistics
	if taskInfo.BytesTransferred > 0 {
		// Compute transfer speed in MB/s
		bytesTransferred := float64(taskInfo.BytesTransferred)
		seconds := elapsedTime.Seconds()
		mbTransferred := bytesTransferred / (1024 * 1024)
		result.TransferSpeedMBs = mbTransferred / seconds
	}

	// Calculate success rate
	result.SuccessRate = float64(taskInfo.FilesTransferred) / float64(config.FileCount) * 100

	// Display results
	fmt.Fprintf(output, "\nBenchmark Results:\n")
	fmt.Fprintf(output, "  Files: %d (%.2f MB each, %.2f MB total)\n",
		config.FileCount, config.FileSizeMB, result.TotalSizeMB)
	fmt.Fprintf(output, "  Time: %s\n", elapsedTime)
	fmt.Fprintf(output, "  Speed: %.2f MB/s\n", result.TransferSpeedMBs)
	fmt.Fprintf(output, "  Success Rate: %.2f%%\n", result.SuccessRate)

	// Clean up if requested
	if config.DeleteAfter {
		fmt.Fprintf(output, "\nCleaning up test data...\n")
		if err := cleanupTestData(ctx, client, config, output); err != nil {
			fmt.Fprintf(output, "Warning: Failed to clean up test data: %v\n", err)
		}
	}

	return result, nil
}

// generateTestData creates test files on the source endpoint
// This is a simplified version that will need to be updated based on the actual API
func generateTestData(
	ctx context.Context,
	client *transfer.Client,
	config *TransferBenchmarkConfig,
	output io.Writer,
) error {
	fmt.Fprintf(output, "Note: Test data generation is disabled in this version\n")
	fmt.Fprintf(output, "Please create test data manually at %s:%s/benchmark\n",
		config.SourceEndpoint, config.SourcePath)

	return nil
}

// cleanupTestData removes test files
func cleanupTestData(
	ctx context.Context,
	client *transfer.Client,
	config *TransferBenchmarkConfig,
	output io.Writer,
) error {
	fmt.Fprintf(output, "Note: Test data cleanup is disabled in this version\n")
	fmt.Fprintf(output, "Please remove test data manually from %s:%s/benchmark\n",
		config.SourceEndpoint, config.SourcePath)

	return nil
}

// waitForTaskCompletion waits for a task to complete
// Returns true if the task succeeded, false otherwise
func waitForTaskCompletion(
	ctx context.Context,
	client *transfer.Client,
	taskID string,
	output io.Writer,
) (bool, error) {
	pollInterval := 2 * time.Second
	maxRetries := 300 // 10 minutes

	for i := 0; i < maxRetries; i++ {
		// Get task status
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			return false, fmt.Errorf("failed to get task status: %w", err)
		}

		// Check if the task is complete
		if task.Status == "SUCCEEDED" {
			if output != nil {
				fmt.Fprintf(output, "Task succeeded\n")
			}
			return true, nil
		} else if task.Status == "FAILED" {
			if output != nil {
				fmt.Fprintf(output, "Task failed: %s\n", task.Status)
			}
			return false, nil
		} else if task.Status != "ACTIVE" {
			if output != nil {
				fmt.Fprintf(output, "Task in unexpected state: %s\n", task.Status)
			}
			return false, nil
		}

		// Task is still running, log progress if output is provided
		if output != nil {
			if task.BytesTransferred > 0 {
				fmt.Fprintf(output, "Progress: %d bytes transferred\n",
					task.BytesTransferred)
			} else {
				fmt.Fprintf(output, "Task is processing (status: %s)\n", task.Status)
			}
		}

		// Wait before polling again
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}

	return false, fmt.Errorf("timeout waiting for task completion")
}

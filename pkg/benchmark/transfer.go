// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package benchmark

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/core"
	"github.com/yourusername/globus-go-sdk/pkg/services/transfer"
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

	// Set up transfer task
	transferItems := make([]transfer.TransferItem, 0, config.FileCount)

	if config.UseRecursive {
		// Use recursive transfer
		fmt.Fprintf(output, "Setting up recursive transfer...\n")
		transferItems = append(transferItems, transfer.TransferItem{
			SourcePath:      filepath.Join(config.SourcePath, "benchmark"),
			DestinationPath: filepath.Join(config.DestPath, "benchmark"),
			Recursive:       true,
		})
	} else {
		// Add each file as a separate transfer item
		fmt.Fprintf(output, "Setting up individual file transfers...\n")
		for i := 0; i < config.FileCount; i++ {
			transferItems = append(transferItems, transfer.TransferItem{
				SourcePath:      filepath.Join(config.SourcePath, "benchmark", fmt.Sprintf("file_%d.dat", i)),
				DestinationPath: filepath.Join(config.DestPath, "benchmark", fmt.Sprintf("file_%d.dat", i)),
			})
		}
	}

	// Create transfer task
	taskOptions := &transfer.SubmitTransferOptions{
		Label:           "Benchmark Transfer",
		SourceEndpoint:  config.SourceEndpoint,
		DestEndpoint:    config.DestEndpoint,
		TransferItems:   transferItems,
		Deadline:        time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		SyncLevel:       transfer.SyncLevelChecksum,
		VerifyChecksum:  true,
		PreserveTimestamp: true,
		EncryptData:     true,
		Notify:          transfer.NotifyOff,
		Parallelism:     config.Parallelism,
	}

	// Submit the transfer task
	fmt.Fprintf(output, "Submitting transfer task...\n")
	submitResult, err := client.SubmitTransfer(ctx, taskOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transfer task: %w", err)
	}

	result.TaskID = submitResult.TaskID
	fmt.Fprintf(output, "Transfer task submitted with ID: %s\n", result.TaskID)

	// Wait for the transfer to complete
	fmt.Fprintf(output, "Waiting for transfer to complete...\n")
	completed, err := waitForTaskCompletion(ctx, client, submitResult.TaskID, output)
	if err != nil {
		return nil, fmt.Errorf("error waiting for task completion: %w", err)
	}

	// Stop timer
	elapsedTime := time.Since(startTime)
	result.ElapsedTime = elapsedTime

	// Get task details
	taskInfo, err := client.GetTask(ctx, submitResult.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task info: %w", err)
	}

	// Calculate statistics
	if completed {
		result.SuccessRate = float64(taskInfo.BytesTransferred) / float64(taskInfo.BytesExpected)
		if result.SuccessRate > 1 {
			result.SuccessRate = 1
		}
	} else {
		result.SuccessRate = 0
	}

	// Calculate transfer speed in MB/s
	if elapsedTime > 0 {
		result.TransferSpeedMBs = result.TotalSizeMB / elapsedTime.Seconds()
	}

	// Print results
	fmt.Fprintf(output, "\nTransfer Benchmark Results:\n")
	fmt.Fprintf(output, "  File Size:       %.2f MB\n", result.FileSizeMB)
	fmt.Fprintf(output, "  File Count:      %d\n", result.FileCount)
	fmt.Fprintf(output, "  Total Size:      %.2f MB\n", result.TotalSizeMB)
	fmt.Fprintf(output, "  Elapsed Time:    %s\n", result.ElapsedTime)
	fmt.Fprintf(output, "  Transfer Speed:  %.2f MB/s\n", result.TransferSpeedMBs)
	fmt.Fprintf(output, "  Success Rate:    %.2f%%\n", result.SuccessRate*100)
	fmt.Fprintf(output, "  Task ID:         %s\n", result.TaskID)

	// Clean up test data if needed
	if config.DeleteAfter {
		fmt.Fprintf(output, "\nCleaning up test data...\n")
		if err := cleanupTestData(ctx, client, config, output); err != nil {
			fmt.Fprintf(output, "Warning: Failed to clean up test data: %v\n", err)
		}
	}

	return result, nil
}

// generateTestData creates test files on the source endpoint
func generateTestData(
	ctx context.Context,
	client *transfer.Client,
	config *TransferBenchmarkConfig,
	output io.Writer,
) error {
	// Create directory structure
	mkdirTask := &transfer.SubmitOperationOptions{
		Endpoint: config.SourceEndpoint,
		Operation: transfer.Operation{
			Op:    transfer.OpMkdir,
			Path:  filepath.Join(config.SourcePath, "benchmark"),
		},
	}

	_, err := client.SubmitOperation(ctx, mkdirTask)
	if err != nil {
		return fmt.Errorf("failed to create benchmark directory: %w", err)
	}

	// Generate files with random data on the local machine, then upload them
	tempDir, err := os.MkdirTemp("", "globus-benchmark")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate files in parallel
	var wg sync.WaitGroup
	errorChan := make(chan error, config.FileCount)
	semaphore := make(chan struct{}, 4) // Limit concurrent file generation

	for i := 0; i < config.FileCount; i++ {
		wg.Add(1)
		go func(fileIndex int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			filePath := filepath.Join(tempDir, fmt.Sprintf("file_%d.dat", fileIndex))
			if err := generateRandomFile(filePath, config.FileSizeMB); err != nil {
				errorChan <- fmt.Errorf("failed to generate file %d: %w", fileIndex, err)
				return
			}

			// Upload file to source endpoint
			uploadTask := &transfer.SubmitTransferOptions{
				Label:          fmt.Sprintf("Upload benchmark file %d", fileIndex),
				SourceEndpoint: core.LocalGCPEndpointID, // Use local GCP endpoint for uploads
				DestEndpoint:   config.SourceEndpoint,
				TransferItems: []transfer.TransferItem{
					{
						SourcePath:      filePath,
						DestinationPath: filepath.Join(config.SourcePath, "benchmark", fmt.Sprintf("file_%d.dat", fileIndex)),
					},
				},
				Deadline:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				SyncLevel:     transfer.SyncLevelExistence,
				VerifyChecksum: true,
			}

			submitResult, err := client.SubmitTransfer(ctx, uploadTask)
			if err != nil {
				errorChan <- fmt.Errorf("failed to submit upload task for file %d: %w", fileIndex, err)
				return
			}

			// Wait for upload to complete
			completed, err := waitForTaskCompletion(ctx, client, submitResult.TaskID, nil)
			if err != nil {
				errorChan <- fmt.Errorf("error waiting for upload task completion for file %d: %w", fileIndex, err)
				return
			}

			if !completed {
				errorChan <- fmt.Errorf("upload task for file %d did not complete successfully", fileIndex)
				return
			}

			fmt.Fprintf(output, "Generated and uploaded file %d/%d\n", fileIndex+1, config.FileCount)
		}(i)
	}

	// Wait for all file operations to complete
	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors generating test data: %v", errors)
	}

	return nil
}

// generateRandomFile creates a file with random data of specified size
func generateRandomFile(filePath string, sizeMB float64) error {
	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate size in bytes
	sizeBytes := int64(sizeMB * 1024 * 1024)

	// Use a buffer to improve performance
	const bufferSize = 64 * 1024 // 64KB buffer
	buffer := make([]byte, bufferSize)

	// Fill file with random data
	remaining := sizeBytes
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	for remaining > 0 {
		writeSize := bufferSize
		if remaining < bufferSize {
			writeSize = int(remaining)
		}

		// Fill buffer with random data
		_, err := rng.Read(buffer[:writeSize])
		if err != nil {
			return err
		}

		// Write buffer to file
		n, err := file.Write(buffer[:writeSize])
		if err != nil {
			return err
		}

		remaining -= int64(n)
	}

	return nil
}

// cleanupTestData removes test files from endpoints
func cleanupTestData(
	ctx context.Context,
	client *transfer.Client,
	config *TransferBenchmarkConfig,
	output io.Writer,
) error {
	// Clean up source endpoint
	rmTask := &transfer.SubmitOperationOptions{
		Endpoint: config.SourceEndpoint,
		Operation: transfer.Operation{
			Op:          transfer.OpDelete,
			Path:        filepath.Join(config.SourcePath, "benchmark"),
			Recursive:   true,
		},
	}

	_, err := client.SubmitOperation(ctx, rmTask)
	if err != nil {
		return fmt.Errorf("failed to clean up source endpoint: %w", err)
	}

	// Clean up destination endpoint
	rmTask = &transfer.SubmitOperationOptions{
		Endpoint: config.DestEndpoint,
		Operation: transfer.Operation{
			Op:          transfer.OpDelete,
			Path:        filepath.Join(config.DestPath, "benchmark"),
			Recursive:   true,
		},
	}

	_, err = client.SubmitOperation(ctx, rmTask)
	if err != nil {
		return fmt.Errorf("failed to clean up destination endpoint: %w", err)
	}

	return nil
}

// waitForTaskCompletion polls a task until it completes or fails
func waitForTaskCompletion(
	ctx context.Context,
	client *transfer.Client,
	taskID string,
	output io.Writer,
) (bool, error) {
	pollInterval := 3 * time.Second
	timeout := 1 * time.Hour
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			// Get task status
			task, err := client.GetTask(ctx, taskID)
			if err != nil {
				return false, err
			}

			// Print status if output is provided
			if output != nil {
				fmt.Fprintf(output, "Task status: %s (Files: %d/%d, Bytes: %d/%d)\n",
					task.Status, task.FilesTransferred, task.FilesExpected,
					task.BytesTransferred, task.BytesExpected)
			}

			// Check if task is done
			switch task.Status {
			case "SUCCEEDED":
				return true, nil
			case "FAILED", "CANCELED":
				return false, fmt.Errorf("task failed with status: %s", task.Status)
			}

			// Wait before polling again
			time.Sleep(pollInterval)
		}
	}

	return false, fmt.Errorf("timeout waiting for task completion")
}

// RunBenchmarkSuite runs a series of transfer benchmarks with varying configurations
func RunBenchmarkSuite(
	ctx context.Context,
	client *transfer.Client,
	baseConfig *TransferBenchmarkConfig,
	output io.Writer,
) ([]*BenchmarkResult, error) {
	if output == nil {
		output = os.Stdout
	}

	// Define test cases
	testCases := []struct {
		name   string
		config func(*TransferBenchmarkConfig) *TransferBenchmarkConfig
	}{
		{
			name: "Small Files (10 x 1MB)",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 1
				newConfig.FileCount = 10
				return &newConfig
			},
		},
		{
			name: "Medium Files (10 x 10MB)",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 10
				newConfig.FileCount = 10
				return &newConfig
			},
		},
		{
			name: "Large Files (2 x 100MB)",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 100
				newConfig.FileCount = 2
				return &newConfig
			},
		},
		{
			name: "Many Small Files (100 x 1MB)",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 1
				newConfig.FileCount = 100
				return &newConfig
			},
		},
		{
			name: "Sequential Transfer",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 10
				newConfig.FileCount = 10
				newConfig.Parallelism = 1
				return &newConfig
			},
		},
		{
			name: "High Parallelism",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 10
				newConfig.FileCount = 10
				newConfig.Parallelism = 8
				return &newConfig
			},
		},
		{
			name: "Individual File Transfers",
			config: func(c *TransferBenchmarkConfig) *TransferBenchmarkConfig {
				newConfig := *c
				newConfig.FileSizeMB = 10
				newConfig.FileCount = 10
				newConfig.UseRecursive = false
				return &newConfig
			},
		},
	}

	// Run benchmarks
	results := make([]*BenchmarkResult, 0, len(testCases))

	for _, tc := range testCases {
		fmt.Fprintf(output, "\n====== Benchmark: %s ======\n\n", tc.name)
		config := tc.config(baseConfig)
		
		result, err := BenchmarkTransfer(ctx, client, config, output)
		if err != nil {
			fmt.Fprintf(output, "Error running benchmark %s: %v\n", tc.name, err)
			continue
		}
		
		results = append(results, result)
		
		// Add a small delay between benchmarks
		time.Sleep(5 * time.Second)
	}

	// Print comparison table
	fmt.Fprintf(output, "\n====== Benchmark Summary ======\n\n")
	fmt.Fprintf(output, "| %-25s | %-10s | %-10s | %-15s | %-15s |\n", 
		"Benchmark", "Size", "Files", "Time", "Speed (MB/s)")
	fmt.Fprintf(output, "|%-25s-|%-10s-|%-10s-|%-15s-|%-15s-|\n",
		"-------------------------", "----------", "----------", "---------------", "---------------")
	
	for i, result := range results {
		fmt.Fprintf(output, "| %-25s | %-10.2f | %-10d | %-15s | %-15.2f |\n",
			testCases[i].name, result.TotalSizeMB, result.FileCount, 
			result.ElapsedTime.Round(time.Millisecond), result.TransferSpeedMBs)
	}

	return results, nil
}
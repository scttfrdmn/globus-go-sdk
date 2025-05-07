// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"
)

// MemoryOptimizedOptions contains options for memory-optimized transfers
type MemoryOptimizedOptions struct {
	// BatchSize is the number of files to process in each batch
	BatchSize int

	// MaxConcurrentTasks is the maximum number of concurrent transfer tasks
	MaxConcurrentTasks int

	// Label is the label for the transfer
	Label string

	// SyncLevel determines when to transfer files (use SyncLevelXXX constants)
	SyncLevel int

	// VerifyChecksum indicates whether to verify checksums
	VerifyChecksum bool

	// PreserveMtime indicates whether to preserve timestamps
	PreserveMtime bool

	// Encrypt indicates whether to encrypt data
	Encrypt bool

	// MaxRetries is the maximum number of retries for failed transfers
	MaxRetries int

	// ShowHidden determines whether to show hidden files
	ShowHidden bool

	// ProgressCallback is called with progress updates
	ProgressCallback func(processed, total int, bytes int64, message string)
}

// DefaultMemoryOptimizedOptions returns default options for memory-optimized transfers
func DefaultMemoryOptimizedOptions() *MemoryOptimizedOptions {
	return &MemoryOptimizedOptions{
		BatchSize:          100,
		MaxConcurrentTasks: 4,
		Label:              fmt.Sprintf("Memory-Optimized Transfer %s", time.Now().Format("2006-01-02 15:04:05")),
		SyncLevel:          SyncLevelChecksum,
		VerifyChecksum:     true,
		PreserveMtime:      true,
		Encrypt:            true,
		MaxRetries:         3,
		ShowHidden:         true,
	}
}

// MemoryOptimizedTransferResult contains the results of a memory-optimized transfer
type MemoryOptimizedTransferResult struct {
	// TaskIDs contains the IDs of all transfer tasks
	TaskIDs []string

	// FilesTransferred is the total number of files transferred
	FilesTransferred int

	// BytesTransferred is the total number of bytes transferred
	BytesTransferred int64

	// FailedFiles is the number of files that failed to transfer
	FailedFiles int

	// ElapsedTime is the total time taken for the transfer
	ElapsedTime time.Duration
}

// SubmitMemoryOptimizedTransfer performs a transfer between endpoints using streaming to minimize memory usage
func (c *Client) SubmitMemoryOptimizedTransfer(
	ctx context.Context,
	sourceEndpointID, sourcePath string,
	destinationEndpointID, destinationPath string,
	options *MemoryOptimizedOptions,
) (*MemoryOptimizedTransferResult, error) {
	if options == nil {
		options = DefaultMemoryOptimizedOptions()
	}

	startTime := time.Now()
	result := &MemoryOptimizedTransferResult{}

	// Create a streaming iterator for the source directory
	iterator, err := NewStreamingFileIterator(ctx, c, sourceEndpointID, sourcePath, &StreamingIteratorOptions{
		Recursive:   true,
		ShowHidden:  options.ShowHidden,
		Concurrency: options.MaxConcurrentTasks,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming iterator: %w", err)
	}

	// Process files in batches to minimize memory usage
	var (
		fileBatch      []FileListItem
		processedCount int64
		totalFiles     int
		mu             sync.Mutex
		wg             sync.WaitGroup
		batchCount     int
		errorCount     int
	)

	// Create a channel for processing batches
	batchChan := make(chan []FileListItem)

	// Start batch processing goroutines
	for i := 0; i < options.MaxConcurrentTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Add(-1)

			for fileBatch := range batchChan {
				// Prepare transfer items
				items := make([]TransferItem, 0, len(fileBatch))
				var batchBytes int64

				for _, file := range fileBatch {
					if file.Type == "file" {
						sourceFilePath := path.Join(sourcePath, file.Name)
						destinationFilePath := path.Join(destinationPath, file.Name)

						items = append(items, TransferItem{
							SourcePath:      sourceFilePath,
							DestinationPath: destinationFilePath,
						})

						batchBytes += file.Size
					}
				}

				if len(items) == 0 {
					continue
				}

				// Create transfer task label
				label := fmt.Sprintf("%s (Batch %d)", options.Label, batchCount+1)

				// Create the transfer task request
				request := &TransferTaskRequest{
					DataType:              "transfer",
					Label:                 label,
					SourceEndpointID:      sourceEndpointID,
					DestinationEndpointID: destinationEndpointID,
					SyncLevel:             options.SyncLevel,
					VerifyChecksum:        options.VerifyChecksum,
					PreserveMtime:         options.PreserveMtime,
					Encrypt:               options.Encrypt,
					Items:                 items,
				}

				// Submit the task
				taskResponse, err := c.CreateTransferTask(ctx, request)

				mu.Lock()
				if err != nil {
					// If we hit an error, we don't fail the whole transfer but log it
					if options.ProgressCallback != nil {
						options.ProgressCallback(
							int(processedCount),
							totalFiles,
							result.BytesTransferred,
							fmt.Sprintf("Error submitting batch %d: %v", batchCount+1, err),
						)
					}
					errorCount++
				} else {
					// Add the task ID to the result
					result.TaskIDs = append(result.TaskIDs, taskResponse.TaskID)
					result.FilesTransferred += len(items)
					result.BytesTransferred += batchBytes

					if options.ProgressCallback != nil {
						options.ProgressCallback(
							int(processedCount),
							totalFiles,
							result.BytesTransferred,
							fmt.Sprintf("Submitted batch %d with %d files (%d bytes)",
								batchCount+1, len(items), batchBytes),
						)
					}
				}
				batchCount++
				mu.Unlock()
			}
		}()
	}

	// Collect files in batches
	fileBatch = make([]FileListItem, 0, options.BatchSize)
	for {
		file, ok := iterator.Next()
		if !ok {
			break
		}

		// Increment the total count
		totalFiles++

		// Add to the current batch
		fileBatch = append(fileBatch, file)

		// If we've reached the batch size, send it for processing
		if len(fileBatch) >= options.BatchSize {
			batchChan <- fileBatch
			fileBatch = make([]FileListItem, 0, options.BatchSize)
		}

		// Increment the processed count
		processedCount++

		// Update progress
		if options.ProgressCallback != nil && processedCount%100 == 0 {
			options.ProgressCallback(
				int(processedCount),
				totalFiles,
				result.BytesTransferred,
				"Scanning files...",
			)
		}
	}

	// Process any remaining files
	if len(fileBatch) > 0 {
		batchChan <- fileBatch
	}

	// Close the batch channel to signal completion
	close(batchChan)

	// Wait for all batches to be processed
	wg.Wait()

	// Check for iterator errors
	if err := iterator.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	// Set the result metrics
	result.ElapsedTime = time.Since(startTime)
	result.FailedFiles = errorCount

	// Final progress update
	if options.ProgressCallback != nil {
		options.ProgressCallback(
			int(processedCount),
			totalFiles,
			result.BytesTransferred,
			fmt.Sprintf("Transfer completed with %d files (%d bytes) in %s",
				result.FilesTransferred, result.BytesTransferred, result.ElapsedTime),
		)
	}

	return result, nil
}

// ListMemoryOptimizedTaskStatus lists the status of all tasks in a memory-optimized transfer
func (c *Client) ListMemoryOptimizedTaskStatus(
	ctx context.Context,
	result *MemoryOptimizedTransferResult,
) ([]Task, error) {
	if result == nil || len(result.TaskIDs) == 0 {
		return nil, fmt.Errorf("no tasks to check")
	}

	tasks := make([]Task, 0, len(result.TaskIDs))

	// Get the status of each task
	for _, taskID := range result.TaskIDs {
		task, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
		}

		tasks = append(tasks, *task)
	}

	return tasks, nil
}

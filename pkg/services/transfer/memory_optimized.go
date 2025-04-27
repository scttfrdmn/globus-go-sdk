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
	// BatchSize is the number of files to transfer in a single Globus task
	BatchSize int
	
	// MaxConcurrentTasks is the maximum number of concurrent transfer tasks
	MaxConcurrentTasks int
	
	// Label is the label prefix for transfer tasks
	Label string
	
	// SyncLevel determines how files are compared
	SyncLevel string
	
	// VerifyChecksum specifies whether to verify checksums
	VerifyChecksum bool
	
	// PreserveTimestamp specifies whether to preserve file timestamps
	PreserveTimestamp bool
	
	// EncryptData specifies whether to encrypt data in transit
	EncryptData bool
	
	// MaxRetries is the maximum number of retries for failed transfers
	MaxRetries int
	
	// ShowHidden specifies whether to include hidden files
	ShowHidden bool
	
	// ProgressCallback is called with progress updates
	ProgressCallback func(processed, total int, bytes int64, message string)
}

// DefaultMemoryOptimizedOptions returns default options for memory-optimized transfers
func DefaultMemoryOptimizedOptions() *MemoryOptimizedOptions {
	return &MemoryOptimizedOptions{
		BatchSize:         100,
		MaxConcurrentTasks: 4,
		Label:             fmt.Sprintf("Memory-Optimized Transfer %s", time.Now().Format("2006-01-02 15:04:05")),
		SyncLevel:         SyncChecksum,
		VerifyChecksum:    true,
		PreserveTimestamp: true,
		EncryptData:       true,
		MaxRetries:        3,
		ShowHidden:        true,
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
		return nil, fmt.Errorf("failed to create file iterator: %w", err)
	}
	defer iterator.Close()
	
	// Create a buffer for batch processing
	buffer := make([]FileListItem, 0, options.BatchSize)
	fileCount := 0
	var bufferMutex sync.Mutex
	
	// Create a wait group for concurrent task submission
	var wg sync.WaitGroup
	taskIDsMutex := sync.Mutex{}
	
	// Create a semaphore to limit concurrent tasks
	semaphore := make(chan struct{}, options.MaxConcurrentTasks)
	
	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	// Process files in batches
	var processedCount, totalBytes int64
	
	// Helper function to submit a batch of files
	submitBatch := func(batch []FileListItem) {
		// Skip empty batches
		if len(batch) == 0 {
			return
		}
		
		// Acquire semaphore
		semaphore <- struct{}{}
		
		// Submit in a goroutine
		wg.Add(1)
		go func(fileBatch []FileListItem) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			// Prepare transfer items
			transferItems := make([]TransferItem, 0, len(fileBatch))
			var batchBytes int64
			
			for _, file := range fileBatch {
				if file.Type == "file" {
					sourcePath := path.Join(sourcePath, file.Path)
					destinationFilePath := path.Join(destinationPath, file.Path)
					
					transferItems = append(transferItems, TransferItem{
						SourcePath:      sourcePath,
						DestinationPath: destinationFilePath,
					})
					
					batchBytes += file.Size
				}
			}
			
			// Skip if no files to transfer
			if len(transferItems) == 0 {
				return
			}
			
			// Create transfer task
			taskOptions := &SubmitTransferOptions{
				Label:             fmt.Sprintf("%s (Batch %d)", options.Label, len(result.TaskIDs)+1),
				SourceEndpoint:    sourceEndpointID,
				DestEndpoint:      destinationEndpointID,
				TransferItems:     transferItems,
				SyncLevel:         options.SyncLevel,
				VerifyChecksum:    options.VerifyChecksum,
				PreserveTimestamp: options.PreserveTimestamp,
				Encrypt:           options.EncryptData,
			}
			
			// Submit the transfer
			taskResult, err := c.SubmitTransfer(ctx, taskOptions)
			if err != nil {
				// If we hit an error, we don't fail the whole transfer but log it
				if options.ProgressCallback != nil {
					options.ProgressCallback(
						int(processedCount),
						-1,
						totalBytes,
						fmt.Sprintf("Error submitting batch: %v", err),
					)
				}
				return
			}
			
			// Update results
			taskIDsMutex.Lock()
			result.TaskIDs = append(result.TaskIDs, taskResult.TaskID)
			result.FilesTransferred += len(transferItems)
			result.BytesTransferred += batchBytes
			taskIDsMutex.Unlock()
			
			// Update progress
			if options.ProgressCallback != nil {
				taskIDsMutex.Lock()
				processed := processedCount
				bytes := totalBytes
				taskIDsMutex.Unlock()
				
				options.ProgressCallback(
					int(processed),
					-1,
					bytes,
					fmt.Sprintf("Submitted batch %d with %d files", len(result.TaskIDs), len(transferItems)),
				)
			}
		}(append([]FileListItem{}, batch...)) // Create a copy of the batch
	}
	
	// Process files using the iterator
	for {
		file, ok := iterator.Next()
		if !ok {
			if err := iterator.Error(); err != nil {
				return nil, fmt.Errorf("error iterating files: %w", err)
			}
			break
		}
		
		fileCount++
		
		// Add file to buffer
		bufferMutex.Lock()
		buffer = append(buffer, file)
		
		// If buffer is full, submit the batch
		if len(buffer) >= options.BatchSize {
			// Create a copy of the buffer
			batch := append([]FileListItem{}, buffer...)
			buffer = buffer[:0] // Clear buffer
			
			// Update progress counters
			processedCount += int64(len(batch))
			for _, f := range batch {
				if f.Type == "file" {
					totalBytes += f.Size
				}
			}
			
			bufferMutex.Unlock()
			submitBatch(batch)
		} else {
			bufferMutex.Unlock()
		}
		
		// If the user provides a progress callback, call it periodically
		if options.ProgressCallback != nil && fileCount%1000 == 0 {
			options.ProgressCallback(
				fileCount,
				-1,
				totalBytes,
				"Processing files...",
			)
		}
	}
	
	// Submit any remaining files
	bufferMutex.Lock()
	remainingBatch := append([]FileListItem{}, buffer...)
	buffer = nil // Release memory
	bufferMutex.Unlock()
	
	// Only submit if there are remaining files
	if len(remainingBatch) > 0 {
		submitBatch(remainingBatch)
	}
	
	// Wait for all tasks to complete
	wg.Wait()
	
	// Calculate elapsed time
	result.ElapsedTime = time.Since(startTime)
	
	return result, nil
}

// WaitForMemoryOptimizedTransfer waits for all tasks in a memory-optimized transfer to complete
func (c *Client) WaitForMemoryOptimizedTransfer(
	ctx context.Context,
	result *MemoryOptimizedTransferResult,
	options *WaitOptions,
) error {
	if options == nil {
		options = &WaitOptions{
			PollInterval: 5 * time.Second,
			Timeout:      1 * time.Hour,
		}
	}
	
	// Create a wait group for concurrent polling
	var wg sync.WaitGroup
	errChan := make(chan error, len(result.TaskIDs))
	
	// Create a semaphore to limit concurrent polling
	semaphore := make(chan struct{}, 4)
	
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	
	// Wait for each task
	for _, taskID := range result.TaskIDs {
		// Acquire semaphore
		semaphore <- struct{}{}
		
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			// Poll until task completes or fails
			ticker := time.NewTicker(options.PollInterval)
			defer ticker.Stop()
			
			for {
				select {
				case <-timeoutCtx.Done():
					errChan <- fmt.Errorf("timeout waiting for task %s", id)
					return
				case <-ticker.C:
					// Check task status
					task, err := c.GetTask(ctx, id)
					if err != nil {
						errChan <- fmt.Errorf("error checking task %s: %w", id, err)
						return
					}
					
					// Check if task is done
					switch task.Status {
					case "SUCCEEDED":
						return
					case "FAILED", "CANCELED":
						errChan <- fmt.Errorf("task %s failed with status: %s", id, task.Status)
						return
					}
				}
			}
		}(taskID)
	}
	
	// Wait for all polling to complete
	wg.Wait()
	
	// Check for errors
	close(errChan)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	
	// Return error if any occurred
	if len(errors) > 0 {
		return fmt.Errorf("errors in transfer tasks: %v", errors)
	}
	
	return nil
}

// WaitOptions contains options for waiting on task completion
type WaitOptions struct {
	// PollInterval is how often to check task status
	PollInterval time.Duration
	
	// Timeout is the maximum time to wait for tasks to complete
	Timeout time.Duration
	
	// ProgressCallback is called with progress updates
	ProgressCallback func(completed, total int, message string)
}
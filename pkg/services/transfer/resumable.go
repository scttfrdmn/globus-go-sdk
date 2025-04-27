// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// ResumableTransferResult contains the results of a resumable transfer
type ResumableTransferResult struct {
	// CheckpointID is the ID of the checkpoint for resuming the transfer
	CheckpointID string

	// CompletedItems is the number of items that were successfully transferred
	CompletedItems int

	// FailedItems is the number of items that failed to transfer
	FailedItems int

	// RemainingItems is the number of items that still need to be transferred
	RemainingItems int

	// TaskIDs is the list of task IDs that were created for this transfer
	TaskIDs []string

	// Completed indicates whether the transfer is complete
	Completed bool

	// Duration is how long the transfer has been running
	Duration time.Duration

	// State is the current state of the transfer
	State *CheckpointState
}

// CreateResumableTransfer creates a new resumable transfer
func (c *Client) CreateResumableTransfer(
	ctx context.Context,
	sourceEndpointID, sourcePath string,
	destinationEndpointID, destinationPath string,
	options *ResumableTransferOptions,
) (string, error) {
	// Use default options if none provided
	if options == nil {
		options = DefaultResumableTransferOptions()
	}

	// Generate a unique checkpoint ID
	checkpointID, err := generateCheckpointID()
	if err != nil {
		return "", fmt.Errorf("failed to generate checkpoint ID: %w", err)
	}

	// Create the initial checkpoint state
	state := &CheckpointState{
		CheckpointID: checkpointID,
	}

	// Set task info
	state.TaskInfo.SourceEndpointID = sourceEndpointID
	state.TaskInfo.DestinationEndpointID = destinationEndpointID
	state.TaskInfo.SourceBasePath = sourcePath
	state.TaskInfo.DestinationBasePath = destinationPath
	state.TaskInfo.Label = fmt.Sprintf("Resumable Transfer %s", time.Now().Format("2006-01-02 15:04:05"))
	state.TaskInfo.StartTime = time.Now()
	state.TaskInfo.LastUpdated = time.Now()

	// Set transfer options
	state.TransferOptions = *options

	// List files recursively
	files, err := c.listRecursiveForResumable(ctx, sourceEndpointID, sourcePath, options.SyncLevel > 0)
	if err != nil {
		return "", fmt.Errorf("failed to list source directory: %w", err)
	}

	// Prepare transfer items
	for _, file := range files {
		if file.Type == "file" {
			// Get relative path
			relPath, err := filepath.Rel(sourcePath, filepath.Join(sourcePath, file.Name))
			if err != nil {
				relPath = file.Name
			}

			// Create transfer item
			item := TransferItem{
				SourcePath:      filepath.Join(sourcePath, file.Name),
				DestinationPath: filepath.Join(destinationPath, relPath),
				Recursive:       false,
			}

			// Add to pending items
			state.PendingItems = append(state.PendingItems, item)

			// Update stats
			state.Stats.TotalItems++
			state.Stats.TotalBytes += file.Size
			state.Stats.RemainingItems++
			state.Stats.RemainingBytes += file.Size
		}
	}

	// Create storage for checkpoints
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		return "", fmt.Errorf("failed to create checkpoint storage: %w", err)
	}

	// Save the initial checkpoint
	if err := storage.SaveCheckpoint(ctx, state); err != nil {
		return "", fmt.Errorf("failed to save initial checkpoint: %w", err)
	}

	return checkpointID, nil
}

// ResumeTransfer resumes a previously created resumable transfer
func (c *Client) ResumeTransfer(
	ctx context.Context,
	checkpointID string,
	options *ResumableTransferOptions,
) (*ResumableTransferResult, error) {
	// Create storage for checkpoints
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint storage: %w", err)
	}

	// Load the checkpoint
	state, err := storage.LoadCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Update options if provided
	if options != nil {
		state.TransferOptions = *options
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start the transfer
	result, err := c.executeResumableTransfer(ctx, state, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to execute resumable transfer: %w", err)
	}

	return result, nil
}

// GetTransferCheckpoint gets the current state of a resumable transfer
func (c *Client) GetTransferCheckpoint(
	ctx context.Context,
	checkpointID string,
) (*CheckpointState, error) {
	// Create storage for checkpoints
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint storage: %w", err)
	}

	// Load the checkpoint
	state, err := storage.LoadCheckpoint(ctx, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	return state, nil
}

// ListTransferCheckpoints lists all available transfer checkpoints
func (c *Client) ListTransferCheckpoints(
	ctx context.Context,
) ([]string, error) {
	// Create storage for checkpoints
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint storage: %w", err)
	}

	// List checkpoints
	checkpoints, err := storage.ListCheckpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	return checkpoints, nil
}

// DeleteTransferCheckpoint deletes a transfer checkpoint
func (c *Client) DeleteTransferCheckpoint(
	ctx context.Context,
	checkpointID string,
) error {
	// Create storage for checkpoints
	storage, err := NewFileCheckpointStorage("")
	if err != nil {
		return fmt.Errorf("failed to create checkpoint storage: %w", err)
	}

	// Delete the checkpoint
	if err := storage.DeleteCheckpoint(ctx, checkpointID); err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}

// executeResumableTransfer executes a resumable transfer
func (c *Client) executeResumableTransfer(
	ctx context.Context,
	state *CheckpointState,
	storage CheckpointStorage,
) (*ResumableTransferResult, error) {
	// Create result
	result := &ResumableTransferResult{
		CheckpointID:   state.CheckpointID,
		CompletedItems: len(state.CompletedItems),
		FailedItems:    len(state.FailedItems),
		RemainingItems: len(state.PendingItems),
		TaskIDs:        state.CurrentTasks,
		State:          state,
	}

	// Check if transfer is already complete
	if len(state.PendingItems) == 0 {
		result.Completed = true
		result.Duration = time.Since(state.TaskInfo.StartTime)
		return result, nil
	}

	// Use a wait group to track batch transfers
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errOccurred bool
	var firstErr error

	// Set up checkpoint ticker
	checkpointTicker := time.NewTicker(state.TransferOptions.CheckpointInterval)
	defer checkpointTicker.Stop()

	// Set up progress ticker if callback is provided
	var progressTicker *time.Ticker
	if state.TransferOptions.ProgressCallback != nil {
		progressTicker = time.NewTicker(time.Second * 5)
		defer progressTicker.Stop()
	}

	// Process pending items in batches
	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Check if we have any pending items
		if len(state.PendingItems) == 0 {
			break
		}

		// Process in batches
		var batch []TransferItem
		if len(state.PendingItems) <= state.TransferOptions.BatchSize {
			batch = state.PendingItems
			state.PendingItems = nil
		} else {
			batch = state.PendingItems[:state.TransferOptions.BatchSize]
			state.PendingItems = state.PendingItems[state.TransferOptions.BatchSize:]
		}

		// Submit this batch
		wg.Add(1)
		go func(items []TransferItem) {
			defer wg.Done()

			// Create transfer request
			request := &TransferTaskRequest{
				DataType:               "transfer",
				Label:                  fmt.Sprintf("%s (Batch)", state.TaskInfo.Label),
				SourceEndpointID:       state.TaskInfo.SourceEndpointID,
				DestinationEndpointID:  state.TaskInfo.DestinationEndpointID,
				SyncLevel:              state.TransferOptions.SyncLevel,
				VerifyChecksum:         state.TransferOptions.VerifyChecksum,
				PreserveMtime:          state.TransferOptions.PreserveMtime,
				Encrypt:                state.TransferOptions.Encrypt,
				DeleteDestinationExtra: state.TransferOptions.DeleteDestinationExtra,
				Items:                  items,
			}

			// Submit the transfer
			response, err := c.CreateTransferTask(ctx, request)
			if err != nil {
				mu.Lock()
				if !errOccurred {
					errOccurred = true
					firstErr = fmt.Errorf("failed to submit transfer batch: %w", err)
				}
				// Move items back to pending
				state.PendingItems = append(state.PendingItems, items...)
				mu.Unlock()
				return
			}

			// Track the task ID
			mu.Lock()
			state.CurrentTasks = append(state.CurrentTasks, response.TaskID)
			mu.Unlock()

			// Monitor the task until completion
			for {
				// Check if context is done
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Get task status
				task, err := c.GetTask(ctx, response.TaskID)
				if err != nil {
					// Ignore errors here, we'll retry on the next iteration
					time.Sleep(time.Second * 5)
					continue
				}

				// Check if task is complete
				if task.Status == "SUCCEEDED" {
					mu.Lock()
					// Move all items to completed
					state.CompletedItems = append(state.CompletedItems, items...)
					// Update stats
					for _, item := range items {
						state.Stats.CompletedItems++
						state.Stats.RemainingItems--
						// We don't know the exact size, so approximate
						state.Stats.CompletedBytes += 1
						state.Stats.RemainingBytes -= 1
					}
					mu.Unlock()
					break
				} else if task.Status == "FAILED" {
					mu.Lock()
					// Move all items to failed
					for _, item := range items {
						failedItem := FailedTransferItem{
							Item:         item,
							ErrorMessage: "Task failed",
							RetryCount:   0,
							LastAttempt:  time.Now(),
						}
						state.FailedItems = append(state.FailedItems, failedItem)
						state.Stats.FailedItems++
						state.Stats.RemainingItems--
					}
					mu.Unlock()
					break
				} else if task.Status == "ACTIVE" || task.Status == "INACTIVE" {
					// Task is still running, wait a bit
					time.Sleep(time.Second * 5)
				} else {
					// Task is in an unexpected state, consider it failed
					mu.Lock()
					// Move all items to failed
					for _, item := range items {
						failedItem := FailedTransferItem{
							Item:         item,
							ErrorMessage: fmt.Sprintf("Unexpected task status: %s", task.Status),
							RetryCount:   0,
							LastAttempt:  time.Now(),
						}
						state.FailedItems = append(state.FailedItems, failedItem)
						state.Stats.FailedItems++
						state.Stats.RemainingItems--
					}
					mu.Unlock()
					break
				}
			}
		}(batch)

		// Handle checkpoint and progress tickers
		select {
		case <-checkpointTicker.C:
			// Save checkpoint
			if err := storage.SaveCheckpoint(ctx, state); err != nil {
				return result, fmt.Errorf("failed to save checkpoint: %w", err)
			}
		case <-progressTicker.C:
			// Call progress callback
			if state.TransferOptions.ProgressCallback != nil {
				state.TransferOptions.ProgressCallback(state)
			}
		default:
			// No need to wait, continue processing
		}
	}

	// Wait for all batches to complete
	wg.Wait()

	// Save final checkpoint
	if err := storage.SaveCheckpoint(ctx, state); err != nil {
		return result, fmt.Errorf("failed to save final checkpoint: %w", err)
	}

	// Update result
	result.CompletedItems = len(state.CompletedItems)
	result.FailedItems = len(state.FailedItems)
	result.RemainingItems = len(state.PendingItems)
	result.TaskIDs = state.CurrentTasks
	result.Completed = len(state.PendingItems) == 0
	result.Duration = time.Since(state.TaskInfo.StartTime)
	result.State = state

	// Check if there was an error
	if errOccurred {
		return result, firstErr
	}

	return result, nil
}

// listRecursiveForResumable is a simplified version of listRecursive specifically for resumable transfers
func (c *Client) listRecursiveForResumable(
	ctx context.Context,
	endpointID, dirPath string,
	recursive bool,
) ([]FileListItem, error) {
	var allFiles []FileListItem
	var dirs = []string{dirPath}

	// Process each directory
	for len(dirs) > 0 {
		// Check if context is done
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Get the next directory to process
		currentDir := dirs[0]
		dirs = dirs[1:]

		// List files in the directory
		listOptions := &ListFileOptions{
			ShowHidden: true,
		}
		listing, err := c.ListFiles(ctx, endpointID, currentDir, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list directory %s: %w", currentDir, err)
		}

		// Process files and collect subdirectories
		var newDirs []string
		for _, file := range listing.Data {
			// Add the file to our list
			allFiles = append(allFiles, file)

			// If it's a directory and we're recursive, add it to the list to process
			if file.Type == "dir" && recursive {
				newDirs = append(newDirs, filepath.Join(currentDir, file.Name))
			}
		}

		// Add new directories to the list
		if len(newDirs) > 0 {
			dirs = append(dirs, newDirs...)
		}
	}

	return allFiles, nil
}

// generateCheckpointID generates a unique ID for checkpoints
func generateCheckpointID() (string, error) {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	return hex.EncodeToString(bytes), nil
}
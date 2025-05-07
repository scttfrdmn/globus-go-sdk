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

// RecursiveTransferOptions contains options for recursive transfers
type RecursiveTransferOptions struct {
	// Recursive specifies whether to transfer directories recursively
	Recursive bool

	// PreserveTimestamp specifies whether to preserve file timestamps
	PreserveTimestamp bool

	// VerifyChecksum specifies whether to verify checksums after transfer
	VerifyChecksum bool

	// EncryptData specifies whether to encrypt data in transit
	EncryptData bool

	// Sync specifies whether to sync source and destination
	Sync bool

	// DeleteDestinationExtra specifies whether to delete files in destination that don't exist in source
	DeleteDestinationExtra bool

	// MaxConcurrentListings is the maximum number of concurrent directory listings
	MaxConcurrentListings int

	// MaxConcurrentTransfers is the maximum number of concurrent transfer tasks
	MaxConcurrentTransfers int

	// Label is the label for the transfer task
	Label string

	// SkipDirSizes specifies whether to skip directory size calculations
	SkipDirSizes bool

	// ProgressCallback is called with progress updates
	ProgressCallback func(current, total int64, message string)
}

// DefaultRecursiveTransferOptions returns default options for recursive transfers
func DefaultRecursiveTransferOptions() *RecursiveTransferOptions {
	return &RecursiveTransferOptions{
		Recursive:              true,
		PreserveTimestamp:      true,
		VerifyChecksum:         true,
		EncryptData:            true,
		Sync:                   false,
		DeleteDestinationExtra: false,
		MaxConcurrentListings:  4,
		MaxConcurrentTransfers: 1,
		Label:                  fmt.Sprintf("Recursive Transfer %s", time.Now().Format("2006-01-02 15:04:05")),
		SkipDirSizes:           true,
	}
}

// RecursiveTransferResult contains the results of a recursive transfer
type RecursiveTransferResult struct {
	// TaskID is the ID of the transfer task
	TaskID string

	// TotalFiles is the total number of files transferred
	TotalFiles int64

	// TotalSize is the total size of all files transferred
	TotalSize int64

	// StartTime is when the transfer started
	StartTime time.Time

	// EndTime is when the transfer completed
	EndTime time.Time

	// Directories is the number of directories transferred
	Directories int

	// Subdirectories is the number of nested directories
	Subdirectories int

	// FailedFiles is the number of files that failed to transfer
	FailedFiles int
}

// SubmitRecursiveTransfer submits a recursive transfer between two endpoints
func (c *Client) SubmitRecursiveTransfer(
	ctx context.Context,
	sourceEndpointID, sourcePath string,
	destinationEndpointID, destinationPath string,
	options *RecursiveTransferOptions,
) (*RecursiveTransferResult, error) {
	if options == nil {
		options = DefaultRecursiveTransferOptions()
	}

	result := &RecursiveTransferResult{
		StartTime: time.Now(),
	}

	// Create a context that we can cancel if needed
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Get source directory listing
	sourceFiles, err := c.listRecursive(ctx, sourceEndpointID, sourcePath, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list source directory: %w", err)
	}

	result.Directories = 1 // Count the root directory
	result.Subdirectories = countDirectories(sourceFiles) - 1

	// Calculate total size and file count
	totalSize, totalFiles := calculateTotals(sourceFiles)
	result.TotalFiles = totalFiles
	result.TotalSize = totalSize

	// Prepare transfer items
	transferItems := prepareTransferItems(sourceFiles, sourcePath, destinationPath)

	// Create transfer request
	transferRequest := &TransferTaskRequest{
		DataType:               "transfer",
		Label:                  options.Label,
		SourceEndpointID:       sourceEndpointID,
		DestinationEndpointID:  destinationEndpointID,
		SyncLevel:              getSyncLevel(options),
		VerifyChecksum:         options.VerifyChecksum,
		PreserveMtime:          options.PreserveTimestamp,
		Encrypt:                options.EncryptData,
		DeleteDestinationExtra: options.DeleteDestinationExtra,
		Items:                  transferItems,
	}

	// Submit the transfer
	response, err := c.CreateTransferTask(ctx, transferRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transfer: %w", err)
	}

	result.TaskID = response.TaskID
	result.EndTime = time.Now()

	return result, nil
}

// listRecursive lists files recursively in a directory
func (c *Client) listRecursive(
	ctx context.Context,
	endpointID, dirPath string,
	options *RecursiveTransferOptions,
) ([]FileListItem, error) {
	// Start with the root directory
	var allFiles []FileListItem
	var dirs = []string{dirPath}
	var mutex sync.Mutex
	var wg sync.WaitGroup
	var semaphore = make(chan struct{}, options.MaxConcurrentListings)

	// Keep track of errors
	var errorOccurred bool
	var errorMutex sync.Mutex
	var firstError error

	// Process each directory
	for len(dirs) > 0 && !errorOccurred {
		// Get the next directory to process
		currentDir := dirs[0]
		dirs = dirs[1:]

		// Use a semaphore to limit concurrency
		semaphore <- struct{}{}
		wg.Add(1)

		go func(dir string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			// Check if we already hit an error
			errorMutex.Lock()
			if errorOccurred {
				errorMutex.Unlock()
				return
			}
			errorMutex.Unlock()

			// List files in the directory
			listOptions := &ListFileOptions{
				ShowHidden: true,
			}
			listing, err := c.ListFiles(ctx, endpointID, dir, listOptions)
			if err != nil {
				errorMutex.Lock()
				errorOccurred = true
				firstError = fmt.Errorf("failed to list directory %s: %w", dir, err)
				errorMutex.Unlock()
				return
			}

			// Process files and collect subdirectories
			var newDirs []string
			mutex.Lock()
			for _, file := range listing.Data {
				// Add the file to our list
				allFiles = append(allFiles, file)

				// If it's a directory and we're recursive, add it to the list to process
				if file.Type == "dir" && options.Recursive {
					newDirs = append(newDirs, path.Join(dir, file.Name))
				}
			}
			mutex.Unlock()

			// Add new directories to the list
			if len(newDirs) > 0 {
				mutex.Lock()
				dirs = append(dirs, newDirs...)
				mutex.Unlock()
			}

			// Report progress if callback is provided
			if options.ProgressCallback != nil {
				mutex.Lock()
				dirCount := len(allFiles)
				mutex.Unlock()
				options.ProgressCallback(int64(dirCount), -1, fmt.Sprintf("Listing directory: %s", dir))
			}
		}(currentDir)
	}

	// Wait for all directory listings to complete
	wg.Wait()

	// Check if an error occurred
	if errorOccurred {
		return nil, firstError
	}

	return allFiles, nil
}

// countDirectories counts the number of directories in a file list
func countDirectories(files []FileListItem) int {
	count := 0
	for _, file := range files {
		if file.Type == "dir" {
			count++
		}
	}
	return count
}

// calculateTotals calculates the total size and file count
func calculateTotals(files []FileListItem) (int64, int64) {
	var totalSize int64
	var totalFiles int64

	for _, file := range files {
		if file.Type == "file" {
			totalSize += file.Size
			totalFiles++
		}
	}

	return totalSize, totalFiles
}

// prepareTransferItems prepares transfer items from file listing
func prepareTransferItems(files []FileListItem, sourcePath, destPath string) []TransferItem {
	var items []TransferItem

	// Process regular files first
	for _, file := range files {
		if file.Type == "file" {
			sourceFilePath := path.Join(sourcePath, file.Name)
			destFilePath := path.Join(destPath, file.Name)

			items = append(items, TransferItem{
				SourcePath:      sourceFilePath,
				DestinationPath: destFilePath,
				Recursive:       false,
			})
		}
	}

	return items
}

// getSyncLevel converts options to the appropriate sync level string
func getSyncLevel(options *RecursiveTransferOptions) int {
	if !options.Sync {
		return SyncLevelExists
	}
	if !options.VerifyChecksum {
		return SyncLevelSize
	}
	return SyncLevelChecksum
}

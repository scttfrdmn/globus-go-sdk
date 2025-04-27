// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package search

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BatchIngestOptions contains options for batch document ingestion
type BatchIngestOptions struct {
	// Maximum number of documents per batch
	BatchSize int

	// Maximum number of concurrent batch operations
	MaxConcurrent int

	// Optional task ID prefix
	TaskIDPrefix string

	// Optional callback for batch progress
	ProgressCallback func(processed, total int)

	// Optional timeout per batch
	BatchTimeout time.Duration
}

// DefaultBatchIngestOptions returns the default options for batch ingestion
func DefaultBatchIngestOptions() *BatchIngestOptions {
	return &BatchIngestOptions{
		BatchSize:     1000,
		MaxConcurrent: 5,
		BatchTimeout:  2 * time.Minute,
	}
}

// BatchIngestResult contains the results of a batch ingest operation
type BatchIngestResult struct {
	// Total number of documents processed
	TotalDocuments int

	// Number of documents successfully processed
	SuccessDocuments int

	// Number of documents that failed
	FailedDocuments int

	// List of task IDs created
	TaskIDs []string

	// Map of task IDs to their results
	TaskResults map[string]*IngestResponse

	// Any errors encountered
	Errors []error
}

// BatchIngestDocuments ingests documents in batches
func (c *Client) BatchIngestDocuments(
	ctx context.Context,
	indexID string,
	documents []SearchDocument,
	options *BatchIngestOptions,
) (*BatchIngestResult, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("at least one document is required")
	}

	// Use default options if none provided
	if options == nil {
		options = DefaultBatchIngestOptions()
	}

	// Set defaults for unset options
	if options.BatchSize <= 0 {
		options.BatchSize = 1000
	}
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 5
	}
	if options.BatchTimeout == 0 {
		options.BatchTimeout = 2 * time.Minute
	}

	// Calculate number of batches
	numDocuments := len(documents)
	numBatches := (numDocuments + options.BatchSize - 1) / options.BatchSize

	// Create result
	result := &BatchIngestResult{
		TotalDocuments: numDocuments,
		TaskResults:    make(map[string]*IngestResponse),
		TaskIDs:        make([]string, 0, numBatches),
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, options.MaxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create error channel
	errCh := make(chan error, numBatches)

	// Process batches
	for i := 0; i < numBatches; i++ {
		start := i * options.BatchSize
		end := start + options.BatchSize
		if end > numDocuments {
			end = numDocuments
		}

		batchDocs := documents[start:end]

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(batchNum int, docs []SearchDocument) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Create timeout context for this batch
			batchCtx, cancel := context.WithTimeout(ctx, options.BatchTimeout)
			defer cancel()

			// Prepare request
			ingestReq := &IngestRequest{
				IndexID:   indexID,
				Documents: docs,
			}

			// If task ID prefix is provided, use it
			if options.TaskIDPrefix != "" {
				ingestReq.Task = &IngestTask{
					TaskID: fmt.Sprintf("%s-batch-%d", options.TaskIDPrefix, batchNum),
				}
			}

			// Submit ingest request
			resp, err := c.IngestDocuments(batchCtx, ingestReq)
			if err != nil {
				errCh <- fmt.Errorf("batch %d failed: %w", batchNum, err)
				return
			}

			// Update result
			mu.Lock()
			result.SuccessDocuments += resp.Succeeded
			result.FailedDocuments += resp.Failed
			result.TaskIDs = append(result.TaskIDs, resp.Task.TaskID)
			result.TaskResults[resp.Task.TaskID] = resp
			mu.Unlock()

			// Call progress callback if provided
			if options.ProgressCallback != nil {
				processed := (batchNum + 1) * options.BatchSize
				if processed > numDocuments {
					processed = numDocuments
				}
				options.ProgressCallback(processed, numDocuments)
			}
		}(i, batchDocs)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Collect errors
	for err := range errCh {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// BatchDeleteOptions contains options for batch document deletion
type BatchDeleteOptions struct {
	// Maximum number of subjects per batch
	BatchSize int

	// Maximum number of concurrent batch operations
	MaxConcurrent int

	// Optional task ID prefix
	TaskIDPrefix string

	// Optional callback for batch progress
	ProgressCallback func(processed, total int)

	// Optional timeout per batch
	BatchTimeout time.Duration
}

// DefaultBatchDeleteOptions returns the default options for batch deletion
func DefaultBatchDeleteOptions() *BatchDeleteOptions {
	return &BatchDeleteOptions{
		BatchSize:     1000,
		MaxConcurrent: 5,
		BatchTimeout:  2 * time.Minute,
	}
}

// BatchDeleteResult contains the results of a batch delete operation
type BatchDeleteResult struct {
	// Total number of subjects processed
	TotalSubjects int

	// Number of subjects successfully processed
	SuccessSubjects int

	// Number of subjects that failed
	FailedSubjects int

	// List of task IDs created
	TaskIDs []string

	// Map of task IDs to their results
	TaskResults map[string]*DeleteDocumentsResponse

	// Any errors encountered
	Errors []error
}

// BatchDeleteDocuments deletes documents in batches
func (c *Client) BatchDeleteDocuments(
	ctx context.Context,
	indexID string,
	subjects []string,
	options *BatchDeleteOptions,
) (*BatchDeleteResult, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index ID is required")
	}

	if len(subjects) == 0 {
		return nil, fmt.Errorf("at least one subject is required")
	}

	// Use default options if none provided
	if options == nil {
		options = DefaultBatchDeleteOptions()
	}

	// Set defaults for unset options
	if options.BatchSize <= 0 {
		options.BatchSize = 1000
	}
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 5
	}
	if options.BatchTimeout == 0 {
		options.BatchTimeout = 2 * time.Minute
	}

	// Calculate number of batches
	numSubjects := len(subjects)
	numBatches := (numSubjects + options.BatchSize - 1) / options.BatchSize

	// Create result
	result := &BatchDeleteResult{
		TotalSubjects: numSubjects,
		TaskResults:   make(map[string]*DeleteDocumentsResponse),
		TaskIDs:       make([]string, 0, numBatches),
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, options.MaxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create error channel
	errCh := make(chan error, numBatches)

	// Process batches
	for i := 0; i < numBatches; i++ {
		start := i * options.BatchSize
		end := start + options.BatchSize
		if end > numSubjects {
			end = numSubjects
		}

		batchSubjects := subjects[start:end]

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(batchNum int, subjs []string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Create timeout context for this batch
			batchCtx, cancel := context.WithTimeout(ctx, options.BatchTimeout)
			defer cancel()

			// Prepare request
			deleteReq := &DeleteDocumentsRequest{
				IndexID:  indexID,
				Subjects: subjs,
			}

			// Submit delete request
			resp, err := c.DeleteDocuments(batchCtx, deleteReq)
			if err != nil {
				errCh <- fmt.Errorf("batch %d failed: %w", batchNum, err)
				return
			}

			// Update result
			mu.Lock()
			result.SuccessSubjects += resp.Succeeded
			result.FailedSubjects += resp.Failed
			result.TaskIDs = append(result.TaskIDs, resp.Task.TaskID)
			result.TaskResults[resp.Task.TaskID] = resp
			mu.Unlock()

			// Call progress callback if provided
			if options.ProgressCallback != nil {
				processed := (batchNum + 1) * options.BatchSize
				if processed > numSubjects {
					processed = numSubjects
				}
				options.ProgressCallback(processed, numSubjects)
			}
		}(i, batchSubjects)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Collect errors
	for err := range errCh {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// WaitForTasks waits for multiple tasks to complete
func (c *Client) WaitForTasks(
	ctx context.Context,
	taskIDs []string,
	pollInterval time.Duration,
) ([]*TaskStatusResponse, error) {
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("at least one task ID is required")
	}

	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	// Create results slice
	results := make([]*TaskStatusResponse, len(taskIDs))

	// Create a set of pending task IDs
	pendingTasks := make(map[string]int) // Map from task ID to index in results
	for i, taskID := range taskIDs {
		pendingTasks[taskID] = i
	}

	// Poll until all tasks are complete or context is cancelled
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for len(pendingTasks) > 0 {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case <-ticker.C:
			for taskID, index := range pendingTasks {
				// Use a copy of context to prevent cancellation
				taskCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				status, err := c.GetTaskStatus(taskCtx, taskID)
				cancel()

				if err != nil {
					// Skip this task for now
					continue
				}

				// Store the result
				results[index] = status

				// If the task is complete, remove it from pending
				if status.State == "SUCCESS" || status.State == "FAILED" {
					delete(pendingTasks, taskID)
				}
			}
		}
	}

	return results, nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package flows

import (
	"context"
	"sync"
)

// BatchOptions configures behavior for batch operations.
type BatchOptions struct {
	// Concurrency is the maximum number of concurrent operations.
	// If <= 0, a default of 10 will be used.
	Concurrency int
}

// defaultBatchOptions provides default batch operation settings.
func defaultBatchOptions() *BatchOptions {
	return &BatchOptions{
		Concurrency: 10,
	}
}

// BatchRunFlowsRequest represents a batch of flow run requests.
type BatchRunFlowsRequest struct {
	Requests []*RunRequest
	Options  *BatchOptions
}

// BatchRunFlowsResponse represents the results of a batch flow run operation.
type BatchRunFlowsResponse struct {
	Responses []*BatchRunFlowResult
}

// BatchRunFlowResult represents the result of a single flow run in a batch.
type BatchRunFlowResult struct {
	Response *RunResponse
	Error    error
	Index    int
}

// BatchRunFlows executes multiple flow runs concurrently.
func (c *Client) BatchRunFlows(ctx context.Context, batch *BatchRunFlowsRequest) *BatchRunFlowsResponse {
	// Apply default options if not provided
	options := batch.Options
	if options == nil {
		options = defaultBatchOptions()
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create response container
	response := &BatchRunFlowsResponse{
		Responses: make([]*BatchRunFlowResult, len(batch.Requests)),
	}

	// Use a semaphore pattern to limit concurrency
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Execute each request
	for i, req := range batch.Requests {
		// Store index for result matching
		index := i
		request := req

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute the run
			result := &BatchRunFlowResult{
				Index: index,
			}

			runResponse, err := c.RunFlow(ctx, request)
			if err != nil {
				result.Error = err
			} else {
				result.Response = runResponse
			}

			// Store the result
			response.Responses[index] = result
		}()
	}

	// Wait for all operations to complete
	wg.Wait()
	return response
}

// BatchActionRoleRequest represents a request to retrieve action roles in batch.
type BatchActionRoleRequest struct {
	ProviderID string
	RoleIDs    []string
	Options    *BatchOptions
}

// BatchActionRoleResponse represents the results of a batch action role operation.
type BatchActionRoleResponse struct {
	Responses []*BatchActionRoleResult
}

// BatchActionRoleResult represents the result of a single action role retrieval in a batch.
type BatchActionRoleResult struct {
	Role  *ActionRole
	Error error
	Index int
}

// BatchGetActionRoles retrieves multiple action roles concurrently.
func (c *Client) BatchGetActionRoles(ctx context.Context, batch *BatchActionRoleRequest) *BatchActionRoleResponse {
	// Apply default options if not provided
	options := batch.Options
	if options == nil {
		options = defaultBatchOptions()
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create response container
	response := &BatchActionRoleResponse{
		Responses: make([]*BatchActionRoleResult, len(batch.RoleIDs)),
	}

	// Use a semaphore pattern to limit concurrency
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Execute each request
	for i, roleID := range batch.RoleIDs {
		// Store index for result matching
		index := i
		id := roleID

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute the get
			result := &BatchActionRoleResult{
				Index: index,
			}

			role, err := c.GetActionRole(ctx, batch.ProviderID, id)
			if err != nil {
				result.Error = err
			} else {
				result.Role = role
			}

			// Store the result
			response.Responses[index] = result
		}()
	}

	// Wait for all operations to complete
	wg.Wait()
	return response
}

// BatchRunsRequest represents a request to retrieve flow runs in batch.
type BatchRunsRequest struct {
	RunIDs  []string
	Options *BatchOptions
}

// BatchRunsResponse represents the results of a batch flow run retrieval operation.
type BatchRunsResponse struct {
	Responses []*BatchRunResult
}

// BatchRunResult represents the result of a single flow run retrieval in a batch.
type BatchRunResult struct {
	Run   *RunResponse
	Error error
	Index int
}

// BatchGetRuns retrieves multiple flow runs concurrently.
func (c *Client) BatchGetRuns(ctx context.Context, batch *BatchRunsRequest) *BatchRunsResponse {
	// Apply default options if not provided
	options := batch.Options
	if options == nil {
		options = defaultBatchOptions()
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create response container
	response := &BatchRunsResponse{
		Responses: make([]*BatchRunResult, len(batch.RunIDs)),
	}

	// Use a semaphore pattern to limit concurrency
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Execute each request
	for i, runID := range batch.RunIDs {
		// Store index for result matching
		index := i
		id := runID

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute the get
			result := &BatchRunResult{
				Index: index,
			}

			run, err := c.GetRun(ctx, id)
			if err != nil {
				result.Error = err
			} else {
				result.Run = run
			}

			// Store the result
			response.Responses[index] = result
		}()
	}

	// Wait for all operations to complete
	wg.Wait()
	return response
}

// BatchCancelRunsRequest represents a request to cancel multiple flow runs.
type BatchCancelRunsRequest struct {
	RunIDs  []string
	Options *BatchOptions
}

// BatchCancelRunsResponse represents the results of a batch run cancellation operation.
type BatchCancelRunsResponse struct {
	Responses []*BatchCancelRunResult
}

// BatchCancelRunResult represents the result of a single flow run cancellation in a batch.
type BatchCancelRunResult struct {
	RunID string
	Error error
	Index int
}

// BatchCancelRuns cancels multiple flow runs concurrently.
func (c *Client) BatchCancelRuns(ctx context.Context, batch *BatchCancelRunsRequest) *BatchCancelRunsResponse {
	// Apply default options if not provided
	options := batch.Options
	if options == nil {
		options = defaultBatchOptions()
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create response container
	response := &BatchCancelRunsResponse{
		Responses: make([]*BatchCancelRunResult, len(batch.RunIDs)),
	}

	// Use a semaphore pattern to limit concurrency
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Execute each request
	for i, runID := range batch.RunIDs {
		// Store index for result matching
		index := i
		id := runID

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute the cancellation
			result := &BatchCancelRunResult{
				RunID: id,
				Index: index,
			}

			err := c.CancelRun(ctx, id)
			if err != nil {
				result.Error = err
			}

			// Store the result
			response.Responses[index] = result
		}()
	}

	// Wait for all operations to complete
	wg.Wait()
	return response
}

// BatchFlowsRequest represents a request to retrieve flows in batch.
type BatchFlowsRequest struct {
	FlowIDs []string
	Options *BatchOptions
}

// BatchFlowsResponse represents the results of a batch flow retrieval operation.
type BatchFlowsResponse struct {
	Responses []*BatchFlowResult
}

// BatchFlowResult represents the result of a single flow retrieval in a batch.
type BatchFlowResult struct {
	Flow  *Flow
	Error error
	Index int
}

// BatchGetFlows retrieves multiple flows concurrently.
func (c *Client) BatchGetFlows(ctx context.Context, batch *BatchFlowsRequest) *BatchFlowsResponse {
	// Apply default options if not provided
	options := batch.Options
	if options == nil {
		options = defaultBatchOptions()
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	// Create response container
	response := &BatchFlowsResponse{
		Responses: make([]*BatchFlowResult, len(batch.FlowIDs)),
	}

	// Use a semaphore pattern to limit concurrency
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Execute each request
	for i, flowID := range batch.FlowIDs {
		// Store index for result matching
		index := i
		id := flowID

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore slot
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute the get
			result := &BatchFlowResult{
				Index: index,
			}

			flow, err := c.GetFlow(ctx, id)
			if err != nil {
				result.Error = err
			} else {
				result.Flow = flow
			}

			// Store the result
			response.Responses[index] = result
		}()
	}

	// Wait for all operations to complete
	wg.Wait()
	return response
}
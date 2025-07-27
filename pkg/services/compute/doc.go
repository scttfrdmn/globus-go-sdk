// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package compute provides a client for interacting with the Globus Compute service.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees. Components listed below are
considered part of the public API and will not change incompatibly within a major version:

  - Client interface and implementation
  - Function management operations (register, list, get, update, delete)
  - Basic task execution methods
  - Core model types (Function, Task, Endpoint)
  - Batch processing capabilities
  - Client configuration options
  - Workflow orchestration features
  - Container integration
  - Dependency management
  - Advanced polling and status tracking
  - Task group functionality

# Compatibility Guarantees

For stable packages:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v3.x to v4.x)

# Synchronized Versioning

Starting with v3.60.0-1, this package follows synchronized versioning with the Globus Python SDK.
This ensures API compatibility and feature parity across language implementations.

# Basic Usage

Create a new compute client:

	computeClient := compute.NewClient(
		compute.WithAuthorizer(authorizer),
	)

Function Management:

	// Register a function
	functionID, err := computeClient.RegisterFunction(ctx, &compute.FunctionRegistration{
		Name:    "example-function",
		Code:    "def example(x, y): return x + y",
		Entry:   "example",
		Runtime: "python3.8",
	})
	if err != nil {
		// Handle error
	}

	// List functions
	functions, err := computeClient.ListFunctions(ctx, nil)
	if err != nil {
		// Handle error
	}

	for _, fn := range functions.Functions {
		fmt.Printf("ID: %s, Name: %s\n", fn.ID, fn.Name)
	}

	// Get a function
	function, err := computeClient.GetFunction(ctx, functionID)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Function: %s (%s)\n", function.Name, function.Entry)

	// Delete a function
	err = computeClient.DeleteFunction(ctx, functionID)
	if err != nil {
		// Handle error
	}

Task Execution:

	// Execute a function
	taskID, err := computeClient.RunFunction(ctx, functionID, []interface{}{2, 3})
	if err != nil {
		// Handle error
	}

	// Get task result
	result, err := computeClient.GetTaskResult(ctx, taskID)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Result: %v\n", result.Result)

	// Wait for task completion
	task, err := computeClient.WaitForTask(ctx, taskID)
	if err != nil {
		// Handle error
	}

	if task.IsSuccessful() {
		fmt.Println("Task completed successfully!")
	} else {
		fmt.Printf("Task failed: %s\n", task.Status)
	}

Batch Processing:

	// Create a batch of tasks
	batch := compute.NewBatch()
	batch.AddTask(functionID, []interface{}{1, 2})
	batch.AddTask(functionID, []interface{}{3, 4})

	// Submit the batch
	batchID, err := computeClient.SubmitBatch(ctx, batch)
	if err != nil {
		// Handle error
	}

	// Get batch status
	batchStatus, err := computeClient.GetBatchStatus(ctx, batchID)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Completed tasks: %d/%d\n", batchStatus.Completed, batchStatus.Total)

Container Support:

	// Register a containerized function
	functionID, err := computeClient.RegisterFunction(ctx, &compute.FunctionRegistration{
		Name:        "container-function",
		Code:        "def example(x, y): return x + y",
		Entry:       "example",
		Container:   "my-container-image:latest",
		ContainerID: "docker://ghcr.io/example/my-container:latest",
	})
	if err != nil {
		// Handle error
	}
*/
package compute

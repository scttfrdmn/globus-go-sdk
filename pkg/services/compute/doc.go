// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package compute provides a client for interacting with the Globus Compute service.

# STABILITY: BETA

This package is approaching stability but may still undergo minor changes.
Components listed below are considered relatively stable, but may have
minor signature changes before the package is marked as stable:

  - Client interface and implementation
  - Function management operations (register, list, get, update, delete)
  - Basic task execution methods
  - Core model types (Function, Task, Endpoint)
  - Batch processing capabilities
  - Client configuration options

The following components are less stable and more likely to evolve:

  - Workflow orchestration features
  - Container integration
  - Dependency management
  - Advanced polling and status tracking
  - Task group functionality

# Compatibility Notes

For beta packages:
  - Minor backward-incompatible changes may still occur in minor releases
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices
  - Migration paths will be provided for any breaking changes

This package is expected to reach stable status in version v1.0.0.
Until then, users should review the CHANGELOG when upgrading.

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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package transfer provides a client for interacting with the Globus Transfer service.

# STABILITY: MIXED

This package contains components at different stability levels:

## STABLE Components

The following components are considered stable and will not change incompatibly
within a major version:

  - Client interface and basic implementation
  - Endpoint operations (ListEndpoints, GetEndpoint)
  - Basic file operations (ListDirectory, Mkdir, Rename)
  - Basic transfer operations (CreateTransferTask, CreateDeleteTask)
  - Task management (GetTask, CancelTask, ListTasks)
  - Core model types (Endpoint, Task, FileListItem)
  - Error handling (TransferError, error checking functions)
  - Client configuration options (WithAuthorizer, WithBaseURL, etc.)

## BETA Components

The following components are approaching stability but may still undergo
minor changes:

  - Recursive transfers
  - Task wait operations and event iterator
  - Checkpoint file format

## EXPERIMENTAL Components

The following components are still in development and may change
significantly or be removed:

  - Resumable transfers
  - Memory-optimized operations
  - Streaming iterator functionality

# Compatibility Notes

- Stable components follow semantic versioning guarantees
- Beta components may change in minor versions, but with migration guidance
- Experimental components may change in any release with minimal notice
- All changes will be documented in the CHANGELOG

# Basic Usage

Create a new transfer client:

	transferClient := transfer.NewClient(
		transfer.WithAuthorizer(authorizer),
	)

List endpoints:

	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		// Handle error
	}

	for _, ep := range endpoints.Data {
		fmt.Printf("ID: %s, Display Name: %s\n", ep.ID, ep.DisplayName)
	}

List files on an endpoint:

	files, err := transferClient.ListDirectory(ctx, "endpoint_id", "/path/to/dir")
	if err != nil {
		// Handle error
	}

	for _, f := range files.Data {
		fmt.Printf("Name: %s, Type: %s\n", f.Name, f.Type)
	}

Create a directory:

	err := transferClient.Mkdir(ctx, "endpoint_id", "/path/to/new_dir")
	if err != nil {
		// Handle error
	}

Create a transfer task:

	transferRequest := &transfer.TransferTaskRequest{
		Label: "My Transfer",
		Source: &transfer.TaskEndpoint{
			ID: "source_endpoint_id",
		},
		Destination: &transfer.TaskEndpoint{
			ID: "destination_endpoint_id",
		},
		Items: []transfer.TransferItem{
			{
				SourcePath: "/source/path/file.txt",
				DestinationPath: "/destination/path/file.txt",
			},
		},
	}

	task, err := transferClient.CreateTransferTask(ctx, transferRequest)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Task ID: %s\n", task.TaskID)

Wait for a task to complete:

	task, err := transferClient.WaitForTask(ctx, taskID)
	if err != nil {
		// Handle error
	}

	if task.IsSuccessful() {
		fmt.Println("Transfer completed successfully!")
	} else {
		fmt.Printf("Transfer failed: %s\n", task.Status)
	}

# Advanced Usage

For recursive transfers (BETA):

	task, err := transferClient.SubmitRecursiveTransfer(
		ctx,
		"source_endpoint_id", "/source/dir",
		"destination_endpoint_id", "/destination/dir",
		transfer.WithTransferLabel("Recursive Transfer"),
	)
	if err != nil {
		// Handle error
	}

For resumable transfers (EXPERIMENTAL):

	// Create a resumable transfer
	resumable, err := transferClient.CreateResumableTransfer(
		ctx,
		"source_endpoint_id", "/source/dir",
		"destination_endpoint_id", "/destination/dir",
		transfer.WithCheckpointDir("/path/to/checkpoints"),
	)
	if err != nil {
		// Handle error
	}

	// Start the transfer
	err = resumable.Start(ctx)
	if err != nil {
		// Handle error
	}

	// Later, resume the transfer after interruption
	resumable, err = transferClient.ResumeTransfer(ctx, resumable.ID)
	if err != nil {
		// Handle error
	}

	err = resumable.Resume(ctx)
	if err != nil {
		// Handle error
	}
*/
package transfer

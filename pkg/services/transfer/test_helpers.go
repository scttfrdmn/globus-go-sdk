// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// TestCreateDirectoryOptions contains options for creating a directory (for tests)
type TestCreateDirectoryOptions struct {
	EndpointID string
	Path       string
}

// TestCreateDirectory creates a directory on an endpoint (for tests)
func (c *Client) TestCreateDirectory(ctx context.Context, options *TestCreateDirectoryOptions) error {
	return c.Mkdir(ctx, options.EndpointID, options.Path)
}

// DeleteItemOptions contains options for deleting a file or directory
type DeleteItemOptions struct {
	EndpointID string
	Path       string
	// Note: The API does not support a "recursive" field for delete_item as of API v0.10
	// All deletions in Globus Transfer appear to be recursive by default
}

// DeleteItem deletes a file or directory on an endpoint
func (c *Client) DeleteItem(ctx context.Context, options *DeleteItemOptions) error {
	// Create a delete task with just one item
	request := &DeleteTaskRequest{
		DataType: "delete",
		Label:    "Delete item " + options.Path,
		EndpointID: options.EndpointID,
		Items: []DeleteItem{
			{
				DataType: "delete_item",
				Path:     options.Path,
			},
		},
	}

	_, err := c.CreateDeleteTask(ctx, request)
	return err
}

// TestListDirectoryOptions contains options for listing a directory (for tests)
type TestListDirectoryOptions struct {
	EndpointID string
	Path       string
	ShowHidden bool
	OrderBy    string
	Marker     string
	Limit      int
}

// TestListDirectory lists files and directories in a path on an endpoint (for tests)
func (c *Client) TestListDirectory(ctx context.Context, options *TestListDirectoryOptions) (*FileList, error) {
	listOptions := &ListFileOptions{
		ShowHidden: options.ShowHidden,
		OrderBy:    options.OrderBy,
		Marker:     options.Marker,
		Limit:      options.Limit,
	}

	return c.ListFiles(ctx, options.EndpointID, options.Path, listOptions)
}

// RenameItemOptions contains options for renaming a file or directory
type RenameItemOptions struct {
	EndpointID string
	OldPath    string
	NewPath    string
}

// RenameItem renames a file or directory on an endpoint
func (c *Client) RenameItem(ctx context.Context, options *RenameItemOptions) error {
	return c.Rename(ctx, options.EndpointID, options.OldPath, options.NewPath)
}

// GetTaskEventsOptions contains options for getting task events
type GetTaskEventsOptions struct {
	FilterCode string
	Limit      int
	Offset     int
}

// GetTaskEvents retrieves events for a specific task
func (c *Client) GetTaskEvents(ctx context.Context, taskID string, options *GetTaskEventsOptions) (*TaskEventList, error) {
	// Convert options to query parameters
	query := map[string][]string{}
	if options != nil {
		if options.FilterCode != "" {
			query["filter_code"] = []string{options.FilterCode}
		}
		if options.Limit > 0 {
			query["limit"] = []string{strconv.Itoa(options.Limit)}
		}
		if options.Offset > 0 {
			query["offset"] = []string{strconv.Itoa(options.Offset)}
		}
	}

	var eventList TaskEventList
	err := c.doRequest(ctx, "GET", "task/"+taskID+"/event_list", query, nil, &eventList)
	if err != nil {
		return nil, err
	}

	return &eventList, nil
}

// TaskEventList represents a list of task events
type TaskEventList struct {
	Data []TaskEvent `json:"data"`
}

// TaskEvent represents a task event
type TaskEvent struct {
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
	Details     string    `json:"details,omitempty"`
}

// UpdateTaskLabelOptions contains options for updating a task label
type UpdateTaskLabelOptions struct {
	TaskID string
	Label  string
}

// UpdateTaskLabel updates the label of a task
func (c *Client) UpdateTaskLabel(ctx context.Context, options *UpdateTaskLabelOptions) error {
	body := map[string]string{
		"label": options.Label,
	}

	var result OperationResult
	err := c.doRequest(ctx, "POST", "task/"+options.TaskID, nil, body, &result)
	if err != nil {
		return err
	}

	if result.Code != "Updated" {
		return fmt.Errorf("update task label failed: %s - %s", result.Code, result.Message)
	}

	return nil
}

// WaitForTaskCompletion waits for a task to complete with a timeout
func (c *Client) WaitForTaskCompletion(ctx context.Context, taskID string, interval time.Duration) (*Task, error) {
	var task *Task
	var err error

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return task, ctx.Err()
		case <-ticker.C:
			task, err = c.GetTask(ctx, taskID)
			if err != nil {
				return nil, err
			}

			// Check if the task is done
			if task.Status == "SUCCEEDED" || task.Status == "FAILED" || task.Status == "CANCELLED" {
				return task, nil
			}
		}
	}
}

// TestRecursiveTransferOptions contains options for recursive transfer in test helpers
type TestRecursiveTransferOptions struct {
	SourceEndpointID      string
	DestinationEndpointID string
	SourcePath            string
	DestinationPath       string
	Label                 string
	Sync                  bool
	VerifyChecksum        bool
}

// TestRecursiveTransfer starts a recursive transfer between endpoints using test options
func (c *Client) TestRecursiveTransfer(ctx context.Context, options *TestRecursiveTransferOptions) error {
	// Create transfer request
	transferRequest := &TransferTaskRequest{
		DataType:              "transfer",
		Label:                 options.Label,
		SourceEndpointID:      options.SourceEndpointID,
		DestinationEndpointID: options.DestinationEndpointID,
		Encrypt:               true,
		VerifyChecksum:        options.VerifyChecksum,
		Items: []TransferItem{
			{
				SourcePath:      options.SourcePath,
				DestinationPath: options.DestinationPath,
				Recursive:       true,
			},
		},
	}

	// Submit the transfer
	_, err := c.CreateTransferTask(ctx, transferRequest)
	return err
}
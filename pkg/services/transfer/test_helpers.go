// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// CreateDirectoryOptions contains options for creating a directory
type CreateDirectoryOptions struct {
	EndpointID string
	Path       string
}

// CreateDirectory creates a directory on an endpoint
func (c *Client) CreateDirectory(ctx context.Context, options *CreateDirectoryOptions) error {
	return c.Mkdir(ctx, options.EndpointID, options.Path)
}

// DeleteItemOptions contains options for deleting a file or directory
type DeleteItemOptions struct {
	EndpointID string
	Path       string
	Recursive  bool
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
				Path:      options.Path,
				Recursive: options.Recursive,
			},
		},
	}

	_, err := c.CreateDeleteTask(ctx, request)
	return err
}

// ListDirectoryOptions contains options for listing a directory
type ListDirectoryOptions struct {
	EndpointID string
	Path       string
	ShowHidden bool
	OrderBy    string
	Marker     string
	Limit      int
}

// ListDirectory lists files and directories in a path on an endpoint
func (c *Client) ListDirectory(ctx context.Context, options *ListDirectoryOptions) (*FileList, error) {
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
	DATA []TaskEvent `json:"DATA"`
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

// RecursiveTransferOptions contains options for recursive transfer
type RecursiveTransferOptions struct {
	SourceEndpointID      string
	DestinationEndpointID string
	SourcePath            string
	DestinationPath       string
	Label                 string
	Sync                  bool
	VerifyChecksum        bool
}

// RecursiveTransfer starts a recursive transfer between endpoints
func (c *Client) RecursiveTransfer(ctx context.Context, options *RecursiveTransferOptions) error {
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
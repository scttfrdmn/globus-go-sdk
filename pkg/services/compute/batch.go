// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package compute

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CreateWorkflow creates a new workflow for orchestrating tasks
func (c *Client) CreateWorkflow(ctx context.Context, request *WorkflowCreateRequest) (*WorkflowResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("workflow create request is required")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}

	if len(request.Tasks) == 0 {
		return nil, fmt.Errorf("at least one task is required in the workflow")
	}

	// Validate tasks
	for i, task := range request.Tasks {
		if task.FunctionID == "" {
			return nil, fmt.Errorf("function ID is required for task at index %d", i)
		}
		if task.EndpointID == "" {
			return nil, fmt.Errorf("endpoint ID is required for task at index %d", i)
		}
	}

	var workflow WorkflowResponse
	err := c.doRequest(ctx, http.MethodPost, "workflows", nil, request, &workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (*WorkflowResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	var workflow WorkflowResponse
	err := c.doRequest(ctx, http.MethodGet, "workflows/"+workflowID, nil, nil, &workflow)
	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

// ListWorkflows lists all workflows the user has access to
func (c *Client) ListWorkflows(ctx context.Context) ([]WorkflowResponse, error) {
	var workflows []WorkflowResponse
	err := c.doRequest(ctx, http.MethodGet, "workflows", nil, nil, &workflows)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

// DeleteWorkflow deletes a workflow
func (c *Client) DeleteWorkflow(ctx context.Context, workflowID string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	return c.doRequest(ctx, http.MethodDelete, "workflows/"+workflowID, nil, nil, nil)
}

// RunWorkflow executes a workflow
func (c *Client) RunWorkflow(ctx context.Context, workflowID string, request *WorkflowRunRequest) (*WorkflowRunResponse, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	var response WorkflowRunResponse
	err := c.doRequest(ctx, http.MethodPost, "workflows/"+workflowID+"/run", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetWorkflowStatus gets the status of a workflow run
func (c *Client) GetWorkflowStatus(ctx context.Context, runID string) (*WorkflowStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	var status WorkflowStatusResponse
	err := c.doRequest(ctx, http.MethodGet, "workflows/runs/"+runID, nil, nil, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// CancelWorkflowRun cancels a running workflow
func (c *Client) CancelWorkflowRun(ctx context.Context, runID string) error {
	if runID == "" {
		return fmt.Errorf("run ID is required")
	}

	return c.doRequest(ctx, http.MethodPost, "workflows/runs/"+runID+"/cancel", nil, nil, nil)
}

// CreateTaskGroup creates a new task group for related tasks
func (c *Client) CreateTaskGroup(ctx context.Context, request *TaskGroupCreateRequest) (*TaskGroupResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("task group create request is required")
	}

	if request.Name == "" {
		return nil, fmt.Errorf("task group name is required")
	}

	var taskGroup TaskGroupResponse
	err := c.doRequest(ctx, http.MethodPost, "task_groups", nil, request, &taskGroup)
	if err != nil {
		return nil, err
	}

	return &taskGroup, nil
}

// RunTaskGroup runs all tasks in a task group
func (c *Client) RunTaskGroup(ctx context.Context, taskGroupID string, request *TaskGroupRunRequest) (*TaskGroupRunResponse, error) {
	if taskGroupID == "" {
		return nil, fmt.Errorf("task group ID is required")
	}

	var response TaskGroupRunResponse
	err := c.doRequest(ctx, http.MethodPost, "task_groups/"+taskGroupID+"/run", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTaskGroupStatus gets the status of a task group run
func (c *Client) GetTaskGroupStatus(ctx context.Context, runID string) (*TaskGroupStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	var status TaskGroupStatusResponse
	err := c.doRequest(ctx, http.MethodGet, "task_groups/runs/"+runID, nil, nil, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// WaitForWorkflowCompletion waits for a workflow run to complete with a timeout
func (c *Client) WaitForWorkflowCompletion(ctx context.Context, runID string, timeout time.Duration, pollInterval time.Duration) (*WorkflowStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Default poll interval if not specified
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	// Poll until completion or timeout
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for workflow completion: %w", timeoutCtx.Err())
		default:
			// Check status
			status, err := c.GetWorkflowStatus(ctx, runID)
			if err != nil {
				return nil, err
			}

			// Check if workflow is complete
			if status.Status == "COMPLETED" || status.Status == "FAILED" || status.Status == "CANCELED" {
				return status, nil
			}

			// Wait before polling again
			time.Sleep(pollInterval)
		}
	}
}

// WaitForTaskGroupCompletion waits for a task group run to complete with a timeout
func (c *Client) WaitForTaskGroupCompletion(ctx context.Context, runID string, timeout time.Duration, pollInterval time.Duration) (*TaskGroupStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Default poll interval if not specified
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	// Poll until completion or timeout
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for task group completion: %w", timeoutCtx.Err())
		default:
			// Check status
			status, err := c.GetTaskGroupStatus(ctx, runID)
			if err != nil {
				return nil, err
			}

			// Check if all tasks are complete
			if status.Status == "COMPLETED" || status.Status == "FAILED" || status.Status == "CANCELED" {
				return status, nil
			}

			// Wait before polling again
			time.Sleep(pollInterval)
		}
	}
}

// RunDependencyGraph runs a task dependency graph
func (c *Client) RunDependencyGraph(ctx context.Context, request *DependencyGraphRequest) (*DependencyGraphResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("dependency graph request is required")
	}

	if len(request.Nodes) == 0 {
		return nil, fmt.Errorf("at least one node is required in the dependency graph")
	}

	// Validate nodes
	for id, node := range request.Nodes {
		if node.Task.FunctionID == "" {
			return nil, fmt.Errorf("function ID is required for node %s", id)
		}
		if node.Task.EndpointID == "" {
			return nil, fmt.Errorf("endpoint ID is required for node %s", id)
		}
	}

	var response DependencyGraphResponse
	err := c.doRequest(ctx, http.MethodPost, "dependency_graph/run", nil, request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetDependencyGraphStatus gets the status of a dependency graph run
func (c *Client) GetDependencyGraphStatus(ctx context.Context, runID string) (*DependencyGraphStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	var status DependencyGraphStatusResponse
	err := c.doRequest(ctx, http.MethodGet, "dependency_graph/runs/"+runID, nil, nil, &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// WaitForDependencyGraphCompletion waits for a dependency graph run to complete with a timeout
func (c *Client) WaitForDependencyGraphCompletion(ctx context.Context, runID string, timeout time.Duration, pollInterval time.Duration) (*DependencyGraphStatusResponse, error) {
	if runID == "" {
		return nil, fmt.Errorf("run ID is required")
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Default poll interval if not specified
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	// Poll until completion or timeout
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("timeout waiting for dependency graph completion: %w", timeoutCtx.Err())
		default:
			// Check status
			status, err := c.GetDependencyGraphStatus(ctx, runID)
			if err != nil {
				return nil, err
			}

			// Check if all nodes are complete
			if status.Status == "COMPLETED" || status.Status == "FAILED" || status.Status == "CANCELED" {
				return status, nil
			}

			// Wait before polling again
			time.Sleep(pollInterval)
		}
	}
}
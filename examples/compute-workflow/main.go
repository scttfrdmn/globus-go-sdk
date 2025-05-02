// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
)

func main() {
	// Get access token from environment variable
	accessToken := os.Getenv("GLOBUS_COMPUTE_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("GLOBUS_COMPUTE_ACCESS_TOKEN environment variable is required")
	}

	// Create a compute client
	client := compute.NewClient(accessToken, 
		core.WithLogLevel(core.LogLevelDebug),
		core.WithHTTPTracing(true),
	)

	// Define the context
	ctx := context.Background()

	// Example: Create a workflow
	fmt.Println("Creating workflow...")
	workflow, err := createWorkflow(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}
	fmt.Printf("Created workflow with ID: %s\n", workflow.ID)

	// Run the workflow
	fmt.Println("Running workflow...")
	runResponse, err := runWorkflow(ctx, client, workflow.ID)
	if err != nil {
		log.Fatalf("Failed to run workflow: %v", err)
	}
	fmt.Printf("Workflow running with run ID: %s\n", runResponse.RunID)

	// Monitor workflow status
	fmt.Println("Monitoring workflow status...")
	err = monitorWorkflow(ctx, client, runResponse.RunID)
	if err != nil {
		log.Fatalf("Error monitoring workflow: %v", err)
	}

	// Example: Create a task group
	fmt.Println("\nCreating task group...")
	taskGroup, err := createTaskGroup(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create task group: %v", err)
	}
	fmt.Printf("Created task group with ID: %s\n", taskGroup.ID)

	// Run the task group
	fmt.Println("Running task group...")
	taskGroupRun, err := runTaskGroup(ctx, client, taskGroup.ID)
	if err != nil {
		log.Fatalf("Failed to run task group: %v", err)
	}
	fmt.Printf("Task group running with run ID: %s\n", taskGroupRun.RunID)

	// Monitor task group status
	fmt.Println("Monitoring task group status...")
	err = monitorTaskGroup(ctx, client, taskGroupRun.RunID)
	if err != nil {
		log.Fatalf("Error monitoring task group: %v", err)
	}

	fmt.Println("Example completed successfully!")
}

// createWorkflow creates a sample workflow
func createWorkflow(ctx context.Context, client *compute.Client) (*compute.WorkflowResponse, error) {
	// For a real application, you would need to:
	// 1. Have registered functions already (using client.RegisterFunction)
	// 2. Have access to compute endpoints (list with client.ListEndpoints)

	// This is a placeholder - in a real app, replace with your actual function and endpoint IDs
	functionID := "function-123"
	endpointID := "endpoint-456"

	// Define tasks in the workflow
	tasks := []compute.WorkflowTask{
		{
			ID:         "task1",
			Name:       "First Task",
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"input1", 42},
		},
		{
			ID:         "task2",
			Name:       "Second Task",
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"input2", 84},
		},
		{
			ID:         "task3",
			Name:       "Final Task",
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"final", 100},
		},
	}

	// Define dependencies (task3 depends on task1 and task2)
	dependencies := map[string][]string{
		"task3": {"task1", "task2"},
	}

	// Create workflow request
	request := &compute.WorkflowCreateRequest{
		Name:         "Example Workflow",
		Description:  "A workflow created by the example application",
		Tasks:        tasks,
		Dependencies: dependencies,
		ErrorHandling: "continue",
		RetryPolicy: &compute.RetryPolicy{
			MaxRetries: 3,
			RetryInterval: 5,
		},
		Public: false,
	}

	return client.CreateWorkflow(ctx, request)
}

// runWorkflow runs the workflow and returns the run response
func runWorkflow(ctx context.Context, client *compute.Client, workflowID string) (*compute.WorkflowRunResponse, error) {
	request := &compute.WorkflowRunRequest{
		GlobalArgs: map[string]interface{}{
			"scale_factor": 2.0,
		},
		RunLabel: "Example Run",
	}

	return client.RunWorkflow(ctx, workflowID, request)
}

// monitorWorkflow monitors the workflow until it completes or times out
func monitorWorkflow(ctx context.Context, client *compute.Client, runID string) error {
	// Set a timeout context
	timeout := 5 * time.Minute
	pollInterval := 5 * time.Second
	
	// Using the built-in wait method
	status, err := client.WaitForWorkflowCompletion(ctx, runID, timeout, pollInterval)
	if err != nil {
		return err
	}
	
	fmt.Printf("Workflow completed with status: %s\n", status.Status)
	
	if status.Status == "FAILED" {
		return fmt.Errorf("workflow failed: %s", status.Error)
	}
	
	return nil
}

// createTaskGroup creates a sample task group
func createTaskGroup(ctx context.Context, client *compute.Client) (*compute.TaskGroupResponse, error) {
	// For a real application, you would need to:
	// 1. Have registered functions already (using client.RegisterFunction)
	// 2. Have access to compute endpoints (list with client.ListEndpoints)

	// This is a placeholder - in a real app, replace with your actual function and endpoint IDs
	functionID := "function-123"
	endpointID := "endpoint-456"

	// Define tasks in the task group
	tasks := []compute.TaskRequest{
		{
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"data1", 10},
			Priority:   1,
		},
		{
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"data2", 20},
			Priority:   1,
		},
		{
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{"data3", 30},
			Priority:   1,
		},
	}

	// Create task group request
	request := &compute.TaskGroupCreateRequest{
		Name:        "Example Task Group",
		Description: "A task group created by the example application",
		Tasks:       tasks,
		Concurrency: 2, // Run at most 2 tasks concurrently
		RetryPolicy: &compute.RetryPolicy{
			MaxRetries: 2,
		},
		Public: false,
	}

	return client.CreateTaskGroup(ctx, request)
}

// runTaskGroup runs the task group and returns the run response
func runTaskGroup(ctx context.Context, client *compute.Client, taskGroupID string) (*compute.TaskGroupRunResponse, error) {
	request := &compute.TaskGroupRunRequest{
		RunLabel: "Example Task Group Run",
	}

	return client.RunTaskGroup(ctx, taskGroupID, request)
}

// monitorTaskGroup monitors the task group until it completes or times out
func monitorTaskGroup(ctx context.Context, client *compute.Client, runID string) error {
	// Set a timeout context
	timeout := 5 * time.Minute
	pollInterval := 5 * time.Second
	
	// Using the built-in wait method
	status, err := client.WaitForTaskGroupCompletion(ctx, runID, timeout, pollInterval)
	if err != nil {
		return err
	}
	
	fmt.Printf("Task group completed with status: %s\n", status.Status)
	
	if status.Status == "FAILED" {
		return fmt.Errorf("task group failed: %s", status.Error)
	}
	
	return nil
}
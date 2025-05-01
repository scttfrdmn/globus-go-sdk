// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
)

// Define several functions to use in our workflow examples
const dataProcessorFunction = `
def process_data(data):
    """
    Process input data and extract key information
    
    Args:
        data: Dictionary of input data
        
    Returns:
        Processed results
    """
    import time
    
    # Simulate processing time
    time.sleep(2)
    
    result = {
        "processed_at": time.time(),
        "input_size": len(str(data)),
        "extracted_values": {},
    }
    
    # Extract values based on data type
    if isinstance(data, dict):
        for key, value in data.items():
            if isinstance(value, (int, float, str, bool)):
                result["extracted_values"][key] = value
            elif isinstance(value, (list, tuple)):
                result["extracted_values"][key] = len(value)
            elif isinstance(value, dict):
                result["extracted_values"][key] = list(value.keys())
    
    return result
`

const dataAggregatorFunction = `
def aggregate_results(results):
    """
    Combine and aggregate results from multiple processors
    
    Args:
        results: List of result dictionaries
        
    Returns:
        Aggregated results
    """
    import time
    
    # Simulate aggregation time
    time.sleep(1)
    
    if not results or not isinstance(results, list):
        return {"error": "Invalid input for aggregation"}
    
    aggregated = {
        "aggregated_at": time.time(),
        "total_inputs": len(results),
        "combined_values": {},
        "timestamps": [],
    }
    
    # Collect all timestamps
    for result in results:
        if isinstance(result, dict) and "processed_at" in result:
            aggregated["timestamps"].append(result["processed_at"])
    
    # Combine extracted values
    for result in results:
        if isinstance(result, dict) and "extracted_values" in result:
            for key, value in result["extracted_values"].items():
                if key in aggregated["combined_values"]:
                    if isinstance(aggregated["combined_values"][key], list):
                        if isinstance(value, list):
                            aggregated["combined_values"][key].extend(value)
                        else:
                            aggregated["combined_values"][key].append(value)
                    elif isinstance(value, list):
                        aggregated["combined_values"][key] = [aggregated["combined_values"][key]] + value
                    else:
                        aggregated["combined_values"][key] = [aggregated["combined_values"][key], value]
                else:
                    aggregated["combined_values"][key] = value
    
    return aggregated
`

const reportGeneratorFunction = `
def generate_report(aggregated_data):
    """
    Generate a final report from aggregated data
    
    Args:
        aggregated_data: Dictionary of aggregated results
        
    Returns:
        Report as a formatted dictionary
    """
    import time
    import statistics
    
    # Simulate report generation time
    time.sleep(1.5)
    
    if not isinstance(aggregated_data, dict):
        return {"error": "Invalid aggregated data for report generation"}
    
    report = {
        "report_generated_at": time.time(),
        "summary": {
            "total_processed_inputs": aggregated_data.get("total_inputs", 0),
            "pipeline_duration": 0,
            "unique_keys": 0,
        },
        "details": {},
        "metrics": {},
    }
    
    # Calculate pipeline duration
    timestamps = aggregated_data.get("timestamps", [])
    if timestamps and len(timestamps) > 0:
        earliest = min(timestamps)
        latest = max(timestamps)
        report["summary"]["pipeline_duration"] = latest - earliest
    
    # Process combined values
    combined_values = aggregated_data.get("combined_values", {})
    report["summary"]["unique_keys"] = len(combined_values)
    
    # Extract details and metrics
    for key, value in combined_values.items():
        report["details"][key] = value
        
        # Calculate metrics for numeric lists
        if isinstance(value, list) and all(isinstance(x, (int, float)) for x in value):
            report["metrics"][key] = {
                "min": min(value),
                "max": max(value),
                "mean": statistics.mean(value) if value else 0,
                "count": len(value),
            }
    
    return report
`

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create a new Auth client
	authClient := config.NewAuthClient()

	// Get token using client credentials
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.ComputeScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)
	accessToken := tokenResp.AccessToken

	// Create Compute client
	computeClient := config.NewComputeClient(accessToken)

	// List available endpoints
	fmt.Println("\n=== Available Compute Endpoints ===")
	endpoints, err := computeClient.ListEndpoints(ctx, &pkg.ListEndpointsOptions{
		PerPage: 5,
	})
	if err != nil {
		log.Fatalf("Failed to list endpoints: %v", err)
	}

	if len(endpoints.Endpoints) == 0 {
		log.Fatalf("No compute endpoints found. Please create an endpoint first.")
	}

	fmt.Printf("Found %d compute endpoints:\n", len(endpoints.Endpoints))
	for i, endpoint := range endpoints.Endpoints {
		fmt.Printf("%d. %s (%s)\n", i+1, endpoint.Name, endpoint.ID)
		fmt.Printf("   Status: %s, Connected: %t\n", endpoint.Status, endpoint.Connected)
	}

	// Select the first endpoint
	selectedEndpoint := endpoints.Endpoints[0]
	fmt.Printf("\nUsing endpoint: %s (%s)\n", selectedEndpoint.Name, selectedEndpoint.ID)

	// Create a timestamp for unique naming
	timestamp := time.Now().Format("20060102_150405")

	// Register functions for our examples
	fmt.Println("\n=== Registering Functions ===")
	
	// Register data processor function
	procFuncName := fmt.Sprintf("data_processor_%s", timestamp)
	processorReq := &pkg.FunctionRegisterRequest{
		Function:    dataProcessorFunction,
		Name:        procFuncName,
		Description: "Function to process data chunks",
	}
	procFunc, err := computeClient.RegisterFunction(ctx, processorReq)
	if err != nil {
		log.Fatalf("Failed to register processor function: %v", err)
	}
	fmt.Printf("Registered processor function: %s (%s)\n", procFunc.Name, procFunc.ID)

	// Register aggregator function
	aggFuncName := fmt.Sprintf("data_aggregator_%s", timestamp)
	aggregatorReq := &pkg.FunctionRegisterRequest{
		Function:    dataAggregatorFunction,
		Name:        aggFuncName,
		Description: "Function to aggregate processed results",
	}
	aggFunc, err := computeClient.RegisterFunction(ctx, aggregatorReq)
	if err != nil {
		log.Fatalf("Failed to register aggregator function: %v", err)
	}
	fmt.Printf("Registered aggregator function: %s (%s)\n", aggFunc.Name, aggFunc.ID)

	// Register report generator function
	reportFuncName := fmt.Sprintf("report_generator_%s", timestamp)
	reportReq := &pkg.FunctionRegisterRequest{
		Function:    reportGeneratorFunction,
		Name:        reportFuncName,
		Description: "Function to generate final report",
	}
	reportFunc, err := computeClient.RegisterFunction(ctx, reportReq)
	if err != nil {
		log.Fatalf("Failed to register report function: %v", err)
	}
	fmt.Printf("Registered report function: %s (%s)\n", reportFunc.Name, reportFunc.ID)

	// Clean up resources at the end
	defer func() {
		fmt.Println("\n=== Cleaning Up Resources ===")
		
		// Clean up functions
		for _, funcID := range []string{procFunc.ID, aggFunc.ID, reportFunc.ID} {
			if err := computeClient.DeleteFunction(ctx, funcID); err != nil {
				log.Printf("Warning: Failed to delete function %s: %v", funcID, err)
			} else {
				fmt.Printf("Function %s deleted successfully\n", funcID)
			}
		}
		
		// Clean up other resources created in examples
		if taskGroupID != "" {
			if err := computeClient.DeleteTaskGroup(ctx, taskGroupID); err != nil {
				log.Printf("Warning: Failed to delete task group %s: %v", taskGroupID, err)
			} else {
				fmt.Printf("Task group %s deleted successfully\n", taskGroupID)
			}
		}
		
		if workflowID != "" {
			if err := computeClient.DeleteWorkflow(ctx, workflowID); err != nil {
				log.Printf("Warning: Failed to delete workflow %s: %v", workflowID, err)
			} else {
				fmt.Printf("Workflow %s deleted successfully\n", workflowID)
			}
		}
	}()

	// Example 1: Task Group Execution
	fmt.Println("\n=== EXAMPLE 1: Task Group Execution ===")
	taskGroupID, err := demonstrateTaskGroupExecution(ctx, computeClient, selectedEndpoint.ID, procFunc.ID)
	if err != nil {
		log.Printf("Task group execution example failed: %v", err)
	}
	
	// Example 2: Workflow with Dependencies
	fmt.Println("\n=== EXAMPLE 2: Workflow with Dependencies ===")
	workflowID, err := demonstrateWorkflow(ctx, computeClient, selectedEndpoint.ID, procFunc.ID, aggFunc.ID, reportFunc.ID)
	if err != nil {
		log.Printf("Workflow example failed: %v", err)
	}
	
	// Example 3: Dependency Graph Execution
	fmt.Println("\n=== EXAMPLE 3: Dependency Graph Execution ===")
	err = demonstrateDependencyGraph(ctx, computeClient, selectedEndpoint.ID, procFunc.ID, aggFunc.ID, reportFunc.ID)
	if err != nil {
		log.Printf("Dependency graph example failed: %v", err)
	}
	
	fmt.Println("\nBatch execution examples complete!")
}

// Global variables to track resources for cleanup
var (
	taskGroupID string
	workflowID  string
)

// demonstrateTaskGroupExecution shows how to execute a group of similar tasks concurrently
func demonstrateTaskGroupExecution(ctx context.Context, client *pkg.Client, endpointID, functionID string) (string, error) {
	fmt.Println("Creating a task group for parallel data processing...")
	
	// Create sample data chunks
	dataChunks := []map[string]interface{}{
		{
			"chunk_id": 1,
			"values": []int{10, 20, 30, 40, 50},
			"metadata": {
				"source": "sensor-1",
				"timestamp": time.Now().Unix(),
			},
		},
		{
			"chunk_id": 2,
			"values": []int{15, 25, 35, 45, 55},
			"metadata": {
				"source": "sensor-2",
				"timestamp": time.Now().Unix(),
			},
		},
		{
			"chunk_id": 3,
			"values": []int{5, 15, 25, 35, 45},
			"metadata": {
				"source": "sensor-3",
				"timestamp": time.Now().Unix(),
			},
		},
		{
			"chunk_id": 4,
			"values": []int{12, 24, 36, 48, 60},
			"metadata": {
				"source": "sensor-4",
				"timestamp": time.Now().Unix(),
			},
		},
		{
			"chunk_id": 5,
			"values": []int{8, 16, 24, 32, 40},
			"metadata": {
				"source": "sensor-5",
				"timestamp": time.Now().Unix(),
			},
		},
	}
	
	// Create tasks for each data chunk
	tasks := make([]pkg.TaskRequest, len(dataChunks))
	for i, chunk := range dataChunks {
		tasks[i] = pkg.TaskRequest{
			FunctionID: functionID,
			EndpointID: endpointID,
			Args:       []interface{}{chunk},
		}
	}
	
	// Create the task group
	taskGroupReq := &pkg.TaskGroupCreateRequest{
		Name:        "parallel_data_processing",
		Description: "Process multiple data chunks in parallel",
		Tasks:       tasks,
		Concurrency: 3, // Process up to 3 tasks at once
		RetryPolicy: &pkg.RetryPolicy{
			MaxRetries: 2,
			RetryInterval: 5,
		},
	}
	
	taskGroup, err := client.CreateTaskGroup(ctx, taskGroupReq)
	if err != nil {
		return "", fmt.Errorf("failed to create task group: %w", err)
	}
	
	fmt.Printf("Created task group: %s (%s)\n", taskGroup.Name, taskGroup.ID)
	fmt.Printf("Task group contains %d tasks with concurrency limit of %d\n", 
		len(taskGroup.Tasks), taskGroup.Concurrency)
	
	// Store for cleanup
	taskGroupID = taskGroup.ID
	
	// Run the task group
	fmt.Println("\nRunning the task group...")
	runReq := &pkg.TaskGroupRunRequest{
		Priority:    2,
		Description: "Batch processing example run",
		RunLabel:    "example-run",
	}
	
	runResp, err := client.RunTaskGroup(ctx, taskGroup.ID, runReq)
	if err != nil {
		return taskGroup.ID, fmt.Errorf("failed to run task group: %w", err)
	}
	
	fmt.Printf("Task group started with run ID: %s\n", runResp.RunID)
	fmt.Printf("Started %d tasks\n", len(runResp.TaskIDs))
	
	// Wait for the task group to complete
	fmt.Println("\nWaiting for task group to complete...")
	status, err := client.WaitForTaskGroupCompletion(ctx, runResp.RunID, 30*time.Second, 2*time.Second)
	if err != nil {
		return taskGroup.ID, fmt.Errorf("error waiting for task group completion: %w", err)
	}
	
	fmt.Printf("Task group %s! Progress: %d/%d tasks completed\n", 
		status.Status, status.Progress.Completed, status.Progress.TotalTasks)
	
	// Print task results
	fmt.Println("\nTask results:")
	for taskID, taskStatus := range status.TaskStatus {
		if taskStatus.Status == "COMPLETED" {
			fmt.Printf("Task %s: Completed successfully\n", taskID)
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Task %s: Failed with error: %s\n", taskID, taskStatus.Error)
		} else {
			fmt.Printf("Task %s: Status=%s\n", taskID, taskStatus.Status)
		}
	}
	
	return taskGroup.ID, nil
}

// demonstrateWorkflow shows how to create and run a workflow with dependencies
func demonstrateWorkflow(ctx context.Context, client *pkg.Client, endpointID, processorID, aggregatorID, reportID string) (string, error) {
	fmt.Println("Creating a workflow with dependencies...")
	
	// Create a workflow with three stages:
	// 1. Three parallel data processing tasks
	// 2. One aggregation task that depends on the processing tasks
	// 3. One report generation task that depends on the aggregation task
	
	// Define the workflow
	workflowReq := &pkg.WorkflowCreateRequest{
		Name:        "data_processing_pipeline",
		Description: "A three-stage data processing pipeline",
		Tasks: []pkg.WorkflowTask{
			{
				ID:         "process1",
				Name:       "Process Data 1",
				FunctionID: processorID,
				EndpointID: endpointID,
				Args: []interface{}{
					map[string]interface{}{
						"chunk_id": 1,
						"values":   []int{10, 20, 30, 40, 50},
						"metadata": map[string]interface{}{"source": "dataset-1"},
					},
				},
			},
			{
				ID:         "process2",
				Name:       "Process Data 2",
				FunctionID: processorID,
				EndpointID: endpointID,
				Args: []interface{}{
					map[string]interface{}{
						"chunk_id": 2,
						"values":   []int{15, 25, 35, 45, 55},
						"metadata": map[string]interface{}{"source": "dataset-2"},
					},
				},
			},
			{
				ID:         "process3",
				Name:       "Process Data 3",
				FunctionID: processorID,
				EndpointID: endpointID,
				Args: []interface{}{
					map[string]interface{}{
						"chunk_id": 3,
						"values":   []int{5, 15, 25, 35, 45},
						"metadata": map[string]interface{}{"source": "dataset-3"},
					},
				},
			},
			{
				ID:         "aggregate",
				Name:       "Aggregate Results",
				FunctionID: aggregatorID,
				EndpointID: endpointID,
				// Args will be supplied dynamically from previous tasks
			},
			{
				ID:         "report",
				Name:       "Generate Report",
				FunctionID: reportID,
				EndpointID: endpointID,
				// Args will be supplied from the aggregate task
			},
		},
		Dependencies: map[string][]string{
			"aggregate": {"process1", "process2", "process3"},
			"report":    {"aggregate"},
		},
		ErrorHandling: "continue",
		RetryPolicy: &pkg.RetryPolicy{
			MaxRetries: 2,
		},
	}
	
	workflow, err := client.CreateWorkflow(ctx, workflowReq)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}
	
	fmt.Printf("Created workflow: %s (%s)\n", workflow.Name, workflow.ID)
	fmt.Printf("Workflow contains %d tasks with dependencies\n", len(workflow.Tasks))
	
	// Store for cleanup
	workflowID = workflow.ID
	
	// Run the workflow
	fmt.Println("\nRunning the workflow...")
	runReq := &pkg.WorkflowRunRequest{
		Priority:    2,
		Description: "Data processing pipeline execution",
		RunLabel:    "pipeline-run",
		GlobalArgs: map[string]interface{}{
			"debug": true,
		},
		// Supply arguments to the aggregate task based on process task outputs
		TaskArgs: map[string]map[string]interface{}{
			"aggregate": {
				"$ARGS": []string{"$OUTPUT.process1", "$OUTPUT.process2", "$OUTPUT.process3"},
			},
			"report": {
				"$ARGS": []string{"$OUTPUT.aggregate"},
			},
		},
	}
	
	runResp, err := client.RunWorkflow(ctx, workflow.ID, runReq)
	if err != nil {
		return workflow.ID, fmt.Errorf("failed to run workflow: %w", err)
	}
	
	fmt.Printf("Workflow started with run ID: %s\n", runResp.RunID)
	
	// Wait for the workflow to complete
	fmt.Println("\nWaiting for workflow to complete...")
	status, err := client.WaitForWorkflowCompletion(ctx, runResp.RunID, 60*time.Second, 2*time.Second)
	if err != nil {
		return workflow.ID, fmt.Errorf("error waiting for workflow completion: %w", err)
	}
	
	fmt.Printf("Workflow %s! Progress: %d/%d tasks completed\n", 
		status.Status, status.Progress.Completed, status.Progress.TotalTasks)
	
	// Print task results
	fmt.Println("\nWorkflow task results:")
	for taskID, taskStatus := range status.TaskStatus {
		if taskStatus.Status == "COMPLETED" {
			fmt.Printf("Task %s: Completed successfully\n", taskID)
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Task %s: Failed with error: %s\n", taskID, taskStatus.Error)
		} else {
			fmt.Printf("Task %s: Status=%s\n", taskID, taskStatus.Status)
		}
	}
	
	// If the final report task completed, print a summary
	if reportStatus, ok := status.TaskStatus["report"]; ok && reportStatus.Status == "COMPLETED" {
		if result, ok := reportStatus.Result.(map[string]interface{}); ok {
			if summary, ok := result["summary"].(map[string]interface{}); ok {
				fmt.Println("\nReport Summary:")
				fmt.Printf("Total Processed Inputs: %v\n", summary["total_processed_inputs"])
				fmt.Printf("Pipeline Duration: %v seconds\n", summary["pipeline_duration"])
				fmt.Printf("Unique Keys: %v\n", summary["unique_keys"])
			}
		}
	}
	
	return workflow.ID, nil
}

// demonstrateDependencyGraph shows how to create and run a dependency graph with dynamic execution
func demonstrateDependencyGraph(ctx context.Context, client *pkg.Client, endpointID, processorID, aggregatorID, reportID string) error {
	fmt.Println("Creating a dependency graph for dynamic execution...")
	
	// Create test data with different shapes
	testData := []map[string]interface{}{
		{
			"id": "data1",
			"values": []int{10, 20, 30},
			"metadata": map[string]string{"type": "integers"},
		},
		{
			"id": "data2",
			"values": []string{"a", "b", "c"},
			"metadata": map[string]string{"type": "strings"},
		},
		{
			"id": "data3",
			"values": []float64{1.1, 2.2, 3.3},
			"metadata": map[string]string{"type": "floats"},
		},
		{
			"id": "data4",
			"values": []bool{true, false, true},
			"metadata": map[string]string{"type": "booleans"},
		},
	}
	
	// Define the dependency graph nodes
	nodes := make(map[string]pkg.DependencyGraphNode)
	
	// Add processing nodes
	for i, data := range testData {
		nodeID := fmt.Sprintf("process_%d", i+1)
		nodes[nodeID] = pkg.DependencyGraphNode{
			Task: pkg.TaskRequest{
				FunctionID: processorID,
				EndpointID: endpointID,
				Args:       []interface{}{data},
			},
			RetryPolicy: &pkg.RetryPolicy{
				MaxRetries: 1,
			},
		}
	}
	
	// Add aggregation node
	nodes["aggregate"] = pkg.DependencyGraphNode{
		Task: pkg.TaskRequest{
			FunctionID: aggregatorID,
			EndpointID: endpointID,
			// Args will be created dynamically from processing results
		},
		Dependencies: []string{"process_1", "process_2", "process_3", "process_4"},
		Condition:    "ALL_COMPLETED", // Wait for all dependencies
		ErrorHandler: &pkg.ErrorHandler{
			Strategy: "RETRY",
			RetryPolicy: &pkg.RetryPolicy{
				MaxRetries: 2,
			},
		},
	}
	
	// Add report node
	nodes["report"] = pkg.DependencyGraphNode{
		Task: pkg.TaskRequest{
			FunctionID: reportID,
			EndpointID: endpointID,
			// Args will be created from aggregation result
		},
		Dependencies: []string{"aggregate"},
		ErrorHandler: &pkg.ErrorHandler{
			Strategy: "FAIL_WORKFLOW", // If report fails, fail the whole workflow
		},
	}
	
	// Create the request
	graphReq := &pkg.DependencyGraphRequest{
		Nodes:       nodes,
		Description: "Dynamic data processing graph",
		ErrorPolicy: "FAIL_FAST", // Fail immediately on any error
	}
	
	// Run the dependency graph
	fmt.Println("Running the dependency graph...")
	resp, err := client.RunDependencyGraph(ctx, graphReq)
	if err != nil {
		return fmt.Errorf("failed to run dependency graph: %w", err)
	}
	
	fmt.Printf("Dependency graph started with run ID: %s\n", resp.RunID)
	
	// Wait for the graph execution to complete
	fmt.Println("\nWaiting for dependency graph to complete...")
	status, err := client.WaitForDependencyGraphCompletion(ctx, resp.RunID, 60*time.Second, 2*time.Second)
	if err != nil {
		return fmt.Errorf("error waiting for dependency graph completion: %w", err)
	}
	
	fmt.Printf("Dependency graph %s! Progress: %d/%d nodes completed\n", 
		status.Status, status.Progress.Completed, status.Progress.TotalNodes)
	
	// Print node results
	fmt.Println("\nDependency graph node results:")
	for nodeID, nodeStatus := range status.NodeStatus {
		if nodeStatus.Status == "COMPLETED" {
			fmt.Printf("Node %s: Completed successfully\n", nodeID)
		} else if nodeStatus.Status == "FAILED" {
			fmt.Printf("Node %s: Failed with error: %s\n", nodeID, nodeStatus.Error)
		} else {
			fmt.Printf("Node %s: Status=%s\n", nodeID, nodeStatus.Status)
		}
	}
	
	return nil
}
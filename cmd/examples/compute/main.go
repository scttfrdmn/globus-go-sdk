// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
)

// Define a simple function to register and run
const sampleFunction = `def hello(name="World"):
    import datetime
    now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    return f"Hello, {name}! The time is {now}"
`

const advancedFunction = `def process_data(data):
    """
    Process some input data and return results
    
    Args:
        data: A dictionary containing input values
        
    Returns:
        A dictionary with processed results
    """
    result = {}
    
    # Calculate sum of values
    if "values" in data and isinstance(data["values"], list):
        result["sum"] = sum(data["values"])
        result["average"] = sum(data["values"]) / len(data["values"])
        result["min"] = min(data["values"])
        result["max"] = max(data["values"])
        result["count"] = len(data["values"])
    
    # Process text
    if "text" in data and isinstance(data["text"], str):
        result["text_length"] = len(data["text"])
        result["words"] = len(data["text"].split())
        result["uppercase"] = data["text"].upper()
    
    # Add timestamp
    import datetime
    result["timestamp"] = datetime.datetime.now().isoformat()
    
    return result
`

func main() {
	// Create a new SDK configuration
	config := pkg.NewConfigFromEnvironment().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

	// Create a new Auth client
	authClient, err := config.NewAuthClient()
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}

	// Get token using client credentials for simplicity
	// In a real application, you would likely use the authorization code flow
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.ComputeScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)
	accessToken := tokenResp.AccessToken

	// Create Compute client
	computeClient, err := config.NewComputeClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}

	// List available endpoints
	fmt.Println("\n=== Available Compute Endpoints ===")
	endpoints, err := computeClient.ListEndpoints(ctx, nil)
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

	// Register a simple function
	fmt.Println("\n=== Registering Function ===")
	timestamp := time.Now().Format("20060102_150405")
	functionName := fmt.Sprintf("example_function_%s", timestamp)

	// Assuming this type is from the compute package
	registerRequest := &compute.FunctionRegisterRequest{
		Function:    sampleFunction,
		Name:        functionName,
		Description: "A simple greeting function created by the Globus Go SDK",
	}

	function, err := computeClient.RegisterFunction(ctx, registerRequest)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Printf("Function registered: %s (%s)\n", function.Name, function.ID)

	// Register an advanced function for later use
	advancedFunctionName := fmt.Sprintf("advanced_function_%s", timestamp)
	advancedRegisterRequest := &compute.FunctionRegisterRequest{
		Function:    advancedFunction,
		Name:        advancedFunctionName,
		Description: "An advanced data processing function created by the Globus Go SDK",
	}

	advancedFunc, err := computeClient.RegisterFunction(ctx, advancedRegisterRequest)
	if err != nil {
		log.Printf("Failed to register advanced function: %v", err)
	} else {
		fmt.Printf("Advanced function registered: %s (%s)\n", advancedFunc.Name, advancedFunc.ID)
	}

	// Make sure to clean up the functions after the example
	defer func() {
		fmt.Println("\n=== Cleaning Up Functions ===")
		if err := computeClient.DeleteFunction(ctx, function.ID); err != nil {
			log.Printf("Warning: Failed to delete function %s: %v", function.ID, err)
		} else {
			fmt.Printf("Function %s deleted successfully\n", function.ID)
		}

		if advancedFunc != nil {
			if err := computeClient.DeleteFunction(ctx, advancedFunc.ID); err != nil {
				log.Printf("Warning: Failed to delete function %s: %v", advancedFunc.ID, err)
			} else {
				fmt.Printf("Function %s deleted successfully\n", advancedFunc.ID)
			}
		}
	}()

	// Execute the simple function
	fmt.Println("\n=== Running Simple Function ===")
	taskRequest := &compute.TaskRequest{
		FunctionID: function.ID,
		EndpointID: selectedEndpoint.ID,
		Args:       []interface{}{"Globus Go SDK"},
	}

	task, err := computeClient.RunFunction(ctx, taskRequest)
	if err != nil {
		log.Fatalf("Failed to run function: %v", err)
	}

	fmt.Printf("Task submitted: %s (Status: %s)\n", task.TaskID, task.Status)

	// Execute the advanced function if available
	var advancedTask *compute.TaskResponse
	if advancedFunc != nil {
		fmt.Println("\n=== Running Advanced Function ===")

		// Prepare sample data
		sampleData := map[string]interface{}{
			"values": []int{10, 20, 30, 40, 50},
			"text":   "This is a sample text for Globus Compute processing",
		}

		advTaskRequest := &compute.TaskRequest{
			FunctionID: advancedFunc.ID,
			EndpointID: selectedEndpoint.ID,
			Args:       []interface{}{sampleData},
		}

		var err error
		advancedTask, err = computeClient.RunFunction(ctx, advTaskRequest)
		if err != nil {
			log.Printf("Failed to run advanced function: %v", err)
		} else {
			fmt.Printf("Advanced task submitted: %s (Status: %s)\n", advancedTask.TaskID, advancedTask.Status)
		}
	}

	// Wait for tasks to complete and get results
	fmt.Println("\nWaiting for tasks to complete...")
	time.Sleep(3 * time.Second)

	// Get simple task status
	fmt.Println("\n=== Simple Task Results ===")
	taskStatus, err := computeClient.GetTaskStatus(ctx, task.TaskID)
	if err != nil {
		log.Printf("Failed to get task status: %v", err)
	} else {
		fmt.Printf("Task ID: %s\n", taskStatus.TaskID)
		fmt.Printf("Status: %s\n", taskStatus.Status)

		if taskStatus.Status == "SUCCESS" {
			fmt.Printf("Result: %v\n", taskStatus.Result)
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Exception: %s\n", taskStatus.Exception)
		} else {
			fmt.Println("Task is still running or in another state")
		}
	}

	// Get advanced task status if available
	if advancedTask != nil {
		fmt.Println("\n=== Advanced Task Results ===")
		advTaskStatus, err := computeClient.GetTaskStatus(ctx, advancedTask.TaskID)
		if err != nil {
			log.Printf("Failed to get advanced task status: %v", err)
		} else {
			fmt.Printf("Task ID: %s\n", advTaskStatus.TaskID)
			fmt.Printf("Status: %s\n", advTaskStatus.Status)

			if advTaskStatus.Status == "SUCCESS" {
				// Pretty print the result
				resultJSON, err := json.MarshalIndent(advTaskStatus.Result, "", "  ")
				if err != nil {
					fmt.Printf("Result: %v\n", advTaskStatus.Result)
				} else {
					fmt.Printf("Result:\n%s\n", resultJSON)
				}
			} else if advTaskStatus.Status == "FAILED" {
				fmt.Printf("Exception: %s\n", advTaskStatus.Exception)
			} else {
				fmt.Println("Task is still running or in another state")
			}
		}
	}

	// Execute a batch of tasks
	fmt.Println("\n=== Running Batch of Tasks ===")
	batchRequest := &compute.BatchTaskRequest{
		Tasks: []compute.TaskRequest{
			{
				FunctionID: function.ID,
				EndpointID: selectedEndpoint.ID,
				Args:       []interface{}{"First Batch Task"},
			},
			{
				FunctionID: function.ID,
				EndpointID: selectedEndpoint.ID,
				Args:       []interface{}{"Second Batch Task"},
			},
		},
	}

	batchResp, err := computeClient.RunBatch(ctx, batchRequest)
	if err != nil {
		log.Printf("Failed to run batch: %v", err)
	} else {
		fmt.Printf("Batch submitted with %d tasks\n", len(batchResp.TaskIDs))

		// Wait for batch to complete
		time.Sleep(3 * time.Second)

		// Get batch status
		batchStatus, err := computeClient.GetBatchStatus(ctx, batchResp.TaskIDs)
		if err != nil {
			log.Printf("Failed to get batch status: %v", err)
		} else {
			fmt.Printf("\n=== Batch Results ===\n")
			fmt.Printf("Completed: %d, Pending: %d, Failed: %d\n",
				len(batchStatus.Completed), len(batchStatus.Pending), len(batchStatus.Failed))

			// Print results for each task
			for i, taskID := range batchResp.TaskIDs {
				status, ok := batchStatus.Tasks[taskID]
				if !ok {
					fmt.Printf("Task %d (%s): Status not available\n", i+1, taskID)
					continue
				}

				fmt.Printf("Task %d (%s): Status = %s\n", i+1, taskID, status.Status)
				if status.Status == "SUCCESS" {
					fmt.Printf("  Result: %v\n", status.Result)
				} else if status.Status == "FAILED" {
					fmt.Printf("  Exception: %s\n", status.Exception)
				}
			}
		}
	}

	fmt.Println("\nCompute example complete!")
}

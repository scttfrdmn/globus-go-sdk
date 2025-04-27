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
	authClient := config.NewAuthClient()

	// Get token using client credentials for simplicity
	// In a real application, you would likely use the authorization code flow
	ctx := context.Background()
	tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.ComputeScope)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Printf("Obtained access token (expires in %d seconds)
", tokenResp.ExpiresIn)
	accessToken := tokenResp.AccessToken

	// Create Compute client
	computeClient := config.NewComputeClient(accessToken)

	// List available endpoints
	fmt.Println("
=== Available Compute Endpoints ===")
	endpoints, err := computeClient.ListEndpoints(ctx, &pkg.ListEndpointsOptions{
		PerPage: 5,
	})
	if err != nil {
		log.Fatalf("Failed to list endpoints: %v", err)
	}

	if len(endpoints.Endpoints) == 0 {
		log.Fatalf("No compute endpoints found. Please create an endpoint first.")
	}

	fmt.Printf("Found %d compute endpoints:
", len(endpoints.Endpoints))
	for i, endpoint := range endpoints.Endpoints {
		fmt.Printf("%d. %s (%s)
", i+1, endpoint.Name, endpoint.ID)
		fmt.Printf("   Status: %s, Connected: %t
", endpoint.Status, endpoint.Connected)
	}

	// Select the first endpoint
	selectedEndpoint := endpoints.Endpoints[0]
	fmt.Printf("
Using endpoint: %s (%s)
", selectedEndpoint.Name, selectedEndpoint.ID)

	// Register a simple function
	fmt.Println("
=== Registering Function ===")
	timestamp := time.Now().Format("20060102_150405")
	functionName := fmt.Sprintf("example_function_%s", timestamp)

	registerRequest := &pkg.FunctionRegisterRequest{
		Function:    sampleFunction,
		Name:        functionName,
		Description: "A simple greeting function created by the Globus Go SDK",
	}

	function, err := computeClient.RegisterFunction(ctx, registerRequest)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Printf("Function registered: %s (%s)
", function.Name, function.ID)

	// Register an advanced function for later use
	advancedFunctionName := fmt.Sprintf("advanced_function_%s", timestamp)
	advancedRegisterRequest := &pkg.FunctionRegisterRequest{
		Function:    advancedFunction,
		Name:        advancedFunctionName,
		Description: "An advanced data processing function created by the Globus Go SDK",
	}

	advancedFunc, err := computeClient.RegisterFunction(ctx, advancedRegisterRequest)
	if err != nil {
		log.Printf("Failed to register advanced function: %v", err)
	} else {
		fmt.Printf("Advanced function registered: %s (%s)
", advancedFunc.Name, advancedFunc.ID)
	}

	// Make sure to clean up the functions after the example
	defer func() {
		fmt.Println("
=== Cleaning Up Functions ===")
		if err := computeClient.DeleteFunction(ctx, function.ID); err != nil {
			log.Printf("Warning: Failed to delete function %s: %v", function.ID, err)
		} else {
			fmt.Printf("Function %s deleted successfully
", function.ID)
		}

		if advancedFunc != nil {
			if err := computeClient.DeleteFunction(ctx, advancedFunc.ID); err != nil {
				log.Printf("Warning: Failed to delete function %s: %v", advancedFunc.ID, err)
			} else {
				fmt.Printf("Function %s deleted successfully
", advancedFunc.ID)
			}
		}
	}()

	// Execute the simple function
	fmt.Println("
=== Running Simple Function ===")
	taskRequest := &pkg.TaskRequest{
		FunctionID: function.ID,
		EndpointID: selectedEndpoint.ID,
		Args:       []interface{}{"Globus Go SDK"},
	}

	task, err := computeClient.RunFunction(ctx, taskRequest)
	if err != nil {
		log.Fatalf("Failed to run function: %v", err)
	}

	fmt.Printf("Task submitted: %s (Status: %s)
", task.TaskID, task.Status)

	// Execute the advanced function if available
	var advancedTask *pkg.TaskResponse
	if advancedFunc != nil {
		fmt.Println("
=== Running Advanced Function ===")
		
		// Prepare sample data
		sampleData := map[string]interface{}{
			"values": []int{10, 20, 30, 40, 50},
			"text":   "This is a sample text for Globus Compute processing",
		}
		
		advTaskRequest := &pkg.TaskRequest{
			FunctionID: advancedFunc.ID,
			EndpointID: selectedEndpoint.ID,
			Args:       []interface{}{sampleData},
		}
		
		var err error
		advancedTask, err = computeClient.RunFunction(ctx, advTaskRequest)
		if err != nil {
			log.Printf("Failed to run advanced function: %v", err)
		} else {
			fmt.Printf("Advanced task submitted: %s (Status: %s)
", advancedTask.TaskID, advancedTask.Status)
		}
	}

	// Wait for tasks to complete and get results
	fmt.Println("
Waiting for tasks to complete...")
	time.Sleep(3 * time.Second)

	// Get simple task status
	fmt.Println("
=== Simple Task Results ===")
	taskStatus, err := computeClient.GetTaskStatus(ctx, task.TaskID)
	if err != nil {
		log.Printf("Failed to get task status: %v", err)
	} else {
		fmt.Printf("Task ID: %s
", taskStatus.TaskID)
		fmt.Printf("Status: %s
", taskStatus.Status)
		
		if taskStatus.Status == "SUCCESS" {
			fmt.Printf("Result: %v
", taskStatus.Result)
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Exception: %s
", taskStatus.Exception)
		} else {
			fmt.Println("Task is still running or in another state")
		}
	}

	// Get advanced task status if available
	if advancedTask != nil {
		fmt.Println("
=== Advanced Task Results ===")
		advTaskStatus, err := computeClient.GetTaskStatus(ctx, advancedTask.TaskID)
		if err != nil {
			log.Printf("Failed to get advanced task status: %v", err)
		} else {
			fmt.Printf("Task ID: %s
", advTaskStatus.TaskID)
			fmt.Printf("Status: %s
", advTaskStatus.Status)
			
			if advTaskStatus.Status == "SUCCESS" {
				// Pretty print the result
				resultJSON, err := json.MarshalIndent(advTaskStatus.Result, "", "  ")
				if err != nil {
					fmt.Printf("Result: %v
", advTaskStatus.Result)
				} else {
					fmt.Printf("Result:
%s
", resultJSON)
				}
			} else if advTaskStatus.Status == "FAILED" {
				fmt.Printf("Exception: %s
", advTaskStatus.Exception)
			} else {
				fmt.Println("Task is still running or in another state")
			}
		}
	}

	// Execute a batch of tasks
	fmt.Println("
=== Running Batch of Tasks ===")
	batchRequest := &pkg.BatchTaskRequest{
		Tasks: []pkg.TaskRequest{
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
		fmt.Printf("Batch submitted with %d tasks
", len(batchResp.TaskIDs))
		
		// Wait for batch to complete
		time.Sleep(3 * time.Second)
		
		// Get batch status
		batchStatus, err := computeClient.GetBatchStatus(ctx, batchResp.TaskIDs)
		if err != nil {
			log.Printf("Failed to get batch status: %v", err)
		} else {
			fmt.Printf("
=== Batch Results ===
")
			fmt.Printf("Completed: %d, Pending: %d, Failed: %d
", 
				len(batchStatus.Completed), len(batchStatus.Pending), len(batchStatus.Failed))
			
			// Print results for each task
			for i, taskID := range batchResp.TaskIDs {
				status, ok := batchStatus.Tasks[taskID]
				if !ok {
					fmt.Printf("Task %d (%s): Status not available
", i+1, taskID)
					continue
				}
				
				fmt.Printf("Task %d (%s): Status = %s
", i+1, taskID, status.Status)
				if status.Status == "SUCCESS" {
					fmt.Printf("  Result: %v
", status.Result)
				} else if status.Status == "FAILED" {
					fmt.Printf("  Exception: %s
", status.Exception)
				}
			}
		}
	}

	fmt.Println("
Compute example complete!")
}

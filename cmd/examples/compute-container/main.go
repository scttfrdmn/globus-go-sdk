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

// Define a simple function to run in a container
const samplePythonFunction = `
def analyze_data(data):
    """
    Analyze input data using container-provided libraries
    
    Args:
        data: A dictionary containing input data
        
    Returns:
        Dictionary with analysis results
    """
    # Import libraries available in the container
    import numpy as np
    import pandas as pd
    import matplotlib.pyplot as plt
    import io
    import base64
    
    # Create a simple dataframe
    df = pd.DataFrame(data)
    
    # Perform basic analysis
    results = {
        "summary": df.describe().to_dict(),
        "correlation": df.corr().to_dict(),
    }
    
    # Generate a simple plot
    plt.figure(figsize=(10, 6))
    df.plot(kind='scatter', x='x', y='y')
    plt.title('Data Visualization')
    plt.grid(True)
    
    # Convert plot to base64 encoded string
    buffer = io.BytesIO()
    plt.savefig(buffer, format='png')
    buffer.seek(0)
    image_base64 = base64.b64encode(buffer.read()).decode('utf-8')
    results["plot"] = image_base64
    
    # Add more analysis
    results["statistics"] = {
        "mean": df.mean().to_dict(),
        "median": df.median().to_dict(),
        "std": df.std().to_dict()
    }
    
    return results
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

	// Get token using client credentials
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

	// Register a container
	fmt.Println("\n=== Registering Container ===")
	timestamp := time.Now().Format("20060102_150405")
	containerName := fmt.Sprintf("data_science_container_%s", timestamp)

	containerReq := &compute.ContainerRegistrationRequest{
		Name:        containerName,
		Description: "Python data science container with numpy, pandas, and matplotlib",
		Image:       "python:3.9-slim",
		Type:        "docker",
		Variables: map[string]string{
			"PYTHONPATH": "/app",
		},
		Arguments: []string{"-m", "pip", "install", "numpy", "pandas", "matplotlib"},
	}

	container, err := computeClient.RegisterContainer(ctx, containerReq)
	if err != nil {
		log.Fatalf("Failed to register container: %v", err)
	}

	fmt.Printf("Container registered: %s (%s)\n", container.Name, container.ID)

	// Register a function
	fmt.Println("\n=== Registering Function ===")
	functionName := fmt.Sprintf("data_analysis_function_%s", timestamp)

	registerRequest := &pkg.FunctionRegisterRequest{
		Function:    samplePythonFunction,
		Name:        functionName,
		Description: "A data analysis function that uses numpy and pandas",
	}

	function, err := computeClient.RegisterFunction(ctx, registerRequest)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Printf("Function registered: %s (%s)\n", function.Name, function.ID)

	// Clean up resources at the end
	defer func() {
		fmt.Println("\n=== Cleaning Up Resources ===")
		if err := computeClient.DeleteFunction(ctx, function.ID); err != nil {
			log.Printf("Warning: Failed to delete function %s: %v", function.ID, err)
		} else {
			fmt.Printf("Function %s deleted successfully\n", function.ID)
		}

		if err := computeClient.DeleteContainer(ctx, container.ID); err != nil {
			log.Printf("Warning: Failed to delete container %s: %v", container.ID, err)
		} else {
			fmt.Printf("Container %s deleted successfully\n", container.ID)
		}
	}()

	// Create sample data for analysis
	sampleData := map[string]interface{}{
		"x": []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"y": []float64{2, 4, 5, 4, 5, 6, 7, 9, 10, 12},
	}

	// Execute the function in the container
	fmt.Println("\n=== Running Function in Container ===")
	containerTaskReq := &pkg.ContainerTaskRequest{
		EndpointID:  selectedEndpoint.ID,
		ContainerID: container.ID,
		FunctionID:  function.ID,
		Args:        []interface{}{sampleData},
		Environment: map[string]string{
			"DEBUG": "true",
		},
	}

	task, err := computeClient.RunContainerFunction(ctx, containerTaskReq)
	if err != nil {
		log.Fatalf("Failed to run container function: %v", err)
	}

	fmt.Printf("Container task submitted: %s (Status: %s)\n", task.TaskID, task.Status)

	// Wait for the task to complete
	fmt.Println("\nWaiting for container task to complete...")
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(3 * time.Second)

		taskStatus, err := computeClient.GetTaskStatus(ctx, task.TaskID)
		if err != nil {
			log.Printf("Error checking task status: %v", err)
			continue
		}

		fmt.Printf("Task status: %s\n", taskStatus.Status)

		if taskStatus.Status == "SUCCESS" {
			fmt.Println("\n=== Container Task Results ===")
			fmt.Printf("Task ID: %s\n", taskStatus.TaskID)

			// Pretty print the result (excluding the plot to avoid large output)
			resultMap, ok := taskStatus.Result.(map[string]interface{})
			if ok {
				// Remove plot from printed output (it's large)
				if _, exists := resultMap["plot"]; exists {
					fmt.Println("Plot: [base64 encoded image - not displayed]")
					delete(resultMap, "plot")
				}

				resultJSON, err := json.MarshalIndent(resultMap, "", "  ")
				if err != nil {
					fmt.Printf("Result: %v\n", taskStatus.Result)
				} else {
					fmt.Printf("Analysis Results:\n%s\n", resultJSON)
				}
			} else {
				fmt.Printf("Result: %v\n", taskStatus.Result)
			}
			break
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Task failed: %s\n", taskStatus.Exception)
			break
		}

		if i == maxAttempts-1 {
			fmt.Println("Task is still running. Check the status manually later.")
		}
	}

	// Execute a code snippet directly in the container (without registering a function)
	fmt.Println("\n=== Running Direct Code in Container ===")
	directCode := `
import numpy as np

def run():
    # Create a random matrix
    matrix = np.random.rand(5, 5)
    
    # Perform operations
    result = {
        "determinant": float(np.linalg.det(matrix)),
        "trace": float(np.trace(matrix)),
        "eigenvalues": [float(x) for x in np.linalg.eigvals(matrix).tolist()],
        "matrix": matrix.tolist()
    }
    return result

output = run()
`

	directTaskReq := &pkg.ContainerTaskRequest{
		EndpointID:  selectedEndpoint.ID,
		ContainerID: container.ID,
		Code:        directCode,
	}

	directTask, err := computeClient.RunContainerFunction(ctx, directTaskReq)
	if err != nil {
		log.Printf("Failed to run direct container code: %v", err)
	} else {
		fmt.Printf("Direct container task submitted: %s (Status: %s)\n", directTask.TaskID, directTask.Status)

		// Wait for the task to complete
		fmt.Println("\nWaiting for direct container task to complete...")
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(2 * time.Second)

			taskStatus, err := computeClient.GetTaskStatus(ctx, directTask.TaskID)
			if err != nil {
				log.Printf("Error checking direct task status: %v", err)
				continue
			}

			fmt.Printf("Direct task status: %s\n", taskStatus.Status)

			if taskStatus.Status == "SUCCESS" {
				fmt.Println("\n=== Direct Container Task Results ===")

				resultJSON, err := json.MarshalIndent(taskStatus.Result, "", "  ")
				if err != nil {
					fmt.Printf("Result: %v\n", taskStatus.Result)
				} else {
					fmt.Printf("Direct Code Results:\n%s\n", resultJSON)
				}
				break
			} else if taskStatus.Status == "FAILED" {
				fmt.Printf("Direct task failed: %s\n", taskStatus.Exception)
				break
			}

			if i == maxAttempts-1 {
				fmt.Println("Direct task is still running. Check the status manually later.")
			}
		}
	}

	fmt.Println("\nContainer example complete!")
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
)

// Define a function that requires specific dependencies
const dependencyBasedFunction = `
def analyze_image(image_url):
    """
    Analyze an image using pre-installed dependencies
    
    Args:
        image_url: URL to an image to analyze
        
    Returns:
        Dictionary with analysis results
    """
    import cv2
    import numpy as np
    import requests
    from PIL import Image
    import io
    
    # Download the image
    response = requests.get(image_url)
    if response.status_code != 200:
        return {"error": f"Failed to download image: HTTP {response.status_code}"}
    
    # Convert to OpenCV format
    image_bytes = io.BytesIO(response.content)
    pil_image = Image.open(image_bytes)
    img = cv2.cvtColor(np.array(pil_image), cv2.COLOR_RGB2BGR)
    
    # Basic image analysis
    height, width, channels = img.shape
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    
    # Extract some features
    results = {
        "dimensions": {
            "width": width,
            "height": height,
            "channels": channels
        },
        "statistics": {
            "mean": float(gray.mean()),
            "std": float(gray.std()),
            "min": int(gray.min()),
            "max": int(gray.max())
        }
    }
    
    # Detect edges with Canny
    edges = cv2.Canny(gray, 100, 200)
    edge_count = cv2.countNonZero(edges)
    results["edge_analysis"] = {
        "edge_pixel_count": edge_count,
        "edge_ratio": float(edge_count) / (width * height)
    }
    
    # Color analysis
    color_analysis = {}
    for i, color in enumerate(['blue', 'green', 'red']):
        channel = img[:,:,i]
        color_analysis[color] = {
            "mean": float(channel.mean()),
            "std": float(channel.std())
        }
    results["color_analysis"] = color_analysis
    
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

	// Create a timestamp for unique naming
	timestamp := time.Now().Format("20060102_150405")

	// Register a dependency
	fmt.Println("\n=== Registering Dependency ===")
	dependencyName := fmt.Sprintf("cv_dependency_%s", timestamp)

	// Define Python requirements
	pythonRequirements := `
opencv-python==4.7.0.72
numpy>=1.22.0
Pillow==9.5.0
requests==2.28.2
`

	dependencyReq := &compute.DependencyRegistrationRequest{
		Name:               dependencyName,
		Description:        "Computer vision dependencies for image processing",
		PythonRequirements: pythonRequirements,
	}

	dependency, err := computeClient.RegisterDependency(ctx, dependencyReq)
	if err != nil {
		log.Fatalf("Failed to register dependency: %v", err)
	}

	fmt.Printf("Dependency registered: %s (%s)\n", dependency.Name, dependency.ID)

	// Register a function
	fmt.Println("\n=== Registering Function ===")
	functionName := fmt.Sprintf("image_analysis_function_%s", timestamp)

	registerRequest := &compute.FunctionRegisterRequest{
		Function:    dependencyBasedFunction,
		Name:        functionName,
		Description: "An image analysis function that uses OpenCV",
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

		if err := computeClient.DeleteDependency(ctx, dependency.ID); err != nil {
			log.Printf("Warning: Failed to delete dependency %s: %v", dependency.ID, err)
		} else {
			fmt.Printf("Dependency %s deleted successfully\n", dependency.ID)
		}
	}()

	// Attach dependency to function
	fmt.Println("\n=== Attaching Dependency to Function ===")
	err = computeClient.AttachDependencyToFunction(ctx, function.ID, dependency.ID)
	if err != nil {
		log.Fatalf("Failed to attach dependency to function: %v", err)
	}
	fmt.Printf("Dependency %s successfully attached to function %s\n", dependency.ID, function.ID)

	// List function dependencies to verify
	fmt.Println("\n=== Listing Function Dependencies ===")
	functionDeps, err := computeClient.ListFunctionDependencies(ctx, function.ID)
	if err != nil {
		log.Printf("Failed to list function dependencies: %v", err)
	} else {
		fmt.Printf("Function %s has %d dependencies:\n", function.ID, len(functionDeps))
		for i, dep := range functionDeps {
			fmt.Printf("%d. %s (%s)\n", i+1, dep.Name, dep.ID)
			if dep.PythonRequirements != "" {
				fmt.Printf("   Python Requirements: %s\n", dep.PythonRequirements)
			}
		}
	}

	// Example image URL to analyze
	imageURL := "https://images.unsplash.com/photo-1617854818583-09e7f077a156?w=800"

	// Execute the function with dependency
	fmt.Println("\n=== Running Function with Dependency ===")
	taskRequest := &compute.TaskRequest{
		FunctionID: function.ID,
		EndpointID: selectedEndpoint.ID,
		Args:       []interface{}{imageURL},
	}

	task, err := computeClient.RunFunction(ctx, taskRequest)
	if err != nil {
		log.Fatalf("Failed to run function: %v", err)
	}

	fmt.Printf("Task submitted: %s (Status: %s)\n", task.TaskID, task.Status)

	// Wait for the task to complete
	fmt.Println("\nWaiting for task to complete...")
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
			fmt.Println("\n=== Image Analysis Results ===")
			fmt.Printf("Task ID: %s\n", taskStatus.TaskID)

			// Pretty print the result
			fmt.Printf("Image dimensions: %v\n", taskStatus.Result.(map[string]interface{})["dimensions"])
			fmt.Printf("Image statistics: %v\n", taskStatus.Result.(map[string]interface{})["statistics"])
			fmt.Printf("Edge analysis: %v\n", taskStatus.Result.(map[string]interface{})["edge_analysis"])
			fmt.Printf("Color analysis: %v\n", taskStatus.Result.(map[string]interface{})["color_analysis"])

			break
		} else if taskStatus.Status == "FAILED" {
			fmt.Printf("Task failed: %s\n", taskStatus.Exception)
			break
		}

		if i == maxAttempts-1 {
			fmt.Println("Task is still running. Check the status manually later.")
		}
	}

	// Demonstrate detaching a dependency
	fmt.Println("\n=== Detaching Dependency from Function ===")
	err = computeClient.DetachDependencyFromFunction(ctx, function.ID, dependency.ID)
	if err != nil {
		log.Printf("Failed to detach dependency from function: %v", err)
	} else {
		fmt.Printf("Dependency %s successfully detached from function %s\n", dependency.ID, function.ID)
	}

	fmt.Println("\nDependency management example complete!")
}

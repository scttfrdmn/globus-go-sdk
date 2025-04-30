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
)

// Define a function that uses environment variables
const environmentBasedFunction = `
def process_data(data_url):
    """
    Process data using environment configuration
    
    Args:
        data_url: URL to data for processing
        
    Returns:
        Dictionary with processing results
    """
    import os
    import requests
    import json
    import time
    
    # Access environment variables
    api_key = os.environ.get('API_KEY', '')
    log_level = os.environ.get('LOG_LEVEL', 'INFO')
    
    # Prepare result object
    result = {
        "status": "success",
        "timestamp": time.time(),
        "environment": {
            "available_vars": {},
            "resource_info": {}
        },
        "data_processing": {}
    }
    
    # Collect environment info
    for env_var in ['API_KEY', 'LOG_LEVEL', 'DEBUG', 'ENDPOINT_NAME', 'MAX_RETRIES']:
        value = os.environ.get(env_var)
        if value:
            # Mask API_KEY value for security
            if env_var == 'API_KEY':
                result["environment"]["available_vars"][env_var] = '****' + value[-4:]
            else:
                result["environment"]["available_vars"][env_var] = value
    
    # Get resource details if available
    import multiprocessing
    import platform
    import psutil
    
    result["environment"]["resource_info"] = {
        "cpu_count": multiprocessing.cpu_count(),
        "platform": platform.platform(),
        "memory_available": psutil.virtual_memory().available,
        "process_memory": psutil.Process().memory_info().rss
    }
    
    # Process the data based on URL
    headers = {}
    if api_key:
        headers['Authorization'] = f'Bearer {api_key}'
    
    try:
        # Request data with retry logic
        max_retries = int(os.environ.get('MAX_RETRIES', '3'))
        retry_count = 0
        
        while retry_count < max_retries:
            response = requests.get(data_url, headers=headers)
            if response.status_code == 200:
                break
            retry_count += 1
            time.sleep(1)
        
        if response.status_code != 200:
            result["status"] = "error"
            result["data_processing"]["error"] = f"Failed to get data after {max_retries} retries: {response.status_code}"
            return result
        
        # Process the data
        data = response.json()
        result["data_processing"]["record_count"] = len(data) if isinstance(data, list) else 1
        result["data_processing"]["content_length"] = len(response.content)
        
        # If in debug mode, include more details
        if os.environ.get('DEBUG') == 'true':
            result["data_processing"]["headers"] = dict(response.headers)
            result["data_processing"]["time_taken"] = response.elapsed.total_seconds()
    
    except Exception as e:
        result["status"] = "error"
        result["data_processing"]["error"] = str(e)
    
    return result
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

	// Create a secret
	fmt.Println("\n=== Creating Secret ===")
	secretName := fmt.Sprintf("API_KEY_%s", timestamp)
	secretRequest := &pkg.SecretCreateRequest{
		Name:        secretName,
		Description: "Sample API key for environment example",
		Value:       "abcd1234-test-api-key-5678efgh",
	}

	secret, err := computeClient.CreateSecret(ctx, secretRequest)
	if err != nil {
		log.Fatalf("Failed to create secret: %v", err)
	}

	fmt.Printf("Secret created: %s (%s)\n", secret.Name, secret.ID)

	// List secrets
	fmt.Println("\n=== Listing Secrets ===")
	secrets, err := computeClient.ListSecrets(ctx)
	if err != nil {
		log.Printf("Failed to list secrets: %v", err)
	} else {
		fmt.Printf("Found %d secrets:\n", len(secrets))
		for i, s := range secrets {
			fmt.Printf("%d. %s (%s)\n", i+1, s.Name, s.ID)
			if s.Description != "" {
				fmt.Printf("   Description: %s\n", s.Description)
			}
		}
	}

	// Create environment configuration
	fmt.Println("\n=== Creating Environment Configuration ===")
	envName := fmt.Sprintf("data_processing_env_%s", timestamp)

	envRequest := &pkg.EnvironmentCreateRequest{
		Name:        envName,
		Description: "Environment for data processing functions",
		Variables: map[string]string{
			"LOG_LEVEL":     "DEBUG",
			"DEBUG":         "true",
			"MAX_RETRIES":   "5",
			"ENDPOINT_NAME": selectedEndpoint.Name,
		},
		Secrets: []string{secret.ID},
		Resources: map[string]interface{}{
			"cpu_cores":      2,
			"memory_limit":   "4GB",
			"execution_time": 300, // seconds
		},
	}

	environment, err := computeClient.CreateEnvironment(ctx, envRequest)
	if err != nil {
		log.Fatalf("Failed to create environment: %v", err)
	}

	fmt.Printf("Environment created: %s (%s)\n", environment.Name, environment.ID)

	// Clean up resources at the end
	defer func() {
		fmt.Println("\n=== Cleaning Up Resources ===")
		if err := computeClient.DeleteEnvironment(ctx, environment.ID); err != nil {
			log.Printf("Warning: Failed to delete environment %s: %v", environment.ID, err)
		} else {
			fmt.Printf("Environment %s deleted successfully\n", environment.ID)
		}

		if err := computeClient.DeleteSecret(ctx, secret.ID); err != nil {
			log.Printf("Warning: Failed to delete secret %s: %v", secret.ID, err)
		} else {
			fmt.Printf("Secret %s deleted successfully\n", secret.ID)
		}

		if functionID != "" {
			if err := computeClient.DeleteFunction(ctx, functionID); err != nil {
				log.Printf("Warning: Failed to delete function %s: %v", functionID, err)
			} else {
				fmt.Printf("Function %s deleted successfully\n", functionID)
			}
		}
	}()

	// List environment configurations
	fmt.Println("\n=== Listing Environment Configurations ===")
	environments, err := computeClient.ListEnvironments(ctx, &pkg.ListEnvironmentsOptions{
		PerPage: 10,
	})
	if err != nil {
		log.Printf("Failed to list environments: %v", err)
	} else {
		fmt.Printf("Found %d environments:\n", len(environments.Environments))
		for i, env := range environments.Environments {
			fmt.Printf("%d. %s (%s)\n", i+1, env.Name, env.ID)
			if env.Description != "" {
				fmt.Printf("   Description: %s\n", env.Description)
			}
			fmt.Printf("   Variables: %d, Secrets: %d\n", len(env.Variables), len(env.Secrets))
		}
	}

	// Register a function
	fmt.Println("\n=== Registering Function ===")
	functionName := fmt.Sprintf("env_function_%s", timestamp)

	registerRequest := &pkg.FunctionRegisterRequest{
		Function:    environmentBasedFunction,
		Name:        functionName,
		Description: "A function that uses environment configuration",
	}

	function, err := computeClient.RegisterFunction(ctx, registerRequest)
	if err != nil {
		log.Fatalf("Failed to register function: %v", err)
	}

	fmt.Printf("Function registered: %s (%s)\n", function.Name, function.ID)
	functionID := function.ID

	// Prepare a task request
	dataURL := "https://jsonplaceholder.typicode.com/posts"
	taskRequest := &pkg.TaskRequest{
		FunctionID: function.ID,
		EndpointID: selectedEndpoint.ID,
		Args:       []interface{}{dataURL},
	}

	// Apply environment to task
	fmt.Println("\n=== Applying Environment to Task ===")
	enrichedRequest, err := computeClient.ApplyEnvironmentToTask(ctx, taskRequest, environment.ID)
	if err != nil {
		log.Fatalf("Failed to apply environment to task: %v", err)
	}

	fmt.Println("Environment successfully applied to task request")

	// Execute the function with environment
	fmt.Println("\n=== Running Function with Environment ===")
	task, err := computeClient.RunFunction(ctx, enrichedRequest)
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
			fmt.Println("\n=== Function Execution Results ===")
			fmt.Printf("Task ID: %s\n", taskStatus.TaskID)
			
			// Extract and display environment information
			result, ok := taskStatus.Result.(map[string]interface{})
			if ok {
				fmt.Println("\nExecution Status:", result["status"])
				
				// Show environment variables used
				if env, ok := result["environment"].(map[string]interface{}); ok {
					if vars, ok := env["available_vars"].(map[string]interface{}); ok {
						fmt.Println("\nEnvironment Variables Used:")
						for k, v := range vars {
							fmt.Printf("  %s: %v\n", k, v)
						}
					}
					
					// Show resource information
					if res, ok := env["resource_info"].(map[string]interface{}); ok {
						fmt.Println("\nResource Information:")
						for k, v := range res {
							fmt.Printf("  %s: %v\n", k, v)
						}
					}
				}
				
				// Show data processing results
				if proc, ok := result["data_processing"].(map[string]interface{}); ok {
					fmt.Println("\nData Processing Results:")
					for k, v := range proc {
						fmt.Printf("  %s: %v\n", k, v)
					}
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

	// Update environment configuration
	fmt.Println("\n=== Updating Environment Configuration ===")
	updateRequest := &pkg.EnvironmentUpdateRequest{
		Variables: map[string]string{
			"LOG_LEVEL": "INFO",
			"DEBUG":     "false",
		},
		Resources: map[string]interface{}{
			"cpu_cores":    4,
			"memory_limit": "8GB",
		},
	}

	updatedEnv, err := computeClient.UpdateEnvironment(ctx, environment.ID, updateRequest)
	if err != nil {
		log.Printf("Failed to update environment: %v", err)
	} else {
		fmt.Printf("Environment updated: %s\n", updatedEnv.ID)
		fmt.Printf("Updated variables - LOG_LEVEL: %s, DEBUG: %s\n", 
			updatedEnv.Variables["LOG_LEVEL"], updatedEnv.Variables["DEBUG"])
	}

	fmt.Println("\nEnvironment configuration example complete!")
}
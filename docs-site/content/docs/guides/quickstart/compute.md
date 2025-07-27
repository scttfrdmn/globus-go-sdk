---
title: "Compute Service Quick Start"
weight: 50
---

# Compute Service Quick Start

This guide will help you get started with the Globus Compute service using the Go SDK. Globus Compute allows you to execute remote functions on Globus endpoints, create and manage containers, and configure execution environments.

## Setup

First, import the required packages and create a context:

```go
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

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    // Continue with the examples below...
}
```

## Creating a Compute Client

There are two main ways to create a Compute client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Compute client
computeClient, err := config.NewComputeClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create compute client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Compute client with options
computeClient, err := compute.NewClient(
    compute.WithAccessToken(os.Getenv("GLOBUS_ACCESS_TOKEN")),
    compute.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create compute client: %v", err)
}
```

## Obtaining Compute Scope Tokens

To use the Compute service, you need a token with the correct scope:

```go
// Create a new SDK configuration
config := pkg.NewConfigFromEnvironment().
    WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
    WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

// Create a new Auth client
authClient, err := config.NewAuthClient()
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}

// Get token using client credentials for the Compute scope
tokenResp, err := authClient.GetClientCredentialsToken(ctx, pkg.ComputeScope)
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

fmt.Printf("Obtained access token (expires in %d seconds)\n", tokenResp.ExpiresIn)
accessToken := tokenResp.AccessToken

// Create Compute client with the token
computeClient, err := config.NewComputeClient(accessToken)
if err != nil {
    log.Fatalf("Failed to create compute client: %v", err)
}
```

## Working with Compute Endpoints

Compute endpoints are the resources where your functions will run.

### Listing Compute Endpoints

```go
// List available endpoints
endpoints, err := computeClient.ListEndpoints(ctx, &pkg.ListEndpointsOptions{
    PerPage: 10,
})
if err != nil {
    log.Fatalf("Failed to list endpoints: %v", err)
}

fmt.Printf("Found %d compute endpoints:\n", len(endpoints.Endpoints))
for i, endpoint := range endpoints.Endpoints {
    fmt.Printf("%d. %s (%s)\n", i+1, endpoint.Name, endpoint.ID)
    fmt.Printf("   Status: %s, Connected: %t\n", endpoint.Status, endpoint.Connected)
}
```

### Getting Endpoint Details

```go
// Get details about a specific endpoint
endpointID := "your-endpoint-id"
endpoint, err := computeClient.GetEndpoint(ctx, endpointID)
if err != nil {
    log.Fatalf("Failed to get endpoint: %v", err)
}

fmt.Printf("Endpoint: %s\n", endpoint.Name)
fmt.Printf("Status: %s\n", endpoint.Status)
fmt.Printf("Connected: %t\n", endpoint.Connected)
```

## Working with Functions

Functions are pieces of code that you can register and execute on Compute endpoints.

### Registering a Function

```go
// Define a simple Python function
sampleFunction := `def hello(name="World"):
    import datetime
    now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    return f"Hello, {name}! The time is {now}"
`

// Register the function
functionName := fmt.Sprintf("example_function_%s", time.Now().Format("20060102_150405"))
registerRequest := &pkg.FunctionRegisterRequest{
    Function:    sampleFunction,
    Name:        functionName,
    Description: "A simple greeting function created by the Globus Go SDK",
}

function, err := computeClient.RegisterFunction(ctx, registerRequest)
if err != nil {
    log.Fatalf("Failed to register function: %v", err)
}

fmt.Printf("Function registered: %s (%s)\n", function.Name, function.ID)
```

### Listing Functions

```go
// List all functions
functions, err := computeClient.ListFunctions(ctx, &pkg.ListFunctionsOptions{
    PerPage: 10,
})
if err != nil {
    log.Fatalf("Failed to list functions: %v", err)
}

fmt.Printf("Found %d functions:\n", len(functions.Functions))
for i, fn := range functions.Functions {
    fmt.Printf("%d. %s (%s)\n", i+1, fn.Name, fn.ID)
    fmt.Printf("   Description: %s\n", fn.Description)
}
```

### Getting Function Details

```go
// Get details about a specific function
functionID := "your-function-id"
function, err := computeClient.GetFunction(ctx, functionID)
if err != nil {
    log.Fatalf("Failed to get function: %v", err)
}

fmt.Printf("Function: %s\n", function.Name)
fmt.Printf("Description: %s\n", function.Description)
fmt.Printf("ID: %s\n", function.ID)
```

### Updating a Function

```go
// Update an existing function
updateRequest := &pkg.FunctionUpdateRequest{
    Description: "Updated description for my function",
}

updatedFunction, err := computeClient.UpdateFunction(ctx, functionID, updateRequest)
if err != nil {
    log.Fatalf("Failed to update function: %v", err)
}

fmt.Printf("Function updated: %s\n", updatedFunction.Name)
```

### Deleting a Function

```go
// Delete a function
err = computeClient.DeleteFunction(ctx, functionID)
if err != nil {
    log.Fatalf("Failed to delete function: %v", err)
}

fmt.Println("Function deleted successfully")
```

## Running Functions

After registering functions, you can execute them on Compute endpoints.

### Running a Single Function

```go
// Execute a function on an endpoint
taskRequest := &pkg.TaskRequest{
    FunctionID: function.ID,
    EndpointID: endpoint.ID,
    Args:       []interface{}{"Globus Go SDK"},
}

task, err := computeClient.RunFunction(ctx, taskRequest)
if err != nil {
    log.Fatalf("Failed to run function: %v", err)
}

fmt.Printf("Task submitted: %s (Status: %s)\n", task.TaskID, task.Status)
```

### Getting Task Status

```go
// Get the status of a task
taskStatus, err := computeClient.GetTaskStatus(ctx, task.TaskID)
if err != nil {
    log.Fatalf("Failed to get task status: %v", err)
}

fmt.Printf("Task ID: %s\n", taskStatus.TaskID)
fmt.Printf("Status: %s\n", taskStatus.Status)

if taskStatus.Status == "SUCCESS" {
    fmt.Printf("Result: %v\n", taskStatus.Result)
} else if taskStatus.Status == "FAILED" {
    fmt.Printf("Exception: %s\n", taskStatus.Exception)
}
```

### Running Batch Functions

```go
// Execute multiple functions in a batch
batchRequest := &pkg.BatchTaskRequest{
    Tasks: []pkg.TaskRequest{
        {
            FunctionID: function.ID,
            EndpointID: endpoint.ID,
            Args:       []interface{}{"First Batch Task"},
        },
        {
            FunctionID: function.ID,
            EndpointID: endpoint.ID,
            Args:       []interface{}{"Second Batch Task"},
        },
    },
}

batchResp, err := computeClient.RunBatch(ctx, batchRequest)
if err != nil {
    log.Fatalf("Failed to run batch: %v", err)
}

fmt.Printf("Batch submitted with %d tasks\n", len(batchResp.TaskIDs))
```

### Getting Batch Status

```go
// Get the status of a batch of tasks
batchStatus, err := computeClient.GetBatchStatus(ctx, batchResp.TaskIDs)
if err != nil {
    log.Fatalf("Failed to get batch status: %v", err)
}

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
```

## Working with Containers

Containers allow you to execute functions in custom environments with specific dependencies.

### Registering a Container

```go
// Register a container
containerName := fmt.Sprintf("data_science_container_%s", time.Now().Format("20060102_150405"))

containerReq := &pkg.ContainerRegistrationRequest{
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
```

### Running a Function in a Container

```go
// Define sample data for the function
sampleData := map[string]interface{}{
    "x": []float64{1, 2, 3, 4, 5},
    "y": []float64{2, 4, 5, 4, 5},
}

// Execute the function in the container
containerTaskReq := &pkg.ContainerTaskRequest{
    EndpointID:  endpoint.ID,
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
```

### Running Direct Code in a Container

```go
// Execute code directly in a container without registering a function
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
    EndpointID:  endpoint.ID,
    ContainerID: container.ID,
    Code:        directCode,
}

directTask, err := computeClient.RunContainerFunction(ctx, directTaskReq)
if err != nil {
    log.Fatalf("Failed to run direct container code: %v", err)
}

fmt.Printf("Direct container task submitted: %s (Status: %s)\n", directTask.TaskID, directTask.Status)
```

### Deleting a Container

```go
// Delete a container
err = computeClient.DeleteContainer(ctx, container.ID)
if err != nil {
    log.Fatalf("Failed to delete container: %v", err)
}

fmt.Println("Container deleted successfully")
```

## Working with Environments

Environments allow you to configure resource requirements, environment variables, and secrets for your functions.

### Creating an Environment

```go
// Create an environment configuration
envName := fmt.Sprintf("data_processing_env_%s", time.Now().Format("20060102_150405"))

envRequest := &pkg.EnvironmentCreateRequest{
    Name:        envName,
    Description: "Environment for data processing functions",
    Variables: map[string]string{
        "LOG_LEVEL":     "DEBUG",
        "DEBUG":         "true",
        "MAX_RETRIES":   "5",
    },
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
```

### Creating and Using Secrets

```go
// Create a secret
secretName := fmt.Sprintf("API_KEY_%s", time.Now().Format("20060102_150405"))
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

// Add the secret to an environment
envRequest.Secrets = []string{secret.ID}
```

### Applying an Environment to a Task

```go
// Prepare a task request
taskRequest := &pkg.TaskRequest{
    FunctionID: function.ID,
    EndpointID: endpoint.ID,
    Args:       []interface{}{"https://jsonplaceholder.typicode.com/posts"},
}

// Apply environment to task
enrichedRequest, err := computeClient.ApplyEnvironmentToTask(ctx, taskRequest, environment.ID)
if err != nil {
    log.Fatalf("Failed to apply environment to task: %v", err)
}

// Execute the function with the environment
task, err := computeClient.RunFunction(ctx, enrichedRequest)
if err != nil {
    log.Fatalf("Failed to run function: %v", err)
}

fmt.Printf("Task with environment submitted: %s (Status: %s)\n", task.TaskID, task.Status)
```

### Updating an Environment

```go
// Update an environment configuration
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
    log.Fatalf("Failed to update environment: %v", err)
}

fmt.Printf("Environment updated: %s\n", updatedEnv.ID)
```

### Deleting an Environment and Secrets

```go
// Delete an environment
err = computeClient.DeleteEnvironment(ctx, environment.ID)
if err != nil {
    log.Fatalf("Failed to delete environment: %v", err)
}

fmt.Println("Environment deleted successfully")

// Delete a secret
err = computeClient.DeleteSecret(ctx, secret.ID)
if err != nil {
    log.Fatalf("Failed to delete secret: %v", err)
}

fmt.Println("Secret deleted successfully")
```

## Complete Example

Here's a complete example that registers a function, runs it, and checks the result:

```go
package main

import (
    "context"
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

    // Register a simple function
    fmt.Println("\n=== Registering Function ===")
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

    fmt.Printf("Function registered: %s (%s)\n", function.Name, function.ID)

    // Make sure to clean up the function after the example
    defer func() {
        fmt.Println("\n=== Cleaning Up Function ===")
        if err := computeClient.DeleteFunction(ctx, function.ID); err != nil {
            log.Printf("Warning: Failed to delete function %s: %v", function.ID, err)
        } else {
            fmt.Printf("Function %s deleted successfully\n", function.ID)
        }
    }()

    // Execute the simple function
    fmt.Println("\n=== Running Function ===")
    taskRequest := &pkg.TaskRequest{
        FunctionID: function.ID,
        EndpointID: selectedEndpoint.ID,
        Args:       []interface{}{"Globus Go SDK"},
    }

    task, err := computeClient.RunFunction(ctx, taskRequest)
    if err != nil {
        log.Fatalf("Failed to run function: %v", err)
    }

    fmt.Printf("Task submitted: %s (Status: %s)\n", task.TaskID, task.Status)

    // Wait for task to complete and get results
    fmt.Println("\nWaiting for task to complete...")
    time.Sleep(3 * time.Second)

    // Get task status
    fmt.Println("\n=== Task Results ===")
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

    fmt.Println("\nCompute example complete!")
}
```

## Error Handling

The Compute service provides error handling for common issues:

```go
// Try to get a non-existent function
_, err = computeClient.GetFunction(ctx, "non-existent-function")
if err != nil {
    if strings.Contains(err.Error(), "404") {
        fmt.Println("Function not found - check the function ID")
    } else if strings.Contains(err.Error(), "403") {
        fmt.Println("Permission denied - check your access token and permissions")
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Next Steps

Now that you understand the basics of the Compute service, you can:

1. **Create Custom Functions**: Develop more complex functions that perform data analysis, transformation, or other tasks
2. **Experiment with Containers**: Use containers to create consistent execution environments with specific dependencies
3. **Set Up Resource Configurations**: Configure environments with appropriate CPU, memory, and execution time limits
4. **Integrate with Other Services**: Combine Compute with other Globus services like Transfer and Flows for comprehensive workflows

For more details, check out the [Compute Service API Reference](/docs/reference/compute/) documentation.
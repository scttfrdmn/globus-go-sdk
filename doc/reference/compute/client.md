# Compute Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Compute client provides access to the Globus Compute API, which allows you to execute functions on remote endpoints, manage environments, and orchestrate computational workflows.

## Client Structure

```go
type Client struct {
    client *core.Client
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*core.Client` | Core client for making HTTP requests |

## Creating a Compute Client

```go
// Create a compute client with options
client, err := compute.NewClient(
    compute.WithAccessToken("access-token"),
    compute.WithHTTPDebugging(),
)
if err != nil {
    // Handle error
}
```

### Options

| Option | Description |
|--------|-------------|
| `WithAccessToken(token string)` | Sets the access token for authorization |
| `WithAuthorizer(auth core.Authorizer)` | Sets a custom authorizer (alternative to access token) |
| `WithBaseURL(url string)` | Sets a custom base URL (default: "https://compute.api.globus.org/v2/") |
| `WithHTTPDebugging()` | Enables HTTP debugging |
| `WithHTTPTracing()` | Enables HTTP tracing |

## Endpoint Management

### Listing Endpoints

```go
// List compute endpoints
endpoints, err := client.ListEndpoints(ctx, nil)
if err != nil {
    // Handle error
}

// List endpoints with options
endpoints, err := client.ListEndpoints(ctx, &compute.ListEndpointsOptions{
    Limit:  100,
    Offset: 0,
    Filter: "owner=me",
})
if err != nil {
    // Handle error
}

// Iterate through endpoints
for _, endpoint := range endpoints.Endpoints {
    fmt.Printf("Endpoint: %s (%s)\n", endpoint.DisplayName, endpoint.ID)
    fmt.Printf("  Status: %s\n", endpoint.Status)
    fmt.Printf("  Owner: %s\n", endpoint.Owner)
}
```

### Getting an Endpoint

```go
// Get a specific endpoint
endpoint, err := client.GetEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

fmt.Printf("Endpoint: %s (%s)\n", endpoint.DisplayName, endpoint.ID)
fmt.Printf("Status: %s\n", endpoint.Status)
fmt.Printf("Description: %s\n", endpoint.Description)
fmt.Printf("Owner: %s\n", endpoint.Owner)
fmt.Printf("Metrics: Queue size %d, active workers %d\n", 
    endpoint.Metrics.QueueSize, endpoint.Metrics.ActiveWorkers)
```

## Function Management

### Registering a Function

```go
// Register a new function
registerRequest := &compute.FunctionRegisterRequest{
    Name:        "my-function",
    Description: "A simple example function",
    Code:        "def my_function(x, y):\n    return x + y",
    Entry:       "my_function",
    Container:   "python:3.9",
}

function, err := client.RegisterFunction(ctx, registerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Function registered: %s (%s)\n", function.Name, function.FunctionID)
```

### Listing Functions

```go
// List functions
functions, err := client.ListFunctions(ctx, nil)
if err != nil {
    // Handle error
}

// Iterate through functions
for _, function := range functions.Functions {
    fmt.Printf("Function: %s (%s)\n", function.Name, function.FunctionID)
    fmt.Printf("  Container: %s\n", function.Container)
    fmt.Printf("  Created: %s\n", function.CreatedTimestamp)
}
```

### Getting a Function

```go
// Get a specific function
function, err := client.GetFunction(ctx, "function-id")
if err != nil {
    // Handle error
}

fmt.Printf("Function: %s (%s)\n", function.Name, function.FunctionID)
fmt.Printf("Description: %s\n", function.Description)
fmt.Printf("Entry point: %s\n", function.Entry)
fmt.Printf("Container: %s\n", function.Container)
fmt.Printf("Code:\n%s\n", function.Code)
```

### Updating a Function

```go
// Update an existing function
updateRequest := &compute.FunctionUpdateRequest{
    Description: "Updated function description",
    Code:        "def my_function(x, y):\n    return x + y + 1",
    Entry:       "my_function",
}

function, err := client.UpdateFunction(ctx, "function-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Function updated: %s\n", function.Name)
```

### Deleting a Function

```go
// Delete a function
err := client.DeleteFunction(ctx, "function-id")
if err != nil {
    // Handle error
}

fmt.Println("Function deleted successfully")
```

## Function Execution

### Running a Function

```go
// Create a task request
taskRequest := &compute.TaskRequest{
    FunctionID:  "function-id",
    EndpointID:  "endpoint-id",
    Parameters:  map[string]interface{}{"x": 5, "y": 10},
    Label:       "Addition Task",
}

// Run the function
taskResponse, err := client.RunFunction(ctx, taskRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task submitted: %s\n", taskResponse.TaskID)
```

### Getting Task Status

```go
// Check task status
taskStatus, err := client.GetTaskStatus(ctx, "task-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task status: %s\n", taskStatus.Status)
if taskStatus.Status == "success" {
    fmt.Printf("Result: %v\n", taskStatus.Result)
} else if taskStatus.Status == "failed" {
    fmt.Printf("Error: %s\n", taskStatus.ErrorMessage)
}
```

### Running a Batch of Functions

```go
// Create a batch task request
batchRequest := &compute.BatchTaskRequest{
    Tasks: []compute.TaskRequest{
        {
            FunctionID:  "function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"x": 1, "y": 2},
            Label:       "Batch Task 1",
        },
        {
            FunctionID:  "function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"x": 3, "y": 4},
            Label:       "Batch Task 2",
        },
        {
            FunctionID:  "function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"x": 5, "y": 6},
            Label:       "Batch Task 3",
        },
    },
}

// Run the batch
batchResponse, err := client.RunBatch(ctx, batchRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Batch submitted with %d tasks\n", len(batchResponse.TaskIDs))
for i, taskID := range batchResponse.TaskIDs {
    fmt.Printf("Task %d ID: %s\n", i+1, taskID)
}
```

### Getting Batch Status

```go
// Check batch status
batchStatus, err := client.GetBatchStatus(ctx, batchResponse.TaskIDs)
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved status for %d tasks\n", len(batchStatus.TaskStatuses))
for taskID, status := range batchStatus.TaskStatuses {
    fmt.Printf("Task %s: %s\n", taskID, status.Status)
    if status.Status == "success" {
        fmt.Printf("  Result: %v\n", status.Result)
    } else if status.Status == "failed" {
        fmt.Printf("  Error: %s\n", status.ErrorMessage)
    }
}
```

### Listing Tasks

```go
// List tasks
tasks, err := client.ListTasks(ctx, nil)
if err != nil {
    // Handle error
}

// List tasks with options
tasks, err := client.ListTasks(ctx, &compute.TaskListOptions{
    FunctionID: "function-id",
    EndpointID: "endpoint-id",
    Status:     "success",
    Limit:      50,
    Offset:     0,
})
if err != nil {
    // Handle error
}

fmt.Printf("Found %d tasks\n", len(tasks.Tasks))
for _, taskID := range tasks.Tasks {
    fmt.Printf("Task ID: %s\n", taskID)
}
```

### Canceling a Task

```go
// Cancel a task
err := client.CancelTask(ctx, "task-id")
if err != nil {
    // Handle error
}

fmt.Println("Task canceled successfully")
```

## Container Management

### Registering a Container

```go
// Register a new container
containerRequest := &compute.ContainerRegistrationRequest{
    Name:        "python-data-science",
    Description: "Container with data science libraries",
    Type:        "docker",
    Location:    "docker://python:3.9-slim",
    Dependencies: map[string]string{
        "numpy":     "1.22.0",
        "pandas":    "1.4.0",
        "matplotlib": "3.5.0",
    },
}

container, err := client.RegisterContainer(ctx, containerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Container registered: %s (%s)\n", container.Name, container.ContainerID)
```

### Listing Containers

```go
// List containers
containers, err := client.ListContainers(ctx, nil)
if err != nil {
    // Handle error
}

// Iterate through containers
for _, container := range containers.Containers {
    fmt.Printf("Container: %s (%s)\n", container.Name, container.ContainerID)
    fmt.Printf("  Type: %s\n", container.Type)
    fmt.Printf("  Location: %s\n", container.Location)
}
```

### Getting a Container

```go
// Get a specific container
container, err := client.GetContainer(ctx, "container-id")
if err != nil {
    // Handle error
}

fmt.Printf("Container: %s (%s)\n", container.Name, container.ContainerID)
fmt.Printf("Description: %s\n", container.Description)
fmt.Printf("Type: %s\n", container.Type)
fmt.Printf("Location: %s\n", container.Location)
```

### Updating a Container

```go
// Update an existing container
updateRequest := &compute.ContainerUpdateRequest{
    Description: "Updated container description",
    Dependencies: map[string]string{
        "numpy":     "1.23.0",
        "pandas":    "1.5.0",
        "matplotlib": "3.6.0",
        "scikit-learn": "1.1.0",
    },
}

container, err := client.UpdateContainer(ctx, "container-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Container updated: %s\n", container.Name)
```

### Deleting a Container

```go
// Delete a container
err := client.DeleteContainer(ctx, "container-id")
if err != nil {
    // Handle error
}

fmt.Println("Container deleted successfully")
```

## Environment Management

### Creating an Environment

```go
// Create a new environment
environmentRequest := &compute.EnvironmentCreateRequest{
    Name:        "production-environment",
    Description: "Environment for production workloads",
    Variables: map[string]string{
        "LOG_LEVEL":   "INFO",
        "API_TIMEOUT": "30",
        "ENV":         "production",
    },
}

environment, err := client.CreateEnvironment(ctx, environmentRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Environment created: %s (%s)\n", environment.Name, environment.EnvironmentID)
```

### Listing Environments

```go
// List environments
environments, err := client.ListEnvironments(ctx, nil)
if err != nil {
    // Handle error
}

// Iterate through environments
for _, env := range environments.Environments {
    fmt.Printf("Environment: %s (%s)\n", env.Name, env.EnvironmentID)
    fmt.Printf("  Description: %s\n", env.Description)
}
```

### Getting an Environment

```go
// Get a specific environment
environment, err := client.GetEnvironment(ctx, "environment-id")
if err != nil {
    // Handle error
}

fmt.Printf("Environment: %s (%s)\n", environment.Name, environment.EnvironmentID)
fmt.Printf("Description: %s\n", environment.Description)
fmt.Printf("Variables:\n")
for key, value := range environment.Variables {
    fmt.Printf("  %s: %s\n", key, value)
}
```

### Updating an Environment

```go
// Update an existing environment
updateRequest := &compute.EnvironmentUpdateRequest{
    Description: "Updated environment description",
    Variables: map[string]string{
        "LOG_LEVEL":   "DEBUG",
        "API_TIMEOUT": "60",
        "ENV":         "production",
        "FEATURE_FLAG": "true",
    },
}

environment, err := client.UpdateEnvironment(ctx, "environment-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Environment updated: %s\n", environment.Name)
```

### Deleting an Environment

```go
// Delete an environment
err := client.DeleteEnvironment(ctx, "environment-id")
if err != nil {
    // Handle error
}

fmt.Println("Environment deleted successfully")
```

### Creating a Secret

```go
// Create a new secret
secretRequest := &compute.SecretCreateRequest{
    Name:  "api-key",
    Value: "secret-api-key-value",
    Description: "API key for external service",
}

secret, err := client.CreateSecret(ctx, "environment-id", secretRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Secret created: %s\n", secret.Name)
```

## Dependency Management

### Registering a Dependency

```go
// Register a new dependency
dependencyRequest := &compute.DependencyRegistrationRequest{
    Name:        "data-science-stack",
    Description: "Common data science packages",
    Type:        "pip",
    Packages: []compute.PythonPackage{
        {
            Name:    "numpy",
            Version: "1.22.0",
        },
        {
            Name:    "pandas",
            Version: "1.4.0",
        },
        {
            Name:    "scikit-learn",
            Version: "1.0.0",
        },
    },
}

dependency, err := client.RegisterDependency(ctx, dependencyRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency registered: %s (%s)\n", dependency.Name, dependency.DependencyID)
```

### Listing Dependencies

```go
// List dependencies
dependencies, err := client.ListDependencies(ctx, nil)
if err != nil {
    // Handle error
}

// Iterate through dependencies
for _, dep := range dependencies.Dependencies {
    fmt.Printf("Dependency: %s (%s)\n", dep.Name, dep.DependencyID)
    fmt.Printf("  Type: %s\n", dep.Type)
}
```

### Getting a Dependency

```go
// Get a specific dependency
dependency, err := client.GetDependency(ctx, "dependency-id")
if err != nil {
    // Handle error
}

fmt.Printf("Dependency: %s (%s)\n", dependency.Name, dependency.DependencyID)
fmt.Printf("Description: %s\n", dependency.Description)
fmt.Printf("Type: %s\n", dependency.Type)
fmt.Printf("Packages:\n")
for _, pkg := range dependency.Packages {
    fmt.Printf("  %s: %s\n", pkg.Name, pkg.Version)
}
```

### Updating a Dependency

```go
// Update an existing dependency
updateRequest := &compute.DependencyUpdateRequest{
    Description: "Updated dependency description",
    Packages: []compute.PythonPackage{
        {
            Name:    "numpy",
            Version: "1.23.0",
        },
        {
            Name:    "pandas",
            Version: "1.5.0",
        },
        {
            Name:    "scikit-learn",
            Version: "1.1.0",
        },
        {
            Name:    "matplotlib",
            Version: "3.6.0",
        },
    },
}

dependency, err := client.UpdateDependency(ctx, "dependency-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency updated: %s\n", dependency.Name)
```

### Deleting a Dependency

```go
// Delete a dependency
err := client.DeleteDependency(ctx, "dependency-id")
if err != nil {
    // Handle error
}

fmt.Println("Dependency deleted successfully")
```

## Workflow Management

### Creating a Workflow

```go
// Create a new workflow
workflowRequest := &compute.WorkflowCreateRequest{
    Name:        "data-processing-workflow",
    Description: "Process and analyze data files",
    Tasks: []compute.WorkflowTask{
        {
            Name:        "preprocess",
            FunctionID:  "preprocess-function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"input_file": "${input.file_path}"},
            RetryPolicy: &compute.RetryPolicy{
                MaxRetries: 3,
                Interval:   5,
            },
        },
        {
            Name:        "analyze",
            FunctionID:  "analyze-function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"input_file": "${tasks.preprocess.output.processed_file}"},
            DependsOn:   []string{"preprocess"},
        },
        {
            Name:        "visualize",
            FunctionID:  "visualize-function-id",
            EndpointID:  "endpoint-id",
            Parameters:  map[string]interface{}{"data": "${tasks.analyze.output.results}"},
            DependsOn:   []string{"analyze"},
        },
    },
}

workflow, err := client.CreateWorkflow(ctx, workflowRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Workflow created: %s (%s)\n", workflow.Name, workflow.WorkflowID)
```

### Running a Workflow

```go
// Run a workflow
runRequest := &compute.WorkflowRunRequest{
    Input: map[string]interface{}{
        "file_path": "/path/to/data.csv",
    },
    Label: "Process data file",
}

run, err := client.RunWorkflow(ctx, "workflow-id", runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Workflow run started: %s\n", run.RunID)
```

### Getting Workflow Status

```go
// Get workflow run status
status, err := client.GetWorkflowStatus(ctx, "workflow-id", "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Workflow status: %s\n", status.Status)
fmt.Printf("Progress: %d/%d tasks completed\n", 
    status.Progress.Completed, status.Progress.Total)

// Check status of individual tasks
for taskName, taskStatus := range status.Tasks {
    fmt.Printf("Task %s: %s\n", taskName, taskStatus.Status)
    if taskStatus.Status == "success" {
        fmt.Printf("  Output: %v\n", taskStatus.Output)
    } else if taskStatus.Status == "failed" {
        fmt.Printf("  Error: %s\n", taskStatus.ErrorMessage)
    }
}
```

### Waiting for Workflow Completion

```go
// Wait for workflow to complete with timeout
status, err := client.WaitForWorkflowCompletion(
    ctx, "workflow-id", "run-id", 5*time.Minute, 5*time.Second)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Workflow did not complete within timeout")
    } else {
        fmt.Printf("Error waiting for workflow: %v\n", err)
    }
    return
}

fmt.Printf("Workflow completed with status: %s\n", status.Status)
if status.Status == "success" {
    // All tasks completed successfully
    fmt.Println("Workflow output:")
    for taskName, taskStatus := range status.Tasks {
        fmt.Printf("  %s: %v\n", taskName, taskStatus.Output)
    }
} else {
    // At least one task failed
    fmt.Println("Workflow failed:")
    for taskName, taskStatus := range status.Tasks {
        if taskStatus.Status == "failed" {
            fmt.Printf("  %s: %s\n", taskName, taskStatus.ErrorMessage)
        }
    }
}
```

## Task Group Management

### Creating a Task Group

```go
// Create a task group
taskGroupRequest := &compute.TaskGroupCreateRequest{
    Name:        "data-processing-group",
    Description: "Process multiple data files",
    FunctionID:  "function-id",
    EndpointID:  "endpoint-id",
    RetryPolicy: &compute.RetryPolicy{
        MaxRetries: 3,
        Interval:   5,
    },
}

taskGroup, err := client.CreateTaskGroup(ctx, taskGroupRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task group created: %s (%s)\n", taskGroup.Name, taskGroup.TaskGroupID)
```

### Running a Task Group

```go
// Run a task group with multiple tasks
runRequest := &compute.TaskGroupRunRequest{
    Label: "Process data files",
    Tasks: []map[string]interface{}{
        {
            "file_path": "/path/to/file1.csv",
            "options":   map[string]interface{}{"headers": true},
        },
        {
            "file_path": "/path/to/file2.csv",
            "options":   map[string]interface{}{"headers": false},
        },
        {
            "file_path": "/path/to/file3.csv",
            "options":   map[string]interface{}{"headers": true},
        },
    },
}

run, err := client.RunTaskGroup(ctx, "task-group-id", runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task group run started: %s\n", run.RunID)
```

### Getting Task Group Status

```go
// Get task group status
status, err := client.GetTaskGroupStatus(ctx, "task-group-id", "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task group status: %s\n", status.Status)
fmt.Printf("Progress: %d/%d tasks completed\n", 
    status.Progress.Completed, status.Progress.Total)
fmt.Printf("Success: %d, Failed: %d, Running: %d\n", 
    status.Progress.Success, status.Progress.Failed, status.Progress.Running)

// Check results of completed tasks
for i, taskStatus := range status.Tasks {
    fmt.Printf("Task %d: %s\n", i+1, taskStatus.Status)
    if taskStatus.Status == "success" {
        fmt.Printf("  Result: %v\n", taskStatus.Result)
    } else if taskStatus.Status == "failed" {
        fmt.Printf("  Error: %s\n", taskStatus.ErrorMessage)
    }
}
```

## Common Patterns

### Function with Dependencies

```go
// Register a function with dependencies
functionRequest := &compute.FunctionRegisterRequest{
    Name:        "data-processor",
    Description: "Process data files with pandas",
    Code: `import pandas as pd

def process_data(file_path, options=None):
    if options is None:
        options = {}
    
    # Read the data
    df = pd.read_csv(file_path, header=options.get('headers', True))
    
    # Process the data
    df = df.dropna()
    
    # Return results
    return {
        "row_count": len(df),
        "column_count": len(df.columns),
        "columns": list(df.columns),
        "summary": df.describe().to_dict()
    }`,
    Entry:       "process_data",
    Container:   "python-data-science",
}

function, err := client.RegisterFunction(ctx, functionRequest)
if err != nil {
    // Handle error
}

// Attach dependency to function
err = client.AttachDependencyToFunction(ctx, function.FunctionID, "dependency-id")
if err != nil {
    // Handle error
}

fmt.Printf("Function registered with dependencies: %s\n", function.FunctionID)
```

### Function with Environment

```go
// Create a task request with environment
taskRequest := &compute.TaskRequest{
    FunctionID:    "function-id",
    EndpointID:    "endpoint-id",
    Parameters:    map[string]interface{}{"file_path": "/path/to/data.csv"},
    Label:         "Process with environment",
    EnvironmentID: "environment-id",
}

// Run the function
taskResponse, err := client.RunFunction(ctx, taskRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task submitted with environment: %s\n", taskResponse.TaskID)
```

### Dependency Graph Execution

```go
// Create a dependency graph
graphRequest := &compute.DependencyGraphRequest{
    Nodes: []compute.DependencyGraphNode{
        {
            Name:       "extract",
            FunctionID: "extract-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "source": "${input.data_source}",
            },
        },
        {
            Name:       "transform",
            FunctionID: "transform-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${nodes.extract.output}",
            },
            DependsOn: []string{"extract"},
        },
        {
            Name:       "load",
            FunctionID: "load-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${nodes.transform.output}",
                "destination": "${input.destination}",
            },
            DependsOn: []string{"transform"},
        },
    },
    Input: map[string]interface{}{
        "data_source": "s3://bucket/data.csv",
        "destination": "database://table",
    },
}

// Run the dependency graph
graphResponse, err := client.RunDependencyGraph(ctx, graphRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency graph execution started: %s\n", graphResponse.RunID)

// Wait for completion
status, err := client.WaitForDependencyGraphCompletion(
    ctx, graphResponse.RunID, 10*time.Minute, 5*time.Second)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency graph completed with status: %s\n", status.Status)
```

## Best Practices

1. Use appropriate error handling for compute operations
2. Set reasonable timeouts for long-running operations
3. Use wait methods with appropriate polling intervals
4. Leverage containers for consistent execution environments
5. Group related tasks using task groups for efficient execution
6. Implement retry policies for transient failures
7. Use workflows for complex, multi-step processes
8. Store secrets securely using the environment secret management
9. Check endpoint status before submitting tasks
10. Monitor task status for long-running operations
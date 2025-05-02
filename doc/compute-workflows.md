# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Compute Workflows and Task Groups

Globus Compute allows you to execute functions on remote computing resources. The SDK's Compute client provides capabilities for executing individual tasks, workflows, and task groups.

## Overview

In v0.9.0, the Compute client has been enhanced with the following capabilities:

- **Workflow Management**: Create, run, and monitor complex workflows with task dependencies
- **Task Groups**: Execute parallel tasks efficiently with concurrency control
- **Dependency Graphs**: Build and execute complex workflow graphs with sophisticated dependencies
- **Environment Management**: Configure execution environments for tasks
- **Container Support**: Execute functions in containers with specific configurations

## Workflows

A workflow is a collection of tasks with defined dependencies between them. This allows you to create complex execution patterns where some tasks depend on the outputs of others.

### Creating a Workflow

```go
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

workflow, err := computeClient.CreateWorkflow(ctx, request)
```

### Listing Workflows

```go
// List all workflows
// Note: ListWorkflows currently returns a slice, not a paginated response structure
workflows, err := computeClient.ListWorkflows(ctx)
if err != nil {
    // Handle error
}

// Process workflow list
for _, workflow := range workflows {
    fmt.Printf("Workflow: %s (ID: %s)\n", workflow.Name, workflow.ID)
}

// In future, if the API is updated to use pagination similar to other services:
// workflowList, err := computeClient.ListWorkflows(ctx, &compute.ListWorkflowsOptions{
//     PerPage: 10,
//     Search: "workflow name",
// })
```

### Running a Workflow

```go
workflowRequest := &compute.WorkflowRunRequest{
    GlobalArgs: map[string]interface{}{
        "scale_factor": 2.0,
    },
    RunLabel: "Example Run",
}

runResponse, err := computeClient.RunWorkflow(ctx, workflowID, workflowRequest)
```

### Monitoring a Workflow

```go
for {
    status, err := computeClient.GetWorkflowStatus(ctx, runID)
    if err != nil {
        return err
    }
    
    fmt.Printf("Status: %s, Progress: %.1f%% (%d/%d tasks completed)\n", 
        status.Status, 
        status.Progress.PercentDone,
        status.Progress.Completed,
        status.Progress.TotalTasks)
    
    if status.Status == "completed" || status.Status == "failed" {
        break
    }
    
    time.Sleep(5 * time.Second)
}
```

## Task Groups

Task groups allow you to execute multiple tasks in parallel with control over concurrency.

### Creating a Task Group

```go
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

taskGroup, err := computeClient.CreateTaskGroup(ctx, request)
```

### Running a Task Group

```go
request := &compute.TaskGroupRunRequest{
    RunLabel: "Example Task Group Run",
}

taskGroupRun, err := computeClient.RunTaskGroup(ctx, taskGroupID, request)
```

### Monitoring a Task Group

```go
for {
    status, err := computeClient.GetTaskGroupStatus(ctx, runID)
    if err != nil {
        return err
    }
    
    fmt.Printf("Status: %s, Progress: %.1f%% (%d/%d tasks completed)\n", 
        status.Status, 
        status.Progress.PercentDone,
        status.Progress.Completed,
        status.Progress.TotalTasks)
    
    if status.Status == "completed" || status.Status == "failed" {
        break
    }
    
    time.Sleep(5 * time.Second)
}
```

## Dependency Graphs

For more complex workflows where dependencies might dynamically change, you can use the dependency graph API:

```go
// Define nodes in the dependency graph
nodes := map[string]compute.DependencyGraphNode{
    "node1": {
        Task: compute.TaskRequest{
            FunctionID: functionID,
            EndpointID: endpointID,
            Args:       []interface{}{"input1"},
        },
    },
    "node2": {
        Task: compute.TaskRequest{
            FunctionID: functionID,
            EndpointID: endpointID,
            Args:       []interface{}{"input2"},
        },
    },
    "node3": {
        Task: compute.TaskRequest{
            FunctionID: functionID,
            EndpointID: endpointID,
            Args:       []interface{}{"input3"},
        },
        Dependencies: []string{"node1", "node2"},
        Condition:    "node1.status == 'success' && node2.status == 'success'",
    },
}

request := &compute.DependencyGraphRequest{
    Nodes:       nodes,
    Description: "Example dependency graph",
    ErrorPolicy: "continue",
}

response, err := computeClient.RunDependencyGraph(ctx, request)
```

## Complete Example

For a complete example, see [examples/compute-workflow/main.go](../examples/compute-workflow/main.go).

## Notes on API Versions

Different Globus Compute deployments might support different API versions. The SDK provides version checking capabilities to ensure compatibility:

```go
// Check if the API version is compatible
version, err := ParseVersion("compute", "v2")
if err != nil {
    // Handle error
}

// Check compatibility with current client
if !version.IsCompatible(computeClient.Version) {
    // Handle incompatibility
}
```

## Performance Considerations

When working with workflows and task groups, consider:

1. **Batch operations** where possible to reduce API calls
2. **Concurrency limits** to avoid overloading endpoints
3. **Error handling strategies** that make sense for your application
4. **Retry policies** that account for transient failures

## Further Reading

- [Globus Compute Documentation](https://docs.globus.org/api/compute/)
- [Function Registration Guide](function-registration.md)
- [Container Configuration Guide](container-configuration.md)
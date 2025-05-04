# Compute Service: Batch Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Batch operations allow you to run multiple compute functions concurrently and manage complex function orchestration. The Compute service provides workflows, task groups, and dependency graphs to organize and execute computational tasks efficiently.

## Workflow Orchestration

Workflows enable you to define and execute a sequence of interdependent functions that process data through multiple steps.

### Workflow Structure

```go
type WorkflowCreateRequest struct {
    Name            string           `json:"name"`
    Description     string           `json:"description,omitempty"`
    Tasks           []WorkflowTask   `json:"tasks"`
    InputSchema     map[string]interface{} `json:"input_schema,omitempty"`
    VisibleTo       []string         `json:"visible_to,omitempty"`
    ManagingUsers   []string         `json:"managing_users,omitempty"`
    ManagingGroups  []string         `json:"managing_groups,omitempty"`
}

type WorkflowTask struct {
    Name            string                 `json:"name"`
    FunctionID      string                 `json:"function_id"`
    EndpointID      string                 `json:"endpoint_id"`
    Parameters      map[string]interface{} `json:"parameters,omitempty"`
    DependsOn       []string               `json:"depends_on,omitempty"`
    RetryPolicy     *RetryPolicy           `json:"retry_policy,omitempty"`
    EnvironmentID   string                 `json:"environment_id,omitempty"`
}

type RetryPolicy struct {
    MaxRetries      int    `json:"max_retries"`
    Interval        int    `json:"interval"`
    BackoffFactor   float64 `json:"backoff_factor,omitempty"`
    MaxInterval     int    `json:"max_interval,omitempty"`
}
```

### Creating a Workflow

To create a new workflow with multiple tasks:

```go
// Create a data processing workflow
workflowRequest := &compute.WorkflowCreateRequest{
    Name:        "data-processing-pipeline",
    Description: "Pipeline for data extraction, transformation, and loading",
    Tasks: []compute.WorkflowTask{
        {
            Name:       "extract",
            FunctionID: "extract-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "source": "${input.data_source}",
                "format": "${input.format}",
            },
            RetryPolicy: &compute.RetryPolicy{
                MaxRetries: 3,
                Interval:   10, // seconds
            },
        },
        {
            Name:       "transform",
            FunctionID: "transform-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${tasks.extract.output}",
                "transformations": "${input.transformations}",
            },
            DependsOn: []string{"extract"},
            RetryPolicy: &compute.RetryPolicy{
                MaxRetries:    3,
                Interval:      10,
                BackoffFactor: 2.0,
                MaxInterval:   60,
            },
        },
        {
            Name:       "load",
            FunctionID: "load-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${tasks.transform.output}",
                "destination": "${input.destination}",
            },
            DependsOn: []string{"transform"},
        },
    ],
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{"data_source", "destination"},
        "properties": map[string]interface{}{
            "data_source": map[string]interface{}{
                "type": "string",
                "description": "URL or path to the data source",
            },
            "format": map[string]interface{}{
                "type": "string",
                "enum": []string{"csv", "json", "parquet"},
                "default": "csv",
            },
            "transformations": map[string]interface{}{
                "type": "array",
                "items": map[string]interface{}{
                    "type": "object",
                },
            },
            "destination": map[string]interface{}{
                "type": "string",
                "description": "Destination for the processed data",
            },
        },
    },
    VisibleTo:      []string{"public"},
    ManagingUsers:  []string{"user@example.com"},
    ManagingGroups: []string{"admin-group-id"},
}

// Create the workflow
workflow, err := client.CreateWorkflow(ctx, workflowRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Workflow created: %s (%s)\n", workflow.Name, workflow.WorkflowID)
```

### Listing Workflows

To list available workflows:

```go
// List workflows
workflows, err := client.ListWorkflows(ctx, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d workflows\n", len(workflows.Workflows))
for _, wf := range workflows.Workflows {
    fmt.Printf("- %s (%s)\n", wf.Name, wf.WorkflowID)
    fmt.Printf("  Description: %s\n", wf.Description)
    fmt.Printf("  Tasks: %d\n", len(wf.Tasks))
}
```

### Getting a Workflow

To retrieve details of a specific workflow:

```go
// Get a specific workflow
workflow, err := client.GetWorkflow(ctx, "workflow-id")
if err != nil {
    // Handle error
}

fmt.Printf("Workflow: %s (%s)\n", workflow.Name, workflow.WorkflowID)
fmt.Printf("Description: %s\n", workflow.Description)
fmt.Printf("Created: %s\n", workflow.CreatedTimestamp)
fmt.Printf("Owner: %s\n", workflow.Owner)

// List tasks
fmt.Println("Tasks:")
for _, task := range workflow.Tasks {
    fmt.Printf("- %s (Function: %s)\n", task.Name, task.FunctionID)
    if len(task.DependsOn) > 0 {
        fmt.Printf("  Depends on: %v\n", task.DependsOn)
    }
}
```

### Running a Workflow

To execute a workflow:

```go
// Create a workflow run request
runRequest := &compute.WorkflowRunRequest{
    Input: map[string]interface{}{
        "data_source": "s3://bucket/data.csv",
        "format": "csv",
        "transformations": []map[string]interface{}{
            {
                "type": "filter",
                "column": "status",
                "value": "active",
            },
            {
                "type": "calculate",
                "column": "total",
                "formula": "price * quantity",
            },
        },
        "destination": "s3://bucket/processed-data/",
    },
    Label: "Process sales data",
}

// Run the workflow
run, err := client.RunWorkflow(ctx, "workflow-id", runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Workflow run started: %s\n", run.RunID)
```

### Getting Workflow Run Status

To check the status of a workflow run:

```go
// Check workflow run status
status, err := client.GetWorkflowStatus(ctx, "workflow-id", "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Workflow status: %s\n", status.Status)
fmt.Printf("Started: %s\n", status.StartTime)
if status.EndTime != "" {
    fmt.Printf("Ended: %s\n", status.EndTime)
}

// Check progress
fmt.Printf("Progress: %d/%d tasks completed\n", 
    status.Progress.Completed, status.Progress.Total)
fmt.Printf("Success: %d, Failed: %d, Running: %d\n", 
    status.Progress.Success, status.Progress.Failed, status.Progress.Running)

// Check individual task status
for taskName, taskStatus := range status.Tasks {
    fmt.Printf("Task %s: %s\n", taskName, taskStatus.Status)
    if taskStatus.Status == "success" {
        fmt.Printf("  Completed: %s\n", taskStatus.EndTime)
        // Output is available in taskStatus.Output
    } else if taskStatus.Status == "failed" {
        fmt.Printf("  Failed: %s\n", taskStatus.ErrorMessage)
    }
}
```

### Waiting for Workflow Completion

For convenience, the SDK provides a method to wait for workflow completion:

```go
// Wait for workflow to complete (timeout after 10 minutes, check every 5 seconds)
status, err := client.WaitForWorkflowCompletion(
    ctx, "workflow-id", "run-id", 10*time.Minute, 5*time.Second)
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
    fmt.Println("Workflow output:")
    // Access outputs from each task
    for taskName, taskStatus := range status.Tasks {
        fmt.Printf("  %s output: %v\n", taskName, taskStatus.Output)
    }
} else {
    fmt.Println("Workflow failed:")
    // Find failed tasks
    for taskName, taskStatus := range status.Tasks {
        if taskStatus.Status == "failed" {
            fmt.Printf("  %s failed: %s\n", taskName, taskStatus.ErrorMessage)
        }
    }
}
```

### Canceling a Workflow Run

To cancel a running workflow:

```go
// Cancel a workflow run
err := client.CancelWorkflowRun(ctx, "workflow-id", "run-id")
if err != nil {
    // Handle error
}

fmt.Println("Workflow run canceled")
```

## Task Groups

Task groups allow you to run multiple instances of the same function with different parameters.

### Task Group Structure

```go
type TaskGroupCreateRequest struct {
    Name           string         `json:"name"`
    Description    string         `json:"description,omitempty"`
    FunctionID     string         `json:"function_id"`
    EndpointID     string         `json:"endpoint_id"`
    RetryPolicy    *RetryPolicy   `json:"retry_policy,omitempty"`
    EnvironmentID  string         `json:"environment_id,omitempty"`
    VisibleTo      []string       `json:"visible_to,omitempty"`
    ManagingUsers  []string       `json:"managing_users,omitempty"`
    ManagingGroups []string       `json:"managing_groups,omitempty"`
}
```

### Creating a Task Group

To create a task group for processing multiple files:

```go
// Create a task group
taskGroupRequest := &compute.TaskGroupCreateRequest{
    Name:        "batch-file-processor",
    Description: "Process multiple data files",
    FunctionID:  "file-processing-function-id",
    EndpointID:  "endpoint-id",
    RetryPolicy: &compute.RetryPolicy{
        MaxRetries:    3,
        Interval:      30,
        BackoffFactor: 2.0,
    },
    EnvironmentID:  "environment-id", // Optional environment
    VisibleTo:      []string{"public"},
    ManagingUsers:  []string{"user@example.com"},
    ManagingGroups: []string{},
}

// Create the task group
taskGroup, err := client.CreateTaskGroup(ctx, taskGroupRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task group created: %s (%s)\n", taskGroup.Name, taskGroup.TaskGroupID)
```

### Running a Task Group

To run a task group with multiple input parameter sets:

```go
// Create a task group run request
runRequest := &compute.TaskGroupRunRequest{
    Label: "Process data files batch",
    Tasks: []map[string]interface{}{
        {
            "file_path": "/path/to/file1.csv",
            "options": map[string]interface{}{
                "headers": true,
                "delimiter": ",",
            },
        },
        {
            "file_path": "/path/to/file2.csv",
            "options": map[string]interface{}{
                "headers": false,
                "delimiter": ";",
            },
        },
        {
            "file_path": "/path/to/file3.csv",
            "options": map[string]interface{}{
                "headers": true,
                "delimiter": "\t",
            },
        },
    },
}

// Run the task group
run, err := client.RunTaskGroup(ctx, "task-group-id", runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task group run started: %s\n", run.RunID)
```

### Getting Task Group Status

To check the status of a task group run:

```go
// Check task group status
status, err := client.GetTaskGroupStatus(ctx, "task-group-id", "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Task group status: %s\n", status.Status)
fmt.Printf("Started: %s\n", status.StartTime)
if status.EndTime != "" {
    fmt.Printf("Ended: %s\n", status.EndTime)
}

// Check progress
fmt.Printf("Progress: %d/%d tasks completed\n", 
    status.Progress.Completed, status.Progress.Total)
fmt.Printf("Success: %d, Failed: %d, Running: %d\n", 
    status.Progress.Success, status.Progress.Failed, status.Progress.Running)

// Check individual task results
for i, taskStatus := range status.Tasks {
    fmt.Printf("Task %d: %s\n", i+1, taskStatus.Status)
    if taskStatus.Status == "success" {
        fmt.Printf("  Completed: %s\n", taskStatus.EndTime)
        fmt.Printf("  Result: %v\n", taskStatus.Result)
    } else if taskStatus.Status == "failed" {
        fmt.Printf("  Failed: %s\n", taskStatus.ErrorMessage)
    }
}
```

### Waiting for Task Group Completion

Similar to workflows, you can wait for a task group to complete:

```go
// Wait for task group to complete
status, err := client.WaitForTaskGroupCompletion(
    ctx, "task-group-id", "run-id", 10*time.Minute, 5*time.Second)
if err != nil {
    // Handle error
    return
}

fmt.Printf("Task group completed with status: %s\n", status.Status)

// Get successful and failed tasks
successCount := 0
failedCount := 0
for i, taskStatus := range status.Tasks {
    if taskStatus.Status == "success" {
        successCount++
    } else if taskStatus.Status == "failed" {
        failedCount++
        fmt.Printf("Task %d failed: %s\n", i+1, taskStatus.ErrorMessage)
    }
}

fmt.Printf("Tasks: %d successful, %d failed\n", successCount, failedCount)
```

## Dependency Graphs

Dependency graphs allow you to create an ad-hoc directed acyclic graph (DAG) of tasks without creating a permanent workflow.

### Dependency Graph Structure

```go
type DependencyGraphRequest struct {
    Nodes []DependencyGraphNode         `json:"nodes"`
    Input map[string]interface{}        `json:"input,omitempty"`
    Label string                        `json:"label,omitempty"`
}

type DependencyGraphNode struct {
    Name         string                 `json:"name"`
    FunctionID   string                 `json:"function_id"`
    EndpointID   string                 `json:"endpoint_id"`
    Parameters   map[string]interface{} `json:"parameters,omitempty"`
    DependsOn    []string               `json:"depends_on,omitempty"`
    RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
    ErrorHandler *ErrorHandler          `json:"error_handler,omitempty"`
}

type ErrorHandler struct {
    FunctionID string                 `json:"function_id"`
    EndpointID string                 `json:"endpoint_id"`
    Parameters map[string]interface{} `json:"parameters,omitempty"`
}
```

### Running a Dependency Graph

To execute a one-time dependency graph:

```go
// Create a dependency graph request
graphRequest := &compute.DependencyGraphRequest{
    Nodes: []compute.DependencyGraphNode{
        {
            Name:       "download",
            FunctionID: "download-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "url": "${input.url}",
                "output_path": "/tmp/downloaded_file",
            },
            RetryPolicy: &compute.RetryPolicy{
                MaxRetries: 3,
                Interval:   10,
            },
            ErrorHandler: &compute.ErrorHandler{
                FunctionID: "error-notification-function-id",
                EndpointID: "endpoint-id",
                Parameters: map[string]interface{}{
                    "error": "${error}",
                    "email": "${input.notification_email}",
                },
            },
        },
        {
            Name:       "process",
            FunctionID: "process-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input_file": "/tmp/downloaded_file",
                "output_file": "/tmp/processed_file",
            },
            DependsOn: []string{"download"},
        },
        {
            Name:       "analyze",
            FunctionID: "analyze-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input_file": "/tmp/processed_file",
            },
            DependsOn: []string{"process"},
        },
        {
            Name:       "notify",
            FunctionID: "notify-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "result": "${nodes.analyze.output}",
                "email": "${input.notification_email}",
            },
            DependsOn: []string{"analyze"},
        },
    },
    Input: map[string]interface{}{
        "url": "https://example.com/data.csv",
        "notification_email": "user@example.com",
    },
    Label: "One-time data processing pipeline",
}

// Run the dependency graph
graphResponse, err := client.RunDependencyGraph(ctx, graphRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency graph execution started: %s\n", graphResponse.RunID)
```

### Getting Dependency Graph Status

To check the status of a dependency graph:

```go
// Check dependency graph status
status, err := client.GetDependencyGraphStatus(ctx, graphResponse.RunID)
if err != nil {
    // Handle error
}

fmt.Printf("Dependency graph status: %s\n", status.Status)
fmt.Printf("Started: %s\n", status.StartTime)
if status.EndTime != "" {
    fmt.Printf("Ended: %s\n", status.EndTime)
}

// Check progress
fmt.Printf("Progress: %d/%d nodes completed\n", 
    status.Progress.Completed, status.Progress.Total)
fmt.Printf("Success: %d, Failed: %d, Running: %d\n", 
    status.Progress.Success, status.Progress.Failed, status.Progress.Running)

// Check individual node status
for nodeName, nodeStatus := range status.Nodes {
    fmt.Printf("Node %s: %s\n", nodeName, nodeStatus.Status)
    if nodeStatus.Status == "success" {
        fmt.Printf("  Completed: %s\n", nodeStatus.EndTime)
        fmt.Printf("  Output: %v\n", nodeStatus.Output)
    } else if nodeStatus.Status == "failed" {
        fmt.Printf("  Failed: %s\n", nodeStatus.ErrorMessage)
        if nodeStatus.ErrorHandlerStatus != nil {
            fmt.Printf("  Error handler: %s\n", nodeStatus.ErrorHandlerStatus.Status)
        }
    }
}
```

### Waiting for Dependency Graph Completion

To wait for a dependency graph to complete:

```go
// Wait for dependency graph to complete
status, err := client.WaitForDependencyGraphCompletion(
    ctx, graphResponse.RunID, 15*time.Minute, 5*time.Second)
if err != nil {
    // Handle error
    return
}

fmt.Printf("Dependency graph completed with status: %s\n", status.Status)

// Get the final outputs from successful nodes
if status.Status == "success" {
    fmt.Println("Node outputs:")
    for nodeName, nodeStatus := range status.Nodes {
        fmt.Printf("  %s: %v\n", nodeName, nodeStatus.Output)
    }
}
```

## Batch Function Execution

For simple batch operations without task dependencies, you can use the batch functionality:

### Running a Batch of Functions

```go
// Create a batch request
batchRequest := &compute.BatchTaskRequest{
    Tasks: []compute.TaskRequest{
        {
            FunctionID: "function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input": "data1",
                "option": "value1",
            },
            Label: "Task 1",
        },
        {
            FunctionID: "function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input": "data2",
                "option": "value2",
            },
            Label: "Task 2",
        },
        {
            FunctionID: "function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input": "data3",
                "option": "value3",
            },
            Label: "Task 3",
        },
    },
}

// Run the batch
batchResponse, err := client.RunBatch(ctx, batchRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Batch with %d tasks submitted\n", len(batchResponse.TaskIDs))
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

fmt.Printf("Batch status retrieved for %d tasks\n", len(batchStatus.TaskStatuses))

// Check individual task status
for taskID, status := range batchStatus.TaskStatuses {
    fmt.Printf("Task %s: %s\n", taskID, status.Status)
    if status.Status == "success" {
        fmt.Printf("  Result: %v\n", status.Result)
    } else if status.Status == "failed" {
        fmt.Printf("  Error: %s\n", status.ErrorMessage)
    }
}
```

## Common Patterns

### Parallel Data Processing

Process multiple datasets in parallel using task groups:

```go
// Create a task group for data processing
taskGroupRequest := &compute.TaskGroupCreateRequest{
    Name:        "dataset-processor",
    Description: "Process multiple datasets in parallel",
    FunctionID:  "process-dataset-function-id",
    EndpointID:  "endpoint-id",
    RetryPolicy: &compute.RetryPolicy{
        MaxRetries: 2,
        Interval:   30,
    },
}

taskGroup, err := client.CreateTaskGroup(ctx, taskGroupRequest)
if err != nil {
    // Handle error
}

// Run the task group with multiple datasets
datasets := []string{
    "dataset1.csv",
    "dataset2.csv",
    "dataset3.csv",
    "dataset4.csv",
    "dataset5.csv",
}

var tasks []map[string]interface{}
for _, dataset := range datasets {
    tasks = append(tasks, map[string]interface{}{
        "dataset": dataset,
        "options": map[string]interface{}{
            "normalize": true,
            "remove_outliers": true,
        },
    })
}

runRequest := &compute.TaskGroupRunRequest{
    Label: "Batch dataset processing",
    Tasks: tasks,
}

run, err := client.RunTaskGroup(ctx, taskGroup.TaskGroupID, runRequest)
if err != nil {
    // Handle error
}

// Wait for completion and collect results
status, err := client.WaitForTaskGroupCompletion(
    ctx, taskGroup.TaskGroupID, run.RunID, 30*time.Minute, 10*time.Second)
if err != nil {
    // Handle error
}

// Process the results
for i, taskStatus := range status.Tasks {
    if taskStatus.Status == "success" {
        fmt.Printf("Dataset %s processed successfully\n", datasets[i])
        // Use taskStatus.Result
    } else {
        fmt.Printf("Failed to process dataset %s: %s\n", 
            datasets[i], taskStatus.ErrorMessage)
    }
}
```

### ETL Pipeline

Create an Extract, Transform, Load (ETL) workflow:

```go
// Create an ETL workflow
etlWorkflow := &compute.WorkflowCreateRequest{
    Name:        "etl-pipeline",
    Description: "Extract, Transform, and Load data",
    Tasks: []compute.WorkflowTask{
        {
            Name:       "extract",
            FunctionID: "extract-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "source": "${input.source}",
                "credentials": "${input.credentials}",
            },
        },
        {
            Name:       "transform",
            FunctionID: "transform-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${tasks.extract.output}",
                "transformations": "${input.transformations}",
            },
            DependsOn: []string{"extract"},
        },
        {
            Name:       "load",
            FunctionID: "load-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "data": "${tasks.transform.output}",
                "destination": "${input.destination}",
                "credentials": "${input.credentials}",
            },
            DependsOn: []string{"transform"},
        },
        {
            Name:       "notify",
            FunctionID: "notify-function-id",
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "email": "${input.notification_email}",
                "message": "ETL pipeline completed successfully",
                "details": {
                    "source": "${input.source}",
                    "destination": "${input.destination}",
                    "records_processed": "${tasks.transform.output.record_count}",
                },
            },
            DependsOn: []string{"load"},
        },
    },
}

workflow, err := client.CreateWorkflow(ctx, etlWorkflow)
if err != nil {
    // Handle error
}

// Run the ETL workflow
runRequest := &compute.WorkflowRunRequest{
    Input: map[string]interface{}{
        "source": "database://source_db/table",
        "destination": "database://dest_db/table",
        "credentials": {
            "username": "db_user",
            "password": "db_password",
        },
        "transformations": []map[string]interface{}{
            {"type": "filter", "field": "status", "value": "active"},
            {"type": "aggregate", "field": "amount", "operation": "sum", "group_by": "customer_id"},
        },
        "notification_email": "user@example.com",
    },
    Label: "Monthly data aggregation",
}

run, err := client.RunWorkflow(ctx, workflow.WorkflowID, runRequest)
if err != nil {
    // Handle error
}

// Wait for the ETL workflow to complete
status, err := client.WaitForWorkflowCompletion(
    ctx, workflow.WorkflowID, run.RunID, 1*time.Hour, 1*time.Minute)
if err != nil {
    // Handle error
}
```

### Fan-out, Fan-in Pattern

Implement a fan-out, fan-in pattern with dependency graphs:

```go
// Create a function to generate a fan-out, fan-in dependency graph
func createFanOutFanInGraph(inputData []string, processingFunction, aggregationFunction string) *compute.DependencyGraphRequest {
    // Create the fan-out nodes
    var nodes []compute.DependencyGraphNode
    var processingNodes []string
    
    // Add the processing nodes (fan-out)
    for i, data := range inputData {
        nodeName := fmt.Sprintf("process_%d", i)
        processingNodes = append(processingNodes, nodeName)
        
        nodes = append(nodes, compute.DependencyGraphNode{
            Name:       nodeName,
            FunctionID: processingFunction,
            EndpointID: "endpoint-id",
            Parameters: map[string]interface{}{
                "input": data,
            },
        })
    }
    
    // Add the aggregation node (fan-in)
    aggregationParams := map[string]interface{}{}
    for i, nodeName := range processingNodes {
        aggregationParams[fmt.Sprintf("result_%d", i)] = fmt.Sprintf("${nodes.%s.output}", nodeName)
    }
    
    nodes = append(nodes, compute.DependencyGraphNode{
        Name:       "aggregate",
        FunctionID: aggregationFunction,
        EndpointID: "endpoint-id",
        Parameters: aggregationParams,
        DependsOn:  processingNodes,
    })
    
    // Create the graph request
    return &compute.DependencyGraphRequest{
        Nodes: nodes,
        Label: "Fan-out, Fan-in Processing",
    }
}

// Use the function to create and run a graph
inputData := []string{"data1", "data2", "data3", "data4", "data5"}
graphRequest := createFanOutFanInGraph(inputData, "processing-function-id", "aggregation-function-id")

// Run the graph
graphResponse, err := client.RunDependencyGraph(ctx, graphRequest)
if err != nil {
    // Handle error
}

// Wait for completion
status, err := client.WaitForDependencyGraphCompletion(
    ctx, graphResponse.RunID, 30*time.Minute, 10*time.Second)
if err != nil {
    // Handle error
}

// Get the aggregated result
if status.Status == "success" {
    aggregateNode := status.Nodes["aggregate"]
    fmt.Printf("Aggregated result: %v\n", aggregateNode.Output)
}
```

## Best Practices

1. **Define Clear Task Boundaries**: Each task should have a clear, single responsibility
2. **Implement Error Handling**: Use retry policies and error handlers for resilience
3. **Manage Data Flow**: Use parameters and input/output mappings to control data flow
4. **Use Appropriate Batch Type**:
   - Use task groups for parallel, independent executions of the same function
   - Use workflows for complex, dependent task sequences that will be reused
   - Use dependency graphs for one-time, ad-hoc task dependencies
   - Use batch execution for simple parallel function calls
5. **Set Realistic Timeouts**: Provide appropriate timeouts when waiting for completion
6. **Monitor Progress**: Track task progress and handle timeouts gracefully
7. **Validate Inputs**: Define input schemas for workflows to catch invalid inputs early
8. **Plan for Scalability**: Design batch operations to handle variable workloads
9. **Control Access**: Use visibility and management permissions appropriately
10. **Test with Small Batches**: Validate your approach with small batches before scaling up
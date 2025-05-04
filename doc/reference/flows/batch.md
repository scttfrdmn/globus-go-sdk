# Flows Service: Batch Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Batch operations allow you to perform multiple actions concurrently for improved efficiency. The Flows service provides batch methods for common operations like running flows, retrieving run information, and canceling runs.

## Batch Options

All batch operations accept options for configuring their behavior:

```go
type BatchOptions struct {
    Concurrency int // Maximum number of concurrent operations
}
```

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Concurrency` | `int` | Maximum concurrent operations | 10 |

## Batch Run Flows

Start multiple flow runs concurrently:

```go
// Create batch runs request
batchRequest := &flows.BatchRunFlowsRequest{
    Runs: []flows.RunRequest{
        {
            FlowID: "flow-id",
            Label:  "Batch Job 1",
            Input: map[string]interface{}{
                "param1": "value1",
            },
        },
        {
            FlowID: "flow-id",
            Label:  "Batch Job 2",
            Input: map[string]interface{}{
                "param1": "value2",
            },
        },
        {
            FlowID: "different-flow-id",
            Label:  "Batch Job 3",
            Input: map[string]interface{}{
                "param1": "value3",
            },
        },
    },
    Options: &flows.BatchOptions{
        Concurrency: 3, // Run 3 flows concurrently
    },
}

// Submit batch runs
batchResponse, err := client.BatchRunFlows(ctx, batchRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Started %d runs\n", len(batchResponse.Results))

// Check for individual run errors
for i, result := range batchResponse.Results {
    if result.Error != nil {
        fmt.Printf("Error starting run %d: %v\n", i+1, result.Error)
    } else {
        fmt.Printf("Run %d started: %s (Status: %s)\n", 
            i+1, result.Run.RunID, result.Run.Status)
    }
}
```

### BatchRunFlowsRequest

```go
type BatchRunFlowsRequest struct {
    Runs    []RunRequest  // List of run requests
    Options *BatchOptions // Batch operation options
}
```

### BatchRunFlowsResponse

```go
type BatchRunFlowsResponse struct {
    Results []BatchRunFlowResult // Results for each run request
}

type BatchRunFlowResult struct {
    Run   *RunResponse // Run information (if successful)
    Error error        // Error (if failed)
}
```

## Batch Get Runs

Retrieve information for multiple runs concurrently:

```go
// Create batch get runs request
batchRequest := &flows.BatchRunsRequest{
    RunIDs: []string{"run-id-1", "run-id-2", "run-id-3"},
    Options: &flows.BatchOptions{
        Concurrency: 3, // Get 3 runs concurrently
    },
}

// Submit batch get runs
batchResponse, err := client.BatchGetRuns(ctx, batchRequest)
if err != nil {
    // Handle error
}

// Process results
for _, result := range batchResponse.Results {
    if result.Error != nil {
        fmt.Printf("Error getting run %s: %v\n", result.RunID, result.Error)
    } else {
        fmt.Printf("Run %s status: %s\n", result.Run.RunID, result.Run.Status)
        if result.Run.Status == "SUCCEEDED" {
            fmt.Printf("  Output: %v\n", result.Run.Output)
        } else if result.Run.Status == "FAILED" {
            fmt.Printf("  Error: %s - %s\n", 
                result.Run.RunError.Code, result.Run.RunError.Description)
        }
    }
}
```

### BatchRunsRequest

```go
type BatchRunsRequest struct {
    RunIDs  []string      // List of run IDs
    Options *BatchOptions // Batch operation options
}
```

### BatchRunsResponse

```go
type BatchRunsResponse struct {
    Results []BatchRunResult // Results for each run ID
}

type BatchRunResult struct {
    RunID string        // Run ID
    Run   *RunResponse  // Run information (if successful)
    Error error         // Error (if failed)
}
```

## Batch Cancel Runs

Cancel multiple runs concurrently:

```go
// Create batch cancel runs request
batchRequest := &flows.BatchCancelRunsRequest{
    RunIDs: []string{"run-id-1", "run-id-2", "run-id-3"},
    Options: &flows.BatchOptions{
        Concurrency: 3, // Cancel 3 runs concurrently
    },
}

// Submit batch cancellation
batchResponse, err := client.BatchCancelRuns(ctx, batchRequest)
if err != nil {
    // Handle error
}

// Check for individual cancellation errors
for _, result := range batchResponse.Results {
    if result.Error != nil {
        fmt.Printf("Error canceling run %s: %v\n", result.RunID, result.Error)
    } else {
        fmt.Printf("Run %s canceled\n", result.RunID)
    }
}
```

### BatchCancelRunsRequest

```go
type BatchCancelRunsRequest struct {
    RunIDs  []string      // List of run IDs to cancel
    Options *BatchOptions // Batch operation options
}
```

### BatchCancelRunsResponse

```go
type BatchCancelRunsResponse struct {
    Results []BatchCancelRunResult // Results for each run ID
}

type BatchCancelRunResult struct {
    RunID string // Run ID
    Error error  // Error (if failed)
}
```

## Batch Get Flows

Retrieve information for multiple flows concurrently:

```go
// Create batch get flows request
batchRequest := &flows.BatchFlowsRequest{
    FlowIDs: []string{"flow-id-1", "flow-id-2", "flow-id-3"},
    Options: &flows.BatchOptions{
        Concurrency: 3, // Get 3 flows concurrently
    },
}

// Submit batch get flows
batchResponse, err := client.BatchGetFlows(ctx, batchRequest)
if err != nil {
    // Handle error
}

// Process results
for _, result := range batchResponse.Results {
    if result.Error != nil {
        fmt.Printf("Error getting flow %s: %v\n", result.FlowID, result.Error)
    } else {
        fmt.Printf("Flow %s: %s\n", result.Flow.ID, result.Flow.Title)
        fmt.Printf("  Description: %s\n", result.Flow.Description)
        fmt.Printf("  Owner: %s\n", result.Flow.OwnerString)
    }
}
```

### BatchFlowsRequest

```go
type BatchFlowsRequest struct {
    FlowIDs []string      // List of flow IDs
    Options *BatchOptions // Batch operation options
}
```

### BatchFlowsResponse

```go
type BatchFlowsResponse struct {
    Results []BatchFlowResult // Results for each flow ID
}

type BatchFlowResult struct {
    FlowID string  // Flow ID
    Flow   *Flow   // Flow information (if successful)
    Error  error   // Error (if failed)
}
```

## Batch Get Action Roles

Retrieve information for multiple action roles concurrently:

```go
// Create batch get action roles request
batchRequest := &flows.BatchActionRoleRequest{
    ProviderID: "provider-id",
    RoleIDs:    []string{"role-id-1", "role-id-2", "role-id-3"},
    Options: &flows.BatchOptions{
        Concurrency: 3, // Get 3 roles concurrently
    },
}

// Submit batch get action roles
batchResponse, err := client.BatchGetActionRoles(ctx, batchRequest)
if err != nil {
    // Handle error
}

// Process results
for _, result := range batchResponse.Results {
    if result.Error != nil {
        fmt.Printf("Error getting role %s: %v\n", result.RoleID, result.Error)
    } else {
        fmt.Printf("Role %s: %s\n", result.Role.ID, result.Role.Name)
        fmt.Printf("  Description: %s\n", result.Role.Description)
        fmt.Printf("  Required scopes: %v\n", result.Role.RequiredScopes)
    }
}
```

### BatchActionRoleRequest

```go
type BatchActionRoleRequest struct {
    ProviderID string        // Action provider ID
    RoleIDs    []string      // List of role IDs
    Options    *BatchOptions // Batch operation options
}
```

### BatchActionRoleResponse

```go
type BatchActionRoleResponse struct {
    Results []BatchActionRoleResult // Results for each role ID
}

type BatchActionRoleResult struct {
    RoleID string      // Role ID
    Role   *ActionRole // Role information (if successful)
    Error  error       // Error (if failed)
}
```

## Common Batch Patterns

### Parallel Flow Processing

Run the same flow multiple times with different inputs:

```go
// Generate multiple run requests for the same flow with different inputs
var runRequests []flows.RunRequest
for i := 1; i <= 10; i++ {
    runRequests = append(runRequests, flows.RunRequest{
        FlowID: "flow-id",
        Label:  fmt.Sprintf("Process File %d", i),
        Input: map[string]interface{}{
            "file_path": fmt.Sprintf("/path/to/file_%d.txt", i),
            "options": map[string]interface{}{
                "parameter": fmt.Sprintf("value_%d", i),
            },
        },
    })
}

// Create batch runs request
batchRequest := &flows.BatchRunFlowsRequest{
    Runs:    runRequests,
    Options: &flows.BatchOptions{
        Concurrency: 5, // Process 5 files concurrently
    },
}

// Submit batch runs
batchResponse, err := client.BatchRunFlows(ctx, batchRequest)
if err != nil {
    // Handle error
}

// Store run IDs for monitoring
var runIDs []string
for _, result := range batchResponse.Results {
    if result.Error == nil {
        runIDs = append(runIDs, result.Run.RunID)
    }
}

fmt.Printf("Started %d runs\n", len(runIDs))
```

### Monitoring Multiple Runs

Check the status of multiple runs concurrently:

```go
// Function to monitor runs until all are complete
func monitorRuns(ctx context.Context, client *flows.Client, runIDs []string) {
    // Track run status
    completed := make(map[string]bool)
    for _, id := range runIDs {
        completed[id] = false
    }
    
    // Continue until all runs complete
    for {
        // Create a batch request to check status
        batchRequest := &flows.BatchRunsRequest{
            RunIDs:  runIDs,
            Options: &flows.BatchOptions{
                Concurrency: 10,
            },
        }
        
        batchResponse, err := client.BatchGetRuns(ctx, batchRequest)
        if err != nil {
            fmt.Printf("Error getting run status: %v\n", err)
            return
        }
        
        // Check status of each run
        totalCompleted := 0
        for _, result := range batchResponse.Results {
            if result.Error != nil {
                fmt.Printf("Error getting run %s: %v\n", result.RunID, result.Error)
                continue
            }
            
            run := result.Run
            status := run.Status
            
            // Mark complete if terminal status
            if status == "SUCCEEDED" || status == "FAILED" || status == "CANCELED" {
                if !completed[run.RunID] {
                    fmt.Printf("Run %s completed with status: %s\n", run.RunID, status)
                    if status == "FAILED" && run.RunError != nil {
                        fmt.Printf("  Error: %s - %s\n", 
                            run.RunError.Code, run.RunError.Description)
                    }
                }
                completed[run.RunID] = true
            } else {
                fmt.Printf("Run %s status: %s\n", run.RunID, status)
            }
        }
        
        // Count completed runs
        for _, isComplete := range completed {
            if isComplete {
                totalCompleted++
            }
        }
        
        fmt.Printf("Progress: %d/%d runs completed\n", totalCompleted, len(runIDs))
        
        // Exit if all runs are complete
        if totalCompleted == len(runIDs) {
            fmt.Println("All runs completed!")
            return
        }
        
        // Wait before checking again
        select {
        case <-ctx.Done():
            fmt.Println("Monitoring canceled")
            return
        case <-time.After(10 * time.Second):
            // Continue checking
        }
    }
}

// Use the monitoring function
go monitorRuns(ctx, client, runIDs)
```

### Canceling Failed Runs

Cancel all runs that have been running too long:

```go
// Function to cancel runs that have been running for too long
func cancelLongRunningRuns(ctx context.Context, client *flows.Client, runIDs []string, maxDuration time.Duration) {
    // Get run information
    batchGetRequest := &flows.BatchRunsRequest{
        RunIDs:  runIDs,
        Options: &flows.BatchOptions{
            Concurrency: 10,
        },
    }
    
    batchGetResponse, err := client.BatchGetRuns(ctx, batchGetRequest)
    if err != nil {
        fmt.Printf("Error getting runs: %v\n", err)
        return
    }
    
    // Collect run IDs that need to be canceled
    var runsToCancel []string
    for _, result := range batchGetResponse.Results {
        if result.Error != nil {
            continue
        }
        
        run := result.Run
        
        // Check if the run is still active
        if run.Status != "ACTIVE" && run.Status != "INACTIVE" {
            continue
        }
        
        // Parse start time
        startTime, err := time.Parse(time.RFC3339, run.StartTime)
        if err != nil {
            fmt.Printf("Error parsing start time for run %s: %v\n", run.RunID, err)
            continue
        }
        
        // Check if the run has been running too long
        duration := time.Since(startTime)
        if duration > maxDuration {
            fmt.Printf("Run %s has been running for %s (max: %s), will cancel\n", 
                run.RunID, duration, maxDuration)
            runsToCancel = append(runsToCancel, run.RunID)
        }
    }
    
    // If no runs need to be canceled, we're done
    if len(runsToCancel) == 0 {
        fmt.Println("No runs need to be canceled")
        return
    }
    
    // Cancel the runs
    batchCancelRequest := &flows.BatchCancelRunsRequest{
        RunIDs:  runsToCancel,
        Options: &flows.BatchOptions{
            Concurrency: 10,
        },
    }
    
    batchCancelResponse, err := client.BatchCancelRuns(ctx, batchCancelRequest)
    if err != nil {
        fmt.Printf("Error canceling runs: %v\n", err)
        return
    }
    
    // Check the results
    canceledCount := 0
    for _, result := range batchCancelResponse.Results {
        if result.Error != nil {
            fmt.Printf("Error canceling run %s: %v\n", result.RunID, result.Error)
        } else {
            canceledCount++
        }
    }
    
    fmt.Printf("Successfully canceled %d runs\n", canceledCount)
}

// Use the function to cancel runs that have been running for more than 1 hour
go cancelLongRunningRuns(ctx, client, runIDs, 1*time.Hour)
```

### Flow Information Collection

Retrieve information about multiple flows efficiently:

```go
// Collect information about multiple flows
func collectFlowInfo(ctx context.Context, client *flows.Client, flowIDs []string) map[string]*flows.Flow {
    // Create batch request
    batchRequest := &flows.BatchFlowsRequest{
        FlowIDs: flowIDs,
        Options: &flows.BatchOptions{
            Concurrency: 10,
        },
    }
    
    // Submit batch request
    batchResponse, err := client.BatchGetFlows(ctx, batchRequest)
    if err != nil {
        fmt.Printf("Error getting flows: %v\n", err)
        return nil
    }
    
    // Collect flow information
    flowInfo := make(map[string]*flows.Flow)
    for _, result := range batchResponse.Results {
        if result.Error != nil {
            fmt.Printf("Error getting flow %s: %v\n", result.FlowID, result.Error)
            continue
        }
        
        flowInfo[result.FlowID] = result.Flow
    }
    
    return flowInfo
}

// Use the function to collect flow information
flowInfo := collectFlowInfo(ctx, client, []string{"flow-id-1", "flow-id-2", "flow-id-3"})
for id, flow := range flowInfo {
    fmt.Printf("Flow %s: %s\n", id, flow.Title)
    fmt.Printf("  Description: %s\n", flow.Description)
    fmt.Printf("  Created: %s by %s\n", flow.Created, flow.CreatedByString)
}
```

## Error Handling in Batch Operations

Batch operations ensure that errors in individual operations don't cause the entire batch to fail. Each operation's result includes its own error:

```go
// Handle errors in batch operations
batchResponse, err := client.BatchGetRuns(ctx, batchRequest)
if err != nil {
    // This is a global error with the batch operation itself
    fmt.Printf("Batch operation failed: %v\n", err)
    return
}

// Count success and failures
successCount := 0
failureCount := 0

// Check individual operation results
for _, result := range batchResponse.Results {
    if result.Error != nil {
        failureCount++
        
        // Handle specific error types
        if flows.IsRunNotFoundError(result.Error) {
            fmt.Printf("Run %s not found\n", result.RunID)
        } else if flows.IsForbiddenError(result.Error) {
            fmt.Printf("No permission to access run %s\n", result.RunID)
        } else {
            fmt.Printf("Error with run %s: %v\n", result.RunID, result.Error)
        }
    } else {
        successCount++
        // Process successful result
    }
}

fmt.Printf("Batch results: %d succeeded, %d failed\n", successCount, failureCount)
```

## Best Practices

1. **Tune Concurrency**: Adjust the concurrency level based on the operation and load
2. **Handle Individual Errors**: Check each result for errors individually
3. **Group Similar Operations**: Group operations that are likely to have similar processing times
4. **Set Reasonable Batch Sizes**: Keep batch sizes reasonable (10-50 operations per batch)
5. **Use Timeouts**: Set appropriate context timeouts for batch operations
6. **Monitor Resources**: Be aware of resource consumption with high concurrency
7. **Implement Retries**: Consider retrying failed operations
8. **Track Partial Success**: Handle cases where some operations succeed and others fail
9. **Use for Bulk Operations**: Batch operations are most beneficial for bulk operations
10. **Optimize Concurrency**: Balance concurrency with server and client limits
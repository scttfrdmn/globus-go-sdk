---
title: "Flows Service: Run Operations"
---
# Flows Service: Run Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Flow runs are instances of flows that are executed with specific inputs. The Flows service provides methods for starting, monitoring, and managing flow runs.

## Run Structure

```go
type RunResponse struct {
    RunID               string                 `json:"run_id"`
    FlowID              string                 `json:"flow_id"`
    FlowTitle           string                 `json:"flow_title,omitempty"`
    FlowDescription     string                 `json:"flow_description,omitempty"`
    FlowDefinition      map[string]interface{} `json:"flow_definition,omitempty"`
    Status              string                 `json:"status"`
    CreatedBy           string                 `json:"created_by"`
    CreatedByString     string                 `json:"created_by_string,omitempty"`
    RunOwner            string                 `json:"run_owner"`
    RunOwnerString      string                 `json:"run_owner_string,omitempty"`
    Label               string                 `json:"label,omitempty"`
    Tags                []string               `json:"tags,omitempty"`
    StartTime           string                 `json:"start_time"`
    EndTime             string                 `json:"end_time,omitempty"`
    RunError            *RunError              `json:"run_error,omitempty"`
    Input               map[string]interface{} `json:"input,omitempty"`
    Output              map[string]interface{} `json:"output,omitempty"`
    MonitoringBy        []string               `json:"monitoring_by,omitempty"`
    MonitoringGroups    []string               `json:"monitoring_groups,omitempty"`
    ManagingBy          []string               `json:"managing_by,omitempty"`
    ManagingGroups      []string               `json:"managing_groups,omitempty"`
}

type RunError struct {
    Code        string `json:"code,omitempty"`
    Description string `json:"description,omitempty"`
}
```

The RunResponse structure contains many fields, with the most important being:

| Field | Type | Description |
|-------|------|-------------|
| `RunID` | `string` | Unique identifier for the run |
| `FlowID` | `string` | ID of the flow being executed |
| `Status` | `string` | Current status of the run |
| `Label` | `string` | Human-readable label for the run |
| `Tags` | `[]string` | Tags for categorizing the run |
| `StartTime` | `string` | When the run started |
| `EndTime` | `string` | When the run completed (if finished) |
| `RunError` | `*RunError` | Error details (if failed) |
| `Input` | `map[string]interface{}` | Input provided to the run |
| `Output` | `map[string]interface{}` | Output produced by the run (if successful) |

## Running a Flow

To start a new flow run:

```go
// Create a run request
runRequest := &flows.RunRequest{
    FlowID: "flow-id",
    Label:  "Processing Job " + time.Now().Format("2006-01-02"),
    Input: map[string]interface{}{
        "source_endpoint":      "source-endpoint-id",
        "destination_endpoint": "destination-endpoint-id",
        "source_path":          "/path/to/source/file.txt",
        "destination_path":     "/path/to/destination/file.txt",
    },
    Tags: []string{"processing", "automated"},
    RunMonitoringBy:     []string{"user@example.com"},
    RunMonitoringGroups: []string{"group-id"},
    RunManagingBy:       []string{"admin@example.com"},
    RunManagingGroups:   []string{"admin-group-id"},
}

// Start the flow run
run, err := client.RunFlow(ctx, runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Flow run started: %s (Status: %s)\n", run.RunID, run.Status)
```

## Listing Runs

To list flow runs:

```go
// List all accessible runs
runList, err := client.ListRuns(ctx, nil)
if err != nil {
    // Handle error
}

// List runs with options
runList, err := client.ListRuns(ctx, &flows.ListRunsOptions{
    Limit:      100,
    Offset:     0,
    FlowID:     "flow-id",           // Filter by flow ID
    Status:     "SUCCEEDED",         // Filter by status
    StartTime:  "2023-01-01",        // Filter by start time
    EndTime:    "2023-12-31",        // Filter by end time
    Label:      "Processing",        // Filter by label
    OrderBy:    "start_time",        // Sort by start time
    OrderDir:   "desc",              // Sort in descending order
    MarkerOnly: false,               // Use marker-based pagination
})
if err != nil {
    // Handle error
}

// Iterate through runs
for _, run := range runList.Runs {
    fmt.Printf("Run: %s (Status: %s)\n", run.RunID, run.Status)
    fmt.Printf("  Started: %s\n", run.StartTime)
    if run.EndTime != "" {
        fmt.Printf("  Ended: %s\n", run.EndTime)
    }
    fmt.Printf("  Label: %s\n", run.Label)
}
```

### Run Status

The status field can have the following values:

| Status | Description |
|--------|-------------|
| `"ACTIVE"` | Run is currently executing |
| `"INACTIVE"` | Run is waiting for activation |
| `"SUCCEEDED"` | Run completed successfully |
| `"FAILED"` | Run failed |
| `"CANCELED"` | Run was canceled |

### Pagination

The run list response includes pagination information:

```go
// Check pagination information
fmt.Printf("Total runs: %d\n", runList.Total)
fmt.Printf("Has next page: %t\n", runList.HasNextPage)
fmt.Printf("Marker: %s\n", runList.Marker)
```

### Using Run Iterator

For easier pagination, use the iterator:

```go
// Create a run iterator
iterator := client.GetRunsIterator(&flows.ListRunsOptions{
    FlowID: "flow-id",
    Status: "SUCCEEDED",
})

// Iterate through all runs
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        // Handle error
        break
    }
    
    run := iterator.Run()
    fmt.Printf("Run: %s (Status: %s)\n", run.RunID, run.Status)
}
```

### Listing All Runs

To retrieve all runs automatically:

```go
// List all runs (handles pagination automatically)
allRuns, err := client.ListAllRuns(ctx, &flows.ListRunsOptions{
    FlowID: "flow-id",
})
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d runs\n", len(allRuns))
```

## Getting a Run

To retrieve a specific run by ID:

```go
// Get a specific run
run, err := client.GetRun(ctx, "run-id")
if err != nil {
    if flows.IsRunNotFoundError(err) {
        fmt.Println("Run not found")
    } else {
        fmt.Printf("Error retrieving run: %v\n", err)
    }
    return
}

fmt.Printf("Run: %s\n", run.RunID)
fmt.Printf("Flow: %s (%s)\n", run.FlowTitle, run.FlowID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Started: %s\n", run.StartTime)
if run.EndTime != "" {
    fmt.Printf("Ended: %s\n", run.EndTime)
}

// Check if the run succeeded
if run.Status == "SUCCEEDED" {
    fmt.Println("Run succeeded with output:")
    for k, v := range run.Output {
        fmt.Printf("  %s: %v\n", k, v)
    }
} else if run.Status == "FAILED" {
    fmt.Printf("Run failed: %s - %s\n", run.RunError.Code, run.RunError.Description)
}
```

## Updating a Run

To update a run's metadata:

```go
// Update a run's metadata
updateRequest := &flows.RunUpdateRequest{
    Label: "Updated Processing Job",
    Tags:  []string{"important", "processing", "updated"},
}

run, err := client.UpdateRun(ctx, "run-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Updated run: %s\n", run.RunID)
fmt.Printf("New label: %s\n", run.Label)
fmt.Printf("New tags: %v\n", run.Tags)
```

## Canceling a Run

To cancel a running flow:

```go
// Cancel a running flow
err := client.CancelRun(ctx, "run-id")
if err != nil {
    if flows.IsRunNotFoundError(err) {
        fmt.Println("Run not found")
    } else {
        fmt.Printf("Error canceling run: %v\n", err)
    }
    return
}

fmt.Println("Run canceled successfully")

// Verify cancellation
run, err := client.GetRun(ctx, "run-id")
if err != nil {
    // Handle error
}

if run.Status == "CANCELED" {
    fmt.Println("Run is now canceled")
} else {
    fmt.Printf("Run status: %s\n", run.Status)
}
```

## Getting Run Logs

To retrieve logs for a run:

```go
// Get logs for a run
logs, err := client.GetRunLogs(ctx, "run-id", 100, 0)
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d log entries\n", len(logs.LogEntries))
for _, entry := range logs.LogEntries {
    fmt.Printf("[%s] %s: %s\n", entry.Time, entry.Code, entry.Description)
}
```

### Log Entry Structure

```go
type RunLogEntry struct {
    Time        string `json:"time"`
    Code        string `json:"code"`
    Description string `json:"description"`
    Details     string `json:"details,omitempty"`
}
```

### Using Log Iterator

For easier pagination of logs, use the iterator:

```go
// Create a run log iterator
iterator := client.GetRunLogsIterator("run-id", 50)

// Iterate through all log entries
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        // Handle error
        break
    }
    
    entry := iterator.LogEntry()
    fmt.Printf("[%s] %s: %s\n", entry.Time, entry.Code, entry.Description)
}
```

### Listing All Run Logs

To retrieve all log entries automatically:

```go
// List all logs for a run (handles pagination automatically)
allLogs, err := client.ListAllRunLogs(ctx, "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d log entries\n", len(allLogs))
```

## Waiting for a Run to Complete

To wait for a run to complete:

```go
// Wait for a run to complete with a 5-second polling interval
run, err := client.WaitForRun(ctx, "run-id", 5*time.Second)
if err != nil {
    // Handle error
}

fmt.Printf("Run completed with status: %s\n", run.Status)
if run.Status == "SUCCEEDED" {
    fmt.Println("Output:", run.Output)
} else if run.Status == "FAILED" {
    fmt.Printf("Run failed: %s - %s\n", run.RunError.Code, run.RunError.Description)
} else if run.Status == "CANCELED" {
    fmt.Println("Run was canceled")
}
```

### Using Context Timeout

You can combine `WaitForRun` with a context timeout:

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

// Wait for the run to complete or timeout
run, err := client.WaitForRun(ctx, "run-id", 5*time.Second)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Timed out waiting for run to complete")
    } else {
        fmt.Printf("Error waiting for run: %v\n", err)
    }
    return
}

fmt.Printf("Run completed with status: %s\n", run.Status)
```

## Run Monitoring

### Checking Run Progress

To check the progress of a running flow:

```go
// Function to monitor a run
func monitorRun(ctx context.Context, client *flows.Client, runID string) {
    for {
        // Get the current run status
        run, err := client.GetRun(ctx, runID)
        if err != nil {
            fmt.Printf("Error getting run: %v\n", err)
            return
        }
        
        // Print status
        fmt.Printf("Run status: %s\n", run.Status)
        
        // Check if the run is no longer active
        if run.Status != "ACTIVE" && run.Status != "INACTIVE" {
            fmt.Println("Run is no longer active")
            
            if run.Status == "SUCCEEDED" {
                fmt.Println("Run succeeded with output:")
                for k, v := range run.Output {
                    fmt.Printf("  %s: %v\n", k, v)
                }
            } else if run.Status == "FAILED" {
                fmt.Printf("Run failed: %s - %s\n", run.RunError.Code, run.RunError.Description)
            } else if run.Status == "CANCELED" {
                fmt.Println("Run was canceled")
            }
            
            return
        }
        
        // Get the latest logs
        logs, err := client.GetRunLogs(ctx, runID, 5, 0)
        if err != nil {
            fmt.Printf("Error getting logs: %v\n", err)
        } else if len(logs.LogEntries) > 0 {
            fmt.Println("Latest log entries:")
            for _, entry := range logs.LogEntries {
                fmt.Printf("[%s] %s: %s\n", entry.Time, entry.Code, entry.Description)
            }
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
go monitorRun(ctx, client, "run-id")
```

## Common Run Patterns

### Start and Wait for Completion

```go
// Run a flow and wait for completion
runRequest := &flows.RunRequest{
    FlowID: "flow-id",
    Label:  "Processing Job",
    Input: map[string]interface{}{
        // Flow input parameters
    },
}

// Start the flow run
run, err := client.RunFlow(ctx, runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Flow run started: %s\n", run.RunID)

// Wait for the run to complete
run, err = client.WaitForRun(ctx, run.RunID, 5*time.Second)
if err != nil {
    // Handle error
}

// Process results
if run.Status == "SUCCEEDED" {
    fmt.Println("Run succeeded with output:")
    for k, v := range run.Output {
        fmt.Printf("  %s: %v\n", k, v)
    }
} else {
    fmt.Printf("Run did not succeed. Status: %s\n", run.Status)
    if run.RunError != nil {
        fmt.Printf("Error: %s - %s\n", run.RunError.Code, run.RunError.Description)
    }
}
```

### Batch Run Multiple Flows

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
            FlowID: "flow-id",
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
        fmt.Printf("Run %d started: %s\n", i+1, result.Run.RunID)
    }
}

// Store the run IDs for later checking
var runIDs []string
for _, result := range batchResponse.Results {
    if result.Error == nil {
        runIDs = append(runIDs, result.Run.RunID)
    }
}
```

### Monitor a Set of Runs

```go
// Function to monitor multiple runs
func monitorRuns(ctx context.Context, client *flows.Client, runIDs []string) {
    // Create a map to track run status
    runStatus := make(map[string]string)
    for _, runID := range runIDs {
        runStatus[runID] = "UNKNOWN"
    }
    
    // Track completion
    completed := 0
    totalRuns := len(runIDs)
    
    for completed < totalRuns {
        // Create batch request to get run status
        batchRequest := &flows.BatchRunsRequest{
            RunIDs: runIDs,
            Options: &flows.BatchOptions{
                Concurrency: 5, // Get 5 runs concurrently
            },
        }
        
        // Get status of all runs
        batchResponse, err := client.BatchGetRuns(ctx, batchRequest)
        if err != nil {
            fmt.Printf("Error getting runs: %v\n", err)
            return
        }
        
        // Update status and check for completion
        completed = 0
        for _, result := range batchResponse.Results {
            if result.Error != nil {
                fmt.Printf("Error getting run %s: %v\n", result.RunID, result.Error)
                continue
            }
            
            // Update status
            oldStatus := runStatus[result.Run.RunID]
            newStatus := result.Run.Status
            runStatus[result.Run.RunID] = newStatus
            
            // Print status changes
            if oldStatus != newStatus {
                fmt.Printf("Run %s status changed: %s -> %s\n", 
                    result.Run.RunID, oldStatus, newStatus)
            }
            
            // Check if the run is complete
            if newStatus == "SUCCEEDED" || newStatus == "FAILED" || newStatus == "CANCELED" {
                completed++
            }
        }
        
        fmt.Printf("Progress: %d/%d runs completed\n", completed, totalRuns)
        
        // Wait before checking again
        select {
        case <-ctx.Done():
            fmt.Println("Monitoring canceled")
            return
        case <-time.After(30 * time.Second):
            // Continue checking
        }
    }
    
    fmt.Println("All runs completed!")
}

// Use the function to monitor runs
go monitorRuns(ctx, client, runIDs)
```

### Cancel Multiple Runs

```go
// Create batch cancel request
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

## Best Practices

1. **Use Descriptive Labels**: Provide meaningful labels for runs to make them easier to identify
2. **Add Tags**: Use tags to categorize and filter runs
3. **Handle Run Failures**: Check run status and error details
4. **Monitor Long-running Flows**: Regularly check status and logs for long-running flows
5. **Use Batch Operations**: Use batch operations for working with multiple runs
6. **Set Appropriate Permissions**: Configure monitoring and management permissions
7. **Check Logs for Errors**: Examine logs to diagnose issues
8. **Use Wait with Timeout**: Combine `WaitForRun` with context timeouts
9. **Implement Timeout Handling**: Handle cases where runs take too long
10. **Cleanup Old Runs**: Consider cleaning up or archiving old runs
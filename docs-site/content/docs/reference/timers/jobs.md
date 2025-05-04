---
title: "Timers Service: Run Operations"
---
# Timers Service: Run Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Timers client provides methods for managing and monitoring timer runs, which are the individual executions of a timer's callback.

## Run Model

The central data type for run operations is the `TimerRun` struct:

```go
type TimerRun struct {
    // ID is the unique identifier for the run
    ID string `json:"id"`

    // TimerID is the ID of the timer that was run
    TimerID string `json:"timer_id"`

    // Status is the status of the run
    Status string `json:"status"`

    // StartTime is when the run started
    StartTime time.Time `json:"start_time"`

    // EndTime is when the run ended
    EndTime *time.Time `json:"end_time,omitempty"`

    // Result contains the result of the run
    Result *RunResult `json:"result,omitempty"`
}

// RunResult represents the result of a timer run
type RunResult struct {
    // Status is the status of the callback execution
    Status string `json:"status"`

    // StatusCode is the HTTP status code for web callbacks
    StatusCode *int `json:"status_code,omitempty"`

    // RunID is the ID of the flow run for flow callbacks
    RunID *string `json:"run_id,omitempty"`

    // Error contains error information if the run failed
    Error *RunError `json:"error,omitempty"`
}

// RunError represents an error that occurred during a timer run
type RunError struct {
    // Code is the error code
    Code string `json:"code"`

    // Message is the error message
    Message string `json:"message"`

    // Detail contains additional error details
    Detail map[string]interface{} `json:"detail,omitempty"`
}
```

## Run Status

Runs can have the following status values:

```go
// RunStatus represents the possible statuses of a timer run
type RunStatus string

const (
    // RunStatusPending indicates the run is pending
    RunStatusPending RunStatus = "pending"

    // RunStatusInProgress indicates the run is in progress
    RunStatusInProgress RunStatus = "in_progress"

    // RunStatusSuccess indicates the run succeeded
    RunStatusSuccess RunStatus = "success"

    // RunStatusFailure indicates the run failed
    RunStatusFailure RunStatus = "failure"
)
```

## Manually Triggering a Run

```go
// Manually trigger a timer run
timerID := "12345678-1234-1234-1234-123456789012"
run, err := client.RunTimer(ctx, timerID)
if err != nil {
    // Handle error
}

fmt.Printf("Manual run initiated: %s\n", run.ID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Start time: %s\n", run.StartTime.Format(time.RFC3339))
```

## Listing Runs for a Timer

```go
// Create options for listing runs
timerID := "12345678-1234-1234-1234-123456789012"
limit := 20
status := "success"
startAfter := time.Now().Add(-7 * 24 * time.Hour) // Runs in the last week
options := &timers.ListRunsOptions{
    Limit: &limit,
    Status: &status,
    StartAfter: &startAfter,
}

// List runs
runList, err := client.ListRuns(ctx, timerID, options)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d runs (total: %d)\n", len(runList.Runs), runList.Total)

// Process results
for _, run := range runList.Runs {
    fmt.Printf("Run: %s\n", run.ID)
    fmt.Printf("Status: %s\n", run.Status)
    fmt.Printf("Started: %s\n", run.StartTime.Format(time.RFC3339))
    if run.EndTime != nil {
        fmt.Printf("Ended: %s\n", run.EndTime.Format(time.RFC3339))
        duration := run.EndTime.Sub(run.StartTime)
        fmt.Printf("Duration: %s\n", duration)
    }
    
    // Check for results
    if run.Result != nil {
        fmt.Printf("Result status: %s\n", run.Result.Status)
        if run.Result.StatusCode != nil {
            fmt.Printf("HTTP status code: %d\n", *run.Result.StatusCode)
        }
        if run.Result.RunID != nil {
            fmt.Printf("Flow run ID: %s\n", *run.Result.RunID)
        }
        
        // Check for errors
        if run.Result.Error != nil {
            fmt.Printf("Error code: %s\n", run.Result.Error.Code)
            fmt.Printf("Error message: %s\n", run.Result.Error.Message)
        }
    }
    
    fmt.Println()
}

// Check if there are more runs
if runList.HasNextPage {
    // Use the NextPage marker to get the next page
    nextMarker := runList.NextPage
    nextOptions := &timers.ListRunsOptions{
        Limit: &limit,
        Marker: nextMarker,
    }
    // Get next page...
}
```

## Getting a Specific Run

```go
// Get a specific run
timerID := "12345678-1234-1234-1234-123456789012"
runID := "abcdef12-abcd-abcd-abcd-abcdef123456"

run, err := client.GetRun(ctx, timerID, runID)
if err != nil {
    // Handle error
}

fmt.Printf("Run: %s\n", run.ID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Start time: %s\n", run.StartTime.Format(time.RFC3339))

if run.EndTime != nil {
    fmt.Printf("End time: %s\n", run.EndTime.Format(time.RFC3339))
    duration := run.EndTime.Sub(run.StartTime)
    fmt.Printf("Duration: %s\n", duration)
}

// Check for results
if run.Result != nil {
    fmt.Printf("Result status: %s\n", run.Result.Status)
    
    // Check for web callback results
    if run.Result.StatusCode != nil {
        fmt.Printf("HTTP status code: %d\n", *run.Result.StatusCode)
    }
    
    // Check for flow callback results
    if run.Result.RunID != nil {
        fmt.Printf("Flow run ID: %s\n", *run.Result.RunID)
    }
    
    // Check for errors
    if run.Result.Error != nil {
        fmt.Printf("Error code: %s\n", run.Result.Error.Code)
        fmt.Printf("Error message: %s\n", run.Result.Error.Message)
        
        if run.Result.Error.Detail != nil {
            fmt.Println("Error details:")
            for k, v := range run.Result.Error.Detail {
                fmt.Printf("  %s: %v\n", k, v)
            }
        }
    }
}
```

## Filtering Runs

You can use the `ListRunsOptions` struct to filter runs:

```go
// Filter runs by status
statusFilter := "success"
options := &timers.ListRunsOptions{
    Status: &statusFilter,
}

// Filter runs by time range
startAfter := time.Now().Add(-24 * time.Hour) // Last 24 hours
startBefore := time.Now()
options := &timers.ListRunsOptions{
    StartAfter: &startAfter,
    StartBefore: &startBefore,
}

// Filter runs with pagination
limit := 10
marker := "last-run-id"
options := &timers.ListRunsOptions{
    Limit: &limit,
    Marker: &marker,
}
```

## Analyzing Run Results

Different callback types have different result formats:

### Web Callback Results

For web callbacks, the `Result` will contain an HTTP status code:

```go
if run.Result != nil && run.Result.StatusCode != nil {
    statusCode := *run.Result.StatusCode
    if statusCode >= 200 && statusCode < 300 {
        fmt.Println("Web callback succeeded with status:", statusCode)
    } else {
        fmt.Println("Web callback failed with status:", statusCode)
    }
}
```

### Flow Callback Results

For flow callbacks, the `Result` will contain a flow run ID:

```go
if run.Result != nil && run.Result.RunID != nil {
    flowRunID := *run.Result.RunID
    fmt.Println("Flow run initiated with ID:", flowRunID)
    
    // You can use the Flows client to retrieve details about the flow run
    flowsClient, _ := config.NewFlowsClient(accessToken)
    flowRun, err := flowsClient.GetRun(ctx, flowRunID)
    // Process flow run details...
}
```

## Error Handling

Run operations can return the following types of errors:

- Validation errors (invalid timer ID or run ID)
- Authentication errors (insufficient permissions)
- Resource not found errors (timer or run doesn't exist)
- API communication errors

Example error handling:

```go
run, err := client.GetRun(ctx, timerID, runID)
if err != nil {
    if strings.Contains(err.Error(), "404") {
        // Run not found
        fmt.Printf("Run %s for timer %s does not exist\n", runID, timerID)
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied
        fmt.Println("You don't have permission to access this timer run")
    } else {
        // Other error
        fmt.Printf("Error retrieving run: %v\n", err)
    }
}
```
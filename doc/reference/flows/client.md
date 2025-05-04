# Flows Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Flows client provides access to the Globus Flows API, which allows you to create, manage, and execute automated workflows.

## Client Structure

```go
type Client struct {
    client *core.Client
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*core.Client` | Core client for making HTTP requests |

## Creating a Flows Client

```go
// Create a flows client with options
client, err := flows.NewClient(
    flows.WithAccessToken("access-token"),
    flows.WithHTTPDebugging(),
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
| `WithBaseURL(url string)` | Sets a custom base URL (default: "https://flows.globus.org/v1/") |
| `WithHTTPDebugging()` | Enables HTTP debugging |
| `WithHTTPTracing()` | Enables HTTP tracing |

## Flow Management

### Listing Flows

```go
// List accessible flows
flowList, err := client.ListFlows(ctx, nil)
if err != nil {
    // Handle error
}

// List flows with options
flowList, err := client.ListFlows(ctx, &flows.ListFlowsOptions{
    Limit:  100,
    Offset: 0,
    Title:  "My Flow",
    Owner:  "owner-id",
})
if err != nil {
    // Handle error
}

// Iterate through flows
for _, flow := range flowList.Flows {
    fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
}
```

### Getting a Flow

```go
// Get a specific flow by ID
flow, err := client.GetFlow(ctx, "flow-id")
if err != nil {
    // Handle error
}

fmt.Printf("Flow Title: %s\n", flow.Title)
fmt.Printf("Description: %s\n", flow.Description)
fmt.Printf("Owner: %s\n", flow.Owner)
```

### Creating a Flow

```go
// Create a new flow
createRequest := &flows.FlowCreateRequest{
    Title:       "Data Processing Flow",
    Description: "Processes incoming data files",
    Definition: map[string]interface{}{
        "StartAt": "Process",
        "States": map[string]interface{}{
            "Process": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/hello_world",
                "Parameters": map[string]interface{}{
                    "echo_string": "Hello, World!",
                },
                "End": true,
            },
        },
    },
    Keywords: []string{"data", "processing"},
}

flow, err := client.CreateFlow(ctx, createRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Created flow: %s (%s)\n", flow.Title, flow.ID)
```

### Updating a Flow

```go
// Update an existing flow
updateRequest := &flows.FlowUpdateRequest{
    Title:       "Updated Data Processing Flow",
    Description: "Improved data processing workflow",
    Keywords:    []string{"data", "processing", "updated"},
}

flow, err := client.UpdateFlow(ctx, "flow-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Updated flow: %s\n", flow.Title)
```

### Deleting a Flow

```go
// Delete a flow
err := client.DeleteFlow(ctx, "flow-id")
if err != nil {
    // Handle error
}

fmt.Println("Flow deleted successfully")
```

## Flow Execution

### Running a Flow

```go
// Run a flow
runRequest := &flows.RunRequest{
    FlowID: "flow-id",
    Label:  "Processing Job " + time.Now().Format("2006-01-02"),
    Input: map[string]interface{}{
        "input_file": "path/to/file.txt",
        "options": map[string]interface{}{
            "process_headers": true,
            "output_format":   "csv",
        },
    },
}

run, err := client.RunFlow(ctx, runRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Flow run started: %s (Status: %s)\n", run.RunID, run.Status)
```

### Listing Runs

```go
// List flow runs
runList, err := client.ListRuns(ctx, nil)
if err != nil {
    // Handle error
}

// List runs with options
runList, err := client.ListRuns(ctx, &flows.ListRunsOptions{
    Limit:    100,
    Offset:   0,
    FlowID:   "flow-id",
    Status:   "SUCCEEDED",
    StartTime: "2023-01-01",
    EndTime:   "2023-12-31",
})
if err != nil {
    // Handle error
}

// Iterate through runs
for _, run := range runList.Runs {
    fmt.Printf("Run: %s (Status: %s)\n", run.RunID, run.Status)
}
```

### Getting a Run

```go
// Get a specific run
run, err := client.GetRun(ctx, "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Run: %s\n", run.RunID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Started: %s\n", run.StartTime)
if run.EndTime != "" {
    fmt.Printf("Ended: %s\n", run.EndTime)
}

// Access input and output
fmt.Println("Input:", run.Input)
fmt.Println("Output:", run.Output)
```

### Canceling a Run

```go
// Cancel a running flow
err := client.CancelRun(ctx, "run-id")
if err != nil {
    // Handle error
}

fmt.Println("Run canceled successfully")
```

### Updating a Run

```go
// Update a run's metadata
updateRequest := &flows.RunUpdateRequest{
    Label: "Updated Label",
    Tags:  []string{"important", "processing"},
}

run, err := client.UpdateRun(ctx, "run-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Updated run: %s\n", run.RunID)
```

### Getting Run Logs

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

### Waiting for a Run to Complete

```go
// Wait for a run to complete with a 5-second polling interval
run, err := client.WaitForRun(ctx, "run-id", 5*time.Second)
if err != nil {
    // Handle error
}

fmt.Printf("Run completed with status: %s\n", run.Status)
if run.Status == "SUCCEEDED" {
    fmt.Println("Output:", run.Output)
} else {
    fmt.Println("Error:", run.RunError)
}
```

## Action Providers

### Listing Action Providers

```go
// List available action providers
providerList, err := client.ListActionProviders(ctx, nil)
if err != nil {
    // Handle error
}

// List with options
providerList, err := client.ListActionProviders(ctx, &flows.ListActionProvidersOptions{
    Limit:  100,
    Offset: 0,
})
if err != nil {
    // Handle error
}

// Iterate through providers
for _, provider := range providerList.ActionProviders {
    fmt.Printf("Provider: %s (%s)\n", provider.DisplayName, provider.ID)
}
```

### Getting an Action Provider

```go
// Get a specific action provider
provider, err := client.GetActionProvider(ctx, "provider-id")
if err != nil {
    // Handle error
}

fmt.Printf("Provider: %s\n", provider.DisplayName)
fmt.Printf("Description: %s\n", provider.Description)
fmt.Printf("Type: %s\n", provider.Type)
```

## Action Roles

### Listing Action Roles

```go
// List roles for an action provider
roleList, err := client.ListActionRoles(ctx, "provider-id", 100, 0)
if err != nil {
    // Handle error
}

// Iterate through roles
for _, role := range roleList.ActionRoles {
    fmt.Printf("Role: %s (%s)\n", role.Name, role.ID)
}
```

### Getting an Action Role

```go
// Get a specific action role
role, err := client.GetActionRole(ctx, "provider-id", "role-id")
if err != nil {
    // Handle error
}

fmt.Printf("Role: %s\n", role.Name)
fmt.Printf("Description: %s\n", role.Description)
fmt.Printf("Required Scopes: %v\n", role.RequiredScopes)
```

## Iterators

The Flows client provides iterator methods for convenient pagination:

### Flow Iterator

```go
// Create a flow iterator
iterator := client.GetFlowsIterator(&flows.ListFlowsOptions{
    Title: "Processing",
})

// Iterate through all flows
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        // Handle error
        break
    }
    
    flow := iterator.Flow()
    fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
}
```

### Run Iterator

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

### Action Provider Iterator

```go
// Create an action provider iterator
iterator := client.GetActionProvidersIterator(&flows.ListActionProvidersOptions{
    Limit: 50,
})

// Iterate through all action providers
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        // Handle error
        break
    }
    
    provider := iterator.ActionProvider()
    fmt.Printf("Provider: %s (%s)\n", provider.DisplayName, provider.ID)
}
```

### Action Role Iterator

```go
// Create an action role iterator
iterator := client.GetActionRolesIterator("provider-id", 50)

// Iterate through all action roles
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        // Handle error
        break
    }
    
    role := iterator.ActionRole()
    fmt.Printf("Role: %s (%s)\n", role.Name, role.ID)
}
```

### Run Log Iterator

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

## Convenience Methods

The Flows client provides convenience methods for retrieving all items:

### Listing All Flows

```go
// List all flows (handles pagination automatically)
allFlows, err := client.ListAllFlows(ctx, &flows.ListFlowsOptions{
    Title: "Processing",
})
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d flows\n", len(allFlows))
```

### Listing All Runs

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

### Listing All Action Providers

```go
// List all action providers (handles pagination automatically)
allProviders, err := client.ListAllActionProviders(ctx, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d action providers\n", len(allProviders))
```

### Listing All Action Roles

```go
// List all action roles for a provider (handles pagination automatically)
allRoles, err := client.ListAllActionRoles(ctx, "provider-id")
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d action roles\n", len(allRoles))
```

### Listing All Run Logs

```go
// List all logs for a run (handles pagination automatically)
allLogs, err := client.ListAllRunLogs(ctx, "run-id")
if err != nil {
    // Handle error
}

fmt.Printf("Retrieved %d log entries\n", len(allLogs))
```

## Error Handling

The flows package provides specific error types and helper functions for common error conditions:

```go
// Try an operation
_, err := client.GetFlow(ctx, "non-existent-flow")
if err != nil {
    switch {
    case flows.IsFlowNotFoundError(err):
        fmt.Println("Flow not found")
    case flows.IsRunNotFoundError(err):
        fmt.Println("Run not found")
    case flows.IsActionProviderNotFoundError(err):
        fmt.Println("Action provider not found")
    case flows.IsActionRoleNotFoundError(err):
        fmt.Println("Action role not found")
    case flows.IsForbiddenError(err):
        fmt.Println("Permission denied")
    case flows.IsValidationError(err):
        fmt.Println("Validation error")
    default:
        fmt.Println("Other error:", err)
    }
}
```

## Common Patterns

### Create and Run a Flow

```go
// Create a new flow
createRequest := &flows.FlowCreateRequest{
    Title:       "File Processing Flow",
    Description: "Processes files uploaded to Globus collection",
    Definition: map[string]interface{}{
        "StartAt": "TransferFile",
        "States": map[string]interface{}{
            "TransferFile": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/transfer/transfer",
                "Parameters": map[string]interface{}{
                    "source_endpoint_id": "${input.source_endpoint}",
                    "destination_endpoint_id": "${input.destination_endpoint}",
                    "transfer_items": []map[string]interface{}{
                        {
                            "source_path": "${input.source_path}",
                            "destination_path": "${input.destination_path}",
                        },
                    },
                },
                "Next": "ProcessFile",
            },
            "ProcessFile": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.function_id}",
                    "function_parameters": map[string]interface{}{
                        "input_file": "${input.destination_path}",
                    },
                },
                "End": true,
            },
        },
    },
}

flow, err := client.CreateFlow(ctx, createRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Created flow: %s (%s)\n", flow.Title, flow.ID)

// Run the flow
runRequest := &flows.RunRequest{
    FlowID: flow.ID,
    Label:  "Process file " + time.Now().Format("2006-01-02"),
    Input: map[string]interface{}{
        "source_endpoint":      "source-endpoint-id",
        "destination_endpoint": "destination-endpoint-id",
        "source_path":          "/path/to/source/file.txt",
        "destination_path":     "/path/to/destination/file.txt",
        "compute_endpoint":     "compute-endpoint-id",
        "function_id":          "function-id",
    },
}

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

if run.Status == "SUCCEEDED" {
    fmt.Println("Flow run completed successfully")
    fmt.Println("Output:", run.Output)
} else {
    fmt.Printf("Flow run failed: %s\n", run.RunError.Description)
}
```

### Monitor All Runs for a Flow

```go
// Get all active runs for a flow
activeRuns, err := client.ListAllRuns(ctx, &flows.ListRunsOptions{
    FlowID: "flow-id",
    Status: "ACTIVE",
})
if err != nil {
    // Handle error
}

fmt.Printf("Found %d active runs\n", len(activeRuns))

// Monitor each run
for _, run := range activeRuns {
    // Get details for the run
    run, err := client.GetRun(ctx, run.RunID)
    if err != nil {
        fmt.Printf("Error getting run %s: %v\n", run.RunID, err)
        continue
    }
    
    fmt.Printf("Run %s (Status: %s)\n", run.RunID, run.Status)
    fmt.Printf("Started: %s\n", run.StartTime)
    
    // Get the latest log entries
    logs, err := client.GetRunLogs(ctx, run.RunID, 5, 0)
    if err != nil {
        fmt.Printf("Error getting logs for run %s: %v\n", run.RunID, err)
        continue
    }
    
    if len(logs.LogEntries) > 0 {
        fmt.Println("Latest log entries:")
        for _, entry := range logs.LogEntries {
            fmt.Printf("[%s] %s: %s\n", entry.Time, entry.Code, entry.Description)
        }
    }
}
```

### Find Action Providers by Type

```go
// List all action providers
allProviders, err := client.ListAllActionProviders(ctx, nil)
if err != nil {
    // Handle error
}

// Find transfer providers
var transferProviders []flows.ActionProvider
for _, provider := range allProviders {
    if strings.Contains(strings.ToLower(provider.DisplayName), "transfer") || 
       strings.Contains(strings.ToLower(provider.Description), "transfer") {
        transferProviders = append(transferProviders, provider)
    }
}

fmt.Printf("Found %d transfer-related action providers\n", len(transferProviders))
for _, provider := range transferProviders {
    fmt.Printf("- %s (%s)\n", provider.DisplayName, provider.ID)
    fmt.Printf("  %s\n", provider.Description)
}
```

## Best Practices

1. **Use Iterators for Pagination**: For large collections, use iterators to handle pagination automatically
2. **Wait for Long-running Operations**: Use the `WaitForRun` method for long-running flows
3. **Handle Errors Appropriately**: Use the type-specific error checking functions
4. **Use Meaningful Labels**: Provide descriptive labels for flows and runs
5. **Include Error Handling in Flows**: Design flows to handle errors gracefully
6. **Manage Flow Access**: Set appropriate access controls when creating flows
7. **Check Provider Compatibility**: Verify action provider compatibility before using in flows
8. **Use Context Timeouts**: Set appropriate context timeouts for operations
9. **Monitor Run Status**: Regularly check the status of long-running flows
10. **Check Run Logs**: Use run logs to diagnose issues with flows
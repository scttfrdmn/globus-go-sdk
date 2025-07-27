---
title: "Flows Service Quick Start"
weight: 40
---

# Flows Service Quick Start

This guide will help you get started with the Globus Flows service using the Go SDK. The Flows service enables you to create, manage, and execute automated workflows that coordinate actions across Globus services and beyond.

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
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    // Continue with the examples below...
}
```

## Creating a Flows Client

There are two main ways to create a Flows client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Flows client
flowsClient, err := config.NewFlowsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create flows client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Flows client with options
flowsClient, err := flows.NewClient(
    flows.WithAccessToken(os.Getenv("GLOBUS_ACCESS_TOKEN")),
    flows.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create flows client: %v", err)
}
```

## Working with Flows

Flows are definitions of automated workflows that can be executed.

### Listing Your Flows

```go
// List all flows you have access to
flowList, err := flowsClient.ListFlows(ctx, nil)
if err != nil {
    log.Fatalf("Failed to list flows: %v", err)
}

fmt.Printf("Found %d flows:\n", len(flowList.Flows))
for i, flow := range flowList.Flows {
    fmt.Printf("%d. %s (%s)\n", i+1, flow.Title, flow.ID)
}
```

### Listing Flows with Options

```go
// List flows with filtering options
flowList, err := flowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
    Limit:  25,           // Limit results per page
    Offset: 0,            // Starting offset
    Title:  "Data",       // Filter by title containing "Data"
    Owner:  "owner-id",   // Filter by owner
})
if err != nil {
    log.Fatalf("Failed to list flows with filters: %v", err)
}
```

### Getting Flow Details

```go
// Get details about a specific flow
flowID := "your-flow-id"
flow, err := flowsClient.GetFlow(ctx, flowID)
if err != nil {
    log.Fatalf("Failed to get flow: %v", err)
}

fmt.Printf("Flow: %s\n", flow.Title)
fmt.Printf("Description: %s\n", flow.Description)
fmt.Printf("Owner: %s\n", flow.Owner)
fmt.Printf("Created: %s\n", flow.Created)
fmt.Printf("Last Updated: %s\n", flow.Updated)
```

### Creating a Simple Flow

Here's a simple "Hello World" flow that uses the hello_world action provider:

```go
// Create a new flow
createRequest := &flows.FlowCreateRequest{
    Title:       "Hello World Flow",
    Description: "A simple Hello World flow",
    Definition: map[string]interface{}{
        "StartAt": "HelloWorld",
        "States": map[string]interface{}{
            "HelloWorld": map[string]interface{}{
                "Type":      "Action",
                "ActionUrl": "https://actions.globus.org/hello_world",
                "Parameters": map[string]interface{}{
                    "echo_string": "Hello from Globus Flows!",
                },
                "End": true,
            },
        },
    },
    Keywords: []string{"example", "hello-world"},
}

newFlow, err := flowsClient.CreateFlow(ctx, createRequest)
if err != nil {
    log.Fatalf("Failed to create flow: %v", err)
}

fmt.Printf("Created new flow with ID: %s\n", newFlow.ID)
```

### Creating a Transfer Flow

Here's a more useful flow that transfers a file between Globus endpoints:

```go
// Create a transfer flow
transferFlow := &flows.FlowCreateRequest{
    Title:       "File Transfer Flow",
    Description: "Transfers a file between Globus endpoints",
    Definition: map[string]interface{}{
        "StartAt": "TransferFile",
        "States": map[string]interface{}{
            "TransferFile": map[string]interface{}{
                "Type":      "Action",
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
                "End": true,
            },
        },
    },
}

flow, err := flowsClient.CreateFlow(ctx, transferFlow)
if err != nil {
    log.Fatalf("Failed to create transfer flow: %v", err)
}

fmt.Printf("Created transfer flow with ID: %s\n", flow.ID)
```

### Updating a Flow

```go
// Update an existing flow
updateRequest := &flows.FlowUpdateRequest{
    Title:       "Updated Hello World Flow",
    Description: "An updated Hello World flow",
    Keywords:    []string{"example", "hello-world", "updated"},
}

updatedFlow, err := flowsClient.UpdateFlow(ctx, flowID, updateRequest)
if err != nil {
    log.Fatalf("Failed to update flow: %v", err)
}

fmt.Printf("Updated flow: %s\n", updatedFlow.Title)
```

### Deleting a Flow

```go
// Delete a flow
err = flowsClient.DeleteFlow(ctx, flowID)
if err != nil {
    log.Fatalf("Failed to delete flow: %v", err)
}

fmt.Println("Flow deleted successfully")
```

## Running Flows

### Running a Simple Flow

```go
// Run the Hello World flow
runRequest := &flows.RunRequest{
    FlowID: flowID,
    Label:  "Hello World Run " + time.Now().Format("2006-01-02"),
    Input: map[string]interface{}{
        // No input needed for the basic hello world flow
    },
}

run, err := flowsClient.RunFlow(ctx, runRequest)
if err != nil {
    log.Fatalf("Failed to run flow: %v", err)
}

fmt.Printf("Flow run started with ID: %s\n", run.RunID)
fmt.Printf("Status: %s\n", run.Status)
```

### Running a Transfer Flow

```go
// Run the transfer flow with specific input parameters
runRequest := &flows.RunRequest{
    FlowID: transferFlowID,
    Label:  "Transfer Run " + time.Now().Format("2006-01-02"),
    Input: map[string]interface{}{
        "source_endpoint":      "source-endpoint-id",
        "destination_endpoint": "destination-endpoint-id",
        "source_path":          "/path/to/source/file.txt",
        "destination_path":     "/path/to/destination/file.txt",
    },
}

run, err := flowsClient.RunFlow(ctx, runRequest)
if err != nil {
    log.Fatalf("Failed to run transfer flow: %v", err)
}

fmt.Printf("Transfer flow run started with ID: %s\n", run.RunID)
```

### Waiting for a Run to Complete

```go
// Run the flow and wait for it to complete
run, err := flowsClient.RunFlow(ctx, runRequest)
if err != nil {
    log.Fatalf("Failed to run flow: %v", err)
}

fmt.Printf("Flow run started with ID: %s\n", run.RunID)

// Wait for the run to complete with a 5-second polling interval
completedRun, err := flowsClient.WaitForRun(ctx, run.RunID, 5*time.Second)
if err != nil {
    log.Fatalf("Error waiting for run to complete: %v", err)
}

fmt.Printf("Run completed with status: %s\n", completedRun.Status)

// Check if the run was successful
if completedRun.Status == "SUCCEEDED" {
    fmt.Println("Flow run succeeded!")
    fmt.Println("Output:", completedRun.Output)
} else {
    fmt.Printf("Flow run failed with status: %s\n", completedRun.Status)
    if completedRun.RunError != nil {
        fmt.Printf("Error: %s\n", completedRun.RunError.Description)
    }
}
```

## Monitoring Flow Runs

### Listing Flow Runs

```go
// List all runs for a specific flow
runList, err := flowsClient.ListRuns(ctx, &flows.ListRunsOptions{
    FlowID: flowID,
    Limit:  25,
})
if err != nil {
    log.Fatalf("Failed to list runs: %v", err)
}

fmt.Printf("Found %d runs for flow %s:\n", len(runList.Runs), flowID)
for i, run := range runList.Runs {
    fmt.Printf("%d. Run ID: %s (Status: %s)\n", i+1, run.RunID, run.Status)
    fmt.Printf("   Started: %s\n", run.StartTime)
    if run.EndTime != "" {
        fmt.Printf("   Ended: %s\n", run.EndTime)
    }
}
```

### Filtering Runs by Status

```go
// List only successful runs
successfulRuns, err := flowsClient.ListRuns(ctx, &flows.ListRunsOptions{
    Status: "SUCCEEDED",
    Limit:  25,
})
if err != nil {
    log.Fatalf("Failed to list successful runs: %v", err)
}

fmt.Printf("Found %d successful runs:\n", len(successfulRuns.Runs))
```

### Getting Run Details

```go
// Get details about a specific run
runID := "your-run-id"
run, err := flowsClient.GetRun(ctx, runID)
if err != nil {
    log.Fatalf("Failed to get run: %v", err)
}

fmt.Printf("Run: %s\n", run.RunID)
fmt.Printf("Flow: %s\n", run.FlowID)
fmt.Printf("Status: %s\n", run.Status)
fmt.Printf("Started: %s\n", run.StartTime)
if run.EndTime != "" {
    fmt.Printf("Ended: %s\n", run.EndTime)
    
    // Calculate duration
    startTime, _ := time.Parse(time.RFC3339, run.StartTime)
    endTime, _ := time.Parse(time.RFC3339, run.EndTime)
    duration := endTime.Sub(startTime)
    fmt.Printf("Duration: %s\n", duration)
}

// Display input and output
fmt.Println("Input:", run.Input)
fmt.Println("Output:", run.Output)
```

### Getting Run Logs

```go
// Get logs for a specific run
logs, err := flowsClient.GetRunLogs(ctx, runID, 100, 0)
if err != nil {
    log.Fatalf("Failed to get run logs: %v", err)
}

fmt.Printf("Retrieved %d log entries for run %s:\n", len(logs.LogEntries), runID)
for i, entry := range logs.LogEntries {
    fmt.Printf("%d. [%s] %s: %s\n", i+1, entry.Time, entry.Code, entry.Description)
}
```

### Canceling a Run

```go
// Cancel a running flow
err = flowsClient.CancelRun(ctx, runID)
if err != nil {
    log.Fatalf("Failed to cancel run: %v", err)
}

fmt.Println("Run canceled successfully")

// Get the updated status
canceledRun, err := flowsClient.GetRun(ctx, runID)
if err != nil {
    log.Fatalf("Failed to get canceled run: %v", err)
}

fmt.Printf("Run status after cancellation: %s\n", canceledRun.Status)
```

## Working with Action Providers

Action providers are the building blocks for flows that perform specific actions.

### Listing Action Providers

```go
// List available action providers
providerList, err := flowsClient.ListActionProviders(ctx, nil)
if err != nil {
    log.Fatalf("Failed to list action providers: %v", err)
}

fmt.Printf("Found %d action providers:\n", len(providerList.ActionProviders))
for i, provider := range providerList.ActionProviders {
    fmt.Printf("%d. %s (%s)\n", i+1, provider.DisplayName, provider.ID)
    fmt.Printf("   %s\n", provider.Description)
}
```

### Getting Action Provider Details

```go
// Get details about a specific action provider
providerID := "https://actions.globus.org/transfer/transfer"
provider, err := flowsClient.GetActionProvider(ctx, providerID)
if err != nil {
    log.Fatalf("Failed to get action provider: %v", err)
}

fmt.Printf("Provider: %s\n", provider.DisplayName)
fmt.Printf("Description: %s\n", provider.Description)
fmt.Printf("Type: %s\n", provider.Type)
fmt.Printf("Admin Contact: %s\n", provider.AdminContact)
```

### Listing Action Roles

```go
// List roles for an action provider
roleList, err := flowsClient.ListActionRoles(ctx, providerID, 100, 0)
if err != nil {
    log.Fatalf("Failed to list action roles: %v", err)
}

fmt.Printf("Found %d roles for provider %s:\n", len(roleList.ActionRoles), providerID)
for i, role := range roleList.ActionRoles {
    fmt.Printf("%d. %s (%s)\n", i+1, role.Name, role.ID)
    fmt.Printf("   %s\n", role.Description)
    fmt.Printf("   Required Scopes: %v\n", role.RequiredScopes)
}
```

## Using Iterators for Pagination

When dealing with large numbers of flows or runs, iterators make pagination easier:

### Flow Iterator

```go
// Create a flow iterator
iterator := flowsClient.GetFlowsIterator(&flows.ListFlowsOptions{
    Title: "Transfer", // Filter by title
})

// Iterate through all flows
fmt.Println("Iterating through all transfer flows:")
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        log.Printf("Error in iterator: %v\n", err)
        break
    }
    
    flow := iterator.Flow()
    fmt.Printf("- %s (%s)\n", flow.Title, flow.ID)
}
```

### Run Iterator

```go
// Create a run iterator for a specific flow
iterator := flowsClient.GetRunsIterator(&flows.ListRunsOptions{
    FlowID: flowID,
    Status: "SUCCEEDED", // Only successful runs
})

// Iterate through all runs
fmt.Printf("Iterating through all successful runs for flow %s:\n", flowID)
for {
    hasNext := iterator.Next(ctx)
    if !hasNext {
        break
    }
    
    if err := iterator.Err(); err != nil {
        log.Printf("Error in iterator: %v\n", err)
        break
    }
    
    run := iterator.Run()
    fmt.Printf("- Run ID: %s (Started: %s)\n", run.RunID, run.StartTime)
}
```

## Convenience Methods for Getting All Items

The Flows client also provides convenient methods to get all items at once:

### Getting All Flows

```go
// Get all flows (handles pagination automatically)
allFlows, err := flowsClient.ListAllFlows(ctx, &flows.ListFlowsOptions{
    Title: "Transfer", // Filter by title
})
if err != nil {
    log.Fatalf("Failed to get all flows: %v", err)
}

fmt.Printf("Retrieved all %d transfer flows\n", len(allFlows))
```

### Getting All Runs

```go
// Get all runs for a flow (handles pagination automatically)
allRuns, err := flowsClient.ListAllRuns(ctx, &flows.ListRunsOptions{
    FlowID: flowID,
})
if err != nil {
    log.Fatalf("Failed to get all runs: %v", err)
}

fmt.Printf("Retrieved all %d runs for flow %s\n", len(allRuns), flowID)
```

### Getting All Action Providers

```go
// Get all action providers (handles pagination automatically)
allProviders, err := flowsClient.ListAllActionProviders(ctx, nil)
if err != nil {
    log.Fatalf("Failed to get all action providers: %v", err)
}

fmt.Printf("Retrieved all %d action providers\n", len(allProviders))
```

## Error Handling

The Flows service provides specific error types for better error handling:

```go
// Try to get a non-existent flow
_, err := flowsClient.GetFlow(ctx, "non-existent-flow")
if err != nil {
    switch {
    case flows.IsFlowNotFoundError(err):
        fmt.Println("Flow not found - check the flow ID")
    case flows.IsRunNotFoundError(err):
        fmt.Println("Run not found - check the run ID")
    case flows.IsActionProviderNotFoundError(err):
        fmt.Println("Action provider not found - check the provider ID")
    case flows.IsForbiddenError(err):
        fmt.Println("Permission denied - check your access token and permissions")
    case flows.IsValidationError(err):
        fmt.Println("Validation error - check your request parameters")
    default:
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Complete Example

Here's a complete example that creates a flow, runs it, and monitors its progress:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
    // Get access token from environment
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if accessToken == "" {
        log.Fatalf("GLOBUS_ACCESS_TOKEN environment variable is required")
    }
    
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    // Create SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create flows client
    flowsClient, err := config.NewFlowsClient(accessToken)
    if err != nil {
        log.Fatalf("Failed to create flows client: %v", err)
    }
    
    // Step 1: Create a Hello World flow
    flowDef := &flows.FlowCreateRequest{
        Title:       "Quick Start Example Flow",
        Description: "A simple flow created for the quick start guide",
        Definition: map[string]interface{}{
            "StartAt": "HelloWorld",
            "States": map[string]interface{}{
                "HelloWorld": map[string]interface{}{
                    "Type":      "Action",
                    "ActionUrl": "https://actions.globus.org/hello_world",
                    "Parameters": map[string]interface{}{
                        "echo_string": "${input.message}",
                    },
                    "End": true,
                },
            },
        },
    }
    
    flow, err := flowsClient.CreateFlow(ctx, flowDef)
    if err != nil {
        log.Fatalf("Failed to create flow: %v", err)
    }
    fmt.Printf("Created flow: %s (ID: %s)\n", flow.Title, flow.ID)
    
    // Step 2: Run the flow
    runReq := &flows.RunRequest{
        FlowID: flow.ID,
        Label:  "Quick Start Example Run",
        Input: map[string]interface{}{
            "message": "Hello from the Globus Go SDK!",
        },
    }
    
    run, err := flowsClient.RunFlow(ctx, runReq)
    if err != nil {
        log.Fatalf("Failed to run flow: %v", err)
    }
    fmt.Printf("Started flow run with ID: %s\n", run.RunID)
    
    // Step 3: Wait for the run to complete
    fmt.Println("Waiting for run to complete...")
    completedRun, err := flowsClient.WaitForRun(ctx, run.RunID, 2*time.Second)
    if err != nil {
        log.Fatalf("Error waiting for run: %v", err)
    }
    
    fmt.Printf("Run completed with status: %s\n", completedRun.Status)
    
    // Step 4: Get the logs
    logs, err := flowsClient.GetRunLogs(ctx, run.RunID, 10, 0)
    if err != nil {
        log.Fatalf("Failed to get run logs: %v", err)
    }
    
    fmt.Println("\nRun Logs:")
    for _, entry := range logs.LogEntries {
        fmt.Printf("[%s] %s: %s\n", entry.Time, entry.Code, entry.Description)
    }
    
    // Step 5: Display the output
    if completedRun.Status == "SUCCEEDED" {
        fmt.Println("\nRun Output:")
        fmt.Println(completedRun.Output)
    } else if completedRun.RunError != nil {
        fmt.Printf("\nRun Error: %s\n", completedRun.RunError.Description)
    }
    
    // Step 6: Clean up by deleting the flow
    fmt.Println("\nCleaning up - deleting the flow...")
    err = flowsClient.DeleteFlow(ctx, flow.ID)
    if err != nil {
        log.Fatalf("Failed to delete flow: %v", err)
    }
    fmt.Println("Flow deleted successfully")
}
```

## Next Steps

Now that you understand the basics of the Flows service, you can:

1. **Explore Action Providers**: Discover the available action providers to understand what actions you can incorporate into your flows
2. **Create Multi-step Flows**: Design flows with multiple actions that work together
3. **Implement Error Handling**: Add error handling states to your flows to make them more robust
4. **Integrate with Other Services**: Combine flows with other Globus services like Transfer, Search, and Compute

For more details, check out the [Flows Service API Reference](/docs/reference/flows/) documentation.
# Globus Flows Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Globus Flows service provides a robust platform for defining and executing workflows that can integrate multiple services. The Flows client in the Globus Go SDK allows you to create, manage, and run workflows with comprehensive support for all Flows operations.

## Key Features

- Complete API coverage for managing flows, runs, and action providers
- Iterator pattern for paginated list operations
- Batch operations for concurrent processing
- Detailed error types for improved error handling
- Helper methods for common use cases like waiting for run completion
- Comprehensive examples

## Creating a Client

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

// Create a flows client with access token
client := flows.NewClient(
    accessToken,
    core.WithLogging(true), // Optional: Enable logging
)
```

## Managing Flows

### Listing Flows

With pagination iterator:

```go
// Create an iterator for all flows
iterator := client.GetFlowsIterator(&flows.ListFlowsOptions{
    Limit: 100,
    FilterPublic: true, // Only get public flows
})

// Iterate through all pages
for iterator.Next(ctx) {
    flow := iterator.Flow()
    fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
}

// Check for errors after iteration
if err := iterator.Err(); err != nil {
    log.Fatalf("Error listing flows: %v", err)
}
```

With convenience method:

```go
// Get all flows matching criteria
allFlows, err := client.ListAllFlows(ctx, &flows.ListFlowsOptions{
    FilterOwner: "username@example.com",
})
if err != nil {
    log.Fatalf("Error listing flows: %v", err)
}

for _, flow := range allFlows {
    fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
}
```

### Getting a Flow

```go
flow, err := client.GetFlow(ctx, "flow-id")
if err != nil {
    // Check for specific error type
    if flows.IsFlowNotFoundError(err) {
        log.Fatalf("Flow not found: %v", err)
    }
    log.Fatalf("Error getting flow: %v", err)
}

fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
```

### Creating a Flow

```go
// Define flow definition (state machine)
definition := map[string]interface{}{
    "Comment": "A simple hello world flow",
    "StartAt": "HelloWorld",
    "States": map[string]interface{}{
        "HelloWorld": map[string]interface{}{
            "Type": "Action",
            "ActionUrl": "https://actions.globus.org/hello_world",
            "Parameters": map[string]interface{}{
                "echo_string": "$.input.message",
            },
            "ResultPath": "$.output",
            "End": true,
        },
    },
}

// Create flow request
createRequest := &flows.FlowCreateRequest{
    Title:       "Hello World Flow",
    Description: "A simple flow that echoes a message",
    Definition:  definition,
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "message": map[string]interface{}{
                "type": "string",
                "description": "Message to echo",
            },
        },
        "required": []string{"message"},
    },
    Keywords: []string{"example", "hello"},
    Public:   true,
}

// Create the flow
newFlow, err := client.CreateFlow(ctx, createRequest)
if err != nil {
    log.Fatalf("Error creating flow: %v", err)
}

fmt.Printf("Created flow: %s (%s)\n", newFlow.Title, newFlow.ID)
```

### Updating a Flow

```go
// Update flow request
updateRequest := &flows.FlowUpdateRequest{
    Title:       "Updated Hello World Flow",
    Description: "An updated simple flow",
    // Update other fields as needed
}

// Update the flow
updatedFlow, err := client.UpdateFlow(ctx, "flow-id", updateRequest)
if err != nil {
    log.Fatalf("Error updating flow: %v", err)
}

fmt.Printf("Updated flow: %s (%s)\n", updatedFlow.Title, updatedFlow.ID)
```

### Deleting a Flow

```go
err := client.DeleteFlow(ctx, "flow-id")
if err != nil {
    log.Fatalf("Error deleting flow: %v", err)
}
```

## Running Flows

### Starting a Flow Run

```go
// Create run request
runRequest := &flows.RunRequest{
    FlowID: "flow-id",
    Label:  "Example run",
    Input: map[string]interface{}{
        "message": "Hello, Globus Flows!",
    },
}

// Start the run
run, err := client.RunFlow(ctx, runRequest)
if err != nil {
    log.Fatalf("Error running flow: %v", err)
}

fmt.Printf("Started run: %s (status: %s)\n", run.RunID, run.Status)
```

### Waiting for a Run to Complete

```go
// Wait for run completion with a timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

finalRun, err := client.WaitForRun(ctx, "run-id", 5*time.Second)
if err != nil {
    log.Fatalf("Error waiting for run: %v", err)
}

fmt.Printf("Run completed with status: %s\n", finalRun.Status)
if finalRun.Status == "SUCCEEDED" {
    fmt.Printf("Output: %v\n", finalRun.Output)
}
```

### Listing Run Logs

```go
// Get all logs for a run
logs, err := client.ListAllRunLogs(ctx, "run-id")
if err != nil {
    log.Fatalf("Error getting run logs: %v", err)
}

fmt.Printf("Run logs (%d entries):\n", len(logs))
for i, entry := range logs {
    fmt.Printf("%d. [%s] %s\n", i+1, entry.Code, entry.Description)
}
```

## Batch Operations

### Running Multiple Flows Concurrently

```go
// Create multiple run requests
runRequests := []*flows.RunRequest{
    {
        FlowID: "flow-id",
        Label:  "Batch run 1",
        Input: map[string]interface{}{
            "message": "Batch message 1",
        },
    },
    {
        FlowID: "flow-id",
        Label:  "Batch run 2",
        Input: map[string]interface{}{
            "message": "Batch message 2",
        },
    },
}

// Run flows in batch
batchResp := client.BatchRunFlows(ctx, &flows.BatchRunFlowsRequest{
    Requests: runRequests,
    Options: &flows.BatchOptions{
        Concurrency: 5, // Run up to 5 flows concurrently
    },
})

// Process results
for i, result := range batchResp.Responses {
    if result.Error != nil {
        fmt.Printf("Error in request %d: %v\n", i, result.Error)
        continue
    }
    fmt.Printf("Run %d started: %s (status: %s)\n", 
        i, result.Response.RunID, result.Response.Status)
}
```

### Getting Multiple Flows Concurrently

```go
// Get multiple flows in batch
batchResp := client.BatchGetFlows(ctx, &flows.BatchFlowsRequest{
    FlowIDs: []string{"flow-id-1", "flow-id-2", "flow-id-3"},
})

// Process results
for i, result := range batchResp.Responses {
    if result.Error != nil {
        fmt.Printf("Error getting flow %d: %v\n", i, result.Error)
        continue
    }
    fmt.Printf("Flow %d: %s (%s)\n", 
        i, result.Flow.Title, result.Flow.ID)
}
```

## Error Handling

The Flows client provides structured error types for improved error handling:

```go
// Try to get a flow
flow, err := client.GetFlow(ctx, "non-existent-flow-id")
if err != nil {
    switch {
    case flows.IsFlowNotFoundError(err):
        fmt.Println("Flow not found")
    case flows.IsForbiddenError(err):
        fmt.Println("Permission denied")
    case flows.IsValidationError(err):
        fmt.Println("Invalid request")
    default:
        fmt.Printf("Unexpected error: %v\n", err)
    }
}
```

Available error type checks:
- `IsFlowNotFoundError(err)`
- `IsRunNotFoundError(err)`
- `IsActionProviderNotFoundError(err)`
- `IsActionRoleNotFoundError(err)`
- `IsForbiddenError(err)`
- `IsValidationError(err)`

## Action Providers

### Listing Action Providers

```go
// List action providers with iterator
iterator := client.GetActionProvidersIterator(&flows.ListActionProvidersOptions{
    FilterGlobus: true, // Only get Globus-managed providers
})

for iterator.Next(ctx) {
    provider := iterator.ActionProvider()
    fmt.Printf("Provider: %s (%s)\n", provider.DisplayName, provider.ID)
}

if err := iterator.Err(); err != nil {
    log.Fatalf("Error listing providers: %v", err)
}
```

### Getting Action Roles

```go
// List all roles for a provider
roles, err := client.ListAllActionRoles(ctx, "provider-id")
if err != nil {
    log.Fatalf("Error listing roles: %v", err)
}

fmt.Printf("Provider roles (%d):\n", len(roles))
for i, role := range roles {
    fmt.Printf("%d. %s (%s)\n", i+1, role.Name, role.ID)
}
```

## Complete Example

See the [Flows example](../examples/flows/main.go) for a complete example application that demonstrates the key features of the Flows client.
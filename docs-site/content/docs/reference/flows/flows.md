---
title: "Flows Service: Flow Operations"
---
# Flows Service: Flow Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Flows are automated workflows that orchestrate actions across different Globus services. The Flows service provides methods for creating, managing, and executing flows.

## Flow Structure

```go
type Flow struct {
    ID               string                 `json:"id"`
    Title            string                 `json:"title"`
    Description      string                 `json:"description,omitempty"`
    Definition       map[string]interface{} `json:"definition"`
    Keywords         []string               `json:"keywords,omitempty"`
    Owner            string                 `json:"owner"`
    OwnerString      string                 `json:"owner_string,omitempty"`
    VisibleTo        []string               `json:"visible_to,omitempty"`
    AdminUsers       []string               `json:"administered_by,omitempty"`
    AdminGroups      []string               `json:"admin_group_ids,omitempty"`
    Globus           bool                   `json:"globus,omitempty"`
    Subscription     string                 `json:"subscription,omitempty"`
    Created          string                 `json:"created"`
    CreatedBy        string                 `json:"created_by"`
    CreatedByString  string                 `json:"created_by_string,omitempty"`
    LastUpdated      string                 `json:"updated"`
    UpdatedBy        string                 `json:"updated_by"`
    UpdatedByString  string                 `json:"updated_by_string,omitempty"`
    InputSchema      map[string]interface{} `json:"input_schema,omitempty"`
}
```

The Flow structure contains many fields, with the most important being:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier for the flow |
| `Title` | `string` | Human-readable name for the flow |
| `Description` | `string` | Detailed description of the flow |
| `Definition` | `map[string]interface{}` | Amazon States Language definition of the workflow |
| `Keywords` | `[]string` | Tags for searching and categorizing flows |
| `Owner` | `string` | Identity ID of the flow owner |
| `VisibleTo` | `[]string` | List of principals who can see the flow |
| `AdminUsers` | `[]string` | List of users who can administer the flow |
| `AdminGroups` | `[]string` | List of groups who can administer the flow |
| `InputSchema` | `map[string]interface{}` | JSON Schema defining the required input |

## Listing Flows

Listing flows allows you to discover flows that you have access to:

```go
// List all accessible flows
flowList, err := client.ListFlows(ctx, nil)
if err != nil {
    // Handle error
}

// List with filtering options
flowList, err := client.ListFlows(ctx, &flows.ListFlowsOptions{
    Limit:       100,
    Offset:      0,
    Title:       "Data Processing",  // Filter by title (substring match)
    Owner:       "owner-id",         // Filter by owner
    OrderBy:     "created",          // Sort by creation time
    OrderDir:    "desc",             // Sort in descending order
    MarkerOnly:  false,              // Use marker-based pagination
    CreatedTime: "2023-01-01",       // Filter by creation time
    UpdatedTime: "2023-01-01",       // Filter by update time
})
if err != nil {
    // Handle error
}

// Iterate through flows
for _, flow := range flowList.Flows {
    fmt.Printf("Flow: %s (%s)\n", flow.Title, flow.ID)
    fmt.Printf("  Created: %s by %s\n", flow.Created, flow.CreatedByString)
    fmt.Printf("  Description: %s\n", flow.Description)
}
```

### Pagination

The flow list response includes pagination information:

```go
// Check pagination information
fmt.Printf("Total flows: %d\n", flowList.Total)
fmt.Printf("Has next page: %t\n", flowList.HasNextPage)
fmt.Printf("Marker: %s\n", flowList.Marker)
```

### Using Iterators for Pagination

For easier pagination, use the iterator:

```go
// Create a flow iterator
iterator := client.GetFlowsIterator(&flows.ListFlowsOptions{
    Title: "Data",
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

### Listing All Flows

To retrieve all flows automatically:

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

## Getting a Flow

Retrieving a specific flow by ID:

```go
// Get a specific flow by ID
flow, err := client.GetFlow(ctx, "flow-id")
if err != nil {
    if flows.IsFlowNotFoundError(err) {
        fmt.Println("Flow not found")
    } else {
        fmt.Printf("Error retrieving flow: %v\n", err)
    }
    return
}

fmt.Printf("Flow Title: %s\n", flow.Title)
fmt.Printf("Description: %s\n", flow.Description)
fmt.Printf("Owner: %s\n", flow.OwnerString)
fmt.Printf("Created: %s\n", flow.Created)
fmt.Printf("Last Updated: %s\n", flow.LastUpdated)
```

## Flow Definition

The flow definition uses Amazon States Language (ASL) to define the workflow:

```go
// Extract the flow definition
definition := flow.Definition

// Access definition components
startAt, ok := definition["StartAt"].(string)
if ok {
    fmt.Printf("Flow starts at state: %s\n", startAt)
}

states, ok := definition["States"].(map[string]interface{})
if ok {
    fmt.Printf("Flow has %d states\n", len(states))
    
    for stateName, stateValue := range states {
        stateMap, ok := stateValue.(map[string]interface{})
        if !ok {
            continue
        }
        
        stateType, ok := stateMap["Type"].(string)
        if ok {
            fmt.Printf("State %s has type: %s\n", stateName, stateType)
        }
    }
}
```

## Creating a Flow

Creating a new flow with a definition:

```go
// Create a new flow
createRequest := &flows.FlowCreateRequest{
    Title:       "Data Processing Flow",
    Description: "Processes data files uploaded to Globus collection",
    Definition: map[string]interface{}{
        "StartAt": "ProcessData",
        "States": map[string]interface{}{
            "ProcessData": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.function_id}",
                    "function_parameters": map[string]interface{}{
                        "input_file": "${input.file_path}",
                    },
                },
                "End": true,
            },
        },
    },
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{"compute_endpoint", "function_id", "file_path"},
        "properties": map[string]interface{}{
            "compute_endpoint": map[string]interface{}{"type": "string"},
            "function_id": map[string]interface{}{"type": "string"},
            "file_path": map[string]interface{}{"type": "string"},
        },
    },
    Keywords:    []string{"data", "processing", "compute"},
    VisibleTo:   []string{"public"},
}

flow, err := client.CreateFlow(ctx, createRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Created flow: %s (%s)\n", flow.Title, flow.ID)
```

### Flow Definition Components

A flow definition consists of:

1. **StartAt**: The name of the first state to execute
2. **States**: A map of all states in the flow
3. **State Types**: Different types of states (Action, Choice, Pass, Wait, etc.)
4. **End**: Indicates if a state is a terminal state

### Input Schema

The input schema defines the expected input for the flow using JSON Schema:

```go
// Define an input schema for a flow
inputSchema := map[string]interface{}{
    "type": "object",
    "required": []string{"source_endpoint", "destination_endpoint", "source_path", "destination_path"},
    "properties": map[string]interface{}{
        "source_endpoint": map[string]interface{}{
            "type": "string",
            "description": "Source endpoint ID",
        },
        "destination_endpoint": map[string]interface{}{
            "type": "string",
            "description": "Destination endpoint ID",
        },
        "source_path": map[string]interface{}{
            "type": "string",
            "description": "Path to source file",
        },
        "destination_path": map[string]interface{}{
            "type": "string",
            "description": "Path to destination file",
        },
    },
}
```

## Updating a Flow

Updating an existing flow:

```go
// Update an existing flow
updateRequest := &flows.FlowUpdateRequest{
    Title:       "Improved Data Processing Flow",
    Description: "Enhanced data processing with error handling",
    Definition: map[string]interface{}{
        "StartAt": "ProcessData",
        "States": map[string]interface{}{
            "ProcessData": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.function_id}",
                    "function_parameters": map[string]interface{}{
                        "input_file": "${input.file_path}",
                    },
                },
                "Catch": []map[string]interface{}{
                    {
                        "ErrorEquals": []string{"States.ALL"},
                        "Next":        "HandleError",
                    },
                },
                "End": true,
            },
            "HandleError": map[string]interface{}{
                "Type":  "Action",
                "ActionUrl": "https://actions.globus.org/notification/notify",
                "Parameters": map[string]interface{}{
                    "message": "Processing failed: ${error.Message}",
                    "email":   "${input.email}",
                },
                "End": true,
            },
        },
    },
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{"compute_endpoint", "function_id", "file_path", "email"},
        "properties": map[string]interface{}{
            "compute_endpoint": map[string]interface{}{"type": "string"},
            "function_id": map[string]interface{}{"type": "string"},
            "file_path": map[string]interface{}{"type": "string"},
            "email": map[string]interface{}{"type": "string"},
        },
    },
    Keywords:  []string{"data", "processing", "compute", "error-handling"},
    VisibleTo: []string{"public"},
}

flow, err := client.UpdateFlow(ctx, "flow-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Updated flow: %s\n", flow.Title)
```

## Deleting a Flow

Deleting a flow:

```go
// Delete a flow
err := client.DeleteFlow(ctx, "flow-id")
if err != nil {
    if flows.IsFlowNotFoundError(err) {
        fmt.Println("Flow not found")
    } else {
        fmt.Printf("Error deleting flow: %v\n", err)
    }
    return
}

fmt.Println("Flow deleted successfully")
```

## Flow Definition Patterns

### Sequential Steps

```go
// Flow with sequential steps
definition := map[string]interface{}{
    "StartAt": "Step1",
    "States": map[string]interface{}{
        "Step1": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/action1",
            "Next":     "Step2",
        },
        "Step2": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/action2",
            "Next":     "Step3",
        },
        "Step3": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/action3",
            "End":      true,
        },
    },
}
```

### Conditional Branching

```go
// Flow with conditional branching
definition := map[string]interface{}{
    "StartAt": "CheckCondition",
    "States": map[string]interface{}{
        "CheckCondition": map[string]interface{}{
            "Type": "Choice",
            "Choices": []map[string]interface{}{
                {
                    "Variable": "$.input.file_size",
                    "NumericGreaterThan": 1000000,
                    "Next": "LargeFileProcessing",
                },
                {
                    "Variable": "$.input.file_size",
                    "NumericLessThanEquals": 1000000,
                    "Next": "SmallFileProcessing",
                },
            },
            "Default": "SmallFileProcessing",
        },
        "LargeFileProcessing": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/process/large",
            "End":      true,
        },
        "SmallFileProcessing": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/process/small",
            "End":      true,
        },
    },
}
```

### Parallel Processing

```go
// Flow with parallel processing
definition := map[string]interface{}{
    "StartAt": "ParallelProcessing",
    "States": map[string]interface{}{
        "ParallelProcessing": map[string]interface{}{
            "Type": "Parallel",
            "Branches": []map[string]interface{}{
                {
                    "StartAt": "Branch1",
                    "States": map[string]interface{}{
                        "Branch1": map[string]interface{}{
                            "Type":     "Action",
                            "ActionUrl": "https://actions.globus.org/branch1",
                            "End":      true,
                        },
                    },
                },
                {
                    "StartAt": "Branch2",
                    "States": map[string]interface{}{
                        "Branch2": map[string]interface{}{
                            "Type":     "Action",
                            "ActionUrl": "https://actions.globus.org/branch2",
                            "End":      true,
                        },
                    },
                },
            },
            "Next": "AfterParallel",
        },
        "AfterParallel": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/after_parallel",
            "End":      true,
        },
    },
}
```

### Error Handling

```go
// Flow with error handling
definition := map[string]interface{}{
    "StartAt": "ProcessData",
    "States": map[string]interface{}{
        "ProcessData": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/process",
            "Catch": []map[string]interface{}{
                {
                    "ErrorEquals": []string{"States.ALL"},
                    "Next":        "HandleError",
                },
            },
            "End": true,
        },
        "HandleError": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/error_handler",
            "End":      true,
        },
    },
}
```

### Retry Logic

```go
// Flow with retry logic
definition := map[string]interface{}{
    "StartAt": "ProcessWithRetry",
    "States": map[string]interface{}{
        "ProcessWithRetry": map[string]interface{}{
            "Type":     "Action",
            "ActionUrl": "https://actions.globus.org/process",
            "Retry": []map[string]interface{}{
                {
                    "ErrorEquals": []string{"States.Timeout", "States.TaskFailed"},
                    "IntervalSeconds": 5,
                    "MaxAttempts": 3,
                    "BackoffRate": 2.0,
                },
            },
            "End": true,
        },
    },
}
```

## Common Flow Patterns

### Data Transfer Flow

```go
// Flow for transferring data
createRequest := &flows.FlowCreateRequest{
    Title:       "Data Transfer Flow",
    Description: "Transfers data between Globus endpoints",
    Definition: map[string]interface{}{
        "StartAt": "TransferData",
        "States": map[string]interface{}{
            "TransferData": map[string]interface{}{
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
                "End": true,
            },
        },
    },
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{"source_endpoint", "destination_endpoint", "source_path", "destination_path"},
        "properties": map[string]interface{}{
            "source_endpoint": map[string]interface{}{"type": "string"},
            "destination_endpoint": map[string]interface{}{"type": "string"},
            "source_path": map[string]interface{}{"type": "string"},
            "destination_path": map[string]interface{}{"type": "string"},
        },
    },
    Keywords:  []string{"transfer", "data"},
    VisibleTo: []string{"public"},
}
```

### Data Processing Pipeline

```go
// Flow for a data processing pipeline
createRequest := &flows.FlowCreateRequest{
    Title:       "Data Processing Pipeline",
    Description: "Transfers, processes, and notifies about data",
    Definition: map[string]interface{}{
        "StartAt": "TransferData",
        "States": map[string]interface{}{
            "TransferData": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/transfer/transfer",
                "Parameters": map[string]interface{}{
                    "source_endpoint_id": "${input.source_endpoint}",
                    "destination_endpoint_id": "${input.compute_endpoint}",
                    "transfer_items": []map[string]interface{}{
                        {
                            "source_path": "${input.source_path}",
                            "destination_path": "${input.compute_path}",
                        },
                    },
                },
                "Next": "ProcessData",
            },
            "ProcessData": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.function_id}",
                    "function_parameters": map[string]interface{}{
                        "input_file": "${input.compute_path}",
                        "output_file": "${input.results_path}",
                    },
                },
                "Next": "TransferResults",
            },
            "TransferResults": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/transfer/transfer",
                "Parameters": map[string]interface{}{
                    "source_endpoint_id": "${input.compute_endpoint}",
                    "destination_endpoint_id": "${input.destination_endpoint}",
                    "transfer_items": []map[string]interface{}{
                        {
                            "source_path": "${input.results_path}",
                            "destination_path": "${input.destination_path}",
                        },
                    },
                },
                "Next": "SendNotification",
            },
            "SendNotification": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/notification/notify",
                "Parameters": map[string]interface{}{
                    "message": "Data processing complete. Results available at: ${input.destination_path}",
                    "email":   "${input.email}",
                },
                "End": true,
            },
        },
    },
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{
            "source_endpoint", "compute_endpoint", "destination_endpoint",
            "source_path", "compute_path", "results_path", "destination_path",
            "function_id", "email",
        },
    },
    Keywords:  []string{"pipeline", "transfer", "compute", "notification"},
    VisibleTo: []string{"public"},
}
```

### Conditional Processing Based on File Type

```go
// Flow for processing based on file type
createRequest := &flows.FlowCreateRequest{
    Title:       "File Type Processor",
    Description: "Processes files differently based on file type",
    Definition: map[string]interface{}{
        "StartAt": "DetermineFileType",
        "States": map[string]interface{}{
            "DetermineFileType": map[string]interface{}{
                "Type": "Choice",
                "Choices": []map[string]interface{}{
                    {
                        "Variable": "$.input.file_path",
                        "StringMatches": "*.csv",
                        "Next": "ProcessCSV",
                    },
                    {
                        "Variable": "$.input.file_path",
                        "StringMatches": "*.json",
                        "Next": "ProcessJSON",
                    },
                    {
                        "Variable": "$.input.file_path",
                        "StringMatches": "*.txt",
                        "Next": "ProcessText",
                    },
                },
                "Default": "UnsupportedType",
            },
            "ProcessCSV": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.csv_function_id}",
                    "function_parameters": map[string]interface{}{
                        "file_path": "${input.file_path}",
                    },
                },
                "End": true,
            },
            "ProcessJSON": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.json_function_id}",
                    "function_parameters": map[string]interface{}{
                        "file_path": "${input.file_path}",
                    },
                },
                "End": true,
            },
            "ProcessText": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/compute/run",
                "Parameters": map[string]interface{}{
                    "endpoint_id": "${input.compute_endpoint}",
                    "function_id": "${input.text_function_id}",
                    "function_parameters": map[string]interface{}{
                        "file_path": "${input.file_path}",
                    },
                },
                "End": true,
            },
            "UnsupportedType": map[string]interface{}{
                "Type":     "Action",
                "ActionUrl": "https://actions.globus.org/notification/notify",
                "Parameters": map[string]interface{}{
                    "message": "Unsupported file type: ${input.file_path}",
                    "email":   "${input.email}",
                },
                "End": true,
            },
        },
    },
    InputSchema: map[string]interface{}{
        "type": "object",
        "required": []string{
            "compute_endpoint", "file_path", 
            "csv_function_id", "json_function_id", "text_function_id",
            "email",
        },
    },
    Keywords:  []string{"conditional", "file-processing"},
    VisibleTo: []string{"public"},
}
```

## Best Practices

1. **Design for Reusability**: Create flows that can be reused with different inputs
2. **Use Input Schema**: Define a clear input schema to validate input before execution
3. **Include Error Handling**: Use Catch blocks to handle errors gracefully
4. **Implement Retry Logic**: Add retry behavior for transient failures
5. **Use Descriptive Names**: Give states and flows clear, descriptive names
6. **Document Flows**: Provide detailed descriptions for flows and expected inputs
7. **Manage Access Control**: Set appropriate visibility and admin permissions
8. **Use Keywords**: Add relevant keywords to make flows discoverable
9. **Structure Complex Flows**: Break complex workflows into logical steps
10. **Test Thoroughly**: Test flows with various inputs and error conditions
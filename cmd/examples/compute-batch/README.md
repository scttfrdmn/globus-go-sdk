<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Globus Compute Batch Execution Examples

This example demonstrates sophisticated batch execution patterns available in the Globus Compute service using the Globus Go SDK. It shows how to:

1. Create and execute task groups for parallel processing
2. Build workflows with dependencies between tasks
3. Create dynamic dependency graphs for complex execution patterns
4. Implement error handling and recovery strategies

## Features Demonstrated

- **Task Groups:** Execute multiple similar tasks concurrently with controlled parallelism
- **Workflows:** Create multi-stage workflows with dependencies between tasks
- **Dependency Graphs:** Build complex execution patterns with conditional dependencies
- **Error Handling:** Implement retry policies and error recovery strategies
- **Dynamic Execution:** Pass results from upstream tasks to downstream tasks
- **Wait Utilities:** Use helper functions to wait for completion of complex executions

## Prerequisites

Before running this example, you need:

1. A Globus account with access to the Compute service
2. A configured Compute endpoint 
3. Valid client credentials (client ID and client secret)

## Environment Variables

The example requires the following environment variables:

```
GLOBUS_CLIENT_ID=your-client-id
GLOBUS_CLIENT_SECRET=your-client-secret
```

## Running the Example

```bash
# Navigate to the example directory
cd cmd/examples/compute-batch

# Run the example
go run main.go
```

## Example Flow

The example demonstrates three different patterns for batch execution:

### 1. Task Group Execution

Shows how to:
- Create a group of similar tasks (processing different data chunks)
- Set concurrency limits (process N tasks at once)
- Configure retry policies for the entire group
- Run the group and monitor execution progress
- Access results from all tasks

### 2. Workflow with Dependencies

Shows how to:
- Create a multi-stage workflow (processing → aggregation → reporting)
- Define dependencies between tasks
- Pass results from upstream tasks to downstream tasks
- Set different retry policies for different stages
- Handle errors at the workflow level
- Monitor the workflow's progress

### 3. Dependency Graph Execution

Shows how to:
- Create a dynamic execution graph with complex dependencies
- Add conditional execution logic
- Implement sophisticated error handling strategies
- Create fallback paths for task failures
- Monitor the execution of the entire graph

## Notes

- The example uses simulated functions and data for demonstration purposes
- In a real application, you would use actual functions that perform meaningful work
- The execution patterns shown can be applied to many data processing scenarios
- All created resources (functions, task groups, workflows) are automatically cleaned up
- The example includes sample error handling strategies that you can adapt for your needs
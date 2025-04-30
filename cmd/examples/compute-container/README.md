<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Compute Container Example

This example demonstrates how to use the Globus Go SDK to:
1. Register a container with Globus Compute
2. Register a function designed to run within a container
3. Execute the function inside the container
4. Execute direct code within the container

## Features Demonstrated

- Container registration with Docker images
- Environment variable configuration for containers
- Container arguments specification
- Function registration and execution in containers
- Direct code execution in containers
- Checking task status and retrieving results
- Cleanup of created resources

## Prerequisites

Before running this example, you need:

1. A Globus account with access to the Compute service
2. A configured Compute endpoint that can run containers
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
cd cmd/examples/compute-container

# Run the example
go run main.go
```

## Example Flow

1. **Authentication**: The example first authenticates with Globus Auth to get an access token with the appropriate scopes for Compute.

2. **Endpoint Discovery**: It then lists available Compute endpoints and selects the first one for demonstration.

3. **Container Registration**: A Docker container is registered with Python and data science libraries.

4. **Function Registration**: A data analysis function is registered that uses libraries from the container.

5. **Container Execution**: The function is executed within the container with sample data.

6. **Result Retrieval**: The example polls for task completion and displays results.

7. **Direct Code Execution**: A code snippet is executed directly in the container without registering a function.

8. **Resource Cleanup**: All created resources (functions and containers) are cleaned up at the end.

## Notes

- The example uses a Python data science container with numpy, pandas, and matplotlib
- It demonstrates generating plots inside the container (returned as base64 encoded images)
- The direct code execution feature shows how to run arbitrary code without pre-registering a function
- Environment variables can be passed to the container execution environment
- All resources created during the example are automatically deleted at the end
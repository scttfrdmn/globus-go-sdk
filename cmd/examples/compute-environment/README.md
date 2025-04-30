<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Compute Environment Configuration Example

This example demonstrates how to use the Globus Go SDK to manage environment configurations for Compute functions, including:
1. Creating and managing environment configurations
2. Setting environment variables for function execution
3. Managing secrets securely
4. Specifying resource allocation settings
5. Applying environment configurations to function execution

## Features Demonstrated

- Creating and managing environment configurations
- Storing and accessing environment variables
- Managing secrets securely
- Configuring resource allocation settings
- Applying environment configurations to functions
- Accessing environment variables within functions
- Monitoring environment and resource usage

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
cd cmd/examples/compute-environment

# Run the example
go run main.go
```

## Example Flow

1. **Authentication**: The example first authenticates with Globus Auth to get an access token with the appropriate scopes for Compute.

2. **Endpoint Discovery**: It then lists available Compute endpoints and selects the first one for demonstration.

3. **Secret Creation**: A secret is created to store an API key securely.

4. **Environment Configuration**: An environment configuration is created with environment variables, secret references, and resource allocation settings.

5. **Function Registration**: A function is registered that utilizes environment variables in its execution.

6. **Environment Application**: The environment configuration is applied to a task request.

7. **Function Execution**: The function is executed with the environment configuration.

8. **Results Analysis**: The example polls for task completion and displays how the environment variables and resources were accessed within the function.

9. **Environment Update**: The environment configuration is updated with new settings.

10. **Resource Cleanup**: All created resources (environments, secrets, and functions) are cleaned up at the end.

## Environment Configuration Features

The Globus Compute service supports several types of environment settings:

1. **Environment Variables**: Key-value pairs accessible to the function
2. **Secrets**: Securely stored sensitive values like API keys and passwords
3. **Resource Allocations**: CPU, memory, and execution time limits
4. **Runtime Parameters**: Additional runtime configuration settings

This example demonstrates all these features and shows how they can be accessed and utilized within a function.

## Notes

- Environment variables are accessible via standard OS environment variable mechanisms
- Secrets are securely stored and only accessible to authorized functions
- Resource allocations define the compute resources available to the function
- Environment configurations can be reused across multiple functions
- For optimal security, use secrets for sensitive information instead of environment variables
<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Compute Dependency Management Example

This example demonstrates how to use the Globus Go SDK to manage dependencies for Globus Compute functions, including:
1. Registering a dependency package with requirements.txt-style specifications
2. Attaching dependencies to functions
3. Running functions with their dependencies
4. Managing dependency relationships

## Features Demonstrated

- Creating and managing dependency packages
- Specifying Python package requirements with version constraints
- Attaching dependencies to functions
- Listing dependencies associated with a function
- Executing functions that use installed dependencies
- Detaching dependencies from functions
- Cleaning up resources

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
cd cmd/examples/compute-dependencies

# Run the example
go run main.go
```

## Example Flow

1. **Authentication**: The example first authenticates with Globus Auth to get an access token with the appropriate scopes for Compute.

2. **Endpoint Discovery**: It then lists available Compute endpoints and selects the first one for demonstration.

3. **Dependency Registration**: A dependency package is registered with computer vision libraries (OpenCV, NumPy, PIL, etc.).

4. **Function Registration**: An image analysis function is registered that uses these dependencies.

5. **Dependency Attachment**: The dependency package is attached to the function.

6. **Execution**: The function is executed with the URL of an image to analyze.

7. **Results**: The example polls for task completion and displays the image analysis results.

8. **Dependency Detachment**: Finally, it demonstrates detaching the dependency from the function.

9. **Resource Cleanup**: All created resources (dependencies and functions) are cleaned up at the end.

## Dependency Types Supported

The Globus Compute service supports multiple types of dependencies:

1. **Python Requirements**: Standard `requirements.txt` format with version constraints
2. **Python Packages**: Individually specified packages with versions and sources
3. **Custom Dependencies**: User-defined dependencies with custom specifications
4. **Git Repositories**: Dependencies sourced from git repositories

This example focuses on the Python Requirements format, which is the most common way to specify dependencies for Python functions.

## Notes

- The execution may take longer than other examples, as the endpoint must install the dependencies
- Dependency installation happens when the function is first executed
- For optimal performance, consider using containers with pre-installed dependencies for production use
- Requirements files can specify exact versions (`==`), minimum versions (`>=`), or leave versions unspecified
- The SDK supports dependency caching and reuse across functions
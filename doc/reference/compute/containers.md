# Compute Service: Containers

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Containers provide isolated, reproducible environments for executing functions on Globus Compute endpoints. The Compute service allows you to register, manage, and use containers for your functions.

## Container Structure

```go
type ContainerResponse struct {
    ContainerID      string            `json:"container_id"`
    Name             string            `json:"name"`
    Description      string            `json:"description,omitempty"`
    Type             string            `json:"type"`
    Location         string            `json:"location"`
    Dependencies     map[string]string `json:"dependencies,omitempty"`
    CreatedTimestamp string            `json:"created_timestamp"`
    UpdatedTimestamp string            `json:"updated_timestamp,omitempty"`
    Owner            string            `json:"owner"`
    Status           string            `json:"status,omitempty"`
}
```

The ContainerResponse structure contains the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `ContainerID` | `string` | Unique identifier for the container |
| `Name` | `string` | Human-readable name for the container |
| `Description` | `string` | Description of the container's purpose |
| `Type` | `string` | Container type (e.g., "docker", "singularity") |
| `Location` | `string` | Container image location (e.g., Docker image URL) |
| `Dependencies` | `map[string]string` | Map of dependencies with versions |
| `CreatedTimestamp` | `string` | When the container was created |
| `UpdatedTimestamp` | `string` | When the container was last updated |
| `Owner` | `string` | Identity ID of the container owner |
| `Status` | `string` | Current status of the container |

## Registering a Container

To register a new container:

```go
// Create container registration request
containerRequest := &compute.ContainerRegistrationRequest{
    Name:        "python-scientific",
    Description: "Scientific Python environment",
    Type:        "docker",
    Location:    "docker://python:3.9-slim",
    Dependencies: map[string]string{
        "numpy":      "1.22.0",
        "scipy":      "1.8.0",
        "pandas":     "1.4.0",
        "matplotlib": "3.5.0",
    },
}

// Register the container
container, err := client.RegisterContainer(ctx, containerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Container registered: %s (%s)\n", container.Name, container.ContainerID)
fmt.Printf("Type: %s\n", container.Type)
fmt.Printf("Location: %s\n", container.Location)
```

### ContainerRegistrationRequest

The `ContainerRegistrationRequest` structure is used to register a new container:

```go
type ContainerRegistrationRequest struct {
    Name         string            `json:"name"`
    Description  string            `json:"description,omitempty"`
    Type         string            `json:"type"`
    Location     string            `json:"location"`
    Dependencies map[string]string `json:"dependencies,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Name of the container (required) |
| `Description` | `string` | Description of the container's purpose |
| `Type` | `string` | Container type (required) |
| `Location` | `string` | Container image location (required) |
| `Dependencies` | `map[string]string` | Map of dependencies with versions |

### Container Types

The container type indicates the container technology being used:

- **docker**: Docker container images
- **singularity**: Singularity container images
- **shifter**: Shifter container images
- **podman**: Podman container images

### Container Locations

The location specifies where the container image can be found. It typically includes a protocol prefix:

- Docker: `docker://python:3.9-slim`
- Docker Hub: `docker://username/repository:tag`
- GitLab: `docker://registry.gitlab.com/username/repository:tag`
- Singularity: `shub://username/repository:tag`
- Custom registry: `docker://registry.example.com/image:tag`

## Listing Containers

To list available containers:

```go
// List all containers
containers, err := client.ListContainers(ctx, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d containers\n", len(containers.Containers))
for _, container := range containers.Containers {
    fmt.Printf("- %s (%s)\n", container.Name, container.ContainerID)
    fmt.Printf("  Type: %s, Location: %s\n", container.Type, container.Location)
}
```

### Container Pagination

You can control pagination for container listings:

```go
// List containers with pagination
containers, err := client.ListContainers(ctx, &compute.ContainerListOptions{
    Limit:  10,
    Offset: 20,
    Filter: "owner=me",
})
if err != nil {
    // Handle error
}

fmt.Printf("Page contains %d containers\n", len(containers.Containers))
fmt.Printf("Total containers: %d\n", containers.Total)
fmt.Printf("Has next page: %t\n", containers.HasNextPage)
```

## Getting a Container

To retrieve details about a specific container:

```go
// Get a container by ID
container, err := client.GetContainer(ctx, "container-id")
if err != nil {
    // Handle error
}

fmt.Printf("Container: %s (%s)\n", container.Name, container.ContainerID)
fmt.Printf("Description: %s\n", container.Description)
fmt.Printf("Type: %s\n", container.Type)
fmt.Printf("Location: %s\n", container.Location)
fmt.Printf("Created: %s\n", container.CreatedTimestamp)
fmt.Printf("Status: %s\n", container.Status)

// List dependencies
fmt.Println("Dependencies:")
for name, version := range container.Dependencies {
    fmt.Printf("  %s: %s\n", name, version)
}
```

## Updating a Container

To update an existing container:

```go
// Create update request
updateRequest := &compute.ContainerUpdateRequest{
    Description: "Updated scientific Python environment",
    Dependencies: map[string]string{
        "numpy":      "1.23.0",
        "scipy":      "1.9.0",
        "pandas":     "1.5.0",
        "matplotlib": "3.6.0",
        "scikit-learn": "1.1.0", // Added new dependency
    },
}

// Update the container
container, err := client.UpdateContainer(ctx, "container-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Container updated: %s\n", container.Name)
fmt.Printf("New description: %s\n", container.Description)
fmt.Printf("Updated timestamp: %s\n", container.UpdatedTimestamp)

// List updated dependencies
fmt.Println("Updated dependencies:")
for name, version := range container.Dependencies {
    fmt.Printf("  %s: %s\n", name, version)
}
```

### ContainerUpdateRequest

The `ContainerUpdateRequest` structure is used to update a container:

```go
type ContainerUpdateRequest struct {
    Description  string            `json:"description,omitempty"`
    Dependencies map[string]string `json:"dependencies,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Description` | `string` | New description for the container |
| `Dependencies` | `map[string]string` | Updated map of dependencies with versions |

Note that you cannot change the container type or location after it's registered. If you need to change these, you should register a new container.

## Deleting a Container

To delete a container:

```go
// Delete a container
err := client.DeleteContainer(ctx, "container-id")
if err != nil {
    // Handle error
}

fmt.Println("Container deleted successfully")
```

## Using Containers with Functions

When registering or updating a function, you can specify which container to use:

```go
// Register a function that uses a specific container
functionRequest := &compute.FunctionRegisterRequest{
    Name:        "data-analysis",
    Description: "Analyze data with pandas and matplotlib",
    Code: `import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from io import BytesIO
import base64

def analyze_data(data_file, columns=None):
    # Read data
    df = pd.read_csv(data_file)
    
    # Select columns to analyze
    if columns:
        df = df[columns]
    
    # Perform analysis
    result = {
        "summary": df.describe().to_dict(),
        "correlation": df.corr().to_dict(),
        "missing_values": df.isnull().sum().to_dict()
    }
    
    # Generate a plot
    plt.figure(figsize=(10, 6))
    for col in df.select_dtypes(include=[np.number]).columns:
        plt.plot(df.index, df[col], label=col)
    plt.legend()
    plt.title("Numeric Data Visualization")
    
    # Convert plot to base64 string
    buffer = BytesIO()
    plt.savefig(buffer, format='png')
    buffer.seek(0)
    plot_data = base64.b64encode(buffer.read()).decode('utf-8')
    plt.close()
    
    result["plot"] = plot_data
    return result
`,
    Entry:       "analyze_data",
    Container:   "python-scientific", // Reference the container by name
}

function, err := client.RegisterFunction(ctx, functionRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Function registered with container: %s\n", function.Container)
```

## Running Container Functions

When running a function that uses a container:

```go
// Create a task request
taskRequest := &compute.TaskRequest{
    FunctionID:  "function-id",
    EndpointID:  "endpoint-id",
    Parameters: map[string]interface{}{
        "data_file": "/path/to/data.csv",
        "columns": []string{"col1", "col2", "col3"},
    },
    Label:       "Data Analysis Task",
}

// Run the function (container is automatically used)
taskResponse, err := client.RunFunction(ctx, taskRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task submitted: %s\n", taskResponse.TaskID)

// Check task status
taskStatus, err := client.GetTaskStatus(ctx, taskResponse.TaskID)
if err != nil {
    // Handle error
}

if taskStatus.Status == "success" {
    fmt.Println("Task completed successfully")
    
    // Access result
    result := taskStatus.Result
    
    // You can now access the analysis results
    summary, _ := result["summary"].(map[string]interface{})
    correlation, _ := result["correlation"].(map[string]interface{})
    missingValues, _ := result["missing_values"].(map[string]interface{})
    
    // Access the plot data
    plotData, _ := result["plot"].(string)
    fmt.Printf("Plot data length: %d bytes\n", len(plotData))
}
```

## Container-specific Task Requests

For additional control, you can create a container-specific task:

```go
// Create a container task request
containerTaskRequest := &compute.ContainerTaskRequest{
    ContainerID: "container-id",
    EndpointID:  "endpoint-id",
    Command:     []string{"python", "-c", "print('Hello from container!')"},
    Environment: map[string]string{
        "DEBUG": "true",
        "DATA_DIR": "/data",
    },
    Volumes: []string{
        "/host/path:/container/path",
        "/tmp:/tmp",
    },
    Label: "Container Command Task",
}

// Run the container task
taskResponse, err := client.RunContainerFunction(ctx, containerTaskRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Container task submitted: %s\n", taskResponse.TaskID)
```

### ContainerTaskRequest

The `ContainerTaskRequest` structure allows you to run arbitrary commands in a container:

```go
type ContainerTaskRequest struct {
    ContainerID string            `json:"container_id"`
    EndpointID  string            `json:"endpoint_id"`
    Command     []string          `json:"command"`
    Environment map[string]string `json:"environment,omitempty"`
    Volumes     []string          `json:"volumes,omitempty"`
    Label       string            `json:"label,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `ContainerID` | `string` | ID of the container to use |
| `EndpointID` | `string` | ID of the compute endpoint |
| `Command` | `[]string` | Command to execute in the container |
| `Environment` | `map[string]string` | Environment variables |
| `Volumes` | `[]string` | Volume mounts in format "host:container" |
| `Label` | `string` | Human-readable label for the task |

## Common Container Patterns

### Multi-language Container

Create a container supporting multiple programming languages:

```go
// Register a multi-language container
containerRequest := &compute.ContainerRegistrationRequest{
    Name:        "multi-language",
    Description: "Container supporting Python, R, and Julia",
    Type:        "docker",
    Location:    "docker://custom/multi-language:latest",
    Dependencies: map[string]string{
        // Python packages
        "numpy":      "1.22.0",
        "pandas":     "1.4.0",
        // R packages
        "r-base":     "4.1.0",
        "r-tidyverse": "1.3.0",
        // Julia packages
        "julia":      "1.7.0",
        "julia-dataframes": "1.3.0",
    },
}

container, err := client.RegisterContainer(ctx, containerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Multi-language container registered: %s\n", container.ContainerID)
```

### Container with GPU Support

Create a container optimized for GPU computing:

```go
// Register a GPU-optimized container
containerRequest := &compute.ContainerRegistrationRequest{
    Name:        "gpu-tensorflow",
    Description: "TensorFlow with GPU support",
    Type:        "docker",
    Location:    "docker://tensorflow/tensorflow:latest-gpu",
    Dependencies: map[string]string{
        "cuda":       "11.2",
        "cudnn":      "8.1",
        "tensorflow": "2.8.0",
        "keras":      "2.8.0",
        "scikit-learn": "1.0.2",
    },
}

container, err := client.RegisterContainer(ctx, containerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("GPU container registered: %s\n", container.ContainerID)
```

### Domain-specific Container

Create a container for a specific domain:

```go
// Register a bioinformatics container
containerRequest := &compute.ContainerRegistrationRequest{
    Name:        "bioinformatics",
    Description: "Container for genomic data analysis",
    Type:        "docker",
    Location:    "docker://biocontainers/biocontainers:latest",
    Dependencies: map[string]string{
        "biopython":  "1.79",
        "blast":      "2.12.0",
        "samtools":   "1.14",
        "bcftools":   "1.14",
        "bedtools":   "2.30.0",
        "pandas":     "1.4.0",
        "matplotlib": "3.5.0",
    },
}

container, err := client.RegisterContainer(ctx, containerRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Bioinformatics container registered: %s\n", container.ContainerID)
```

## Best Practices

1. **Use Specific Versions**: Always specify exact version numbers for dependencies
2. **Minimize Container Size**: Include only necessary dependencies to reduce container size
3. **Test Containers**: Verify containers work correctly before using in production
4. **Use Official Base Images**: Start with trusted base images whenever possible
5. **Document Containers**: Provide detailed descriptions of container contents and usage
6. **Update Regularly**: Keep containers updated with security patches
7. **Version Containers**: Use tags or versioning in container names 
8. **Share Containers**: Make common containers available to team members
9. **Use Container Registries**: Store containers in a proper registry for versioning and access control
10. **Monitor Container Status**: Check container status before using in functions
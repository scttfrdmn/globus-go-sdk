# Compute Service: Environments

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Environments provide a way to manage configuration, settings, and secrets for compute functions. The Compute service allows you to create, manage, and apply environment configurations to your tasks.

## Environment Structure

```go
type EnvironmentResponse struct {
    EnvironmentID    string            `json:"environment_id"`
    Name             string            `json:"name"`
    Description      string            `json:"description,omitempty"`
    Variables        map[string]string `json:"variables,omitempty"`
    Secrets          []string          `json:"secrets,omitempty"`
    CreatedTimestamp string            `json:"created_timestamp"`
    UpdatedTimestamp string            `json:"updated_timestamp,omitempty"`
    Owner            string            `json:"owner"`
    Status           string            `json:"status,omitempty"`
}
```

The EnvironmentResponse structure contains the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `EnvironmentID` | `string` | Unique identifier for the environment |
| `Name` | `string` | Human-readable name for the environment |
| `Description` | `string` | Description of the environment's purpose |
| `Variables` | `map[string]string` | Environment variables |
| `Secrets` | `[]string` | Secret names available in this environment |
| `CreatedTimestamp` | `string` | When the environment was created |
| `UpdatedTimestamp` | `string` | When the environment was last updated |
| `Owner` | `string` | Identity ID of the environment owner |
| `Status` | `string` | Current status of the environment |

## Creating an Environment

To create a new environment with variables:

```go
// Create an environment for development
environmentRequest := &compute.EnvironmentCreateRequest{
    Name:        "development-environment",
    Description: "Environment for development and testing",
    Variables: map[string]string{
        "LOG_LEVEL":     "DEBUG",
        "API_URL":       "https://api-dev.example.com",
        "TIMEOUT":       "30",
        "ENABLE_CACHE":  "true",
        "MAX_RETRIES":   "3",
    },
}

// Create the environment
environment, err := client.CreateEnvironment(ctx, environmentRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Environment created: %s (%s)\n", environment.Name, environment.EnvironmentID)
```

### EnvironmentCreateRequest

The `EnvironmentCreateRequest` structure is used to create a new environment:

```go
type EnvironmentCreateRequest struct {
    Name            string            `json:"name"`
    Description     string            `json:"description,omitempty"`
    Variables       map[string]string `json:"variables,omitempty"`
    VisibleTo       []string          `json:"visible_to,omitempty"`
    ManagingUsers   []string          `json:"managing_users,omitempty"`
    ManagingGroups  []string          `json:"managing_groups,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Name of the environment (required) |
| `Description` | `string` | Description of the environment's purpose |
| `Variables` | `map[string]string` | Environment variables |
| `VisibleTo` | `[]string` | Principals who can see the environment |
| `ManagingUsers` | `[]string` | Users who can manage the environment |
| `ManagingGroups` | `[]string` | Groups who can manage the environment |

## Listing Environments

To list available environments:

```go
// List all environments
environments, err := client.ListEnvironments(ctx, nil)
if err != nil {
    // Handle error
}

fmt.Printf("Found %d environments\n", len(environments.Environments))
for _, env := range environments.Environments {
    fmt.Printf("- %s (%s)\n", env.Name, env.EnvironmentID)
    fmt.Printf("  Description: %s\n", env.Description)
}
```

### Pagination and Filtering

You can control pagination and filtering for environment listings:

```go
// List environments with options
environments, err := client.ListEnvironments(ctx, &compute.ListEnvironmentsOptions{
    Limit:  10,
    Offset: 0,
    Filter: "owner=me",
})
if err != nil {
    // Handle error
}

fmt.Printf("Page contains %d environments\n", len(environments.Environments))
fmt.Printf("Total environments: %d\n", environments.Total)
fmt.Printf("Has next page: %t\n", environments.HasNextPage)
```

## Getting an Environment

To retrieve details about a specific environment:

```go
// Get an environment by ID
environment, err := client.GetEnvironment(ctx, "environment-id")
if err != nil {
    // Handle error
}

fmt.Printf("Environment: %s (%s)\n", environment.Name, environment.EnvironmentID)
fmt.Printf("Description: %s\n", environment.Description)
fmt.Printf("Created: %s\n", environment.CreatedTimestamp)
if environment.UpdatedTimestamp != "" {
    fmt.Printf("Last updated: %s\n", environment.UpdatedTimestamp)
}

// List variables
fmt.Println("Variables:")
for key, value := range environment.Variables {
    fmt.Printf("  %s: %s\n", key, value)
}

// List secrets
if len(environment.Secrets) > 0 {
    fmt.Println("Secrets:")
    for _, secretName := range environment.Secrets {
        fmt.Printf("  %s\n", secretName)
    }
}
```

## Updating an Environment

To update an existing environment:

```go
// Create update request
updateRequest := &compute.EnvironmentUpdateRequest{
    Description: "Updated development environment",
    Variables: map[string]string{
        "LOG_LEVEL":     "INFO",           // Changed from DEBUG
        "API_URL":       "https://api-dev.example.com",
        "TIMEOUT":       "60",             // Changed from 30
        "ENABLE_CACHE":  "true",
        "MAX_RETRIES":   "5",              // Changed from 3
        "NEW_FEATURE":   "enabled",        // Added new variable
    },
}

// Update the environment
environment, err := client.UpdateEnvironment(ctx, "environment-id", updateRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Environment updated: %s\n", environment.Name)
fmt.Printf("New description: %s\n", environment.Description)
fmt.Printf("Updated timestamp: %s\n", environment.UpdatedTimestamp)

// List updated variables
fmt.Println("Updated variables:")
for key, value := range environment.Variables {
    fmt.Printf("  %s: %s\n", key, value)
}
```

### EnvironmentUpdateRequest

The `EnvironmentUpdateRequest` structure is used to update an environment:

```go
type EnvironmentUpdateRequest struct {
    Description     string            `json:"description,omitempty"`
    Variables       map[string]string `json:"variables,omitempty"`
    VisibleTo       []string          `json:"visible_to,omitempty"`
    ManagingUsers   []string          `json:"managing_users,omitempty"`
    ManagingGroups  []string          `json:"managing_groups,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Description` | `string` | New description for the environment |
| `Variables` | `map[string]string` | Updated environment variables (replaces existing) |
| `VisibleTo` | `[]string` | Updated principals who can see the environment |
| `ManagingUsers` | `[]string` | Updated users who can manage the environment |
| `ManagingGroups` | `[]string` | Updated groups who can manage the environment |

Note that updating `Variables` will replace all existing variables. To preserve existing variables while adding or changing some, first retrieve the current environment, update the variables map, and then submit the update.

## Deleting an Environment

To delete an environment:

```go
// Delete an environment
err := client.DeleteEnvironment(ctx, "environment-id")
if err != nil {
    // Handle error
}

fmt.Println("Environment deleted successfully")
```

## Managing Secrets

Secrets provide a way to store and use sensitive information like API keys, passwords, and tokens.

### Creating a Secret

To create a new secret in an environment:

```go
// Create a secret
secretRequest := &compute.SecretCreateRequest{
    Name:        "api-key",
    Value:       "secret-api-key-value",
    Description: "API key for external service",
}

// Add the secret to an environment
secret, err := client.CreateSecret(ctx, "environment-id", secretRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Secret created: %s\n", secret.Name)
fmt.Printf("Description: %s\n", secret.Description)
```

### SecretCreateRequest

The `SecretCreateRequest` structure is used to create a secret:

```go
type SecretCreateRequest struct {
    Name        string `json:"name"`
    Value       string `json:"value"`
    Description string `json:"description,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Name of the secret (required) |
| `Value` | `string` | Secret value (required) |
| `Description` | `string` | Description of the secret's purpose |

### Secret Response

The `SecretResponse` structure is returned when creating a secret:

```go
type SecretResponse struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Created     string `json:"created"`
}
```

Note that the secret value is never returned in responses for security reasons.

### Listing Secrets

To list all secrets in an environment:

```go
// List secrets in an environment
secrets, err := client.ListSecrets(ctx, "environment-id")
if err != nil {
    // Handle error
}

fmt.Printf("Found %d secrets in the environment\n", len(secrets.Secrets))
for _, secret := range secrets.Secrets {
    fmt.Printf("- %s\n", secret.Name)
    fmt.Printf("  Description: %s\n", secret.Description)
    fmt.Printf("  Created: %s\n", secret.Created)
}
```

### Deleting a Secret

To delete a secret from an environment:

```go
// Delete a secret
err := client.DeleteSecret(ctx, "environment-id", "secret-name")
if err != nil {
    // Handle error
}

fmt.Println("Secret deleted successfully")
```

## Using Environments with Tasks

When running a function, you can specify an environment to use:

```go
// Create a task request with an environment
taskRequest := &compute.TaskRequest{
    FunctionID:    "function-id",
    EndpointID:    "endpoint-id",
    Parameters:    map[string]interface{}{"input": "value"},
    Label:         "Task with environment",
    EnvironmentID: "environment-id", // Specify the environment to use
}

// Run the function with the environment
taskResponse, err := client.RunFunction(ctx, taskRequest)
if err != nil {
    // Handle error
}

fmt.Printf("Task submitted with environment: %s\n", taskResponse.TaskID)
```

### Environment Variables in Functions

In Python functions, environment variables can be accessed using the `os` module:

```python
import os

def my_function(param1, param2):
    # Access environment variables
    log_level = os.environ.get('LOG_LEVEL', 'INFO')
    api_url = os.environ.get('API_URL')
    timeout = int(os.environ.get('TIMEOUT', '30'))
    
    # Use environment variables in the function
    print(f"Running with log level: {log_level}")
    print(f"API URL: {api_url}")
    print(f"Timeout: {timeout}")
    
    # Function logic...
    return {"result": "success"}
```

### Accessing Secrets in Functions

Secrets are also made available as environment variables:

```python
import os
import requests

def external_service_call(query):
    # Access the API key secret
    api_key = os.environ.get('api-key')
    if not api_key:
        raise ValueError("API key not found in environment")
    
    # Use the secret in an API call
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    
    response = requests.get(
        f"https://api.example.com/search?q={query}",
        headers=headers
    )
    
    return response.json()
```

## Common Environment Patterns

### Environment per Stage

Create separate environments for different stages of development:

```go
// Create development environment
devEnv := &compute.EnvironmentCreateRequest{
    Name:        "development",
    Description: "Environment for development",
    Variables: map[string]string{
        "API_URL":       "https://api-dev.example.com",
        "LOG_LEVEL":     "DEBUG",
        "FEATURE_FLAGS": "all=true",
    },
}

// Create staging environment
stagingEnv := &compute.EnvironmentCreateRequest{
    Name:        "staging",
    Description: "Environment for staging/testing",
    Variables: map[string]string{
        "API_URL":       "https://api-staging.example.com",
        "LOG_LEVEL":     "INFO",
        "FEATURE_FLAGS": "beta=true,alpha=false",
    },
}

// Create production environment
prodEnv := &compute.EnvironmentCreateRequest{
    Name:        "production",
    Description: "Environment for production",
    Variables: map[string]string{
        "API_URL":       "https://api.example.com",
        "LOG_LEVEL":     "WARN",
        "FEATURE_FLAGS": "all=false",
    },
}

// Create each environment
client.CreateEnvironment(ctx, devEnv)
client.CreateEnvironment(ctx, stagingEnv)
client.CreateEnvironment(ctx, prodEnv)
```

### Environment for Services

Create environment configurations for different services:

```go
// Create an environment for database connections
dbEnv := &compute.EnvironmentCreateRequest{
    Name:        "database-config",
    Description: "Database connection settings",
    Variables: map[string]string{
        "DB_HOST":     "db.example.com",
        "DB_PORT":     "5432",
        "DB_NAME":     "mydatabase",
        "DB_USER":     "dbuser",
        "POOL_SIZE":   "10",
        "TIMEOUT":     "30",
    },
}

// Create the environment and add secrets
dbEnvironment, err := client.CreateEnvironment(ctx, dbEnv)
if err != nil {
    // Handle error
}

// Add database password as a secret
passwordSecret := &compute.SecretCreateRequest{
    Name:        "DB_PASSWORD",
    Value:       "database-password",
    Description: "Database password",
}

client.CreateSecret(ctx, dbEnvironment.EnvironmentID, passwordSecret)
```

### Feature Flag Environment

Create an environment to manage feature flags:

```go
// Create a feature flag environment
featureFlagEnv := &compute.EnvironmentCreateRequest{
    Name:        "feature-flags",
    Description: "Configuration for feature flags",
    Variables: map[string]string{
        "FEATURE_NEW_UI":          "true",
        "FEATURE_ADVANCED_SEARCH": "true",
        "FEATURE_EXPORT":          "false",
        "FEATURE_NOTIFICATIONS":   "true",
        "FEATURE_BETA":            "false",
    },
}

// Create the environment
environment, err := client.CreateEnvironment(ctx, featureFlagEnv)
if err != nil {
    // Handle error
}

// Use the feature flag environment in a function
taskRequest := &compute.TaskRequest{
    FunctionID:    "function-id",
    EndpointID:    "endpoint-id",
    Parameters:    map[string]interface{}{},
    EnvironmentID: environment.EnvironmentID,
}

client.RunFunction(ctx, taskRequest)
```

### Environment with Multiple Secrets

Create an environment with multiple secrets for API access:

```go
// Create an environment for external API access
apiEnv := &compute.EnvironmentCreateRequest{
    Name:        "external-apis",
    Description: "Configuration for external API access",
    Variables: map[string]string{
        "WEATHER_API_URL": "https://api.weather.com",
        "MAPS_API_URL":    "https://api.maps.com",
        "STOCK_API_URL":   "https://api.stocks.com",
        "API_TIMEOUT":     "30",
        "RETRY_COUNT":     "3",
    },
}

// Create the environment
environment, err := client.CreateEnvironment(ctx, apiEnv)
if err != nil {
    // Handle error
}

// Add secrets for each API
secrets := []compute.SecretCreateRequest{
    {
        Name:        "WEATHER_API_KEY",
        Value:       "weather-api-key-value",
        Description: "API key for weather service",
    },
    {
        Name:        "MAPS_API_KEY",
        Value:       "maps-api-key-value",
        Description: "API key for maps service",
    },
    {
        Name:        "STOCK_API_KEY",
        Value:       "stock-api-key-value",
        Description: "API key for stock market data",
    },
}

// Add each secret to the environment
for _, secretReq := range secrets {
    _, err := client.CreateSecret(ctx, environment.EnvironmentID, &secretReq)
    if err != nil {
        fmt.Printf("Error creating secret %s: %v\n", secretReq.Name, err)
    }
}
```

## Best Practices

1. **Naming Conventions**: Use consistent, descriptive naming for environments and variables
2. **Separate Concerns**: Create separate environments for different purposes or stages
3. **Document Usage**: Provide clear descriptions for environments and variables
4. **Secret Management**: Store sensitive information as secrets, not as regular variables
5. **Version Control**: Track environment configurations in version control when possible
6. **Default Values**: Implement fallback values in functions for optional variables
7. **Validation**: Validate required environment variables in your functions
8. **Access Control**: Restrict access to production environments and secrets
9. **Audit Changes**: Keep track of who changes environment configurations
10. **Environment Segregation**: Prevent leakage of secrets between environments
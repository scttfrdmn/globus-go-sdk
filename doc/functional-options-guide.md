# Functional Options Pattern Best Practices

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This guide provides best practices for using the functional options pattern in the Globus Go SDK. The pattern was introduced in v0.9.0 to provide a consistent API across all service clients.

## What is the Functional Options Pattern?

The functional options pattern uses variadic functions to provide a flexible and extensible way to configure objects. It allows for:

- Default values for options
- Optional parameters
- Self-documenting API with named functions
- Future extensibility without breaking changes

## Basic Pattern in the Globus Go SDK

All service clients in the Globus Go SDK follow this basic pattern:

```go
// ClientOption configures a service client
type ClientOption func(*clientOptions)

// defaultOptions returns the default client options
func defaultOptions() *clientOptions {
    return &clientOptions{
        // Default values for options
    }
}

// WithX sets option X
func WithX(x X) ClientOption {
    return func(o *clientOptions) {
        o.x = x
    }
}

// NewClient creates a new client with the provided options
func NewClient(opts ...ClientOption) (*Client, error) {
    // Apply default options
    options := defaultOptions()
    
    // Apply user options
    for _, opt := range opts {
        opt(options)
    }
    
    // Validate options
    if err := validateOptions(options); err != nil {
        return nil, err
    }
    
    // Create client
    return &Client{
        // Initialize with options
    }, nil
}
```

## Best Practices

### 1. Always Check for Errors

All client constructors now return errors. Always check for these errors to catch configuration issues early:

```go
// Good
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
)
if err != nil {
    // Handle error
}

// Bad - error is ignored
client, _ := flows.NewClient(
    flows.WithAccessToken(accessToken),
)
```

### 2. Use the Service-Specific Option Functions

Each service provides its own option functions. Use these instead of core options when possible:

```go
// Good
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithBaseURL("https://custom-url.com/"),
)

// Not recommended
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    core.WithBaseURL("https://custom-url.com/"), // Using core option
)
```

### 3. Group Options Logically

When using multiple options, group them logically for better readability:

```go
// Good - options grouped logically
client, err := flows.NewClient(
    // Authentication options
    flows.WithAccessToken(accessToken),
    
    // Networking options
    flows.WithBaseURL(customURL),
    flows.WithHTTPDebugging(true),
    
    // Performance options
    flows.WithTimeout(30 * time.Second),
)

// Not as good - options not grouped
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithTimeout(30 * time.Second),
    flows.WithBaseURL(customURL),
    flows.WithHTTPDebugging(true),
)
```

### 4. Set Only the Options You Need to Change

The functional options pattern provides sensible defaults. Only set the options you need to change:

```go
// Good - only setting options that differ from defaults
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
)

// Not necessary - setting options to their default values
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithHTTPDebugging(false), // Already false by default
    flows.WithTimeout(10 * time.Second), // Already 10s by default
)
```

### 5. Use Constants for Common Option Values

For commonly used option values, define constants to ensure consistency:

```go
// Good - using constants for common values
const (
    DefaultTimeout = 30 * time.Second
    ProductionBaseURL = "https://production.example.com/"
    StagingBaseURL = "https://staging.example.com/"
)

// Production client
prodClient, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithBaseURL(ProductionBaseURL),
    flows.WithTimeout(DefaultTimeout),
)

// Staging client
stagingClient, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithBaseURL(StagingBaseURL),
    flows.WithTimeout(DefaultTimeout),
)
```

### 6. Create Helper Functions for Common Option Combinations

If you frequently use the same combination of options, create helper functions:

```go
// Good - helper function for common option combinations
func debugOptions() []flows.ClientOption {
    return []flows.ClientOption{
        flows.WithHTTPDebugging(true),
        flows.WithHTTPTracing(true),
        flows.WithTimeout(60 * time.Second),
    }
}

// Usage
client, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    // Spread the debug options
    debugOptions()...,
)
```

### 7. Be Consistent with Option Naming

All option functions in the SDK follow the `WithX` naming convention. If you create custom options, follow this convention:

```go
// Good - follows naming convention
func WithCustomHeader(header, value string) ClientOption {
    return func(o *clientOptions) {
        if o.headers == nil {
            o.headers = make(map[string]string)
        }
        o.headers[header] = value
    }
}

// Not recommended - inconsistent naming
func SetCustomHeader(header, value string) ClientOption {
    // ...
}
```

### 8. Create Factory Functions for Related Clients

When working with multiple related clients, create factory functions to ensure consistent configuration:

```go
// Good - factory function for related clients
func NewDataServices(accessToken string) (
    *transfer.Client, *search.Client, *flows.Client, error) {
    
    // Common options
    baseURL := "https://example.com/"
    timeout := 30 * time.Second
    
    // Create transfer client
    transferClient, err := transfer.NewClient(
        transfer.WithAccessToken(accessToken),
        transfer.WithBaseURL(baseURL+"/transfer/"),
        transfer.WithTimeout(timeout),
    )
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to create transfer client: %w", err)
    }
    
    // Create search client
    searchClient, err := search.NewClient(
        search.WithAccessToken(accessToken),
        search.WithBaseURL(baseURL+"/search/"),
        search.WithTimeout(timeout),
    )
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to create search client: %w", err)
    }
    
    // Create flows client
    flowsClient, err := flows.NewClient(
        flows.WithAccessToken(accessToken),
        flows.WithBaseURL(baseURL+"/flows/"),
        flows.WithTimeout(timeout),
    )
    if err != nil {
        return nil, nil, nil, fmt.Errorf("failed to create flows client: %w", err)
    }
    
    return transferClient, searchClient, flowsClient, nil
}
```

### 9. Use the SDK Factory Methods When Possible

The SDK provides factory methods for creating clients with consistent configuration:

```go
// Good - using SDK factory methods
config := pkg.NewSDKConfig()
flowsClient, err := config.NewFlowsClient(accessToken)
if err != nil {
    // Handle error
}

// For custom options
searchClient, err := config.NewSearchClient(
    accessToken,
    search.WithHTTPDebugging(true),
)
if err != nil {
    // Handle error
}
```

### 10. Be Careful with Option Order

In general, option order doesn't matter, as each option function modifies independent fields. However, if multiple options modify the same field, the last one wins:

```go
// BaseURL will be "https://second-url.com/"
client, err := flows.NewClient(
    flows.WithBaseURL("https://first-url.com/"),
    flows.WithBaseURL("https://second-url.com/"),
)
```

## Common Options Available in All Service Clients

Most service clients provide these common options:

```go
// Authentication options
WithAccessToken(token string)
WithAuthorizer(authorizer auth.Authorizer)

// URL options
WithBaseURL(url string)

// Debugging options
WithHTTPDebugging(enable bool)
WithHTTPTracing(enable bool)

// Timeout options
WithTimeout(timeout time.Duration)

// Core options
WithCoreOption(option core.ClientOption) // For advanced use cases
```

## Tokens Package Options

The tokens package provides these options:

```go
// Storage options
WithStorage(storage Storage)
WithFileStorage(path string)
WithMemoryStorage()

// Refresh options
WithRefreshHandler(handler RefreshHandler)
WithRefreshThreshold(threshold time.Duration)
WithAutoRefresh(enable bool)
```

## Examples

### Basic Client Creation

```go
// Create a flows client with an access token
flowsClient, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
)
if err != nil {
    log.Fatalf("Failed to create flows client: %v", err)
}
```

### Client with Multiple Options

```go
// Create a search client with multiple options
searchClient, err := search.NewClient(
    search.WithAccessToken(accessToken),
    search.WithBaseURL("https://custom-search-url.com/"),
    search.WithHTTPDebugging(true),
    search.WithTimeout(30 * time.Second),
)
if err != nil {
    log.Fatalf("Failed to create search client: %v", err)
}
```

### Token Manager with Options

```go
// Create a token manager with options
manager, err := tokens.NewManager(
    tokens.WithFileStorage("~/.globus-tokens"),
    tokens.WithRefreshThreshold(30 * time.Minute),
    tokens.WithRefreshHandler(authClient),
)
if err != nil {
    log.Fatalf("Failed to create token manager: %v", err)
}
```

## Conclusion

The functional options pattern provides a flexible, extensible, and self-documenting API for configuring clients in the Globus Go SDK. By following these best practices, you can ensure that your code is maintainable, readable, and future-proof.

For more information, see:
- [V0.9.0 Migration Guide](V0.9.0_MIGRATION_GUIDE.md)
- [Client Initialization](CLIENT_INITIALIZATION.md)
- [Error Handling](ERROR_HANDLING.md)
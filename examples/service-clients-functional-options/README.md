# Service Clients with Functional Options

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This example demonstrates the consistent functional options pattern used across all service clients in the Globus Go SDK v0.9.0.

## Overview

The example showcases multiple ways to create and configure each service client:

1. Creating clients directly with functional options
2. Creating clients using SDK config helpers
3. Applying advanced configuration with core client options

All service clients in the SDK now follow the same consistent pattern for configuration and initialization.

## Running the Example

### With Globus Credentials

To run the example with real Globus credentials:

```bash
export GLOBUS_CLIENT_ID=your-client-id
export GLOBUS_CLIENT_SECRET=your-client-secret
export GLOBUS_ACCESS_TOKEN=your-access-token
go run main.go
```

### Without Credentials

The example will run with placeholder credentials if not provided:

```bash
go run main.go
```

### With HTTP Debugging

To see HTTP request/response details:

```bash
export GLOBUS_SDK_HTTP_DEBUG=1
go run main.go
```

## Code Highlights

### Creating Clients Directly with Functional Options

```go
// Auth client with functional options
authClient, err := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
    auth.WithHTTPDebugging(true),
)

// Flows client with functional options
flowsClient, err := flows.NewClient(
    flows.WithAccessToken(accessToken),
    flows.WithHTTPDebugging(true),
)

// Transfer client with functional options
transferClient, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
    transfer.WithHTTPDebugging(true),
)
```

### Creating Clients Using SDK Config

```go
// Create SDK config with credentials
config := pkg.NewConfig().
    WithClientID(clientID).
    WithClientSecret(clientSecret)

// Create clients using the config
authClient, err := config.NewAuthClient()
flowsClient, err := config.NewFlowsClient(accessToken)
transferClient, err := config.NewTransferClient(accessToken)
```

### Advanced Configuration with Core Options

```go
// Create custom core client options
coreOptions := []core.ClientOption{
    core.WithAuthorizer(authorizer),
    core.WithUserAgent("globus-go-sdk-example/1.0"),
    core.WithRequestTimeout(30 * time.Second),
}

// Apply core options to a client
advancedClient, err := flows.NewClient(
    flows.WithCoreOptions(coreOptions...),
    flows.WithHTTPDebugging(true),
)
```

## Key Concepts

1. **Functional Options Pattern**: Allows for flexible configuration with sensible defaults
2. **Consistent API**: All service clients follow the same pattern
3. **Error Handling**: All constructors return errors for validation
4. **SDK Config**: Helper methods to create clients with shared configuration
5. **Core Options**: Advanced configuration that applies to all client types

## Error Handling

The example demonstrates proper error handling for all client creation:

```go
client, err := service.NewClient(...)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```
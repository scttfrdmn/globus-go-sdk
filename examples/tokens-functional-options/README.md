# Token Management with Functional Options

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This example demonstrates the token management capabilities in the Globus Go SDK using the functional options pattern introduced in v0.9.0.

## Overview

The example showcases different ways to create and configure token managers using the functional options pattern:

1. Creating a token manager with individual options
2. Using SDK helper methods to create a token manager
3. Using in-memory storage with custom options
4. Starting background token refresh

The example also includes a mock implementation for demonstration without Globus credentials.

## Running the Example

### With Globus Credentials

To run the example with real Globus credentials:

```bash
export GLOBUS_CLIENT_ID=your-client-id
export GLOBUS_CLIENT_SECRET=your-client-secret
go run main.go
```

### Without Credentials

The example will automatically use a mock implementation if credentials are not provided:

```bash
go run main.go
```

## Code Highlights

### Creating a Token Manager with Individual Options

```go
manager, err := tokens.NewManager(
    tokens.WithFileStorage(tokenDir),
    tokens.WithAuthClient(authClient),
    tokens.WithRefreshThreshold(15 * time.Minute),
)
```

### Using SDK Helper Methods

```go
manager, err := config.WithClientID(clientID).
    WithClientSecret(clientSecret).
    NewTokenManagerWithAuth(tokenDir)
```

### Using In-Memory Storage

```go
memoryStorage := tokens.NewMemoryStorage()
    
manager, err := tokens.NewManager(
    tokens.WithStorage(memoryStorage),
    tokens.WithAuthClient(authClient),
    tokens.WithRefreshThreshold(30 * time.Minute),
)
```

### Starting Background Refresh

```go
stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Stop the refresh when the application exits
```

## Key Concepts

1. **Functional Options Pattern**: Allows for flexible configuration with sensible defaults
2. **Token Storage**: Both in-memory and file-based storage options are demonstrated
3. **Token Refresh**: Automatic token refresh when tokens are close to expiry
4. **Background Refresh**: Proactive token refresh in the background for long-running applications

## Error Handling

The example demonstrates proper error handling for all operations:

```go
manager, err := tokens.NewManager(...)
if err != nil {
    log.Fatalf("Failed to create token manager: %v", err)
}
```
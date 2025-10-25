<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4

[![Go Reference](https://pkg.go.dev/badge/github.com/scttfrdmn/globus-go-sdk/v4.svg)](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4)
[![Go Report Card](https://goreportcard.com/badge/github.com/scttfrdmn/globus-go-sdk/v4)](https://goreportcard.com/report/github.com/scttfrdmn/globus-go-sdk/v4)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**IMPORTANT: This is v4 of the Globus Go SDK. It contains breaking changes from v3.**

The Globus Go SDK v4 provides idiomatic Go bindings for the [Globus Platform](https://www.globus.org/) services, synchronized with the [Globus Python SDK v4.x](https://github.com/globus/globus-sdk-python).

## What's New in v4

v4 introduces several breaking changes for better alignment with the Python SDK and improved Go best practices:

### Key Changes

1. **Context-First Design**: All methods now require `context.Context` as the first parameter
2. **Explicit Scopes**: OAuth2 scopes must be explicitly specified in the `Config` for security
3. **Unified Config**: All clients use a `core.Config` struct instead of options pattern
4. **Enhanced Errors**: New `APIError` type with structured error details and request IDs
5. **Import Paths**: Use `/v4` in import paths: `github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth`

### Migration from v3

See the [V4_MIGRATION_GUIDE.md](../V4_MIGRATION_GUIDE.md) for detailed migration instructions.

**Side-by-side Installation**: You can use both v3 and v4 in the same project:

```go
import (
    authv3 "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    authv4 "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)
```

## Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@v4.1.0-1
```

## Requirements

- Go 1.22 or later
- A [Globus account](https://www.globus.org/)
- OAuth2 access tokens (from [Globus Auth](https://docs.globus.org/api/auth/))

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Configure client with explicit scopes (required in v4)
    config := &core.Config{
        AccessToken: "your-access-token",
        Scopes: []string{
            core.Scopes.AuthOpenID,
            core.Scopes.AuthEmail,
        },
    }

    // Create auth client
    client, err := auth.NewClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }

    // Get user info (context is always first parameter)
    userInfo, err := client.GetUserInfo(ctx)
    if err != nil {
        // Enhanced error handling in v4
        if apiErr, ok := err.(*core.APIError); ok {
            log.Printf("API error (HTTP %d): %s", apiErr.StatusCode, apiErr.Message)
            log.Printf("Request ID: %s", apiErr.RequestID)
        } else {
            log.Fatal(err)
        }
        return
    }

    fmt.Printf("User: %s (%s)\n", userInfo.Name, userInfo.Email)
}
```

### Working with Groups

```go
import "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"

// Configure with groups scopes
config := &core.Config{
    AccessToken: "your-access-token",
    Scopes: []string{
        core.Scopes.GroupsView,
    },
}

// Create groups client
client, err := groups.NewClient(ctx, config)
if err != nil {
    log.Fatal(err)
}

// List your groups
options := &groups.ListGroupsOptions{
    MyGroups: true,
    PageSize: 10,
}

groupList, err := client.ListGroups(ctx, options)
if err != nil {
    log.Fatal(err)
}

for _, group := range groupList.Groups {
    fmt.Printf("Group: %s (ID: %s)\n", group.Name, group.ID)
}
```

## Features

### Implemented Services

- ✅ **Auth**: OAuth2 flows, token management, user info, projects
- ✅ **Groups**: Group management, membership operations
- 🚧 **Transfer**: File transfer operations (coming soon)
- 🚧 **Search**: Search index operations (coming soon)
- 🚧 **Flows**: Flow management and execution (coming soon)
- 🚧 **Timers**: Timer scheduling (coming soon)
- 🚧 **Compute**: Compute endpoint management (coming soon)

### Core Features

- **Context-First Design**: Full context support for cancellation and timeouts
- **Enhanced Error Handling**: Structured `APIError` with request IDs and notes
- **Explicit Scopes**: Security-focused scope specification
- **Retry Logic**: Automatic retries with exponential backoff
- **Well-Known Scopes**: Pre-defined scope constants (`core.Scopes`)
- **Type Safety**: Strongly-typed request and response models

## Configuration

### Config Structure

```go
config := &core.Config{
    // Required
    AccessToken: "your-token",
    Scopes: []string{core.Scopes.AuthOpenID},

    // Optional
    HTTPClient: customHTTPClient,
    BaseURL: "https://custom-api.globus.org",
    Timeout: 60 * time.Second,
    RetryConfig: customRetryConfig,
    UserAgent: "MyApp/1.0",
    Environment: "production",
}
```

### Retry Configuration

```go
retryConfig := &core.RetryConfig{
    MaxRetries: 5,
    InitialBackoff: 2 * time.Second,
    MaxBackoff: 60 * time.Second,
    BackoffMultiplier: 2.0,
    RetryableStatusCodes: []int{429, 500, 502, 503, 504},
}

config := &core.Config{
    AccessToken: token,
    Scopes: scopes,
    RetryConfig: retryConfig,
}
```

## Error Handling

v4 introduces enhanced error types:

```go
userInfo, err := client.GetUserInfo(ctx)
if err != nil {
    // Check for specific error types
    if apiErr, ok := err.(*core.APIError); ok {
        fmt.Printf("HTTP Status: %d\n", apiErr.StatusCode)
        fmt.Printf("Error Code: %s\n", apiErr.Code)
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)

        // Use helper methods
        if apiErr.IsAuthError() {
            fmt.Println("Authentication failed")
        } else if apiErr.IsNotFound() {
            fmt.Println("Resource not found")
        } else if apiErr.IsRateLimited() {
            fmt.Println("Rate limited")
        }
    } else if valErr, ok := err.(*core.ValidationError); ok {
        fmt.Printf("Validation error: %s\n", valErr.Message)
    } else if netErr, ok := err.(*core.NetworkError); ok {
        fmt.Printf("Network error: %s\n", netErr.Message)
    }
    return
}
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/basic-usage/` - Basic Auth and Groups operations

## Documentation

- [GoDoc](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4)
- [V4 Migration Guide](../V4_MIGRATION_GUIDE.md)
- [V4 Implementation Plan](../V4_IMPLEMENTATION_PLAN.md)
- [Globus Platform Docs](https://docs.globus.org/)

## Versioning

This SDK follows semantic versioning and is synchronized with the Globus Python SDK:

- **v4.1.0-1**: Synchronized with Python SDK v4.1.0 (first Go SDK v4 release)
- Format: `vX.Y.Z-N` where `N` is the Go SDK release number for that Python SDK version

## Python SDK Compatibility

| Go SDK | Python SDK | Status |
|--------|------------|--------|
| v4.1.0-1 | v4.1.0 | ✅ In Development |

## Contributing

Contributions are welcome! This project maintains synchronization with the Globus Python SDK, so please check existing issues and the implementation plan before contributing.

## License

Apache 2.0 - See [LICENSE](../LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/scttfrdmn/globus-go-sdk/issues)
- **Discussions**: [GitHub Discussions](https://github.com/scttfrdmn/globus-go-sdk/discussions)
- **Globus Support**: [support@globus.org](mailto:support@globus.org)

## Acknowledgments

This SDK is synchronized with the excellent [Globus Python SDK](https://github.com/globus/globus-sdk-python) maintained by the Globus team.

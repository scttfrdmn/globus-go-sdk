---
title: "Globus Go SDK"
type: docs
---

# Globus Go SDK

The Globus Go SDK provides a client library for interacting with Globus platform services using the Go programming language. It enables Go applications to authenticate users, transfer files, search datasets, manage groups, schedule tasks, and more using Globus infrastructure.

> **Current Version: v0.9.2** - See the [Changelog](https://github.com/scttfrdmn/globus-go-sdk/blob/main/CHANGELOG.md) for details on all releases.

## Features

- **Authentication**: OAuth2 authentication flows, token management, and MFA support
- **Transfer**: File transfer operations between Globus endpoints, including recursion and resumability
- **Search**: Index and query data using the Globus Search service
- **Flows**: Create and manage automated workflows
- **Compute**: Execute functions on remote compute endpoints with workflow support
- **Groups**: Manage collaborative groups and memberships
- **Timers**: Schedule tasks to run at specific times

## Getting Started

```go
import (
    "context"
    "log"

    "github.com/scttfrdmn/globus-go-sdk/pkg/globus"
)

func main() {
    // Create SDK configuration with options
    sdk, err := globus.NewSDKFromEnvironment(
        globus.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
        globus.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
    )
    if err != nil {
        log.Fatalf("Failed to create SDK: %v", err)
    }

    // Get an auth client
    authClient := sdk.Auth()

    // Use the client
    userInfo, err := authClient.GetUserInfo(context.Background())
    if err != nil {
        log.Fatalf("Failed to get user info: %v", err)
    }

    log.Printf("Logged in as: %s <%s>", userInfo.Name, userInfo.Email)
}
```

## Documentation

- [Quick Start Guides](/docs/guides/quickstart/) - Get up and running with each service
- [Comprehensive Guides](/docs/guides/) - In-depth guides for common tasks
- [API Reference](/docs/reference/) - Complete API documentation for all services
- [Examples](/docs/examples/) - Example applications and use cases
- [FAQ](/docs/faq/) - Answers to frequently asked questions

## Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk@v0.9.2
```

## Advanced Features

- **Advanced Compute Workflows**: Create and manage complex compute workflows with task dependencies
- **Recursive Transfers**: Transfer entire directory structures with resumability
- **Multi-Factor Authentication**: Support for TOTP, WebAuthn, and backup codes
- **Advanced Search Queries**: Build complex search queries with full-text, geolocation, and more
- **Connection Pooling**: Optimized connection management for high-performance applications
- **Rate Limiting & Circuit Breaking**: Protect against API rate limits and service outages

## License

Apache License 2.0
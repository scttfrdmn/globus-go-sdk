---
title: "Globus Go SDK"
type: docs
---

# Globus Go SDK

The Globus Go SDK provides a client library for interacting with Globus platform services using the Go programming language. It enables Go applications to authenticate users, transfer files, search datasets, manage groups, schedule tasks, and more using Globus infrastructure.

## Features

- **Authentication**: OAuth2 authentication flows, token management, and MFA support
- **Transfer**: File transfer operations between Globus endpoints, including recursion and resumability
- **Search**: Index and query data using the Globus Search service
- **Flows**: Create and manage automated workflows
- **Compute**: Execute functions on remote compute endpoints
- **Groups**: Manage collaborative groups and memberships
- **Timers**: Schedule tasks to run at specific times

## Getting Started

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// Create SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create an auth client
authClient, err := config.NewAuthClient()
if err != nil {
    // Handle error
}

// Use the client
userInfo, err := authClient.GetUserInfo(context.Background())
```

## Documentation

- [API Reference](/docs/reference/) - Complete API documentation for all services
- [Guides](/docs/guides/) - How-to guides for common tasks
- [Examples](/docs/examples/) - Example applications and use cases

## Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk
```

## License

Apache License 2.0
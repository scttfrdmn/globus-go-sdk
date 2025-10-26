# Globus Go SDK

Welcome to the Globus Go SDK documentation! This SDK provides a simple and idiomatic Go interface to Globus APIs.

## Overview

The Globus Go SDK enables Go developers to integrate with Globus services for file transfer, search, groups, workflows, and more. It provides a clean, type-safe API with comprehensive error handling and testing support.

## Key Features

- **🎯 Idiomatic Go** - Follows Go best practices and conventions
- **🔒 Type Safety** - Strong typing with explicit struct definitions
- **⚡ High Performance** - Efficient connection pooling and resource management
- **🔄 Python SDK Parity** - Maintains compatibility with Python SDK patterns
- **📦 Comprehensive Services** - Full support for all Globus services
- **🧪 Well Tested** - Extensive test coverage with mock support
- **📚 Well Documented** - Complete API documentation and examples
- **🔐 Secure** - OAuth2 authentication with token management

## Python SDK Alignment

This SDK maintains synchronization with the upstream [Python Globus SDK](https://github.com/globus/globus-sdk-python), currently tracking:

- **v3.x line**: Latest v3.43.0 features
- **v4.x line**: Latest v4.1.0 features

See [Python SDK Alignment](alignment.md) for details on feature parity.

## Supported Services

| Service | Status | Description |
|---------|--------|-------------|
| **Auth** | ✅ Complete | Authentication and identity management |
| **Transfer** | ✅ Complete | File and directory transfer operations |
| **Search** | ✅ Complete | Search index management and queries |
| **Groups** | ✅ Complete | Group membership and policy management |
| **Flows** | ✅ Complete | Workflow automation and management |
| **Timers** | ✅ Complete | Scheduled task execution |
| **Compute** | ✅ Complete | Distributed computing operations |

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
    // Create authorizer with access token
    auth := authorizers.NewAccessTokenAuthorizer("your-access-token")

    // Create transfer client
    client := transfer.NewClient(auth)

    // List endpoints
    ctx := context.Background()
    endpoints, err := client.EndpointList(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, ep := range endpoints.Data {
        fmt.Printf("Endpoint: %s (%s)\n", ep.DisplayName, ep.ID)
    }
}
```

## Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v3
```

See the [Installation Guide](getting-started/installation.md) for detailed instructions.

## Quick Links

- [Installation Guide](getting-started/installation.md) - Get started with the SDK
- [Quick Start](getting-started/quickstart.md) - Your first SDK code
- [API Reference](api/index.md) - Complete API documentation
- [Examples](examples/transfer.md) - Practical code examples
- [pkg.go.dev](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3) - Go package documentation

## Getting Help

If you need help or want to report an issue:

- Check the [API Reference](api/index.md) for detailed service documentation
- Review [Common Patterns](guides/common-patterns.md) for best practices
- Check [pkg.go.dev](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3) for API docs
- Open an issue on [GitHub](https://github.com/scttfrdmn/globus-go-sdk/issues)

## About This Project

The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus or the University of Chicago. It is maintained by independent contributors and designed to maintain parity with the upstream Python SDK.

## Next Steps

Ready to get started? Head over to the [Installation Guide](getting-started/installation.md) to add the SDK to your project.

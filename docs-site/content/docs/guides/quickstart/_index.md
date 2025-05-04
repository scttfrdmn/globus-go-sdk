---
title: "Quick Start Guides"
weight: 10
bookCollapseSection: true
---

# Quick Start Guides

Welcome to the Globus Go SDK Quick Start Guides. These guides are designed to help you get up and running with the SDK's services as quickly as possible. Each guide focuses on a single service and provides step-by-step instructions for common tasks.

## Available Quick Start Guides

- [Auth Service](auth) - Authentication, OAuth2 flows, and token management
- Transfer Service - File transfers between Globus endpoints
- Search Service - Indexing and searching data
- Flows Service - Managing and executing automated workflows
- Compute Service - Remote function execution and container management
- Groups Service - Group management and membership operations
- Timers Service - Scheduling tasks and operations

## Prerequisites

Before using any of the services, make sure you have:

1. **Go 1.18 or higher** installed
2. Added the Globus Go SDK to your project:
   ```bash
   go get github.com/scttfrdmn/globus-go-sdk
   ```
3. **Globus Account** with appropriate permissions for the services you want to use
4. **Client Credentials** (client ID and secret) for your application

## Common Import Pattern

All quick start guides use the following import pattern:

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    // Service-specific imports, e.g.:
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)
```

## Using Environment Variables

The SDK supports configuration via environment variables. For local development, you can use a `.env.test` file with the following variables:

```bash
GLOBUS_CLIENT_ID=your-client-id
GLOBUS_CLIENT_SECRET=your-client-secret
GLOBUS_ACCESS_TOKEN=your-access-token
GLOBUS_REFRESH_TOKEN=your-refresh-token
```

Then load these variables in your code:

```go
config := pkg.NewConfigFromEnvironment()
```

## Next Steps

After going through the quick start guides, check out the [API Reference](/docs/reference/) for comprehensive documentation on all services and their methods.
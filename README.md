# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

<p align="center">
  <img src="doc/images/globus-go-sdk-logo.png" alt="Globus Go SDK Logo" width="400"/>
</p>

<h1 align="center">Globus Go SDK</h1>

<p align="center">
  <a href="https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk"><img src="https://pkg.go.dev/badge/github.com/scttfrdmn/globus-go-sdk.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/scttfrdmn/globus-go-sdk"><img src="https://goreportcard.com/badge/github.com/scttfrdmn/globus-go-sdk" alt="Go Report Card"></a>
  <a href="https://github.com/scttfrdmn/globus-go-sdk/actions/workflows/go.yml"><img src="https://github.com/scttfrdmn/globus-go-sdk/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/scttfrdmn/globus-go-sdk/actions/workflows/docs.yml"><img src="https://github.com/scttfrdmn/globus-go-sdk/actions/workflows/docs.yml/badge.svg" alt="Documentation Status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/scttfrdmn/globus-go-sdk" alt="License"></a>
  <a href="https://github.com/scttfrdmn/globus-go-sdk/releases"><img src="https://img.shields.io/github/v/release/scttfrdmn/globus-go-sdk" alt="Release"></a>
  <a href="https://codecov.io/gh/scttfrdmn/globus-go-sdk"><img src="https://codecov.io/gh/scttfrdmn/globus-go-sdk/branch/main/graph/badge.svg" alt="Coverage"></a>
</p>

A Go SDK for interacting with Globus services, providing a simple and idiomatic Go interface to Globus APIs.

> **STATUS**: Version 0.9.0 is now available! This version introduces a consistent API pattern across all service clients using the functional options pattern, provides comprehensive token management capabilities, improves error handling, and enhances the Compute client with workflow and task group capabilities. See the [CHANGELOG](doc/project/changelog.md) for information on all features and improvements. If you're upgrading from v0.8.0, check the [Migration Guide](doc/V0.9.0_MIGRATION_GUIDE.md).

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations. This SDK is maintained by independent contributors and is not a product of Globus or the University of Chicago.

## Features

- **Authentication**: OAuth2 flows with token management and automatic refreshing
- **Token Management**: Complete token lifecycle management with automatic refreshing
- **Token Storage**: Persistent token storage (memory and file-based implementations)
- **Groups**: Group management and membership operations
- **Transfer**: File transfer with recursive directory support and resumable transfers
- **Search**: Advanced search capabilities with query building and pagination
- **Flows**: Automation workflows with batch operations
- **Compute**: Function execution, endpoint management, workflows, and task groups
- **Performance**: Connection pooling, rate limiting, and backoff strategies
- **Observability**: Structured logging and distributed tracing
- **Reliability**: Comprehensive error handling with retries and circuit breakers
- **Compatibility**: API version compatibility checking and management
- **Integration**: Extensive testing infrastructure
- **Examples**: CLI, token management, compute workflows, and web application examples
- **Utilities**: Verification tools to test credentials

## Installation

### Requirements

- Go 1.18 or higher (uses generics for some utility functions)
- No external dependencies except for testing

### Using `go get`

```bash
go get github.com/scttfrdmn/globus-go-sdk
```

### Using Go modules in your project

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg"
```

## Testing with Globus Credentials

This SDK includes a standalone credential verification tool to test your Globus credentials:

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/globus-go-sdk.git
cd globus-go-sdk

# Copy the example .env.test file and add your credentials
cp .env.test.example .env.test
# Edit .env.test with your favorite editor and add your credentials

# Build and run the verification tool
cd cmd/verify-credentials
go build
./verify-credentials
```

### Comprehensive Testing

For complete SDK testing, including all services with real credentials:

```bash
# After setting up your .env.test file
./scripts/comprehensive_testing.sh  # Run all tests with real credentials
```

This script runs all unit tests, examples, integration tests, and verifies all services work with your real credentials. Results are logged to `comprehensive_testing.log`.

See [doc/COMPREHENSIVE_TESTING.md](doc/COMPREHENSIVE_TESTING.md) for detailed testing guidelines.

### Integration Testing

For targeted integration testing:

```bash
# After setting up your .env.test file
./scripts/run_integration_tests.sh --verify  # Verify credentials only
./scripts/run_integration_tests.sh           # Run all integration tests
```

See [doc/INTEGRATION_TESTING.md](doc/INTEGRATION_TESTING.md) for more details on integration testing with credentials.

## SDK Organization

This SDK follows the structure and patterns established by the official Globus Python and JavaScript SDKs. It is organized into service-specific packages that each provide clients for interacting with Globus services.

### Core Components

- `pkg/core`: Base client functionality, error handling, logging, etc.
- `pkg/core/authorizers`: Authentication mechanisms
- `pkg/core/transport`: HTTP transport layer
- `pkg/core/config`: Configuration management

### Services

- `pkg/services/auth`: OAuth2 authentication and authorization
- `pkg/services/tokens`: Token management with storage and automatic refreshing
- `pkg/services/groups`: Group management
- `pkg/services/transfer`: File transfer and endpoint management
- `pkg/services/search`: Data search and discovery
- `pkg/services/flows`: Automation and workflow orchestration
- `pkg/services/compute`: Distributed computation and function execution

## Quick Start

### Authentication with Token Management

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
    // Create a token storage for persisting tokens
    storage, err := tokens.NewFileStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }
    
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

    // Create a new Auth client
    authClient, err := config.NewAuthClient()
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    authClient.SetRedirectURL("http://localhost:8080/callback")
    
    // Create a token manager for automatic refresh using the functional options pattern
    tokenManager, err := tokens.NewManager(
        tokens.WithStorage(storage),
        tokens.WithRefreshHandler(authClient),
    )
    if err != nil {
        log.Fatalf("Failed to create token manager: %v", err)
    }
    
    // Configure token refresh settings
    tokenManager.SetRefreshThreshold(5 * time.Minute)
    
    // Start background refresh
    stopRefresh := tokenManager.StartBackgroundRefresh(15 * time.Minute)
    defer stopRefresh() // Stop background refresh when done
    
    // Check if we already have tokens
    entry, err := tokenManager.GetToken(context.Background(), "default")
    if err == nil && !entry.TokenSet.IsExpired() {
        // We have valid tokens, use them
        fmt.Printf("Using existing token (expires at: %s)\n", entry.TokenSet.ExpiresAt.Format(time.RFC3339))
        return
    }
    
    // We need new tokens, start the OAuth2 flow
    authURL := authClient.GetAuthorizationURL("my-state")
    fmt.Printf("Visit this URL to log in: %s\n", authURL)
    
    // Start a local server to handle the callback
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        
        // Exchange code for tokens
        tokenResponse, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
        if err != nil {
            log.Fatalf("Failed to exchange code: %v", err)
            http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
            return
        }
        
        // Create a token entry
        entry := &tokens.Entry{
            Resource: "default",
            TokenSet: &tokens.TokenSet{
                AccessToken:  tokenResponse.AccessToken,
                RefreshToken: tokenResponse.RefreshToken,
                ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
                Scope:        tokenResponse.Scope,
            },
        }
        
        // Store the tokens
        err = storage.Store(entry)
        if err != nil {
            log.Fatalf("Failed to store token: %v", err)
            http.Error(w, "Failed to store token", http.StatusInternalServerError)
            return
        }
        
        fmt.Fprintf(w, "Authentication successful! You can close this window.")
        fmt.Printf("Authentication successful! Tokens stored.\n")
    })
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Groups

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
)

func main() {
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()

    // Create a new Groups client with an access token using the functional options pattern
    groupsClient, err := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create groups client: %v", err)
    }
    
    // List groups the user is a member of
    groupList, err := groupsClient.ListGroups(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to list groups: %v", err)
    }
    
    fmt.Printf("You are a member of %d groups:\n", len(groupList.Groups))
    for _, group := range groupList.Groups {
        fmt.Printf("- %s (%s)\n", group.Name, group.ID)
    }
}
```

### Transfer

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()

    // Create a new Transfer client with an access token using the functional options pattern
    transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // List endpoints the user has access to
    endpoints, err := transferClient.ListEndpoints(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to list endpoints: %v", err)
    }
    
    fmt.Printf("Found %d endpoints:\n", len(endpoints.DATA))
    for i, endpoint := range endpoints.DATA {
        fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
    }
    
    // Submit a file transfer (if source and destination endpoints are provided)
    sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
    destEndpointID := os.Getenv("DEST_ENDPOINT_ID")
    
    if sourceEndpointID != "" && destEndpointID != "" {
        // Regular file transfer
        task, err := transferClient.SubmitTransfer(
            context.Background(),
            sourceEndpointID,
            destEndpointID,
            &transfer.TransferData{
                Label: "SDK Example Transfer",
                Items: []transfer.TransferItem{
                    {
                        Source:      "/~/source.txt",
                        Destination: "/~/destination.txt",
                    },
                },
                SyncLevel: transfer.SyncChecksum,
                Verify:    true,
            },
        )
        if err != nil {
            log.Fatalf("Failed to submit transfer: %v", err)
        }
        
        fmt.Printf("Transfer submitted, task ID: %s\n", task.TaskID)
        
        // Monitor the task
        for {
            status, err := transferClient.GetTaskStatus(context.Background(), task.TaskID)
            if err != nil {
                log.Fatalf("Failed to get task status: %v", err)
            }
            
            fmt.Printf("Task status: %s (%d/%d files)\n", 
                status.Status, status.FilesTransferred, status.FilesTotal)
                
            if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
                break
            }
            
            time.Sleep(2 * time.Second)
        }
        
        // Submit a recursive directory transfer
        result, err := transferClient.SubmitRecursiveTransfer(
            context.Background(),
            sourceEndpointID, "/~/source_dir",
            destEndpointID, "/~/dest_dir",
            &transfer.RecursiveTransferOptions{
                Label:     "SDK Example Recursive Transfer",
                SyncLevel: transfer.SyncChecksum,
                BatchSize: 100,
            },
        )
        if err != nil {
            log.Fatalf("Failed to submit recursive transfer: %v", err)
        }
        
        fmt.Printf("Recursive transfer submitted with %d tasks\n", len(result.TaskIDs))
        fmt.Printf("Transferred %d items\n", result.ItemsTransferred)
    }
}
```

### Search

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
)

func main() {
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()

    // Create a new Search client with an access token using the functional options pattern
    searchClient, err := config.NewSearchClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create search client: %v", err)
    }
    
    // List indexes the user has access to
    indexes, err := searchClient.ListIndexes(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to list indexes: %v", err)
    }
    
    fmt.Printf("Found %d indexes:\n", len(indexes.Indexes))
    for i, index := range indexes.Indexes {
        fmt.Printf("%d. %s (%s)\n", i+1, index.DisplayName, index.ID)
    }
    
    // Search an index (if index ID is provided)
    indexID := os.Getenv("GLOBUS_SEARCH_INDEX_ID")
    if indexID != "" {
        searchReq := &pkg.SearchRequest{
            IndexID: indexID,
            Query:   "example",
            Options: &pkg.SearchOptions{
                Limit: 10,
            },
        }
        
        results, err := searchClient.Search(context.Background(), searchReq)
        if err != nil {
            log.Fatalf("Failed to search: %v", err)
        }
        
        fmt.Printf("Search found %d results\n", results.Count)
        for i, result := range results.Results {
            fmt.Printf("%d. %s (Score: %.2f)\n", i+1, result.Subject, result.Score)
        }
    }
}
```

### Flows

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
)

func main() {
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()

    // Create a new Flows client with an access token using the functional options pattern
    flowsClient, err := config.NewFlowsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create flows client: %v", err)
    }
    
    // List flows the user has access to
    flows, err := flowsClient.ListFlows(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to list flows: %v", err)
    }
    
    fmt.Printf("Found %d flows:\n", len(flows.Flows))
    for i, flow := range flows.Flows {
        fmt.Printf("%d. %s (%s)\n", i+1, flow.Title, flow.ID)
    }
    
    // Run a flow (if flow ID is provided)
    flowID := os.Getenv("GLOBUS_FLOW_ID")
    if flowID != "" {
        runReq := &pkg.RunRequest{
            FlowID: flowID,
            Label:  "Example Run " + time.Now().Format("20060102"),
            Input: map[string]interface{}{
                "message": "Hello from Globus Go SDK!",
            },
        }
        
        run, err := flowsClient.RunFlow(context.Background(), runReq)
        if err != nil {
            log.Fatalf("Failed to run flow: %v", err)
        }
        
        fmt.Printf("Flow run started: %s (Status: %s)\n", run.RunID, run.Status)
    }
}
```

## Documentation

For detailed documentation, see:

- [GoDoc Reference](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/)
- [User Guide](doc/user-guide.md)
- [Quick Start Examples](doc/QUICK_START_EXAMPLES.md)
- [v0.9.0 Migration Guide](doc/V0.9.0_MIGRATION_GUIDE.md)
- [v0.8.0 Migration Guide](doc/V0.8.0_MIGRATION_GUIDE.md)
- [Client Initialization](doc/CLIENT_INITIALIZATION.md)
- [Functional Options Pattern Best Practices](doc/functional-options-guide.md)
- [Error Handling](doc/ERROR_HANDLING.md)
- [Token Management Example](examples/token-management/README.md)
- [Tokens Package Guide](doc/tokens-package.md)
- [Token Storage Guide](doc/token-storage.md)
- [Recursive Transfers Guide](doc/recursive-transfers.md)
- [Resumable Transfers Guide](doc/resumable-transfers.md)
- [Logging and Tracing Guide](doc/logging-and-tracing.md)
- [Integration Testing Guide](doc/INTEGRATION_TESTING.md)
- [Search Client Guide](doc/search-client.md)
- [Timers Client Guide](doc/timers-client.md)
- [Memory Optimization Guide](doc/memory-optimization.md)
- [Connection Pooling Guide](doc/connection-pooling.md)
- [MFA Authentication Guide](doc/mfa-authentication.md)
- [Shell Testing Guide](doc/shell-testing.md)
- [Security Guidelines](doc/SECURITY_GUIDELINES.md)
- [Security Audit Plan](doc/SECURITY_AUDIT_PLAN.md)
- [Data Schemas](doc/data-schemas.md)
- [Extending the SDK](doc/extending-the-sdk.md)
- [CLI Example](cmd/globus-cli/README.md)
- [Compute Workflows Example](examples/compute-workflow/main.go) ← New for v0.9.0
- [Compute Workflows Guide](doc/compute-workflows.md) ← New for v0.9.0

## Development Status

This SDK is under active development. Current version: **v0.9.0**

| Component | Status | Details |
|-----------|--------|---------|
| Core Infrastructure | ✅ Complete | Base client, transport, authorizers, logging |
| Auth Client | ✅ Complete | OAuth flows, token management, validation utilities |
| Token Storage | ✅ Complete | Interface with memory and file-based implementations |
| Token Manager | ✅ Complete | Automatic token refreshing and management |
| Tokens Package | ✅ Complete | Unified token management package with storage and refresh |
| Groups Client | ✅ Complete | Group management, membership operations |
| Transfer Client | ✅ Complete | Basic operations, recursive directory transfers, resumable transfers |
| Search Client | ✅ Complete | Advanced queries, batch operations, pagination |
| Flows Client | ✅ Complete | Flow discovery, execution, management |
| Timers Client | ✅ Complete | Creating and managing scheduled tasks |
| CLI Example | ✅ Complete | Command-line application showcasing SDK features |
| Compute Client | ✅ Complete | Function execution, endpoint management, workflows, task groups |

See [RELEASE_CHECKLIST.md](RELEASE_CHECKLIST.md) for detailed status, [KNOWN_ISSUES.md](doc/KNOWN_ISSUES.md) for current limitations, and [ROADMAP.md](doc/ROADMAP.md) for upcoming features.

## Alignment with Official SDKs

This SDK is designed to follow the patterns established by the official Globus SDKs for [Python](https://github.com/globus/globus-sdk-python) and [JavaScript](https://github.com/globus/globus-sdk-javascript). See [ALIGNMENT.md](ALIGNMENT.md) for details.

## Contributing

Contributions are welcome! Please see the project's GitHub repository for more information.

## License

Apache 2.0

## Resources

- [Globus API Documentation](https://docs.globus.org/api/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
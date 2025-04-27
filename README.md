# Globus Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/scttfrdmn/globus-go-sdk.svg)](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/scttfrdmn/globus-go-sdk)](https://goreportcard.com/report/github.com/scttfrdmn/globus-go-sdk)
[![Build Status](https://github.com/scttfrdmn/globus-go-sdk/workflows/Go/badge.svg)](https://github.com/scttfrdmn/globus-go-sdk/actions)
[![License](https://img.shields.io/github/license/scttfrdmn/globus-go-sdk)](LICENSE)
[![Release](https://img.shields.io/github/v/release/scttfrdmn/globus-go-sdk)](https://github.com/scttfrdmn/globus-go-sdk/releases)
[![Coverage](https://codecov.io/gh/scttfrdmn/globus-go-sdk/branch/main/graph/badge.svg)](https://codecov.io/gh/scttfrdmn/globus-go-sdk)

A Go SDK for interacting with Globus services, providing a simple and idiomatic Go interface to Globus APIs.

## Features

- OAuth2 authentication support
- Token management with automatic refreshing
- Persistent token storage (memory and file-based)
- Groups management
- File transfer with recursive directory and resumable transfer support
- Context-based API with cancellation support
- Structured logging and distributed tracing
- Timers for scheduling tasks
- Integration testing infrastructure
- Minimal dependencies
- Comprehensive error handling
- CLI example application

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

## SDK Organization

This SDK follows the structure and patterns established by the official Globus Python and JavaScript SDKs. It is organized into service-specific packages that each provide clients for interacting with Globus services.

### Core Components

- `pkg/core`: Base client functionality, error handling, logging, etc.
- `pkg/core/authorizers`: Authentication mechanisms
- `pkg/core/transport`: HTTP transport layer
- `pkg/core/config`: Configuration management

### Services

- `pkg/services/auth`: OAuth2 authentication and authorization
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
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

func main() {
    // Create a token storage for persisting tokens
    storage, err := auth.NewFileTokenStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }
    
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

    // Create a new Auth client
    authClient := config.NewAuthClient()
    authClient.SetRedirectURL("http://localhost:8080/callback")
    
    // Create a token manager for automatic refresh
    tokenManager := &auth.TokenManager{
        Storage:          storage,
        RefreshThreshold: 5 * time.Minute,
        RefreshFunc: func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
            return authClient.RefreshToken(ctx, token.RefreshToken)
        },
    }
    
    // Check if we already have tokens
    token, err := tokenManager.GetToken(context.Background(), "default")
    if err == nil && !token.IsExpired() {
        // We have valid tokens, use them
        fmt.Printf("Using existing token (expires in %v)\n", token.Lifetime())
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
        
        // Store the tokens
        err = tokenManager.StoreToken(context.Background(), "default", tokenResponse)
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

    // Create a new Groups client with an access token
    groupsClient := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    
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

    // Create a new Transfer client with an access token
    transferClient := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    
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

    // Create a new Search client with an access token
    searchClient := config.NewSearchClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    
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

    // Create a new Flows client with an access token
    flowsClient := config.NewFlowsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    
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
- [Token Storage Guide](doc/token-storage.md)
- [Recursive Transfers Guide](doc/recursive-transfers.md)
- [Resumable Transfers Guide](doc/resumable-transfers.md)
- [Logging and Tracing Guide](doc/logging-and-tracing.md)
- [Integration Testing Guide](doc/INTEGRATION_TESTING.md)
- [Search Client Guide](doc/search-client.md)
- [Timers Client Guide](doc/timers-client.md)
- [Memory Optimization Guide](doc/memory-optimization.md)
- [Data Schemas](doc/data-schemas.md)
- [Error Handling](doc/error-handling.md)
- [Extending the SDK](doc/extending-the-sdk.md)
- [CLI Example](cmd/globus-cli/README.md)

## Development Status

This SDK is under active development. Current version: **v0.1.0-dev**

| Component | Status | Details |
|-----------|--------|---------|
| Core Infrastructure | ✅ Complete | Base client, transport, authorizers, logging |
| Auth Client | ✅ Complete | OAuth flows, token management, validation utilities |
| Token Storage | ✅ Complete | Interface with memory and file-based implementations |
| Token Manager | ✅ Complete | Automatic token refreshing and management |
| Groups Client | ✅ Complete | Group management, membership operations |
| Transfer Client | ✅ Complete | Basic operations, recursive directory transfers, resumable transfers |
| Search Client | ✅ Complete | Advanced queries, batch operations, pagination |
| Flows Client | ✅ Complete | Flow discovery, execution, management |
| Timers Client | ✅ Complete | Creating and managing scheduled tasks |
| CLI Example | ✅ Complete | Command-line application showcasing SDK features |
| Compute Client | 📅 Planned | Initial structure defined |

See [PROJECT_STATUS.md](doc/PROJECT_STATUS.md) for detailed status and [ROADMAP.md](doc/ROADMAP.md) for upcoming features.

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
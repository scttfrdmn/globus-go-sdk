# Globus Go SDK

A Go SDK for interacting with Globus services, providing a simple and idiomatic Go interface to Globus APIs.

## Features

- OAuth2 authentication support
- Groups management
- Context-based API with cancellation support
- Minimal dependencies
- Comprehensive error handling

## Installation

```bash
go get github.com/yourusername/globus-go-sdk
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

## Quick Start

### Authentication

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/yourusername/globus-go-sdk/pkg"
)

func main() {
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

    // Create a new Auth client
    authClient := config.NewAuthClient()
    authClient.SetRedirectURL("http://localhost:8080/callback")
    
    // Get authorization URL for user authentication
    authURL := authClient.GetAuthorizationURL("my-state")
    fmt.Printf("Visit this URL to log in: %s\n", authURL)
    
    // Start a local server to handle the callback
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        
        // Exchange code for tokens
        tokenResponse, err := authClient.ExchangeAuthorizationCode(context.Background(), code)
        if err != nil {
            log.Fatalf("Failed to exchange code: %v", err)
        }
        
        fmt.Printf("Access Token: %s\n", tokenResponse.AccessToken)
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

    "github.com/yourusername/globus-go-sdk/pkg"
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

    "github.com/yourusername/globus-go-sdk/pkg"
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
    
    // Submit a transfer (if source and destination endpoints are provided)
    sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
    destEndpointID := os.Getenv("DEST_ENDPOINT_ID")
    
    if sourceEndpointID != "" && destEndpointID != "" {
        task, err := transferClient.SubmitTransfer(
            context.Background(),
            sourceEndpointID, "/~/source.txt",
            destEndpointID, "/~/destination.txt",
            "SDK Example Transfer",
            nil,
        )
        if err != nil {
            log.Fatalf("Failed to submit transfer: %v", err)
        }
        
        fmt.Printf("Transfer submitted, task ID: %s\n", task.TaskID)
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

    "github.com/yourusername/globus-go-sdk/pkg"
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

## Documentation

For detailed documentation, see the [GoDoc](https://pkg.go.dev/github.com/yourusername/globus-go-sdk/).

## Development Status

This SDK is under active development. The current status:

- ✅ Core infrastructure
- ✅ Auth client
- ✅ Groups client
- ✅ Transfer client
- ✅ Search client
- 🔄 Flows client (coming soon)

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
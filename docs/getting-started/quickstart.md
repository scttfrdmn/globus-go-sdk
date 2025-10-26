# Quick Start

This guide will help you write your first code using the Globus Go SDK.

## Prerequisites

- [SDK installed](installation.md)
- A Globus account
- An access token (see [Authentication](authentication.md) for details)

## Basic Usage Pattern

All Globus SDK operations follow this pattern:

1. Create an authorizer with your credentials
2. Create a service client
3. Call service methods
4. Handle responses and errors

## Your First Program

### 1. Create a New Project

```bash
mkdir my-globus-app
cd my-globus-app
go mod init my-globus-app
go get github.com/scttfrdmn/globus-go-sdk/v3
```

### 2. Write the Code

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
)

func main() {
    // Get access token from environment
    token := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if token == "" {
        log.Fatal("GLOBUS_ACCESS_TOKEN environment variable not set")
    }

    // Create authorizer
    authorizer := authorizers.NewAccessTokenAuthorizer(token)

    // Create auth client
    client := auth.NewClient(authorizer)

    // Get user info
    ctx := context.Background()
    userInfo, err := client.GetUserInfo(ctx)
    if err != nil {
        log.Fatalf("Error getting user info: %v", err)
    }

    fmt.Printf("Username: %s\n", userInfo.PreferredUsername)
    fmt.Printf("Email: %s\n", userInfo.Email)
}
```

### 3. Get an Access Token

For testing, get a token from the [Globus developers console](https://developers.globus.org) or use the [Globus CLI](https://github.com/scttfrdmn/globus-go-cli):

```bash
# Using Globus CLI
globus login
```

### 4. Run Your Program

```bash
export GLOBUS_ACCESS_TOKEN="your-access-token"
go run main.go
```

## Working with Transfer Service

### List Endpoints

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
    token := os.Getenv("GLOBUS_ACCESS_TOKEN")
    authorizer := authorizers.NewAccessTokenAuthorizer(token)
    client := transfer.NewClient(authorizer)

    ctx := context.Background()

    // List endpoints
    endpoints, err := client.EndpointList(ctx, &transfer.EndpointListOptions{
        Limit: 10,
    })
    if err != nil {
        log.Fatalf("Error listing endpoints: %v", err)
    }

    fmt.Printf("Found %d endpoints:\n", len(endpoints.Data))
    for _, ep := range endpoints.Data {
        fmt.Printf("- %s (%s)\n", ep.DisplayName, ep.ID)
    }
}
```

### Initiate a Transfer

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
    token := os.Getenv("GLOBUS_ACCESS_TOKEN")
    authorizer := authorizers.NewAccessTokenAuthorizer(token)
    client := transfer.NewClient(authorizer)

    ctx := context.Background()

    // Create transfer submission
    submission := &transfer.TransferSubmission{
        SourceEndpoint: "SOURCE_ENDPOINT_ID",
        DestinationEndpoint: "DEST_ENDPOINT_ID",
        Label: "My first transfer",
        Data: []transfer.TransferItem{
            {
                SourcePath: "/~/source/file.txt",
                DestinationPath: "/~/dest/file.txt",
            },
        },
    }

    // Submit transfer
    task, err := client.SubmitTransfer(ctx, submission)
    if err != nil {
        log.Fatalf("Error submitting transfer: %v", err)
    }

    fmt.Printf("Transfer submitted! Task ID: %s\n", task.TaskID)
}
```

## Error Handling

The SDK uses Go's standard error handling:

```go
result, err := client.SomeMethod(ctx, options)
if err != nil {
    // Check for specific error types
    var apiErr *errors.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
        return
    }

    // Handle other errors
    log.Fatal(err)
}

// Use result
fmt.Println(result)
```

## Context Usage

Always pass a context for cancellation and timeouts:

```go
import (
    "context"
    "time"
)

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.SomeMethod(ctx, options)
```

## Common Patterns

### Pagination

```go
options := &transfer.TaskListOptions{
    Limit: 100,
}

for {
    tasks, err := client.TaskList(ctx, options)
    if err != nil {
        log.Fatal(err)
    }

    for _, task := range tasks.Data {
        fmt.Printf("Task: %s\n", task.TaskID)
    }

    if !tasks.HasNextPage {
        break
    }

    options.Offset += options.Limit
}
```

### Retry Logic

```go
import "time"

maxRetries := 3
for i := 0; i < maxRetries; i++ {
    result, err := client.SomeMethod(ctx, options)
    if err == nil {
        // Success
        break
    }

    if i < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(i+1))
        continue
    }

    // Final retry failed
    log.Fatal(err)
}
```

## Next Steps

- Learn about [Authentication](authentication.md) in detail
- Explore [Common Patterns](../guides/common-patterns.md)
- Check [API Reference](../api/index.md) for all services
- Review [Examples](../examples/transfer.md) for more code

## See Also

- [Authentication Guide](authentication.md)
- [Client Configuration](configuration.md)
- [Error Handling](../guides/error-handling.md)
- [API Reference](../api/index.md)

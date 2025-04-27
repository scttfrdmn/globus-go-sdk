<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK User Guide

This guide provides comprehensive information about using the Globus Go SDK for interacting with Globus services.

## Table of Contents

- [Installation](#installation)
- [Authentication](#authentication)
- [Token Management](#token-management)
- [Transfer Service](#transfer-service)
  - [Basic Transfers](#basic-transfers)
  - [Recursive Transfers](#recursive-directory-transfer)
  - [Resumable Transfers](#resumable-transfers)
  - [Monitoring Transfers](#monitoring-transfer-status)
- [Groups Service](#groups-service)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Installation

Install the Globus Go SDK using `go get`:

```bash
go get github.com/scttfrdmn/globus-go-sdk
```

Include it in your Go code:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
)
```

## Authentication

### Creating an Auth Client

```go
package main

import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

func main() {
    ctx := context.Background()
    
    // Create an auth client with your Globus application credentials
    authClient := auth.NewClient(auth.ClientConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret", // Optional for native apps
        RedirectURL:  "https://your-app.example/callback",
    })
    
    // Use the auth client...
}
```

### OAuth2 Authorization Flow

To authenticate a user:

1. Generate an authorization URL:

```go
authURL := authClient.GetAuthorizationURL([]string{
    auth.ScopeTransferAll,
    auth.ScopeOpenID,
    auth.ScopeEmail,
})

// Direct user to authURL in a browser
fmt.Println("Visit this URL to authenticate:", authURL)
```

2. Exchange the authorization code for tokens:

```go
// After the user authorizes and is redirected to your app
code := "authorization-code-from-redirect"

tokens, err := authClient.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    fmt.Printf("Error exchanging code: %v\n", err)
    return
}

// Use tokens.AccessToken for API calls
fmt.Printf("Access Token: %s\n", tokens.AccessToken)
```

## Token Management

### Using Token Storage

The SDK provides flexible token storage mechanisms:

```go
// Create file-based token storage
storage, err := auth.NewFileTokenStorage("/path/to/tokens")
if err != nil {
    fmt.Printf("Error creating storage: %v\n", err)
    return
}

// Store tokens
err = storage.StoreToken(ctx, "user1", tokens)
if err != nil {
    fmt.Printf("Error storing token: %v\n", err)
    return
}

// Retrieve tokens
savedTokens, err := storage.GetToken(ctx, "user1")
if err != nil {
    fmt.Printf("Error retrieving token: %v\n", err)
    return
}
```

### Using Token Manager for Auto-Refresh

```go
// Create a token manager with refresh capabilities
refreshFunc := func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
    return authClient.RefreshToken(ctx, token.RefreshToken)
}

manager := &auth.TokenManager{
    Storage:          storage,
    RefreshThreshold: 5 * time.Minute,
    RefreshFunc:      refreshFunc,
}

// Get tokens with automatic refresh if needed
freshTokens, err := manager.GetToken(ctx, "user1")
if err != nil {
    fmt.Printf("Error getting token: %v\n", err)
    return
}

// Use freshTokens.AccessToken for API calls
```

## Transfer Service

### Creating a Transfer Client

```go
// Create a transfer client using the auth client
transferClient := transfer.NewClient(authClient)
```

### List Endpoint Contents

```go
items, err := transferClient.ListEndpointContents(ctx, "endpoint-id", "/path", nil)
if err != nil {
    fmt.Printf("Error listing contents: %v\n", err)
    return
}

for _, item := range items {
    fmt.Printf("%s - %s (%d bytes)\n", item.Type, item.Name, item.Size)
}
```

### Submitting a Transfer

```go
// Basic file transfer
transfer := &transfer.TransferData{
    Label: "My First Transfer",
    Items: []transfer.TransferItem{
        {
            Source:      "/source/path/file.txt",
            Destination: "/destination/path/file.txt",
        },
    },
}

task, err := transferClient.SubmitTransfer(ctx, "source-endpoint-id", "destination-endpoint-id", transfer)
if err != nil {
    fmt.Printf("Error submitting transfer: %v\n", err)
    return
}

fmt.Printf("Transfer submitted with task ID: %s\n", task.TaskID)
```

### Recursive Directory Transfer

```go
result, err := transferClient.SubmitRecursiveTransfer(
    ctx,
    "source-endpoint-id", "/source/directory",
    "destination-endpoint-id", "/destination/directory",
    &transfer.RecursiveTransferOptions{
        Label:     "Directory Transfer",
        SyncLevel: transfer.SyncChecksum,
    },
)

if err != nil {
    fmt.Printf("Error submitting recursive transfer: %v\n", err)
    return
}

fmt.Printf("Transfer submitted with %d tasks\n", len(result.TaskIDs))
fmt.Printf("Transferred %d items in total\n", result.ItemsTransferred)
```

### Resumable Transfers

Create and start a resumable transfer:

```go
// Set up options
options := transfer.DefaultResumableTransferOptions()
options.BatchSize = 50
options.ProgressCallback = func(state *transfer.CheckpointState) {
    fmt.Printf("Progress: %d/%d files completed\n", 
        state.Stats.CompletedItems, state.Stats.TotalItems)
}

// Start a new resumable transfer
checkpointID, err := transferClient.SubmitResumableTransfer(
    ctx,
    "source-endpoint-id",
    "/source/path",
    "destination-endpoint-id",
    "/destination/path",
    options,
)

if err != nil {
    fmt.Printf("Error creating resumable transfer: %v\n", err)
    return
}

fmt.Printf("Transfer created with checkpoint ID: %s\n", checkpointID)
```

Resuming a previously started transfer:

```go
// Resume a transfer
result, err := transferClient.ResumeResumableTransfer(ctx, checkpointID, options)
if err != nil {
    fmt.Printf("Error resuming transfer: %v\n", err)
    return
}

fmt.Printf("Transfer completed: %d/%d files\n", 
    result.CompletedItems, result.CompletedItems + result.FailedItems)
```

For more details about resumable transfers, see the [Resumable Transfers Guide](resumable-transfers.md).

### Monitoring Transfer Status

```go
status, err := transferClient.GetTaskStatus(ctx, "task-id")
if err != nil {
    fmt.Printf("Error getting task status: %v\n", err)
    return
}

fmt.Printf("Task status: %s\n", status.Status)
fmt.Printf("Files: %d/%d\n", status.FilesTransferred, status.FilesTotal)
```

## Groups Service

### Creating a Groups Client

```go
// Create a groups client using the auth client
groupsClient := groups.NewClient(authClient)
```

### Listing Groups

```go
myGroups, err := groupsClient.ListGroups(ctx, nil)
if err != nil {
    fmt.Printf("Error listing groups: %v\n", err)
    return
}

for _, group := range myGroups {
    fmt.Printf("Group: %s (%s)\n", group.Name, group.ID)
}
```

### Getting Group Members

```go
members, err := groupsClient.ListMembers(ctx, "group-id", nil)
if err != nil {
    fmt.Printf("Error listing members: %v\n", err)
    return
}

for _, member := range members {
    fmt.Printf("Member: %s (%s)\n", member.Username, member.ID)
}
```

## Error Handling

The SDK provides typed errors to help with specific error conditions:

```go
_, err := transferClient.ListEndpointContents(ctx, "endpoint-id", "/path", nil)
if err != nil {
    switch {
    case transfer.IsAuthenticationError(err):
        // Handle authentication errors
        fmt.Println("Authentication failed, please log in again")
    
    case transfer.IsEndpointNotFoundError(err):
        // Handle endpoint not found
        fmt.Println("Endpoint not found, check the endpoint ID")
    
    case transfer.IsPermissionError(err):
        // Handle permission issues
        fmt.Println("Permission denied, you don't have access")
        
    case transfer.IsRateLimitError(err):
        // Handle rate limiting
        fmt.Println("Rate limit exceeded, please try again later")
        
    default:
        // Handle other errors
        fmt.Printf("Unexpected error: %v\n", err)
    }
    return
}
```

## Examples

### Complete File Transfer Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    ctx := context.Background()
    
    // Create auth client
    authClient := auth.NewClient(auth.ClientConfig{
        ClientID: "your-client-id",
        RedirectURL: "https://your-app.example/callback",
    })
    
    // Set up token storage and manager
    storage, _ := auth.NewFileTokenStorage("~/.globus-tokens")
    manager := &auth.TokenManager{
        Storage: storage,
        RefreshThreshold: 5 * time.Minute,
        RefreshFunc: func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
            return authClient.RefreshToken(ctx, token.RefreshToken)
        },
    }
    
    // Check if we have a token
    _, err := manager.GetToken(ctx, "default")
    if err != nil {
        // No token, prompt for login
        authURL := authClient.GetAuthorizationURL([]string{
            auth.ScopeTransferAll,
        })
        
        fmt.Printf("Please visit this URL to log in:\n%s\n", authURL)
        fmt.Print("Enter the authorization code from the redirect: ")
        
        var code string
        fmt.Scanln(&code)
        
        tokens, err := authClient.ExchangeAuthorizationCode(ctx, code)
        if err != nil {
            fmt.Printf("Error exchanging code: %v\n", err)
            return
        }
        
        // Store the new tokens
        err = manager.StoreToken(ctx, "default", tokens)
        if err != nil {
            fmt.Printf("Error storing token: %v\n", err)
            return
        }
    }
    
    // Create transfer client
    transferClient := transfer.NewClient(authClient)
    
    // Submit a transfer
    task, err := transferClient.SubmitTransfer(
        ctx,
        "source-endpoint-id",
        "destination-endpoint-id",
        &transfer.TransferData{
            Label: "SDK Example Transfer",
            Items: []transfer.TransferItem{
                {
                    Source:      "/source/path/file.txt",
                    Destination: "/destination/path/file.txt",
                },
            },
        },
    )
    
    if err != nil {
        fmt.Printf("Error submitting transfer: %v\n", err)
        return
    }
    
    fmt.Printf("Transfer submitted with task ID: %s\n", task.TaskID)
    
    // Monitor the task until completion
    for {
        status, err := transferClient.GetTaskStatus(ctx, task.TaskID)
        if err != nil {
            fmt.Printf("Error checking status: %v\n", err)
            break
        }
        
        fmt.Printf("Task status: %s - Files: %d/%d\n", 
            status.Status, 
            status.FilesTransferred, 
            status.FilesTotal)
        
        if status.Status == "SUCCEEDED" || status.Status == "FAILED" {
            break
        }
        
        time.Sleep(5 * time.Second)
    }
}
```

For more examples, see the `examples` directory in the SDK repository.
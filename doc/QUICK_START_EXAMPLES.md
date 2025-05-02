# Globus Go SDK Quick Start Examples

This document provides updated examples using the v0.8.0+ client initialization patterns and error handling.

## Authentication with Token Management

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
    // Create a token storage for persisting tokens
    storage, err := tokens.NewFileStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }
    
    // Create a new Auth client with the options pattern
    authClient, err := auth.NewClient(
        auth.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
        auth.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
        auth.WithRedirectURL("http://localhost:8080/callback"),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    
    // Create a token manager for automatic refresh
    tokenManager := tokens.NewManager(storage, authClient)
    
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

## Transfer Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    // Create a Transfer client with the new options pattern
    transferClient, err := transfer.NewClient(
        transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
    )
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    ctx := context.Background()
    
    // List endpoints with retry for rate limiting
    var endpoints *transfer.EndpointList
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var listErr error
            endpoints, listErr = transferClient.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
                FilterScope: "my-endpoints",
                Limit:       10,
            })
            return listErr
        },
        ratelimit.DefaultBackoff(),
        transfer.IsRetryableTransferError,
    )
    if err != nil {
        log.Fatalf("Failed to list endpoints: %v", err)
    }
    
    fmt.Printf("Found %d endpoints:\n", len(endpoints.Data))
    for i, endpoint := range endpoints.Data {
        fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
    }
    
    // Submit a file transfer (if source and destination endpoints are provided)
    sourceEndpointID := os.Getenv("SOURCE_ENDPOINT_ID")
    destEndpointID := os.Getenv("DEST_ENDPOINT_ID")
    
    if sourceEndpointID != "" && destEndpointID != "" {
        // Create a transfer request
        transferRequest := &transfer.TransferTaskRequest{
            DataType:              "transfer",
            Label:                 "SDK Example Transfer",
            SourceEndpointID:      sourceEndpointID,
            DestinationEndpointID: destEndpointID,
            SyncLevel:             2, // Checksum verification
            VerifyChecksum:        true,
            Encrypt:               true,
            Items: []transfer.TransferItem{
                {
                    DataType:        "transfer_item",
                    SourcePath:      "/~/source.txt",
                    DestinationPath: "/~/destination.txt",
                },
            },
        }
        
        // Submit transfer with retry
        var taskResponse *transfer.TaskResponse
        err = ratelimit.RetryWithBackoff(
            ctx,
            func(ctx context.Context) error {
                var taskErr error
                taskResponse, taskErr = transferClient.CreateTransferTask(ctx, transferRequest)
                return taskErr
            },
            ratelimit.DefaultBackoff(),
            transfer.IsRetryableTransferError,
        )
        if err != nil {
            log.Fatalf("Failed to submit transfer: %v", err)
        }
        
        fmt.Printf("Transfer submitted, task ID: %s\n", taskResponse.TaskID)
        
        // Monitor the task
        taskID := taskResponse.TaskID
        waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()
        
        err = ratelimit.RetryWithBackoff(
            waitCtx,
            func(ctx context.Context) error {
                task, err := transferClient.GetTask(ctx, taskID)
                if err != nil {
                    return err
                }
                
                fmt.Printf("Task status: %s\n", task.Status)
                
                if task.Status == "SUCCEEDED" || task.Status == "FAILED" {
                    fmt.Printf("Task completed with status: %s\n", task.Status)
                    return nil
                }
                
                return fmt.Errorf("task still in progress")
            },
            &ratelimit.ExponentialBackoff{
                InitialDelay: 2 * time.Second,
                MaxDelay:     30 * time.Second,
                Factor:       1.5,
                Jitter:       true,
                MaxAttempt:   20,
            },
            func(err error) bool {
                return err != nil && (err.Error() == "task still in progress" || 
                    transfer.IsRetryableTransferError(err))
            },
        )
        
        if err != nil {
            if waitCtx.Err() == context.DeadlineExceeded {
                fmt.Println("Timed out waiting for task completion")
            } else {
                fmt.Printf("Error monitoring task: %v\n", err)
            }
        }
    }
}
```

## Groups Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
)

func main() {
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    // Create a Groups client with the new options pattern
    groupsClient, err := groups.NewClient(
        groups.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
    )
    if err != nil {
        log.Fatalf("Failed to create groups client: %v", err)
    }
    
    ctx := context.Background()
    
    // List groups with retry for rate limiting
    var groupList *groups.GroupList
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var listErr error
            groupList, listErr = groupsClient.ListGroups(ctx, nil)
            return listErr
        },
        ratelimit.DefaultBackoff(),
        func(err error) bool {
            // Retry on any error for this example
            return err != nil
        },
    )
    if err != nil {
        log.Fatalf("Failed to list groups: %v", err)
    }
    
    fmt.Printf("You are a member of %d groups:\n", len(groupList.Groups))
    for _, group := range groupList.Groups {
        fmt.Printf("- %s (%s)\n", group.Name, group.ID)
    }
}
```

## Search Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

func main() {
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    // Create a Search client with the new options pattern
    searchClient, err := search.NewClient(
        search.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
    )
    if err != nil {
        log.Fatalf("Failed to create search client: %v", err)
    }
    
    ctx := context.Background()
    
    // List indexes with retry for rate limiting
    var indexes *search.IndexList
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var listErr error
            indexes, listErr = searchClient.ListIndexes(ctx)
            return listErr
        },
        ratelimit.DefaultBackoff(),
        func(err error) bool {
            // Retry on any error for this example
            return err != nil
        },
    )
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
        // Create search request
        searchReq := &search.SearchRequest{
            Q:       "example",
            Limit:   10,
            Offset:  0,
            Filters: []search.Filter{},
        }
        
        // Perform search with retry
        var results *search.SearchResults
        err = ratelimit.RetryWithBackoff(
            ctx,
            func(ctx context.Context) error {
                var searchErr error
                results, searchErr = searchClient.Search(ctx, indexID, searchReq)
                return searchErr
            },
            ratelimit.DefaultBackoff(),
            func(err error) bool {
                // Retry on any error for this example
                return err != nil
            },
        )
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

## Flows Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

func main() {
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    
    // Create a Flows client with the new options pattern
    flowsClient, err := flows.NewClient(
        flows.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
    )
    if err != nil {
        log.Fatalf("Failed to create flows client: %v", err)
    }
    
    ctx := context.Background()
    
    // List flows with retry for rate limiting
    var flowsList *flows.ListFlowsResponse
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var listErr error
            flowsList, listErr = flowsClient.ListFlows(ctx, nil)
            return listErr
        },
        ratelimit.DefaultBackoff(),
        func(err error) bool {
            // Retry on any error for this example
            return err != nil
        },
    )
    if err != nil {
        log.Fatalf("Failed to list flows: %v", err)
    }
    
    fmt.Printf("Found %d flows:\n", len(flowsList.Flows))
    for i, flow := range flowsList.Flows {
        fmt.Printf("%d. %s (%s)\n", i+1, flow.Title, flow.ID)
    }
    
    // Run a flow (if flow ID is provided)
    flowID := os.Getenv("GLOBUS_FLOW_ID")
    if flowID != "" {
        // Create run request
        runReq := &flows.RunFlowRequest{
            FlowID: flowID,
            Label:  fmt.Sprintf("Example Run %s", time.Now().Format("20060102")),
            Input: map[string]interface{}{
                "message": "Hello from Globus Go SDK!",
            },
        }
        
        // Run flow with retry
        var run *flows.FlowRun
        err = ratelimit.RetryWithBackoff(
            ctx,
            func(ctx context.Context) error {
                var runErr error
                run, runErr = flowsClient.RunFlow(ctx, runReq)
                return runErr
            },
            ratelimit.DefaultBackoff(),
            func(err error) bool {
                // Retry on any error for this example
                return err != nil
            },
        )
        if err != nil {
            log.Fatalf("Failed to run flow: %v", err)
        }
        
        fmt.Printf("Flow run started: %s (Status: %s)\n", run.RunID, run.Status)
    }
}
```

## Using Multiple Services Together

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create an Auth client
    authClient, err := auth.NewClient(
        auth.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
        auth.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    
    ctx := context.Background()
    
    // Get a token for the transfer service
    var tokenResp *auth.TokenResponse
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var tokenErr error
            tokenResp, tokenErr = authClient.GetClientCredentialsToken(
                ctx, 
                "urn:globus:auth:scope:transfer.api.globus.org:all",
            )
            return tokenErr
        },
        ratelimit.DefaultBackoff(),
        func(err error) bool {
            return err != nil
        },
    )
    if err != nil {
        log.Fatalf("Failed to get token: %v", err)
    }
    
    // Create a Transfer client with the token
    transferClient, err := transfer.NewClient(
        transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(tokenResp.AccessToken)),
    )
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // List endpoints
    var endpoints *transfer.EndpointList
    err = ratelimit.RetryWithBackoff(
        ctx,
        func(ctx context.Context) error {
            var listErr error
            endpoints, listErr = transferClient.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
                FilterScope: "my-endpoints",
                Limit:       10,
            })
            return listErr
        },
        ratelimit.DefaultBackoff(),
        transfer.IsRetryableTransferError,
    )
    if err != nil {
        log.Fatalf("Failed to list endpoints: %v", err)
    }
    
    fmt.Printf("Found %d endpoints:\n", len(endpoints.Data))
    for i, endpoint := range endpoints.Data {
        fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
    }
}
```
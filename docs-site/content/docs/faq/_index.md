---
title: "Frequently Asked Questions"
weight: 80
bookCollapseSection: true
---

# Globus Go SDK: Frequently Asked Questions

This document provides answers to commonly asked questions about the Globus Go SDK. If you don't find the information you're looking for, please consider checking the [documentation](../reference/) or [filing an issue](https://github.com/scttfrdmn/globus-go-sdk/issues/new) on our GitHub repository.

## Installation and Setup

### How do I install the Globus Go SDK?

Install the SDK using Go modules:

```bash
# In your Go module
go get github.com/scttfrdmn/globus-go-sdk
```

Make sure you have Go 1.19 or later installed. Add the import to your Go files:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg"
```

### What configuration is required to use the SDK?

The SDK supports configuration through environment variables, which is the recommended approach:

```bash
# Required for authentication
export GLOBUS_CLIENT_ID=your_client_id
export GLOBUS_CLIENT_SECRET=your_client_secret

# Optional settings
export GLOBUS_ENVIRONMENT=production  # default is production, can be sandbox for testing
export GLOBUS_HTTP_TIMEOUT=30         # request timeout in seconds
```

You can also configure the SDK programmatically:

```go
config := core.NewConfig().
    WithClientID("your_client_id").
    WithClientSecret("your_client_secret")

client := pkg.NewSDKClient(pkg.WithConfig(config))
```

### How do I set up logging with the SDK?

The SDK provides integrated logging that you can configure:

```go
// Import the logging package
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"

// Configure logging when creating a client
client := pkg.NewSDKClient(
    pkg.WithLogging(logging.NewLogger(
        logging.WithLevel(logging.LevelInfo),
        logging.WithOutput(os.Stdout),
    )),
)
```

You can set different log levels: `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`, or `LevelNone` to disable logging.

### Can I use the SDK with a pre-existing Go application?

Yes, the SDK is designed to integrate seamlessly with existing Go applications. It provides:

- Context-aware operations for proper cancellation support
- Configurable connection pooling for resource management
- Customizable HTTP transport that can be shared with other components
- Flexible logging that can be integrated with your application's logging framework

## Authentication

### What authentication methods does the SDK support?

The SDK supports multiple authentication methods:

1. **Client Credentials Flow**: For server-to-server applications
2. **Authorization Code Flow**: For web applications that interact with user data
3. **Refresh Token Flow**: For long-running applications
4. **Static Token Authentication**: For simple scripts or testing

Example of client credentials flow:

```go
client := pkg.NewSDKClient()
authClient := client.Auth()

// Get client credentials tokens
tokenResponse, err := authClient.ClientCredentials(
    context.Background(),
    []string{"openid", "profile", "email", "urn:globus:auth:scope:transfer.api.globus.org:all"},
)
```

### How do I handle token refresh in my application?

The SDK provides automatic token refresh capabilities through the tokens package:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"

// Create a token manager
tokenManager := tokens.NewManager()

// Add a refreshable token from a previous authentication flow
// This could be from AuthCodeFlow or ClientCredentials
tokenManager.AddRefreshableToken("transfer.api.globus.org", tokenResponse)

// Create a client using this token manager
client := pkg.NewSDKClient(pkg.WithTokenManager(tokenManager))

// The client will automatically refresh tokens when needed
transferClient := client.Transfer()
```

### How do I implement Multi-Factor Authentication (MFA) with the SDK?

MFA is supported through the Auth client's MFA-specific methods:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"

// Initialize the client
client := pkg.NewSDKClient()
authClient := client.Auth()

// Create a URL for an authorization flow that requires MFA
authURL := authClient.GetAuthorizationCodeURL(
    "https://your-app.example.com/callback",
    []string{"openid", "profile", "email", "urn:globus:auth:scope:transfer.api.globus.org:all"},
    auth.WithMFA(),  // This enables MFA requirement
)

// Direct the user to this URL
// After they authenticate, exchange the code for a token
code := "code_from_callback"
tokenResponse, err := authClient.ExchangeAuthorizationCode(
    context.Background(),
    "https://your-app.example.com/callback",
    code,
)
```

### Where should I store access tokens securely?

For production applications, consider these token storage strategies:

1. **Memory-only**: For short-lived applications
2. **Encrypted file storage**: For desktop applications
3. **Secure database**: For web applications
4. **Secret management services**: For cloud deployments (AWS Secrets Manager, HashiCorp Vault, etc.)

The SDK provides a simple file-based implementation for development:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"

// Create a file-based token storage
tokenStorage, err := tokens.NewFileStorage("/path/to/secure/tokens.json")
if err != nil {
    log.Fatalf("Failed to create token storage: %v", err)
}

// Create a token manager with this storage
tokenManager := tokens.NewManager(tokens.WithStorage(tokenStorage))

// Create a client using this token manager
client := pkg.NewSDKClient(pkg.WithTokenManager(tokenManager))
```

## Transfer Operations

### How do I perform a basic file transfer?

A basic transfer operation involves creating a transfer task:

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// Initialize the client
client := pkg.NewSDKClient()
transferClient := client.Transfer()

// Create a transfer task
transferTask, err := transferClient.SubmitTransfer(
    context.Background(),
    transfer.TransferOptions{
        SourceEndpointID: "source_endpoint_id",
        DestinationEndpointID: "destination_endpoint_id",
        Items: []transfer.TransferItem{
            {
                SourcePath: "/path/on/source/file.txt",
                DestinationPath: "/path/on/destination/file.txt",
            },
        },
        Label: "My file transfer",
    },
)

if err != nil {
    log.Fatalf("Transfer submission failed: %v", err)
}

// Get the task ID for monitoring
taskID := transferTask.TaskID
```

### How do I monitor transfer progress?

The SDK provides methods for tracking transfer progress:

```go
// After submitting a transfer, get the task ID
taskID := transferTask.TaskID

// Check task status
status, err := transferClient.GetTaskStatus(context.Background(), taskID)
if err != nil {
    log.Fatalf("Failed to get task status: %v", err)
}

fmt.Printf("Task status: %s\n", status.Status)

// For detailed progress monitoring, use the metrics package
import "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"

// Create a progress tracker
progress := metrics.NewTransferProgress(taskID)

// Start tracking (this runs in a goroutine)
progress.Start(context.Background(), transferClient)

// Register a callback for updates
progress.OnUpdate(func(stats metrics.TransferStats) {
    fmt.Printf("Transferred: %d/%d files, %s/%s\n",
        stats.FilesTransferred, stats.FilesTotal,
        metrics.FormatBytes(stats.BytesTransferred), metrics.FormatBytes(stats.BytesTotal),
    )
})

// When done monitoring
progress.Stop()
```

### How do I implement recursive directory transfers?

Recursive transfers can be done using the recursive transfer helper:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"

// Create a recursive transfer job
recursiveJob := transfer.NewRecursiveTransfer(
    transferClient,
    "source_endpoint_id",
    "destination_endpoint_id",
    "/source/directory",
    "/destination/directory",
)

// Optionally configure the job
recursiveJob.WithConcurrentListLimit(10)  // Number of concurrent directory listings
recursiveJob.WithTransferConcurrency(5)   // Number of concurrent transfer tasks

// Start the transfer
taskID, err := recursiveJob.Start(context.Background())
if err != nil {
    log.Fatalf("Failed to start recursive transfer: %v", err)
}

// Monitor the job
for status := range recursiveJob.StatusChannel() {
    fmt.Printf("Directories processed: %d, Files queued: %d\n", 
        status.DirectoriesProcessed, status.FilesQueued)
}
```

### How do I implement resumable transfers for large files?

The SDK supports resumable transfers for handling large files or unreliable networks:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"

// Create a resumable transfer
resumable := transfer.NewResumableTransfer(
    transferClient,
    transfer.ResumableOptions{
        SourceEndpointID: "source_endpoint_id",
        DestinationEndpointID: "destination_endpoint_id",
        Label: "Large file transfer",
        Items: []transfer.TransferItem{
            {
                SourcePath: "/path/on/source/largefile.iso",
                DestinationPath: "/path/on/destination/largefile.iso",
            },
        },
        // Checkpoint after every 10GB
        CheckpointInterval: 10 * 1024 * 1024 * 1024,
    },
)

// Start the transfer
taskID, err := resumable.Start(context.Background())
if err != nil {
    log.Fatalf("Failed to start resumable transfer: %v", err)
}

// Monitor the transfer
for event := range resumable.StatusChannel() {
    switch event.Type {
    case transfer.EventCheckpoint:
        fmt.Printf("Transfer checkpoint: %d bytes transferred\n", event.BytesTransferred)
    case transfer.EventResume:
        fmt.Printf("Transfer resumed from checkpoint\n")
    case transfer.EventComplete:
        fmt.Printf("Transfer completed successfully\n")
    case transfer.EventError:
        fmt.Printf("Transfer error: %v\n", event.Error)
    }
}
```

## Search Usage

### How do I perform a basic search query?

Basic search operations can be performed with the Search client:

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
)

// Initialize the client
client := pkg.NewSDKClient()
searchClient := client.Search()

// Perform a basic search
results, err := searchClient.Search(
    context.Background(),
    "my_index_id",
    search.NewQuery("neutron scattering"),
    search.WithLimit(100),
)

if err != nil {
    log.Fatalf("Search failed: %v", err)
}

// Process results
fmt.Printf("Found %d results\n", results.Count)
for _, entry := range results.Entries {
    fmt.Printf("Title: %s\n", entry.Content["title"])
    fmt.Printf("Subject: %s\n", entry.Content["subject"])
}
```

### How do I create advanced search queries?

For more complex searches, you can use the advanced query builder:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/services/search"

// Create an advanced query
query := search.NewAdvancedQuery().
    Field("type", search.Equals, "dataset").
    And().
    Group(
        search.NewAdvancedQuery().
            Field("author", search.Contains, "Smith").
            Or().
            Field("author", search.Contains, "Johnson"),
    ).
    And().
    Field("year", search.GreaterThan, 2020).
    And().
    Field("keywords", search.In, []string{"climate", "weather", "temperature"})

// Use the query
results, err := searchClient.Search(
    context.Background(),
    "my_index_id",
    query,
    search.WithLimit(100),
)
```

### How do I handle search pagination?

For large result sets, you can use either manual pagination or the iterator pattern:

```go
// Method 1: Manual pagination
var allResults []search.Entry
pageSize := 100
offset := 0

for {
    results, err := searchClient.Search(
        context.Background(),
        "my_index_id",
        search.NewQuery("climate data"),
        search.WithLimit(pageSize),
        search.WithOffset(offset),
    )
    
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }
    
    allResults = append(allResults, results.Entries...)
    
    if len(results.Entries) < pageSize {
        break // No more results
    }
    
    offset += pageSize
}

// Method 2: Using the iterator
iterator := searchClient.Iterator(
    context.Background(),
    "my_index_id",
    search.NewQuery("climate data"),
    search.WithBatchSize(100),
)

for iterator.Next() {
    entry := iterator.Entry()
    fmt.Printf("Found: %s\n", entry.Content["title"])
}

if err := iterator.Error(); err != nil {
    log.Fatalf("Search iteration failed: %v", err)
}
```

## Flows Implementation

### How do I create and manage flows?

Creating and managing flows involves the Flows client:

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
)

// Initialize the client
client := pkg.NewSDKClient()
flowsClient := client.Flows()

// Create a new flow
flow, err := flowsClient.CreateFlow(
    context.Background(),
    flows.CreateFlowRequest{
        Title: "Data Processing Flow",
        Definition: flows.FlowDefinition{
            // Flow definition goes here
            // This typically involves a JSON structure defining the flow's steps
        },
    },
)

if err != nil {
    log.Fatalf("Flow creation failed: %v", err)
}

// Get a flow by ID
flow, err = flowsClient.GetFlow(context.Background(), flow.ID)
if err != nil {
    log.Fatalf("Failed to get flow: %v", err)
}

// List all flows
flowsList, err := flowsClient.ListFlows(context.Background())
if err != nil {
    log.Fatalf("Failed to list flows: %v", err)
}

// Update a flow
updated, err := flowsClient.UpdateFlow(
    context.Background(),
    flow.ID,
    flows.UpdateFlowRequest{
        Title: "Updated Data Processing Flow",
    },
)
```

### How do I execute a flow?

Running a flow requires starting a flow run:

```go
// Start a flow run
run, err := flowsClient.StartRun(
    context.Background(),
    flows.StartRunRequest{
        FlowID: "flow_id",
        Label: "Processing batch #123",
        Input: map[string]interface{}{
            "source_endpoint": "endpoint_id",
            "source_path": "/path/to/data",
            "destination_endpoint": "destination_id",
            "destination_path": "/path/to/output",
            "parameters": map[string]string{
                "resolution": "high",
                "format": "netcdf",
            },
        },
    },
)

if err != nil {
    log.Fatalf("Flow run failed to start: %v", err)
}

// Get the run ID for monitoring
runID := run.RunID
```

### How do I monitor flow execution status?

Monitoring a flow run involves checking its status:

```go
// Get a run by ID
run, err := flowsClient.GetRun(context.Background(), runID)
if err != nil {
    log.Fatalf("Failed to get run: %v", err)
}

fmt.Printf("Run status: %s\n", run.Status)

// For continuous monitoring, poll at intervals
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
defer cancel()

for {
    select {
    case <-ticker.C:
        run, err := flowsClient.GetRun(ctx, runID)
        if err != nil {
            log.Printf("Error checking run status: %v", err)
            continue
        }
        
        fmt.Printf("Run status: %s\n", run.Status)
        
        // Check if the run has completed or failed
        if run.Status == "SUCCEEDED" || run.Status == "FAILED" {
            // Get detailed results if needed
            details, err := flowsClient.GetRunDetails(ctx, runID)
            if err == nil {
                fmt.Printf("Run details: %+v\n", details)
            }
            return
        }
        
    case <-ctx.Done():
        log.Printf("Monitoring timed out: %v", ctx.Err())
        return
    }
}
```

## Compute Integration

### How do I create and manage compute environments?

Working with compute environments involves the Compute client:

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
)

// Initialize the client
client := pkg.NewSDKClient()
computeClient := client.Compute()

// Create a compute environment
env, err := computeClient.CreateEnvironment(
    context.Background(),
    compute.CreateEnvironmentRequest{
        DisplayName: "My Analysis Environment",
        Environment: compute.Environment{
            GlobusConnect: &compute.GlobusConnectEnvironment{
                EndpointID: "endpoint_id",
                Path: "/path/to/compute/dir",
            },
        },
    },
)

if err != nil {
    log.Fatalf("Failed to create environment: %v", err)
}

// List environments
envs, err := computeClient.ListEnvironments(context.Background())
if err != nil {
    log.Fatalf("Failed to list environments: %v", err)
}

// Get environment by ID
env, err = computeClient.GetEnvironment(context.Background(), env.ID)
if err != nil {
    log.Fatalf("Failed to get environment: %v", err)
}
```

### How do I submit compute tasks?

Submitting compute tasks involves creating batches:

```go
// Create a compute batch
batch, err := computeClient.CreateBatch(
    context.Background(),
    compute.CreateBatchRequest{
        EnvironmentID: "environment_id",
        Label: "Data analysis batch",
        Tasks: []compute.Task{
            {
                Label: "Analysis task 1",
                Command: []string{"python", "/scripts/analyze.py", "--input", "/data/sample1.csv"},
                Dependencies: []compute.TaskDependency{}, // No dependencies for first task
            },
            {
                Label: "Analysis task 2",
                Command: []string{"python", "/scripts/analyze.py", "--input", "/data/sample2.csv"},
                Dependencies: []compute.TaskDependency{
                    {
                        TaskIndex: 0, // Depends on first task
                        Type:      "TASK_COMPLETE",
                    },
                },
            },
        },
    },
)

if err != nil {
    log.Fatalf("Failed to create batch: %v", err)
}

batchID := batch.ID
```

### How do I work with task dependencies in compute?

Managing task dependencies allows for complex workflows:

```go
// Create a batch with complex dependencies
batch, err := computeClient.CreateBatch(
    context.Background(),
    compute.CreateBatchRequest{
        EnvironmentID: "environment_id",
        Label: "Multi-stage analysis",
        Tasks: []compute.Task{
            {
                Label: "Data preparation",
                Command: []string{"python", "/scripts/prepare.py"},
                Dependencies: []compute.TaskDependency{}, // No dependencies
            },
            {
                Label: "Analysis stage 1",
                Command: []string{"python", "/scripts/analyze_stage1.py"},
                Dependencies: []compute.TaskDependency{
                    {
                        TaskIndex: 0,
                        Type:      "TASK_COMPLETE",
                    },
                },
            },
            {
                Label: "Analysis stage 2",
                Command: []string{"python", "/scripts/analyze_stage2.py"},
                Dependencies: []compute.TaskDependency{
                    {
                        TaskIndex: 1,
                        Type:      "TASK_COMPLETE",
                    },
                },
            },
            {
                Label: "Generate reports",
                Command: []string{"python", "/scripts/report.py"},
                Dependencies: []compute.TaskDependency{
                    {
                        TaskIndex: 2,
                        Type:      "TASK_COMPLETE",
                    },
                },
            },
        },
    },
)
```

### How do I integrate compute with transfer operations?

Combining compute and transfer operations is a common pattern:

```go
// First, perform the transfer
transferClient := client.Transfer()
transferTask, err := transferClient.SubmitTransfer(
    context.Background(),
    transfer.TransferOptions{
        SourceEndpointID: "source_endpoint_id",
        DestinationEndpointID: "compute_endpoint_id", // Same as the compute environment
        Items: []transfer.TransferItem{
            {
                SourcePath: "/path/on/source/data.csv",
                DestinationPath: "/path/on/compute/data.csv",
            },
        },
    },
)

if err != nil {
    log.Fatalf("Transfer failed: %v", err)
}

// Wait for the transfer to complete
transferClient.WaitForTask(context.Background(), transferTask.TaskID)

// Now submit the compute job that uses the transferred data
computeClient := client.Compute()
batch, err := computeClient.CreateBatch(
    context.Background(),
    compute.CreateBatchRequest{
        EnvironmentID: "environment_id",
        Label: "Process transferred data",
        Tasks: []compute.Task{
            {
                Label: "Process data",
                Command: []string{"python", "/scripts/process.py", "--input", "/path/on/compute/data.csv"},
            },
        },
    },
)
```

## Error Handling and Debugging

### How should I handle API errors?

The SDK provides structured error handling:

```go
import (
    "errors"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// Attempt an operation
result, err := transferClient.SubmitTransfer(
    context.Background(),
    transfer.TransferOptions{
        // ...
    },
)

if err != nil {
    // Check for specific error types
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: Code=%d, Message=%s\n", apiErr.Code, apiErr.Message)
        
        // Check for specific API error codes
        if apiErr.Code == 409 {
            fmt.Println("Conflict error - resource already exists")
        }
    }
    
    // Check for service-specific errors
    var transferErr *transfer.TransferError
    if errors.As(err, &transferErr) {
        fmt.Printf("Transfer Error: Type=%s, Detail=%s\n", 
            transferErr.ErrorType, transferErr.Detail)
            
        if transferErr.ErrorType == "PERMISSION_DENIED" {
            fmt.Println("Permission denied - check endpoint access")
        }
    }
    
    // Check for context errors
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Operation timed out")
    }
    
    return
}
```

### How do I enable debug logging?

For troubleshooting, you can enable detailed logging:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"

// Enable debug logging
logger := logging.NewLogger(
    logging.WithLevel(logging.LevelDebug),
    logging.WithOutput(os.Stdout),
)

// Create client with debug logging
client := pkg.NewSDKClient(pkg.WithLogging(logger))

// For HTTP transport debugging (very verbose)
client := pkg.NewSDKClient(
    pkg.WithLogging(logger),
    pkg.WithHTTPDebug(true),  // Enables HTTP request/response logging
)
```

### How can I trace requests for debugging?

The SDK supports request tracing for debugging:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"

// Enable tracing
logger := logging.NewLogger(
    logging.WithTracing(true),  // Enables request tracing
)

// Create client with tracing
client := pkg.NewSDKClient(pkg.WithLogging(logger))

// Make a request with a trace ID
ctx := logging.WithTraceID(context.Background(), "my-trace-id")
result, err := transferClient.ListEndpoints(ctx)

// Any logs associated with this request will include the trace ID
```

### How do I implement retry logic for failed operations?

The SDK has built-in retry logic that you can configure:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
)

// Configure retry behavior
backoff := ratelimit.NewExponentialBackoff(
    ratelimit.WithInitialInterval(1*time.Second),
    ratelimit.WithMaxInterval(30*time.Second),
    ratelimit.WithMaxRetries(5),
)

// Create client with custom retry policy
client := pkg.NewSDKClient(pkg.WithRetryPolicy(backoff))

// For specific service operations
transferClient := client.Transfer()
result, err := transferClient.SubmitTransfer(
    context.Background(),
    transfer.TransferOptions{
        // ...
    },
)
```

## Performance Optimization

### How do I optimize the SDK for high-throughput operations?

For high-throughput scenarios, configure the connection pool and concurrency settings:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
)

// Configure the connection pool
poolConfig := http.PoolConfig{
    MaxIdleConns:        100,     // Maximum idle connections
    MaxIdleConnsPerHost: 10,      // Maximum idle connections per host
    MaxConnsPerHost:     100,     // Maximum connections per host
    IdleConnTimeout:     90 * time.Second,
}

// Create client with optimized connection pool
client := pkg.NewSDKClient(pkg.WithConnectionPool(poolConfig))

// For batch operations, specify concurrency limits
transferClient := client.Transfer()
transferClient.SubmitBatchTransfer(
    context.Background(),
    transfer.BatchTransferOptions{
        // Basic transfer options
        SourceEndpointID:      "source_endpoint_id",
        DestinationEndpointID: "destination_endpoint_id",
        
        // Items to transfer
        Items: []transfer.TransferItem{
            // Many items...
        },
        
        // Performance settings
        BatchSize:      100,  // Items per batch
        Concurrency:    5,    // Concurrent batches
    },
)
```

### How can I monitor performance metrics?

The SDK provides metrics collection:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/metrics"

// Create a metrics collector
collector := metrics.NewCollector()

// Create client with metrics
client := pkg.NewSDKClient(pkg.WithMetrics(collector))

// Use the client as normal
transferClient := client.Transfer()
result, err := transferClient.ListEndpoints(context.Background())

// Get metrics
stats := collector.GetStats()
fmt.Printf("API Calls: %d\n", stats.APICallCount)
fmt.Printf("Average Response Time: %v\n", stats.AverageResponseTime)
fmt.Printf("Error Rate: %.2f%%\n", stats.ErrorRate * 100)

// Reset metrics for a new measurement period
collector.Reset()
```

### How do I implement rate limiting to avoid API throttling?

Configure rate limiting to avoid hitting API limits:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
)

// Configure rate limiting
limiter := ratelimit.NewLimiter(
    ratelimit.WithQPS(10),  // 10 queries per second maximum
    ratelimit.WithBurst(20), // Allow bursts up to 20 queries
)

// Create client with rate limiting
client := pkg.NewSDKClient(pkg.WithRateLimiter(limiter))
```

### How do I handle large data sets efficiently?

For large data sets, use streaming iterators and pagination:

```go
// Transfer a large number of files efficiently
transferClient := client.Transfer()
options := transfer.StreamingOptions{
    SourceEndpointID:      "source_endpoint_id",
    SourcePath:            "/source/directory",
    DestinationEndpointID: "destination_endpoint_id",
    DestinationPath:       "/destination/directory",
    BatchSize:             1000,   // Files per transfer task
    Recursive:             true,   // Process directories recursively
}

iterator := transferClient.NewStreamingIterator(context.Background(), options)

for iterator.Next() {
    batch := iterator.Batch()
    taskID, err := transferClient.SubmitBatch(context.Background(), batch)
    if err != nil {
        // Handle error
        continue
    }
    
    // Store task ID for monitoring if needed
    fmt.Printf("Submitted batch, task ID: %s\n", taskID)
}

if err := iterator.Error(); err != nil {
    log.Fatalf("Iterator error: %v", err)
}
```

## SDK Architecture

### How is the SDK structured internally?

The SDK follows a layered architecture:

1. **Core Layer**: Provides fundamental building blocks
   - Base client with context support
   - HTTP transport handling
   - Authentication mechanisms
   - Error handling and logging

2. **Service Layer**: Implements Globus service APIs
   - Auth service
   - Transfer service
   - Search service
   - Flow service
   - Compute service
   - Groups service
   - Timers service

3. **Utility Layer**: Provides cross-cutting functionality
   - Metrics collection
   - Progress tracking
   - Token management
   - Connection pooling

The design emphasizes:
- Consistent interfaces across services
- Context awareness for proper cancellation
- Error handling with structured errors
- Configurability via options pattern

### How do I extend the SDK with custom functionality?

You can extend the SDK in several ways:

```go
// 1. Create custom authorizers
type MyCustomAuthorizer struct {
    // Custom state
}

func (a *MyCustomAuthorizer) GetToken(ctx context.Context) (string, error) {
    // Custom token acquisition logic
    return "Bearer my-custom-token", nil
}

// Use your custom authorizer
client := pkg.NewSDKClient(pkg.WithAuthorizer(myCustomAuthorizer))

// 2. Create custom HTTP middleware
func myLoggingMiddleware(next http.RoundTripper) http.RoundTripper {
    return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
        fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
        resp, err := next.RoundTrip(req)
        if err == nil {
            fmt.Printf("Response: %d %s\n", resp.StatusCode, resp.Status)
        }
        return resp, err
    })
}

// Use your custom middleware
transport := pkg.NewHTTPTransport(pkg.WithMiddleware(myLoggingMiddleware))
client := pkg.NewSDKClient(pkg.WithTransport(transport))

// 3. Implement custom token storage
type MyTokenStorage struct {
    // Custom storage details
}

func (s *MyTokenStorage) StoreToken(resource string, token auth.Token) error {
    // Custom token storage logic
    return nil
}

func (s *MyTokenStorage) GetToken(resource string) (auth.Token, error) {
    // Custom token retrieval logic
    return auth.Token{}, nil
}

// Use your custom token storage
tokenManager := tokens.NewManager(tokens.WithStorage(myTokenStorage))
client := pkg.NewSDKClient(pkg.WithTokenManager(tokenManager))
```

### How does the SDK handle concurrency and goroutines?

The SDK uses several concurrency patterns:

1. **Context-based cancellation**: All operations accept a context.Context for timeout and cancellation
2. **Goroutine safety**: All client methods are safe for concurrent use
3. **Connection pooling**: HTTP connections are pooled for efficient reuse
4. **Synchronization**: Mutexes protect shared resources where needed

Example of context-based cancellation:

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Make a request that can be cancelled
result, err := transferClient.ListEndpoints(ctx)
if errors.Is(err, context.DeadlineExceeded) {
    fmt.Println("Operation timed out")
}

// For long-running operations with manual cancellation
ctx, cancel := context.WithCancel(context.Background())

// In another goroutine or signal handler
go func() {
    time.Sleep(5 * time.Second)
    fmt.Println("Cancelling operation")
    cancel()
}()

// Start an operation that might be cancelled
err := transferClient.WaitForTask(ctx, taskID)
if errors.Is(err, context.Canceled) {
    fmt.Println("Operation was cancelled")
}
```

### What's the difference between the various client implementations?

The SDK provides different client types for different needs:

1. **Base Client**: Core implementation used by all services
   - Handles HTTP requests, auth, errors, and logging
   - Not typically used directly

2. **SDK Client**: Main entry point for the SDK
   - Factory for service-specific clients
   - Shares configuration across services

3. **Service Clients**: Implement specific Globus services
   - Auth, Transfer, Search, etc.
   - Provide service-specific methods

4. **Pool Client**: Optimized for high-throughput scenarios
   - Uses connection pooling
   - Configurable concurrent connections

Example of client relationships:

```go
// Create the main SDK client
sdkClient := pkg.NewSDKClient()

// Get service-specific clients
authClient := sdkClient.Auth()
transferClient := sdkClient.Transfer()
searchClient := sdkClient.Search()

// All share the same underlying configuration
// Changes to one affect all (e.g., token refresh)
```

### How do I contribute to the SDK?

To contribute to the Globus Go SDK:

1. **Fork the repository**: Create your own fork on GitHub
2. **Set up development environment**: 
   ```bash
   git clone https://github.com/your-username/globus-go-sdk.git
   cd globus-go-sdk
   make deps  # Install dependencies
   ```
3. **Run tests**: 
   ```bash
   make test  # Run unit tests
   make integration  # Run integration tests (requires credentials)
   ```
4. **Make changes**: Implement your feature or fix
5. **Add tests**: Ensure your changes are covered by tests
6. **Document**: Update documentation as needed
7. **Format and lint**: 
   ```bash
   make fmt  # Format code
   make lint  # Run linters
   ```
8. **Submit PR**: Create a pull request with a clear description

For detailed contribution guidelines, see the [CONTRIBUTING.md](https://github.com/scttfrdmn/globus-go-sdk/blob/main/CONTRIBUTING.md) file.
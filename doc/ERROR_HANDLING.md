# Error Handling and Rate Limiting

This document describes the error handling and rate limiting strategies in the Globus Go SDK (v0.8.0+).

## Error Types

The Globus Go SDK provides several error types to help identify and handle specific error cases:

1. **Service-specific errors**: Each service (transfer, auth, etc.) has its own error types for specific error conditions.
2. **Transport errors**: Errors related to HTTP communication.
3. **Rate limiting errors**: Errors indicating that a rate limit has been reached.
4. **Permission errors**: Errors indicating that the user doesn't have permission to perform an action.
5. **Resource not found errors**: Errors indicating that a requested resource doesn't exist.

## Error Checking

To check for specific error types, use the helper functions provided by each service package:

```go
// Check for rate limit errors
if transfer.IsRateLimitExceeded(err) {
    log.Printf("Rate limit exceeded: %v", err)
    // Implement backoff or retry logic
}

// Check for permission errors
if transfer.IsPermissionDenied(err) {
    log.Printf("Permission denied: %v", err)
    // Handle permission issues
}

// Check for resource not found errors
if transfer.IsResourceNotFound(err) {
    log.Printf("Resource not found: %v", err)
    // Handle missing resources
}
```

## Rate Limiting and Retries

The SDK provides built-in support for handling rate limiting through the `ratelimit` package.

### Using RetryWithBackoff

The `RetryWithBackoff` function automatically retries operations when they fail due to rate limiting or transient errors:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// Retry a transfer operation with backoff
err := ratelimit.RetryWithBackoff(
    ctx,
    func(ctx context.Context) error {
        return client.Mkdir(ctx, endpointID, dirPath)
    },
    ratelimit.DefaultBackoff(),
    transfer.IsRetryableTransferError,
)
```

### Customizing Backoff Behavior

You can customize the backoff behavior by creating a custom backoff configuration:

```go
// Create a custom backoff with longer delays
customBackoff := &ratelimit.ExponentialBackoff{
    InitialDelay: 1 * time.Second,
    MaxDelay:     60 * time.Second,
    Factor:       2.0,
    Jitter:       true,
    MaxAttempt:   10,
}

// Use the custom backoff with RetryWithBackoff
err := ratelimit.RetryWithBackoff(
    ctx,
    func(ctx context.Context) error {
        return client.Mkdir(ctx, endpointID, dirPath)
    },
    customBackoff,
    transfer.IsRetryableTransferError,
)
```

### Customizing Retry Conditions

You can also customize which errors should trigger a retry by providing a custom retry function:

```go
// Custom retry function that only retries on rate limit errors
customRetryFunc := func(err error) bool {
    return transfer.IsRateLimitExceeded(err)
}

// Use the custom retry function with RetryWithBackoff
err := ratelimit.RetryWithBackoff(
    ctx,
    func(ctx context.Context) error {
        return client.Mkdir(ctx, endpointID, dirPath)
    },
    ratelimit.DefaultBackoff(),
    customRetryFunc,
)
```

## Circuit Breaker

For more advanced error handling, the SDK includes a circuit breaker implementation that can prevent repeated failures:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
)

// Create a circuit breaker
cb := ratelimit.NewCircuitBreaker(
    ratelimit.WithFailureThreshold(5),
    ratelimit.WithResetTimeout(30 * time.Second),
)

// Use the circuit breaker
result, err := cb.Execute(func() (interface{}, error) {
    return client.ListEndpoints(ctx, nil)
})
if err != nil {
    if ratelimit.IsCircuitOpenError(err) {
        log.Printf("Circuit is open, too many failures")
    } else {
        log.Printf("Operation failed: %v", err)
    }
    return
}

// Cast the result to the expected type
endpoints := result.(*transfer.EndpointList)
```

## Best Practices

1. **Always check errors**: Always check and handle errors properly.
2. **Use RetryWithBackoff for API calls**: Use `RetryWithBackoff` for operations that might be affected by rate limiting.
3. **Provide context timeouts**: Set appropriate context timeouts for operations to prevent indefinite retries.
4. **Log and monitor**: Log rate limiting errors and monitor their frequency to adjust your application's behavior.
5. **Implement backoff with jitter**: Use exponential backoff with jitter to avoid thundering herd problems.

```go
// Example of a complete operation with proper error handling
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
defer cancel()

err := ratelimit.RetryWithBackoff(
    ctx,
    func(ctx context.Context) error {
        return client.Mkdir(ctx, endpointID, dirPath)
    },
    ratelimit.DefaultBackoff(),
    transfer.IsRetryableTransferError,
)

if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Printf("Operation timed out after retries: %v", err)
    } else if transfer.IsPermissionDenied(err) {
        log.Printf("Permission denied: %v", err)
    } else if transfer.IsRateLimitExceeded(err) {
        log.Printf("Rate limit exceeded even after retries: %v", err)
    } else {
        log.Printf("Operation failed: %v", err)
    }
    return
}

log.Printf("Operation succeeded")
```
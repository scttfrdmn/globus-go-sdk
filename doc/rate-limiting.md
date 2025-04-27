# Rate Limiting and Retry Strategies

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This guide explains how to use the rate limiting, backoff, and circuit breaker functionality in the Globus Go SDK.

## Overview

The Globus Go SDK provides robust mechanisms for handling API rate limits and transient failures:

1. **Rate Limiting**: Controls request rates to avoid hitting service limits
2. **Backoff Strategy**: Implements retry logic with exponential backoff
3. **Circuit Breaker**: Prevents cascading failures when services degrade

These components work together to make your applications more resilient and respectful of service constraints.

## Rate Limiting

The `RateLimiter` interface provides rate limiting functionality:

```go
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
```

### Creating a Rate Limiter

```go
// Create a token bucket rate limiter
options := &ratelimit.RateLimiterOptions{
    RequestsPerSecond: 10.0,
    BurstSize:         20,
    UseAdaptive:       true,
    MaxRetryCount:     5,
    MinRetryDelay:     100 * time.Millisecond,
    MaxRetryDelay:     60 * time.Second,
    UseJitter:         true,
    JitterFactor:      0.2,
}

limiter := ratelimit.NewTokenBucketLimiter(options)
```

### Using the Rate Limiter

```go
// Wait before making a request
ctx := context.Background()
err := limiter.Wait(ctx)
if err != nil {
    return err // Context canceled or other error
}

// Make your API request
response, err := client.MakeAPICall()

// Update rate limiter from response
ratelimit.UpdateRateLimiterFromResponse(limiter, response)
```

### Adaptive Rate Limiting

The limiter can adapt to server-provided rate limits:

```go
// Extract rate limit information from response headers
info, found := ratelimit.ExtractRateLimitInfo(response)
if found {
    fmt.Printf("Rate limit: %d, Remaining: %d, Reset: %d\n",
        info.Limit, info.Remaining, info.Reset)
}

// Update limiter based on response headers (automatic)
ratelimit.UpdateRateLimiterFromResponse(limiter, response)
```

## Backoff and Retry

The `BackoffStrategy` interface provides retry logic with exponential backoff:

```go
// Create a backoff strategy
backoff := ratelimit.NewExponentialBackoff(
    100*time.Millisecond, // Initial delay
    60*time.Second,       // Max delay
    2.0,                  // Factor (doubles each retry)
    5,                    // Max attempts
)

// Use retry with backoff
err := ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
    // Make your API call
    return client.MakeAPICall(ctx)
}, backoff, ratelimit.IsRetryableError)
```

### Customizing Retry Logic

You can customize retry behavior:

```go
// Custom retry decision function
shouldRetry := func(err error) bool {
    // Custom logic to determine if the error is retryable
    if err == nil {
        return false
    }
    
    // Example: retry rate limit errors
    if strings.Contains(err.Error(), "rate limit") {
        return true
    }
    
    return false
}

// Use custom retry logic
err := ratelimit.RetryWithBackoff(ctx, apiCall, backoff, shouldRetry)
```

## Circuit Breaker

The circuit breaker pattern prevents cascading failures:

```go
// Create a circuit breaker
options := &ratelimit.CircuitBreakerOptions{
    Threshold:         5,                // Open after 5 failures
    Timeout:           30 * time.Second, // Half-open after 30 seconds
    HalfOpenSuccesses: 2,                // Close after 2 successes
    OnStateChange: func(from, to ratelimit.CircuitBreakerState) {
        fmt.Printf("Circuit state changed: %v -> %v\n", from, to)
    },
}

cb := ratelimit.NewCircuitBreaker(options)

// Use the circuit breaker
err := cb.Execute(ctx, func(ctx context.Context) error {
    // Make your API call
    return client.MakeAPICall(ctx)
})

if errors.Is(err, ratelimit.ErrCircuitOpen) {
    // Circuit is open, handle accordingly
    return fallbackOperation()
}
```

## Putting It All Together

Here's a complete example that combines all three mechanisms:

```go
func resilientAPICall(ctx context.Context, client *MyClient, req *Request) (*Response, error) {
    // Create rate limiter
    limiter := ratelimit.NewTokenBucketLimiter(nil) // Use defaults
    
    // Create backoff strategy
    backoff := ratelimit.DefaultBackoff()
    
    // Create circuit breaker
    cb := ratelimit.NewCircuitBreaker(nil) // Use defaults
    
    // Execute with all resilience mechanisms
    var response *Response
    
    err := cb.Execute(ctx, func(ctx context.Context) error {
        // Wait for rate limiter
        if err := limiter.Wait(ctx); err != nil {
            return err
        }
        
        // Call API with retries
        return ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
            resp, err := client.MakeAPICall(req)
            if err != nil {
                return err
            }
            
            // Update rate limiter from response
            ratelimit.UpdateRateLimiterFromResponse(limiter, resp.RawResponse)
            
            // Store response for outer function
            response = resp
            return nil
        }, backoff, ratelimit.IsRetryableError)
    })
    
    return response, err
}
```

## Integration with Core Client

The SDK integrates these resilience mechanisms into the core client:

```go
// Create a client with rate limiting
client := core.NewClient(
    core.WithRateLimit(10.0),          // 10 requests per second
    core.WithRetryCount(3),            // Retry up to 3 times
    core.WithCircuitBreaker(true),     // Enable circuit breaker
)
```

## Best Practices

1. **Use Appropriate Limits**: Start with conservative rates and adjust based on experience
2. **Enable Adaptive Limits**: Allow the limiter to adjust based on server responses
3. **Add Jitter to Retries**: Prevent thundering herd problems with randomized delays
4. **Monitor Circuit Breaker**: Log state changes to understand service health
5. **Use Timeouts**: Set reasonable timeouts on contexts to avoid waiting too long
6. **Implement Fallbacks**: Have fallback mechanisms when services are unavailable

## Configuration Recommendations

| Service | Recommended Rate | Burst Size | Max Retries |
|---------|-----------------|------------|------------|
| Auth    | 10 req/s        | 20         | 3          |
| Transfer| 5 req/s         | 10         | 5          |
| Search  | 3 req/s         | 6          | 3          |
| Flows   | 5 req/s         | 10         | 3          |

## Further Reading

- [Example applications](../examples/ratelimit/)
- [API reference documentation](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit)
- [Globus API rate limits documentation](https://docs.globus.org/api/)
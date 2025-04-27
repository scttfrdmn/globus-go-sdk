<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Error Handling in the Globus Go SDK

This document explains how to handle errors when using the Globus Go SDK.

## Table of Contents

- [Error Types](#error-types)
- [Error Checking Functions](#error-checking-functions)
- [Common Error Patterns](#common-error-patterns)
- [Service-Specific Errors](#service-specific-errors)
- [Best Practices](#best-practices)

## Error Types

The SDK uses a structured error system with several key error types:

### SDKError

The `SDKError` is the base error interface for all SDK errors:

```go
type SDKError interface {
    error
    StatusCode() int
    ErrorCode() string
    ErrorMessage() string
    RequestID() string
    Unwrap() error
}
```

### APIError

The `APIError` type represents errors returned by the Globus API:

```go
type APIError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    Status     int    `json:"status"`
    RequestID  string `json:"request_id"`
    Cause      error  `json:"-"`
}
```

### ValidationError

The `ValidationError` type represents errors in client-side validation:

```go
type ValidationError struct {
    Field   string
    Message string
    Cause   error
}
```

### RateLimitError

The `RateLimitError` type is returned when API rate limits are exceeded:

```go
type RateLimitError struct {
    APIError
    RetryAfter time.Duration
}
```

### AuthenticationError

The `AuthenticationError` type represents authentication failures:

```go
type AuthenticationError struct {
    APIError
    TokenInfo *auth.TokenInfo
}
```

## Error Checking Functions

The SDK provides helper functions to check for specific error types:

### Common Error Checkers

```go
// Check if an error is an API error
isAPIError := errors.IsAPIError(err)

// Check if an error is a validation error
isValidationError := errors.IsValidationError(err)

// Check if an error is a rate limit error
isRateLimitError := errors.IsRateLimitError(err)

// Check if an error is an authentication error
isAuthError := errors.IsAuthenticationError(err)

// Check if an error is a context cancellation error
isContextCanceled := errors.IsContextCanceledError(err)
```

### Status Code Checkers

```go
// Check if an error has a specific status code
isNotFound := errors.IsStatusCode(err, http.StatusNotFound)
isBadRequest := errors.IsStatusCode(err, http.StatusBadRequest)
isServerError := errors.IsStatusCode(err, http.StatusInternalServerError)
```

## Common Error Patterns

### Handling API Errors

```go
result, err := client.SomeOperation(ctx, params)
if err != nil {
    if apiErr, ok := errors.AsAPIError(err); ok {
        fmt.Printf("API Error: %s (Code: %s, Status: %d, Request ID: %s)\n",
            apiErr.Message, apiErr.Code, apiErr.Status, apiErr.RequestID)
        
        // You can handle specific API error codes
        if apiErr.Code == "UnauthorizedRequest" {
            // Handle unauthorized request
        } else if apiErr.Code == "ResourceNotFound" {
            // Handle resource not found
        }
        return
    }
    
    // Handle other errors
    fmt.Printf("Error: %v\n", err)
    return
}
```

### Handling Rate Limits

```go
result, err := client.SomeOperation(ctx, params)
if err != nil {
    if rateErr, ok := errors.AsRateLimitError(err); ok {
        fmt.Printf("Rate limit exceeded. Retry after: %v\n", rateErr.RetryAfter)
        
        // You can implement automatic retry with exponential backoff
        time.Sleep(rateErr.RetryAfter)
        return client.SomeOperation(ctx, params)
    }
    
    // Handle other errors
    fmt.Printf("Error: %v\n", err)
    return
}
```

### Handling Authentication Errors

```go
result, err := client.SomeOperation(ctx, params)
if err != nil {
    if authErr, ok := errors.AsAuthenticationError(err); ok {
        fmt.Println("Authentication error. The token may have expired.")
        
        // You can attempt to refresh the token
        if tokenManager != nil {
            // Force refresh the token
            newToken, refreshErr := tokenManager.RefreshToken(ctx, "user-key")
            if refreshErr != nil {
                fmt.Printf("Error refreshing token: %v\n", refreshErr)
                return
            }
            
            // Try the operation again with the new token
            return client.SomeOperation(ctx, params)
        }
        return
    }
    
    // Handle other errors
    fmt.Printf("Error: %v\n", err)
    return
}
```

### Handling Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.SomeOperation(ctx, params)
if err != nil {
    if errors.IsContextCanceledError(err) {
        fmt.Println("Operation timed out or was canceled")
        return
    }
    
    // Handle other errors
    fmt.Printf("Error: %v\n", err)
    return
}
```

## Service-Specific Errors

### Auth Service Errors

```go
token, err := authClient.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    switch {
    case auth.IsInvalidGrantError(err):
        fmt.Println("Invalid authorization code")
    case auth.IsInvalidClientError(err):
        fmt.Println("Invalid client credentials")
    case auth.IsInvalidScopeError(err):
        fmt.Println("Invalid scope requested")
    default:
        fmt.Printf("Auth error: %v\n", err)
    }
    return
}
```

### Transfer Service Errors

```go
task, err := transferClient.SubmitTransfer(ctx, sourceEP, destEP, transferData)
if err != nil {
    switch {
    case transfer.IsEndpointNotFoundError(err):
        fmt.Println("One of the endpoints was not found")
    case transfer.IsEndpointPermissionDeniedError(err):
        fmt.Println("Permission denied on one of the endpoints")
    case transfer.IsPathNotFoundError(err):
        fmt.Println("One of the paths was not found")
    case transfer.IsTaskQueueFullError(err):
        fmt.Println("Transfer queue is full, try again later")
    default:
        fmt.Printf("Transfer error: %v\n", err)
    }
    return
}
```

### Groups Service Errors

```go
group, err := groupsClient.GetGroup(ctx, groupID)
if err != nil {
    switch {
    case groups.IsGroupNotFoundError(err):
        fmt.Println("Group not found")
    case groups.IsPermissionDeniedError(err):
        fmt.Println("Permission denied to access this group")
    default:
        fmt.Printf("Groups error: %v\n", err)
    }
    return
}
```

## Best Practices

### 1. Check for Specific Error Types First

When handling errors, always check for specific error types before falling back to generic handling:

```go
if err != nil {
    // Check for specific error types first
    if errors.IsRateLimitError(err) {
        // Handle rate limiting
    } else if errors.IsAuthenticationError(err) {
        // Handle authentication issues
    } else if errors.IsContextCanceledError(err) {
        // Handle context cancellation
    } else {
        // Fall back to generic error handling
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

### 2. Use Error Wrapping

When creating your own errors, wrap the original error to preserve the error chain:

```go
if err != nil {
    return nil, fmt.Errorf("failed to submit transfer: %w", err)
}
```

### 3. Log Request IDs

When an error contains a request ID, always log it to help with troubleshooting:

```go
if apiErr, ok := errors.AsAPIError(err); ok {
    log.Printf("API Error: %s (Request ID: %s)", apiErr.Message, apiErr.RequestID)
}
```

### 4. Implement Retry Logic for Transient Errors

Some errors are transient and can be retried. Implement retry logic with exponential backoff:

```go
func retryOperation(ctx context.Context, maxRetries int, fn func() error) error {
    var err error
    
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        // Only retry certain types of errors
        if errors.IsRateLimitError(err) || errors.IsStatusCode(err, http.StatusServiceUnavailable) {
            // Calculate backoff duration
            backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
            
            // Add some jitter
            backoff = backoff + time.Duration(rand.Intn(1000))*time.Millisecond
            
            // Respect context cancellation
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(backoff):
                // Continue with retry
            }
            continue
        }
        
        // For non-retriable errors, return immediately
        return err
    }
    
    return err
}
```

### 5. Use Context for Cancellation

Always use context for cancellation and timeout handling:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.SomeOperation(ctx, params)
if err != nil {
    if errors.IsContextCanceledError(err) {
        // Handle timeout or cancellation
    }
    return
}
```

### 6. Provide User-Friendly Error Messages

When displaying errors to users, translate technical error messages into user-friendly terms:

```go
func userFriendlyError(err error) string {
    if errors.IsAuthenticationError(err) {
        return "Your session has expired. Please log in again."
    }
    
    if errors.IsRateLimitError(err) {
        return "The service is currently busy. Please try again in a few moments."
    }
    
    if transfer.IsPathNotFoundError(err) {
        return "The specified file or folder could not be found."
    }
    
    if transfer.IsEndpointPermissionDeniedError(err) {
        return "You don't have permission to access this location."
    }
    
    // Default user-friendly message
    return "An error occurred while processing your request. Please try again later."
}
```
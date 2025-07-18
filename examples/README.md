# Globus Go SDK Examples

This directory contains examples demonstrating the use of the Globus Go SDK v3.60.0 with unified systems.

## v3.60.0 Unified Systems Features

All examples have been updated to showcase the new unified systems introduced in v3.60.0:

### 🔄 Unified Configuration
- `client.AuthConfig()`, `client.TransferConfig()`, etc. for consistent configuration
- Shared options across all services (timeout, retries, debug mode)
- Backwards compatibility with existing service-specific options

### 🚨 Unified Error Handling
- `GlobusError` type with service context and request tracking
- Consistent error classification (auth, not found, retryable, etc.)
- Rich error metadata with request IDs and context

### 📦 Unified Response System
- `Response[T]` wrapper with metadata for all responses
- `PaginatedResponse[T]` with iterator pattern for large datasets
- Consistent response metadata across all services

### ⚠️ Deprecation System
- Automatic warnings for deprecated methods
- Clear migration guidance
- Environment variable control (`GLOBUS_SDK_DEPRECATION_WARNINGS`)

## Examples Overview

### Core Examples

#### [unified-systems/](unified-systems/)
**Complete demonstration of v3.60.0 unified systems**
- Shows all unified systems working together
- Deprecation warnings in action
- Error handling examples
- Response system with pagination
- **Start here** for v3.60.0 features

#### [service-clients-functional-options/](service-clients-functional-options/)
**Service client creation with unified configuration**
- Updated to show both old and new configuration methods
- Demonstrates deprecation warnings
- Shows unified vs. legacy patterns

### Authentication Examples

#### [mfa-auth/](mfa-auth/)
**Multi-Factor Authentication with v3.60.0**
- Updated to use unified auth configuration
- Shows backwards compatibility
- Demonstrates unified response handling

#### [token-management/](token-management/)
**Token storage and management with unified systems**
- Memory and file-based token storage
- Automatic token refreshing
- Updated to show v3.60.0 patterns

### Transfer Examples

#### [resumable-transfer/](resumable-transfer/)
**Resumable transfers with unified configuration**
- Updated to use unified transfer config
- Shows advanced configuration options
- Demonstrates error handling improvements

#### [data-pipeline/](data-pipeline/)
**Data pipeline with multiple services**
- Shows unified error handling across services
- Demonstrates response metadata usage
- Service integration patterns

### Other Examples

#### [webapp/](webapp/)
**Web application integration**
- OAuth2 flow with unified systems
- Session management with unified responses
- Error handling in web context

#### [benchmark/](benchmark/)
**Performance testing with unified systems**
- Shows unified configuration for high-performance scenarios
- Demonstrates error handling in concurrent operations

## Running Examples

### Prerequisites
```bash
# Set up environment variables
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
export GLOBUS_ACCESS_TOKEN="your-access-token"  # For some examples

# Optional: Enable deprecation warnings
export GLOBUS_SDK_DEPRECATION_WARNINGS="true"

# Optional: Enable debug mode
export GLOBUS_SDK_DEBUG="1"
```

### Running Individual Examples
```bash
# Start with the unified systems example
cd unified-systems
go run main.go

# Try the service clients example
cd ../service-clients-functional-options
go run main.go

# Test token management
cd ../token-management
go run main.go
```

## Migration from v0.9.x

### Configuration Changes
```go
// Old (v0.9.x)
client, err := auth.NewClient(
    auth.WithClientID("id"),
    auth.WithClientSecret("secret"),
)

// New (v3.60.0) - both approaches work
// Option 1: Unified configuration
config, err := client.AuthConfig(
    client.WithClientCredentials("id", "secret"),
    client.WithTimeout(30*time.Second),
)

// Option 2: Legacy method (still works, shows deprecation warning)
client, err := auth.NewClient(
    auth.WithClientID("id"),
    auth.WithClientSecret("secret"),
)
```

### Error Handling Changes
```go
// Old (v0.9.x)
if err != nil {
    log.Printf("Error: %v", err)
}

// New (v3.60.0) - enhanced error handling
if err != nil {
    var globusErr *errors.GlobusError
    if errors.As(err, &globusErr) {
        log.Printf("Service: %s, Code: %s, RequestID: %s", 
            globusErr.Service, globusErr.Code, globusErr.RequestID)
        
        if globusErr.IsRetryable() {
            // Handle retryable errors
        }
    }
}
```

### Response Handling Changes
```go
// Old (v0.9.x)
tokenResp, err := client.ExchangeAuthorizationCode(ctx, code)

// New (v3.60.0) - enhanced responses
authResp, err := client.ExchangeAuthorizationCodeV2(ctx, code)
// authResp.Data contains the token
// authResp.Metadata contains response metadata
// authResp.RequestID contains the request ID
```

## Best Practices

### 1. Use Unified Configuration
```go
config, err := client.AuthConfig(
    client.WithClientCredentials(clientID, clientSecret),
    client.WithTimeout(30*time.Second),
    client.WithMaxRetries(3),
)
```

### 2. Handle Errors Consistently
```go
if err != nil {
    var globusErr *errors.GlobusError
    if errors.As(err, &globusErr) {
        // Handle Globus-specific errors
        fmt.Printf("Request ID: %s\n", globusErr.RequestID)
    }
}
```

### 3. Use Response Metadata
```go
resp, err := client.SomeMethodV2(ctx, params)
if err == nil {
    fmt.Printf("Request ID: %s\n", resp.RequestID)
    fmt.Printf("Service: %s\n", resp.Metadata.Service)
}
```

### 4. Enable Deprecation Warnings
```bash
export GLOBUS_SDK_DEPRECATION_WARNINGS="true"
```

### 5. Use Iterators for Pagination
```go
iterator := response.NewIterator(paginatedResp, fetchNextPage)
for {
    item, ok := iterator.Next()
    if !ok {
        break
    }
    // Process item
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `GLOBUS_CLIENT_ID` | OAuth2 client ID | Required |
| `GLOBUS_CLIENT_SECRET` | OAuth2 client secret | Required |
| `GLOBUS_ACCESS_TOKEN` | Access token for API calls | Required for some examples |
| `GLOBUS_SDK_DEPRECATION_WARNINGS` | Enable deprecation warnings | `true` |
| `GLOBUS_SDK_DEBUG` | Enable debug logging | `false` |
| `GLOBUS_SDK_HTTP_DEBUG` | Enable HTTP request/response debugging | `false` |
| `GLOBUS_SDK_HTTP_TRACE` | Enable HTTP tracing | `false` |

## Support

For questions about the examples or the SDK:
- Check the [main documentation](../README.md)
- Review the [CHANGELOG](../CHANGELOG.md) for v3.60.0 changes
- Look at the [unified systems example](unified-systems/) for comprehensive usage
- File issues at [GitHub Issues](https://github.com/scttfrdmn/globus-go-sdk/issues)

## Version Compatibility

- ✅ **v3.60.0**: Full unified systems support
- ⚠️ **v0.9.x**: Deprecated patterns (still work with warnings)
- ❌ **< v0.9.x**: Not supported

All examples are tested with Go 1.19+ and Globus API current versions.
# Unified Systems Example

This example demonstrates the new unified systems introduced in Globus Go SDK v3.60.0.

## Features Demonstrated

### 1. Deprecation System
- Shows how deprecation warnings are issued
- Demonstrates the new warning system that prevents spam
- Shows how to enable/disable deprecation warnings

### 2. Unified Client Configuration
- Demonstrates the new unified `client.Config` system
- Shows service-specific configuration functions
- Demonstrates consistent configuration across all services

### 3. Unified Error Handling
- Shows the new `GlobusError` type
- Demonstrates error context and metadata
- Shows error classification methods (auth, not found, retryable, etc.)

### 4. Unified Response System
- Demonstrates the new `Response[T]` wrapper
- Shows `PaginatedResponse[T]` for paginated data
- Demonstrates the iterator pattern for large datasets
- Shows response metadata extraction

## Running the Example

```bash
cd examples/unified-systems
go run main.go
```

## Expected Output

The example will show:

1. **Deprecation warnings** when calling deprecated methods
2. **Unified client configuration** for different services
3. **Consistent error handling** across all services
4. **Unified response handling** with metadata and pagination

## Environment Variables

- `GLOBUS_SDK_DEPRECATION_WARNINGS`: Set to "false" to disable deprecation warnings

## Key Benefits

### For Developers
- **Consistent APIs**: All services use the same patterns
- **Rich Error Context**: Detailed error information with request IDs
- **Unified Pagination**: Same iterator pattern across all services
- **Deprecation Guidance**: Clear migration paths for deprecated features

### For Maintainers
- **Easier Updates**: Consistent patterns make updates simpler
- **Better Testing**: Unified systems enable comprehensive testing
- **Documentation**: Single set of patterns to document

## Migration from v0.9.x

### Error Handling
```go
// Old (v0.9.x)
if err != nil {
    log.Printf("Error: %v", err)
}

// New (v3.60.0)
if err != nil {
    var globusErr *errors.GlobusError
    if errors.As(err, &globusErr) {
        log.Printf("Service: %s, Code: %s, RequestID: %s", 
            globusErr.Service, globusErr.Code, globusErr.RequestID)
    }
}
```

### Response Handling
```go
// Old (v0.9.x)
tokenResp, err := client.ExchangeAuthorizationCode(ctx, code)

// New (v3.60.0)
authResp, err := client.ExchangeAuthorizationCodeV2(ctx, code)
// authResp.Data contains the token
// authResp.Metadata contains response metadata
// authResp.RequestID contains the request ID
```

### Client Configuration
```go
// Old (v0.9.x)
client, err := auth.NewClient(
    auth.WithClientID("id"),
    auth.WithClientSecret("secret"),
)

// New (v3.60.0) - still works, but new unified approach available
config, err := client.AuthConfig(
    client.WithAccessToken("token"),
    client.WithTimeout(30*time.Second),
)
```

## Related Examples

- [Auth Example](../auth/main.go) - OAuth2 authentication flows
- [Transfer Example](../transfer/main.go) - File transfer operations
- [Groups Example](../groups/main.go) - Group management
- [Error Handling Example](../error-handling/main.go) - Advanced error handling

## Documentation

- [Error Handling Guide](../../doc/error-handling.md)
- [Response System Guide](../../doc/response-system.md)
- [Deprecation Guide](../../doc/deprecation.md)
- [Migration Guide](../../doc/V3.60.0_MIGRATION_GUIDE.md)
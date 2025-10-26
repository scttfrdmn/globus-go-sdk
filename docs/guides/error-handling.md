# Error Handling

Guide to handling errors in the Globus Go SDK.

## Error Types

### APIError

Errors returned by the Globus API:

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"

result, err := client.Method(ctx, options)
if err != nil {
    var apiErr *errors.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
    }
}
```

## Common Error Codes

- `NotFound` - Resource not found
- `Unauthorized` - Authentication failed
- `PermissionDenied` - Insufficient permissions
- `RateLimitExceeded` - Rate limit exceeded

## Retry Logic

```go
import "time"

maxRetries := 3
for i := 0; i < maxRetries; i++ {
    result, err := client.Method(ctx, options)
    if err == nil {
        break
    }

    if i < maxRetries-1 {
        time.Sleep(time.Second * time.Duration(i+1))
    }
}
```

## See Also

- [API Reference](../api/index.md)
- [Common Patterns](common-patterns.md)

# Client Configuration

Configure SDK clients with custom options for your needs.

## Default Configuration

All clients work with default configuration:

```go
client := transfer.NewClient(authorizer)
```

## Custom HTTP Client

Use a custom HTTP client with connection pooling:

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

client := transfer.NewClientWithHTTPClient(authorizer, httpClient)
```

## Base URL Override

For testing or custom deployments:

```go
client := transfer.NewClient(authorizer)
client.BaseURL = "https://custom-transfer.example.com"
```

## Timeouts

Set operation timeouts:

```go
import (
    "context"
    "time"
)

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.SomeMethod(ctx, options)
```

## Rate Limiting

The SDK handles rate limiting automatically with exponential backoff.

## Logging

Enable debug logging:

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/logging"

logging.SetLevel(logging.DebugLevel)
```

## Next Steps

- Review [Common Patterns](../guides/common-patterns.md)
- Check [Rate Limiting Guide](../guides/rate-limiting.md)
- Explore [API Reference](../api/index.md)

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Connection Pooling

This guide explains how to use the connection pooling functionality in the Globus Go SDK. Connection pooling can dramatically improve performance by reusing HTTP connections across multiple requests.

## Overview

HTTP requests can be expensive to set up, particularly when using HTTPS, due to the overhead of establishing TCP connections and performing TLS handshakes. Connection pooling helps mitigate this overhead by:

1. Reusing existing connections for multiple HTTP requests
2. Maintaining a pool of idle connections for each host
3. Controlling the number of concurrent connections
4. Providing automatic connection lifecycle management

The Globus Go SDK includes connection pooling functionality to optimize performance when making multiple API requests to Globus services.

## Basic Usage

The connection pooling functionality is built into the SDK and enabled by default. The default settings are suitable for most use cases, but you can customize them if needed.

### Using the Default Connection Pool

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

func main() {
    // Create SDK configuration - connection pooling is enabled by default
    config := pkg.NewConfigFromEnvironment()
    
    // Create clients for multiple services
    authClient := config.NewAuthClient()
    transferClient := config.NewTransferClient("your-access-token")
    searchClient := config.NewSearchClient("your-access-token")
    
    // Make requests - connections are automatically pooled and reused
    // ...
}
```

### Customizing Connection Pool Settings

For advanced use cases, you can customize the connection pool settings:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

func main() {
    // Create custom connection pool configuration
    poolConfig := &transport.ConnectionPoolConfig{
        MaxIdleConnsPerHost:   20,      // Max idle connections per host
        MaxIdleConns:          100,     // Max idle connections total
        MaxConnsPerHost:       50,      // Max connections per host
        IdleConnTimeout:       120 * time.Second,
        DisableKeepAlives:     false,
        ResponseHeaderTimeout: 45 * time.Second,
    }
    
    // Get a connection pool for the transfer service
    pool := transport.GetServicePool("transfer", poolConfig)
    
    // Create an HTTP client with the connection pool
    httpClient := pool.GetClient()
    
    // Create an SDK configuration with the custom HTTP client
    config := pkg.NewConfigFromEnvironment().
        WithHTTPClient(httpClient)
    
    // Create a transfer client that uses the connection pool
    transferClient := config.NewTransferClient("your-access-token")
    
    // Make requests - connections are pooled according to your settings
    // ...
}
```

### Connection Pool Manager

For applications that need to manage multiple connection pools across different services:

```go
// Create a connection pool manager
manager := transport.NewConnectionPoolManager(nil)

// Get connection pools for different services
transferPool := manager.GetPool("transfer", nil)
authPool := manager.GetPool("auth", nil)

// Get clients using the pools
transferClient := transferPool.GetClient()
authClient := authPool.GetClient()

// Get stats on all pools
stats := manager.GetAllStats()
fmt.Printf("Transfer pool stats: %+v\n", stats["transfer"])
fmt.Printf("Auth pool stats: %+v\n", stats["auth"])

// Close idle connections across all pools
manager.CloseAllIdleConnections()
```

## Monitoring Connection Pool Usage

You can monitor the usage of the connection pool to understand its behavior and fine-tune the configuration:

```go
// Get a connection pool
pool := transport.GetServicePool("transfer", nil)

// Get and print statistics
stats := pool.GetStats()
fmt.Printf("Connection pool stats:\n")
fmt.Printf("  Active hosts:         %d\n", stats.ActiveHosts)
fmt.Printf("  Total active:         %d\n", stats.TotalActive)
fmt.Printf("  Max idle per host:    %d\n", stats.Config.MaxIdleConnsPerHost)
fmt.Printf("  Max connections/host: %d\n", stats.Config.MaxConnsPerHost)
fmt.Printf("  Active connections by host:\n")
for host, count := range stats.ActiveByHost {
    fmt.Printf("    %s: %d\n", host, count)
}
```

## Connection Pool Configuration Options

Here's a detailed explanation of the available configuration options:

| Option | Description | Default |
|--------|-------------|---------|
| `MaxIdleConnsPerHost` | Maximum number of idle connections to keep per host | CPU count × 2 |
| `MaxIdleConns` | Maximum number of idle connections across all hosts | 100 |
| `MaxConnsPerHost` | Limit on the total number of connections per host | CPU count × 4 |
| `IdleConnTimeout` | How long an idle connection will remain idle before being closed | 90 seconds |
| `DisableKeepAlives` | Disables HTTP keep-alives (not recommended) | false |
| `ResponseHeaderTimeout` | Time to wait for a server's response headers | 30 seconds |
| `ExpectContinueTimeout` | Time to wait for a 100-continue response | 5 seconds |
| `TLSHandshakeTimeout` | Maximum time for a TLS handshake | 10 seconds |
| `TLSClientConfig` | Custom TLS configuration | nil |

## When to Customize Connection Pooling

You might want to customize connection pooling settings in these scenarios:

1. **High-volume requests**: Increase `MaxIdleConnsPerHost` and `MaxConnsPerHost` for better throughput
2. **Many distinct hosts**: Increase `MaxIdleConns` when connecting to many different endpoints
3. **Long-lived applications**: Decrease `IdleConnTimeout` to free up resources sooner
4. **Custom TLS requirements**: Set `TLSClientConfig` for specific certificate validation needs
5. **Slow networks**: Increase `ResponseHeaderTimeout` and `TLSHandshakeTimeout`

## Performance Considerations

Connection pooling can significantly improve performance:

- **Reduced latency**: Eliminates TCP and TLS handshake overhead for subsequent requests
- **Improved throughput**: Allows more efficient use of connections with HTTP/2 multiplexing
- **Reduced resource usage**: Manages the number of connections to prevent resource exhaustion
- **Better error handling**: Detects and removes stale connections automatically

## Implementation Details

The connection pooling system is built on top of Go's standard `http.Transport` and adds:

1. Connection tracking by host
2. Dynamic pool sizing based on system resources
3. Statistics gathering for monitoring
4. Centralized management of multiple pools
5. Automatic HTTP/2 support where available

## Related Features

- [Rate Limiting Guide](doc/rate-limiting.md) - Works well with connection pooling for controlled API usage
- [Logging and Tracing Guide](doc/logging-and-tracing.md) - Can be used to monitor connection pool activity
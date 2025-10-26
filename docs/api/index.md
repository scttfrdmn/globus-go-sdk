# API Reference

Complete API reference for all Globus SDK services.

## Service Clients

The SDK provides clients for these Globus services:

### [Auth Service](auth.md)

Authentication and identity management.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"

client := auth.NewClient(authorizer)
```

### [Transfer Service](transfer.md)

File and directory transfer operations.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"

client := transfer.NewClient(authorizer)
```

### [Search Service](search.md)

Search index management and queries.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/search"

client := search.NewClient(authorizer)
```

### [Groups Service](groups.md)

Group membership and policy management.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/groups"

client := groups.NewClient(authorizer)
```

### [Flows Service](flows.md)

Workflow automation and management.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/flows"

client := flows.NewClient(authorizer)
```

### [Timers Service](timers.md)

Scheduled task execution.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/timers"

client := timers.NewClient(authorizer)
```

### [Compute Service](compute.md)

Distributed computing operations.

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"

client := compute.NewClient(authorizer)
```

## Common Patterns

### Creating Clients

All clients follow the same pattern:

```go
authorizer := authorizers.NewAccessTokenAuthorizer(token)
client := service.NewClient(authorizer)
```

### Method Signatures

All service methods follow this pattern:

```go
result, err := client.MethodName(ctx context.Context, options *MethodOptions)
```

### Error Handling

See [Error Handling](errors.md) for details on SDK error types.

### Pagination

Methods that return lists support pagination:

```go
options := &service.ListOptions{
    Limit:  100,
    Offset: 0,
}

result, err := client.List(ctx, options)
```

## Go Package Documentation

Complete Go API documentation is available at:

**[pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3)**

## See Also

- [Quick Start](../getting-started/quickstart.md)
- [Common Patterns](../guides/common-patterns.md)
- [Examples](../examples/transfer.md)

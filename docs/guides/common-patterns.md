# Common Patterns

Best practices and common patterns for using the Globus Go SDK.

## Client Initialization

### Reuse Clients

Create clients once and reuse them:

```go
// Good: Create once
var transferClient *transfer.Client

func init() {
    authorizer := authorizers.NewAccessTokenAuthorizer(token)
    transferClient = transfer.NewClient(authorizer)
}

func doTransfer() error {
    return transferClient.SubmitTransfer(ctx, submission)
}
```

### Don't Create Per-Request

```go
// Bad: Creates new client every time
func doTransfer() error {
    client := transfer.NewClient(authorizer)
    return client.SubmitTransfer(ctx, submission)
}
```

## Context Management

### Always Use Context

```go
ctx := context.Background()
result, err := client.Method(ctx, options)
```

### Use Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.Method(ctx, options)
```

### Propagate Context

```go
func ProcessTransfer(ctx context.Context, taskID string) error {
    task, err := transferClient.GetTask(ctx, taskID)
    if err != nil {
        return err
    }

    // Context automatically propagated
    return handleTask(ctx, task)
}
```

## Error Handling

### Check Errors

```go
result, err := client.Method(ctx, options)
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Type Assertions

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"

task, err := client.GetTask(ctx, taskID)
if err != nil {
    var apiErr *errors.APIError
    if errors.As(err, &apiErr) {
        if apiErr.Code == "NotFound" {
            return ErrTaskNotFound
        }
    }
    return err
}
```

## Pagination

### Iterate All Pages

```go
offset := 0
limit := 100

for {
    tasks, err := client.TaskList(ctx, &transfer.TaskListOptions{
        Limit:  limit,
        Offset: offset,
    })
    if err != nil {
        return err
    }

    for _, task := range tasks.Data {
        // Process task
    }

    if len(tasks.Data) < limit {
        break
    }

    offset += limit
}
```

## Concurrency

### Concurrent Requests

```go
import "golang.org/x/sync/errgroup"

func ProcessMultipleEndpoints(ctx context.Context, endpointIDs []string) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, id := range endpointIDs {
        id := id // Capture variable
        g.Go(func() error {
            _, err := client.EndpointDetails(ctx, id)
            return err
        })
    }

    return g.Wait()
}
```

### Rate Limiting

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(10), 1) // 10 requests per second

for _, item := range items {
    if err := limiter.Wait(ctx); err != nil {
        return err
    }

    _, err := client.Process(ctx, item)
    if err != nil {
        return err
    }
}
```

## Testing

### Use Interfaces

```go
type TransferClient interface {
    SubmitTransfer(ctx context.Context, submission *transfer.TransferSubmission) (*transfer.Task, error)
}

type MyService struct {
    client TransferClient
}

// Easy to mock in tests
type MockTransferClient struct{}

func (m *MockTransferClient) SubmitTransfer(ctx context.Context, submission *transfer.TransferSubmission) (*transfer.Task, error) {
    return &transfer.Task{TaskID: "test-task"}, nil
}
```

## See Also

- [Error Handling Guide](error-handling.md)
- [Rate Limiting Guide](rate-limiting.md)
- [Testing Guide](testing.md)

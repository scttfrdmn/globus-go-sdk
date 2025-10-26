# Transfer Service

File and directory transfer operations.

## Package

```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
```

## Create Client

```go
client := transfer.NewClient(authorizer)
```

## Key Methods

### EndpointList

List endpoints:

```go
endpoints, err := client.EndpointList(ctx, &transfer.EndpointListOptions{
    Limit: 10,
})
```

### ListDirectory

List directory contents:

```go
items, err := client.ListDirectory(ctx, endpointID, "/path/to/dir", nil)
```

### SubmitTransfer

Submit a transfer task:

```go
submission := &transfer.TransferSubmission{
    SourceEndpoint:      sourceID,
    DestinationEndpoint: destID,
    Label:               "My transfer",
    Data: []transfer.TransferItem{
        {
            SourcePath:      "/source/file.txt",
            DestinationPath: "/dest/file.txt",
        },
    },
}

task, err := client.SubmitTransfer(ctx, submission)
```

### GetTask

Get transfer task status:

```go
task, err := client.GetTask(ctx, taskID)
fmt.Printf("Status: %s\n", task.Status)
```

### TaskList

List recent tasks:

```go
tasks, err := client.TaskList(ctx, &transfer.TaskListOptions{
    Limit:        100,
    FilterStatus: "ACTIVE",
})
```

## See Also

- [Transfer Examples](../examples/transfer.md)
- [pkg.go.dev API Docs](https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer)

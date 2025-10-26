# Transfer Examples

Complete examples for file transfer operations.

## Basic File Transfer

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func main() {
    // Setup
    token := os.Getenv("GLOBUS_ACCESS_TOKEN")
    authorizer := authorizers.NewAccessTokenAuthorizer(token)
    client := transfer.NewClient(authorizer)
    ctx := context.Background()

    // Submit transfer
    submission := &transfer.TransferSubmission{
        SourceEndpoint:      "SOURCE_ENDPOINT_ID",
        DestinationEndpoint: "DEST_ENDPOINT_ID",
        Label:               "Example transfer",
        Data: []transfer.TransferItem{
            {
                SourcePath:      "/~/source/file.txt",
                DestinationPath: "/~/dest/file.txt",
            },
        },
    }

    task, err := client.SubmitTransfer(ctx, submission)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Transfer submitted: %s\n", task.TaskID)
}
```

## Directory Transfer

```go
submission := &transfer.TransferSubmission{
    SourceEndpoint:      sourceID,
    DestinationEndpoint: destID,
    Label:               "Directory backup",
    Data: []transfer.TransferItem{
        {
            SourcePath:      "/~/data/",
            DestinationPath: "/~/backup/data/",
            Recursive:       true,
        },
    },
    SyncLevel: "mtime", // Only transfer changed files
}

task, err := client.SubmitTransfer(ctx, submission)
```

## Monitor Transfer

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/transfer"
)

func MonitorTransfer(client *transfer.Client, taskID string) error {
    ctx := context.Background()

    for {
        task, err := client.GetTask(ctx, taskID)
        if err != nil {
            return err
        }

        fmt.Printf("Status: %s - %d/%d bytes\n",
            task.Status,
            task.BytesTransferred,
            task.BytesChecksummed,
        )

        if task.Status == "SUCCEEDED" {
            fmt.Println("Transfer completed successfully!")
            return nil
        }

        if task.Status == "FAILED" {
            return fmt.Errorf("transfer failed")
        }

        time.Sleep(5 * time.Second)
    }
}
```

## List Directory Contents

```go
func ListDirectory(client *transfer.Client, endpointID, path string) error {
    ctx := context.Background()

    items, err := client.ListDirectory(ctx, endpointID, path, nil)
    if err != nil {
        return err
    }

    for _, item := range items.Data {
        if item.Type == "dir" {
            fmt.Printf("[DIR]  %s\n", item.Name)
        } else {
            fmt.Printf("[FILE] %s (%d bytes)\n", item.Name, item.Size)
        }
    }

    return nil
}
```

## See Also

- [Transfer API Reference](../api/transfer.md)
- [Common Patterns](../guides/common-patterns.md)

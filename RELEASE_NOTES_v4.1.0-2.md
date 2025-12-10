<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4.1.0-2

**Release Date:** October 25, 2025
**Python SDK Parity:** v4.1.0 (Complete)
**Module Path:** `github.com/scttfrdmn/globus-go-sdk/v4`

This release completes the v4 SDK implementation with **all Globus services** now available!

## 🎉 What's New

### Complete Service Coverage

v4.1.0-2 adds five major services, bringing the SDK to **100% Python SDK v4.1.0 parity**:

#### ✅ Transfer Service
Complete file transfer operations:
- Endpoint management (`GetEndpoint`, `ListEndpoints`)
- Transfer operations (`SubmitTransfer`, `GetTask`, `CancelTask`)
- Delete operations (`SubmitDelete`)
- Directory operations (`ListDirectory`, `MakeDirectory`, `Rename`)

```go
transferClient, _ := transfer.NewClient(ctx, config)

// Submit a transfer
transfer := &transfer.Transfer{
    SourceEndpoint: "source-id",
    DestinationEndpoint: "dest-id",
    Items: []transfer.TransferItem{
        {SourcePath: "/source/file.txt", DestinationPath: "/dest/file.txt"},
    },
}
response, _ := transferClient.SubmitTransfer(ctx, transfer)
```

#### ✅ Search Service
Complete search index operations:
- Index management (`CreateIndex`, `UpdateIndex`, `DeleteIndex`)
- Search queries (`Search`)
- Document ingestion (`IngestEntry`, `IngestBatch`)
- Role management (`AddRole`, `RemoveRole`)

```go
searchClient, _ := search.NewClient(ctx, config)

// Perform a search
query := &search.SearchQuery{
    Q: "genome data",
    Limit: 10,
}
results, _ := searchClient.Search(ctx, "index-id", query)
```

#### ✅ Flows Service
Flow management and execution:
- Flow operations (`GetFlow`, `ListFlows`)
- Flow execution (`RunFlow`, `GetRun`, `CancelRun`)
- Run management (`ListRuns`)

```go
flowsClient, _ := flows.NewClient(ctx, config)

// Run a flow
input := &flows.FlowInput{
    Input: map[string]interface{}{
        "source": "/data/input",
        "dest": "/data/output",
    },
}
run, _ := flowsClient.RunFlow(ctx, "flow-id", input)
```

#### ✅ Timers Service
Timer scheduling with flow integration:
- Timer management (`CreateTimer`, `UpdateTimer`, `DeleteTimer`)
- Timer control (`PauseTimer`, `ResumeTimer`)
- Flow timer helpers (`CreateFlowTimer`, `CreateOnceTimer`, `CreateRecurringTimer`)

```go
timersClient, _ := timers.NewClient(ctx, config)

// Create a flow timer
timer, _ := timersClient.CreateFlowTimer(
    ctx,
    "Daily Backup",
    schedule,
    "flow-id",
    "flow-scope",
    flowInput,
)
```

#### ✅ Compute Service
Compute endpoint and function execution:
- Endpoint operations (`GetEndpoint`, `ListEndpoints`)
- Function execution (`SubmitFunction`, `GetFunction`, `CancelFunction`)
- Function management (`ListFunctions`)

```go
computeClient, _ := compute.NewClient(ctx, config)

// Submit a function
submission := &compute.FunctionSubmission{
    FunctionID: "func-id",
    Args: []interface{}{"arg1", "arg2"},
}
run, _ := computeClient.SubmitFunction(ctx, "endpoint-id", submission)
```

## 📊 Service Status

| Service | v4.1.0-1 | v4.1.0-2 | Status |
|---------|----------|----------|--------|
| Auth | ✅ | ✅ | Complete |
| Groups | ✅ | ✅ | Complete |
| Transfer | ❌ | ✅ | **NEW** |
| Search | ❌ | ✅ | **NEW** |
| Flows | ❌ | ✅ | **NEW** |
| Timers | ❌ | ✅ | **NEW** |
| Compute | ❌ | ✅ | **NEW** |

## 🎯 Python SDK Parity: 100%

The Go SDK v4.1.0-2 now has **complete feature parity** with Python SDK v4.1.0:

| Feature | Python SDK | Go SDK v4.1.0-2 |
|---------|------------|-----------------|
| Context-first design | ✅ | ✅ |
| Explicit scopes | ✅ | ✅ |
| Enhanced errors | ✅ | ✅ |
| Unified config | ✅ | ✅ |
| Auth service | ✅ | ✅ |
| Groups service | ✅ | ✅ |
| Transfer service | ✅ | ✅ |
| Search service | ✅ | ✅ |
| Flows service | ✅ | ✅ |
| Timers service | ✅ | ✅ |
| Compute service | ✅ | ✅ |

## 📦 Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@v4.1.0-2
```

## 🚀 Quick Start with All Services

```go
import (
    "context"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/search"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/flows"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/timers"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/compute"
)

func main() {
    ctx := context.Background()

    // All clients use the same v4 pattern
    config := &core.Config{
        AccessToken: token,
        Scopes: []string{core.Scopes.TransferAll},
    }

    transferClient, _ := transfer.NewClient(ctx, config)
    searchClient, _ := search.NewClient(ctx, config)
    flowsClient, _ := flows.NewClient(ctx, config)
    // ... and so on
}
```

## 💡 Key Features

All v4 services share these improvements:

- ✅ **Context-first**: Every method accepts `context.Context` as first parameter
- ✅ **Explicit scopes**: Security-focused scope specification required
- ✅ **Enhanced errors**: Structured `APIError` with request IDs
- ✅ **Type safety**: Strongly-typed models for all operations
- ✅ **Retry logic**: Automatic retries with exponential backoff
- ✅ **Well-known scopes**: Pre-defined constants in `core.Scopes`

## 📚 Documentation

- **API Reference**: https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4
- **Migration Guide**: [V4_MIGRATION_GUIDE.md](V4_MIGRATION_GUIDE.md)
- **Implementation Plan**: [V4_IMPLEMENTATION_PLAN.md](V4_IMPLEMENTATION_PLAN.md)
- **Globus Platform**: https://docs.globus.org/

## 🔄 Upgrading from v4.1.0-1

If you're already using v4.1.0-1 (Auth + Groups only), simply update your import:

```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@v4.1.0-2
```

No code changes needed - all existing v4.1.0-1 code continues to work!

## 🎉 Status

**The v4 SDK is now feature-complete!**

This release brings the Go SDK to full parity with the Python SDK v4.1.0. All Globus services are now available with the improved v4 patterns.

## 🙏 Acknowledgments

This release completes the synchronization with the excellent [Globus Python SDK v4.1.0](https://github.com/globus/globus-sdk-python). Thank you to the Globus team for maintaining such a comprehensive Python SDK as our reference.

## 📞 Support

- **Documentation**: https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4
- **Issues**: https://github.com/scttfrdmn/globus-go-sdk/issues
- **Discussions**: https://github.com/scttfrdmn/globus-go-sdk/discussions

---

**Full Changelog:** v4.1.0-1...v4.1.0-2

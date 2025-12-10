<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4.2.0-1

**Release Date:** December 9, 2025
**Python SDK Parity:** v4.2.0
**Module Path:** `github.com/scttfrdmn/globus-go-sdk/v4`

This release brings the v4 SDK into sync with Python SDK v4.2.0, adding resource cleanup functionality through Close() methods.

## 🎉 What's New

### Resource Management (Python SDK v4.2.0 Feature)

All v4 clients now implement the `Close()` method for proper resource cleanup, matching the Python SDK v4.2.0 context manager pattern.

#### Core Client

```go
import (
    "context"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/transfer"
)

config := &core.Config{
    AccessToken: "your-token",
    Scopes:      []string{core.Scopes.TransferAll},
}

client, err := transfer.NewClient(context.Background(), config)
if err != nil {
    return err
}
defer client.Close() // Automatically cleanup resources

// Use the client...
```

#### All Service Clients

Close() is now available on all service clients:
- ✅ `auth.Client.Close()`
- ✅ `groups.Client.Close()`
- ✅ `transfer.Client.Close()`
- ✅ `search.Client.Close()`
- ✅ `flows.Client.Close()`
- ✅ `timers.Client.Close()`
- ✅ `compute.Client.Close()`

### Behavior Details

**Internally Created HTTP Clients:**
When you don't provide an HTTPClient in the config, the SDK creates one automatically. In this case, `Close()` will clean up idle connections:

```go
config := &core.Config{
    AccessToken: "token",
    Scopes:      []string{core.Scopes.TransferAll},
    // HTTPClient: nil (not provided)
}

client, _ := transfer.NewClient(ctx, config)
defer client.Close() // Closes idle connections on SDK-created HTTP client
```

**User-Provided HTTP Clients:**
If you provide your own HTTPClient, you remain responsible for managing it. The SDK's `Close()` method will NOT close your client:

```go
httpClient := &http.Client{
    Timeout: 60 * time.Second,
}

config := &core.Config{
    AccessToken: "token",
    Scopes:      []string{core.Scopes.TransferAll},
    HTTPClient:  httpClient, // Your client
}

client, _ := transfer.NewClient(ctx, config)
defer client.Close()     // Does NOT close your httpClient
defer httpClient.CloseIdleConnections() // You manage your own client
```

**Safe for Multiple Calls:**
`Close()` is safe to call multiple times and will not panic.

```go
client.Close()
client.Close() // Safe, no error
```

## 📝 Python SDK v4.2.0 Alignment

This release implements the following Python SDK v4.2.0 features:

### ✅ Implemented
- **Close() methods** - Resource cleanup for all clients (context manager equivalent)
- **Safe resource management** - Automatically closes internally-created HTTP clients
- **User client protection** - Respects user-provided HTTP clients

### ⏳ Future Implementation
The following Python SDK v4.2.0 features are documented for future implementation:

- **TimersClient.AddAppFlowUserScope** - Register required flow scope dependencies
- **GARE Auto-Retry** - Automatic Globus Auth Requirements Error handling
- **Enhanced config options** - Additional retry and timeout configuration

See [PYTHON_SDK_V4.2.0_TRACKING.md](../PYTHON_SDK_V4.2.0_TRACKING.md) for implementation details.

## 🔧 Technical Details

### Internal Changes
- Added `httpClientCreated` tracking to `core.Client`
- Updated `NewClient()` to track HTTP client ownership
- Implemented `Close()` in core and all service clients
- Added comprehensive tests for Close() behavior

### Testing
- ✅ Core client Close() tests
- ✅ Internally-created HTTP client cleanup
- ✅ User-provided HTTP client preservation
- ✅ Multiple Close() call safety

## 📊 Version Information

| Component | Version |
|-----------|---------|
| **Go SDK v4** | 4.2.0-1 |
| **Python SDK Parity** | v4.2.0 |
| **Go Version** | 1.22+ |

## 🔗 Migration from v4.1.0-2

No breaking changes! Simply add `defer client.Close()` after creating clients:

**Before (v4.1.0-2):**
```go
client, err := transfer.NewClient(ctx, config)
if err != nil {
    return err
}
// No cleanup
```

**After (v4.2.0-1):**
```go
client, err := transfer.NewClient(ctx, config)
if err != nil {
    return err
}
defer client.Close() // Add this line
```

## 📚 Documentation

### Close() Method Usage Patterns

#### Pattern 1: Defer immediately after creation
```go
client, err := transfer.NewClient(ctx, config)
if err != nil {
    return err
}
defer client.Close()

// Use client...
```

#### Pattern 2: Explicit cleanup in error handling
```go
client, err := transfer.NewClient(ctx, config)
if err != nil {
    return err
}

result, err := client.DoSomething(ctx)
if err != nil {
    client.Close()
    return err
}
defer client.Close()
```

#### Pattern 3: Long-lived client
```go
type Service struct {
    transferClient *transfer.Client
}

func (s *Service) Start(ctx context.Context) error {
    client, err := transfer.NewClient(ctx, config)
    if err != nil {
        return err
    }
    s.transferClient = client
    return nil
}

func (s *Service) Stop() error {
    if s.transferClient != nil {
        return s.transferClient.Close()
    }
    return nil
}
```

## 🐛 Known Limitations

### v4 SDK Status
The v4 SDK is currently a **minimal implementation** suitable for basic use cases. The following features from v3 are not yet available in v4:

- Connection pooling
- Advanced rate limiting
- Circuit breaker patterns
- Comprehensive retry strategies
- Token storage and automatic refresh
- Recursive transfer support

For production applications requiring these features, we recommend using the **v3 SDK** (v3.65.0-1), which is feature-complete and production-ready.

## 🔜 What's Next

See [PYTHON_SDK_V4.2.0_TRACKING.md](../PYTHON_SDK_V4.2.0_TRACKING.md) for:
- Remaining v4.2.0 features (AddAppFlowUserScope, GARE auto-retry)
- Full v4 infrastructure roadmap
- Implementation timeline and priorities

## 🆚 Version Comparison

| Feature | v3.65.0-1 | v4.1.0-2 | v4.2.0-1 |
|---------|-----------|----------|----------|
| **All Services** | ✅ | ✅ | ✅ |
| **Close() Methods** | ❌ | ❌ | ✅ |
| **Connection Pooling** | ✅ | ❌ | ❌ |
| **Rate Limiting** | ✅ | ❌ | ❌ |
| **Token Storage** | ✅ | ❌ | ❌ |
| **Context-First API** | ❌ | ✅ | ✅ |
| **Explicit Scopes** | ❌ | ✅ | ✅ |
| **Production Ready** | ✅ | ⚠️ | ⚠️ |

## 📦 Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@v4.2.0-1
```

## 🔗 Links

- [Python SDK v4.2.0 Release](https://github.com/globus/globus-sdk-python/releases/tag/4.2.0)
- [Go SDK v3.65.0-1 (Production)](https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v3.65.0-1)
- [v4.2.0 Implementation Tracking](../PYTHON_SDK_V4.2.0_TRACKING.md)

---

**Full Changelog:** v4.1.0-2...v4.2.0-1

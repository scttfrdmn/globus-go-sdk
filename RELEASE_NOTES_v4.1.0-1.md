<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4.1.0-1

**Release Date:** October 25, 2025
**Python SDK Parity:** v4.1.0
**Module Path:** `github.com/scttfrdmn/globus-go-sdk/v4`

**⚠️ BREAKING CHANGES: This is a major version release with breaking changes from v3.x**

This release introduces the Globus Go SDK v4, bringing full compatibility with the upstream Globus Python SDK v4.1.0 and implementing significant improvements based on v4 breaking changes.

## 🚀 What's New in v4

### Context-First Design

All methods now require `context.Context` as the first parameter for proper cancellation and timeout support:

```go
// v3.x (old)
userInfo, err := client.GetUserInfo()

// v4.x (new)
userInfo, err := client.GetUserInfo(ctx)
```

### Explicit Scopes for Security

OAuth2 scopes must now be explicitly specified in the `Config` for improved security:

```go
config := &core.Config{
    AccessToken: token,
    Scopes: []string{
        core.Scopes.AuthOpenID,
        core.Scopes.AuthEmail,
    },
}
```

### Unified Configuration

All clients now use a unified `core.Config` struct instead of the options pattern:

```go
// v3.x (old)
client := auth.NewClient(
    auth.WithAuthorizer(authorizer),
    auth.WithHTTPClient(httpClient),
)

// v4.x (new)
config := &core.Config{
    AccessToken: token,
    Scopes: []string{core.Scopes.AuthOpenID},
    HTTPClient: httpClient,
}
client, err := auth.NewClient(ctx, config)
```

### Enhanced Error Handling

New structured error types with detailed information:

```go
userInfo, err := client.GetUserInfo(ctx)
if err != nil {
    if apiErr, ok := err.(*core.APIError); ok {
        fmt.Printf("HTTP Status: %d\n", apiErr.StatusCode)
        fmt.Printf("Error Code: %s\n", apiErr.Code)
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)

        // Use helper methods
        if apiErr.IsAuthError() {
            // Handle auth errors
        }
    }
}
```

### Well-Known Scopes

Pre-defined scope constants for all Globus services:

```go
// Auth scopes
core.Scopes.AuthOpenID
core.Scopes.AuthEmail
core.Scopes.AuthProfile

// Transfer scopes
core.Scopes.TransferAll

// Groups scopes
core.Scopes.GroupsAll
core.Scopes.GroupsView

// And more...
```

## 📦 Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@v4.1.0-1
```

## 🔄 Side-by-Side Usage with v3

You can use both v3 and v4 in the same project during migration:

```go
import (
    authv3 "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    authv4 "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)
```

## 📚 What's Included

### Core Infrastructure

- **pkg/core/version.go** - v4.0.0 version constants and build info
- **pkg/core/errors.go** - Enhanced error types (APIError, ValidationError, NetworkError)
- **pkg/core/config.go** - Unified Config struct with explicit scopes
- **pkg/core/client.go** - Context-first base client with retry logic

### Service Clients (Initial)

- **Auth Service** (`pkg/services/auth`)
  - User info operations
  - Token introspection and revocation
  - Authorization code exchange
  - Token refresh
  - Project management

- **Groups Service** (`pkg/services/groups`)
  - Group management (create, list, update, delete)
  - Member operations
  - Status filtering

### Documentation & Examples

- **v4/README.md** - Complete v4 documentation with quick start guide
- **v4/examples/basic-usage** - Working example demonstrating v4 patterns
- **V4_MIGRATION_GUIDE.md** - Detailed migration guide from v3 to v4
- **V4_IMPLEMENTATION_PLAN.md** - Complete v4 implementation roadmap

## 🔨 Breaking Changes

1. **Context Required**: All methods now require `context.Context` as first parameter
2. **Explicit Scopes**: Scopes must be explicitly specified in `Config.Scopes`
3. **Import Paths**: Use `/v4` in import paths instead of `/v3`
4. **Client Constructors**: All clients require `Config` struct and return `(client, error)`
5. **Error Types**: New error types replace old error handling patterns
6. **No Options Pattern**: Replaced with struct-based configuration

## 🎯 Implemented Features

### ✅ Core Features
- Context-first design
- Enhanced error handling with APIError
- Explicit scope configuration
- Retry logic with exponential backoff
- Well-known scope constants
- Static and refreshable token providers

### ✅ Service Coverage
- Auth service (complete)
- Groups service (complete)

### 🚧 Coming Soon
- Transfer service
- Search service
- Flows service
- Timers service
- Compute service

## 📖 Example Usage

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    config := &core.Config{
        AccessToken: "your-access-token",
        Scopes: []string{
            core.Scopes.AuthOpenID,
            core.Scopes.AuthEmail,
        },
    }

    client, err := auth.NewClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }

    userInfo, err := client.GetUserInfo(ctx)
    if err != nil {
        if apiErr, ok := err.(*core.APIError); ok {
            log.Printf("API error (HTTP %d): %s", apiErr.StatusCode, apiErr.Message)
            log.Printf("Request ID: %s", apiErr.RequestID)
        } else {
            log.Fatal(err)
        }
        return
    }

    fmt.Printf("User: %s (%s)\n", userInfo.Name, userInfo.Email)
}
```

## 🔗 Migration Resources

- **V4 Migration Guide**: [V4_MIGRATION_GUIDE.md](../V4_MIGRATION_GUIDE.md)
- **V4 Implementation Plan**: [V4_IMPLEMENTATION_PLAN.md](../V4_IMPLEMENTATION_PLAN.md)
- **V4 Quick Start**: [V4_QUICK_START.md](../V4_QUICK_START.md)
- **API Documentation**: https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4

## ⚡ Performance & Reliability

- Automatic retry with exponential backoff
- Configurable retry behavior
- Request timeout support via context
- Connection pooling via custom HTTPClient
- Enhanced error details for debugging

## 🧪 Testing

v4 includes comprehensive unit tests for:
- Core configuration and validation
- Error type handling
- Client initialization
- Context propagation
- Scope validation

## 🔄 Python SDK Synchronization

| Feature | Python SDK v4.x | Go SDK v4.1.0-1 |
|---------|----------------|-----------------|
| Context-first design | ✅ | ✅ |
| Explicit scopes | ✅ | ✅ |
| Enhanced errors | ✅ | ✅ |
| Unified config | ✅ | ✅ |
| Auth service | ✅ | ✅ |
| Groups service | ✅ | ✅ |
| Transfer service | ✅ | 🚧 Coming |
| Other services | ✅ | 🚧 Coming |

## 🎯 Roadmap

See [V4_IMPLEMENTATION_PLAN.md](../V4_IMPLEMENTATION_PLAN.md) for the complete implementation roadmap.

**Target Completion**: End of 2025

**Upcoming in v4.0.0-2**:
- Transfer service implementation
- Search service implementation
- Additional examples

## 🙏 Acknowledgments

This release maintains synchronization with the excellent [Globus Python SDK](https://github.com/globus/globus-sdk-python) maintained by the Globus team. The v4 breaking changes are designed to improve API consistency and developer experience across both SDKs.

## 📞 Support

- **Documentation**: https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4
- **Issues**: https://github.com/scttfrdmn/globus-go-sdk/issues
- **Discussions**: https://github.com/scttfrdmn/globus-go-sdk/discussions
- **Globus Platform**: https://www.globus.org
- **Globus Support**: support@globus.org

---

**Full Changelog:** v3.65.0-1...v4.1.0-1

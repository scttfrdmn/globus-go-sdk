<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4 Quick Start Guide

**Status:** Planning Phase
**Target Release:** December 2025
**Current Version:** v4.0.0-1 (to be released)

## For Developers Starting v4 Work

This guide is for contributors working on implementing the v4 SDK.

---

## Getting Started

### 1. Review the Plan

Read these documents in order:

1. **V4_IMPLEMENTATION_PLAN.md** - Complete implementation plan
2. **V4_MIGRATION_GUIDE.md** - User migration guide (helps understand what's changing)
3. This document - Quick reference

### 2. Set Up Development Environment

```bash
cd /Users/scttfrdmn/src/globus-go-sdk

# Create v4 directory structure
mkdir -p v4/pkg/{core,services}
mkdir -p v4/pkg/services/{auth,transfer,groups,search,flows,compute,timers,tokens}
mkdir -p v4/cmd/examples

# Initialize v4 module
cd v4
go mod init github.com/scttfrdmn/globus-go-sdk/v4
```

### 3. Implementation Order

Follow this order for best results:

```
Phase 1: Core Infrastructure
├── v4/pkg/core/config.go
├── v4/pkg/core/transport/
├── v4/pkg/core/retry/
├── v4/pkg/core/errors/
└── v4/pkg/core/scopes/

Phase 2: Auth Service
├── v4/pkg/services/auth/client.go
├── v4/pkg/services/auth/authorizers/
└── v4/pkg/services/auth/*_test.go

Phase 3: Data Services
├── v4/pkg/services/transfer/
├── v4/pkg/services/search/
└── v4/pkg/services/groups/

Phase 4: Automation Services
├── v4/pkg/services/flows/
├── v4/pkg/services/compute/
└── v4/pkg/services/timers/

Phase 5: Examples & Docs
├── v4/cmd/examples/
└── Documentation updates
```

---

## Key Design Decisions

### 1. Module Path: `/v4`

**Decision:** Use `/v4` subdirectory with `github.com/scttfrdmn/globus-go-sdk/v4` module path

**Rationale:**
- Go standard for major version modules
- Allows v3 and v4 to coexist
- Clear versioning in import paths

### 2. Explicit Scopes Required

**Decision:** All authorizers require explicit scope specification

**Example:**
```go
// v4 - Explicit scopes REQUIRED
authorizer := auth.NewClientCredentialsAuthorizer(
    clientID,
    clientSecret,
    []string{"urn:globus:auth:scope:transfer.api.globus.org:all"},
)
```

**Rationale:**
- Matches Python SDK v4.x
- Improves security (no hidden defaults)
- Forces developers to think about permissions

### 3. Unified Config Pattern

**Decision:** All clients accept a `core.Config` object

**Example:**
```go
config := core.NewConfig(core.WithAuthorizer(authorizer))

authClient := auth.NewClient(config)
transferClient := transfer.NewClient(config)
groupsClient := groups.NewClient(config)
```

**Rationale:**
- Consistent API across services
- Easy to share configuration
- Cleaner than per-service options

### 4. Context Everywhere

**Decision:** All methods require `context.Context` as first parameter

**Example:**
```go
group, err := client.GetGroup(ctx, groupID)
```

**Rationale:**
- Go best practice
- Enables timeouts, cancellation, tracing
- Consistent with modern Go libraries

### 5. Enhanced Errors

**Decision:** All API errors implement `core.APIError` interface

**Example:**
```go
err := client.DoSomething(ctx)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        log.Printf("[%s] %s", apiErr.Code, apiErr.Message)
        log.Printf("Request ID: %s", apiErr.RequestID)
    }
}
```

**Rationale:**
- Better error introspection
- Matches Python SDK v4.x error enhancements
- Easier debugging

---

## Code Templates

### Core Config (v4/pkg/core/config.go)

```go
package core

type Config struct {
    authorizer  Authorizer
    transport   Transport
    retryConfig RetryConfig
    logger      Logger
}

func NewConfig(opts ...ConfigOption) *Config {
    config := &Config{
        transport:   DefaultTransport(),
        retryConfig: DefaultRetryConfig(),
    }
    for _, opt := range opts {
        opt(config)
    }
    return config
}

func WithAuthorizer(auth Authorizer) ConfigOption {
    return func(c *Config) { c.authorizer = auth }
}

func WithTransport(t Transport) ConfigOption {
    return func(c *Config) { c.transport = t }
}

func WithRetryConfig(rc RetryConfig) ConfigOption {
    return func(c *Config) { c.retryConfig = rc }
}
```

### Service Client Template

```go
package auth

import "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"

type Client struct {
    config *core.Config
}

func NewClient(config *core.Config) *Client {
    return &Client{config: config}
}

// Example method
func (c *Client) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
    // Implementation
}
```

### Error Handling Template

```go
package core

type APIError struct {
    StatusCode int
    Code       string
    Message    string
    RequestID  string
    Notes      []string
    Detail     map[string]interface{}
}

func (e *APIError) Error() string {
    return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
}
```

---

## Testing Strategy

### Unit Tests

Every package needs:
- Client construction tests
- Method call tests
- Error handling tests
- Mock server tests

```go
func TestClientConstruction(t *testing.T) {
    config := core.NewConfig()
    client := auth.NewClient(config)
    assert.NotNil(t, client)
}
```

### Integration Tests

Test against real API:
```go
func TestIntegrationAuth(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx := context.Background()
    // Use test credentials
    // Call real API
    // Verify response
}
```

### Compatibility Tests

Ensure v3 and v4 can coexist:
```go
import (
    v3auth "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    v4auth "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

func TestV3V4Coexist(t *testing.T) {
    // Create both clients
    // Ensure no conflicts
}
```

---

## Documentation Standards

### Package Documentation

Every package needs `doc.go`:

```go
/*
Package auth provides v4 client for Globus Auth service.

# STABILITY: STABLE

This is the v4 major version with breaking changes from v3.

## Breaking Changes from v3

- Explicit scopes required for all authorizers
- Context required for all methods
- Unified config pattern

## Basic Usage

    config := core.NewConfig(core.WithAuthorizer(authorizer))
    client := auth.NewClient(config)

    userInfo, err := client.GetUserInfo(ctx, token)

See V4_MIGRATION_GUIDE.md for complete migration instructions.
*/
package auth
```

### Method Documentation

```go
// GetUserInfo retrieves user information for the given access token.
//
// This method requires an active access token and returns details about
// the authenticated user.
//
// Parameters:
//   - ctx: Context for the request (required in v4)
//   - token: Valid Globus access token
//
// Returns:
//   - UserInfo containing user details
//   - error if the request fails
//
// Example:
//
//	userInfo, err := client.GetUserInfo(ctx, accessToken)
//	if err != nil {
//	    var apiErr *core.APIError
//	    if errors.As(err, &apiErr) {
//	        log.Printf("API error: %s", apiErr.Message)
//	    }
//	}
func (c *Client) GetUserInfo(ctx context.Context, token string) (*UserInfo, error) {
    // Implementation
}
```

---

## Checklist for Each Service

When implementing a service, complete these steps:

- [ ] Create service directory structure
- [ ] Implement `Client` struct
- [ ] Implement `NewClient(config)` constructor
- [ ] Port all v3 methods with context parameter
- [ ] Update error handling to use `core.APIError`
- [ ] Write unit tests (100% coverage goal)
- [ ] Write integration tests
- [ ] Create `doc.go` with examples
- [ ] Update CHANGELOG.md
- [ ] Create example in `cmd/examples/`

---

## Common Pitfalls

### Pitfall 1: Forgetting Context

```go
// ❌ Wrong
func (c *Client) GetGroup(groupID string) (*Group, error)

// ✅ Right
func (c *Client) GetGroup(ctx context.Context, groupID string) (*Group, error)
```

### Pitfall 2: Not Using core.Config

```go
// ❌ Wrong (v3 pattern)
func NewClient(opts ...Option) (*Client, error)

// ✅ Right (v4 pattern)
func NewClient(config *core.Config) *Client
```

### Pitfall 3: Generic Errors

```go
// ❌ Wrong
return nil, fmt.Errorf("request failed")

// ✅ Right
return nil, &core.APIError{
    StatusCode: resp.StatusCode,
    Code:       errorCode,
    Message:    errorMessage,
    RequestID:  resp.Header.Get("X-Request-ID"),
}
```

---

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/v4-auth-service
```

### 2. Implement Service

Follow the implementation checklist above.

### 3. Test Locally

```bash
cd v4
go test ./pkg/services/auth/...
go test -race ./...
go test -cover ./...
```

### 4. Create Pull Request

Include:
- Description of what was implemented
- Link to Python SDK equivalent
- Test coverage report
- Example usage

### 5. Code Review

Address feedback and update.

### 6. Merge

Once approved, merge to main v4 branch.

---

## Resources

- **Python SDK v4 Docs:** https://globus-sdk-python.readthedocs.io/en/stable/
- **Python SDK GitHub:** https://github.com/globus/globus-sdk-python
- **Globus API Docs:** https://docs.globus.org/api/
- **Go Module Reference:** https://go.dev/ref/mod

---

## Questions?

- Check `V4_IMPLEMENTATION_PLAN.md` for detailed planning
- Check `V4_MIGRATION_GUIDE.md` for user-facing changes
- Open an issue for discussion

---

**Quick Start Guide Version:** 1.0
**Last Updated:** October 25, 2025

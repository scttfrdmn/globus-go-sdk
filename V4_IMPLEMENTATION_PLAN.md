<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v4.0.0-1 Implementation Plan

**Target Date:** End of 2025
**Python SDK Parity:** v4.1.0
**Module Path:** `github.com/scttfrdmn/globus-go-sdk/v4`

This document outlines the complete implementation plan for migrating the Globus Go SDK to v4.x, including module structure, breaking changes, and migration strategy.

---

## Table of Contents

1. [Module Structure Strategy](#module-structure-strategy)
2. [Python SDK v4.x Breaking Changes](#python-sdk-v4x-breaking-changes)
3. [Go SDK v4.x Breaking Changes](#go-sdk-v4x-breaking-changes)
4. [Implementation Phases](#implementation-phases)
5. [Migration Guide for Users](#migration-guide-for-users)
6. [Testing Strategy](#testing-strategy)
7. [Timeline and Milestones](#timeline-and-milestones)

---

## Module Structure Strategy

### Dual-Module Approach

The Go SDK will maintain **both v3 and v4** simultaneously using Go's module versioning:

```
github.com/scttfrdmn/globus-go-sdk/
├── v3/          # Existing v3.x releases (maintenance mode)
│   └── pkg/
│       ├── core/
│       ├── services/
│       └── ...
└── v4/          # New v4.x releases (active development)
    └── pkg/
        ├── core/
        ├── services/
        └── ...
```

### Module Path Structure

**v3 Module:**
```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
```

**v4 Module (New):**
```go
import "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
```

### Benefits of Dual-Module Approach

1. **Side-by-Side Installation:** Users can use both v3 and v4 in the same project during migration
2. **Gradual Migration:** Teams can migrate one service at a time
3. **No Breaking Surprises:** v3 users are not forced to upgrade
4. **Clear Versioning:** Import paths make version explicit

---

## Python SDK v4.x Breaking Changes

### v4.0.0 Breaking Changes (October 8, 2024)

#### 1. Transport & Retry Refactoring
**Python Change:**
```python
# Old (v3.x)
client = TransferClient(transport_params={...})

# New (v4.x)
transport = RequestsTransport(...)
retry_config = RetryConfig(...)
client = TransferClient(transport=transport, retry_config=retry_config)
```

**Go SDK Impact:** Medium
- Current Go SDK already has separate `transport` package (`pkg/core/transport/`)
- Need to refactor client constructors to accept separate transport and retry config

**Go SDK v4 Approach:**
```go
// v3.x (current)
client := auth.NewClient(
    auth.WithAuthorizer(authorizer),
    auth.WithHTTPClient(httpClient),
)

// v4.x (proposed)
transport := core.NewTransport(transportConfig)
retryConfig := core.NewRetryConfig(retryOptions)
client := auth.NewClient(
    auth.WithTransport(transport),
    auth.WithRetryConfig(retryConfig),
    auth.WithAuthorizer(authorizer),
)
```

#### 2. Removed Default Scopes
**Python Change:**
- SDK no longer auto-sets default scopes for client credentials
- Users MUST explicitly specify scopes

**Go SDK Impact:** High
- Breaking change for users relying on default scopes
- Improves security by forcing explicit scope declaration

**Go SDK v4 Approach:**
```go
// v3.x (auto-defaults)
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret)
// Scopes auto-set to some defaults

// v4.x (explicit required)
authorizer := auth.NewClientCredentialsAuthorizer(
    clientID,
    clientSecret,
    []string{
        "urn:globus:auth:scope:transfer.api.globus.org:all",
        "urn:globus:auth:scope:search.api.globus.org:all",
    }, // Explicit scopes REQUIRED
)
```

#### 3. Scope Handling Changes
**Python Change:**
- Removed `scopes_to_str()`, replaced with `ScopeParser.serialize()`
- Removed `scopes_to_scope_list()`

**Go SDK Impact:** Low
- Go SDK has different scope handling patterns
- Consider adding unified scope serialization

**Go SDK v4 Approach:**
```go
// New scope utilities
type ScopeParser struct {}

func (sp *ScopeParser) Serialize(scopes []string) string {
    return strings.Join(scopes, " ")
}

func (sp *ScopeParser) Parse(scopeStr string) []string {
    return strings.Split(scopeStr, " ")
}
```

#### 4. Error Handling Enhancement
**Python Change:**
- API error notes populate `__notes__` (Python 3.11+)
- Richer error context

**Go SDK Impact:** Medium
- Go doesn't have `__notes__`, but can enhance error wrapping

**Go SDK v4 Approach:**
```go
// Enhanced error types with more context
type APIError struct {
    StatusCode int
    Code       string
    Message    string
    RequestID  string
    Notes      []string  // NEW: Additional context
    Detail     map[string]interface{}
}

func (e *APIError) Error() string {
    msg := fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
    if len(e.Notes) > 0 {
        msg += "\nNotes:\n  - " + strings.Join(e.Notes, "\n  - ")
    }
    return msg
}
```

#### 5. Type Safety Improvements
**Python Change:**
- Passing non-Scope types raises `TypeError`
- Stricter type checking

**Go SDK Impact:** Low
- Go already has compile-time type safety
- Consider adding runtime validation helpers

---

### v4.0.1 Bug Fix (October 13, 2024)

**Python Fix:** Fixed route for `TransferClient.set_subscription_admin_verified()`

**Go SDK Status:** ✅ Not applicable
- Go SDK has correct route in Groups service (not Transfer)
- No changes needed

---

### v4.1.0 New Feature (October 23, 2024)

**Feature:** Flow Authentication Policy Support

**Python Change:**
```python
client.create_flow(
    title="My Flow",
    definition={...},
    authentication_policy="policy-uuid"  # NEW
)
```

**Go SDK v4 Approach:**
```go
// pkg/services/flows/models.go
type FlowCreateRequest struct {
    Title                string                 `json:"title"`
    Description          string                 `json:"description,omitempty"`
    Definition           map[string]interface{} `json:"definition"`
    InputSchema          map[string]interface{} `json:"input_schema,omitempty"`
    Keywords             []string               `json:"keywords,omitempty"`
    Public               bool                   `json:"public,omitempty"`
    Managed              bool                   `json:"managed,omitempty"`
    AdminOnly            bool                   `json:"admin_only,omitempty"`
    RunsRequired         bool                   `json:"runs_required,omitempty"`
    RunAsApprover        bool                   `json:"run_as_approver,omitempty"`
    AuthenticationPolicy string                 `json:"authentication_policy,omitempty"` // NEW in v4.1.0
}

type FlowUpdateRequest struct {
    // ... existing fields ...
    AuthenticationPolicy *string `json:"authentication_policy,omitempty"` // NEW (pointer for PATCH)
}
```

**Stability:** BETA initially (service support pending), then STABLE

---

## Go SDK v4.x Breaking Changes

Beyond Python SDK parity, the Go SDK v4.x will introduce additional Go-specific improvements:

### 1. **Simplified Client Construction** (BREAKING)

**v3.x (Current):**
```go
// Different patterns per service
authClient, err := auth.NewClient(
    auth.WithAuthorizer(authorizer),
)

transferClient, err := transfer.NewClient(
    transfer.WithAccessToken(token),
)

groupsClient, err := groups.NewClient(
    groups.WithAuthorizer(authorizer),
)
```

**v4.x (Proposed):**
```go
// Unified pattern across all services
config := core.NewConfig(
    core.WithAuthorizer(authorizer),
    core.WithTransport(transport),
    core.WithRetryConfig(retryConfig),
)

authClient := auth.NewClient(config)
transferClient := transfer.NewClient(config)
groupsClient := groups.NewClient(config)
```

### 2. **Context-First Parameters** (BREAKING)

**v3.x:**
```go
// Some methods have context, some don't
client.GetGroup(groupID) // No context
client.ListGroups(ctx, options) // Has context
```

**v4.x:**
```go
// ALL methods require context
client.GetGroup(ctx, groupID) // Context required
client.ListGroups(ctx, options) // Context required
```

### 3. **Consistent Error Types** (BREAKING)

**v3.x:**
```go
// Mixed error types
err := client.DoSomething()
if err != nil {
    // Could be many different error types
}
```

**v4.x:**
```go
// All API errors implement common interface
err := client.DoSomething(ctx)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (code: %s)\n", apiErr.Message, apiErr.Code)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)
    }
}
```

### 4. **Option Functions Cleanup** (BREAKING)

**v3.x:**
```go
// Different option patterns per package
client, err := auth.NewClient(
    auth.WithAuthorizer(authorizer),
    auth.WithHTTPDebugging(true),
    auth.WithLogger(logger),
)
```

**v4.x:**
```go
// Standardized option functions
client := auth.NewClient(
    config,
    auth.WithDebug(true),  // Simplified
    auth.WithLogger(logger),
)
```

### 5. **Deprecated Method Removal** (BREAKING)

Remove methods deprecated in v3.x:
- `groups.SetSubscriptionAdminVerifiedID()` → Use `SetSubscriptionAdminVerified()`
- Any other methods marked deprecated for > 6 months

### 6. **Generic Response Types** (Enhancement)

Leverage Go 1.18+ generics more extensively:

**v3.x:**
```go
type AuthResponse struct {
    Data  TokenResponse
    Meta  ResponseMetadata
    Error *APIError
}

type TransferResponse struct {
    Data  TaskResponse
    Meta  ResponseMetadata
    Error *APIError
}
```

**v4.x:**
```go
// Single generic response type
type Response[T any] struct {
    Data  T
    Meta  ResponseMetadata
    Error *APIError
}

// Usage
func GetToken(ctx context.Context) (*Response[TokenResponse], error)
func GetTask(ctx context.Context, taskID string) (*Response[TaskResponse], error)
```

---

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)

**Goal:** Set up v4 module structure and core infrastructure

1. **Create v4 Module Structure**
   ```bash
   mkdir -p v4/pkg/core
   mkdir -p v4/pkg/services/{auth,transfer,groups,search,flows,compute,timers}
   ```

2. **Initialize v4 go.mod**
   ```bash
   cd v4
   go mod init github.com/scttfrdmn/globus-go-sdk/v4
   ```

3. **Implement Core v4 Infrastructure**
   - `v4/pkg/core/config.go` - Unified configuration
   - `v4/pkg/core/transport/transport.go` - Refactored transport
   - `v4/pkg/core/retry/retry.go` - Separate retry configuration
   - `v4/pkg/core/errors/errors.go` - Enhanced error types
   - `v4/pkg/core/scopes/scopes.go` - Scope parser and utilities

4. **Testing Framework**
   - Set up v4-specific test helpers
   - Create mock servers for testing
   - Establish CI/CD for v4 module

**Deliverables:**
- v4 module structure
- Core packages implemented
- Basic tests passing

---

### Phase 2: Auth Service Migration (Weeks 3-4)

**Goal:** Migrate Auth service to v4 patterns

1. **Implement v4 Auth Client**
   - Remove default scopes
   - Require explicit scope specification
   - Unified client construction

2. **Update Authorizers**
   - Refactor `StaticTokenAuthorizer`
   - Refactor `RefreshableTokenAuthorizer`
   - Refactor `ClientCredentialsAuthorizer`
   - All authorizers require explicit scopes

3. **Testing**
   - Port all v3 auth tests to v4
   - Add new tests for breaking changes
   - Integration tests

**Deliverables:**
- v4 Auth service fully functional
- 100% test coverage
- Migration examples

---

### Phase 3: Data Services Migration (Weeks 5-8)

**Goal:** Migrate Transfer, Search, and Groups services

1. **Week 5: Transfer Service**
   - Migrate Transfer client
   - Port all operations
   - Update tests

2. **Week 6: Search Service**
   - Migrate Search client
   - Port all operations
   - Update tests

3. **Week 7: Groups Service**
   - Migrate Groups client
   - Port all operations
   - Update tests

4. **Week 8: Integration Testing**
   - Cross-service integration tests
   - Performance benchmarks
   - Documentation

**Deliverables:**
- Transfer, Search, Groups services in v4
- Complete test coverage
- Performance validation

---

### Phase 4: Automation Services Migration (Weeks 9-11)

**Goal:** Migrate Flows, Compute, and Timers services

1. **Week 9: Flows Service**
   - Migrate Flows client
   - Add authentication policy support (v4.1.0 feature)
   - Port all operations
   - Update tests

2. **Week 10: Compute Service**
   - Migrate Compute client
   - Port all operations
   - Update tests

3. **Week 11: Timers Service**
   - Migrate Timers client
   - Port FlowTimer helpers
   - Update tests

**Deliverables:**
- Flows, Compute, Timers services in v4
- v4.1.0 feature parity
- Complete test coverage

---

### Phase 5: Documentation & Examples (Week 12)

**Goal:** Complete documentation and migration guides

1. **Documentation**
   - Update all package documentation
   - Create v4 API reference
   - Write migration guides

2. **Examples**
   - Port all v3 examples to v4
   - Create v4-specific examples
   - Add comparison examples (v3 vs v4)

3. **Migration Tools**
   - Create migration checklist
   - Document breaking changes
   - Provide code rewrite examples

**Deliverables:**
- Complete v4 documentation
- Migration guide
- Example applications

---

### Phase 6: Release & Maintenance (Ongoing)

**Goal:** Release v4.0.0-1 and establish maintenance cycle

1. **Pre-Release**
   - Beta testing period
   - Community feedback
   - Bug fixes

2. **v4.0.0-1 Release**
   - Official release
   - GitHub release notes
   - Announcement blog post

3. **Post-Release**
   - Monitor issues
   - Quick patches as needed
   - User support

---

## Migration Guide for Users

### Quick Migration Checklist

- [ ] Update import paths: `v3` → `v4`
- [ ] Add explicit scopes to authorizers
- [ ] Update client construction patterns
- [ ] Add context to all method calls
- [ ] Update error handling to use `core.APIError`
- [ ] Remove deprecated method calls
- [ ] Update tests
- [ ] Verify functionality

### Side-by-Side Comparison

#### Import Paths
```go
// v3
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"

// v4
import "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
```

#### Client Construction
```go
// v3
client, err := auth.NewClient(
    auth.WithAuthorizer(authorizer),
)

// v4
config := core.NewConfig(core.WithAuthorizer(authorizer))
client := auth.NewClient(config)
```

#### Authorizers with Scopes
```go
// v3 (implicit scopes)
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret)

// v4 (explicit scopes REQUIRED)
authorizer := auth.NewClientCredentialsAuthorizer(
    clientID,
    clientSecret,
    []string{
        "urn:globus:auth:scope:transfer.api.globus.org:all",
    },
)
```

#### Error Handling
```go
// v3
err := client.GetGroup("group-id")
if err != nil {
    log.Fatal(err)
}

// v4
err := client.GetGroup(ctx, "group-id")
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API Error [%s]: %s", apiErr.Code, apiErr.Message)
    } else {
        log.Fatal(err)
    }
}
```

---

## Testing Strategy

### Unit Tests
- **Target:** 100% code coverage for new v4 code
- **Approach:** Port all v3 tests to v4, add tests for new features

### Integration Tests
- **Target:** All major workflows tested against live API
- **Approach:** Use test credentials, clean up resources

### Compatibility Tests
- **Target:** Ensure v3 and v4 can coexist
- **Approach:** Create test project using both v3 and v4 simultaneously

### Performance Tests
- **Target:** v4 performance >= v3 performance
- **Approach:** Benchmark common operations

### Migration Tests
- **Target:** Validate migration paths work
- **Approach:** Create migration examples and test them

---

## Timeline and Milestones

### 2025 Timeline

| Week | Phase | Milestone |
|------|-------|-----------|
| 1-2 | Foundation | v4 module structure complete |
| 3-4 | Auth Service | Auth service migrated |
| 5 | Transfer | Transfer service migrated |
| 6 | Search | Search service migrated |
| 7 | Groups | Groups service migrated |
| 8 | Integration | Data services integration complete |
| 9 | Flows | Flows service migrated |
| 10 | Compute | Compute service migrated |
| 11 | Timers | Timers service migrated |
| 12 | Documentation | Documentation complete |
| 13+ | Release | v4.0.0-1 release |

**Target Release:** December 2025

### Maintenance Strategy

**v3.x (Maintenance Mode):**
- Critical bug fixes only
- Security patches
- No new features
- EOL: 6 months after v4.0.0 release (June 2026)

**v4.x (Active Development):**
- New features
- Python SDK parity updates
- Regular releases
- Active support

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes break user code | High | Clear migration guide, long v3 support |
| Python SDK v5.x released before v4 done | Medium | Focus on v4.0-v4.1 parity first |
| Performance regression in v4 | Medium | Benchmark early, optimize as needed |
| User adoption slow | Low | Side-by-side support eases migration |
| Module path confusion | Medium | Clear documentation, examples |

---

## Success Criteria

v4.0.0-1 is ready for release when:

- ✅ All services migrated and tested
- ✅ Python SDK v4.1.0 feature parity achieved
- ✅ 100% test coverage maintained
- ✅ Performance >= v3.x
- ✅ Complete documentation and migration guide
- ✅ Examples for all services
- ✅ No critical bugs in beta testing
- ✅ Community feedback addressed

---

## Next Steps

1. **Review and approve this plan**
2. **Create v4 branch for development**
3. **Set up v4 module structure** (Phase 1)
4. **Begin Auth service migration** (Phase 2)
5. **Establish regular progress reviews**

---

**Document Version:** 1.0
**Last Updated:** October 25, 2025
**Author:** Globus Go SDK Team

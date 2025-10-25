<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Migrating from Globus Go SDK v3 to v4

This guide helps you migrate your code from Globus Go SDK v3.x to v4.x.

## Table of Contents

1. [Overview](#overview)
2. [Breaking Changes](#breaking-changes)
3. [Step-by-Step Migration](#step-by-step-migration)
4. [Common Patterns](#common-patterns)
5. [Service-Specific Changes](#service-specific-changes)
6. [Testing Your Migration](#testing-your-migration)
7. [Troubleshooting](#troubleshooting)

---

## Overview

### Why v4?

- **Python SDK Parity:** Synchronized with upstream Globus Python SDK v4.x
- **Better Type Safety:** Enhanced error handling and type checking
- **Improved Architecture:** Cleaner separation of concerns
- **Explicit Scopes:** More secure by requiring explicit scope declarations
- **Future-Proof:** Foundation for future Globus platform features

### Migration Strategy

You have two options:

1. **Gradual Migration:** Use both v3 and v4 side-by-side (recommended)
2. **Complete Migration:** Update everything at once

This guide focuses on gradual migration for minimal disruption.

---

## Breaking Changes

### 1. Import Paths Changed

**Before (v3):**
```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
```

**After (v4):**
```go
import "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
```

### 2. Explicit Scopes Required

**Before (v3):**
```go
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret)
// Scopes automatically set
```

**After (v4):**
```go
authorizer := auth.NewClientCredentialsAuthorizer(
    clientID,
    clientSecret,
    []string{
        "urn:globus:auth:scope:transfer.api.globus.org:all",
        "urn:globus:auth:scope:search.api.globus.org:all",
    }, // Explicit scopes REQUIRED
)
```

### 3. Unified Client Construction

**Before (v3):**
```go
client, err := auth.NewClient(
    auth.WithAuthorizer(authorizer),
)
```

**After (v4):**
```go
config := core.NewConfig(core.WithAuthorizer(authorizer))
client := auth.NewClient(config)
```

### 4. Context Required Everywhere

**Before (v3):**
```go
// Some methods didn't require context
group, err := client.GetGroup(groupID)
```

**After (v4):**
```go
// ALL methods require context
group, err := client.GetGroup(ctx, groupID)
```

### 5. Enhanced Error Handling

**Before (v3):**
```go
err := client.DoSomething()
if err != nil {
    log.Fatal(err)
}
```

**After (v4):**
```go
err := client.DoSomething(ctx)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        log.Printf("API Error [%s]: %s", apiErr.Code, apiErr.Message)
        log.Printf("Request ID: %s", apiErr.RequestID)
    } else {
        log.Fatal(err)
    }
}
```

### 6. Deprecated Methods Removed

These methods are removed in v4:

- `groups.SetSubscriptionAdminVerifiedID()` → Use `SetSubscriptionAdminVerified()`
- Any other methods deprecated in v3.x for > 6 months

---

## Step-by-Step Migration

### Step 1: Install v4 Module

Keep v3 installed and add v4:

```bash
# v3 remains installed
go get github.com/scttfrdmn/globus-go-sdk/v3

# Add v4
go get github.com/scttfrdmn/globus-go-sdk/v4
```

Your `go.mod` will have both:
```go
require (
    github.com/scttfrdmn/globus-go-sdk/v3 v3.65.0-1
    github.com/scttfrdmn/globus-go-sdk/v4 v4.0.0-1
)
```

### Step 2: Create New Config

Create a v4 configuration:

```go
import (
    v4core "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Define your scopes explicitly
scopes := []string{
    "urn:globus:auth:scope:transfer.api.globus.org:all",
    "urn:globus:auth:scope:groups.api.globus.org:all",
    "urn:globus:auth:scope:search.api.globus.org:all",
}

// Create authorizer with explicit scopes
authorizer := v4auth.NewClientCredentialsAuthorizer(
    os.Getenv("GLOBUS_CLIENT_ID"),
    os.Getenv("GLOBUS_CLIENT_SECRET"),
    scopes,
)

// Create unified config
config := v4core.NewConfig(
    v4core.WithAuthorizer(authorizer),
)
```

### Step 3: Migrate One Service at a Time

Start with the service you use most:

```go
// Keep v3 imports with aliases
import (
    v3auth "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    v4auth "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// v3 client (keep using for now)
v3Client, _ := v3auth.NewClient(v3auth.WithAuthorizer(v3Authorizer))

// v4 client (new)
v4Client := v4auth.NewClient(config)

// Gradually switch from v3Client to v4Client
```

### Step 4: Update Method Calls

Add context to all method calls:

```go
ctx := context.Background()

// v3
group, err := v3Client.GetGroup(groupID)

// v4
group, err := v4Client.GetGroup(ctx, groupID)
```

### Step 5: Update Error Handling

```go
import "errors"

err := v4Client.DoSomething(ctx)
if err != nil {
    var apiErr *v4core.APIError
    if errors.As(err, &apiErr) {
        // Handle API-specific error
        fmt.Printf("API Error [%d]: %s\n", apiErr.StatusCode, apiErr.Message)
        fmt.Printf("Request ID: %s\n", apiErr.RequestID)

        // Check specific error types
        if apiErr.StatusCode == 401 {
            // Handle unauthorized
        } else if apiErr.StatusCode == 404 {
            // Handle not found
        }
    } else {
        // Handle other errors
        log.Fatal(err)
    }
}
```

### Step 6: Test Thoroughly

```bash
# Run tests
go test ./...

# Test manually
go run main.go
```

### Step 7: Remove v3 Dependency

Once fully migrated, remove v3:

```bash
go get github.com/scttfrdmn/globus-go-sdk/v3@none
go mod tidy
```

---

## Common Patterns

### Pattern 1: OAuth2 Authorization Code Flow

**v3:**
```go
import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"

authClient, _ := auth.NewClient(
    auth.WithAccessToken(token),
)

authURL := authClient.GetAuthorizationURL(clientID, redirectURI, scopes)
// ... user authorizes ...
tokens, _ := authClient.ExchangeAuthorizationCode(ctx, clientID, clientSecret, code, redirectURI)
```

**v4:**
```go
import (
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// Create config (no scopes needed for Auth client itself)
config := core.NewConfig()
authClient := auth.NewClient(config)

// Explicitly specify scopes you want
scopes := []string{
    "urn:globus:auth:scope:transfer.api.globus.org:all",
}

authURL := authClient.GetAuthorizationURL(ctx, clientID, redirectURI, scopes)
// ... user authorizes ...
tokens, err := authClient.ExchangeAuthorizationCode(ctx, clientID, clientSecret, code, redirectURI)
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) {
        log.Fatalf("OAuth exchange failed: %s", apiErr.Message)
    }
}
```

### Pattern 2: Client Credentials Flow

**v3:**
```go
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret)
client, _ := transfer.NewClient(transfer.WithAuthorizer(authorizer))
```

**v4:**
```go
// Explicitly define required scopes
scopes := []string{
    "urn:globus:auth:scope:transfer.api.globus.org:all",
}

authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret, scopes)
config := core.NewConfig(core.WithAuthorizer(authorizer))
client := transfer.NewClient(config)
```

### Pattern 3: Token Refresh

**v3:**
```go
refreshAuthorizer := auth.NewRefreshableTokenAuthorizer(
    clientID,
    clientSecret,
    refreshToken,
)
client, _ := groups.NewClient(groups.WithAuthorizer(refreshAuthorizer))
```

**v4:**
```go
refreshAuthorizer := auth.NewRefreshableTokenAuthorizer(
    clientID,
    clientSecret,
    refreshToken,
    scopes, // Explicit scopes required
)
config := core.NewConfig(core.WithAuthorizer(refreshAuthorizer))
client := groups.NewClient(config)
```

---

## Service-Specific Changes

### Auth Service

```go
// v4 example
import (
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

config := core.NewConfig() // Auth client doesn't need authorizer
authClient := auth.NewClient(config)

// All methods require context
userInfo, err := authClient.GetUserInfo(ctx, accessToken)
```

### Transfer Service

```go
// v4 example
config := core.NewConfig(core.WithAuthorizer(authorizer))
transferClient := transfer.NewClient(config)

// Context required
endpoints, err := transferClient.ListEndpoints(ctx, nil)

// Enhanced error handling
if err != nil {
    var apiErr *core.APIError
    if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
        fmt.Println("Endpoint not found")
    }
}
```

### Groups Service

```go
// v4 example
config := core.NewConfig(core.WithAuthorizer(authorizer))
groupsClient := groups.NewClient(config)

// Use new method name (deprecated method removed)
groups, err := groupsClient.ListGroups(ctx, &groups.ListGroupsOptions{
    Statuses: []string{"active"},
})
```

### Search Service

```go
// v4 example
config := core.NewConfig(core.WithAuthorizer(authorizer))
searchClient := search.NewClient(config)

// Context required, enhanced types
results, err := searchClient.Search(ctx, indexID, "query")
```

### Flows Service

```go
// v4 example (with v4.1.0 feature)
config := core.NewConfig(core.WithAuthorizer(authorizer))
flowsClient := flows.NewClient(config)

// New authentication policy support
flow, err := flowsClient.CreateFlow(ctx, &flows.FlowCreateRequest{
    Title: "My Flow",
    Definition: flowDef,
    AuthenticationPolicy: "policy-uuid", // NEW in v4.1.0
})
```

### Timers Service

```go
// v4 example (FlowTimer preserved)
config := core.NewConfig(core.WithAuthorizer(authorizer))
timersClient := timers.NewClient(config)

// FlowTimer still available
flowTimer := &timers.FlowTimer{
    FlowID:    "flow-id",
    FlowScope: "flow-scope",
    FlowInput: map[string]interface{}{...},
}

timer, err := timersClient.CreateFlowTimerOnce(ctx, "name", startTime, flowTimer, nil)
```

---

## Testing Your Migration

### Unit Tests

```go
func TestMigrationToV4(t *testing.T) {
    ctx := context.Background()

    // Set up v4 client
    config := core.NewConfig(core.WithAuthorizer(mockAuthorizer))
    client := auth.NewClient(config)

    // Test with context
    userInfo, err := client.GetUserInfo(ctx, "token")
    if err != nil {
        t.Fatalf("GetUserInfo failed: %v", err)
    }

    // Verify result
    assert.NotNil(t, userInfo)
}
```

### Integration Tests

```go
func TestIntegrationV4(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx := context.Background()

    // Use real credentials
    scopes := []string{"urn:globus:auth:scope:transfer.api.globus.org:all"}
    authorizer := auth.NewClientCredentialsAuthorizer(
        os.Getenv("CLIENT_ID"),
        os.Getenv("CLIENT_SECRET"),
        scopes,
    )

    config := core.NewConfig(core.WithAuthorizer(authorizer))
    client := transfer.NewClient(config)

    // Test real API call
    endpoints, err := client.ListEndpoints(ctx, nil)
    if err != nil {
        t.Fatalf("ListEndpoints failed: %v", err)
    }

    assert.NotEmpty(t, endpoints)
}
```

---

## Troubleshooting

### Problem: "cannot find package"

**Error:**
```
cannot find package "github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
```

**Solution:**
```bash
go get github.com/scttfrdmn/globus-go-sdk/v4@latest
go mod tidy
```

### Problem: "too many arguments in call"

**Error:**
```
too many arguments in call to client.GetGroup
```

**Solution:** Add context parameter:
```go
// Before
group, err := client.GetGroup(groupID)

// After
group, err := client.GetGroup(ctx, groupID)
```

### Problem: "scope required but not provided"

**Error:**
```
Error: scopes are required for ClientCredentialsAuthorizer
```

**Solution:** Explicitly provide scopes:
```go
// Before (v3)
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret)

// After (v4)
scopes := []string{
    "urn:globus:auth:scope:transfer.api.globus.org:all",
}
authorizer := auth.NewClientCredentialsAuthorizer(clientID, clientSecret, scopes)
```

### Problem: "undefined: SetSubscriptionAdminVerifiedID"

**Error:**
```
undefined: groups.Client.SetSubscriptionAdminVerifiedID
```

**Solution:** Use the new method name:
```go
// Before (deprecated in v3, removed in v4)
err := client.SetSubscriptionAdminVerifiedID(ctx, groupID, subID)

// After
err := client.SetSubscriptionAdminVerified(ctx, groupID, subID)
```

### Problem: Import conflicts with v3 and v4

**Error:**
```
redeclared in this block: auth
```

**Solution:** Use import aliases:
```go
import (
    v3auth "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
    v4auth "github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// Use with aliases
v3Client, _ := v3auth.NewClient(...)
v4Client := v4auth.NewClient(...)
```

---

## Checklist

Use this checklist to track your migration:

- [ ] Install v4 module
- [ ] Update import paths
- [ ] Create v4 config with explicit scopes
- [ ] Migrate Auth service
- [ ] Migrate Transfer service
- [ ] Migrate Groups service
- [ ] Migrate Search service
- [ ] Migrate Flows service
- [ ] Migrate Compute service
- [ ] Migrate Timers service
- [ ] Add context to all method calls
- [ ] Update error handling
- [ ] Remove deprecated method calls
- [ ] Update unit tests
- [ ] Run integration tests
- [ ] Update documentation
- [ ] Remove v3 dependency
- [ ] Deploy to production

---

## Need Help?

- **GitHub Issues:** https://github.com/scttfrdmn/globus-go-sdk/issues
- **Documentation:** https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v4
- **Examples:** See `v4/cmd/examples/`

---

**Migration Guide Version:** 1.0
**Last Updated:** October 25, 2025

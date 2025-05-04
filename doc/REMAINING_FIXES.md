<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Remaining Issues After Import Cycle Fix

This document outlines the remaining issues that need to be addressed after fixing the import cycle between `pkg/core` and `pkg/core/transport`.

## Current Status

The primary import cycle between `pkg/core` and `pkg/core/transport` has been fixed. We can now successfully build:

```bash
go build ./pkg/core/...
go build ./pkg/core/transport/...
```

However, there are still issues when building the entire package:

```bash
go build ./pkg/...
```

## Remaining Issues

### 1. Auth Client API Changes

The auth client constructor has been updated to use a functional options pattern, but not all code has been updated to use this new API.

In `pkg/globus.go` line 58:
```go
authClient := auth.NewClient(c.ClientID, c.ClientSecret)
```

Needs to be updated to match the new API that returns a client and an error:
```go
authClient, err := auth.NewClient(auth.WithClientID(c.ClientID), auth.WithClientSecret(c.ClientSecret))
if err != nil {
    // Handle error
}
```

### 2. Missing Client Options in Auth Package

The auth package needs to define options like `WithClientID` and `WithClientSecret` to support the functional options pattern:

```go
// WithClientID sets the client ID
func WithClientID(clientID string) ClientOption {
    return func(c *Client) {
        c.ClientID = clientID
    }
}

// WithClientSecret sets the client secret
func WithClientSecret(clientSecret string) ClientOption {
    return func(c *Client) {
        c.ClientSecret = clientSecret
    }
}
```

### 3. Similar Issues in Other Service Packages

There are likely similar API mismatches in other service packages that need to be checked and updated.

### 4. HTTP Pool Management

The HTTP pool system (`httppool.GetHTTPClientForService`) may need to be updated to use the new interfaces-based approach.

### 5. Example Files

Many example files may need updates to work with the new API changes.

## Action Plan

1. **Fix Auth Package**:
   - Create ClientOption functions
   - Update method signatures
   - Fix all return values to include error handling

2. **Update Globus.go**:
   - Update NewAuthClient to use the new Auth API
   - Add proper error handling
   - Ensure backwards compatibility where possible

3. **Verify Other Services**:
   - Check each service for similar issues
   - Update service packages to use consistent patterns  

4. **Revisit HTTP Pool**:
   - Ensure HTTP pool management works with new interface-based approach
   - Update any direct uses of concrete transport types

5. **Fix Examples**:
   - Update example code to work with the updated API
   - Ensure examples follow current best practices

## Testing Strategy

1. Fix issues incrementally
2. Build modules one at a time:
   - `go build ./pkg/services/auth/...`
   - `go build ./pkg/...`
   - `go build ./cmd/...`
3. Run tests with each change: `go test ./pkg/...`
4. Once the code builds successfully, re-enable CI workflows

## Next Steps

Once the import cycle and related issues are fully resolved, we should:

1. Update documentation to reflect the new design
2. Re-enable GitHub Actions workflows
3. Consider creating a standardized functional options pattern across all service clients

## See Also

- [IMPORT_CYCLE_FIX.md](IMPORT_CYCLE_FIX.md)
- [IMPORT_CYCLE_RESOLUTION.md](IMPORT_CYCLE_RESOLUTION.md)
- [GITHUB_ACTIONS_STATUS.md](GITHUB_ACTIONS_STATUS.md)
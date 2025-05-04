# API Consistency Changes

## Overview

This PR implements a consistent API pattern across all service client packages in the SDK. The functional options pattern has been applied to all client constructors, ensuring a uniform API experience for SDK users.

## Changes

1. **Implemented the functional options pattern for all service clients:**
   - Flows client
   - Search client
   - Compute client
   - Timers client

2. **Created `options.go` files in each client package with:**
   - ClientOption type
   - clientOptions struct
   - Helper functions (WithX) for various configuration options
   - Default options implementation

3. **Updated `pkg/globus.go` with:**
   - Consistent error handling for all client constructors
   - Service-specific option handling
   - Better pool configuration integration

4. **Updated all example applications and tests to use the new API pattern**

## Motivation

Previously, client constructors had inconsistent signatures:
- Some returned errors, others didn't
- Some used the functional options pattern, others took direct parameters
- Some required access tokens as the first parameter, others handled this differently

This made the SDK harder to use consistently and made it difficult for users to switch between different service clients.

## Implementation

All client constructors now follow this pattern:
```go
func NewClient(opts ...ClientOption) (*Client, error)
```

Where ClientOption is a functional option type defined in each package:
```go
type ClientOption func(*clientOptions)
```

This approach provides several benefits:
1. Better extensibility - new options can be added without breaking existing code
2. More readable code - method calls clearly indicate what options are being set
3. Better default handling - default options are applied first, then overridden by user options
4. Error handling - constructors can now return errors for invalid configurations

## Testing

All example applications and integration tests have been updated to use the new API pattern and have been tested to ensure they work correctly.

## Example Usage

Before:
```go
// Different patterns for different clients
authClient := auth.NewClient("token")
transferClient := transfer.NewClient(options...)
searchClient := search.NewClient("token", coreOptions...)
```

After:
```go
// Consistent pattern for all clients
authClient, err := auth.NewClient(auth.WithAccessToken("token"))
transferClient, err := transfer.NewClient(transfer.WithAccessToken("token"))
searchClient, err := search.NewClient(search.WithAccessToken("token"))
```
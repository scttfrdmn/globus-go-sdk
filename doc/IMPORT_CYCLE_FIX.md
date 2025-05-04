<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Import Cycle Fix: Core and Transport Package

This document explains the recent fix for the import cycle between the `pkg/core` and `pkg/core/transport` packages.

## Problem Description

An import cycle was present in the codebase between the following packages:

- **core package** (`pkg/core/client.go`): imports `pkg/core/transport` to create a Transport instance
- **transport package** (`pkg/core/transport/transport.go`): imports `pkg/core/interfaces` to reference the ClientInterface

This circular dependency prevented the code from compiling and caused CI/CD workflows to fail.

## Solution Implemented

We implemented a solution based on the **deferred initialization** pattern with the following components:

1. **DeferredTransport**: 
   - Added a new struct in the transport package to hold transport configuration
   - Allows creating transport settings without requiring the client instance

2. **Transport Factory Function**:
   - Created `InitTransport()` in a new file `pkg/core/transport_init.go`
   - This function handles creating the transport with deferred initialization
   - Breaks the import cycle by putting the client-to-transport connection in the core package

3. **Interface-based Design**:
   - Updated Client struct to use the Transport interface instead of concrete Transport type
   - Simplified client options to rely on transport initialization in NewClient

## Changes Made

### 1. Updated `pkg/core/transport/transport.go`

Added a DeferredTransport type and supporting functions:

```go
// DeferredTransport holds the transport configuration until a client is available
type DeferredTransport struct {
    Debug  bool
    Trace  bool
    Logger *log.Logger
}

// NewDeferredTransport creates a configuration for a transport that can be attached to a client later
func NewDeferredTransport(options *Options) *DeferredTransport {
    // Configuration setup logic
    return &DeferredTransport{...}
}

// AttachClient creates a Transport by attaching a client to a DeferredTransport
func (dt *DeferredTransport) AttachClient(client interfaces.ClientInterface) *Transport {
    return &Transport{
        Client: client,
        Debug:  dt.Debug,
        Trace:  dt.Trace,
        Logger: dt.Logger,
    }
}
```

### 2. Created `pkg/core/transport_init.go`

Created a new file to handle transport initialization:

```go
// InitTransport initializes a transport for a client
// This function helps break the import cycle between core and transport packages
func InitTransport(client interfaces.ClientInterface, debug, trace bool) interfaces.Transport {
    // Create deferred transport first
    dt := transport.NewDeferredTransport(&transport.Options{
        Debug:  debug,
        Trace:  trace,
        Logger: loggerForTransport,
    })
    
    // Now attach the client to create the actual transport
    return dt.AttachClient(client)
}
```

### 3. Updated `pkg/core/client.go`

- Removed import of `github.com/scttfrdmn/globus-go-sdk/pkg/core/transport`
- Changed `Transport` field type from `*transport.Transport` to `interfaces.Transport`
- Updated `NewClient()` to use `InitTransport()` function
- Simplified HTTP debugging and tracing options

## Advantages of This Approach

1. **No Import Cycle**: The circular dependency has been eliminated
2. **Interface-based Design**: Code now depends on interfaces rather than concrete implementations
3. **Better Testability**: Easier to create mock transports for testing
4. **Cleaner Separation**: Clear boundaries between packages
5. **Minimally Invasive**: No major refactoring required in the rest of the codebase

## How It Works

The key insight is pushing the integration code (creating the Transport with a Client) into the package that already has visibility of both types. By creating a "deferred" transport configuration in the transport package, we can later attach the client in the core package without directly importing the concrete Transport type.

## Testing

To verify this fix:
1. Run `go build ./pkg/...` to ensure compilation works
2. Run the unit tests for both packages to ensure functionality 
3. Enable GitHub Actions workflows to verify CI/CD passes

## Related Work

For more comprehensive information about import cycle resolution in this project, see [IMPORT_CYCLE_RESOLUTION.md](IMPORT_CYCLE_RESOLUTION.md).
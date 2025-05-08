<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v0.9.12 Release Notes

## Overview

Globus Go SDK v0.9.12 is a maintenance release that fixes critical issues from v0.9.11 and adds device authentication flow support for CLI applications.

## Fixed Issues

### Issue #13: Missing Functions in SDK v0.9.11

In v0.9.11, the functions `SetConnectionPoolManager` and `EnableDefaultConnectionPool` were removed, but references to them remained in `pkg/core/transport_init.go`. This caused compilation errors when building projects that depend on the SDK.

This issue has been fixed by:
- Updating the `SetConnectionPoolManager` function to accept both `ConnectionPoolProvider` and `interfaces.ConnectionPoolManager` types
- Implementing an adapter pattern to ensure compatibility with both interface types
- Ensuring proper initialization of connection pools

### Issue #12: Missing Device Authentication Flow

The device authentication flow was not available in previous SDK versions, making it difficult for CLI applications to authenticate without a browser. This release adds full support for device authentication with a clean, user-friendly API.

## New Features

### Device Authentication Flow

Added complete support for the device authentication flow:

- **New Models:**
  - `DeviceCodeResponse` to handle the response from device code requests
  - Error types with device-specific error codes (`authorization_pending`, `slow_down`, etc.)

- **New Methods:**
  - `RequestDeviceCode`: Initiates the device flow and returns verification information
  - `PollDeviceCode`: Polls for user authorization completion
  - `CompleteDeviceFlow`: Convenience method that handles the entire flow

- **Error Handling:**
  - `DeviceAuthError` type with specific error codes
  - Helper functions for checking error types (e.g., `IsAuthorizationPending`, `IsSlowDown`)

- **Example and Documentation:**
  - Complete example in `cmd/examples/device-auth/`
  - Comprehensive guide in the documentation

## Migration Notes

This release contains no breaking changes and should be a straightforward upgrade from v0.9.11.

### For CLI Applications

CLI applications can now implement device authentication, which provides a better user experience in non-browser environments:

```go
// Get a device code
deviceCode, err := authClient.RequestDeviceCode(ctx, scopes...)

// Display instructions to the user
fmt.Printf("Visit %s and enter code: %s\n", deviceCode.VerificationURI, deviceCode.UserCode)

// Poll for completion
token, err := authClient.CompleteDeviceFlow(ctx, displayCallback, 0, scopes...)
```

## Documentation Updates

- Added a device authentication guide
- Updated the auth client reference documentation
- Added error handling documentation for device-specific errors

## Additional Information

- All tests pass for core and auth packages
- Documentation site builds successfully with the new content
- Verified compatibility with dependent projects
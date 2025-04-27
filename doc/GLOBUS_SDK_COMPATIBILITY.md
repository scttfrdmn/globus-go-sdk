# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Globus API Compatibility

_Last Updated: April 27, 2025_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

This document defines the Globus API compatibility for the Globus Go SDK, specifying which API versions are supported and tested, along with any limitations or constraints.

## Table of Contents

- [API Version Support](#api-version-support)
- [Service Compatibility Matrix](#service-compatibility-matrix)
- [Feature Support](#feature-support)
- [Authentication Methods](#authentication-methods)
- [Testing Status](#testing-status)
- [Known Limitations](#known-limitations)
- [Versioning Policy](#versioning-policy)
- [Compatibility Checks](#compatibility-checks)

## API Version Support

The Globus Go SDK targets specific API versions for each Globus service:

| Service | API Version | Base URL |
|---------|-------------|----------|
| Auth    | v2          | https://auth.globus.org/v2/ |
| Transfer | v0.10      | https://transfer.api.globus.org/v0.10/ |
| Search  | v1.0        | https://search.api.globus.org/v1/ |
| Groups  | v2          | https://groups.api.globus.org/v2/ |
| Flows   | Beta        | https://flows.globus.org/api/ |

The SDK's behavior with newer or older API versions is undefined. The service clients are designed to work with the specific API versions listed above.

## Service Compatibility Matrix

| Service Feature | Status | Notes |
|-----------------|--------|-------|
| **Auth Service** | ✅ Supported | |
| OAuth2 Authorization Code Flow | ✅ Tested | Primary flow for interactive applications |
| OAuth2 Client Credentials Flow | ✅ Tested | Primary flow for non-interactive applications |
| Token Introspection | ✅ Tested | |
| Token Revocation | ✅ Tested | |
| Dependent Tokens | ✅ Tested | |
| Identity Management | ⚠️ Limited | Basic identity operations are supported |
| MFA Support | ✅ Supported | |
| **Transfer Service** | ✅ Supported | |
| Endpoint Management | ✅ Tested | List, get, create operations |
| Transfer Task Submission | ✅ Tested | Basic transfer submissions |
| Transfer Task Management | ✅ Tested | Monitoring, cancellation |
| Recursive Directory Transfers | ✅ Tested | |
| Resumable Transfers | ✅ Tested | |
| Endpoint Activation | ⚠️ Limited | Basic activation is supported |
| Bookmark Management | ⚠️ Limited | |
| **Search Service** | ✅ Supported | |
| Index Management | ✅ Tested | |
| Query Operations | ✅ Tested | |
| Advanced Queries | ✅ Tested | |
| Ingest Operations | ✅ Tested | |
| **Groups Service** | ✅ Supported | |
| Group Management | ✅ Tested | |
| Membership Management | ✅ Tested | |
| Access Control | ✅ Tested | |
| **Flows Service** | ⚠️ Limited | |
| Flow Management | ✅ Tested | |
| Flow Execution | ✅ Tested | |
| Flow Monitoring | ✅ Tested | |
| Custom Components | ❌ Unsupported | |

## Feature Support

### Core Features

| Feature | Status | Notes |
|---------|--------|-------|
| Context-based API | ✅ Supported | All API calls support context for cancellation |
| Custom HTTP Client | ✅ Supported | Customizable HTTP clients |
| Connection Pooling | ✅ Supported | Optimized connection reuse |
| Logging | ✅ Supported | Structured logging with levels |
| Rate Limiting | ✅ Supported | Client-side rate limiting with backoff |
| Circuit Breaking | ✅ Supported | Automatic failure detection |
| Error Handling | ✅ Supported | Typed errors with predicates |
| Configuration | ✅ Supported | Environment and code-based config |

### Authentication and Authorization

| Feature | Status | Notes |
|---------|--------|-------|
| Static Tokens | ✅ Supported | For applications with pre-existing tokens |
| Refresh Token Management | ✅ Supported | Automatic token refresh |
| Token Storage | ✅ Supported | Memory and file-based storage |
| Custom Token Storage | ✅ Supported | Interface for custom storage |
| Scope Management | ✅ Supported | Service-specific and custom scopes |

## Authentication Methods

The SDK supports these authentication methods:

| Method | Support | Use Case |
|--------|---------|----------|
| Authorization Code | ✅ Tested | Interactive applications where a user authorizes access |
| Client Credentials | ✅ Tested | Server-to-server applications with client secret |
| Refresh Token | ✅ Tested | Long-running applications that maintain access |
| Static Token | ✅ Tested | Applications with pre-existing tokens |
| Device Code | ❌ Planned | Devices without input capabilities |

## Testing Status

The SDK's test coverage for Globus API integration:

| Test Type | Status | Description |
|-----------|--------|-------------|
| Unit Tests | ✅ Complete | Tests for SDK components in isolation |
| Mock API Tests | ✅ Complete | Tests against mock Globus API |
| Integration Tests | ⚠️ Partial | Tests against actual Globus APIs |
| Performance Tests | ⚠️ Partial | Tests for performance characteristics |
| Security Tests | ✅ Complete | Tests for security issues |

### Integration Testing Requirements

To run integration tests against the Globus API:

- A Globus account
- Client application credentials (ID and secret)
- Optionally, Globus endpoints with appropriate permissions

See [Integration Testing Guide](development/testing.md#integration-testing) for details on setting up integration tests.

## Known Limitations

1. **API Version Constraints**: The SDK is designed for specific API versions and may break with future API changes.

2. **Endpoint Activation**: The SDK supports basic endpoint activation but does not handle all activation methods or requirements.

3. **Error Handling**: Some API-specific errors may not be properly typed or handled.

4. **Performance**: Large directory transfers might have performance implications due to memory usage.

5. **Search Query Limitations**: Some advanced search query features might not be fully supported.

6. **Flows Compatibility**: The Flows API is in beta and compatibility may change.

## Versioning Policy

The Globus Go SDK follows [Semantic Versioning](https://semver.org):

- **MAJOR** version increments for incompatible API changes
- **MINOR** version increments for backward-compatible functionality
- **PATCH** version increments for backward-compatible bug fixes

### Compatibility Guarantees

- **Major Versions**: No backward compatibility guarantees between major versions
- **Minor Versions**: Backward compatible within the same major version
- **Patch Versions**: Fully compatible within the same minor version

### Globus API Changes

When Globus makes changes to their APIs:

- **Breaking Changes**: We will increment our major version
- **New Features**: We will increment our minor version
- **Bugfixes**: We will increment our patch version

## Compatibility Checks

The SDK includes runtime compatibility checks:

```go
// Example of version compatibility check
config := pkg.NewConfigFromEnvironment().
    WithAPIVersionCheck(true) // Enable API version validation

client := config.NewTransferClient(accessToken)

// This will fail if the Transfer API version is not compatible
endpoints, err := client.ListEndpoints(ctx, nil)
if err != nil {
    // Handle API version incompatibility
}
```

### Disabling Compatibility Checks

For advanced use cases, you can disable compatibility checks:

```go
// Disable API version checking (USE WITH CAUTION)
config := pkg.NewConfigFromEnvironment().
    WithAPIVersionCheck(false)
```

**WARNING**: Disabling compatibility checks may lead to unexpected behavior if the API has changed.

### Custom API Versions

For testing purposes or when using custom Globus deployments:

```go
// Use a custom API version for a specific service
config := pkg.NewConfigFromEnvironment().
    WithCustomAPIVersion("transfer", "v0.11")
```

**NOTE**: Custom API versions are not officially supported and may cause unexpected behavior.
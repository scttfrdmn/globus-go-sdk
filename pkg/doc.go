// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package pkg is the root package for the Globus Go SDK.

# STABILITY: STABLE

The Globus Go SDK v3.x is synchronized with the Globus Python SDK and follows
a stable API guarantee. Individual packages have different stability levels,
documented in their respective doc.go files.

# Package Stability Overview

- core: STABLE - Foundation components with stable connection pool features
- services/auth: STABLE - Authentication, token management, and GARE support
- services/transfer: STABLE - Data transfer operations
- services/tokens: STABLE - Token storage and management
- services/search: STABLE - Search index operations
- services/flows: STABLE - Globus Flows automation
- services/groups: STABLE - Group management
- services/compute: STABLE - Compute operations
- services/timers: STABLE - Timers service

# Versioning Strategy

Starting with v3.60.0-1, this SDK uses synchronized versioning with the Python SDK:

Format: [PYTHON_SDK_VERSION]-[GO_SDK_BUILD]
- v3.60.0-1 = First Go SDK release synchronized with Python SDK v3.60.0
- v3.60.0-2 = Second Go SDK release (Go-specific improvements)
- v3.61.0-1 = First Go SDK release synchronized with Python SDK v3.61.0

# API Stability Guidelines

The SDK follows semantic versioning (https://semver.org/) with these guarantees:

1. BUILD increments (-1, -2) contain only backward-compatible changes
2. MINOR Python SDK updates may add functionality in backward-compatible ways
3. MAJOR Python SDK updates may contain breaking changes

All stable components maintain API compatibility within the Python SDK major version.

# Version Compatibility

This SDK requires Go 1.18 or later and is compatible with
the following minimum Globus API versions:

- Transfer API v0.10
- Auth API v2
- Search API v1.0
- Flows API v1.0
- Groups API v1
- Compute API v2

# Basic Usage

The recommended way to create service clients is through the main SDK entry point:

	import "github.com/scttfrdmn/globus-go-sdk/v3/pkg"

	// Create the SDK instance
	sdk := pkg.NewSDK()

	// Create an auth client
	authClient, err := sdk.Auth(
		auth.WithClientID("your-client-id"),
		auth.WithClientSecret("your-client-secret"),
	)
	if err != nil {
		// Handle error
	}

	// Create a transfer client with an authorizer
	transferClient, err := sdk.Transfer(
		transfer.WithAuthorizer(authorizer),
	)
	if err != nil {
		// Handle error
	}

# Documentation Resources

For detailed documentation on each package, refer to:

- https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3/pkg
- https://github.com/scttfrdmn/globus-go-sdk/tree/main/doc

For examples, see:

- https://github.com/scttfrdmn/globus-go-sdk/tree/main/examples
- https://github.com/scttfrdmn/globus-go-sdk/tree/main/cmd/examples
*/
package pkg

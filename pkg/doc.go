// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package pkg is the root package for the Globus Go SDK.

# STABILITY: BETA

The Globus Go SDK is currently in a beta state, approaching v1.0.0.
Individual packages have different stability levels, which are
documented in their respective doc.go files.

# Package Stability Overview

- core: BETA - Foundation components with some evolving connection pool features
- services/auth: STABLE - Authentication and token management
- services/transfer: MIXED - Core features stable, advanced features in development
- services/tokens: BETA - Token storage and management
- services/search: BETA - Search index operations
- services/flows: BETA - Globus Flows automation
- services/groups: BETA - Group management
- services/compute: BETA - Compute operations
- services/timers: BETA - Timers service

# API Stability Guidelines

The SDK follows semantic versioning (https://semver.org/) with these guarantees:

1. PATCH versions (0.9.x) contain only backward-compatible bug fixes.
2. MINOR versions (0.x.0) may add functionality in a backward-compatible manner.
3. MAJOR versions (x.0.0) may contain breaking changes.

As this SDK is pre-1.0, minor versions may occasionally contain
breaking changes, but these will be clearly documented in the CHANGELOG.

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

	import "github.com/scttfrdmn/globus-go-sdk/pkg"

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

- https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/pkg
- https://github.com/scttfrdmn/globus-go-sdk/tree/main/doc

For examples, see:

- https://github.com/scttfrdmn/globus-go-sdk/tree/main/examples
- https://github.com/scttfrdmn/globus-go-sdk/tree/main/cmd/examples
*/
package pkg

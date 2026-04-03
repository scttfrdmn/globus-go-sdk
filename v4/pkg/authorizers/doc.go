// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package authorizers provides implementations of the core.Authorizer interface,
// mirroring the Python SDK's globus_sdk.authorizers module.
//
// # Available Authorizers
//
//   - AccessTokenAuthorizer: static bearer token, never refreshes
//   - RefreshTokenAuthorizer: automatically refreshes using a refresh token
//   - ClientCredentialsAuthorizer: obtains tokens via client credentials grant
//
// # Stability: BETA
//
// The API of this package may have minor changes in minor releases.
package authorizers

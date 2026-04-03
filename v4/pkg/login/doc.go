// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package login provides interactive OAuth2 login flow implementations,
// mirroring the Python SDK's globus_sdk.login_flows module.
//
// # Available Implementations
//
//   - CommandLineLoginFlowManager: prints an authorization URL, reads the
//     auth code from stdin, and exchanges it for tokens.
//
// # Stability: BETA
//
// The API of this package may have minor changes in minor releases.
package login

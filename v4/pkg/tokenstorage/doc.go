// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package tokenstorage provides persistent storage for OAuth2 token data,
// mirroring the Python SDK's globus_sdk.token_storage module.
//
// # Available Implementations
//
//   - MemoryTokenStorage: in-memory storage, cleared when the process exits
//   - JSONTokenStorage: file-backed storage with atomic writes
//
// # Stability: BETA
//
// The API of this package may have minor changes in minor releases.
package tokenstorage

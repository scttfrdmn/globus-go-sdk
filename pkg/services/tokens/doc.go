// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package tokens provides functionality for managing OAuth 2.0 tokens in Globus SDK applications.

# STABILITY: STABLE

This package is part of the Globus Go SDK v3.x which is synchronized with the
Globus Python SDK and follows stable API guarantees. Components listed below are
considered part of the public API and will not change incompatibly within a major version:

  - Storage interface - Core interface for token persistence
  - TokenSet and Entry structs - Token data structures
  - Manager struct - Token management capabilities
  - MemoryStorage and FileStorage implementations
  - RefreshHandler interface - Token refresh capabilities

# Compatibility Guarantees

For stable packages:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v3.x to v4.x)

# Synchronized Versioning

Starting with v3.60.0-1, this package follows synchronized versioning with the Globus Python SDK.
This ensures API compatibility and feature parity across language implementations.

This package is particularly useful for web applications and other applications that need
to store, retrieve, and refresh OAuth 2.0 tokens. It provides interfaces and implementations
for token storage and management, including automatic token refreshing.

# Basic Usage

To use the tokens package, you need to create a token storage implementation and a token manager:

	// Create a file-based token storage
	storage, err := tokens.NewFileStorage("./tokens")
	if err != nil {
		log.Fatalf("Failed to initialize token storage: %v", err)
	}

	// Create an auth client (implements RefreshHandler)
	authClient := auth.NewClient("client_id", "client_secret")

	// Initialize the token manager
	tokenManager := tokens.NewManager(storage, authClient)

# Token Storage

The package provides two storage implementations:

1. MemoryStorage: In-memory token storage for testing or simple applications.
2. FileStorage: File-based token storage for persisting tokens across application restarts.

Both implementations implement the Storage interface:

	type Storage interface {
		Store(entry *Entry) error
		Lookup(resource string) (*Entry, error)
		Delete(resource string) error
		List() ([]string, error)
	}

# Token Management

The Manager handles token storage, retrieval, and automatic refreshing:

	// Get a token (will refresh if needed)
	entry, err := tokenManager.GetToken(ctx, "user_123")
	if err != nil {
		// Handle error
	}

	// Use the access token
	accessToken := entry.TokenSet.AccessToken

# Automatic Token Refreshing

The Manager will automatically refresh tokens when they are close to expiry:

	// Configure refresh threshold (default is 5 minutes)
	tokenManager.SetRefreshThreshold(10 * time.Minute)

	// Start background refresh (refreshes tokens every 15 minutes)
	stopRefresh := tokenManager.StartBackgroundRefresh(15 * time.Minute)
	defer stopRefresh() // Call when done to stop background refresh

# Thread Safety

All implementations in this package are thread-safe and can be used concurrently.
*/
package tokens

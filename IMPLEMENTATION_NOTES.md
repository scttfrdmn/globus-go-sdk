<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Implementation Notes: Tokens Package

This document describes the implementation of the tokens package for the Globus Go SDK.

## Overview

The tokens package provides functionality for storing, retrieving, and refreshing OAuth2 tokens, which is particularly useful for web applications and other long-running applications that need to maintain authentication.

## Implementation Details

### Structure

The tokens package is implemented in two main files:

1. `tokens.go`: Contains core data structures and storage implementations
2. `manager.go`: Contains the token manager for automatic token refreshing

### Key Components

#### TokenSet

The `TokenSet` struct represents a set of OAuth2 tokens:

```go
type TokenSet struct {
    AccessToken  string    // The access token
    RefreshToken string    // The refresh token (optional)
    ExpiresAt    time.Time // When the access token expires
    Scope        string    // The scopes associated with the token
    ResourceID   string    // Optional resource ID
}
```

#### Entry

The `Entry` struct represents a token entry in storage:

```go
type Entry struct {
    Resource    string    // The resource identifier (e.g., user ID)
    AccessToken string    // The access token
    RefreshToken string   // The refresh token
    ExpiresAt   time.Time // When the access token expires
    Scope       string    // The scopes associated with the token
    TokenSet    *TokenSet // For convenience (not serialized)
}
```

#### Storage Interface

The `Storage` interface defines methods for storing and retrieving tokens:

```go
type Storage interface {
    Store(entry *Entry) error
    Lookup(resource string) (*Entry, error)
    Delete(resource string) error
    List() ([]string, error)
}
```

#### Storage Implementations

Two implementations are provided:

1. `MemoryStorage`: In-memory token storage (for testing or simple applications)
2. `FileStorage`: File-based token storage for persistence across restarts

#### Token Manager

The `Manager` handles token storage, retrieval, and automatic refreshing:

```go
type Manager struct {
    Storage          Storage
    RefreshThreshold time.Duration
    RefreshHandler   RefreshHandler
}
```

### Features

- **Token Storage**: Both in-memory and file-based storage options
- **Token Refresh**: Automatic refreshing of tokens when they are close to expiry
- **Concurrent Access**: Thread-safe implementations for use in multi-goroutine environments
- **Background Refresh**: Option to start a background goroutine that periodically refreshes tokens
- **Configurable Thresholds**: Adjustable refresh thresholds

## Usage Example

```go
// Initialize token storage
tokenStorage, err := tokens.NewFileStorage("./tokens")
if err != nil {
    log.Fatalf("Failed to initialize token storage: %v", err)
}

// Create an auth client
authClient := auth.NewClient(clientID, clientSecret)

// Initialize the token manager
tokenManager := tokens.NewManager(tokenStorage, authClient)

// Configure refresh threshold (optional)
tokenManager.SetRefreshThreshold(10 * time.Minute)

// Start background refresh (optional)
stopRefresh := tokenManager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Call when done

// Get a token (will refresh if needed)
entry, err := tokenManager.GetToken(ctx, "user_123")
if err != nil {
    // Handle error
}

// Use the access token
accessToken := entry.TokenSet.AccessToken
```

## Integration with Web Applications

The tokens package is particularly useful for web applications that need to:

1. Store tokens for multiple users
2. Automatically refresh tokens
3. Maintain token state across application restarts

The webapp example demonstrates these features, showing how to integrate the tokens package with a web application.

## Future Improvements

- **Redis/Database Storage**: Add support for storing tokens in Redis or a database
- **Token Encryption**: Add encryption for stored tokens
- **Token Stats**: Add methods to track token usage and refresh statistics
- **Token Revocation**: Add support for token revocation
- **More OAuth Flows**: Add support for more OAuth flows (device code, implicit, etc.)

## Security Considerations

- Tokens are stored in plain text, so secure the storage location
- For production use, consider adding encryption for stored tokens
- The file storage implementation uses the file system permissions for security
- Background token refresh should be used carefully to avoid excessive API calls
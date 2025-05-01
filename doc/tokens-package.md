<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Tokens Package

The `tokens` package provides comprehensive token management functionality for Globus authentication. It is designed to handle the entire lifecycle of OAuth 2.0 tokens, including storage, retrieval, and automatic refreshing.

## Overview

The tokens package consists of these main components:

1. **TokenSet**: Core structure containing token data
2. **Entry**: Storage-friendly wrapper for token data
3. **Storage Interface**: Abstract interface for token persistence
4. **Storage Implementations**:
   - MemoryStorage: In-memory token storage
   - FileStorage: File-based token storage
5. **Manager**: Handles token retrieval and automatic refreshing

## TokenSet Structure

The `TokenSet` struct contains all the information related to an OAuth2 token:

```go
type TokenSet struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scope        string    `json:"scope,omitempty"`
    ResourceID   string    `json:"resource_id,omitempty"`
}
```

This structure provides methods for checking token validity and refresh capability:

```go
// Check if token is expired
if tokenSet.IsExpired() {
    // Token needs to be refreshed
}

// Check if token can be refreshed
if tokenSet.CanRefresh() {
    // Token has a refresh token available
}
```

## Entry Structure

The `Entry` struct provides a storage-friendly wrapper for token data:

```go
type Entry struct {
    Resource     string    `json:"resource"`
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scope        string    `json:"scope,omitempty"`
    TokenSet     *TokenSet `json:"-"` // Not serialized, for convenience
}
```

This structure combines the metadata needed for storage with the TokenSet for convenience.

## Storage Interface

The `Storage` interface defines the operations for token persistence:

```go
type Storage interface {
    // Store saves a token entry
    Store(entry *Entry) error

    // Lookup retrieves a token entry for a specific resource
    Lookup(resource string) (*Entry, error)

    // Delete removes a token entry for a specific resource
    Delete(resource string) error

    // List returns all stored token resources
    List() ([]string, error)
}
```

## Storage Implementations

### Memory Storage

The `MemoryStorage` implementation stores tokens in memory:

```go
// Create a memory storage
storage := tokens.NewMemoryStorage()

// Store a token
entry := &tokens.Entry{
    Resource: "user-123",
    TokenSet: &tokens.TokenSet{
        AccessToken:  "example-access-token",
        RefreshToken: "example-refresh-token",
        ExpiresAt:    time.Now().Add(1 * time.Hour),
        Scope:        "openid profile email",
    },
}
storage.Store(entry)
```

### File Storage

The `FileStorage` implementation persists tokens to disk in JSON format:

```go
// Create a file storage
storage, err := tokens.NewFileStorage("/path/to/tokens")
if err != nil {
    // Handle error
}

// Store a token
entry := &tokens.Entry{
    Resource: "user-123",
    TokenSet: &tokens.TokenSet{
        AccessToken:  "example-access-token",
        RefreshToken: "example-refresh-token",
        ExpiresAt:    time.Now().Add(1 * time.Hour),
        Scope:        "openid profile email",
    },
}
storage.Store(entry)
```

## Token Manager

The `Manager` provides advanced token management capabilities:

```go
// Create a token manager
manager := tokens.NewManager(storage, authClient)

// Configure refresh threshold
manager.SetRefreshThreshold(10 * time.Minute)

// Start background refresh
stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Call when done

// Get a token (will refresh if needed)
entry, err := manager.GetToken(ctx, "user-123")
if err != nil {
    // Handle error
}

// Use the access token
accessToken := entry.TokenSet.AccessToken
```

### Automatic Token Refreshing

The token manager automatically handles token refreshing:

1. When `GetToken` is called and the token is close to expiry
2. In the background if background refresh is enabled

### RefreshHandler Interface

The token manager uses the `RefreshHandler` interface for token refreshing:

```go
type RefreshHandler interface {
    RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, error)
}
```

The auth.Client implements this interface, so it can be used directly with the token manager.

## Thread Safety

All implementations in the tokens package are thread-safe and can be used concurrently:

- `MemoryStorage` uses a mutex to protect its internal map
- `FileStorage` uses a mutex to protect file operations
- `Manager` uses a mutex to prevent multiple simultaneous refreshes

## Examples

For complete examples of using the tokens package, see:

- [Token Management Example](../examples/token-management/README.md)
- [Web Application Example](../examples/webapp/README.md)

## Best Practices

1. **Security**: Always store tokens securely
2. **Refresh Thresholds**: Set appropriate refresh thresholds based on your application's needs
3. **Background Refresh**: Use background refresh for long-running applications
4. **Error Handling**: Implement proper error handling for token operations
5. **Resource IDs**: Use consistent resource IDs for storing tokens
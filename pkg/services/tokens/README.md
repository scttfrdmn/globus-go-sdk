<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Tokens Package

The tokens package provides functionality for storing, retrieving, and refreshing OAuth2 tokens. It is particularly useful for web applications that need to manage user tokens.

## Features

- Storage interfaces for both in-memory and file-based token storage
- Automatic token refreshing when tokens are close to expiry
- Background refresh capabilities to keep tokens fresh
- Thread-safe implementations for concurrent access

## Usage

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

// Initialize token storage
tokenStorage, err := tokens.NewFileStorage("./tokens")
if err != nil {
    // Handle error
}

// Create an auth client for refreshing tokens
authClient := auth.NewClient("client_id", "client_secret")

// Initialize the token manager
tokenManager := tokens.NewManager(tokenStorage, authClient)

// Configure refresh threshold (optional)
tokenManager.SetRefreshThreshold(10 * time.Minute)

// Start background refresh (optional)
stopRefresh := tokenManager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Call when done to stop background refresh

// Store a token
entry := &tokens.Entry{
    Resource: "user_123",
    TokenSet: &tokens.TokenSet{
        AccessToken:  "access_token",
        RefreshToken: "refresh_token",
        ExpiresAt:    time.Now().Add(1 * time.Hour),
        Scope:        "openid profile email",
    },
}
err = tokenManager.StoreToken(ctx, entry)

// Retrieve a token (will refresh automatically if needed)
entry, err = tokenManager.GetToken(ctx, "user_123")
if err != nil {
    // Handle error
}

// Use the access token
accessToken := entry.TokenSet.AccessToken
```

## Components

### TokenSet

`TokenSet` represents a set of OAuth2 tokens:

```go
type TokenSet struct {
    AccessToken  string    // The access token
    RefreshToken string    // The refresh token (may be empty)
    ExpiresAt    time.Time // When the access token expires
    Scope        string    // The scopes associated with the token
    ResourceID   string    // Optional resource ID
}
```

### Entry

`Entry` represents a token entry in storage:

```go
type Entry struct {
    Resource    string    // The resource identifier (e.g. user ID)
    AccessToken string    // The access token
    RefreshToken string   // The refresh token
    ExpiresAt   time.Time // When the access token expires
    Scope       string    // The scopes associated with the token
    TokenSet    *TokenSet // For convenience (not serialized)
}
```

### Storage Interface

```go
type Storage interface {
    Store(entry *Entry) error
    Lookup(resource string) (*Entry, error)
    Delete(resource string) error
    List() ([]string, error)
}
```

### Manager

```go
type Manager struct {
    Storage          Storage
    RefreshThreshold time.Duration
    RefreshHandler   RefreshHandler
}
```

## Implementation Notes

- The file-based storage uses the filesystem with proper file locking
- All operations are thread-safe for concurrent access
- Refresh operations use a mutex to prevent multiple simultaneous refreshes
- Background refresh runs in a goroutine and can be stopped when no longer needed
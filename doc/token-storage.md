<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Token Storage Interface

The Globus Go SDK includes a flexible token storage system that allows applications to persist OAuth2 tokens in different ways. This document explains how to use the token storage interface and its available implementations.

## Overview

The token storage system consists of:

1. A `TokenInfo` struct that contains token data
2. A `TokenStorage` interface for storing and retrieving tokens
3. Implementations of the interface (memory, file)
4. A `TokenManager` that handles token refreshing

## TokenInfo Structure

The `TokenInfo` struct contains all the information related to an OAuth2 token:

```go
type TokenInfo struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    TokenType    string    `json:"token_type,omitempty"`
    ExpiresAt    time.Time `json:"expires_at,omitempty"`
    Scope        string    `json:"scope,omitempty"`
    Resource     string    `json:"resource,omitempty"`
    ClientID     string    `json:"client_id,omitempty"`
}
```

## TokenStorage Interface

The `TokenStorage` interface defines the basic operations for token management:

```go
type TokenStorage interface {
    // StoreToken saves a token for a specific user or resource
    StoreToken(ctx context.Context, key string, token TokenInfo) error
    
    // GetToken retrieves a token for a specific user or resource
    GetToken(ctx context.Context, key string) (TokenInfo, error)
    
    // DeleteToken removes a token for a specific user or resource
    DeleteToken(ctx context.Context, key string) error
    
    // ListTokens returns all stored token keys
    ListTokens(ctx context.Context) ([]string, error)
}
```

## Available Implementations

### Memory Token Storage

The `MemoryTokenStorage` implementation stores tokens in memory. It's useful for short-lived applications or testing:

```go
storage := auth.NewMemoryTokenStorage()
```

### File Token Storage

The `FileTokenStorage` implementation persists tokens to disk in JSON format:

```go
storage, err := auth.NewFileTokenStorage("/path/to/token/directory")
if err != nil {
    // Handle error
}
```

## Using the Token Manager

The `TokenManager` provides advanced token management capabilities, including automatic token refreshing:

```go
// Create a token manager with file storage
storage, _ := auth.NewFileTokenStorage("/path/to/tokens")
refreshFunc := func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
    // Implementation of refresh logic using auth client
    return authClient.RefreshToken(ctx, token.RefreshToken)
}

manager := &auth.TokenManager{
    Storage:          storage,
    RefreshThreshold: 5 * time.Minute,  // Refresh tokens when they're 5 minutes from expiry
    RefreshFunc:      refreshFunc,
}

// Get a token, refreshing if necessary
token, err := manager.GetToken(ctx, "user1")
if err != nil {
    // Handle error
}

// Use token.AccessToken for API calls
```

## Implementing Custom Storage

You can implement the `TokenStorage` interface to create custom storage solutions, such as:

- Database storage (SQL, NoSQL)
- Encrypted storage
- Remote storage (via API calls)

Example implementation skeleton:

```go
type CustomTokenStorage struct {
    // Your fields here
}

func (s *CustomTokenStorage) StoreToken(ctx context.Context, key string, token auth.TokenInfo) error {
    // Your implementation
}

func (s *CustomTokenStorage) GetToken(ctx context.Context, key string) (auth.TokenInfo, error) {
    // Your implementation
}

func (s *CustomTokenStorage) DeleteToken(ctx context.Context, key string) error {
    // Your implementation
}

func (s *CustomTokenStorage) ListTokens(ctx context.Context) ([]string, error) {
    // Your implementation
}
```

## Best Practices

1. **Security**: Always store tokens securely, especially refresh tokens
2. **Concurrency**: The built-in implementations are thread-safe; custom implementations should be too
3. **Error Handling**: Return appropriate errors, especially for not found conditions
4. **Context Support**: Honor context cancellation for long operations
5. **Key Naming**: Use consistent key naming for tokens (e.g., user IDs, resource IDs)
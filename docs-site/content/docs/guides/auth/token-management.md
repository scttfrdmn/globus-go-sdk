---
title: "Token Management"
weight: 10
---

# Token Management and Storage

This guide explains how to effectively use the token management and storage capabilities of the Globus Go SDK.

## Overview

The `tokens` package provides a comprehensive solution for managing OAuth2 tokens in your applications. Key features include:

- **Flexible Storage Options**: Memory-based for testing or short-lived applications, file-based for persistence
- **Automatic Token Refresh**: Tokens are automatically refreshed when they're close to expiration
- **Background Refresh**: Keep tokens fresh with a background refresh process
- **Thread Safety**: All operations are thread-safe for concurrent access

## Storage Options

The SDK provides two built-in storage implementations:

### Memory Storage

Memory storage is ideal for:
- Testing environments
- Short-lived applications
- Situations where persistence isn't required

```go
// Create an in-memory token storage
storage := tokens.NewMemoryStorage()
```

### File Storage

File storage provides persistence across application restarts:
- Tokens are stored as encrypted files on disk
- Uses file-system locking for concurrent access
- Automatically handles token serialization and deserialization

```go
// Create a file-based token storage
storage, err := tokens.NewFileStorage("~/.globus-tokens")
if err != nil {
    log.Fatalf("Failed to create token storage: %v", err)
}
```

### Custom Storage

You can implement your own storage mechanism by implementing the `tokens.Storage` interface:

```go
type Storage interface {
    Store(entry *tokens.Entry) error
    Lookup(resource string) (*tokens.Entry, error)
    Delete(resource string) error
    List() ([]string, error)
}
```

This allows you to create database-backed storage or other custom implementations.

## Working with Tokens

### Token Structure

Tokens in the SDK are represented by two main types:

1. `TokenSet`: The core token data structure containing:
   - `AccessToken`: The token used for API requests
   - `RefreshToken`: Used to obtain a new access token
   - `ExpiresAt`: When the access token expires
   - `Scope`: The permissions associated with the token
   - `ResourceID`: Optional identifier for the token's resource

2. `Entry`: A storage representation wrapping a TokenSet with additional metadata:
   - `Resource`: The identifier used to lookup the token (e.g., user ID)
   - Fields mirroring TokenSet for easier serialization
   - `TokenSet`: The actual token set (not serialized)

### Storing Tokens

```go
// Create a token entry
tokenSet := &tokens.TokenSet{
    AccessToken:  "your-access-token",
    RefreshToken: "your-refresh-token",
    ExpiresAt:    time.Now().Add(1 * time.Hour),
    Scope:        "openid profile email",
}

entry := &tokens.Entry{
    Resource:     "user-123", // Identifier for this token
    AccessToken:  tokenSet.AccessToken,
    RefreshToken: tokenSet.RefreshToken,
    ExpiresAt:    tokenSet.ExpiresAt,
    Scope:        tokenSet.Scope,
    TokenSet:     tokenSet,
}

// Store the token
err := storage.Store(entry)
if err != nil {
    log.Fatalf("Failed to store token: %v", err)
}
```

### Retrieving Tokens

```go
// Look up a token by resource identifier
entry, err := storage.Lookup("user-123")
if err != nil {
    log.Fatalf("Failed to lookup token: %v", err)
}

// Use the token
fmt.Printf("Access Token: %s\n", entry.TokenSet.AccessToken)
fmt.Printf("Expires At: %s\n", entry.TokenSet.ExpiresAt.Format(time.RFC3339))
```

### Listing and Deleting Tokens

```go
// List all stored token resources
resources, err := storage.List()
if err != nil {
    log.Fatalf("Failed to list tokens: %v", err)
}

for _, resource := range resources {
    fmt.Printf("Found token for resource: %s\n", resource)
}

// Delete a token
err = storage.Delete("user-123")
if err != nil {
    log.Fatalf("Failed to delete token: %v", err)
}
```

## Token Manager

The Token Manager provides high-level functionality for token management, including automatic refreshing:

### Creating a Token Manager

```go
// Create Auth client for token refreshing
authClient, err := auth.NewClient(
    auth.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
    auth.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
)
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}

// Create token storage
storage, err := tokens.NewFileStorage("~/.globus-tokens")
if err != nil {
    log.Fatalf("Failed to create token storage: %v", err)
}

// Create token manager using the functional options pattern
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithAuthClient(authClient),
    tokens.WithRefreshThreshold(30 * time.Minute),
)
if err != nil {
    log.Fatalf("Failed to create token manager: %v", err)
}
```

### Automatic Token Refreshing

The Token Manager automatically handles token refreshing based on a configurable threshold:

```go
// Set when tokens should be refreshed (30 minutes before expiry)
manager.SetRefreshThreshold(30 * time.Minute)

// Get a token - will be automatically refreshed if needed
entry, err := manager.GetToken(context.Background(), "user-123")
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

// The returned token is guaranteed to be valid for at least the refresh threshold
fmt.Printf("Token valid until: %s\n", entry.TokenSet.ExpiresAt.Format(time.RFC3339))
```

### Background Refresh

For long-running applications, you can start a background refresh process:

```go
// Start background refresh that runs every 15 minutes
stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Stop the refresh when your application exits

// The background process will periodically check all tokens
// and refresh any that are close to expiration
```

## Common Use Cases

### Web Application Authentication

In a web application, you might want to store tokens per user:

```go
func handleCallback(w http.ResponseWriter, r *http.Request) {
    // Exchange authorization code for tokens
    code := r.URL.Query().Get("code")
    tokenResponse, err := authClient.ExchangeAuthorizationCode(
        context.Background(), 
        code,
    )
    if err != nil {
        http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
        return
    }
    
    // Store the token using the user ID as the resource identifier
    userID := getUserIDFromSession(r)
    entry := &tokens.Entry{
        Resource: userID,
        TokenSet: &tokens.TokenSet{
            AccessToken:  tokenResponse.AccessToken,
            RefreshToken: tokenResponse.RefreshToken,
            ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
            Scope:        tokenResponse.Scope,
        },
    }
    
    err = tokenManager.StoreToken(context.Background(), entry)
    if err != nil {
        http.Error(w, "Failed to store token", http.StatusInternalServerError)
        return
    }
    
    // Redirect to the application home page
    http.Redirect(w, r, "/app", http.StatusFound)
}
```

Then, in your API handlers, you can retrieve and use the tokens:

```go
func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromSession(r)
    
    // Get token (will be refreshed automatically if needed)
    entry, err := tokenManager.GetToken(context.Background(), userID)
    if err != nil {
        http.Error(w, "Failed to get token", http.StatusInternalServerError)
        return
    }
    
    // Use the token to call Globus services
    // The token is guaranteed to be valid
    accessToken := entry.TokenSet.AccessToken
    
    // ... make Globus API calls ...
}
```

### Service Account Authentication

For a service account or backend application:

```go
func main() {
    // Create Auth client
    authClient, err := auth.NewClient(
        auth.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
        auth.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    
    // Get client credentials token
    tokenResponse, err := authClient.GetClientCredentialsToken(
        context.Background(),
        []string{"https://auth.globus.org/scopes/transfer.api.globus.org/all"},
    )
    if err != nil {
        log.Fatalf("Failed to get client credentials token: %v", err)
    }
    
    // Create token storage
    storage := tokens.NewMemoryStorage()
    
    // Create token manager
    manager, err := tokens.NewManager(
        tokens.WithStorage(storage),
        tokens.WithAuthClient(authClient),
    )
    if err != nil {
        log.Fatalf("Failed to create token manager: %v", err)
    }
    
    // Store the service token
    entry := &tokens.Entry{
        Resource: "service-account",
        TokenSet: &tokens.TokenSet{
            AccessToken:  tokenResponse.AccessToken,
            RefreshToken: tokenResponse.RefreshToken,
            ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
            Scope:        tokenResponse.Scope,
        },
    }
    
    err = manager.StoreToken(context.Background(), entry)
    if err != nil {
        log.Fatalf("Failed to store token: %v", err)
    }
    
    // Start background refresh
    stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
    defer stopRefresh()
    
    // Now the application can run for a long time, and tokens will be
    // automatically refreshed as needed
    // ...
}
```

### Testing with Mock Refresh Handler

For testing, you can create a mock refresh handler:

```go
// MockRefreshHandler implements the tokens.RefreshHandler interface
type MockRefreshHandler struct {
    refreshCount int
}

func NewMockRefreshHandler() *MockRefreshHandler {
    return &MockRefreshHandler{
        refreshCount: 0,
    }
}

func (m *MockRefreshHandler) RefreshToken(_ context.Context, refreshToken string) (*auth.TokenResponse, error) {
    m.refreshCount++
    
    // Return a mock token response
    return &auth.TokenResponse{
        AccessToken:  fmt.Sprintf("mock-access-token-%d", m.refreshCount),
        RefreshToken: fmt.Sprintf("mock-refresh-token-%d", m.refreshCount),
        ExpiresIn:    3600,
        ExpiryTime:   time.Now().Add(1 * time.Hour),
        TokenType:    "Bearer",
        Scope:        "openid profile email",
    }, nil
}

// Usage
func TestTokenManager() {
    mockHandler := NewMockRefreshHandler()
    storage := tokens.NewMemoryStorage()
    
    manager, _ := tokens.NewManager(
        tokens.WithStorage(storage),
        tokens.WithRefreshHandler(mockHandler),
    )
    
    // Continue testing with the mock handler
}
```

## Best Practices

1. **Choose the right storage** for your use case:
   - Memory storage for tests and short-lived applications
   - File storage for most applications that require persistence

2. **Set an appropriate refresh threshold** based on your application's needs:
   - Shorter for critical applications (e.g., 30 minutes)
   - Longer for less critical applications (e.g., 2 hours)

3. **Use background refresh** for long-running applications:
   - Always call the returned stop function with `defer` to prevent leaks
   - Choose a refresh interval based on how many tokens you store

4. **Handle errors gracefully**:
   - Have fallback mechanisms for when token refresh fails
   - Consider implementing retry logic for transient failures

5. **Secure your tokens**:
   - Never log access or refresh tokens
   - Ensure file storage is on a secure filesystem
   - Set appropriate permissions for token files

6. **Clean up unused tokens**:
   - Implement a mechanism to periodically remove unused or expired tokens
   - Use the `storage.List()` and `storage.Delete()` methods for cleanup

## Complete Example

Here's a complete example showing how to use token management and storage:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // Get credentials from environment variables
    clientID := os.Getenv("GLOBUS_CLIENT_ID")
    clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")

    if clientID == "" || clientSecret == "" {
        log.Fatal("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET must be set")
    }

    // Create Auth client
    authClient, err := auth.NewClient(
        auth.WithClientID(clientID),
        auth.WithClientSecret(clientSecret),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }

    // Create file storage
    storage, err := tokens.NewFileStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }

    // Create token manager
    manager, err := tokens.NewManager(
        tokens.WithStorage(storage),
        tokens.WithAuthClient(authClient),
        tokens.WithRefreshThreshold(30 * time.Minute),
    )
    if err != nil {
        log.Fatalf("Failed to create token manager: %v", err)
    }

    // Check if we already have a token
    entry, err := manager.GetToken(ctx, "default")
    if err == nil && entry != nil && !entry.TokenSet.IsExpired() {
        fmt.Println("Found existing token!")
        fmt.Printf("Access Token: %s...\n", entry.TokenSet.AccessToken[:10])
        fmt.Printf("Expires At: %s\n", entry.TokenSet.ExpiresAt.Format(time.RFC3339))
    } else {
        fmt.Println("No valid token found. Let's get a new one.")

        // For a CLI application, you might use the device code flow:
        deviceCode, err := authClient.GetDeviceCode(ctx, []string{
            "openid", "profile", "email",
            "urn:globus:auth:scope:transfer.api.globus.org:all",
        })
        if err != nil {
            log.Fatalf("Failed to get device code: %v", err)
        }

        fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
        fmt.Printf("And enter the code: %s\n", deviceCode.UserCode)
        fmt.Println("Waiting for you to authenticate...")

        // Poll for the token
        tokenResponse, err := authClient.PollForTokens(ctx, deviceCode)
        if err != nil {
            log.Fatalf("Failed to get token: %v", err)
        }

        // Create and store the token
        entry = &tokens.Entry{
            Resource: "default",
            TokenSet: &tokens.TokenSet{
                AccessToken:  tokenResponse.AccessToken,
                RefreshToken: tokenResponse.RefreshToken,
                ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
                Scope:        tokenResponse.Scope,
            },
        }

        err = manager.StoreToken(ctx, entry)
        if err != nil {
            log.Fatalf("Failed to store token: %v", err)
        }

        fmt.Println("Token obtained and stored successfully!")
    }

    // Start background refresh
    fmt.Println("Starting background refresh...")
    stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
    defer stopRefresh()

    // Now you can use the token for API calls...
    // For example, create a transfer client:
    // transferClient, err := transfer.NewClient(
    //     transfer.WithAccessToken(entry.TokenSet.AccessToken),
    // )

    // Keep the application running
    fmt.Println("Application running. Press Ctrl+C to exit.")
    select {} // Wait forever
}
```

## Next Steps

Now that you understand token management, you might want to explore:

- [Authentication Flows](../authentication-flows.md) - Different ways to authenticate with Globus
- [Multi-Factor Authentication](../multi-factor-authentication.md) - How to handle MFA requirements
- [Token Security](../token-security.md) - Best practices for token security
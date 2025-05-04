---
title: "Auth Service Quick Start"
weight: 10
---

# Auth Service Quick Start

This guide walks you through basic authentication operations using the Globus Go SDK's Auth service. The Auth service is the foundation for interacting with Globus, handling authentication, token management, and user information.

## Setup

First, import the required packages and create a context:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Continue with the examples below...
}
```

## Creating an Auth Client

There are two main ways to create an Auth client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Auth client
authClient, err := config.NewAuthClient()
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Auth client with options
authClient, err := auth.NewClient(
    auth.WithClientID(os.Getenv("GLOBUS_CLIENT_ID")),
    auth.WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET")),
    auth.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}
```

## OAuth2 Authorization Code Flow

The authorization code flow is used for web applications where a user needs to grant permission:

```go
// Set the redirect URL for the OAuth2 flow
authClient.SetRedirectURL("http://localhost:8080/callback")

// Generate the authorization URL
state := "my-random-state"
authURL := authClient.GetAuthorizationURL(state)

fmt.Printf("Visit this URL to log in: %s\n", authURL)

// In a real application, you would set up a web server to handle the callback
// Here's a simplified example:
http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
    // Verify the state parameter
    if r.URL.Query().Get("state") != state {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    
    // Exchange the authorization code for tokens
    code := r.URL.Query().Get("code")
    tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
        return
    }
    
    fmt.Printf("Received tokens: %+v\n", tokenResponse)
    
    // Use the tokens to access Globus services
    // Store the tokens for future use
    
    fmt.Fprintf(w, "Authentication successful! You can close this window.")
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

## Client Credentials Flow

The client credentials flow is used for server-to-server authentication without user interaction:

```go
// Get tokens using client credentials
tokenResponse, err := authClient.GetClientCredentialsTokens(ctx, auth.AllScopes)
if err != nil {
    log.Fatalf("Failed to get tokens: %v", err)
}

fmt.Printf("Access Token: %s\n", tokenResponse.AccessToken)
fmt.Printf("Expires In: %d seconds\n", tokenResponse.ExpiresIn)
```

## Token Management

The SDK provides a token manager for storing and refreshing tokens:

```go
// Create a memory-based token storage
storage := tokens.NewMemoryStorage()

// Create a token manager
tokenManager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithRefreshHandler(authClient),
)
if err != nil {
    log.Fatalf("Failed to create token manager: %v", err)
}

// Store tokens
entry := &tokens.Entry{
    Resource: "my-service",
    TokenSet: &tokens.TokenSet{
        AccessToken:  tokenResponse.AccessToken,
        RefreshToken: tokenResponse.RefreshToken,
        ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
        Scope:        tokenResponse.Scope,
    },
}

err = tokenManager.StoreToken(ctx, entry)
if err != nil {
    log.Fatalf("Failed to store token: %v", err)
}

// Later, retrieve the token
retrievedEntry, err := tokenManager.GetToken(ctx, "my-service")
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

// Check if the token is expired and refresh if needed
if retrievedEntry.TokenSet.IsExpired() {
    refreshedEntry, err := tokenManager.RefreshToken(ctx, retrievedEntry)
    if err != nil {
        log.Fatalf("Failed to refresh token: %v", err)
    }
    
    fmt.Printf("Token refreshed! New expiry: %s\n", refreshedEntry.TokenSet.ExpiresAt)
}
```

## Automatic Token Refresh

You can set up automatic token refresh in the background:

```go
// Configure refresh settings
tokenManager.SetRefreshThreshold(5 * time.Minute) // Refresh tokens when they're within 5 minutes of expiring

// Start background refresh that checks tokens every 15 minutes
stopRefresh := tokenManager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Stop background refresh when done
```

## Getting User Information

Once you have an access token, you can get information about the authenticated user:

```go
// Create an auth client with a valid access token
authClient, err := auth.NewClient(
    auth.WithAccessToken(accessToken),
)
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}

// Get user information
userInfo, err := authClient.GetUserInfo(ctx)
if err != nil {
    log.Fatalf("Failed to get user info: %v", err)
}

fmt.Printf("User ID: %s\n", userInfo.SubjectID)
fmt.Printf("Username: %s\n", userInfo.Username)
fmt.Printf("Email: %s\n", userInfo.Email)
fmt.Printf("Name: %s %s\n", userInfo.Name, userInfo.LastName)
```

## Validating Tokens

You can validate tokens to ensure they are still valid:

```go
// Introspect a token to check its validity
introspectResult, err := authClient.IntrospectToken(ctx, accessToken)
if err != nil {
    log.Fatalf("Failed to introspect token: %v", err)
}

if introspectResult.Active {
    fmt.Println("Token is active and valid")
    fmt.Printf("Scopes: %s\n", introspectResult.Scope)
    fmt.Printf("Expires: %d\n", introspectResult.Exp)
} else {
    fmt.Println("Token is no longer valid")
}
```

## Revoking Tokens

When you're done with a token, you should revoke it:

```go
// Revoke a token
err = authClient.RevokeToken(ctx, accessToken)
if err != nil {
    log.Fatalf("Failed to revoke token: %v", err)
}

fmt.Println("Token successfully revoked")
```

## Complete Example

Here's a complete example that combines several operations:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

func main() {
    // Create context
    ctx := context.Background()
    
    // Create SDK configuration from environment
    config := pkg.NewConfigFromEnvironment()
    
    // Create auth client
    authClient, err := config.NewAuthClient()
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    
    // Get tokens using client credentials
    tokenResponse, err := authClient.GetClientCredentialsTokens(ctx, auth.AllScopes)
    if err != nil {
        log.Fatalf("Failed to get tokens: %v", err)
    }
    
    fmt.Printf("Access Token: %s (truncated for security)\n", tokenResponse.AccessToken[:10]+"...")
    fmt.Printf("Expires In: %d seconds\n", tokenResponse.ExpiresIn)
    
    // Create token storage and manager
    storage := tokens.NewMemoryStorage()
    tokenManager, err := tokens.NewManager(
        tokens.WithStorage(storage),
        tokens.WithRefreshHandler(authClient),
    )
    if err != nil {
        log.Fatalf("Failed to create token manager: %v", err)
    }
    
    // Store the token
    entry := &tokens.Entry{
        Resource: "default",
        TokenSet: &tokens.TokenSet{
            AccessToken:  tokenResponse.AccessToken,
            RefreshToken: tokenResponse.RefreshToken,
            ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
            Scope:        tokenResponse.Scope,
        },
    }
    
    err = tokenManager.StoreToken(ctx, entry)
    if err != nil {
        log.Fatalf("Failed to store token: %v", err)
    }
    
    // Get user information
    authClientWithToken, err := auth.NewClient(
        auth.WithAccessToken(tokenResponse.AccessToken),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client with token: %v", err)
    }
    
    userInfo, err := authClientWithToken.GetUserInfo(ctx)
    if err != nil {
        log.Fatalf("Failed to get user info: %v", err)
    }
    
    fmt.Printf("User ID: %s\n", userInfo.SubjectID)
    fmt.Printf("Username: %s\n", userInfo.Username)
    fmt.Printf("Email: %s\n", userInfo.Email)
    
    // Clean up - revoke the token when done
    err = authClientWithToken.RevokeToken(ctx, tokenResponse.AccessToken)
    if err != nil {
        log.Fatalf("Failed to revoke token: %v", err)
    }
    
    fmt.Println("Token successfully revoked")
}
```

## Next Steps

Now that you understand the basics of authentication with the Globus Go SDK, you can:

1. **Set up token storage**: Implement persistent token storage using `tokens.FileStorage`
2. **Implement Web Authentication**: Build a complete web application with the authorization code flow
3. **Connect to other services**: Use your tokens to authenticate with other Globus services

Check out the [Auth Service API Reference](/docs/reference/auth/) for comprehensive documentation on all Auth service methods and options.
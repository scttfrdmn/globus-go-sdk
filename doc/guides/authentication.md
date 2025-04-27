# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Authentication Guide

_Last Updated: April 27, 2025_
_Compatible with SDK versions: v0.1.0 and above_

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations.

This guide covers authentication and authorization with the Globus Auth service, which is the foundation for accessing all Globus services.

## Table of Contents

- [Overview](#overview)
- [Authentication Flows](#authentication-flows)
  - [Authorization Code Flow](#authorization-code-flow)
  - [Client Credentials Flow](#client-credentials-flow)
  - [Refresh Token Flow](#refresh-token-flow)
- [Token Management](#token-management)
  - [Token Storage](#token-storage)
  - [Automatic Token Refresh](#automatic-token-refresh)
  - [Token Validation](#token-validation)
  - [Token Revocation](#token-revocation)
- [Scopes](#scopes)
- [Multi-Factor Authentication](#multi-factor-authentication)
- [Advanced Features](#advanced-features)
  - [Token Introspection](#token-introspection)
  - [Client Registration](#client-registration)
  - [Dependent Tokens](#dependent-tokens)
- [Error Handling](#error-handling)
- [Examples](#examples)
  - [Authorization Code Example](#authorization-code-example)
  - [Client Credentials Example](#client-credentials-example)
  - [Token Management Example](#token-management-example)
- [Troubleshooting](#troubleshooting)
- [Related Topics](#related-topics)

## Overview

The Globus Auth service provides OAuth 2.0-based authentication and authorization for all Globus services. The Globus Go SDK offers a comprehensive client for interacting with Auth, supporting various authentication flows, token management, and advanced features.

## Authentication Flows

### Authorization Code Flow

The Authorization Code flow is used for applications where a user must interactively authenticate:

1. Generate an authorization URL:

```go
// Create an Auth client
authClient := pkg.NewAuthClient()
authClient.SetRedirectURL("http://localhost:8000/callback")

// Generate the URL
state := generateRandomState() // For CSRF protection
authURL := authClient.GetAuthorizationURL(
    state,
    pkg.TransferScope, // Request transfer scope
    pkg.GroupsScope,   // Request groups scope
)

// Redirect the user to this URL
fmt.Printf("Visit this URL to log in: %s\n", authURL)
```

2. Handle the callback and exchange the code:

```go
// In your HTTP callback handler
http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
    // Get the authorization code from the query
    code := r.URL.Query().Get("code")
    receivedState := r.URL.Query().Get("state")
    
    // Verify state to prevent CSRF
    if receivedState != state {
        http.Error(w, "Invalid state parameter", http.StatusBadRequest)
        return
    }
    
    // Exchange the code for tokens
    tokenResp, err := authClient.ExchangeAuthorizationCode(
        context.Background(),
        code,
    )
    if err != nil {
        http.Error(w, "Failed to exchange code: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Store the tokens
    // ...
    
    fmt.Fprintf(w, "Authentication successful! You can close this window.")
})
```

### Client Credentials Flow

The Client Credentials flow is for machine-to-machine communication without user interaction:

```go
// Create an Auth client with client credentials
authClient := pkg.NewAuthClient(
    os.Getenv("GLOBUS_CLIENT_ID"),
    os.Getenv("GLOBUS_CLIENT_SECRET"),
)

// Get a token with specified scopes
tokenResp, err := authClient.GetClientCredentialsToken(
    context.Background(),
    []string{
        "urn:globus:auth:scope:transfer.api.globus.org:all",
        "urn:globus:auth:scope:groups.api.globus.org:all",
    },
)
if err != nil {
    log.Fatalf("Failed to get client credentials token: %v", err)
}

// Use the access token
accessToken := tokenResp.AccessToken
```

### Refresh Token Flow

When you have a refresh token, you can use it to get a new access token:

```go
// Create an Auth client
authClient := pkg.NewAuthClient()

// Refresh the token
newTokenResp, err := authClient.RefreshToken(
    context.Background(),
    refreshToken,
)
if err != nil {
    log.Fatalf("Failed to refresh token: %v", err)
}

// Use the new access token
newAccessToken := newTokenResp.AccessToken
```

## Token Management

### Token Storage

The SDK provides interfaces and implementations for storing tokens securely:

```go
// Create a file-based token storage
storage, err := auth.NewFileTokenStorage("~/.globus-tokens")
if err != nil {
    log.Fatalf("Failed to create token storage: %v", err)
}

// Store a token
err = storage.StoreToken(
    context.Background(),
    "default",  // A name for this token
    tokenResp,  // The token response from authentication
)
if err != nil {
    log.Fatalf("Failed to store token: %v", err)
}

// Retrieve a token
token, err := storage.GetToken(context.Background(), "default")
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}
```

For more details, see [Token Storage](../topics/token-storage.md).

### Automatic Token Refresh

The SDK provides a token manager that automatically refreshes tokens before they expire:

```go
// Create a token manager
tokenManager := &auth.TokenManager{
    Storage:          storage,
    RefreshThreshold: 5 * time.Minute,  // Refresh tokens 5 minutes before expiry
    RefreshFunc: func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
        return authClient.RefreshToken(ctx, token.RefreshToken)
    },
}

// Get a token (will be refreshed if needed)
token, err := tokenManager.GetToken(context.Background(), "default")
if err != nil {
    log.Fatalf("Failed to get token: %v", err)
}

// Use the token with confidence it's valid
accessToken := token.AccessToken
```

### Token Validation

Validate tokens to ensure they're still usable:

```go
// Check if a token is expired
if token.IsExpired() {
    fmt.Println("Token has expired")
}

// Check how much time is left before expiry
lifetime := token.Lifetime()
fmt.Printf("Token expires in %v\n", lifetime)

// Get detailed information about a token
tokenInfo, err := authClient.IntrospectToken(
    context.Background(),
    token.AccessToken,
)
if err != nil {
    log.Fatalf("Failed to introspect token: %v", err)
}

if !tokenInfo.Active {
    fmt.Println("Token is no longer active")
}
```

### Token Revocation

Revoke tokens when they're no longer needed:

```go
// Revoke a token
err := authClient.RevokeToken(
    context.Background(),
    token.AccessToken,
)
if err != nil {
    log.Fatalf("Failed to revoke token: %v", err)
}
```

## Scopes

Scopes define the permissions granted to an application. The SDK provides constants for common scopes:

```go
// Common scopes
pkg.OpenIDScope      // "openid"
pkg.EmailScope       // "email"
pkg.ProfileScope     // "profile"
pkg.TransferScope    // "urn:globus:auth:scope:transfer.api.globus.org:all"
pkg.GroupsScope      // "urn:globus:auth:scope:groups.api.globus.org:all"
pkg.SearchScope      // "urn:globus:auth:scope:search.api.globus.org:all"
pkg.FlowsScope       // "urn:globus:auth:scope:flows.globus.org:all"
```

Request only the scopes your application needs:

```go
// For a transfer-only application
authURL := authClient.GetAuthorizationURL(state, pkg.TransferScope)

// For an application that needs multiple services
authURL := authClient.GetAuthorizationURL(
    state,
    pkg.TransferScope,
    pkg.GroupsScope,
    pkg.OpenIDScope,
    pkg.ProfileScope,
)
```

## Multi-Factor Authentication

The SDK supports multi-factor authentication (MFA) for enhanced security:

```go
// Exchange code with MFA support
tokenResp, err := authClient.ExchangeAuthorizationCodeWithMFA(
    context.Background(),
    code,
    func(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
        // Prompt the user for the MFA code
        fmt.Printf("Enter the code from your %s: ", challenge.Type)
        var code string
        fmt.Scanln(&code)
        
        return &auth.MFAResponse{
            ChallengeID: challenge.ChallengeID,
            Type:        challenge.Type,
            Value:       code,
        }, nil
    },
)
```

For more details, see [Multi-Factor Authentication](../advanced/mfa.md).

## Advanced Features

### Token Introspection

Introspect tokens to get detailed information:

```go
// Introspect a token
info, err := authClient.IntrospectToken(
    context.Background(),
    accessToken,
)
if err != nil {
    log.Fatalf("Failed to introspect token: %v", err)
}

fmt.Printf("Token info: active=%v, scopes=%v, exp=%v\n",
    info.Active, info.Scope, time.Unix(info.Exp, 0))
```

### Client Registration

Register a new client dynamically:

```go
// Register a new client
client, err := authClient.RegisterClient(
    context.Background(),
    &auth.ClientRegistrationRequest{
        ClientName:    "My New App",
        RedirectURIs:  []string{"https://myapp.example.com/callback"},
        Scopes:        []string{pkg.TransferScope, pkg.GroupsScope},
        RedirectTypes: []string{"native_app", "web_application"},
    },
)
if err != nil {
    log.Fatalf("Failed to register client: %v", err)
}

fmt.Printf("Client registered with ID: %s\n", client.ClientID)
```

### Dependent Tokens

Get tokens for dependent services:

```go
// Get dependent tokens
dependentTokens, err := authClient.GetDependentTokens(
    context.Background(),
    accessToken,
    []string{pkg.TransferScope, pkg.GroupsScope},
)
if err != nil {
    log.Fatalf("Failed to get dependent tokens: %v", err)
}

// Use the dependent tokens
transferToken := dependentTokens["transfer.api.globus.org"].AccessToken
groupsToken := dependentTokens["groups.api.globus.org"].AccessToken
```

## Error Handling

The Auth client defines specific error types:

```go
// Check for specific error types
if err != nil {
    switch {
    case auth.IsInvalidGrantError(err):
        fmt.Println("Invalid grant (code or refresh token)")
    case auth.IsInvalidClientError(err):
        fmt.Println("Invalid client credentials")
    case auth.IsInvalidScopeError(err):
        fmt.Println("Invalid scope requested")
    case auth.IsUnsupportedGrantTypeError(err):
        fmt.Println("Unsupported grant type")
    case auth.IsAccessDeniedError(err):
        fmt.Println("Access denied")
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## Examples

### Authorization Code Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

func main() {
    // Create a token storage for persisting tokens
    storage, err := auth.NewFileTokenStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }
    
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))

    // Create a new Auth client
    authClient := config.NewAuthClient()
    authClient.SetRedirectURL("http://localhost:8080/callback")
    
    // Create a token manager for automatic refresh
    tokenManager := &auth.TokenManager{
        Storage:          storage,
        RefreshThreshold: 5 * time.Minute,
        RefreshFunc: func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
            return authClient.RefreshToken(ctx, token.RefreshToken)
        },
    }
    
    // Check if we already have tokens
    token, err := tokenManager.GetToken(context.Background(), "default")
    if err == nil && !token.IsExpired() {
        // We have valid tokens, use them
        fmt.Printf("Using existing token (expires in %v)\n", token.Lifetime())
        return
    }
    
    // We need new tokens, start the OAuth2 flow
    state := fmt.Sprintf("%d", time.Now().UnixNano())
    authURL := authClient.GetAuthorizationURL(
        state, 
        pkg.TransferScope,
        pkg.GroupsScope,
    )
    fmt.Printf("Visit this URL to log in: %s\n", authURL)
    
    // Start a local server to handle the callback
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        receivedState := r.URL.Query().Get("state")
        
        // Verify state
        if receivedState != state {
            http.Error(w, "Invalid state parameter", http.StatusBadRequest)
            return
        }
        
        // Exchange code for tokens
        tokenResponse, err := authClient.ExchangeAuthorizationCode(
            context.Background(), 
            code,
        )
        if err != nil {
            log.Printf("Failed to exchange code: %v", err)
            http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
            return
        }
        
        // Store the tokens
        err = tokenManager.StoreToken(context.Background(), "default", tokenResponse)
        if err != nil {
            log.Printf("Failed to store token: %v", err)
            http.Error(w, "Failed to store token", http.StatusInternalServerError)
            return
        }
        
        fmt.Fprintf(w, "Authentication successful! You can close this window.")
        fmt.Printf("Authentication successful! Tokens stored.\n")
    })
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Client Credentials Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
)

func main() {
    // Create a configuration with client credentials
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))
    
    // Create an Auth client
    authClient := config.NewAuthClient()
    
    // Get a token with client credentials
    tokenResp, err := authClient.GetClientCredentialsToken(
        context.Background(),
        []string{pkg.TransferScope},
    )
    if err != nil {
        log.Fatalf("Failed to get client credentials token: %v", err)
    }
    
    // Create a Transfer client with the token
    transferClient := config.NewTransferClient(tokenResp.AccessToken)
    
    // Use the Transfer client
    endpoints, err := transferClient.ListEndpoints(context.Background(), nil)
    if err != nil {
        log.Fatalf("Failed to list endpoints: %v", err)
    }
    
    fmt.Printf("Found %d endpoints\n", len(endpoints.DATA))
    for i, endpoint := range endpoints.DATA {
        fmt.Printf("%d. %s (%s)\n", i+1, endpoint.DisplayName, endpoint.ID)
    }
}
```

### Token Management Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

func main() {
    // Create a token storage
    storage, err := auth.NewFileTokenStorage("~/.globus-tokens")
    if err != nil {
        log.Fatalf("Failed to create token storage: %v", err)
    }
    
    // Create a configuration
    config := pkg.NewConfigFromEnvironment().
        WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
        WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))
    
    // Create an Auth client
    authClient := config.NewAuthClient()
    
    // Create a token manager
    tokenManager := &auth.TokenManager{
        Storage:          storage,
        RefreshThreshold: 5 * time.Minute,
        RefreshFunc: func(ctx context.Context, token auth.TokenInfo) (auth.TokenInfo, error) {
            return authClient.RefreshToken(ctx, token.RefreshToken)
        },
    }
    
    // Get a stored token
    tokenName := "my-token"
    token, err := tokenManager.GetToken(context.Background(), tokenName)
    if err != nil {
        if auth.IsTokenNotFoundError(err) {
            fmt.Printf("Token '%s' not found, please authenticate first\n", tokenName)
        } else {
            log.Fatalf("Error getting token: %v", err)
        }
        return
    }
    
    // Check if the token is valid
    if token.IsExpired() {
        fmt.Println("Token has expired, refreshing...")
        token, err = tokenManager.RefreshToken(context.Background(), tokenName)
        if err != nil {
            log.Fatalf("Failed to refresh token: %v", err)
        }
        fmt.Println("Token refreshed successfully")
    }
    
    // Display token information
    fmt.Printf("Token: %s...\n", token.AccessToken[:10])
    fmt.Printf("Expires in: %v\n", token.Lifetime())
    
    // List all tokens
    tokens, err := storage.ListTokens(context.Background())
    if err != nil {
        log.Fatalf("Failed to list tokens: %v", err)
    }
    
    fmt.Printf("Found %d tokens in storage:\n", len(tokens))
    for _, name := range tokens {
        t, err := storage.GetToken(context.Background(), name)
        if err != nil {
            fmt.Printf("- %s (error: %v)\n", name, err)
            continue
        }
        fmt.Printf("- %s (expires in %v)\n", name, t.Lifetime())
    }
}
```

## Troubleshooting

### Common Issues

1. **Invalid Client Credentials**

   ```
   Error: invalid_client
   ```

   **Solution**:
   - Double-check your client ID and client secret
   - Ensure the client has been registered correctly
   - Verify the client has the necessary scopes

2. **Invalid Grant**

   ```
   Error: invalid_grant
   ```

   **Solution**:
   - For authorization code flow: verify the code hasn't expired (they are short-lived)
   - For refresh token flow: verify the refresh token hasn't been revoked
   - Check if the token has been used already (authorization codes are one-time use)

3. **Access Denied**

   ```
   Error: access_denied
   ```

   **Solution**:
   - The user declined to authorize your application
   - The requested scopes exceed what the client is allowed to request

4. **Unsupported Grant Type**

   ```
   Error: unsupported_grant_type
   ```

   **Solution**:
   - Verify your client is allowed to use the requested grant type
   - Check if you're using the correct endpoint for the grant type

## Related Topics

- [Token Storage](../topics/token-storage.md)
- [Multi-Factor Authentication](../advanced/mfa.md)
- [Error Handling](../topics/error-handling.md)
- [Transfer Guide](transfer.md)
- [Groups Guide](groups.md)
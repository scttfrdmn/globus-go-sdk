# Auth Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Auth client provides access to the Globus Auth API, which handles authentication and authorization for all Globus services. It supports various OAuth2 flows, token management, and multi-factor authentication.

## Client Structure

```go
type Client struct {
    client      *core.Client
    transport   *core.HTTPTransport
    clientID    string
    clientSecret string
    redirectURL string
}
```

| Field | Type | Description |
|-------|------|-------------|
| `client` | `*core.Client` | Core client for making HTTP requests |
| `transport` | `*core.HTTPTransport` | HTTP transport for request/response handling |
| `clientID` | `string` | OAuth2 client ID |
| `clientSecret` | `string` | OAuth2 client secret |
| `redirectURL` | `string` | OAuth2 redirect URL |

## Creating an Auth Client

```go
// Create an auth client with options
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithClientSecret("client-secret"),
    auth.WithRedirectURL("https://example.com/callback"),
)
if err != nil {
    // Handle error
}
```

### Options

| Option | Description |
|--------|-------------|
| `WithClientID(id string)` | Sets the client ID (required for some flows) |
| `WithClientSecret(secret string)` | Sets the client secret (required for some flows) |
| `WithRedirectURL(url string)` | Sets the redirect URL (required for authorization code flow) |
| `WithBaseURL(url string)` | Sets a custom base URL (default: "https://auth.globus.org/v2/") |
| `WithAuthorizer(auth core.Authorizer)` | Sets an authorizer for the client |
| `WithHTTPDebugging()` | Enables HTTP debugging |
| `WithHTTPTracing()` | Enables HTTP tracing |

## Authorization Code Flow

The authorization code flow is used for web applications and native applications where a user needs to authenticate:

### Getting an Authorization URL

```go
// Generate an authorization URL
url := client.GetAuthorizationURL([]string{
    auth.ScopeOpenID,
    auth.ScopeProfile,
    auth.ScopeEmail,
    auth.ScopeOfflineAccess,
})

// Redirect the user to this URL or display it
fmt.Println("Visit this URL to authenticate:", url)
```

### Exchanging an Authorization Code

```go
// Exchange the code for tokens
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, "authorization-code")
if err != nil {
    if auth.IsMFAError(err) {
        // Handle MFA required
        challenge := auth.GetMFAChallenge(err)
        // Prompt user for MFA response
        // ...
        tokenResponse, err = client.ExchangeAuthorizationCodeWithMFA(
            ctx, 
            "authorization-code", 
            challenge.ChallengeID, 
            "mfa-response",
        )
    } else {
        // Handle other errors
    }
}

// Use the tokens
accessToken := tokenResponse.AccessToken
refreshToken := tokenResponse.RefreshToken
expiresIn := tokenResponse.ExpiresIn
```

## Client Credentials Flow

The client credentials flow is used for server-to-server applications:

```go
// Get a token using client credentials
tokenResponse, err := client.GetClientCredentialsToken(ctx, []string{
    "https://auth.globus.org/scopes/example-resource-server/scope",
})
if err != nil {
    // Handle error
}

// Use the token
accessToken := tokenResponse.AccessToken
```

## Token Refresh

```go
// Refresh a token
tokenResponse, err := client.RefreshToken(ctx, "refresh-token")
if err != nil {
    if auth.IsMFAError(err) {
        // Handle MFA required
        challenge := auth.GetMFAChallenge(err)
        // Prompt user for MFA response
        // ...
        tokenResponse, err = client.RefreshTokenWithMFA(
            ctx, 
            "refresh-token", 
            challenge.ChallengeID, 
            "mfa-response",
        )
    } else {
        // Handle other errors
    }
}

// Use the refreshed token
newAccessToken := tokenResponse.AccessToken
```

## Token Introspection

```go
// Introspect a token
tokenInfo, err := client.IntrospectToken(ctx, "access-token")
if err != nil {
    // Handle error
}

// Check if the token is active
if tokenInfo.IsActive() {
    // Token is valid
    fmt.Println("Token expires at:", tokenInfo.ExpiresAt())
} else {
    // Token is invalid or expired
    fmt.Println("Token is invalid or expired")
}
```

## Token Revocation

```go
// Revoke a token
err := client.RevokeToken(ctx, "token", "access_token")
if err != nil {
    // Handle error
}
```

## MFA Authentication

Globus Auth supports multi-factor authentication (MFA) for additional security:

### Handling MFA Challenges

```go
// Try to exchange code
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, "code")
if err != nil {
    if auth.IsMFAError(err) {
        // Get the MFA challenge
        challenge := auth.GetMFAChallenge(err)
        
        // Display the challenge to the user
        fmt.Println("MFA Required:", challenge.Prompt)
        
        // Get the user's response
        var response string
        fmt.Print("Enter MFA code: ")
        fmt.Scanln(&response)
        
        // Exchange with MFA
        tokenResponse, err = client.ExchangeAuthorizationCodeWithMFA(
            ctx,
            "code",
            challenge.ChallengeID,
            response,
        )
        if err != nil {
            // Handle error
        }
    } else {
        // Handle other errors
    }
}
```

## User Information

```go
// Get information about the user associated with a token
userInfo, err := client.GetUserInfo(ctx, "access-token")
if err != nil {
    // Handle error
}

// Use the user info
fmt.Println("User:", userInfo.Name)
fmt.Println("Email:", userInfo.Email)
fmt.Println("Identity:", userInfo.Sub)
```

## Creating Authorizers

The auth client provides methods for creating authorizers that can be used with other service clients:

### Static Token Authorizer

```go
// Create a static token authorizer
authorizer := client.CreateStaticTokenAuthorizer("access-token")

// Use with another service client
transferClient, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
if err != nil {
    // Handle error
}
```

### Client Credentials Authorizer

```go
// Create a client credentials authorizer
authorizer, err := client.CreateClientCredentialsAuthorizer(ctx, []string{
    "https://auth.globus.org/scopes/transfer.api.globus.org/all",
})
if err != nil {
    // Handle error
}

// Use with another service client
transferClient, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
if err != nil {
    // Handle error
}
```

### Refreshable Token Authorizer

```go
// Create a refreshable token authorizer
authorizer, err := client.CreateRefreshableTokenAuthorizer(ctx, "refresh-token")
if err != nil {
    // Handle error
}

// Use with another service client
transferClient, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
if err != nil {
    // Handle error
}
```

## Error Handling

The auth package provides specific error types and helper functions for common error conditions:

### Error Constants

```go
// Check for specific error types
if err != nil {
    switch {
    case auth.IsInvalidGrant(err):
        fmt.Println("Invalid grant (e.g., expired code or refresh token)")
    case auth.IsInvalidClient(err):
        fmt.Println("Invalid client credentials")
    case auth.IsInvalidScope(err):
        fmt.Println("Invalid scope")
    case auth.IsAccessDenied(err):
        fmt.Println("Access denied")
    case auth.IsServerError(err):
        fmt.Println("Server error")
    case auth.IsUnauthorized(err):
        fmt.Println("Unauthorized")
    case auth.IsBadRequest(err):
        fmt.Println("Bad request")
    case auth.IsMFAError(err):
        fmt.Println("MFA required")
    default:
        fmt.Println("Unknown error:", err)
    }
}
```

### Error Details

```go
// Get more details from an auth error
if authErr, ok := err.(*auth.AuthError); ok {
    fmt.Println("Error code:", authErr.Code)
    fmt.Println("Error description:", authErr.Description)
    fmt.Println("HTTP status:", authErr.StatusCode)
}
```

## Token Response Structure

```go
type TokenResponse struct {
    AccessToken      string `json:"access_token"`
    RefreshToken     string `json:"refresh_token,omitempty"`
    ExpiresIn        int    `json:"expires_in"`
    ResourceServer   string `json:"resource_server,omitempty"`
    TokenType        string `json:"token_type"`
    Scope            string `json:"scope"`
    OtherTokens      map[string]interface{} `json:"other_tokens,omitempty"`
    DependentTokens  map[string]interface{} `json:"dependent_tokens,omitempty"`
}
```

### Token Response Methods

#### HasRefreshToken

```go
// Check if a token response has a refresh token
if tokenResponse.HasRefreshToken() {
    // Refresh token is available
}
```

#### IsValid

```go
// Check if a token response is valid
if tokenResponse.IsValid() {
    // Token is valid
}
```

#### GetOtherTokens

```go
// Get other tokens from a token response
otherTokens := tokenResponse.GetOtherTokens()
for rs, token := range otherTokens {
    fmt.Println("Resource server:", rs)
    fmt.Println("Token:", token)
}
```

#### GetDependentTokens

```go
// Get dependent tokens from a token response
dependentTokens := tokenResponse.GetDependentTokens()
for rs, token := range dependentTokens {
    fmt.Println("Resource server:", rs)
    fmt.Println("Token:", token)
}
```

## Common Patterns

### Web Application Flow

```go
// Step 1: Redirect user to auth URL
authURL := client.GetAuthorizationURL([]string{
    auth.ScopeOpenID,
    auth.ScopeProfile,
    auth.ScopeEmail,
    auth.ScopeOfflineAccess,
})
// Redirect to authURL

// Step 2: Handle the callback
// code = URL parameter from callback

// Step 3: Exchange code for tokens
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    // Handle error
}

// Step 4: Store tokens
// ... store tokenResponse.AccessToken, tokenResponse.RefreshToken

// Step 5: Use tokens
// ... use tokenResponse.AccessToken

// Step 6: Refresh when needed
newTokenResponse, err := client.RefreshToken(ctx, tokenResponse.RefreshToken)
if err != nil {
    // Handle error
}
// ... use newTokenResponse.AccessToken
```

### Server-to-Server Authorization

```go
// Create an auth client with client credentials
authClient, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithClientSecret("client-secret"),
)
if err != nil {
    // Handle error
}

// Get a token for a specific scope
tokenResponse, err := authClient.GetClientCredentialsToken(ctx, []string{
    "https://auth.globus.org/scopes/transfer.api.globus.org/all",
})
if err != nil {
    // Handle error
}

// Create a service client with the token
transferClient, err := transfer.NewClient(
    transfer.WithAccessToken(tokenResponse.AccessToken),
)
if err != nil {
    // Handle error
}

// Use the transfer client
// ...
```

### Native Application Flow with Token Storage

```go
// Create an auth client
authClient, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithRedirectURL("https://example.com/callback"),
)
if err != nil {
    // Handle error
}

// Create a token manager with file storage
manager, err := tokens.NewManager(
    tokens.WithFileStorage("~/.globus-tokens"),
    tokens.WithAuthClient(authClient),
)
if err != nil {
    // Handle error
}

// Check if we have tokens
entry, err := manager.GetToken(ctx, "default")
if err != nil {
    // No tokens, need to authenticate
    
    // Generate an auth URL
    authURL := authClient.GetAuthorizationURL([]string{
        auth.ScopeOpenID,
        auth.ScopeProfile,
        auth.ScopeEmail,
        auth.ScopeOfflineAccess,
    })
    
    // Display the URL to the user
    fmt.Println("Visit this URL to authenticate:", authURL)
    
    // Get the code from the user
    var code string
    fmt.Print("Enter the code: ")
    fmt.Scanln(&code)
    
    // Exchange the code for tokens
    tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
    if err != nil {
        // Handle error
    }
    
    // Store the tokens
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
        // Handle error
    }
} else {
    // We have tokens, use them
    fmt.Println("Using existing tokens")
}

// Use the tokens
accessToken := entry.TokenSet.AccessToken

// Create a service client with the token
transferClient, err := transfer.NewClient(
    transfer.WithAccessToken(accessToken),
)
if err != nil {
    // Handle error
}

// Use the transfer client
// ...
```

## Best Practices

1. Always use HTTPS for redirect URLs in production
2. Store client secrets securely and never expose them in client-side code
3. Request only the scopes you need
4. Implement token refresh to ensure uninterrupted access
5. Use the tokens package for managing tokens in long-running applications
6. Handle MFA challenges gracefully in interactive applications
7. Use appropriate error handling to provide meaningful feedback
8. Validate tokens regularly to ensure they are still valid
9. Revoke tokens when they are no longer needed
10. Consider using the client credentials flow for server-to-server applications
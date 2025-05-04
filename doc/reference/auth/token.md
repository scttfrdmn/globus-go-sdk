# Auth Service: Token Validation

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Token validation is a critical part of OAuth2 security. The Globus Auth service provides several methods for validating and managing tokens.

## Token Information

The `TokenInfo` struct represents information about a token obtained through introspection:

```go
type TokenInfo struct {
    Active    bool   `json:"active"`
    Scope     string `json:"scope,omitempty"`
    ClientID  string `json:"client_id,omitempty"`
    UserName  string `json:"username,omitempty"`
    Exp       int64  `json:"exp,omitempty"`
    Sub       string `json:"sub,omitempty"`
    Iss       string `json:"iss,omitempty"`
    Nbf       int64  `json:"nbf,omitempty"`
    Iat       int64  `json:"iat,omitempty"`
    Jti       string `json:"jti,omitempty"`
    TokenType string `json:"token_type,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Active` | `bool` | Whether the token is active (not expired or revoked) |
| `Scope` | `string` | Space-separated list of scopes |
| `ClientID` | `string` | ID of the client that requested the token |
| `UserName` | `string` | Username of the user who authorized the token (if available) |
| `Exp` | `int64` | Expiration time of the token (Unix timestamp) |
| `Sub` | `string` | Subject of the token (usually a user ID) |
| `Iss` | `string` | Issuer of the token |
| `Nbf` | `int64` | Not before time (Unix timestamp) |
| `Iat` | `int64` | Issued at time (Unix timestamp) |
| `Jti` | `string` | JWT ID |
| `TokenType` | `string` | Type of token |

## Token Introspection

Token introspection is the process of getting information about a token, including whether it's active (valid and not expired):

```go
// Create an auth client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithClientSecret("client-secret"),
)
if err != nil {
    // Handle error
}

// Introspect a token
tokenInfo, err := client.IntrospectToken(ctx, "access-token")
if err != nil {
    // Handle error
}

// Check if the token is active
if tokenInfo.IsActive() {
    fmt.Println("Token is valid")
    fmt.Println("Scope:", tokenInfo.Scope)
    fmt.Println("Subject:", tokenInfo.Sub)
    fmt.Println("Expires at:", tokenInfo.ExpiresAt())
} else {
    fmt.Println("Token is invalid or expired")
}
```

### TokenInfo Methods

#### IsActive

```go
// Check if a token is active
if tokenInfo.IsActive() {
    // Token is valid and not expired
} else {
    // Token is invalid or expired
}
```

#### ExpiresAt

```go
// Get the expiration time as a time.Time
expiresAt := tokenInfo.ExpiresAt()
fmt.Println("Token expires at:", expiresAt)
```

#### IsExpired

```go
// Check if a token is expired
if tokenInfo.IsExpired() {
    // Token is expired
} else {
    // Token is not expired
}
```

## Token Validation

The auth package provides utility functions for validating tokens:

### ValidateToken

```go
// Validate a token using introspection
isValid, err := auth.ValidateToken(ctx, client, "access-token")
if err != nil {
    // Handle error
}

if isValid {
    fmt.Println("Token is valid")
} else {
    fmt.Println("Token is invalid or expired")
}
```

### GetRemainingValidity

```go
// Get the remaining validity duration of a token
remaining, err := auth.GetRemainingValidity(ctx, client, "access-token")
if err != nil {
    // Handle error
}

fmt.Println("Token is valid for:", remaining)

// Check if token needs refresh
if remaining < 10*time.Minute {
    fmt.Println("Token will expire soon, refreshing...")
    // Refresh token
}
```

## Token Revocation

Revoking a token is the process of invalidating it before its expiration time:

```go
// Revoke an access token
err := client.RevokeToken(ctx, "access-token", "access_token")
if err != nil {
    // Handle error
}

// Revoke a refresh token
err = client.RevokeToken(ctx, "refresh-token", "refresh_token")
if err != nil {
    // Handle error
}
```

The second parameter specifies the token type. It can be:
- `"access_token"` for access tokens
- `"refresh_token"` for refresh tokens

## Token Expiry Checks

The auth client provides methods to check if a token is expired or should be refreshed:

### IsTokenValid

```go
// Check if a token is valid (not expired)
isValid, expiresAt, err := client.IsTokenValid(ctx, "access-token")
if err != nil {
    // Handle error
}

if isValid {
    fmt.Println("Token is valid until:", expiresAt)
} else {
    fmt.Println("Token is invalid or expired")
}
```

### GetTokenExpiry

```go
// Get the expiry time of a token
expiresAt, err := client.GetTokenExpiry(ctx, "access-token")
if err != nil {
    // Handle error
}

fmt.Println("Token expires at:", expiresAt)
```

### ShouldRefresh

```go
// Check if a token should be refreshed (expiry is within the threshold)
shouldRefresh, err := client.ShouldRefresh(ctx, "access-token", 30*time.Minute)
if err != nil {
    // Handle error
}

if shouldRefresh {
    fmt.Println("Token should be refreshed")
    // Refresh token
} else {
    fmt.Println("Token is still valid")
}
```

## Token Response Structure

The `TokenResponse` struct represents the response from token endpoints:

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

| Field | Type | Description |
|-------|------|-------------|
| `AccessToken` | `string` | The access token |
| `RefreshToken` | `string` | The refresh token (if requested) |
| `ExpiresIn` | `int` | Number of seconds until the access token expires |
| `ResourceServer` | `string` | The resource server for the token |
| `TokenType` | `string` | The type of token (usually "Bearer") |
| `Scope` | `string` | Space-separated list of granted scopes |
| `OtherTokens` | `map[string]interface{}` | Additional tokens for other resource servers |
| `DependentTokens` | `map[string]interface{}` | Dependent tokens for dependent resource servers |

### TokenResponse Methods

#### HasRefreshToken

```go
// Check if a token response has a refresh token
if tokenResponse.HasRefreshToken() {
    // Refresh token is available
    fmt.Println("Refresh token:", tokenResponse.RefreshToken)
} else {
    fmt.Println("No refresh token provided")
}
```

#### IsValid

```go
// Check if a token response is valid
if tokenResponse.IsValid() {
    // Token response is valid
    fmt.Println("Token response is valid")
} else {
    fmt.Println("Token response is invalid")
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

## Token Lifecycle Management

Managing tokens throughout their lifecycle involves:

1. **Obtaining**: Through an authorization flow
2. **Validating**: Before using, especially if stored
3. **Refreshing**: When they're about to expire
4. **Revoking**: When they're no longer needed

### Example: Complete Token Lifecycle

```go
// Create an auth client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithClientSecret("client-secret"),
    auth.WithRedirectURL("https://example.com/callback"),
)
if err != nil {
    // Handle error
}

// 1. Obtain tokens (authorization code flow)
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, "authorization-code")
if err != nil {
    // Handle error
}

// Use the tokens
accessToken := tokenResponse.AccessToken
refreshToken := tokenResponse.RefreshToken

// 2. Validate before use (especially after storage/retrieval)
isValid, expiresAt, err := client.IsTokenValid(ctx, accessToken)
if err != nil {
    // Handle error
}

if !isValid {
    fmt.Println("Token is invalid or expired")
    // Skip to refresh
} else {
    fmt.Println("Token is valid until:", expiresAt)
    
    // Use the token
    // ...
    
    // 3. Check if token should be refreshed
    shouldRefresh, err := client.ShouldRefresh(ctx, accessToken, 30*time.Minute)
    if err != nil {
        // Handle error
    }
    
    if shouldRefresh {
        fmt.Println("Token should be refreshed")
    } else {
        fmt.Println("Token is still valid")
    }
}

// 3. Refresh token if needed
if !isValid || shouldRefresh {
    newTokenResponse, err := client.RefreshToken(ctx, refreshToken)
    if err != nil {
        // Handle error
    }
    
    // Update tokens
    accessToken = newTokenResponse.AccessToken
    if newTokenResponse.HasRefreshToken() {
        refreshToken = newTokenResponse.RefreshToken
    }
    
    fmt.Println("Token refreshed")
}

// 4. Revoke tokens when done
err = client.RevokeToken(ctx, accessToken, "access_token")
if err != nil {
    // Handle error
}

err = client.RevokeToken(ctx, refreshToken, "refresh_token")
if err != nil {
    // Handle error
}

fmt.Println("Tokens revoked")
```

## Best Practices

1. **Validate Tokens**: Always validate tokens before use, especially if they've been stored or retrieved from storage
2. **Refresh Threshold**: Set a reasonable refresh threshold (15-30 minutes) to ensure tokens are refreshed before they expire
3. **Error Handling**: Handle token validation and refresh errors gracefully
4. **Revocation**: Revoke tokens when they're no longer needed, especially refresh tokens
5. **Token Storage**: Store tokens securely, ideally using the tokens package
6. **Automatic Refresh**: Use the tokens package for automatic token refresh
7. **Scopes**: Validate that tokens have the required scopes before using them
8. **Client Credentials**: Use client ID and secret when introspecting tokens for additional security
9. **MFA Handling**: Be prepared to handle MFA challenges during token refresh
10. **Dependent Tokens**: Check for and use dependent tokens when interacting with multiple services
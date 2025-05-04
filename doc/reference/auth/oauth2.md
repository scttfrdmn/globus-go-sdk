# Auth Service: OAuth2 Flows

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Globus Auth service supports various OAuth2 authentication flows, which are standardized ways for applications to obtain access tokens. This document explains each flow and provides examples for their implementation.

## Authorization Code Flow

The authorization code flow is designed for web applications and native applications where a user needs to authenticate. It's the most secure flow as it involves a redirect to the Globus Auth service, user authentication, and a code that's exchanged for tokens.

### When to Use

- Web applications with server-side components
- Native applications (desktop, mobile)
- Applications where users need to authenticate
- When you need refresh tokens

### Flow Overview

1. The application redirects the user to Globus Auth
2. The user authenticates and authorizes the application
3. Globus Auth redirects back to the application with an authorization code
4. The application exchanges the code for tokens
5. The application uses the tokens to access Globus services

### Implementation

```go
// Create an auth client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithRedirectURL("https://example.com/callback"),
)
if err != nil {
    // Handle error
}

// Step 1: Generate an authorization URL
authURL := client.GetAuthorizationURL([]string{
    auth.ScopeOpenID,
    auth.ScopeProfile,
    auth.ScopeEmail,
    auth.ScopeOfflineAccess,
})

// Step 2: Redirect the user to the authorization URL
// (In a web application, this would be an HTTP redirect)
fmt.Println("Please visit this URL to authorize the application:", authURL)

// Step 3: Get the code from the redirect
// (In a web application, this would come from the callback URL)
var code string
fmt.Print("Enter the code from the redirect: ")
fmt.Scanln(&code)

// Step 4: Exchange the code for tokens
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    // Handle error
}

// Step 5: Use the tokens
fmt.Println("Access Token:", tokenResponse.AccessToken)
fmt.Println("Refresh Token:", tokenResponse.RefreshToken)
fmt.Println("Expires In:", tokenResponse.ExpiresIn, "seconds")
```

## Client Credentials Flow

The client credentials flow is designed for server-to-server applications where user interaction is not possible or needed. It uses client credentials (ID and secret) to obtain tokens.

### When to Use

- Server-to-server applications
- Background processes
- When no user interaction is required
- When the client is trusted (can securely store client secret)

### Flow Overview

1. The application authenticates with client ID and secret
2. Globus Auth returns a token for the requested scopes
3. The application uses the token to access Globus services

### Implementation

```go
// Create an auth client with client credentials
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithClientSecret("client-secret"),
)
if err != nil {
    // Handle error
}

// Request a token with specific scopes
tokenResponse, err := client.GetClientCredentialsToken(ctx, []string{
    "https://auth.globus.org/scopes/transfer.api.globus.org/all",
})
if err != nil {
    // Handle error
}

// Use the token
fmt.Println("Access Token:", tokenResponse.AccessToken)
fmt.Println("Expires In:", tokenResponse.ExpiresIn, "seconds")
```

## Refresh Token Flow

The refresh token flow allows applications to obtain new access tokens without requiring user interaction when the original access token expires.

### When to Use

- To extend sessions without user re-authentication
- Long-running applications where tokens may expire
- To maintain uninterrupted access to Globus services

### Flow Overview

1. The application detects that an access token is expired or about to expire
2. The application uses the refresh token to request a new access token
3. Globus Auth returns a new access token
4. The application continues using the new access token

### Implementation

```go
// Create an auth client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
)
if err != nil {
    // Handle error
}

// Refresh a token
tokenResponse, err := client.RefreshToken(ctx, "refresh-token")
if err != nil {
    // Handle error
}

// Use the new tokens
fmt.Println("New Access Token:", tokenResponse.AccessToken)
fmt.Println("New Refresh Token:", tokenResponse.RefreshToken) // May or may not be provided
fmt.Println("Expires In:", tokenResponse.ExpiresIn, "seconds")
```

## Device Code Flow

While not directly supported in the SDK, the device code flow is useful for devices with limited input capabilities.

### When to Use

- Devices with limited input capabilities (e.g., IoT devices, smart TVs)
- Applications without a web browser
- When a user can use a secondary device for authentication

### Flow Overview

1. The application requests a device code from Globus Auth
2. Globus Auth returns a device code, user code, and verification URL
3. The application displays the user code and URL to the user
4. The user visits the URL on another device and enters the code
5. The application polls Globus Auth until the user completes authentication
6. Globus Auth returns tokens to the application

### Implementation Note

While this flow isn't directly supported in the SDK, you can implement it using the core HTTP client:

```go
// Create a core client
client := core.NewClient()

// Request a device code
resp, err := client.Post(ctx, "https://auth.globus.org/v2/oauth2/device_authorization", map[string]string{
    "client_id": "client-id",
    "scope": "openid profile email offline_access",
})
if err != nil {
    // Handle error
}

// Parse the response
var deviceCodeResp struct {
    DeviceCode              string `json:"device_code"`
    UserCode                string `json:"user_code"`
    VerificationURI         string `json:"verification_uri"`
    VerificationURIComplete string `json:"verification_uri_complete"`
    ExpiresIn               int    `json:"expires_in"`
    Interval                int    `json:"interval"`
}
if err := json.Unmarshal(resp, &deviceCodeResp); err != nil {
    // Handle error
}

// Display instructions to the user
fmt.Println("Please visit this URL on another device:", deviceCodeResp.VerificationURI)
fmt.Println("And enter this code:", deviceCodeResp.UserCode)
fmt.Println("Or visit:", deviceCodeResp.VerificationURIComplete)

// Poll for tokens
for {
    time.Sleep(time.Duration(deviceCodeResp.Interval) * time.Second)
    
    resp, err := client.Post(ctx, "https://auth.globus.org/v2/oauth2/token", map[string]string{
        "client_id": "client-id",
        "device_code": deviceCodeResp.DeviceCode,
        "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
    })
    if err != nil {
        // Check if it's a "not yet authorized" error
        if strings.Contains(err.Error(), "authorization_pending") {
            continue
        }
        // Handle other errors
        break
    }
    
    // Parse the token response
    var tokenResponse auth.TokenResponse
    if err := json.Unmarshal(resp, &tokenResponse); err != nil {
        // Handle error
        break
    }
    
    // Use the tokens
    fmt.Println("Access Token:", tokenResponse.AccessToken)
    fmt.Println("Refresh Token:", tokenResponse.RefreshToken)
    fmt.Println("Expires In:", tokenResponse.ExpiresIn, "seconds")
    break
}
```

## Choosing the Right Flow

When developing an application that integrates with Globus Auth, it's important to choose the most appropriate flow:

| Flow | Best For | User Interaction | Refresh Tokens |
|------|----------|------------------|----------------|
| Authorization Code | Web and native apps with user authentication | Required | Yes |
| Client Credentials | Server-to-server applications | None | No |
| Refresh Token | Extending sessions | None (after initial authentication) | Yes (input) |
| Device Code | Limited-input devices | Required on another device | Yes |

## Scope Management

Scopes define what access the application is requesting. Always request only the scopes needed:

### Common Scopes

```go
// OpenID Connect scopes
auth.ScopeOpenID      // "openid"
auth.ScopeProfile     // "profile"
auth.ScopeEmail       // "email"
auth.ScopeOfflineAccess // "offline_access" (for refresh tokens)

// Convenience combinations
auth.ScopeDefaultOpenID // "openid profile email"
auth.ScopeDefaultOpenIDOffline // "openid profile email offline_access"
```

### Service-Specific Scopes

```go
// Transfer service (full access)
"https://auth.globus.org/scopes/transfer.api.globus.org/all"

// Groups service (full access)
"https://auth.globus.org/scopes/groups.api.globus.org/all"

// Search service (full access)
"https://auth.globus.org/scopes/search.api.globus.org/all"

// Flows service (full access)
"https://auth.globus.org/scopes/flows.globus.org/all"
```

## Token Validation

It's important to validate tokens before using them, especially if they've been stored:

```go
// Check if a token is still valid
tokenInfo, err := client.IntrospectToken(ctx, "access-token")
if err != nil {
    // Handle error
}

if tokenInfo.IsActive() {
    // Token is valid
    fmt.Println("Token is valid until:", tokenInfo.ExpiresAt())
} else {
    // Token is invalid or expired
    fmt.Println("Token is invalid or expired")
}
```

## Error Handling

Different flows may encounter different errors. Handle them appropriately:

```go
// Exchange code for tokens
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    switch {
    case auth.IsInvalidGrant(err):
        // Code is expired or already used
        fmt.Println("The authorization code is invalid or expired. Please try again.")
    case auth.IsInvalidClient(err):
        // Client ID/secret is incorrect
        fmt.Println("Invalid client credentials. Check your client ID and secret.")
    case auth.IsInvalidScope(err):
        // Requested scopes are invalid
        fmt.Println("Invalid scopes requested. Check your scope configuration.")
    case auth.IsMFAError(err):
        // MFA is required
        challenge := auth.GetMFAChallenge(err)
        fmt.Println("MFA required:", challenge.Prompt)
        // Handle MFA
    default:
        // Other error
        fmt.Println("Authentication error:", err)
    }
    return
}
```

## Best Practices

1. **Security**:
   - Keep client secrets secure and never expose them in client-side code
   - Use HTTPS for all redirects
   - Validate the `state` parameter in authorization code flow to prevent CSRF

2. **Scopes**:
   - Request only the scopes your application needs
   - Request `offline_access` scope if you need refresh tokens

3. **Token Management**:
   - Store tokens securely
   - Refresh tokens before they expire to avoid service interruptions
   - Use the tokens package for automatic token management

4. **Error Handling**:
   - Handle authentication errors gracefully
   - Provide clear instructions to users when authentication fails
   - Implement proper MFA handling for interactive applications

5. **User Experience**:
   - Make authentication flows as seamless as possible
   - Provide clear instructions for users during authentication
   - Consider persisting tokens to reduce the need for re-authentication
# Auth Service: MFA Support

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Multi-Factor Authentication (MFA) provides an additional layer of security beyond passwords. The Globus Auth service supports MFA, and the SDK provides utilities for handling MFA challenges during authentication flows.

## MFA Challenge Structure

When MFA is required, an error will be returned from authentication methods. This error contains information about the MFA challenge:

```go
type MFAChallenge struct {
    ChallengeID string `json:"challenge_id"`
    Type        string `json:"type"`
    Prompt      string `json:"prompt"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `ChallengeID` | `string` | Unique identifier for the MFA challenge |
| `Type` | `string` | Type of MFA challenge (e.g., "totp", "sms") |
| `Prompt` | `string` | Human-readable prompt for the user |

## MFA Required Error

When MFA is required, an `MFARequiredError` is returned:

```go
type MFARequiredError struct {
    Challenge MFAChallenge
    Err       error
}
```

You can check if an error is an MFA error using the `IsMFAError` function:

```go
// Try to exchange authorization code
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, "code")
if err != nil {
    if auth.IsMFAError(err) {
        // Handle MFA
        challenge := auth.GetMFAChallenge(err)
        fmt.Println("MFA Required:", challenge.Prompt)
    } else {
        // Handle other errors
    }
}
```

## Handling MFA in Authorization Code Flow

When exchanging an authorization code, MFA might be required:

```go
// Create an auth client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
    auth.WithRedirectURL("https://example.com/callback"),
)
if err != nil {
    // Handle error
}

// Try to exchange authorization code
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

// Use the tokens
fmt.Println("Access Token:", tokenResponse.AccessToken)
```

## Handling MFA in Token Refresh

MFA might also be required when refreshing a token:

```go
// Try to refresh token
tokenResponse, err := client.RefreshToken(ctx, "refresh-token")
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
        
        // Refresh with MFA
        tokenResponse, err = client.RefreshTokenWithMFA(
            ctx,
            "refresh-token",
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

// Use the tokens
fmt.Println("Access Token:", tokenResponse.AccessToken)
```

## MFA Response Structure

When responding to an MFA challenge, you can provide additional information:

```go
type MFAResponse struct {
    Response     string `json:"response"`
    RememberMe   bool   `json:"remember_me,omitempty"`
    RememberDays int    `json:"remember_days,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Response` | `string` | The user's response to the MFA challenge |
| `RememberMe` | `bool` | Whether to remember the MFA for a period of time |
| `RememberDays` | `int` | Number of days to remember the MFA (if `RememberMe` is true) |

## Using Custom MFA Response

You can create a custom MFA response with additional options:

```go
// Create an MFA response with remember me
mfaResponse := &auth.MFAResponse{
    Response:     "123456", // User's input
    RememberMe:   true,     // Remember this MFA
    RememberDays: 30,       // Remember for 30 days
}

// Exchange with custom MFA response
tokenResponse, err := client.ExchangeAuthorizationCodeWithMFAResponse(
    ctx,
    "code",
    challenge.ChallengeID,
    mfaResponse,
)
if err != nil {
    // Handle error
}
```

```go
// Refresh with custom MFA response
tokenResponse, err := client.RefreshTokenWithMFAResponse(
    ctx,
    "refresh-token",
    challenge.ChallengeID,
    mfaResponse,
)
if err != nil {
    // Handle error
}
```

## Getting MFA Challenge Details

You can get more details about an MFA challenge:

```go
// Get detailed information about an MFA challenge
challengeDetails, err := client.GetMFAChallenge(ctx, challenge.ChallengeID)
if err != nil {
    // Handle error
}

fmt.Println("Challenge Type:", challengeDetails.Type)
fmt.Println("Challenge Prompt:", challengeDetails.Prompt)
```

## Responding to MFA Challenge

You can respond to an MFA challenge directly:

```go
// Respond to an MFA challenge
err := client.RespondToMFAChallenge(
    ctx,
    challenge.ChallengeID,
    "123456", // User's input
)
if err != nil {
    // Handle error
}
```

## Common MFA Patterns

### Interactive CLI Application

```go
// Try to exchange authorization code
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    if auth.IsMFAError(err) {
        // Get MFA challenge
        challenge := auth.GetMFAChallenge(err)
        
        // Display prompt to user
        fmt.Println(challenge.Prompt)
        
        // Get MFA code from user
        var mfaCode string
        fmt.Print("Enter MFA code: ")
        fmt.Scanln(&mfaCode)
        
        // Ask if user wants to remember MFA
        var rememberMFA string
        fmt.Print("Remember MFA for 30 days? (y/n): ")
        fmt.Scanln(&rememberMFA)
        
        // Create MFA response
        mfaResponse := &auth.MFAResponse{
            Response: mfaCode,
        }
        
        // Add remember me if requested
        if strings.ToLower(rememberMFA) == "y" {
            mfaResponse.RememberMe = true
            mfaResponse.RememberDays = 30
        }
        
        // Exchange with MFA
        tokenResponse, err = client.ExchangeAuthorizationCodeWithMFAResponse(
            ctx,
            code,
            challenge.ChallengeID,
            mfaResponse,
        )
        if err != nil {
            fmt.Println("MFA authentication failed:", err)
            return
        }
    } else {
        fmt.Println("Authentication failed:", err)
        return
    }
}

// Use tokens
fmt.Println("Authentication successful!")
```

### Web Application

```go
// In a web application, you would typically handle MFA in a multi-step process

// Step 1: Try to exchange the code
tokenResponse, err := client.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    if auth.IsMFAError(err) {
        // Get MFA challenge
        challenge := auth.GetMFAChallenge(err)
        
        // Store the code and challenge ID in session
        session.Set("auth_code", code)
        session.Set("mfa_challenge_id", challenge.ChallengeID)
        
        // Redirect to MFA page
        http.Redirect(w, r, "/mfa?prompt="+url.QueryEscape(challenge.Prompt), http.StatusFound)
        return
    } else {
        // Handle other errors
        http.Error(w, "Authentication failed: "+err.Error(), http.StatusBadRequest)
        return
    }
}

// ... normal flow if MFA wasn't required ...

// Step 2: Handle MFA form submission
// This would be a separate handler for the MFA form
func handleMFASubmit(w http.ResponseWriter, r *http.Request) {
    // Get MFA code from form
    mfaCode := r.FormValue("mfa_code")
    rememberMFA := r.FormValue("remember_mfa") == "on"
    
    // Get code and challenge ID from session
    code := session.Get("auth_code")
    challengeID := session.Get("mfa_challenge_id")
    
    // Create MFA response
    mfaResponse := &auth.MFAResponse{
        Response: mfaCode,
    }
    
    // Add remember me if requested
    if rememberMFA {
        mfaResponse.RememberMe = true
        mfaResponse.RememberDays = 30
    }
    
    // Exchange with MFA
    tokenResponse, err := client.ExchangeAuthorizationCodeWithMFAResponse(
        r.Context(),
        code,
        challengeID,
        mfaResponse,
    )
    if err != nil {
        // Handle error
        http.Error(w, "MFA authentication failed: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // Clear session data
    session.Delete("auth_code")
    session.Delete("mfa_challenge_id")
    
    // Store tokens and redirect to dashboard
    session.Set("access_token", tokenResponse.AccessToken)
    session.Set("refresh_token", tokenResponse.RefreshToken)
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

## Error Handling

MFA-related errors should be handled specifically:

```go
if err != nil {
    switch {
    case auth.IsMFAError(err):
        // MFA is required
        challenge := auth.GetMFAChallenge(err)
        // Handle MFA challenge
    case auth.IsInvalidGrant(err):
        // Code or refresh token is invalid
        fmt.Println("Invalid code or refresh token")
    case auth.IsInvalidClient(err):
        // Client credentials are invalid
        fmt.Println("Invalid client credentials")
    case auth.IsAccessDenied(err):
        // Access denied
        fmt.Println("Access denied")
    default:
        // Other error
        fmt.Println("Authentication error:", err)
    }
}
```

## Testing with MFA

For testing purposes, you can check if MFA would be required:

```go
// Create a test client
client, err := auth.NewClient(
    auth.WithClientID("client-id"),
)
if err != nil {
    // Handle error
}

// Check if MFA would be required for a specific user
wouldRequireMFA, err := client.CheckForMFARequired(ctx, "example@globus.org")
if err != nil {
    // Handle error
}

if wouldRequireMFA {
    fmt.Println("MFA would be required for this user")
} else {
    fmt.Println("MFA would not be required for this user")
}
```

## Best Practices

1. **Always Handle MFA**: Always check for MFA errors in authentication flows
2. **Clear Instructions**: Provide clear instructions to users when MFA is required
3. **Remember Me**: Offer a "remember me" option for better user experience
4. **Timeout Handling**: Be prepared to handle MFA challenge timeouts
5. **Error Messages**: Provide helpful error messages for failed MFA attempts
6. **Fallback Mechanisms**: Provide fallback mechanisms for users who can't access their MFA device
7. **Security**: Treat MFA challenge IDs and responses with the same security as passwords
8. **Cancellation**: Allow users to cancel MFA and start over if needed
9. **Progressive Enhancement**: Design your application to work even if MFA is unexpectedly required
10. **Testing**: Test both MFA and non-MFA authentication flows
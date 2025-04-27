# Multi-Factor Authentication (MFA) Guide

This guide explains how to handle Multi-Factor Authentication (MFA) requirements when authenticating with Globus Auth in the Globus Go SDK.

## Overview

Globus supports Multi-Factor Authentication (MFA) to enhance security for sensitive operations. When MFA is required, the authentication process involves these steps:

1. The initial authentication request is made (e.g., exchanging an authorization code)
2. If MFA is required, the server returns an MFA challenge
3. The application must obtain an MFA code from the user
4. The MFA code is submitted in response to the challenge
5. If the MFA code is valid, authentication proceeds

The Globus Go SDK provides built-in support for handling MFA challenges through a callback-based approach.

## Using MFA-Enabled Authentication Methods

The Auth client includes MFA-aware versions of the authentication methods:

- `ExchangeAuthorizationCodeWithMFA`: Exchanges an authorization code with MFA support
- `RefreshTokenWithMFA`: Refreshes a token with MFA support

These methods take an additional parameter called an "MFA handler" which is a function that gets called when an MFA challenge is received. The handler is responsible for obtaining the MFA code from the user and returning an appropriate response.

### Example: Exchanging an Authorization Code with MFA

```go
// Create an Auth client
authClient := config.NewAuthClient()

// Exchange the authorization code with MFA support
tokenResp, err := authClient.ExchangeAuthorizationCodeWithMFA(
    ctx, 
    code,
    func(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
        // Display information about the MFA challenge
        fmt.Printf("MFA Required (%s): %s\n", challenge.Type, challenge.Prompt)
        
        // Ask the user for the MFA code
        fmt.Print("Enter your MFA code: ")
        var code string
        fmt.Scanln(&code)
        
        // Return the MFA response
        return &auth.MFAResponse{
            ChallengeID: challenge.ChallengeID,
            Type:        challenge.Type,
            Value:       code,
        }, nil
    },
)

if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}

// Use the token response
fmt.Printf("Authentication successful! Token: %s\n", tokenResp.AccessToken)
```

### Example: Refreshing a Token with MFA

```go
// Refresh a token with MFA support
tokenResp, err := authClient.RefreshTokenWithMFA(
    ctx, 
    refreshToken,
    func(challenge *auth.MFAChallenge) (*auth.MFAResponse, error) {
        // Handle the MFA challenge (as shown above)
        // ...
        
        return &auth.MFAResponse{
            ChallengeID: challenge.ChallengeID,
            Type:        challenge.Type,
            Value:       code,
        }, nil
    },
)
```

## Working with MFA Challenges

The `MFAChallenge` struct contains information about the MFA challenge:

```go
type MFAChallenge struct {
    // ChallengeID is the unique identifier for this challenge
    ChallengeID string `json:"challenge_id"`

    // Type indicates the type of MFA challenge (e.g., "totp", "webauthn", "backup_code")
    Type string `json:"type"`

    // Prompt is a human-readable prompt to display to the user
    Prompt string `json:"prompt"`

    // AllowedTypes contains all MFA types that can be used to satisfy this challenge
    AllowedTypes []string `json:"allowed_types"`

    // Additional information specific to the challenge type
    Extra map[string]interface{} `json:"extra,omitempty"`
}
```

The MFA handler function should:

1. Present the challenge information to the user
2. Obtain the appropriate response (e.g., a TOTP code)
3. Return an `MFAResponse` struct with the challenge ID, selected type, and response value

## MFA Response Types

Globus supports several MFA types:

- **TOTP**: Time-based One-Time Passwords from authenticator apps
- **WebAuthn**: Web Authentication using security keys or biometrics 
- **Backup Codes**: One-time use backup codes

The `AllowedTypes` field in the MFA challenge indicates which types can be used to satisfy the challenge.

## Handling MFA Errors

When MFA is required but not provided, or if the provided MFA code is invalid, the SDK returns an `MFARequiredError` error. You can check for this specific error type:

```go
// Try to exchange the code
tokenResp, err := authClient.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    // Check if this is an MFA error
    if auth.IsMFAError(err) {
        // Get the MFA challenge
        challenge := auth.GetMFAChallenge(err)
        if challenge != nil {
            // Handle the MFA challenge
            // ...
        }
    } else {
        // Handle other errors
        log.Fatalf("Authentication failed: %v", err)
    }
}
```

## Best Practices

1. **UI Integration**: For graphical applications, integrate MFA prompts smoothly into your UI
2. **Fallback Options**: Support multiple MFA types when possible (e.g., TOTP and backup codes)
3. **Clear Instructions**: Provide clear instructions to users about where to find their MFA code
4. **Error Handling**: Handle MFA errors gracefully and provide helpful error messages
5. **Retry Logic**: Allow users to retry if they enter an incorrect MFA code

## Complete Example

See [`examples/mfa-auth/main.go`](../examples/mfa-auth/main.go) for a complete example of handling MFA in the authentication flow.

The example demonstrates:
- Setting up a web server for the OAuth2 callback
- Exchanging an authorization code with MFA support
- Refreshing a token with MFA support
- Handling MFA challenges by prompting the user for input

## Related Resources

- [Globus Auth Documentation](https://docs.globus.org/api/auth/)
- [Globus MFA Guide](https://docs.globus.org/api/auth/mfa/)
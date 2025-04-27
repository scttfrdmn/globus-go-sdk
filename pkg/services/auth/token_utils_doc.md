<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Token Utilities Documentation

The `token_utils.go` file provides several helpful utilities for working with Globus Auth tokens. These utilities make it easier to validate, check, and manage token lifetimes.

## Key Functions

### ValidateToken

```go
func (c *Client) ValidateToken(ctx context.Context, token string) error
```

Validates a token by calling the introspection endpoint. Returns `nil` if the token is valid and active, or one of the following errors:
- `ErrTokenInvalid`: The token is invalid or has been revoked
- `ErrTokenExpired`: The token has expired
- Other errors if the introspection request fails

### GetTokenExpiry

```go
func (c *Client) GetTokenExpiry(ctx context.Context, token string) (time.Time, bool, error)
```

Returns the expiry time of a token, a boolean indicating if the token is valid, and any error that occurred during introspection.

### IsTokenValid

```go
func (c *Client) IsTokenValid(ctx context.Context, token string) bool
```

Convenience method that returns `true` if the token is valid, `false` otherwise.

### GetRemainingValidity

```go
func (c *Client) GetRemainingValidity(ctx context.Context, token string) (time.Duration, error)
```

Returns the remaining validity duration of a token. Returns `0` if the token is already expired or invalid.

### ShouldRefresh

```go
func (c *Client) ShouldRefresh(ctx context.Context, token string, threshold time.Duration) (bool, error)
```

Determines if a token should be refreshed based on a threshold duration. Returns `true` if the token will expire within the threshold period.

## Example Usage

```go
// Create an auth client
client := auth.NewClient(clientID, clientSecret)

// Validate a token
if err := client.ValidateToken(ctx, accessToken); err != nil {
    if errors.Is(err, auth.ErrTokenExpired) {
        // Token has expired, refresh it
        tokenResp, err := client.RefreshToken(ctx, refreshToken)
        if err != nil {
            // Handle refresh error
        }
        accessToken = tokenResp.AccessToken
    } else {
        // Handle other validation errors
    }
}

// Check if a token should be refreshed
shouldRefresh, err := client.ShouldRefresh(ctx, accessToken, 5*time.Minute)
if err != nil {
    // Handle error
}
if shouldRefresh {
    // Refresh the token
}
```
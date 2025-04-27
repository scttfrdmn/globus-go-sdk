<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Error Handling Documentation

The `errors.go` file provides enhanced error handling for the Globus Auth client. It defines common error types, error checking functions, and utilities for parsing error responses from the Globus Auth API.

## Error Types

The package defines several standard error variables that can be checked using Go's `errors.Is()`:

- `ErrInvalidGrant`: The authorization code or refresh token is invalid
- `ErrInvalidClient`: The client credentials are invalid
- `ErrInvalidScope`: The requested scope is invalid
- `ErrAccessDenied`: The user denied the authorization request
- `ErrServerError`: The server encountered an error
- `ErrUnauthorized`: The request is not authorized
- `ErrBadRequest`: The request is malformed
- `ErrTokenExpired`: The token has expired (special case of invalid grant)

## Error Checking Functions

Convenient functions for checking specific error types:

- `IsInvalidGrant(err error) bool`
- `IsInvalidClient(err error) bool`
- `IsInvalidScope(err error) bool`
- `IsAccessDenied(err error) bool`
- `IsServerError(err error) bool`
- `IsUnauthorized(err error) bool`
- `IsBadRequest(err error) bool`

These functions check both standard errors using `errors.Is()` and custom `AuthError` types.

## Error Structures

The `AuthError` type represents errors returned by the Globus Auth API:

```go
type AuthError struct {
    Code        string `json:"error"`
    Description string `json:"error_description,omitempty"`
    StatusCode  int    `json:"-"`
}
```

## Example Usage

```go
// Try to refresh a token
tokenResp, err := client.RefreshToken(ctx, refreshToken)
if err != nil {
    if auth.IsInvalidGrant(err) {
        // The refresh token is invalid, need to re-authenticate
        fmt.Println("Your session has expired. Please log in again.")
    } else if auth.IsServerError(err) {
        // A server error occurred
        fmt.Println("The service is currently unavailable. Please try again later.")
    } else {
        // Handle other errors
        fmt.Printf("An error occurred: %v\n", err)
    }
    return
}
```
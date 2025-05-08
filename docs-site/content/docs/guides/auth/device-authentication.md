<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
---
title: "Device Authentication"
weight: 30
---

# Device Authentication Flow

The device authentication flow is designed for non-browser environments like CLI applications, scripts, or other headless services. This flow allows an application to authenticate with Globus without requiring a web browser on the same device.

## Overview

The device authentication flow follows these steps:

1. **Request a Device Code**: The application requests a device code from Globus Auth
2. **User Authorization**: The user is directed to a verification URL and enters a user code
3. **Polling**: The application polls Globus Auth to check for authorization
4. **Token Access**: Once authorized, the application receives access and refresh tokens

## Using Device Authentication

### Step 1: Create an Auth Client

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// Create the auth client
authClient, err := auth.NewClient(
    auth.WithClientID("your-client-id"),
    auth.WithCoreOption(core.WithHTTPDebugging(false)),
)
if err != nil {
    log.Fatalf("Failed to create auth client: %v", err)
}
```

### Step 2: Request a Device Code

```go
// Define the scopes needed for your application
scopes := []string{
    "openid",
    "profile",
    "email",
    "urn:globus:auth:scope:transfer.api.globus.org:all",
}

// Request a device code
deviceCode, err := authClient.RequestDeviceCode(ctx, scopes...)
if err != nil {
    log.Fatalf("Failed to request device code: %v", err)
}

// Display the verification URL and user code to the user
fmt.Println("Please visit this URL to authorize this application:")
fmt.Printf("  %s\n\n", deviceCode.VerificationURI)
fmt.Println("Enter the following code when prompted:")
fmt.Printf("  %s\n", deviceCode.UserCode)
fmt.Printf("This code will expire in %d seconds.\n", deviceCode.ExpiresIn)
```

### Step 3: Poll for Authorization

```go
// Create a ticker for polling
interval := time.Duration(deviceCode.Interval) * time.Second
ticker := time.NewTicker(interval)
defer ticker.Stop()

// Poll until success, error, or context cancellation
for {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-ticker.C:
        // Poll for the token
        token, err := authClient.PollDeviceCode(ctx, deviceCode.DeviceCode)
        if err != nil {
            // Check if this is a retriable error
            if auth.IsAuthorizationPending(err) {
                // This is normal, continue polling
                continue
            }
            if auth.IsSlowDown(err) {
                // Slow down polling
                ticker.Reset(interval * 2)
                continue
            }
            // Other errors are terminal
            return nil, err
        }

        // Success! Use the token
        return token, nil
    }
}
```

### Simplified Approach: Using CompleteDeviceFlow

For a simpler implementation, you can use the `CompleteDeviceFlow` method, which handles the entire flow:

```go
// This callback will be called with the device code information
displayCallback := func(deviceCode *auth.DeviceCodeResponse) {
    fmt.Println("Please visit this URL to authorize this application:")
    fmt.Printf("  %s\n\n", deviceCode.VerificationURI)
    fmt.Println("Enter the following code when prompted:")
    fmt.Printf("  %s\n", deviceCode.UserCode)
    fmt.Printf("This code will expire in %d seconds.\n", deviceCode.ExpiresIn)
}

// Start the device flow and wait for user authorization
// Pass 0 for pollInterval to use the recommended interval from the device code
tokenResp, err := authClient.CompleteDeviceFlow(ctx, displayCallback, 0, scopes...)
if err != nil {
    log.Fatalf("Device flow failed: %v", err)
}

// Use the tokens
fmt.Printf("Access Token: %s...\n", tokenResp.AccessToken[:15])
```

## Error Handling

The device flow can result in several specific error types:

- `authorization_pending`: The user has not yet completed the authorization
- `slow_down`: The application is polling too frequently
- `expired_token`: The device code has expired
- `access_denied`: The user denied the authorization request

You can check for these errors using the appropriate helper functions:

```go
if auth.IsAuthorizationPending(err) {
    // Continue polling
}

if auth.IsSlowDown(err) {
    // Increase the polling interval
}

if auth.IsExpiredToken(err) {
    // Request a new device code
}

if auth.IsAccessDenied(err) {
    // Inform the user and abort
}
```

## Complete Example

A complete example implementation is available in the [device-auth example](https://github.com/scttfrdmn/globus-go-sdk/tree/main/cmd/examples/device-auth).

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

func main() {
    // Get client ID from environment variable
    clientID := os.Getenv("GLOBUS_CLIENT_ID")
    if clientID == "" {
        log.Fatal("GLOBUS_CLIENT_ID environment variable must be set")
    }

    // Create the auth client
    authClient, err := auth.NewClient(
        auth.WithClientID(clientID),
        auth.WithCoreOption(core.WithHTTPDebugging(false)),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }

    // Create a context with a timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    // Define the scopes we need
    scopes := []string{
        "openid",
        "profile",
        "email",
        "urn:globus:auth:scope:transfer.api.globus.org:all",
    }

    // Display callback function
    displayCallback := func(deviceCode *auth.DeviceCodeResponse) {
        fmt.Println("\n===== Device Authorization Required =====")
        fmt.Println("Please visit this URL to authorize this application:")
        fmt.Printf("  %s\n\n", deviceCode.VerificationURI)
        fmt.Println("Enter the following code when prompted:")
        fmt.Printf("  %s\n", deviceCode.UserCode)
        fmt.Println("=======================================")
        fmt.Printf("This code will expire in %d seconds.\n", deviceCode.ExpiresIn)
        fmt.Printf("Waiting for authorization...\n\n")
    }

    // Start the device flow and wait for user authorization
    tokenResp, err := authClient.CompleteDeviceFlow(ctx, displayCallback, 0, scopes...)
    if err != nil {
        log.Fatalf("Device flow failed: %v", err)
    }

    fmt.Println("Authentication successful!")
    fmt.Printf("Access Token: %s...\n", tokenResp.AccessToken[:15])
    fmt.Printf("Token expires in: %d seconds\n", tokenResp.ExpiresIn)
    
    if tokenResp.RefreshToken != "" {
        fmt.Printf("Refresh Token: %s...\n", tokenResp.RefreshToken[:15])
    }
}
```
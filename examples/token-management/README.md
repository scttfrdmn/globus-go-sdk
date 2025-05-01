<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Token Management Example

This example demonstrates how to use the Globus Go SDK's token management features to store, retrieve, and automatically refresh OAuth 2.0 tokens.

## Overview

The token management example showcases three main features of the Globus Go SDK token system:

1. **In-Memory Token Storage**: Demonstrates storing tokens in memory for short-lived applications.
2. **File-Based Token Storage**: Shows how to persist tokens to disk for applications that need to maintain auth state across restarts.
3. **Token Manager**: Illustrates how to use the token manager for automatic token refreshing.

## Prerequisites

To run this example, you need:

1. Go 1.19 or later
2. A Globus developer account
3. A registered Globus application with:
   - Client ID and Client Secret
   - Appropriate scopes for the services you want to access

## Running the Example

### Without Globus Credentials (Demo Mode)

You can run the example without any Globus credentials. In this mode, it will use mock implementations to demonstrate the token management features:

```bash
go run .
```

This is the easiest way to see how the tokens package works without needing to set up Globus credentials.

### With Globus Credentials (Real Mode)

For a complete demonstration with real Globus authentication, set the following environment variables:

```bash
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
```

Then run the example:

```bash
go run .
```

## Testing with Real Credentials

For release testing, you should verify that the tokens package works correctly with real Globus credentials. A test script is provided for this purpose:

```bash
# Make sure you have a .env.test file in the project root with your credentials
./test_tokens.sh
```

This will:

1. Verify that you have a .env.test file with the required credentials
2. Run test_with_credentials.go to test the tokens package with real authentication
3. Run the standard example in demo mode

The test will:
- Obtain tokens using client credentials flow
- Store and retrieve tokens
- Test token refreshing (if refresh tokens are available)
- Verify that all token operations work as expected

## Key Features

### 1. Mock Token Handler

The example includes a mock implementation of the RefreshHandler interface that demonstrates:

- How to implement the RefreshHandler interface
- Automatic token refreshing without real Globus credentials
- How to simulate token expiration and refresh

This is useful for:
- Testing token-related functionality
- Development without requiring real credentials
- Understanding the token refresh flow

### 2. In-Memory Token Storage

The example demonstrates how to:
- Create an in-memory token storage
- Store token entries
- Retrieve token entries
- List all stored tokens
- Delete tokens

This is useful for:
- Command-line applications
- Short-lived processes
- Testing environments

### 3. File-Based Token Storage

The example shows how to:
- Create a file-based token storage
- Store tokens to disk
- Retrieve tokens from disk
- List all stored tokens
- Delete token files

This is useful for:
- Desktop applications
- Long-running services
- Applications that need to maintain authentication across restarts

### 4. Token Manager

The example demonstrates how to:
- Create a token manager with a storage backend
- Configure token refresh thresholds
- Set up automatic token refreshing
- Use background refresh for proactive token management

This is useful for:
- Web applications
- APIs and services
- Any application interacting with Globus services

## Token Manager Usage

The token manager provides automatic token refreshing when tokens are close to expiry.

```go
// Create a storage backend
storage := tokens.NewFileStorage("./tokens")

// Create a token manager with an auth client
manager := tokens.NewManager(storage, authClient)

// Set refresh threshold (when to refresh tokens)
manager.SetRefreshThreshold(30 * time.Minute)

// Start background refresh
stopRefresh := manager.StartBackgroundRefresh(15 * time.Minute)
defer stopRefresh() // Call when done

// Use the manager to get tokens (will refresh if needed)
tokenEntry, err := manager.GetToken(ctx, "user-id")
if err != nil {
    // Handle error
}

// Use the access token
accessToken := tokenEntry.TokenSet.AccessToken
```

## Security Considerations

When using the tokens package, keep the following security considerations in mind:

1. **Token Storage Security**:
   - For file-based storage, ensure the token directory has appropriate permissions
   - Consider encrypting tokens at rest for sensitive applications

2. **Client Secret Handling**:
   - Store client secrets securely and never commit them to source control
   - Use environment variables or a secure configuration mechanism

3. **Token Refresh**:
   - Set appropriate refresh thresholds to balance security and performance
   - Implement proper error handling for refresh failures

## Advanced Usage

For more advanced token management scenarios:

1. **Custom Storage Backend**:
   - Implement the `tokens.Storage` interface for custom storage solutions
   - Consider using a database or a distributed cache for multi-instance applications

2. **Error Handling**:
   - Implement retry logic for token refresh failures
   - Have a fallback mechanism for when tokens cannot be refreshed

3. **Token Scope Management**:
   - Store tokens with different scopes for different services
   - Use appropriate token keys to organize tokens by user and purpose
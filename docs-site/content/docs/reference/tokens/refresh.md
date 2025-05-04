---
title: "Tokens Package: Refresh"
---
# Tokens Package: Refresh

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The tokens package provides functionality for refreshing OAuth2 tokens automatically.

## Refresh Mechanism

The Token Manager automatically refreshes tokens when:

1. A token is requested via `GetToken` and it's expired or close to expiry
2. The background refresh process runs and finds tokens that are expired or close to expiry

## Refresh Handler Interface

```go
type RefreshHandler interface {
    RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, error)
}
```

The `RefreshHandler` interface defines a method for refreshing tokens. The Globus Auth client (`auth.Client`) implements this interface, so it can be used directly as a refresh handler.

## Setting a Refresh Handler

```go
// Create an auth client
authClient, err := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
)
if err != nil {
    // Handle error
}

// Create a token manager with the auth client as refresh handler
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithRefreshHandler(authClient),
)
if err != nil {
    // Handle error
}
```

## Refresh Threshold

The refresh threshold determines how close a token can be to expiry before it's refreshed:

```go
// Set the refresh threshold to 30 minutes
manager.SetRefreshThreshold(30 * time.Minute)
```

With a 30-minute threshold, tokens will be refreshed when they are less than 30 minutes from expiry.

You can also set the refresh threshold when creating the manager:

```go
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithRefreshHandler(authClient),
    tokens.WithRefreshThreshold(30 * time.Minute),
)
if err != nil {
    // Handle error
}
```

The default refresh threshold is 5 minutes.

## Token Refresh Flow

When the Token Manager needs to refresh a token, it follows this process:

1. Acquire a mutex to prevent concurrent refreshes of the same token
2. Check if the token has already been refreshed by another process
3. Call the refresh handler with the refresh token
4. Create a new token entry with the refreshed token
5. Store the refreshed token
6. Return the refreshed token

## Refresh Logic for `GetToken`

When you call `GetToken`, the Token Manager will:

1. Retrieve the token from storage
2. Check if it's expired or close to expiry (within the refresh threshold)
3. If it needs refreshing and can be refreshed (has a refresh token), refresh it
4. If refreshing fails but the token is still valid, return the original token
5. If refreshing fails and the token is expired, return an error
6. If the token doesn't need refreshing, return it as-is

## Implementing a Custom Refresh Handler

You can implement your own refresh handler by implementing the `RefreshHandler` interface:

```go
type MyRefreshHandler struct {
    // Your fields here
}

func (h *MyRefreshHandler) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, error) {
    // Your implementation here
    return &auth.TokenResponse{
        AccessToken:  "new-access-token",
        RefreshToken: "new-refresh-token",
        ExpiresIn:    3600, // 1 hour
        Scope:        "example-scope",
    }, nil
}
```

This is useful if you have a custom mechanism for refreshing tokens, such as:

- A token server that handles refresh token exchange
- A different OAuth2 provider
- A mock implementation for testing

## Error Handling

Token refresh can fail for several reasons:

1. The refresh token is invalid or expired
2. The refresh handler returns an error
3. Storing the refreshed token fails

When refresh fails, the behavior depends on the context:

- If called via `GetToken` and the original token is still valid, the original token is returned
- If called via `GetToken` and the original token is expired, an error is returned
- If called during background refresh, the error is logged but not propagated

## Best Practices

1. **Use the auth client as refresh handler**: The auth client is designed to handle token refresh
2. **Set a reasonable refresh threshold**: 5-30 minutes is a good range
3. **Handle refresh errors**: Check for errors when getting tokens
4. **Enable background refresh**: This keeps tokens fresh without explicit calls to `GetToken`
5. **Use a persistent storage mechanism**: This allows tokens to be refreshed across application restarts

## Example: Handling Refresh Errors

```go
// Get a token (which may trigger a refresh)
entry, err := manager.GetToken(ctx, "example-resource")
if err != nil {
    if strings.Contains(err.Error(), "failed to refresh token") {
        // Handle refresh failure (e.g., re-authenticate the user)
    } else {
        // Handle other errors
    }
    return err
}

// Use the token
accessToken := entry.TokenSet.AccessToken
```

## Example: Refreshing a Specific Token

If you want to explicitly refresh a token (regardless of whether it's close to expiry), you can use this approach:

```go
// Get the current token
entry, err := manager.GetToken(ctx, "example-resource")
if err != nil {
    // Handle error
}

// Create a modified entry with an expired expiry time
expiredEntry := &tokens.Entry{
    Resource:     entry.Resource,
    AccessToken:  entry.AccessToken,
    RefreshToken: entry.RefreshToken,
    ExpiresAt:    time.Now().Add(-1 * time.Hour), // Force expiry
    Scope:        entry.Scope,
    TokenSet: &tokens.TokenSet{
        AccessToken:  entry.AccessToken,
        RefreshToken: entry.RefreshToken,
        ExpiresAt:    time.Now().Add(-1 * time.Hour), // Force expiry
        Scope:        entry.Scope,
        ResourceID:   entry.Resource,
    },
}

// Store the expired entry
err = manager.StoreToken(ctx, expiredEntry)
if err != nil {
    // Handle error
}

// Get the token again (which will trigger a refresh)
refreshedEntry, err := manager.GetToken(ctx, "example-resource")
if err != nil {
    // Handle error
}

// Use the refreshed token
accessToken := refreshedEntry.TokenSet.AccessToken
```

Note: This is a bit of a hack and not the recommended approach. It's better to rely on the automatic refresh mechanism.

## Example: Auth Client as Refresh Handler

```go
// Create an auth client
authClient, err := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
)
if err != nil {
    // Handle error
}

// Create a token manager with the auth client as refresh handler
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithRefreshHandler(authClient),
)
if err != nil {
    // Handle error
}
```

## Implementation Details

The token refresh logic is implemented in the `refreshToken` method of the `Manager` struct:

```go
func (m *Manager) refreshToken(ctx context.Context, resource string, entry *Entry) (*Entry, error) {
    // Use a mutex to prevent multiple simultaneous refreshes for the same token
    m.refreshMutex.Lock()
    defer m.refreshMutex.Unlock()

    // Check if another goroutine already refreshed the token while we were waiting
    latestEntry, err := m.Storage.Lookup(resource)
    if err == nil && latestEntry != nil && latestEntry.AccessToken != entry.AccessToken &&
        latestEntry.TokenSet != nil && !latestEntry.TokenSet.IsExpired() &&
        time.Until(latestEntry.TokenSet.ExpiresAt) > m.RefreshThreshold {
        return latestEntry, nil
    }

    // Refresh the token
    tokenResponse, err := m.RefreshHandler.RefreshToken(ctx, entry.TokenSet.RefreshToken)
    if err != nil {
        return nil, err
    }

    // Calculate expiry time if not set directly in the response
    expiryTime := tokenResponse.ExpiryTime
    if expiryTime.IsZero() {
        expiryTime = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
    }

    // Create a new entry with the refreshed values
    refreshedEntry := &Entry{
        Resource:     resource,
        AccessToken:  tokenResponse.AccessToken,
        RefreshToken: tokenResponse.RefreshToken,
        ExpiresAt:    expiryTime,
        Scope:        tokenResponse.Scope,
    }

    // If the refresh token wasn't updated, use the original one
    if refreshedEntry.RefreshToken == "" {
        refreshedEntry.RefreshToken = entry.RefreshToken
    }

    // Create TokenSet for convenience
    refreshedEntry.TokenSet = &TokenSet{
        AccessToken:  refreshedEntry.AccessToken,
        RefreshToken: refreshedEntry.RefreshToken,
        ExpiresAt:    refreshedEntry.ExpiresAt,
        Scope:        refreshedEntry.Scope,
        ResourceID:   refreshedEntry.Resource,
    }

    // Store the refreshed token
    if err := m.Storage.Store(refreshedEntry); err != nil {
        return nil, fmt.Errorf("failed to store refreshed token: %w", err)
    }

    return refreshedEntry, nil
}
```
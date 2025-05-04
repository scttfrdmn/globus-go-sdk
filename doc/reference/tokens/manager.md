# Tokens Package: Manager

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Token Manager is responsible for managing OAuth2 tokens, including storage, retrieval, and automatic refreshing.

## Manager Structure

```go
type Manager struct {
    Storage          Storage
    RefreshThreshold time.Duration
    RefreshHandler   RefreshHandler
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Storage` | `Storage` | The storage mechanism for tokens |
| `RefreshThreshold` | `time.Duration` | The threshold before token expiry at which to trigger a refresh |
| `RefreshHandler` | `RefreshHandler` | Handler for refreshing tokens |

## Creating a Token Manager

```go
// Create a token manager with options
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithRefreshHandler(authClient),
    tokens.WithRefreshThreshold(30 * time.Minute),
)
if err != nil {
    // Handle error
}
```

### Options

| Option | Description |
|--------|-------------|
| `WithStorage(storage Storage)` | Sets the storage mechanism for tokens (required) |
| `WithFileStorage(path string)` | Creates and sets a file-based storage at the specified path |
| `WithMemoryStorage()` | Creates and sets a memory-based storage |
| `WithRefreshHandler(handler RefreshHandler)` | Sets the handler for refreshing tokens |
| `WithAuthClient(client *auth.Client)` | Sets an auth client as the refresh handler |
| `WithRefreshThreshold(threshold time.Duration)` | Sets the threshold before token expiry at which to trigger a refresh (default: 5 minutes) |

## Getting Tokens

```go
// Get a token, automatically refreshing it if needed
tokenEntry, err := manager.GetToken(ctx, "resource-name")
if err != nil {
    // Handle error
}

// Use the token
accessToken := tokenEntry.TokenSet.AccessToken
```

The `GetToken` method will:

1. Retrieve the token from storage
2. Check if it's expired or close to expiry
3. Refresh it if necessary and possible
4. Return the token (either the original or the refreshed one)

## Storing Tokens

```go
// Create a token entry
entry := &tokens.Entry{
    Resource: "resource-name",
    TokenSet: &tokens.TokenSet{
        AccessToken:  "access-token",
        RefreshToken: "refresh-token",
        ExpiresAt:    time.Now().Add(1 * time.Hour),
        Scope:        "scope1 scope2",
    },
}

// Store the token
err := manager.StoreToken(ctx, entry)
if err != nil {
    // Handle error
}
```

## Background Refresh

The token manager can automatically refresh tokens in the background:

```go
// Start background refresh (check every 15 minutes)
stop := manager.StartBackgroundRefresh(15 * time.Minute)

// Stop background refresh when done
defer stop()
```

The background refresh process will:

1. Periodically check all tokens in storage
2. Refresh any tokens that are expired or close to expiry
3. Update the storage with the refreshed tokens

## Refresh Threshold

The refresh threshold determines how close a token can be to expiry before it's refreshed:

```go
// Set the refresh threshold
manager.SetRefreshThreshold(30 * time.Minute)
```

With a 30-minute threshold, tokens will be refreshed when they are less than 30 minutes from expiry.

## Token Entry

```go
type Entry struct {
    Resource     string
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scope        string
    TokenSet     *TokenSet
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Resource` | `string` | A unique identifier for the token |
| `AccessToken` | `string` | The OAuth2 access token |
| `RefreshToken` | `string` | The OAuth2 refresh token |
| `ExpiresAt` | `time.Time` | When the access token expires |
| `Scope` | `string` | The scope of the token |
| `TokenSet` | `*TokenSet` | A convenience wrapper for the token |

## Token Set

```go
type TokenSet struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Scope        string
    ResourceID   string
}
```

### Token Set Methods

#### IsExpired

```go
// Check if a token is expired
if tokenSet.IsExpired() {
    // Token is expired
}
```

#### CanRefresh

```go
// Check if a token can be refreshed
if tokenSet.CanRefresh() {
    // Token can be refreshed
}
```

## Refresh Handler Interface

```go
type RefreshHandler interface {
    RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenResponse, error)
}
```

The `auth.Client` implements this interface, so it can be used directly as a refresh handler.

## Common Patterns

### Creating a Token Manager with File Storage

```go
// Create file storage
storage, err := tokens.NewFileStorage("~/.globus-tokens")
if err != nil {
    // Handle error
}

// Create a token manager
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithAuthClient(authClient),
)
if err != nil {
    // Handle error
}
```

### Creating a Token Manager with Memory Storage

```go
// Create memory storage
storage := tokens.NewMemoryStorage()

// Create a token manager
manager, err := tokens.NewManager(
    tokens.WithStorage(storage),
    tokens.WithAuthClient(authClient),
)
if err != nil {
    // Handle error
}
```

### Simplified Creation with Helper Functions

```go
// Create a token manager with file storage
manager, err := tokens.NewManager(
    tokens.WithFileStorage("~/.globus-tokens"),
    tokens.WithAuthClient(authClient),
)
if err != nil {
    // Handle error
}
```

### Storing Tokens from an Authorization Code Flow

```go
// Exchange authorization code for tokens
tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
if err != nil {
    // Handle error
}

// Create a token entry
entry := &tokens.Entry{
    Resource: "default",
    TokenSet: &tokens.TokenSet{
        AccessToken:  tokenResponse.AccessToken,
        RefreshToken: tokenResponse.RefreshToken,
        ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
        Scope:        tokenResponse.Scope,
        ResourceID:   "default",
    },
}

// Store the token
err = manager.StoreToken(ctx, entry)
if err != nil {
    // Handle error
}
```

### Getting a Token for a Service Client

```go
// Get the token
entry, err := manager.GetToken(ctx, "flows-service")
if err != nil {
    // Handle error
}

// Create a flows client with the token
flowsClient, err := flows.NewClient(
    flows.WithAccessToken(entry.TokenSet.AccessToken),
)
if err != nil {
    // Handle error
}
```

## Error Handling

The token manager can return the following errors:

- `"no storage provided"`: When creating a manager without a storage mechanism
- `"no token found for resource: %s"`: When trying to get a token that doesn't exist
- `"failed to refresh token: %w"`: When token refresh fails
- `"failed to store refreshed token: %w"`: When storing a refreshed token fails

## Best Practices

1. Always provide a storage mechanism when creating a token manager
2. Use file storage for persistent tokens across application restarts
3. Set a reasonable refresh threshold (5-30 minutes)
4. Enable background refresh to keep tokens fresh
5. Always check for errors when creating a token manager or getting tokens
6. Use the token manager to handle all token operations
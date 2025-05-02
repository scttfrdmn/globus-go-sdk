# Client Initialization Patterns

This document describes the preferred client initialization patterns for the Globus Go SDK (v0.8.0+).

## Options Pattern

The Globus Go SDK uses an options pattern for client initialization, which provides several benefits:

- More flexible and extensible API
- Better support for default values
- Clearer parameter naming
- Easier to add new options without breaking changes

## Basic Client Initialization

To create a new client for any Globus service, use the `NewClient` function with appropriate options:

```go
// Create a Transfer client with a static token
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
)
if err != nil {
    // Handle error
}
```

## Common Options

### Authentication Options

```go
// Static token authentication
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
)

// Client credentials authentication (Auth client)
authClient, err := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
)

// Create a client credentials authorizer
authorizer := authClient.CreateClientCredentialsAuthorizer(scopes...)

// Use the authorizer with another service
transferClient, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
```

### Core Options

Core options are available for all service clients:

```go
// Custom base URL (for testing or non-standard environments)
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
    transfer.WithBaseURL("https://custom-transfer-api.example.com/v0.10/"),
)

// HTTP debugging
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
    transfer.WithHTTPDebugging(true),
    transfer.WithHTTPTracing(true),
)
```

## Service-Specific Options

Each service may provide additional options specific to their functionality:

### Auth Options

```go
client, err := auth.NewClient(
    auth.WithClientID(clientID),
    auth.WithClientSecret(clientSecret),
    auth.WithRedirectURL("https://example.com/callback"),
)
```

### Transfer Options

```go
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
    transfer.WithDefaultTransferOptions(&transfer.TransferOptions{
        SyncLevel:      2,
        VerifyChecksum: true,
        Encrypt:        true,
    }),
)
```

## Error Handling

Client initialization may return an error if the provided options are invalid or if required options are missing:

```go
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
if err != nil {
    log.Fatalf("Failed to create transfer client: %v", err)
}
```

## Using Core Options Directly

You can also use core options directly for more advanced customization:

```go
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
    transfer.WithCoreOption(core.WithRequestTimeout(30 * time.Second)),
    transfer.WithCoreOption(core.WithUserAgent("MyApp/1.0")),
)
```

## Migration from Previous Versions

If you're updating from a previous version of the SDK, here's how to migrate your client initialization code:

### Before (pre-v0.8.0):

```go
// Old pattern
client := transfer.NewClient(accessToken)
```

### After (v0.8.0+):

```go
// New pattern
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizers.NewStaticTokenAuthorizer(accessToken)),
)
if err != nil {
    // Handle error
}
```

This new pattern provides better error handling, more flexibility, and a more consistent API across all services.
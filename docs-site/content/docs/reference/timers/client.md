---
title: "Timers Service: Client"
---
# Timers Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Timers client provides access to the Globus Timers API, which allows you to schedule and manage automated tasks at specific times or intervals.

## Client Structure

```go
type Client struct {
    Client *core.Client
}
```

The Timers client is a wrapper around the core client that provides specific methods for working with Globus Timers.

## Creating a Timers Client

```go
// Create a Timers client with functional options
client, err := timers.NewClient(
    timers.WithAccessToken("access-token"),
    timers.WithHTTPDebugging(true),
)
if err != nil {
    // Handle error
}

// Or create a Timers client using the SDK config
config := pkg.NewConfigFromEnvironment()
timersClient, err := config.NewTimersClient(accessToken)
if err != nil {
    // Handle error
}
```

## Configuration Options

The Timers client supports the following configuration options:

| Option | Description |
| ------ | ----------- |
| `WithAccessToken(token)` | Sets the access token for authorization |
| `WithBaseURL(url)` | Sets the base URL for the Timers API |
| `WithAuthorizer(authorizer)` | Sets a custom authorizer for the client |
| `WithHTTPDebugging(enable)` | Enables or disables HTTP debugging |
| `WithHTTPTracing(enable)` | Enables or disables HTTP tracing |
| `WithCoreOption(option)` | Applies additional core client options |

## Authorization Scope

The Timers client requires the following authorization scope:

```go
const TimersScope = "https://auth.globus.org/scopes/a1a171d5-48fb-4c77-a7ba-b8c628c20fd5/timers.api"
```

## Error Handling

The Timers client methods return errors for various failure conditions:

- Network communication errors
- Invalid parameters
- Unauthorized access
- Resource not found
- Validation errors
- Server errors

Example error handling:

```go
timer, err := client.GetTimer(ctx, timerID)
if err != nil {
    // Handle error - check for specific error types if needed
    return err
}
```

## Basic Usage Example

```go
// Get access token from environment variable
accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
if accessToken == "" {
    fmt.Println("Please set the GLOBUS_ACCESS_TOKEN environment variable")
    os.Exit(1)
}

// Create a new SDK configuration
config := pkg.NewConfigFromEnvironment()

// Create a new Timers client with the access token
timersClient, err := config.NewTimersClient(accessToken)
if err != nil {
    fmt.Printf("Error creating timers client: %v\n", err)
    os.Exit(1)
}

// Get information about the current user
user, err := timersClient.GetCurrentUser(ctx)
if err != nil {
    fmt.Printf("Error getting user information: %v\n", err)
} else {
    fmt.Printf("Current user: %s (ID: %s)\n", user.Username, user.ID)
}

// Create a one-time timer with a web callback
startTime := time.Now().Add(5 * time.Minute)
webCallback := timers.CreateWebCallback(
    "https://example.com/webhook", 
    "POST", 
    map[string]string{
        "Content-Type": "application/json",
    },
    nil,
)

webTimer, err := timersClient.CreateOnceTimer(
    ctx,
    "Example One-Time Timer",
    startTime,
    webCallback,
    map[string]interface{}{
        "description": "This is an example timer",
    },
)

if err != nil {
    fmt.Printf("Error creating timer: %v\n", err)
} else {
    fmt.Printf("Created timer with ID: %s\n", webTimer.ID)
    fmt.Printf("Timer will run at: %s\n", webTimer.NextDue.Format(time.RFC3339))
}
```
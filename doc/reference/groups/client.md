# Groups Service: Client

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Groups client provides access to the Globus Groups API, which allows you to create, manage, and interact with Globus Groups - a collaborative system for organizing users and controlling access to shared resources.

## Client Structure

```go
type Client struct {
    Client *core.Client
}
```

The Groups client is a wrapper around the core client that provides specific methods for working with Globus Groups.

## Creating a Groups Client

```go
// Create a Groups client with functional options
client, err := groups.NewClient(
    groups.WithAuthorizer(authorizer),
    groups.WithHTTPDebugging(true),
)
if err != nil {
    // Handle error
}

// Or create a Groups client using the SDK config
config := pkg.NewConfigFromEnvironment()
groupsClient, err := config.NewGroupsClient(accessToken)
if err != nil {
    // Handle error
}
```

## Configuration Options

The Groups client supports the following configuration options:

| Option | Description |
| ------ | ----------- |
| `WithAuthorizer(authorizer)` | Sets the authorizer for the client (required) |
| `WithHTTPDebugging(enable)` | Enables or disables HTTP debugging |
| `WithHTTPTracing(enable)` | Enables or disables HTTP tracing |
| `WithLogger(logger)` | Sets a custom logger for the client |
| `WithCoreOptions(options...)` | Applies additional core client options |

## Authorization Scope

The Groups client requires the following authorization scope:

```go
const GroupsScope = "urn:globus:auth:scope:groups.api.globus.org:all"
```

## Error Handling

The Groups client methods return errors for various failure conditions:

- Network communication errors
- Invalid parameters
- Unauthorized access
- Resource not found
- Validation errors
- Server errors

Example error handling:

```go
group, err := client.GetGroup(ctx, groupID)
if err != nil {
    // Handle error - check for specific error types if needed
    return err
}
```

## Basic Usage Example

```go
// Create a new SDK configuration
config := pkg.NewConfigFromEnvironment()

// Create a new Groups client with an access token
groupsClient, err := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create groups client: %v", err)
}

// List groups the user is a member of
groupList, err := groupsClient.ListGroups(context.Background(), &groups.ListGroupsOptions{
    MyGroups: true,
    PageSize: 100,
})
if err != nil {
    log.Fatalf("Failed to list groups: %v", err)
}

fmt.Printf("You are a member of %d groups:\n", len(groupList.Groups))
for _, group := range groupList.Groups {
    fmt.Printf("- %s (%s)\n", group.Name, group.ID)
}
```
---
title: "Transfer Service: Endpoint Operations"
---
# Transfer Service: Endpoint Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

Endpoints are access points for data in the Globus transfer ecosystem. The Transfer client provides methods for discovering, managing, and interacting with endpoints.

## Endpoint Structure

```go
type Endpoint struct {
    ID                     string   `json:"id"`
    DisplayName            string   `json:"display_name"`
    CanonicalName          string   `json:"canonical_name,omitempty"`
    Description            string   `json:"description,omitempty"`
    OwnerID                string   `json:"owner_id,omitempty"`
    OwnerString            string   `json:"owner_string,omitempty"`
    Organization           string   `json:"organization,omitempty"`
    Department             string   `json:"department,omitempty"`
    Keywords               string   `json:"keywords,omitempty"`
    ContactEmail           string   `json:"contact_email,omitempty"`
    ContactInfo            string   `json:"contact_info,omitempty"`
    InfoLink               string   `json:"info_link,omitempty"`
    SubscriptionID         string   `json:"subscription_id,omitempty"`
    DefaultDirectory       string   `json:"default_directory,omitempty"`
    Force_encryption       bool     `json:"force_encryption,omitempty"`
    Public                 bool     `json:"public,omitempty"`
    Activated              bool     `json:"activated,omitempty"`
    GlobusConnectSetupKey  string   `json:"globus_connect_setup_key,omitempty"`
    HighAssurance          bool     `json:"high_assurance,omitempty"`
    HostEndpointID         string   `json:"host_endpoint_id,omitempty"`
    HostPath               string   `json:"host_path,omitempty"`
    HostBoundPath          string   `json:"host_bound_path,omitempty"`
    LocalUserInfo          bool     `json:"local_user_info,omitempty"`
    GSSAudienceRequired    bool     `json:"gss_audience_required,omitempty"`
    IsGlobusConnect        bool     `json:"is_globus_connect,omitempty"`
    LocationVerification   bool     `json:"location_verification,omitempty"`
    ManagedEndpointID      string   `json:"managed_endpoint_id,omitempty"`
    ManagedByServiceID     string   `json:"managed_by_service_id,omitempty"`
    ManagedByServiceName   string   `json:"managed_by_service_name,omitempty"`
    MyTasks                string   `json:"my_tasks,omitempty"`
    NetworkUse             string   `json:"network_use,omitempty"`
    NonTransferable        bool     `json:"non_transferable,omitempty"`
    SharingTargetRoot      string   `json:"sharing_target_root,omitempty"`
    RequireVerification    bool     `json:"require_verification,omitempty"`
    SupportEmail           string   `json:"support_email,omitempty"`
    VirtualNFSRoot         string   `json:"virtual_nfs_root,omitempty"`
    InUse                  bool     `json:"in_use,omitempty"`
}
```

The Endpoint structure contains many fields, with the most commonly used being:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Unique identifier for the endpoint |
| `DisplayName` | `string` | Human-readable name for the endpoint |
| `Description` | `string` | Detailed description of the endpoint |
| `OwnerID` | `string` | Identity ID of the endpoint owner |
| `OwnerString` | `string` | Human-readable representation of the owner |
| `Public` | `bool` | Whether the endpoint is publicly accessible |
| `Activated` | `bool` | Whether the endpoint is currently activated |
| `DefaultDirectory` | `string` | Default directory when connecting to the endpoint |

## Listing Endpoints

Listing endpoints allows you to discover endpoints that you have access to.

```go
// List all endpoints
endpoints, err := client.ListEndpoints(ctx, nil)
if err != nil {
    // Handle error
}

// List with filtering options
endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    Filter:       "my-endpoints",
    Limit:        100,
    OwnerID:      "owner-id",
    SearchString: "cluster",
})
if err != nil {
    // Handle error
}

// Iterate through endpoints
for _, endpoint := range endpoints.DATA {
    fmt.Printf("Endpoint: %s (%s)\n", endpoint.DisplayName, endpoint.ID)
}
```

### ListEndpointsOptions

The `ListEndpointsOptions` struct provides filtering options for the endpoint listing:

```go
type ListEndpointsOptions struct {
    Filter       string `url:"filter,omitempty"`
    Limit        int    `url:"limit,omitempty"`
    Offset       int    `url:"offset,omitempty"`
    Fields       string `url:"fields,omitempty"`
    OwnerID      string `url:"owner_id,omitempty"`
    SearchString string `url:"search_string,omitempty"`
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Filter` | `string` | Filter for the results (e.g., "my-endpoints", "recently-used") |
| `Limit` | `int` | Maximum number of results to return |
| `Offset` | `int` | Offset for pagination |
| `Fields` | `string` | Comma-separated list of fields to include in the response |
| `OwnerID` | `string` | Filter by endpoint owner |
| `SearchString` | `string` | Search string to filter endpoints by name or description |

### Common Filters

The `Filter` parameter supports several pre-defined values:

- `"my-endpoints"` - Only endpoints owned by the current user
- `"recently-used"` - Recently used endpoints
- `"shared-by-me"` - Endpoints shared by the current user
- `"shared-with-me"` - Endpoints shared with the current user
- `"administered-by-me"` - Endpoints administered by the current user
- `"bookmarked"` - Bookmarked endpoints

```go
// List endpoints shared with me
endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    Filter: "shared-with-me",
})
```

## Getting an Endpoint

Retrieving a specific endpoint by ID:

```go
// Get a specific endpoint by ID
endpoint, err := client.GetEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

fmt.Printf("Endpoint Name: %s\n", endpoint.DisplayName)
fmt.Printf("Description: %s\n", endpoint.Description)
fmt.Printf("Owner: %s\n", endpoint.OwnerString)
```

## Activating an Endpoint

Some endpoints require activation before use. Activation usually involves an authentication step.

```go
// Activate an endpoint with auto-activation
activationResult, err := client.AutoActivateEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

if activationResult.Code == "AutoActivated" {
    fmt.Println("Endpoint activated successfully")
} else if activationResult.Code == "AutoActivationFailed" {
    fmt.Println("Auto-activation failed, manual activation required")
}
```

## Endpoint Activation Requirements

For endpoints that require manual activation, you can check the activation requirements:

```go
// Get activation requirements
requirements, err := client.GetEndpointActivationRequirements(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

if len(requirements.DATA) > 0 {
    fmt.Println("Activation requirements:")
    for _, req := range requirements.DATA {
        fmt.Printf("- %s: %s (Required: %t)\n", req.Type, req.Name, req.Required)
    }
}
```

## Manually Activating an Endpoint

For endpoints that require credentials:

```go
// Create activation data
activationData := &transfer.ActivationData{
    DATA: []transfer.ActivationRequirement{
        {
            Type:  "myproxy",
            Name:  "username",
            Value: "username",
        },
        {
            Type:  "myproxy",
            Name:  "passphrase",
            Value: "password",
        },
    },
}

// Activate the endpoint
activationResult, err := client.ActivateEndpoint(ctx, "endpoint-id", activationData)
if err != nil {
    // Handle error
}

if activationResult.Code == "Activated" {
    fmt.Println("Endpoint activated successfully")
}
```

## Deactivating an Endpoint

When you're done with an endpoint, you can deactivate it:

```go
// Deactivate an endpoint
deactivationResult, err := client.DeactivateEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

if deactivationResult.Code == "Deactivated" {
    fmt.Println("Endpoint deactivated successfully")
}
```

## Endpoint Management

### Searching for Endpoints

Searching for endpoints by name or description:

```go
// Search for endpoints
endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    SearchString: "science data",
})
if err != nil {
    // Handle error
}

fmt.Printf("Found %d matching endpoints\n", len(endpoints.DATA))
```

### Finding Recently Used Endpoints

```go
// Get recently used endpoints
endpoints, err := client.ListEndpoints(ctx, &transfer.ListEndpointsOptions{
    Filter: "recently-used",
    Limit:  5,
})
if err != nil {
    // Handle error
}

fmt.Println("Recently used endpoints:")
for _, endpoint := range endpoints.DATA {
    fmt.Printf("- %s (%s)\n", endpoint.DisplayName, endpoint.ID)
}
```

## Endpoint Properties

### Checking Endpoint Activation Status

```go
// Check if an endpoint is activated
endpoint, err := client.GetEndpoint(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

if endpoint.Activated {
    fmt.Println("Endpoint is activated")
} else {
    fmt.Println("Endpoint is not activated")
}
```

### Getting Endpoint Server Configuration

```go
// Get server configuration
config, err := client.GetEndpointServerConfiguration(ctx, "endpoint-id")
if err != nil {
    // Handle error
}

fmt.Printf("Server hostname: %s\n", config.Hostname)
fmt.Printf("Supports encryption: %t\n", config.SupportsEncryption)
```

## Working with Collections

Collections are subsets of endpoints with specific permissions.

### Listing Collections

```go
// List collections
collections, err := client.ListCollections(ctx, nil)
if err != nil {
    // Handle error
}

for _, collection := range collections.DATA {
    fmt.Printf("Collection: %s (%s)\n", collection.DisplayName, collection.ID)
}
```

### Getting a Collection

```go
// Get a specific collection
collection, err := client.GetCollection(ctx, "collection-id")
if err != nil {
    // Handle error
}

fmt.Printf("Collection Name: %s\n", collection.DisplayName)
fmt.Printf("Host Endpoint: %s\n", collection.HostEndpointID)
fmt.Printf("Path: %s\n", collection.HostPath)
```

## Endpoint Role Assignment

For managed endpoints, you can manage role assignments:

```go
// List endpoint role assignments
roles, err := client.ListEndpointRoleAssignments(ctx, "endpoint-id", nil)
if err != nil {
    // Handle error
}

for _, role := range roles.DATA {
    fmt.Printf("Role: %s - Principal: %s\n", role.Role, role.PrincipalString)
}
```

## Best Practices

1. **Cache Endpoint IDs**: Store frequently used endpoint IDs rather than searching for them repeatedly
2. **Check Activation**: Always check if an endpoint is activated before attempting transfers
3. **Use Auto-Activation**: Try auto-activation first before requesting credentials from users
4. **Deactivate When Done**: Deactivate endpoints when they're no longer needed
5. **Use Filters**: Use appropriate filters when listing endpoints to reduce the number of results
6. **Handle Pagination**: For large endpoint lists, handle pagination using limit and offset
7. **Verify Access**: Check that endpoints are accessible before attempting transfers
8. **Provide Context**: When displaying endpoints to users, include descriptive information
9. **Error Handling**: Handle common errors like "endpoint not found" or "activation required"
10. **Search Efficiently**: Use specific search terms when searching for endpoints
---
title: "Groups Service Quick Start"
weight: 60
---

# Groups Service Quick Start

This guide will help you get started with the Globus Groups service using the Go SDK. The Groups service allows you to create, manage, and work with Globus groups, including memberships and roles.

## Setup

First, import the required packages and create a context:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
)

func main() {
    // Create a context
    ctx := context.Background()
    
    // Continue with the examples below...
}
```

## Creating a Groups Client

There are two main ways to create a Groups client:

### Option 1: Using the SDK Configuration

```go
// Create a new SDK configuration from environment variables
config := pkg.NewConfigFromEnvironment()

// Create a new Groups client
groupsClient, err := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
if err != nil {
    log.Fatalf("Failed to create groups client: %v", err)
}
```

### Option 2: Using Functional Options

```go
// Create a new Groups client with options
import "github.com/scttfrdmn/globus-go-sdk/pkg/core/authorizers"

// Create a token authorizer
authorizer := authorizers.NewAccessTokenAuthorizer(os.Getenv("GLOBUS_ACCESS_TOKEN"))

// Create the groups client with options
groupsClient, err := groups.NewClient(
    groups.WithAuthorizer(authorizer),
    groups.WithHTTPDebugging(true),
)
if err != nil {
    log.Fatalf("Failed to create groups client: %v", err)
}
```

## Working with Groups

### Listing Groups

You can list groups that you are a member of:

```go
// List groups the user is a member of
groupList, err := groupsClient.ListGroups(ctx, &groups.ListGroupsOptions{
    MyGroups: true,
    PageSize: 100,
})
if err != nil {
    log.Fatalf("Failed to list groups: %v", err)
}

fmt.Printf("You are a member of %d groups:\n", len(groupList.Groups))
for _, group := range groupList.Groups {
    fmt.Printf("- %s (%s)\n", group.Name, group.ID)
    fmt.Printf("  Description: %s\n", group.Description)
    fmt.Printf("  Member count: %d\n", group.MemberCount)
}
```

### Listing with Pagination

If you need to handle large numbers of groups, you can use pagination:

```go
// Set up options for pagination
options := &groups.ListGroupsOptions{
    PageSize: 10, // Smaller page size
}

// Iterate through all pages
for {
    groupList, err := groupsClient.ListGroups(ctx, options)
    if err != nil {
        log.Fatalf("Failed to list groups: %v", err)
    }
    
    // Process the current page of groups
    for _, group := range groupList.Groups {
        fmt.Printf("- %s (%s)\n", group.Name, group.ID)
    }
    
    // Check if there are more pages
    if !groupList.HasNextPage {
        break
    }
    
    // Update the page token for the next page
    options.PageToken = groupList.NextPageToken
}
```

### Getting Group Details

You can retrieve details about a specific group:

```go
// Get details for a specific group
groupID := "your-group-id"
group, err := groupsClient.GetGroup(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to get group: %v", err)
}

fmt.Printf("Group: %s (%s)\n", group.Name, group.ID)
fmt.Printf("Description: %s\n", group.Description)
fmt.Printf("Member count: %d\n", group.MemberCount)
fmt.Printf("Created: %s\n", group.Created)
fmt.Printf("Last updated: %s\n", group.LastUpdated)
fmt.Printf("Public group: %t\n", group.PublicGroup)
```

### Creating a Group

To create a new group:

```go
// Create a new group
newGroup := &groups.GroupCreate{
    Name:        "My Test Group",
    Description: "A test group created using the Globus Go SDK",
    PublicGroup: true,
}

createdGroup, err := groupsClient.CreateGroup(ctx, newGroup)
if err != nil {
    log.Fatalf("Failed to create group: %v", err)
}

fmt.Printf("Created group: %s (%s)\n", createdGroup.Name, createdGroup.ID)
```

### Updating a Group

You can update a group's properties:

```go
// Update an existing group
updateRequest := &groups.GroupUpdate{
    Name:        "Updated Group Name",
    Description: "Updated description for the group",
}

updatedGroup, err := groupsClient.UpdateGroup(ctx, groupID, updateRequest)
if err != nil {
    log.Fatalf("Failed to update group: %v", err)
}

fmt.Printf("Updated group: %s (%s)\n", updatedGroup.Name, updatedGroup.ID)
fmt.Printf("Updated description: %s\n", updatedGroup.Description)
```

### Deleting a Group

To delete a group:

```go
// Delete a group
err = groupsClient.DeleteGroup(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to delete group: %v", err)
}

fmt.Println("Group deleted successfully.")
```

## Working with Members

### Listing Group Members

You can retrieve the members of a group:

```go
// List members of a group
memberList, err := groupsClient.ListMembers(ctx, groupID, &groups.ListMembersOptions{
    PageSize: 100,
})
if err != nil {
    log.Fatalf("Failed to list members: %v", err)
}

fmt.Printf("Group has %d members:\n", len(memberList.Members))
for _, member := range memberList.Members {
    fmt.Printf("- %s (%s)\n", member.Username, member.IdentityID)
    fmt.Printf("  Email: %s\n", member.Email)
    fmt.Printf("  Role: %s (%s)\n", member.Role.Name, member.Role.ID)
    fmt.Printf("  Status: %s\n", member.Status)
}
```

### Adding a Member

To add a new member to a group:

```go
// Add a member to a group
userID := "user@example.com" // This can be an email address or identity ID
roleID := "member"           // Common roles include "admin", "member"

err = groupsClient.AddMember(ctx, groupID, userID, roleID)
if err != nil {
    log.Fatalf("Failed to add member: %v", err)
}

fmt.Printf("Added %s to group with role %s\n", userID, roleID)
```

### Removing a Member

To remove a member from a group:

```go
// Remove a member from a group
err = groupsClient.RemoveMember(ctx, groupID, userID)
if err != nil {
    log.Fatalf("Failed to remove member: %v", err)
}

fmt.Printf("Removed %s from group\n", userID)
```

### Updating a Member's Role

You can change a member's role within a group:

```go
// Update a member's role
newRoleID := "admin"
err = groupsClient.UpdateMemberRole(ctx, groupID, userID, newRoleID)
if err != nil {
    log.Fatalf("Failed to update member role: %v", err)
}

fmt.Printf("Updated %s's role to %s\n", userID, newRoleID)
```

## Working with Roles

### Listing Roles

You can list the roles defined for a group:

```go
// List roles for a group
roleList, err := groupsClient.ListRoles(ctx, groupID)
if err != nil {
    log.Fatalf("Failed to list roles: %v", err)
}

fmt.Printf("Group has %d roles:\n", len(roleList.Roles))
for _, role := range roleList.Roles {
    fmt.Printf("- %s (%s)\n", role.Name, role.ID)
    fmt.Printf("  Description: %s\n", role.Description)
}
```

### Getting Role Details

To get details about a specific role:

```go
// Get details for a specific role
roleID := "admin"
role, err := groupsClient.GetRole(ctx, groupID, roleID)
if err != nil {
    log.Fatalf("Failed to get role: %v", err)
}

fmt.Printf("Role: %s (%s)\n", role.Name, role.ID)
fmt.Printf("Description: %s\n", role.Description)
```

### Creating a Custom Role

You can create custom roles in a group:

```go
// Create a new role
newRole := &groups.RoleCreate{
    Name:        "contributor",
    Description: "Can contribute to group resources but cannot administer",
}

createdRole, err := groupsClient.CreateRole(ctx, groupID, newRole)
if err != nil {
    log.Fatalf("Failed to create role: %v", err)
}

fmt.Printf("Created role: %s (%s)\n", createdRole.Name, createdRole.ID)
```

### Updating a Role

To update an existing role:

```go
// Update an existing role
updateRoleRequest := &groups.RoleUpdate{
    Name:        "updated-contributor",
    Description: "Updated description for the contributor role",
}

updatedRole, err := groupsClient.UpdateRole(ctx, groupID, customRoleID, updateRoleRequest)
if err != nil {
    log.Fatalf("Failed to update role: %v", err)
}

fmt.Printf("Updated role: %s (%s)\n", updatedRole.Name, updatedRole.ID)
```

### Deleting a Role

To delete a custom role:

```go
// Delete a role
err = groupsClient.DeleteRole(ctx, groupID, customRoleID)
if err != nil {
    log.Fatalf("Failed to delete role: %v", err)
}

fmt.Println("Role deleted successfully.")
```

## Complete Example

Here's a complete example that creates a group, adds a member, and then lists all members:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
)

func main() {
    // Create a context
    ctx := context.Background()
    
    // Create a new SDK configuration
    config := pkg.NewConfigFromEnvironment()
    
    // Create a new Groups client
    groupsClient, err := config.NewGroupsClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
    if err != nil {
        log.Fatalf("Failed to create groups client: %v", err)
    }
    
    // Create a new group with a timestamp to ensure uniqueness
    timestamp := time.Now().Format("20060102-150405")
    groupName := fmt.Sprintf("Test Group %s", timestamp)
    
    newGroup := &groups.GroupCreate{
        Name:        groupName,
        Description: "A test group created using the Globus Go SDK",
        PublicGroup: true,
    }
    
    // Create the group
    createdGroup, err := groupsClient.CreateGroup(ctx, newGroup)
    if err != nil {
        log.Fatalf("Failed to create group: %v", err)
    }
    
    fmt.Printf("Created group: %s (%s)\n", createdGroup.Name, createdGroup.ID)
    
    // Add a member to the group (use your own email or ID)
    userEmail := "your.email@example.com"
    roleID := "member"
    
    err = groupsClient.AddMember(ctx, createdGroup.ID, userEmail, roleID)
    if err != nil {
        log.Printf("Failed to add member: %v", err)
    } else {
        fmt.Printf("Added %s to group with role %s\n", userEmail, roleID)
    }
    
    // List the members of the group
    memberList, err := groupsClient.ListMembers(ctx, createdGroup.ID, &groups.ListMembersOptions{
        PageSize: 100,
    })
    if err != nil {
        log.Printf("Failed to list members: %v", err)
    } else {
        fmt.Printf("\nGroup has %d members:\n", len(memberList.Members))
        for _, member := range memberList.Members {
            fmt.Printf("- %s (%s)\n", member.Username, member.IdentityID)
            fmt.Printf("  Email: %s\n", member.Email)
            fmt.Printf("  Role: %s (%s)\n", member.Role.Name, member.Role.ID)
            fmt.Printf("  Status: %s\n", member.Status)
        }
    }
    
    // List the roles for the group
    roleList, err := groupsClient.ListRoles(ctx, createdGroup.ID)
    if err != nil {
        log.Printf("Failed to list roles: %v", err)
    } else {
        fmt.Printf("\nGroup has %d roles:\n", len(roleList.Roles))
        for _, role := range roleList.Roles {
            fmt.Printf("- %s (%s)\n", role.Name, role.ID)
            fmt.Printf("  Description: %s\n", role.Description)
        }
    }
    
    // Clean up: Delete the group
    fmt.Println("\nCleaning up - deleting the group...")
    err = groupsClient.DeleteGroup(ctx, createdGroup.ID)
    if err != nil {
        log.Printf("Failed to delete group: %v", err)
    } else {
        fmt.Println("Group deleted successfully.")
    }
}
```

## Error Handling

The Groups service methods return errors when operations fail. You can handle these errors by checking the error message:

```go
// Try to get a non-existent group
_, err = groupsClient.GetGroup(ctx, "non-existent-group-id")
if err != nil {
    if err.Error() == "request failed with status 404: " {
        fmt.Println("Group not found - check the group ID")
    } else if err.Error() == "request failed with status 403: " {
        fmt.Println("Permission denied - check your access token and permissions")
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

## Next Steps

Now that you understand the basics of the Groups service, you can:

1. **Implement Group Management**: Build applications that create and manage Globus groups for collaboration or access control
2. **Integrate with Other Services**: Use groups for access control with other Globus services like Transfer and Flows
3. **Create Custom Roles**: Define specialized roles for your organization's needs
4. **Automate Membership**: Build systems that automatically manage group membership based on external events or data

For more details, check out the [Groups Service API Reference](/docs/reference/groups/) documentation.
---
title: "Groups Service: Role Operations"
---
# Groups Service: Role Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Groups client provides methods for managing roles within groups, including listing, creating, updating, and deleting roles.

## Role Model

The central data type for role operations is the `Role` struct:

```go
type Role struct {
    DATA_TYPE   string `json:"DATA_TYPE"`
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

Related structs for creating and updating roles:

```go
// Used when creating a new role
type RoleCreate struct {
    DATA_TYPE   string `json:"DATA_TYPE"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
}

// Used when updating an existing role
type RoleUpdate struct {
    DATA_TYPE   string `json:"DATA_TYPE"`
    Name        string `json:"name,omitempty"`
    Description string `json:"description,omitempty"`
}
```

## Default Roles

Globus Groups typically come with default roles:

- **admin**: Full administrative access to the group, including member management
- **manager**: Can manage group membership but has limited administrative capabilities
- **member**: Basic group membership with limited permissions

## Listing Roles

To retrieve a list of roles defined for a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"

roleList, err := client.ListRoles(ctx, groupID)
if err != nil {
    // Handle error
}

// Process results
for _, role := range roleList.Roles {
    fmt.Printf("Role: %s (%s) - %s\n", 
        role.Name, role.ID, role.Description)
}
```

## Getting a Specific Role

To retrieve information about a specific role:

```go
groupID := "00000000-0000-0000-0000-000000000000"
roleID := "admin"  // Role ID to retrieve

role, err := client.GetRole(ctx, groupID, roleID)
if err != nil {
    // Handle error
}

fmt.Printf("Role: %s\nDescription: %s\n", 
    role.Name, role.Description)
```

## Creating a Role

To create a new custom role in a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"

newRole := &groups.RoleCreate{
    Name: "Reviewer",
    Description: "Can review content but cannot modify group settings",
}

createdRole, err := client.CreateRole(ctx, groupID, newRole)
if err != nil {
    // Handle error
}

fmt.Printf("Created role with ID: %s\n", createdRole.ID)
```

## Updating a Role

To update an existing role:

```go
groupID := "00000000-0000-0000-0000-000000000000"
roleID := "reviewer"  // Role ID to update

update := &groups.RoleUpdate{
    Description: "Updated description for reviewer role",
}

updatedRole, err := client.UpdateRole(ctx, groupID, roleID, update)
if err != nil {
    // Handle error
}

fmt.Printf("Updated role: %s\n", updatedRole.Name)
```

## Deleting a Role

To delete a custom role:

```go
groupID := "00000000-0000-0000-0000-000000000000"
roleID := "reviewer"  // Role ID to delete

err := client.DeleteRole(ctx, groupID, roleID)
if err != nil {
    // Handle error
}

fmt.Println("Role deleted successfully")
```

## Using Roles with Members

Roles are typically used when adding or updating members:

```go
// Add a member with a specific role
err := client.AddMember(ctx, groupID, userID, "admin")

// Update a member's role
err := client.UpdateMemberRole(ctx, groupID, userID, "member")
```

## Error Handling

Role operations can return the following types of errors:

- Validation errors (invalid group ID, role ID, or missing required fields)
- Authentication errors (insufficient permissions)
- Resource not found errors
- Default role modification errors (some default roles can't be modified)
- API communication errors

Example error handling:

```go
role, err := client.GetRole(ctx, groupID, roleID)
if err != nil {
    if strings.Contains(err.Error(), "404") {
        // Role not found
        fmt.Printf("Role %s does not exist in group %s\n", roleID, groupID)
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied
        fmt.Println("You don't have permission to access this role")
    } else {
        // Other error
        fmt.Printf("Error retrieving role: %v\n", err)
    }
}
```

## Provisioning Rules

Roles can also be used with provisioning rules to automatically assign roles to members:

```go
// Provision rule model
type ProvisionRule struct {
    DATA_TYPE    string `json:"DATA_TYPE"`
    ID           string `json:"id"`
    Label        string `json:"label"`
    Expression   string `json:"expression"`
    MappedRoleID string `json:"mapped_role_id"`
}

// Creating a provision rule
ruleCreate := &groups.ProvisionRuleCreate{
    Label: "Organization Members",
    Expression: "identity.organization == 'Globus'",
    MappedRoleID: "member",
}
```

Provisioning rules allow for automatic role assignment based on identity attributes, enabling more dynamic group membership management.
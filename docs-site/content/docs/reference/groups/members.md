---
title: "Groups Service: Member Operations"
---
# Groups Service: Member Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Groups client provides methods for managing group membership, including listing, adding, removing, and updating member roles.

## Member Model

The central data type for member operations is the `Member` struct:

```go
type Member struct {
    DATA_TYPE         string    `json:"DATA_TYPE"`
    IdentityID        string    `json:"identity_id"`
    Username          string    `json:"username"`
    Email             string    `json:"email"`
    Status            string    `json:"status"`
    RoleID            string    `json:"role_id"`
    Role              Role      `json:"role"`
    // Additional fields
    Name              string    `json:"name,omitempty"`
    Organization      string    `json:"organization,omitempty"`
    JoinedDate        time.Time `json:"joined_date,omitempty"`
    LastUpdateDate    time.Time `json:"last_update_date,omitempty"`
    ProvisionedByRule string    `json:"provisioned_by_rule,omitempty"`
}
```

The possible values for the `Status` field include:
- `"active"` - Member is active in the group
- `"invited"` - User has been invited but hasn't accepted
- `"denied"` - Member request or invitation was denied
- `"requested"` - User has requested to join the group

## Listing Group Members

To retrieve a list of members in a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"

// Create options for listing members
options := &groups.ListMembersOptions{
    PageSize: 50,          // Number of members per page
    // Additional filtering options
    RoleID: "",            // Filter by specific role ID
    Status: "active",      // Filter by status: active, invited, etc.
    PageToken: "",         // Token for pagination
}

// List members
memberList, err := client.ListMembers(ctx, groupID, options)
if err != nil {
    // Handle error
}

// Process results
for _, member := range memberList.Members {
    fmt.Printf("Member: %s (%s) - Role: %s\n", 
        member.Username, member.IdentityID, member.Role.Name)
}

// Check if there are more pages
if memberList.HasNextPage {
    // Use the NextPageToken to get the next page
    options.PageToken = memberList.NextPageToken
    nextPage, err := client.ListMembers(ctx, groupID, options)
    // Process next page...
}
```

## Adding a Member to a Group

To add a user to a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"
userID := "user@example.com"  // Can be email, username, or Globus Identity ID
roleID := "member"            // Role ID, typically "admin" or "member"

err := client.AddMember(ctx, groupID, userID, roleID)
if err != nil {
    // Handle error
}

fmt.Println("Member added successfully")
```

## Removing a Member from a Group

To remove a user from a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"
userID := "user@example.com"  // Can be email, username, or Globus Identity ID

err := client.RemoveMember(ctx, groupID, userID)
if err != nil {
    // Handle error
}

fmt.Println("Member removed successfully")
```

## Updating a Member's Role

To change a member's role within a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"
userID := "user@example.com"
newRoleID := "admin"  // Change role to admin

err := client.UpdateMemberRole(ctx, groupID, userID, newRoleID)
if err != nil {
    // Handle error
}

fmt.Println("Member role updated successfully")
```

## Low-Level Member Operations

The Groups client also provides lower-level methods for more advanced member operations:

```go
// Low-level member listing with more direct control
memberList, err := client.ListMembersLowLevel(ctx, groupID, options)

// Low-level member addition 
err := client.AddMemberLowLevel(ctx, groupID, userID, roleID)

// Low-level member removal
err := client.RemoveMemberLowLevel(ctx, groupID, userID)

// Low-level role update
err := client.UpdateMemberRoleLowLevel(ctx, groupID, userID, roleID)
```

These low-level methods provide more direct access to the underlying API for advanced use cases.

## Error Handling

Member operations can return the following types of errors:

- Validation errors (invalid group ID, user ID, or role ID)
- Authentication errors (insufficient permissions)
- Resource not found errors (group or user doesn't exist)
- Role validation errors (invalid role for the group)
- API communication errors

Example error handling:

```go
err := client.AddMember(ctx, groupID, userID, roleID)
if err != nil {
    if strings.Contains(err.Error(), "404") {
        // Group or user not found
        fmt.Println("Group or user not found")
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied
        fmt.Println("You don't have permission to add members to this group")
    } else {
        // Other error
        fmt.Printf("Error adding member: %v\n", err)
    }
}
```
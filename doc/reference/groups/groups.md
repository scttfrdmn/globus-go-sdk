# Groups Service: Group Operations

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

The Groups client provides methods for creating, retrieving, updating, and deleting Globus Groups.

## Group Model

The central data type for Groups operations is the `Group` struct:

```go
type Group struct {
    DATA_TYPE             string    `json:"DATA_TYPE"`
    ID                    string    `json:"id"`
    Name                  string    `json:"name"`
    Description           string    `json:"description"`
    ParentID              string    `json:"parent_id,omitempty"`
    IdentityID            string    `json:"identity_id"`
    MemberCount           int       `json:"member_count"`
    IsGroupAdmin          bool      `json:"is_group_admin"`
    IsMember              bool      `json:"is_member"`
    Created               time.Time `json:"created"`
    LastUpdated           time.Time `json:"last_updated"`
    PublicGroup           bool      `json:"public_group"`
    RequiresSignAgreement bool      `json:"requires_sign_agreement"`
    SignAgreementMessage  string    `json:"sign_agreement_message,omitempty"`
    Policies              map[string]interface{} `json:"policies,omitempty"`
    EnforceProvisionRules bool              `json:"enforce_provision_rules,omitempty"`
    ProvisionRules        []ProvisionRule   `json:"provision_rules,omitempty"`
}
```

Related structs for creating and updating groups:

```go
// Used when creating a new group
type GroupCreate struct {
    DATA_TYPE             string            `json:"DATA_TYPE"`
    Name                  string            `json:"name"`
    Description           string            `json:"description,omitempty"`
    ParentID              string            `json:"parent_id,omitempty"`
    PublicGroup           bool              `json:"public_group,omitempty"`
    RequiresSignAgreement bool              `json:"requires_sign_agreement,omitempty"`
    SignAgreementMessage  string            `json:"sign_agreement_message,omitempty"`
    EnforceProvisionRules bool              `json:"enforce_provision_rules,omitempty"`
    Policies              map[string]interface{} `json:"policies,omitempty"`
}

// Used when updating an existing group
type GroupUpdate struct {
    DATA_TYPE             string            `json:"DATA_TYPE"`
    Name                  string            `json:"name,omitempty"`
    Description           string            `json:"description,omitempty"`
    ParentID              string            `json:"parent_id,omitempty"`
    PublicGroup           *bool             `json:"public_group,omitempty"`
    RequiresSignAgreement *bool             `json:"requires_sign_agreement,omitempty"`
    SignAgreementMessage  string            `json:"sign_agreement_message,omitempty"`
    EnforceProvisionRules *bool             `json:"enforce_provision_rules,omitempty"`
    Policies              map[string]interface{} `json:"policies,omitempty"`
}
```

## Listing Groups

To retrieve a list of groups that the current user is a member of:

```go
// Create options for listing groups
options := &groups.ListGroupsOptions{
    MyGroups: true,        // Only show groups where I'm a member
    PageSize: 50,          // Number of groups per page
    // Additional filtering options
    IncludeGroupMembership: false,  // Include full membership details
    IncludeIdentitySet: false,      // Include identity set information
    ForUserID: "",                  // Filter for a specific user ID
    PageToken: "",                  // Token for pagination
}

// List groups
groupList, err := client.ListGroups(ctx, options)
if err != nil {
    // Handle error
}

// Process results
for _, group := range groupList.Groups {
    fmt.Printf("Group: %s (%s)\n", group.Name, group.ID)
}

// Check if there are more pages
if groupList.HasNextPage {
    // Use the NextPageToken to get the next page
    options.PageToken = groupList.NextPageToken
    nextPage, err := client.ListGroups(ctx, options)
    // Process next page...
}
```

## Getting a Specific Group

To retrieve information about a specific group:

```go
groupID := "00000000-0000-0000-0000-000000000000"
group, err := client.GetGroup(ctx, groupID)
if err != nil {
    // Handle error
}

fmt.Printf("Group: %s\nDescription: %s\nMembers: %d\n", 
    group.Name, group.Description, group.MemberCount)
```

## Creating a Group

To create a new group:

```go
newGroup := &groups.GroupCreate{
    Name: "Research Team Alpha",
    Description: "Group for coordinating research activities",
    PublicGroup: false,  // Private group
    Policies: map[string]interface{}{
        "is_high_assurance": true,
    },
}

createdGroup, err := client.CreateGroup(ctx, newGroup)
if err != nil {
    // Handle error
}

fmt.Printf("Created group with ID: %s\n", createdGroup.ID)
```

## Updating a Group

To update an existing group:

```go
groupID := "00000000-0000-0000-0000-000000000000"

// Create an update request
// Note: Use pointers for boolean fields to distinguish between 
// "not set" and explicit true/false values
isPublic := true
update := &groups.GroupUpdate{
    Description: "Updated description for the research group",
    PublicGroup: &isPublic,
}

updatedGroup, err := client.UpdateGroup(ctx, groupID, update)
if err != nil {
    // Handle error
}

fmt.Printf("Updated group: %s\n", updatedGroup.Name)
```

## Deleting a Group

To delete a group:

```go
groupID := "00000000-0000-0000-0000-000000000000"

err := client.DeleteGroup(ctx, groupID)
if err != nil {
    // Handle error
}

fmt.Println("Group deleted successfully")
```

## Error Handling

Group operations can return the following types of errors:

- Validation errors (invalid group ID, missing required fields)
- Authentication errors (insufficient permissions)
- Resource not found errors
- API communication errors

Example error handling:

```go
group, err := client.GetGroup(ctx, groupID)
if err != nil {
    if strings.Contains(err.Error(), "404") {
        // Group not found
        fmt.Printf("Group %s does not exist\n", groupID)
    } else if strings.Contains(err.Error(), "403") {
        // Permission denied
        fmt.Println("You don't have permission to access this group")
    } else {
        // Other error
        fmt.Printf("Error retrieving group: %v\n", err)
    }
}
```
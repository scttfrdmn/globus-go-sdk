// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package groups provides a client for interacting with the Globus Groups service.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Client interface and implementation
  - Group management methods (ListGroups, GetGroup, CreateGroup, etc.)
  - Membership management methods (ListMembers, AddMember, RemoveMember, etc.)
  - Role management methods (ListRoles, GetRole, UpdateRole, etc.)
  - Core model types (Group, Member, Role, etc.)
  - Client configuration options (WithAuthorizer, WithBaseURL, etc.)

Methods marked as "LowLevel" in the code are considered internal and
may change in future versions.

# Compatibility Guarantees

For stable packages:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create a new groups client:

	groupsClient := groups.NewClient(
		groups.WithAuthorizer(authorizer),
	)

List groups:

	groupList, err := groupsClient.ListGroups(ctx, nil)
	if err != nil {
		// Handle error
	}

	for _, group := range groupList.Groups {
		fmt.Printf("Group ID: %s, Name: %s\n", group.ID, group.Name)
	}

Get a specific group:

	group, err := groupsClient.GetGroup(ctx, "group_id")
	if err != nil {
		// Handle error
	}

	fmt.Printf("Group: %s (%s)\n", group.Name, group.Description)

Create a group:

	newGroup := &groups.GroupCreate{
		Name:        "My New Group",
		Description: "A group for my team",
		Visibility:  "private",
	}

	created, err := groupsClient.CreateGroup(ctx, newGroup)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Created group with ID: %s\n", created.ID)

Update a group:

	update := &groups.GroupUpdate{
		Description: "Updated description",
	}

	updated, err := groupsClient.UpdateGroup(ctx, "group_id", update)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Updated group: %s\n", updated.Name)

Delete a group:

	err := groupsClient.DeleteGroup(ctx, "group_id")
	if err != nil {
		// Handle error
	}

# Membership Management

List members:

	memberList, err := groupsClient.ListMembers(ctx, "group_id", nil)
	if err != nil {
		// Handle error
	}

	for _, member := range memberList.Members {
		fmt.Printf("Member: %s, Role: %s\n", member.Username, member.RoleID)
	}

Add a member:

	err := groupsClient.AddMember(ctx, "group_id", "user_id", "role_id")
	if err != nil {
		// Handle error
	}

Remove a member:

	err := groupsClient.RemoveMember(ctx, "group_id", "user_id")
	if err != nil {
		// Handle error
	}

Update a member's role:

	err := groupsClient.UpdateMemberRole(ctx, "group_id", "user_id", "new_role_id")
	if err != nil {
		// Handle error
	}

# Role Management

List roles:

	roles, err := groupsClient.ListRoles(ctx, "group_id")
	if err != nil {
		// Handle error
	}

	for _, role := range roles.Roles {
		fmt.Printf("Role: %s (%s)\n", role.Name, role.ID)
	}

Create a role:

	newRole := &groups.Role{
		Name:        "Custom Role",
		Description: "A custom role with specific permissions",
		// Set permissions...
	}

	created, err := groupsClient.CreateRole(ctx, "group_id", newRole)
	if err != nil {
		// Handle error
	}

	fmt.Printf("Created role with ID: %s\n", created.ID)
*/
package groups

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
)

// RoleChange represents a request to change a member's role in a group
// v3.65.0: Added for batch role change operations
type RoleChange struct {
	GroupID    string `json:"group_id"`
	IdentityID string `json:"identity_id"`
	RoleID     string `json:"role_id"`
}

// BatchRoleChangeRequest represents a batch role change request
type BatchRoleChangeRequest struct {
	DATA_TYPE string       `json:"DATA_TYPE"`
	Changes   []RoleChange `json:"changes"`
}

// BatchRoleChangeResult represents the result of a batch role change operation
type BatchRoleChangeResult struct {
	DATA_TYPE string                     `json:"DATA_TYPE"`
	Results   []BatchRoleChangeItemResult `json:"results"`
}

// BatchRoleChangeItemResult represents the result of a single role change
type BatchRoleChangeItemResult struct {
	GroupID    string `json:"group_id"`
	IdentityID string `json:"identity_id"`
	RoleID     string `json:"role_id"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

// ChangeRole changes a member's role in a group
// v3.65.0: Added for single role change operation
func (c *Client) ChangeRole(ctx context.Context, groupID, identityID, roleID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if identityID == "" {
		return fmt.Errorf("identity ID is required")
	}
	if roleID == "" {
		return fmt.Errorf("role ID is required")
	}

	// Use the existing update member role endpoint
	path := fmt.Sprintf("groups/%s/members/%s", groupID, identityID)

	updateReq := map[string]interface{}{
		"DATA_TYPE": "group_membership_update",
		"role_id":   roleID,
	}

	var result map[string]interface{}
	err := c.doRequestLowLevel(ctx, "PUT", path, nil, updateReq, &result)
	if err != nil {
		return fmt.Errorf("failed to change role: %w", err)
	}

	return nil
}

// ChangeRoles performs batch role changes across multiple groups/members
// v3.65.0: Added for batch role change operations
func (c *Client) ChangeRoles(ctx context.Context, changes []RoleChange) (*BatchRoleChangeResult, error) {
	if len(changes) == 0 {
		return nil, fmt.Errorf("at least one role change is required")
	}

	// Note: The actual Globus API might not have a dedicated batch endpoint.
	// This implementation performs changes sequentially and aggregates results.
	// In production, check if Globus has added a true batch endpoint.

	results := make([]BatchRoleChangeItemResult, 0, len(changes))

	for _, change := range changes {
		err := c.ChangeRole(ctx, change.GroupID, change.IdentityID, change.RoleID)

		result := BatchRoleChangeItemResult{
			GroupID:    change.GroupID,
			IdentityID: change.IdentityID,
			RoleID:     change.RoleID,
			Success:    err == nil,
		}

		if err != nil {
			result.Error = err.Error()
		}

		results = append(results, result)
	}

	return &BatchRoleChangeResult{
		DATA_TYPE: "batch_role_change_result",
		Results:   results,
	}, nil
}

// BatchMembershipActions provides a builder pattern for batch membership operations
// v3.65.0: Added for fluent API batch operations
type BatchMembershipActions struct {
	client  *Client
	changes []RoleChange
}

// NewBatchMembershipActions creates a new batch membership actions builder
func (c *Client) NewBatchMembershipActions() *BatchMembershipActions {
	return &BatchMembershipActions{
		client:  c,
		changes: make([]RoleChange, 0),
	}
}

// ChangeRole adds a role change to the batch
// v3.65.0: Fluent API for building batch operations
func (b *BatchMembershipActions) ChangeRole(groupID, identityID, roleID string) *BatchMembershipActions {
	b.changes = append(b.changes, RoleChange{
		GroupID:    groupID,
		IdentityID: identityID,
		RoleID:     roleID,
	})
	return b
}

// Execute performs all queued role changes
func (b *BatchMembershipActions) Execute(ctx context.Context) (*BatchRoleChangeResult, error) {
	if len(b.changes) == 0 {
		return &BatchRoleChangeResult{
			DATA_TYPE: "batch_role_change_result",
			Results:   []BatchRoleChangeItemResult{},
		}, nil
	}

	return b.client.ChangeRoles(ctx, b.changes)
}

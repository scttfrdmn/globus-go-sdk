// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
)

// SubscriptionUpdate represents an update to a group's subscription
type SubscriptionUpdate struct {
	DATA_TYPE      string `json:"DATA_TYPE"`
	SubscriptionID string `json:"subscription_id"`
}

// SetSubscriptionAdminVerified sets the subscription ID for a group
// and marks it as admin-verified.
//
// v3.63.0: Renamed from set_subscription_admin_verified_id in v3.62.0
// v3.62.0: Originally added as set_subscription_admin_verified_id
//
// This method allows administrators to associate a group with a subscription
// and mark the association as verified by an administrator.
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID, subscriptionID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID is required")
	}

	path := fmt.Sprintf("groups/%s/subscription", groupID)

	updateReq := &SubscriptionUpdate{
		DATA_TYPE:      "subscription_update",
		SubscriptionID: subscriptionID,
	}

	var result map[string]interface{}
	err := c.doRequestLowLevel(ctx, "PUT", path, nil, updateReq, &result)
	if err != nil {
		return fmt.Errorf("failed to set subscription: %w", err)
	}

	return nil
}

// GetSubscription retrieves the subscription associated with a group
func (c *Client) GetSubscription(ctx context.Context, groupID string) (string, error) {
	if groupID == "" {
		return "", fmt.Errorf("group ID is required")
	}

	path := fmt.Sprintf("groups/%s/subscription", groupID)

	var result struct {
		DATA_TYPE      string `json:"DATA_TYPE"`
		SubscriptionID string `json:"subscription_id"`
	}

	err := c.doRequestLowLevel(ctx, "GET", path, nil, nil, &result)
	if err != nil {
		return "", fmt.Errorf("failed to get subscription: %w", err)
	}

	return result.SubscriptionID, nil
}

// RemoveSubscription removes the subscription association from a group
func (c *Client) RemoveSubscription(ctx context.Context, groupID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	path := fmt.Sprintf("groups/%s/subscription", groupID)

	var result map[string]interface{}
	err := c.doRequestLowLevel(ctx, "DELETE", path, nil, nil, &result)
	if err != nil {
		return fmt.Errorf("failed to remove subscription: %w", err)
	}

	return nil
}

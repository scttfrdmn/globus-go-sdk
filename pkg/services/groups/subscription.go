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

// SetSubscriptionAdminVerified sets a subscription ID for a group (admin-only operation).
// This method follows Python SDK v3.63.0 naming convention.
// Renamed from SetSubscriptionAdminVerifiedID in v3.63.0.
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID, subscriptionID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if subscriptionID == "" {
		return fmt.Errorf("subscription ID is required")
	}

	body := map[string]string{
		"subscription_id": subscriptionID,
		"DATA_TYPE":       "subscription_update",
	}

	return c.doRequestLowLevel(ctx, "PUT", "groups/"+groupID+"/subscription", nil, body, nil)
}

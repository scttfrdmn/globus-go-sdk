// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package groups

import (
	"context"
	"fmt"
)

// SetSubscriptionAdminVerified sets (or clears) the subscription that
// admin-verifies a group (PUT /groups/{id}/subscription_admin_verified). Pass an
// empty subscriptionID to disassociate (sends a JSON null).
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID, subscriptionID string) error {
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	var body map[string]interface{}
	if subscriptionID != "" {
		body = map[string]interface{}{"subscription_admin_verified_id": subscriptionID}
	} else {
		body = map[string]interface{}{"subscription_admin_verified_id": nil}
	}
	return c.doRequestLowLevel(ctx, "PUT", "groups/"+groupID+"/subscription_admin_verified", nil, body, nil)
}

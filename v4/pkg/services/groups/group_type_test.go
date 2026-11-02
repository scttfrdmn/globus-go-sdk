// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// Entries mirror upstream globus-sdk-python 4.10.0 fixtures:
// testing/data/groups/get_my_groups.py (subscription_id absent),
// _common.py SUBSCRIPTION_GROUP_DOC ("plus", subscription set) and
// BASE_GROUP_DOC (subscription_id null). Issue #84.
const myGroupsFixture = `[
  {"name": "Claptrap's Rough Riders", "parent_id": null,
   "id": "d3974728-6458-11e4-b72d-123139141556", "group_type": "regular",
   "enforce_session": false, "session_limit": 28800, "session_timeouts": {},
   "my_memberships": [{"group_id": "d3974728-6458-11e4-b72d-123139141556",
     "identity_id": "ae332d86-d274-11e5-b885-b31714a110e9",
     "username": "sirosen@globusid.org", "role": "member", "status": "active"}],
   "policies": {"group_visibility": "private", "group_members_visibility": "managers"}},
  {"name": "University of Chicago, Department of Magical Automata",
   "description": "The finest in machines that go 'ping!' and 'whomp!' and 'bzzt!'",
   "parent_id": null, "id": "5e8f6a3a-9b1d-11ee-8c90-0242ac120002", "group_type": "plus",
   "enforce_session": true, "session_limit": 1800, "session_timeouts": {},
   "my_memberships": [{"group_id": "5e8f6a3a-9b1d-11ee-8c90-0242ac120002",
     "identity_id": "ae332d86-d274-11e5-b885-b31714a110e9",
     "username": "claptrap@globus.org", "role": "admin", "status": "active"}],
   "policies": {"group_visibility": "private", "group_members_visibility": "managers"},
   "subscription_id": "6a1c2e4e-9b1d-11ee-8c90-0242ac120002",
   "subscription_info": {"name": "University of Chicago"}},
  {"name": "No Subscription", "description": "No stairs allowed.", "parent_id": null,
   "id": "7c580b9a-4592-11e3-a2a0-12313d2d6e7f", "group_type": "regular",
   "my_memberships": [], "subscription_id": null, "subscription_info": null}
]`

func TestGetMyGroups_GroupTypeAndSubscriptionID(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/groups/my_groups", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, myGroupsFixture)
	})
	c, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	require.NoError(t, err)
	defer c.Close()

	gs, err := c.GetMyGroups(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, gs, 3)

	assert.Equal(t, "regular", gs[0].GroupType)
	assert.Empty(t, gs[0].SubscriptionID, "absent subscription_id")

	assert.Equal(t, "plus", gs[1].GroupType)
	assert.Equal(t, "6a1c2e4e-9b1d-11ee-8c90-0242ac120002", gs[1].SubscriptionID)
	require.Len(t, gs[1].MyMemberships, 1)
	assert.Equal(t, "admin", gs[1].MyMemberships[0].Role)

	assert.Equal(t, "regular", gs[2].GroupType)
	assert.Empty(t, gs[2].SubscriptionID, "null subscription_id")
}

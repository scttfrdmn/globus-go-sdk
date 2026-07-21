// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package groups_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// TestGetMyGroupsWithOptions_MyMembershipsRole verifies that include=my_memberships
// is sent and that the caller's role decodes into Group.MyMemberships — the fix
// for the manager-vs-member gap (issue #50).
func TestGetMyGroupsWithOptions_MyMembershipsRole(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("include"); got != "my_memberships" {
			t.Errorf("include = %q, want my_memberships", got)
		}
		if got := r.URL.Query().Get("statuses"); got != "active" {
			t.Errorf("statuses = %q, want active", got)
		}
		testhelpers.RespondJSON(w, http.StatusOK, []map[string]interface{}{
			{
				"id":   "grp-1",
				"name": "Managed Group",
				"my_memberships": []map[string]interface{}{
					{"identity_id": "id-1", "role": "manager", "status": "active"},
				},
			},
		})
	})
	client, err := groups.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	out, err := client.GetMyGroupsWithOptions(context.Background(), &groups.GetMyGroupsOptions{
		Statuses: []string{"active"},
		Include:  []string{"my_memberships"},
	})
	if err != nil {
		t.Fatalf("GetMyGroupsWithOptions: %v", err)
	}
	if len(out) != 1 || len(out[0].MyMemberships) != 1 {
		b, _ := json.Marshal(out)
		t.Fatalf("unexpected result: %s", b)
	}
	if role := out[0].MyMemberships[0].Role; role != "manager" {
		t.Errorf("caller role = %q, want manager (must be distinguishable from member)", role)
	}
}

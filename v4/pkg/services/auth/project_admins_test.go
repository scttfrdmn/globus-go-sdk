// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth_test

import (
	"encoding/json"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// TestProjectAdminsDecode is a regression test for issue #61: the Globus Auth
// Projects API returns admins.identities (and admins.groups) as arrays of
// objects, not ID strings. The previous []string typing failed to decode any
// project carrying the expanded admins object, blocking `project list`/`show`.
func TestProjectAdminsDecode(t *testing.T) {
	// Shape taken from docs.globus.org/api/auth/reference (Projects Resource).
	body := `{
		"id": "927c2fbf-e962-4207-8f8e-2c48e7fd3f00",
		"display_name": "Demo Project",
		"contact_email": "admin@example.org",
		"admin_ids": ["ae341a98-d274-11e5-b888-dbae3a8ba545"],
		"admins": {
			"identities": [
				{
					"id": "ae341a98-d274-11e5-b888-dbae3a8ba545",
					"username": "admin@example.org",
					"name": "Demo Admin",
					"email": "admin@example.org",
					"identity_provider": "927d7238-f917-4eb2-9ace-c523a2af1234",
					"identity_type": "login",
					"organization": "Example Org",
					"status": "used"
				}
			],
			"groups": []
		}
	}`

	var p auth.Project
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatalf("failed to decode project with expanded admins: %v", err)
	}

	if p.Admins == nil {
		t.Fatal("expected admins object to be populated")
	}
	if len(p.Admins.Identities) != 1 {
		t.Fatalf("expected 1 admin identity, got %d", len(p.Admins.Identities))
	}
	got := p.Admins.Identities[0]
	if got.Username != "admin@example.org" || got.IdentityType != "login" || got.Organization != "Example Org" {
		t.Errorf("admin identity fields not decoded correctly: %+v", got)
	}
	if got.ID != "ae341a98-d274-11e5-b888-dbae3a8ba545" {
		t.Errorf("admin identity ID = %q, want the identity UUID", got.ID)
	}

	// A populated group object must also decode (docs show groups empty, but the
	// field is an object array, not a string array).
	groupBody := `{"identities": [], "groups": [{"id": "g-123", "name": "Admins"}]}`
	var admins auth.ProjectAdmins
	if err := json.Unmarshal([]byte(groupBody), &admins); err != nil {
		t.Fatalf("failed to decode admins with group object: %v", err)
	}
	if len(admins.Groups) != 1 || admins.Groups[0].Name != "Admins" {
		t.Errorf("admin group not decoded correctly: %+v", admins.Groups)
	}
}

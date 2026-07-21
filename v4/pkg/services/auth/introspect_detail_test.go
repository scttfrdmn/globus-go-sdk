// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/testhelpers"
)

// TestIntrospectIdentitySetDetail verifies the fields added for issue #51:
// organization on the introspection result, and a detailed linked-identity set
// (with identity_type) via include=identity_set_detail.
func TestIntrospectIdentitySetDetail(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.PostForm.Get("include"); got != "identity_set_detail" {
			t.Errorf("include = %q, want identity_set_detail", got)
		}
		testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"active":       true,
			"sub":          "primary-uuid",
			"username":     "user@example.org",
			"organization": "Example Org",
			"identity_set_detail": []map[string]interface{}{
				{"id": "primary-uuid", "username": "user@example.org", "organization": "Example Org", "identity_type": "login"},
				{"id": "linked-uuid", "username": "user@orcid", "identity_type": "link"},
			},
		})
	})
	client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	res, err := client.IntrospectToken(context.Background(), "tok", &auth.IntrospectOptions{Include: "identity_set_detail"})
	if err != nil {
		t.Fatalf("IntrospectToken: %v", err)
	}
	if res.Organization != "Example Org" {
		t.Errorf("Organization = %q, want Example Org", res.Organization)
	}
	if len(res.IdentitySetDetail) != 2 {
		t.Fatalf("IdentitySetDetail len = %d, want 2", len(res.IdentitySetDetail))
	}
	if res.IdentitySetDetail[0].IdentityType != "login" || res.IdentitySetDetail[1].IdentityType != "link" {
		t.Errorf("identity_type not decoded: %+v", res.IdentitySetDetail)
	}
	if res.IdentitySetDetail[1].Username != "user@orcid" {
		t.Errorf("linked username = %q", res.IdentitySetDetail[1].Username)
	}
}

// TestIntrospectSessionInfo verifies that include=session_info decodes the
// session detail (session_id + per-identity authentications), the basis for a
// `session show` CLI command.
func TestIntrospectSessionInfo(t *testing.T) {
	server := testhelpers.NewMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.PostForm.Get("include"); got != "session_info" {
			t.Errorf("include = %q, want session_info", got)
		}
		testhelpers.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"active": true,
			"sub":    "primary-uuid",
			"session_info": map[string]interface{}{
				"session_id": "bf133a70-5e59-404a-b3ab-83728e030372",
				"authentications": map[string]interface{}{
					"id-a": map[string]interface{}{
						"auth_time": 1535238478,
						"idp":       "idp-uuid",
						"acr":       "https://refeds.org/profile/mfa",
						"amr":       []string{"mfa"},
					},
					"id-b": map[string]interface{}{
						"auth_time": 1535238000,
					},
				},
			},
		})
	})
	client, err := auth.NewClient(context.Background(), testhelpers.NewTestConfig(server.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	res, err := client.IntrospectToken(context.Background(), "tok", &auth.IntrospectOptions{Include: "session_info"})
	if err != nil {
		t.Fatalf("IntrospectToken: %v", err)
	}
	if res.SessionInfo == nil {
		t.Fatal("SessionInfo is nil, want populated")
	}
	if res.SessionInfo.SessionID != "bf133a70-5e59-404a-b3ab-83728e030372" {
		t.Errorf("SessionID = %q", res.SessionInfo.SessionID)
	}
	if len(res.SessionInfo.Authentications) != 2 {
		t.Fatalf("Authentications len = %d, want 2", len(res.SessionInfo.Authentications))
	}
	a := res.SessionInfo.Authentications["id-a"]
	if a.AuthTime != 1535238478 {
		t.Errorf("id-a auth_time = %d", a.AuthTime)
	}
	if a.ACR != "https://refeds.org/profile/mfa" || len(a.AMR) != 1 || a.AMR[0] != "mfa" {
		t.Errorf("id-a acr/amr not decoded: %+v", a)
	}
	if res.SessionInfo.Authentications["id-b"].AuthTime != 1535238000 {
		t.Errorf("id-b auth_time = %d", res.SessionInfo.Authentications["id-b"].AuthTime)
	}
}

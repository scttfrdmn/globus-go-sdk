// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package login

import (
	"net/url"
	"strings"
	"testing"
)

// TestBuildAuthURLSessionParams verifies that the session-enforcement
// (step-up auth) parameters on AuthParams are emitted on the authorization URL,
// which is what powers a CLI `session update`.
func TestBuildAuthURLSessionParams(t *testing.T) {
	m := NewCommandLineLoginFlowManager("client-abc", "")

	params := AuthParams{
		Scopes:                      []string{"openid"},
		SessionRequiredIdentities:   []string{"id-1", "id-2"},
		SessionRequiredSingleDomain: []string{"example.org"},
		SessionRequiredPolicies:     []string{"pol-1"},
		SessionRequiredMFA:          true,
		SessionMessage:              "please re-authenticate",
	}

	raw := m.buildAuthURL(params.Scopes, "https://example.org/cb", "", params)
	u, err := parseQuery(raw)
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}

	checks := map[string]string{
		"session_required_identities":    "id-1,id-2",
		"session_required_single_domain": "example.org",
		"session_required_policies":      "pol-1",
		"session_required_mfa":           "true",
		"session_message":                "please re-authenticate",
	}
	for k, want := range checks {
		if got := u.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	// Absent when unset: a bare params object must emit no session_* keys.
	bare := m.buildAuthURL([]string{"openid"}, "https://example.org/cb", "", AuthParams{Scopes: []string{"openid"}})
	if strings.Contains(bare, "session_required") || strings.Contains(bare, "session_message") {
		t.Errorf("unexpected session params in URL with none set: %s", bare)
	}
}

// parseQuery extracts the query values from an authorization URL.
func parseQuery(raw string) (url.Values, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return u.Query(), nil
}

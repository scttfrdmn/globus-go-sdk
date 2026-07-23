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

	raw := m.buildAuthURL(params.Scopes, "https://example.org/cb", "", params, "challenge-abc")
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
		"code_challenge":                 "challenge-abc",
		"code_challenge_method":          "S256",
	}
	for k, want := range checks {
		if got := u.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}

	// Absent when unset: a bare params object (and no PKCE challenge) must emit
	// no session_* or code_challenge keys.
	bare := m.buildAuthURL([]string{"openid"}, "https://example.org/cb", "", AuthParams{Scopes: []string{"openid"}}, "")
	if strings.Contains(bare, "session_required") || strings.Contains(bare, "session_message") {
		t.Errorf("unexpected session params in URL with none set: %s", bare)
	}
	if strings.Contains(bare, "code_challenge") {
		t.Errorf("unexpected code_challenge in URL when none set: %s", bare)
	}
}

// TestPKCEVerifierAndChallenge checks the RFC 7636 S256 derivation: a known
// verifier produces the documented challenge, and generated verifiers are
// unique and URL-safe.
func TestPKCEVerifierAndChallenge(t *testing.T) {
	// RFC 7636 Appendix B worked example.
	const rfcVerifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	const rfcChallenge = "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := pkceChallengeS256(rfcVerifier); got != rfcChallenge {
		t.Errorf("pkceChallengeS256(RFC example) = %q, want %q", got, rfcChallenge)
	}

	v1, err := generatePKCEVerifier()
	if err != nil {
		t.Fatalf("generatePKCEVerifier: %v", err)
	}
	if len(v1) < 43 || len(v1) > 128 {
		t.Errorf("verifier length %d out of RFC range 43-128", len(v1))
	}
	v2, _ := generatePKCEVerifier()
	if v1 == v2 {
		t.Error("consecutive verifiers should differ")
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

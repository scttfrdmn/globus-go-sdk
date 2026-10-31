// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/services/auth"
)

// Upstream v4.9.1 (#1430) fixed authorize URLs to comma-join session_required_*
// lists and to omit them when empty. Pin that the Go builder does the same.
func TestGetAuthorizationURL_SessionRequiredParams(t *testing.T) {
	c, err := auth.NewClient(context.Background(), &core.Config{AccessToken: "unused", Scopes: []string{"openid"}})
	require.NoError(t, err)

	t.Run("lists are comma-joined", func(t *testing.T) {
		raw, err := c.GetAuthorizationURL(&auth.AuthorizationURLOptions{
			ClientID: "cid", RedirectURI: "https://example.org/cb",
			SessionRequiredIdentities:   []string{"id-1", "id-2"},
			SessionRequiredSingleDomain: []string{"uchicago.edu", "example.org"},
			SessionRequiredPolicies:     []string{"pol-1"},
			SessionRequiredMFA:          true,
		})
		require.NoError(t, err)
		u, err := url.Parse(raw)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "/v2/oauth2/authorize", u.Path)
		assert.Equal(t, "id-1,id-2", q.Get("session_required_identities"))
		assert.Equal(t, "uchicago.edu,example.org", q.Get("session_required_single_domain"))
		assert.Equal(t, "pol-1", q.Get("session_required_policies"))
		assert.Equal(t, "true", q.Get("session_required_mfa"))
	})

	t.Run("empty lists are omitted", func(t *testing.T) {
		raw, err := c.GetAuthorizationURL(&auth.AuthorizationURLOptions{
			ClientID: "cid", RedirectURI: "https://example.org/cb",
			SessionRequiredIdentities:   []string{},
			SessionRequiredSingleDomain: []string{},
			SessionRequiredPolicies:     nil,
		})
		require.NoError(t, err)
		u, err := url.Parse(raw)
		require.NoError(t, err)
		for _, k := range []string{"session_required_identities", "session_required_single_domain", "session_required_policies", "session_required_mfa"} {
			_, present := u.Query()[k]
			assert.False(t, present, "%s should be omitted, got %s", k, raw)
		}
	})
}

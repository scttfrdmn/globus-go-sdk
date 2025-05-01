// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"testing"
	"time"
)

func TestTokenSetBasics(t *testing.T) {
	// Create a TokenSet
	tokenSet := &TokenSet{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "test-scope",
		ResourceID:   "test-resource",
	}

	// Test IsExpired
	if tokenSet.IsExpired() {
		t.Error("TokenSet.IsExpired() = true, want false")
	}

	// Test CanRefresh
	if !tokenSet.CanRefresh() {
		t.Error("TokenSet.CanRefresh() = false, want true")
	}

	// Create an expired TokenSet
	expiredTokenSet := &TokenSet{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		Scope:        "test-scope",
		ResourceID:   "test-resource",
	}

	// Test IsExpired
	if !expiredTokenSet.IsExpired() {
		t.Error("expiredTokenSet.IsExpired() = false, want true")
	}

	// Create a TokenSet without refresh token
	noRefreshTokenSet := &TokenSet{
		AccessToken: "test-access-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Scope:       "test-scope",
		ResourceID:  "test-resource",
	}

	// Test CanRefresh
	if noRefreshTokenSet.CanRefresh() {
		t.Error("noRefreshTokenSet.CanRefresh() = true, want false")
	}
}
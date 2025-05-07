// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

// Scope constants for Globus Auth
const (
	// ScopeOpenID is the OpenID Connect scope
	ScopeOpenID = "openid"

	// ScopeProfile is the profile scope
	ScopeProfile = "profile"

	// ScopeEmail is the email scope
	ScopeEmail = "email"

	// ScopeOfflineAccess is the offline access scope (for refresh tokens)
	ScopeOfflineAccess = "offline_access"

	// ScopeAllScopes is a convenience constant for all common scopes
	ScopeAllScopes = "openid profile email offline_access"
)

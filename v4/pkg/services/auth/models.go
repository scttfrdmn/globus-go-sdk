// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "time"

// UserInfo represents user information from the /oauth2/userinfo endpoint
type UserInfo struct {
	Sub               string `json:"sub"`
	Email             string `json:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty"`
	Name              string `json:"name,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Organization      string `json:"organization,omitempty"`
}

// TokenIntrospection represents the result of token introspection
type TokenIntrospection struct {
	Active       bool     `json:"active"`
	Scope        string   `json:"scope,omitempty"`
	ClientID     string   `json:"client_id,omitempty"`
	Username     string   `json:"username,omitempty"`
	TokenType    string   `json:"token_type,omitempty"`
	Exp          int64    `json:"exp,omitempty"`
	Iat          int64    `json:"iat,omitempty"`
	Nbf          int64    `json:"nbf,omitempty"`
	Sub          string   `json:"sub,omitempty"`
	Aud          []string `json:"aud,omitempty"`
	Iss          string   `json:"iss,omitempty"`
	Jti          string   `json:"jti,omitempty"`
	Name         string   `json:"name,omitempty"`
	Email        string   `json:"email,omitempty"`
	Organization string   `json:"organization,omitempty"`
	// IdentitySet holds the linked identity UUIDs, populated when
	// include=identity_set.
	IdentitySet []string `json:"identity_set,omitempty"`
	// IdentitySetDetail holds the full linked-identity records, populated when
	// include=identity_set_detail — avoiding a second GetIdentities round trip.
	IdentitySetDetail []Identity `json:"identity_set_detail,omitempty"`
	// SessionInfo describes the Globus Auth session the token derives from,
	// populated when include=session_info. Used to answer "which identities have
	// authenticated in this session and when" (the basis for `session show`).
	SessionInfo *SessionInfo `json:"session_info,omitempty"`
}

// SessionInfo is the session detail returned by token introspection when
// include=session_info is requested. See the Globus Auth Sessions guide.
type SessionInfo struct {
	// SessionID is the session's UUID.
	SessionID string `json:"session_id,omitempty"`
	// Authentications maps an identity ID (UUID) to when and how that identity
	// last authenticated in this session.
	Authentications map[string]SessionAuthentication `json:"authentications,omitempty"`
}

// SessionAuthentication records a single identity's authentication within a
// session.
type SessionAuthentication struct {
	// AuthTime is the Unix timestamp of the authentication.
	AuthTime int64 `json:"auth_time"`
	// IDP is the UUID of the identity provider used.
	IDP string `json:"idp,omitempty"`
	// ACR is the Authentication Context Class Reference (e.g. an MFA profile
	// URL), when present.
	ACR string `json:"acr,omitempty"`
	// AMR lists the Authentication Methods References (e.g. ["mfa"]).
	AMR []string `json:"amr,omitempty"`
}

// TokenResponse represents an OAuth2 token response. When a single token request
// yields tokens for multiple resource servers, the primary token's fields are at
// the top level and the remainder appear in OtherTokens.
type TokenResponse struct {
	AccessToken    string          `json:"access_token"`
	RefreshToken   string          `json:"refresh_token,omitempty"`
	ExpiresIn      int             `json:"expires_in"`
	TokenType      string          `json:"token_type"`
	Scope          string          `json:"scope,omitempty"`
	ResourceServer string          `json:"resource_server,omitempty"`
	IDToken        string          `json:"id_token,omitempty"`
	State          string          `json:"state,omitempty"`
	OtherTokens    []TokenResponse `json:"other_tokens,omitempty"`
}

// ByResourceServer returns a map of resource server to token, including the
// primary token and every entry in OtherTokens. It mirrors the Python SDK's
// by_resource_server accessor.
func (t *TokenResponse) ByResourceServer() map[string]TokenResponse {
	out := make(map[string]TokenResponse, 1+len(t.OtherTokens))
	if t.ResourceServer != "" {
		primary := *t
		primary.OtherTokens = nil
		out[t.ResourceServer] = primary
	}
	for _, o := range t.OtherTokens {
		if o.ResourceServer != "" {
			out[o.ResourceServer] = o
		}
	}
	return out
}

// ExpiresAt returns the time when the access token expires
func (t *TokenResponse) ExpiresAt() time.Time {
	return time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
}

// IsExpired returns true if the access token has expired
func (t *TokenResponse) IsExpired() bool {
	return time.Now().After(t.ExpiresAt())
}

// Project represents a Globus Auth project
type Project struct {
	ID            string         `json:"id"`
	DisplayName   string         `json:"display_name"`
	ContactEmail  string         `json:"contact_email"`
	AdminIDs      []string       `json:"admin_ids"`
	AdminGroupIDs []string       `json:"admin_group_ids,omitempty"`
	Admins        *ProjectAdmins `json:"admins,omitempty"`
	ProjectName   string         `json:"project_name,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	LastUpdated   time.Time      `json:"last_updated"`
}

// ProjectAdmins is the expanded admins object on a project. The Globus Auth
// Projects API returns identities and groups as arrays of objects (not ID
// strings); the string ID arrays live on Project.AdminIDs/AdminGroupIDs.
type ProjectAdmins struct {
	Identities []ProjectAdminIdentity `json:"identities"`
	Groups     []ProjectAdminGroup    `json:"groups"`
}

// ProjectAdminIdentity is an expanded admin identity in ProjectAdmins.Identities.
type ProjectAdminIdentity struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	IdentityProvider string `json:"identity_provider"`
	IdentityType     string `json:"identity_type"`
	Organization     string `json:"organization"`
	Status           string `json:"status"`
}

// ProjectAdminGroup is an expanded admin group in ProjectAdmins.Groups.
type ProjectAdminGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProjectCreate is the create-project body (nested under a "project" key on the
// wire). Upstream 4.8.1 accepts only these fields; at least one of AdminIDs or
// AdminGroupIDs is required.
type ProjectCreate struct {
	DisplayName   string   `json:"display_name"`
	ContactEmail  string   `json:"contact_email,omitempty"`
	AdminIDs      []string `json:"admin_ids,omitempty"`
	AdminGroupIDs []string `json:"admin_group_ids,omitempty"`
}

// ProjectUpdate is the update-project body (nested under "project"). Only set
// fields are sent.
type ProjectUpdate struct {
	DisplayName   string   `json:"display_name,omitempty"`
	ContactEmail  string   `json:"contact_email,omitempty"`
	AdminIDs      []string `json:"admin_ids,omitempty"`
	AdminGroupIDs []string `json:"admin_group_ids,omitempty"`
}

// AuthorizationURLOptions controls how the authorization URL is constructed.
type AuthorizationURLOptions struct {
	ClientID    string
	RedirectURI string
	Scopes      []string
	State       string
	AccessType  string
	Prompt      string

	// Session enforcement parameters (Globus Auth "high assurance" / step-up auth).
	SessionRequiredIdentities   []string // comma-joined into session_required_identities
	SessionRequiredSingleDomain []string // comma-joined into session_required_single_domain
	SessionRequiredPolicies     []string // comma-joined into session_required_policies
	SessionRequiredMFA          bool     // session_required_mfa
	SessionMessage              string   // session_message
}

// DeviceAuthorizationResponse is returned by StartDeviceAuthorization.
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

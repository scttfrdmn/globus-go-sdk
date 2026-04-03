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
	Active    bool     `json:"active"`
	Scope     string   `json:"scope,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	TokenType string   `json:"token_type,omitempty"`
	Exp       int64    `json:"exp,omitempty"`
	Iat       int64    `json:"iat,omitempty"`
	Nbf       int64    `json:"nbf,omitempty"`
	Sub       string   `json:"sub,omitempty"`
	Aud       []string `json:"aud,omitempty"`
	Iss       string   `json:"iss,omitempty"`
	Jti       string   `json:"jti,omitempty"`
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
	ResourceServer string `json:"resource_server,omitempty"`
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
	ID                string                 `json:"id"`
	DisplayName       string                 `json:"display_name"`
	ContactEmail      string                 `json:"contact_email"`
	AdminIDs          []string               `json:"admin_ids"`
	AdminGroupIDs     []string               `json:"admin_group_ids,omitempty"`
	ProjectName       string                 `json:"project_name,omitempty"`
	PublicContactInfo string                 `json:"public_contact_info,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	LastUpdated       time.Time              `json:"last_updated"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ProjectCreate represents the data needed to create a new project
type ProjectCreate struct {
	DisplayName       string                 `json:"display_name"`
	ContactEmail      string                 `json:"contact_email"`
	AdminIDs          []string               `json:"admin_ids,omitempty"`
	AdminGroupIDs     []string               `json:"admin_group_ids,omitempty"`
	ProjectName       string                 `json:"project_name,omitempty"`
	PublicContactInfo string                 `json:"public_contact_info,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// AuthorizationURLOptions controls how the authorization URL is constructed.
type AuthorizationURLOptions struct {
	ClientID    string
	RedirectURI string
	Scopes      []string
	State       string
	AccessType  string
	Prompt      string
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

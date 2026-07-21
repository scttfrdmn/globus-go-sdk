// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import "time"

// Identity represents a Globus Auth identity (GET /api/identities).
type Identity struct {
	ID               string `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Organization     string `json:"organization"`
	IdentityProvider string `json:"identity_provider"`
	Status           string `json:"status"`
	// IdentityType distinguishes a login identity from a linked identity
	// ("login" vs "link"), returned by GetIdentities and in identity_set_detail.
	IdentityType string `json:"identity_type,omitempty"`
}

// IdentityProvider represents a Globus Auth identity provider.
type IdentityProvider struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ShortName        string   `json:"short_name"`
	Domains          []string `json:"domains"`
	AlternativeNames []string `json:"alternative_names"`
}

// Policy represents a Globus Auth authentication policy.
type Policy struct {
	ID                             string   `json:"id"`
	ProjectID                      string   `json:"project_id"`
	DisplayName                    string   `json:"display_name"`
	Description                    string   `json:"description"`
	HighAssurance                  bool     `json:"high_assurance"`
	AuthenticationAssuranceTimeout int      `json:"authentication_assurance_timeout"`
	RequiredMFA                    bool     `json:"required_mfa"`
	DomainConstraintsInclude       []string `json:"domain_constraints_include"`
	DomainConstraintsExclude       []string `json:"domain_constraints_exclude"`
}

// PolicyCreate is the create-policy body (nested under "policy").
type PolicyCreate struct {
	ProjectID                      string   `json:"project_id"`
	DisplayName                    string   `json:"display_name"`
	Description                    string   `json:"description"`
	HighAssurance                  *bool    `json:"high_assurance,omitempty"`
	AuthenticationAssuranceTimeout *int     `json:"authentication_assurance_timeout,omitempty"`
	RequiredMFA                    *bool    `json:"required_mfa,omitempty"`
	DomainConstraintsInclude       []string `json:"domain_constraints_include,omitempty"`
	DomainConstraintsExclude       []string `json:"domain_constraints_exclude,omitempty"`
}

// PolicyUpdate is the update-policy body (nested under "policy"); only set fields
// are sent.
type PolicyUpdate struct {
	ProjectID                      string   `json:"project_id,omitempty"`
	DisplayName                    string   `json:"display_name,omitempty"`
	Description                    string   `json:"description,omitempty"`
	HighAssurance                  *bool    `json:"high_assurance,omitempty"`
	AuthenticationAssuranceTimeout *int     `json:"authentication_assurance_timeout,omitempty"`
	RequiredMFA                    *bool    `json:"required_mfa,omitempty"`
	DomainConstraintsInclude       []string `json:"domain_constraints_include,omitempty"`
	DomainConstraintsExclude       []string `json:"domain_constraints_exclude,omitempty"`
}

// AuthClientInfo represents a registered Globus Auth client (the /api/clients
// resource). Named to avoid collision with the auth.Client service type.
type AuthClientInfo struct {
	ID                            string      `json:"id"`
	Name                          string      `json:"name"`
	Project                       string      `json:"project"`
	ClientType                    string      `json:"client_type"`
	Visibility                    string      `json:"visibility"`
	RedirectURIs                  []string    `json:"redirect_uris"`
	Scopes                        []string    `json:"scopes"`
	GrantTypes                    []string    `json:"grant_types"`
	FQDNs                         []string    `json:"fqdns"`
	Links                         ClientLinks `json:"links"`
	ParentClient                  *string     `json:"parent_client"`
	RequiredIDP                   *string     `json:"required_idp"`
	PreselectIDP                  *string     `json:"preselect_idp"`
	PublicClient                  bool        `json:"public_client"`
	PromptForNamedGrant           bool        `json:"prompt_for_named_grant"`
	UserinfoFromEffectiveIdentity bool        `json:"userinfo_from_effective_identity"`
}

// ClientLinks holds the terms/privacy links on a client.
type ClientLinks struct {
	PrivacyPolicy      *string `json:"privacy_policy"`
	TermsAndConditions *string `json:"terms_and_conditions"`
}

// ClientLinksInput is the writable links object for client create/update. When
// present, both fields must be supplied together (both set or both null).
type ClientLinksInput struct {
	TermsAndConditions string `json:"terms_and_conditions"`
	PrivacyPolicy      string `json:"privacy_policy"`
}

// ClientCreate is the create-client body (nested under "client"). Exactly one of
// PublicClient or ClientType must be set.
type ClientCreate struct {
	Name         string            `json:"name"`
	Project      string            `json:"project"`
	Visibility   string            `json:"visibility,omitempty"`
	RedirectURIs []string          `json:"redirect_uris,omitempty"`
	RequiredIDP  string            `json:"required_idp,omitempty"`
	PreselectIDP string            `json:"preselect_idp,omitempty"`
	PublicClient *bool             `json:"public_client,omitempty"`
	ClientType   string            `json:"client_type,omitempty"`
	Links        *ClientLinksInput `json:"links,omitempty"`
}

// ClientUpdate is the update-client body (nested under "client"); only set fields
// are sent. project / public_client / client_type are not updatable.
type ClientUpdate struct {
	Name         string            `json:"name,omitempty"`
	Visibility   string            `json:"visibility,omitempty"`
	RequiredIDP  string            `json:"required_idp,omitempty"`
	PreselectIDP string            `json:"preselect_idp,omitempty"`
	RedirectURIs []string          `json:"redirect_uris,omitempty"`
	Links        *ClientLinksInput `json:"links,omitempty"`
}

// ChildClientCreate is the create-child-client body (nested under "client"). Like
// ClientCreate but without a Project field (a child inherits its parent's project).
type ChildClientCreate struct {
	Name         string            `json:"name"`
	Visibility   string            `json:"visibility,omitempty"`
	RequiredIDP  string            `json:"required_idp,omitempty"`
	PreselectIDP string            `json:"preselect_idp,omitempty"`
	PublicClient *bool             `json:"public_client,omitempty"`
	ClientType   string            `json:"client_type,omitempty"`
	RedirectURIs []string          `json:"redirect_uris,omitempty"`
	Links        *ClientLinksInput `json:"links,omitempty"`
}

// Credential represents a client credential. Secret is populated only on create.
type Credential struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Client  string    `json:"client"`
	Created time.Time `json:"created"`
	Secret  *string   `json:"secret"`
}

// DependentScopeSpec describes a dependency of a scope.
type DependentScopeSpec struct {
	Scope                string `json:"scope"`
	Optional             bool   `json:"optional"`
	RequiresRefreshToken bool   `json:"requires_refresh_token"`
}

// Scope represents a Globus Auth scope.
type Scope struct {
	ID                 string               `json:"id"`
	ScopeString        string               `json:"scope_string"`
	AllowsRefreshToken bool                 `json:"allows_refresh_token"`
	Advertised         bool                 `json:"advertised"`
	RequiredDomains    []string             `json:"required_domains"`
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	Client             string               `json:"client"`
	DependentScopes    []DependentScopeSpec `json:"dependent_scopes"`
}

// ScopeCreate is the create-scope body (nested under "scope").
type ScopeCreate struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	ScopeSuffix        string               `json:"scope_suffix"`
	Advertised         *bool                `json:"advertised,omitempty"`
	AllowsRefreshToken *bool                `json:"allows_refresh_token,omitempty"`
	RequiredDomains    []string             `json:"required_domains,omitempty"`
	DependentScopes    []DependentScopeSpec `json:"dependent_scopes,omitempty"`
}

// ScopeUpdate is the update-scope body (nested under "scope"); only set fields
// are sent.
type ScopeUpdate struct {
	Name               string               `json:"name,omitempty"`
	Description        string               `json:"description,omitempty"`
	ScopeSuffix        string               `json:"scope_suffix,omitempty"`
	Advertised         *bool                `json:"advertised,omitempty"`
	AllowsRefreshToken *bool                `json:"allows_refresh_token,omitempty"`
	RequiredDomains    []string             `json:"required_domains,omitempty"`
	DependentScopes    []DependentScopeSpec `json:"dependent_scopes,omitempty"`
}

// Consent represents a consent granted to a client by an identity. ID and the
// dependency path elements are integers, not UUIDs.
type Consent struct {
	ID                  int       `json:"id"`
	Client              string    `json:"client"`
	Scope               string    `json:"scope"`
	ScopeName           string    `json:"scope_name"`
	EffectiveIdentity   string    `json:"effective_identity"`
	DependencyPath      []int     `json:"dependency_path"`
	Created             time.Time `json:"created"`
	Updated             time.Time `json:"updated"`
	LastUsed            time.Time `json:"last_used"`
	Status              string    `json:"status"`
	AllowsRefresh       bool      `json:"allows_refresh"`
	AutoApproved        bool      `json:"auto_approved"`
	AtomicallyRevocable bool      `json:"atomically_revocable"`
}

// DependentTokenInfo is one element of the dependent-token grant response array.
type DependentTokenInfo struct {
	AccessToken    string `json:"access_token"`
	Scope          string `json:"scope"`
	RefreshToken   string `json:"refresh_token,omitempty"`
	TokenType      string `json:"token_type,omitempty"`
	ExpiresIn      int    `json:"expires_in"`
	ResourceServer string `json:"resource_server"`
}

// --- Option structs ---

// IntrospectOptions carries optional parameters for IntrospectToken.
type IntrospectOptions struct {
	// Include is the comma-separated introspection include set, e.g.
	// "identity_set" (UUIDs, into IdentitySet) or "identity_set_detail" (full
	// records, into IdentitySetDetail).
	Include string
}

// GetIdentitiesOptions controls GetIdentities. Usernames and IDs are mutually
// exclusive; Provision is only sent when Usernames is set.
type GetIdentitiesOptions struct {
	Usernames []string
	IDs       []string
	Provision bool
}

// GetIdentityProvidersOptions controls GetIdentityProviders. Domains and IDs are
// mutually exclusive.
type GetIdentityProvidersOptions struct {
	Domains []string
	IDs     []string
}

// GetScopesOptions controls GetScopes. ScopeStrings and IDs are mutually exclusive.
type GetScopesOptions struct {
	ScopeStrings []string
	IDs          []string
}

// GetClientOptions selects a client by ID or FQDN (exactly one).
type GetClientOptions struct {
	ClientID string
	FQDN     string
}

// DependentTokensOptions controls GetDependentTokens.
type DependentTokensOptions struct {
	RefreshTokens    bool              // sets access_type=offline
	Scopes           []string          // space-joined into scope
	AdditionalParams map[string]string // extra form fields
}

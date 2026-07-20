// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// doJSON performs a JSON request against a path relative to the client base URL
// (which ends in /v2/), decoding the response into result when non-nil.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body, result interface{}) error {
	u := c.Client.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

// doJSONAbsolute is like doJSON but against a fully-qualified URL (used for the
// host-root OIDC/JWK endpoints outside the /v2/ base).
func (c *Client) doJSONAbsolute(ctx context.Context, method, fullURL string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

// hostRoot returns the Auth host root (scheme+host, no /v2/) for the OIDC
// discovery and JWKS endpoints.
func (c *Client) hostRoot() string {
	base := c.Client.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	if u, err := url.Parse(base); err == nil {
		u.Path, u.RawQuery = "", ""
		return strings.TrimRight(u.String(), "/")
	}
	return "https://auth.globus.org"
}

// --- OIDC ---

// GetOpenIDConfiguration fetches the OIDC discovery document
// (GET /.well-known/openid-configuration at the host root).
func (c *Client) GetOpenIDConfiguration(ctx context.Context) (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := c.doJSONAbsolute(ctx, http.MethodGet, c.hostRoot()+"/.well-known/openid-configuration", &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// GetJWK fetches the JWKS from the jwks_uri advertised in the OIDC config.
func (c *Client) GetJWK(ctx context.Context) (map[string]interface{}, error) {
	config, err := c.GetOpenIDConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	jwksURI, ok := config["jwks_uri"].(string)
	if !ok || jwksURI == "" {
		return nil, fmt.Errorf("OIDC configuration did not contain a jwks_uri")
	}
	var jwks map[string]interface{}
	if err := c.doJSONAbsolute(ctx, http.MethodGet, jwksURI, &jwks); err != nil {
		return nil, err
	}
	return jwks, nil
}

// Userinfo returns the OIDC userinfo document (GET /v2/oauth2/userinfo).
func (c *Client) Userinfo(ctx context.Context) (map[string]interface{}, error) {
	var info map[string]interface{}
	if err := c.doJSON(ctx, http.MethodGet, "oauth2/userinfo", nil, nil, &info); err != nil {
		return nil, err
	}
	return info, nil
}

// --- Identities ---

// GetIdentitiesOptions controls GetIdentities. Usernames and IDs are mutually
// exclusive; Provision is only sent with Usernames.
type GetIdentitiesOptions struct {
	Usernames []string
	IDs       []string
	Provision bool
}

// GetIdentities looks up identities (GET /v2/api/identities).
func (c *Client) GetIdentities(ctx context.Context, opts *GetIdentitiesOptions) ([]Identity, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.Usernames) > 0 && len(opts.IDs) > 0 {
			return nil, fmt.Errorf("usernames and ids are mutually exclusive")
		}
		if len(opts.Usernames) > 0 {
			query.Set("usernames", strings.Join(opts.Usernames, ","))
			if opts.Provision {
				query.Set("provision", "true")
			} else {
				query.Set("provision", "false")
			}
		}
		if len(opts.IDs) > 0 {
			query.Set("ids", strings.Join(opts.IDs, ","))
		}
	}
	var env struct {
		Identities []Identity `json:"identities"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/identities", query, nil, &env); err != nil {
		return nil, err
	}
	return env.Identities, nil
}

// IdentityProvider represents a Globus Auth identity provider.
type IdentityProvider struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ShortName        string   `json:"short_name"`
	Domains          []string `json:"domains"`
	AlternativeNames []string `json:"alternative_names"`
}

// GetIdentityProvidersOptions controls GetIdentityProviders (Domains and IDs are
// mutually exclusive).
type GetIdentityProvidersOptions struct {
	Domains []string
	IDs     []string
}

// GetIdentityProviders looks up identity providers (GET /v2/api/identity_providers).
func (c *Client) GetIdentityProviders(ctx context.Context, opts *GetIdentityProvidersOptions) ([]IdentityProvider, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.Domains) > 0 && len(opts.IDs) > 0 {
			return nil, fmt.Errorf("domains and ids are mutually exclusive")
		}
		if len(opts.Domains) > 0 {
			query.Set("domains", strings.Join(opts.Domains, ","))
		}
		if len(opts.IDs) > 0 {
			query.Set("ids", strings.Join(opts.IDs, ","))
		}
	}
	var env struct {
		IdentityProviders []IdentityProvider `json:"identity_providers"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/identity_providers", query, nil, &env); err != nil {
		return nil, err
	}
	return env.IdentityProviders, nil
}

// --- Projects ---

// Project represents a Globus Auth project.
type Project struct {
	ID            string   `json:"id"`
	DisplayName   string   `json:"display_name"`
	ContactEmail  string   `json:"contact_email"`
	AdminIDs      []string `json:"admin_ids"`
	AdminGroupIDs []string `json:"admin_group_ids,omitempty"`
}

// ProjectCreate is the create/update body for a project (nested under "project").
type ProjectCreate struct {
	DisplayName   string   `json:"display_name,omitempty"`
	ContactEmail  string   `json:"contact_email,omitempty"`
	AdminIDs      []string `json:"admin_ids,omitempty"`
	AdminGroupIDs []string `json:"admin_group_ids,omitempty"`
}

// GetProjects lists the caller's projects (GET /v2/api/projects).
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var env struct {
		Projects []Project `json:"projects"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/projects", nil, nil, &env); err != nil {
		return nil, err
	}
	return env.Projects, nil
}

// GetProject retrieves a project (GET /v2/api/projects/{id}).
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	var env struct {
		Project Project `json:"project"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/projects/"+projectID, nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Project, nil
}

// CreateProject creates a project (POST /v2/api/projects).
func (c *Client) CreateProject(ctx context.Context, project *ProjectCreate) (*Project, error) {
	if project == nil {
		return nil, fmt.Errorf("project is required")
	}
	var env struct {
		Project Project `json:"project"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "api/projects", nil, map[string]interface{}{"project": project}, &env); err != nil {
		return nil, err
	}
	return &env.Project, nil
}

// UpdateProject updates a project (PUT /v2/api/projects/{id}).
func (c *Client) UpdateProject(ctx context.Context, projectID string, update *ProjectCreate) (*Project, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update document is required")
	}
	var env struct {
		Project Project `json:"project"`
	}
	if err := c.doJSON(ctx, http.MethodPut, "api/projects/"+projectID, nil, map[string]interface{}{"project": update}, &env); err != nil {
		return nil, err
	}
	return &env.Project, nil
}

// DeleteProject deletes a project (DELETE /v2/api/projects/{id}).
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "api/projects/"+projectID, nil, nil, nil)
}

// --- Policies ---

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

// PolicyCreate is the create/update body for a policy (nested under "policy").
type PolicyCreate struct {
	ProjectID                      string   `json:"project_id,omitempty"`
	DisplayName                    string   `json:"display_name,omitempty"`
	Description                    string   `json:"description,omitempty"`
	HighAssurance                  *bool    `json:"high_assurance,omitempty"`
	AuthenticationAssuranceTimeout *int     `json:"authentication_assurance_timeout,omitempty"`
	RequiredMFA                    *bool    `json:"required_mfa,omitempty"`
	DomainConstraintsInclude       []string `json:"domain_constraints_include,omitempty"`
	DomainConstraintsExclude       []string `json:"domain_constraints_exclude,omitempty"`
}

// GetPolicies lists policies (GET /v2/api/policies).
func (c *Client) GetPolicies(ctx context.Context) ([]Policy, error) {
	var env struct {
		Policies []Policy `json:"policies"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/policies", nil, nil, &env); err != nil {
		return nil, err
	}
	return env.Policies, nil
}

// GetPolicy retrieves a policy (GET /v2/api/policies/{id}).
func (c *Client) GetPolicy(ctx context.Context, policyID string) (*Policy, error) {
	if policyID == "" {
		return nil, fmt.Errorf("policy ID is required")
	}
	var env struct {
		Policy Policy `json:"policy"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/policies/"+policyID, nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Policy, nil
}

// CreatePolicy creates a policy (POST /v2/api/policies).
func (c *Client) CreatePolicy(ctx context.Context, policy *PolicyCreate) (*Policy, error) {
	if policy == nil {
		return nil, fmt.Errorf("policy is required")
	}
	var env struct {
		Policy Policy `json:"policy"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "api/policies", nil, map[string]interface{}{"policy": policy}, &env); err != nil {
		return nil, err
	}
	return &env.Policy, nil
}

// UpdatePolicy updates a policy (PUT /v2/api/policies/{id}).
func (c *Client) UpdatePolicy(ctx context.Context, policyID string, update *PolicyCreate) (*Policy, error) {
	if policyID == "" {
		return nil, fmt.Errorf("policy ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update document is required")
	}
	var env struct {
		Policy Policy `json:"policy"`
	}
	if err := c.doJSON(ctx, http.MethodPut, "api/policies/"+policyID, nil, map[string]interface{}{"policy": update}, &env); err != nil {
		return nil, err
	}
	return &env.Policy, nil
}

// DeletePolicy deletes a policy (DELETE /v2/api/policies/{id}).
func (c *Client) DeletePolicy(ctx context.Context, policyID string) error {
	if policyID == "" {
		return fmt.Errorf("policy ID is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "api/policies/"+policyID, nil, nil, nil)
}

// --- Clients ---

// ClientInfo represents a registered Globus Auth client.
type ClientInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Project      string   `json:"project"`
	ClientType   string   `json:"client_type"`
	Visibility   string   `json:"visibility"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	FQDNs        []string `json:"fqdns"`
	PublicClient bool     `json:"public_client"`
	ParentClient *string  `json:"parent_client"`
}

// ClientCreate is the create body for a client (nested under "client").
type ClientCreate struct {
	Name         string   `json:"name"`
	Project      string   `json:"project,omitempty"`
	Visibility   string   `json:"visibility,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	RequiredIDP  string   `json:"required_idp,omitempty"`
	PreselectIDP string   `json:"preselect_idp,omitempty"`
	PublicClient *bool    `json:"public_client,omitempty"`
	ClientType   string   `json:"client_type,omitempty"`
	TemplateID   string   `json:"template_id,omitempty"`
}

// GetClientOptions selects a client by ID or FQDN (exactly one).
type GetClientOptions struct {
	ClientID string
	FQDN     string
}

// GetClient retrieves a client by ID or FQDN (GET /v2/api/clients/{id} or ?fqdn=).
func (c *Client) GetClient(ctx context.Context, opts *GetClientOptions) (*ClientInfo, error) {
	if opts == nil || (opts.ClientID == "" && opts.FQDN == "") {
		return nil, fmt.Errorf("exactly one of ClientID or FQDN is required")
	}
	path := "api/clients"
	var query url.Values
	if opts.ClientID != "" {
		path = "api/clients/" + opts.ClientID
	} else {
		query = url.Values{}
		query.Set("fqdn", opts.FQDN)
	}
	var env struct {
		Client ClientInfo `json:"client"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path, query, nil, &env); err != nil {
		return nil, err
	}
	return &env.Client, nil
}

// GetClients lists registered clients (GET /v2/api/clients).
func (c *Client) GetClients(ctx context.Context) ([]ClientInfo, error) {
	var env struct {
		Clients []ClientInfo `json:"clients"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/clients", nil, nil, &env); err != nil {
		return nil, err
	}
	return env.Clients, nil
}

// CreateClient creates a client (POST /v2/api/clients).
func (c *Client) CreateClient(ctx context.Context, req *ClientCreate) (*ClientInfo, error) {
	if req == nil {
		return nil, fmt.Errorf("client create document is required")
	}
	var env struct {
		Client ClientInfo `json:"client"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "api/clients", nil, map[string]interface{}{"client": req}, &env); err != nil {
		return nil, err
	}
	return &env.Client, nil
}

// CreateChildClient creates a child client under the calling confidential client
// (POST /v2/api/clients).
func (c *Client) CreateChildClient(ctx context.Context, req *ClientCreate) (*ClientInfo, error) {
	return c.CreateClient(ctx, req)
}

// CreateNativeAppInstance creates a native app instance from a template
// (POST /v2/api/clients).
func (c *Client) CreateNativeAppInstance(ctx context.Context, templateID, name string) (*ClientInfo, error) {
	if templateID == "" {
		return nil, fmt.Errorf("template ID is required")
	}
	return c.CreateClient(ctx, &ClientCreate{Name: name, TemplateID: templateID})
}

// UpdateClient updates a client (PUT /v2/api/clients/{id}).
func (c *Client) UpdateClient(ctx context.Context, clientID string, update *ClientCreate) (*ClientInfo, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update document is required")
	}
	var env struct {
		Client ClientInfo `json:"client"`
	}
	if err := c.doJSON(ctx, http.MethodPut, "api/clients/"+clientID, nil, map[string]interface{}{"client": update}, &env); err != nil {
		return nil, err
	}
	return &env.Client, nil
}

// DeleteClient deletes a client (DELETE /v2/api/clients/{id}).
func (c *Client) DeleteClient(ctx context.Context, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client ID is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "api/clients/"+clientID, nil, nil, nil)
}

// --- Client credentials ---

// ClientCredential represents a client credential (secret populated on create).
type ClientCredential struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Client  string  `json:"client"`
	Created string  `json:"created"`
	Secret  *string `json:"secret"`
}

// GetClientCredentials lists a client's credentials
// (GET /v2/api/clients/{id}/credentials).
func (c *Client) GetClientCredentials(ctx context.Context, clientID string) ([]ClientCredential, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	var env struct {
		Credentials []ClientCredential `json:"credentials"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/clients/"+clientID+"/credentials", nil, nil, &env); err != nil {
		return nil, err
	}
	return env.Credentials, nil
}

// CreateClientCredential creates a client credential
// (POST /v2/api/clients/{id}/credentials).
func (c *Client) CreateClientCredential(ctx context.Context, clientID, name string) (*ClientCredential, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	body := map[string]interface{}{"credential": map[string]string{"name": name}}
	var env struct {
		Credential ClientCredential `json:"credential"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "api/clients/"+clientID+"/credentials", nil, body, &env); err != nil {
		return nil, err
	}
	return &env.Credential, nil
}

// DeleteClientCredential deletes a client credential
// (DELETE /v2/api/clients/{id}/credentials/{credential_id}).
func (c *Client) DeleteClientCredential(ctx context.Context, clientID, credentialID string) error {
	if clientID == "" {
		return fmt.Errorf("client ID is required")
	}
	if credentialID == "" {
		return fmt.Errorf("credential ID is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "api/clients/"+clientID+"/credentials/"+credentialID, nil, nil, nil)
}

// --- Scopes ---

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
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	Advertised         bool                 `json:"advertised"`
	AllowsRefreshToken bool                 `json:"allows_refresh_token"`
	RequiredDomains    []string             `json:"required_domains"`
	DependentScopes    []DependentScopeSpec `json:"dependent_scopes"`
}

// ScopeCreate is the create/update body for a scope (nested under "scope").
type ScopeCreate struct {
	Name               string               `json:"name,omitempty"`
	Description        string               `json:"description,omitempty"`
	ScopeSuffix        string               `json:"scope_suffix,omitempty"`
	Advertised         *bool                `json:"advertised,omitempty"`
	AllowsRefreshToken *bool                `json:"allows_refresh_token,omitempty"`
	RequiredDomains    []string             `json:"required_domains,omitempty"`
	DependentScopes    []DependentScopeSpec `json:"dependent_scopes,omitempty"`
}

// GetScopesOptions controls GetScopes (ScopeStrings and IDs are mutually exclusive).
type GetScopesOptions struct {
	ScopeStrings []string
	IDs          []string
}

// GetScopes lists scopes (GET /v2/api/scopes).
func (c *Client) GetScopes(ctx context.Context, opts *GetScopesOptions) ([]Scope, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.ScopeStrings) > 0 {
			query.Set("scope_strings", strings.Join(opts.ScopeStrings, ","))
		}
		if len(opts.IDs) > 0 {
			query.Set("ids", strings.Join(opts.IDs, ","))
		}
	}
	var env struct {
		Scopes []Scope `json:"scopes"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/scopes", query, nil, &env); err != nil {
		return nil, err
	}
	return env.Scopes, nil
}

// GetScope retrieves a scope (GET /v2/api/scopes/{id}).
func (c *Client) GetScope(ctx context.Context, scopeID string) (*Scope, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID is required")
	}
	var env struct {
		Scope Scope `json:"scope"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/scopes/"+scopeID, nil, nil, &env); err != nil {
		return nil, err
	}
	return &env.Scope, nil
}

// CreateScope creates a scope on a client (POST /v2/api/clients/{id}/scopes).
func (c *Client) CreateScope(ctx context.Context, clientID string, req *ScopeCreate) (*Scope, error) {
	if clientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("scope create document is required")
	}
	var env struct {
		Scope Scope `json:"scope"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "api/clients/"+clientID+"/scopes", nil, map[string]interface{}{"scope": req}, &env); err != nil {
		return nil, err
	}
	return &env.Scope, nil
}

// UpdateScope updates a scope (PUT /v2/api/scopes/{id}).
func (c *Client) UpdateScope(ctx context.Context, scopeID string, update *ScopeCreate) (*Scope, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID is required")
	}
	if update == nil {
		return nil, fmt.Errorf("update document is required")
	}
	var env struct {
		Scope Scope `json:"scope"`
	}
	if err := c.doJSON(ctx, http.MethodPut, "api/scopes/"+scopeID, nil, map[string]interface{}{"scope": update}, &env); err != nil {
		return nil, err
	}
	return &env.Scope, nil
}

// DeleteScope deletes a scope (DELETE /v2/api/scopes/{id}).
func (c *Client) DeleteScope(ctx context.Context, scopeID string) error {
	if scopeID == "" {
		return fmt.Errorf("scope ID is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "api/scopes/"+scopeID, nil, nil, nil)
}

// --- Consents ---

// Consent represents a consent granted to a client by an identity.
type Consent struct {
	ID                int    `json:"id"`
	Client            string `json:"client"`
	Scope             string `json:"scope"`
	ScopeName         string `json:"scope_name"`
	EffectiveIdentity string `json:"effective_identity"`
	DependencyPath    []int  `json:"dependency_path"`
	Status            string `json:"status"`
	AllowsRefresh     bool   `json:"allows_refresh"`
}

// GetConsents lists the consents for an identity
// (GET /v2/api/identities/{id}/consents).
func (c *Client) GetConsents(ctx context.Context, identityID string, all bool) ([]Consent, error) {
	if identityID == "" {
		return nil, fmt.Errorf("identity ID is required")
	}
	query := url.Values{}
	if all {
		query.Set("all", "true")
	}
	var env struct {
		Consents []Consent `json:"consents"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/identities/"+identityID+"/consents", query, nil, &env); err != nil {
		return nil, err
	}
	return env.Consents, nil
}

// --- OAuth2 grants ---

// DependentTokenInfo is one element of the dependent-token grant response array.
type DependentTokenInfo struct {
	AccessToken    string `json:"access_token"`
	Scope          string `json:"scope"`
	RefreshToken   string `json:"refresh_token,omitempty"`
	TokenType      string `json:"token_type,omitempty"`
	ExpiresIn      int    `json:"expires_in"`
	ResourceServer string `json:"resource_server"`
}

// OAuth2GetDependentTokens performs the dependent-token grant (POST /v2/oauth2/token).
func (c *Client) OAuth2GetDependentTokens(ctx context.Context, token string, refreshTokens bool, scopes []string) ([]DependentTokenInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	form := url.Values{}
	form.Set("grant_type", "urn:globus:auth:grant_type:dependent_token")
	form.Set("token", token)
	if refreshTokens {
		form.Set("access_type", "offline")
	}
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	var tokens []DependentTokenInfo
	if err := c.doForm(ctx, "oauth2/token", form, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// OAuth2ValidateToken validates a token (POST /v2/oauth2/token/validate). This
// endpoint is deprecated upstream but present at 3.65.0.
func (c *Client) OAuth2ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	form := url.Values{}
	form.Set("token", token)
	if c.ClientID != "" {
		form.Set("client_id", c.ClientID)
	}
	var result map[string]interface{}
	if err := c.doForm(ctx, "oauth2/token/validate", form, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// doForm performs a form-encoded POST and decodes the JSON response.
func (c *Client) doForm(ctx context.Context, path string, form url.Values, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Client.BaseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return nil
}

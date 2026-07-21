// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// UpdateProject updates a project (PUT /api/projects/{id}).
func (c *Client) UpdateProject(ctx context.Context, projectID string, update *ProjectUpdate) (*Project, error) {
	if projectID == "" {
		return nil, &core.ValidationError{Field: "projectID", Message: "project ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update document is required"}
	}

	body := map[string]interface{}{"project": update}
	var result struct {
		Project Project `json:"project"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/projects/%s", projectID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Project, nil
}

// GetIdentities looks up identities by username or by ID (GET /api/identities).
// Usernames and IDs are mutually exclusive; Provision is only sent with Usernames.
func (c *Client) GetIdentities(ctx context.Context, opts *GetIdentitiesOptions) ([]Identity, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.Usernames) > 0 && len(opts.IDs) > 0 {
			return nil, &core.ValidationError{Field: "opts", Message: "usernames and ids are mutually exclusive"}
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

	var envelope struct {
		Identities []Identity `json:"identities"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/identities", query, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Identities, nil
}

// GetIdentityProviders looks up identity providers by domain or ID
// (GET /api/identity_providers). Domains and IDs are mutually exclusive.
func (c *Client) GetIdentityProviders(ctx context.Context, opts *GetIdentityProvidersOptions) ([]IdentityProvider, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.Domains) > 0 && len(opts.IDs) > 0 {
			return nil, &core.ValidationError{Field: "opts", Message: "domains and ids are mutually exclusive"}
		}
		if len(opts.Domains) > 0 {
			query.Set("domains", strings.Join(opts.Domains, ","))
		}
		if len(opts.IDs) > 0 {
			query.Set("ids", strings.Join(opts.IDs, ","))
		}
	}

	var envelope struct {
		IdentityProviders []IdentityProvider `json:"identity_providers"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/identity_providers", query, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.IdentityProviders, nil
}

// GetPolicy retrieves a policy (GET /api/policies/{id}).
func (c *Client) GetPolicy(ctx context.Context, policyID string) (*Policy, error) {
	if policyID == "" {
		return nil, &core.ValidationError{Field: "policyID", Message: "policy ID is required"}
	}
	var result struct {
		Policy Policy `json:"policy"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/policies/%s", policyID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result.Policy, nil
}

// GetPolicies lists policies (GET /api/policies).
func (c *Client) GetPolicies(ctx context.Context) ([]Policy, error) {
	var envelope struct {
		Policies []Policy `json:"policies"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/policies", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Policies, nil
}

// CreatePolicy creates a policy (POST /api/policies).
func (c *Client) CreatePolicy(ctx context.Context, policy *PolicyCreate) (*Policy, error) {
	if policy == nil {
		return nil, &core.ValidationError{Field: "policy", Message: "policy is required"}
	}
	body := map[string]interface{}{"policy": policy}
	var result struct {
		Policy Policy `json:"policy"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/api/policies", nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Policy, nil
}

// UpdatePolicy updates a policy (PUT /api/policies/{id}).
func (c *Client) UpdatePolicy(ctx context.Context, policyID string, update *PolicyUpdate) (*Policy, error) {
	if policyID == "" {
		return nil, &core.ValidationError{Field: "policyID", Message: "policy ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update document is required"}
	}
	body := map[string]interface{}{"policy": update}
	var result struct {
		Policy Policy `json:"policy"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/policies/%s", policyID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Policy, nil
}

// DeletePolicy deletes a policy (DELETE /api/policies/{id}).
func (c *Client) DeletePolicy(ctx context.Context, policyID string) error {
	if policyID == "" {
		return &core.ValidationError{Field: "policyID", Message: "policy ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/policies/%s", policyID), nil, nil, nil)
}

// GetClient retrieves a registered client by ID or FQDN (exactly one)
// (GET /api/clients/{id} or GET /api/clients?fqdn=...).
func (c *Client) GetClient(ctx context.Context, opts *GetClientOptions) (*AuthClientInfo, error) {
	if opts == nil || (opts.ClientID == "" && opts.FQDN == "") {
		return nil, &core.ValidationError{Field: "opts", Message: "exactly one of ClientID or FQDN is required"}
	}
	if opts.ClientID != "" && opts.FQDN != "" {
		return nil, &core.ValidationError{Field: "opts", Message: "ClientID and FQDN are mutually exclusive"}
	}

	endpoint := "/api/clients"
	var query url.Values
	if opts.ClientID != "" {
		endpoint = fmt.Sprintf("/api/clients/%s", opts.ClientID)
	} else {
		query = url.Values{}
		query.Set("fqdn", opts.FQDN)
	}

	var result struct {
		Client AuthClientInfo `json:"client"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, endpoint, query, nil, &result); err != nil {
		return nil, err
	}
	return &result.Client, nil
}

// GetClients lists registered clients (GET /api/clients).
func (c *Client) GetClients(ctx context.Context) ([]AuthClientInfo, error) {
	var envelope struct {
		Clients []AuthClientInfo `json:"clients"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/clients", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Clients, nil
}

// CreateClient creates a client (POST /api/clients).
func (c *Client) CreateClient(ctx context.Context, req *ClientCreate) (*AuthClientInfo, error) {
	if req == nil {
		return nil, &core.ValidationError{Field: "req", Message: "client create document is required"}
	}
	body := map[string]interface{}{"client": req}
	var result struct {
		Client AuthClientInfo `json:"client"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/api/clients", nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Client, nil
}

// UpdateClient updates a client (PUT /api/clients/{id}).
func (c *Client) UpdateClient(ctx context.Context, clientID string, update *ClientUpdate) (*AuthClientInfo, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update document is required"}
	}
	body := map[string]interface{}{"client": update}
	var result struct {
		Client AuthClientInfo `json:"client"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/clients/%s", clientID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Client, nil
}

// DeleteClient deletes a client (DELETE /api/clients/{id}).
func (c *Client) DeleteClient(ctx context.Context, clientID string) error {
	if clientID == "" {
		return &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/clients/%s", clientID), nil, nil, nil)
}

// GetClientCredentials lists a client's credentials
// (GET /api/clients/{id}/credentials).
func (c *Client) GetClientCredentials(ctx context.Context, clientID string) ([]Credential, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	var envelope struct {
		Credentials []Credential `json:"credentials"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/clients/%s/credentials", clientID), nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Credentials, nil
}

// CreateClientCredential creates a client credential
// (POST /api/clients/{id}/credentials). The secret is returned only on create.
func (c *Client) CreateClientCredential(ctx context.Context, clientID, name string) (*Credential, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	body := map[string]interface{}{"credential": map[string]string{"name": name}}
	var result struct {
		Credential Credential `json:"credential"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/clients/%s/credentials", clientID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Credential, nil
}

// DeleteClientCredential deletes a client credential
// (DELETE /api/clients/{id}/credentials/{credential_id}).
func (c *Client) DeleteClientCredential(ctx context.Context, clientID, credentialID string) error {
	if clientID == "" {
		return &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if credentialID == "" {
		return &core.ValidationError{Field: "credentialID", Message: "credential ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/clients/%s/credentials/%s", clientID, credentialID), nil, nil, nil)
}

// GetScope retrieves a scope (GET /api/scopes/{id}).
func (c *Client) GetScope(ctx context.Context, scopeID string) (*Scope, error) {
	if scopeID == "" {
		return nil, &core.ValidationError{Field: "scopeID", Message: "scope ID is required"}
	}
	var result struct {
		Scope Scope `json:"scope"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/scopes/%s", scopeID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result.Scope, nil
}

// GetScopes lists scopes (GET /api/scopes). ScopeStrings and IDs are mutually
// exclusive.
func (c *Client) GetScopes(ctx context.Context, opts *GetScopesOptions) ([]Scope, error) {
	query := url.Values{}
	if opts != nil {
		if len(opts.ScopeStrings) > 0 && len(opts.IDs) > 0 {
			return nil, &core.ValidationError{Field: "opts", Message: "scope_strings and ids are mutually exclusive"}
		}
		if len(opts.ScopeStrings) > 0 {
			query.Set("scope_strings", strings.Join(opts.ScopeStrings, ","))
		}
		if len(opts.IDs) > 0 {
			query.Set("ids", strings.Join(opts.IDs, ","))
		}
	}

	var envelope struct {
		Scopes []Scope `json:"scopes"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, "/api/scopes", query, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Scopes, nil
}

// CreateScope creates a scope on a client (POST /api/clients/{id}/scopes).
func (c *Client) CreateScope(ctx context.Context, clientID string, req *ScopeCreate) (*Scope, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if req == nil {
		return nil, &core.ValidationError{Field: "req", Message: "scope create document is required"}
	}
	body := map[string]interface{}{"scope": req}
	var result struct {
		Scope Scope `json:"scope"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, fmt.Sprintf("/api/clients/%s/scopes", clientID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Scope, nil
}

// UpdateScope updates a scope (PUT /api/scopes/{id}).
func (c *Client) UpdateScope(ctx context.Context, scopeID string, update *ScopeUpdate) (*Scope, error) {
	if scopeID == "" {
		return nil, &core.ValidationError{Field: "scopeID", Message: "scope ID is required"}
	}
	if update == nil {
		return nil, &core.ValidationError{Field: "update", Message: "update document is required"}
	}
	body := map[string]interface{}{"scope": update}
	var result struct {
		Scope Scope `json:"scope"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPut, fmt.Sprintf("/api/scopes/%s", scopeID), nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Scope, nil
}

// DeleteScope deletes a scope (DELETE /api/scopes/{id}).
func (c *Client) DeleteScope(ctx context.Context, scopeID string) error {
	if scopeID == "" {
		return &core.ValidationError{Field: "scopeID", Message: "scope ID is required"}
	}
	return c.baseClient.DoRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/scopes/%s", scopeID), nil, nil, nil)
}

// GetConsents lists the consents for an identity
// (GET /api/identities/{id}/consents). When all is true, all consents are
// returned rather than only the active set.
func (c *Client) GetConsents(ctx context.Context, identityID string, all bool) ([]Consent, error) {
	if identityID == "" {
		return nil, &core.ValidationError{Field: "identityID", Message: "identity ID is required"}
	}
	query := url.Values{}
	if all {
		query.Set("all", "true")
	}
	var envelope struct {
		Consents []Consent `json:"consents"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodGet, fmt.Sprintf("/api/identities/%s/consents", identityID), query, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Consents, nil
}

// GetDependentTokens performs the dependent-token grant (POST /oauth2/token),
// returning one token per dependent resource server.
//
// Like introspection, this grant authenticates the *client*. A confidential
// client must construct the auth Client with
// authorizers.NewBasicAuthAuthorizer(clientID, clientSecret) so the request is
// client-authenticated via HTTP Basic auth.
func (c *Client) GetDependentTokens(ctx context.Context, token string, opts *DependentTokensOptions) ([]DependentTokenInfo, error) {
	if token == "" {
		return nil, &core.ValidationError{Field: "token", Message: "token is required"}
	}

	data := url.Values{}
	data.Set("grant_type", "urn:globus:auth:grant_type:dependent_token")
	data.Set("token", token)
	if opts != nil {
		if opts.RefreshTokens {
			data.Set("access_type", "offline")
		}
		if len(opts.Scopes) > 0 {
			data.Set("scope", strings.Join(opts.Scopes, " "))
		}
		for k, v := range opts.AdditionalParams {
			data.Set(k, v)
		}
	}

	var tokens []DependentTokenInfo
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token", nil, data, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// ClientCredentialsTokens performs the client_credentials grant
// (POST /oauth2/token), returning a token response. The client_id/client_secret
// are sent in the form body.
func (c *Client) ClientCredentialsTokens(ctx context.Context, clientID, clientSecret string, scopes []string) (*TokenResponse, error) {
	if clientID == "" {
		return nil, &core.ValidationError{Field: "clientID", Message: "client ID is required"}
	}
	if len(scopes) == 0 {
		return nil, &core.ValidationError{Field: "scopes", Message: "at least one scope is required"}
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", strings.Join(scopes, " "))
	data.Set("client_id", clientID)
	if clientSecret != "" {
		data.Set("client_secret", clientSecret)
	}

	var tokenResp TokenResponse
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/oauth2/token", nil, data, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

// CreateChildClient creates a child client under the calling confidential client
// (POST /api/clients). The child inherits the parent's project.
func (c *Client) CreateChildClient(ctx context.Context, req *ChildClientCreate) (*AuthClientInfo, error) {
	if req == nil {
		return nil, &core.ValidationError{Field: "req", Message: "child client create document is required"}
	}
	body := map[string]interface{}{"client": req}
	var result struct {
		Client AuthClientInfo `json:"client"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/api/clients", nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Client, nil
}

// CreateNativeAppInstance creates a native app instance from a template
// (POST /api/clients).
func (c *Client) CreateNativeAppInstance(ctx context.Context, templateID, name string) (*AuthClientInfo, error) {
	if templateID == "" {
		return nil, &core.ValidationError{Field: "templateID", Message: "template ID is required"}
	}
	body := map[string]interface{}{"client": map[string]string{"name": name, "template_id": templateID}}
	var result struct {
		Client AuthClientInfo `json:"client"`
	}
	if err := c.baseClient.DoRequest(ctx, http.MethodPost, "/api/clients", nil, body, &result); err != nil {
		return nil, err
	}
	return &result.Client, nil
}

// GetOpenIDConfiguration fetches the OIDC discovery document
// (GET /.well-known/openid-configuration at the Auth host root).
func (c *Client) GetOpenIDConfiguration(ctx context.Context) (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := c.baseClient.DoRequestURL(ctx, http.MethodGet, c.rootURL("/.well-known/openid-configuration"), nil, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// GetJWK fetches the JSON Web Key Set from the jwks_uri advertised in the OIDC
// discovery document.
func (c *Client) GetJWK(ctx context.Context) (map[string]interface{}, error) {
	config, err := c.GetOpenIDConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	jwksURI, ok := config["jwks_uri"].(string)
	if !ok || jwksURI == "" {
		return nil, &core.ValidationError{Field: "jwks_uri", Message: "OIDC configuration did not contain a jwks_uri"}
	}

	var jwks map[string]interface{}
	if err := c.baseClient.DoRequestURL(ctx, http.MethodGet, jwksURI, nil, &jwks); err != nil {
		return nil, err
	}
	return jwks, nil
}

// rootURL returns an absolute URL against the Auth host root, dropping any /v2
// (or other) path suffix carried by the configured base URL. Used for the
// host-root .well-known endpoints that live outside the /v2 API surface.
func (c *Client) rootURL(path string) string {
	base := c.baseURL
	if base == "" {
		base = "https://auth.globus.org"
	}
	if u, err := url.Parse(base); err == nil {
		u.Path = ""
		u.RawQuery = ""
		return strings.TrimRight(u.String(), "/") + path
	}
	return "https://auth.globus.org" + path
}

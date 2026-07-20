// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Config represents the configuration for all Globus SDK clients in v4
// This replaces the options pattern from v3 with explicit configuration
type Config struct {
	// AccessToken is the OAuth2 access token for authentication
	// Required if Authorizer is not set
	AccessToken string

	// Authorizer provides dynamic authorization headers (e.g., auto-refreshing tokens)
	// If set, takes precedence over AccessToken
	Authorizer Authorizer

	// Scopes are the explicitly required OAuth2 scopes for this client
	// v4 requires explicit scope specification for security
	Scopes []string

	// HTTPClient is the HTTP client to use for requests
	// If nil, a default client will be created
	HTTPClient *http.Client

	// BaseURL is the base URL for API requests
	// If empty, the default service URL will be used
	BaseURL string

	// Timeout is the request timeout
	// Default: 30 seconds
	Timeout time.Duration

	// RetryConfig specifies retry behavior
	RetryConfig *RetryConfig

	// UserAgent specifies a custom user agent string
	// If empty, the default SDK user agent will be used
	UserAgent string

	// Environment specifies the Globus environment (production, sandbox, preview)
	// Default: production
	Environment string
}

// RetryConfig specifies retry behavior for API requests
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	// Default: 3
	MaxRetries int

	// InitialBackoff is the initial backoff duration
	// Default: 1 second
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration
	// Default: 30 seconds
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier for exponential backoff
	// Default: 2.0
	BackoffMultiplier float64

	// RetryableStatusCodes are HTTP status codes that should trigger a retry
	// Default: [429, 500, 502, 503, 504]
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatusCodes: []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// In v4, either an access token or an authorizer must be provided
	if c.AccessToken == "" && c.Authorizer == nil {
		return &ValidationError{
			Field:   "AccessToken",
			Message: "access token or authorizer is required",
		}
	}

	// In v4, scopes must be explicitly specified
	if len(c.Scopes) == 0 {
		return &ValidationError{
			Field:   "Scopes",
			Message: "explicit scopes are required in v4 for security",
		}
	}

	return nil
}

// WithDefaults returns a config with default values filled in
func (c *Config) WithDefaults() *Config {
	config := *c

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryConfig == nil {
		config.RetryConfig = DefaultRetryConfig()
	}

	if config.UserAgent == "" {
		config.UserAgent = UserAgent()
	}

	if config.Environment == "" {
		config.Environment = "production"
	}

	return &config
}

// Authorizer provides authorization headers for HTTP requests.
// It is the primary abstraction for credential management in v4, mirroring
// the Python SDK's globus_sdk.authorizers module.
type Authorizer interface {
	// GetAuthorizationHeader returns the full Authorization header value (e.g. "Bearer <token>")
	GetAuthorizationHeader(ctx context.Context) (string, error)

	// HandleMissingAuthorization is called when the server returns 401.
	// Returns true if the authorizer believes a retry may succeed (e.g. after a refresh).
	HandleMissingAuthorization(ctx context.Context) bool
}

// TokenProvider is an interface for providing access tokens
// This supports both static and refreshable tokens
type TokenProvider interface {
	// GetAccessToken returns a valid access token
	GetAccessToken(ctx context.Context) (string, error)

	// RequiresRefresh returns true if the token needs refreshing
	RequiresRefresh() bool

	// Refresh refreshes the access token
	Refresh(ctx context.Context) error
}

// StaticTokenProvider provides a static access token
type StaticTokenProvider struct {
	Token string
}

// GetAccessToken returns the static token
func (p *StaticTokenProvider) GetAccessToken(ctx context.Context) (string, error) {
	if p.Token == "" {
		return "", &ValidationError{
			Message: "static token is empty",
		}
	}
	return p.Token, nil
}

// RequiresRefresh always returns false for static tokens
func (p *StaticTokenProvider) RequiresRefresh() bool {
	return false
}

// Refresh is a no-op for static tokens
func (p *StaticTokenProvider) Refresh(ctx context.Context) error {
	return nil
}

// Scopes defines well-known OAuth2 scopes for Globus services
var Scopes = struct {
	// Auth service scopes
	AuthOpenID  string
	AuthEmail   string
	AuthProfile string
	AuthManage  string

	// Transfer service scopes
	TransferAll string

	// Groups service scopes
	GroupsAll    string
	GroupsView   string
	GroupsManage string

	// Search service scopes
	SearchAll   string
	SearchIndex string

	// Flows service scopes
	FlowsAll string
	FlowsRun string

	// Timers service scopes
	TimersAll string

	// Compute service scopes
	ComputeAll string
}{
	// Auth
	AuthOpenID:  "openid",
	AuthEmail:   "email",
	AuthProfile: "profile",
	AuthManage:  "urn:globus:auth:scope:auth.globus.org:manage_projects",

	// Transfer
	TransferAll: "urn:globus:auth:scope:transfer.api.globus.org:all",

	// Groups
	GroupsAll:    "urn:globus:auth:scope:groups.api.globus.org:all",
	GroupsView:   "urn:globus:auth:scope:groups.api.globus.org:view_my_groups_and_memberships",
	GroupsManage: "urn:globus:auth:scope:groups.api.globus.org:manage_groups",

	// Search
	SearchAll:   "urn:globus:auth:scope:search.api.globus.org:all",
	SearchIndex: "urn:globus:auth:scope:index.globus.org:all",

	// Flows
	FlowsAll: "https://auth.globus.org/scopes/flows.globus.org/all",
	FlowsRun: "https://auth.globus.org/scopes/flows.globus.org/run",

	// Timers
	TimersAll: "https://auth.globus.org/scopes/timer.automate.globus.org/all",

	// Compute
	ComputeAll: "https://auth.globus.org/scopes/compute.api.globus.org/all",
}

// MakeFlowScope creates a flow-specific scope for a given flow ID
func MakeFlowScope(flowID string) string {
	return fmt.Sprintf("https://auth.globus.org/scopes/%s/flow_run", flowID)
}

// MakeCollectionScope creates a collection-specific scope for Transfer
func MakeCollectionScope(collectionID string, access string) string {
	if access == "" {
		access = "all"
	}
	return fmt.Sprintf("https://auth.globus.org/scopes/%s/%s", collectionID, access)
}

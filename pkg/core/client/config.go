// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package client provides unified client configuration for the Globus Go SDK.
//
// This package implements consistent client initialization patterns across
// all Globus services, following the patterns established by the Python SDK
// for compatibility and familiarity.
package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/authorizers"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/deprecation"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/errors"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"
)

// Config contains configuration for Globus service clients
type Config struct {
	// Authorizer handles authentication for requests
	Authorizer interfaces.Authorizer

	// BaseURL is the base URL for the service
	BaseURL string

	// HTTPClient is the HTTP client to use for requests
	HTTPClient *http.Client

	// UserAgent is the User-Agent header to send with requests
	UserAgent string

	// Logger is used for logging requests and responses
	Logger interfaces.Logger

	// RetryPolicy defines retry behavior for failed requests
	RetryPolicy *RetryPolicy

	// RateLimit defines rate limiting configuration
	RateLimit *RateLimitConfig

	// Timeout is the default timeout for requests
	Timeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int

	// EnableTLS controls whether TLS is enabled
	EnableTLS bool

	// TLSConfig provides custom TLS configuration
	TLSConfig *TLSConfig

	// Service is the name of the service (e.g., "auth", "transfer")
	Service string

	// APIVersion is the API version to use
	APIVersion string

	// Environment is the environment to use (e.g., "production", "preview")
	Environment string

	// Context is the default context for requests
	Context context.Context

	// Debug enables debug logging
	Debug bool
}

// RetryPolicy defines retry behavior for failed requests
type RetryPolicy struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int

	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the backoff multiplier
	Multiplier float64

	// Jitter adds randomness to delays
	Jitter bool

	// RetryableStatusCodes are HTTP status codes that should be retried
	RetryableStatusCodes []int

	// RetryableErrors are error types that should be retried
	RetryableErrors []error
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	// Enabled controls whether rate limiting is enabled
	Enabled bool

	// RequestsPerSecond is the maximum requests per second
	RequestsPerSecond float64

	// BurstSize is the burst size for rate limiting
	BurstSize int

	// RespectServerLimits controls whether to respect server-provided limits
	RespectServerLimits bool
}

// TLSConfig provides custom TLS configuration
type TLSConfig struct {
	// InsecureSkipVerify controls whether to skip TLS verification
	InsecureSkipVerify bool

	// ServerName is the server name for TLS verification
	ServerName string

	// CertFile is the path to the client certificate file
	CertFile string

	// KeyFile is the path to the client key file
	KeyFile string

	// CAFile is the path to the CA certificate file
	CAFile string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent:   core.UserAgent(),
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableTLS:   true,
		Environment: "production",
		Context:     context.Background(),
		RetryPolicy: &RetryPolicy{
			MaxRetries:   3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
			RetryableStatusCodes: []int{
				http.StatusTooManyRequests,
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
			},
		},
		RateLimit: &RateLimitConfig{
			Enabled:             true,
			RequestsPerSecond:   10.0,
			BurstSize:           20,
			RespectServerLimits: true,
		},
	}
}

// NewConfig creates a new client configuration
func NewConfig(service string, options ...ConfigOption) (*Config, error) {
	config := DefaultConfig()
	config.Service = service

	// Apply options
	for _, option := range options {
		if err := option(config); err != nil {
			return nil, fmt.Errorf("failed to apply config option: %w", err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Service == "" {
		return errors.NewGlobusError("client", "InvalidConfig", "service name is required")
	}

	if c.BaseURL == "" {
		return errors.NewGlobusError("client", "InvalidConfig", "base URL is required")
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return errors.NewGlobusError("client", "InvalidConfig", "invalid base URL").WithDetail(err.Error())
	}

	if c.Timeout <= 0 {
		return errors.NewGlobusError("client", "InvalidConfig", "timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return errors.NewGlobusError("client", "InvalidConfig", "max retries must be non-negative")
	}

	return nil
}

// ConfigOption is a function that configures a Config
type ConfigOption func(*Config) error

// WithAuthorizer sets the authorizer
func WithAuthorizer(authorizer interfaces.Authorizer) ConfigOption {
	return func(c *Config) error {
		c.Authorizer = authorizer
		return nil
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(baseURL string) ConfigOption {
	return func(c *Config) error {
		c.BaseURL = baseURL
		return nil
	}
}

// WithHTTPClient sets the HTTP client
func WithHTTPClient(client *http.Client) ConfigOption {
	return func(c *Config) error {
		c.HTTPClient = client
		return nil
	}
}

// WithUserAgent sets the user agent
func WithUserAgent(userAgent string) ConfigOption {
	return func(c *Config) error {
		c.UserAgent = userAgent
		return nil
	}
}

// WithLogger sets the logger
func WithLogger(logger interfaces.Logger) ConfigOption {
	return func(c *Config) error {
		c.Logger = logger
		return nil
	}
}

// WithTimeout sets the timeout
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) error {
		c.Timeout = timeout
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) ConfigOption {
	return func(c *Config) error {
		c.MaxRetries = maxRetries
		return nil
	}
}

// WithRetryPolicy sets the retry policy
func WithRetryPolicy(policy *RetryPolicy) ConfigOption {
	return func(c *Config) error {
		c.RetryPolicy = policy
		return nil
	}
}

// WithRateLimit sets the rate limit configuration
func WithRateLimit(config *RateLimitConfig) ConfigOption {
	return func(c *Config) error {
		c.RateLimit = config
		return nil
	}
}

// WithAPIVersion sets the API version
func WithAPIVersion(version string) ConfigOption {
	return func(c *Config) error {
		c.APIVersion = version
		return nil
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env string) ConfigOption {
	return func(c *Config) error {
		c.Environment = env
		return nil
	}
}

// WithContext sets the default context
func WithContext(ctx context.Context) ConfigOption {
	return func(c *Config) error {
		c.Context = ctx
		return nil
	}
}

// WithDebug enables debug logging
func WithDebug(debug bool) ConfigOption {
	return func(c *Config) error {
		c.Debug = debug
		return nil
	}
}

// WithTLSConfig sets the TLS configuration
func WithTLSConfig(config *TLSConfig) ConfigOption {
	return func(c *Config) error {
		c.TLSConfig = config
		return nil
	}
}

// WithAccessToken creates an authorizer from an access token
func WithAccessToken(token string) ConfigOption {
	return func(c *Config) error {
		c.Authorizer = authorizers.NewStaticTokenAuthorizer(token)
		return nil
	}
}

// WithClientCredentials creates an authorizer from client credentials
func WithClientCredentials(clientID, clientSecret string) ConfigOption {
	return func(c *Config) error {
		// For now, we'll create a null authorizer as a placeholder
		// This would need to be implemented with the actual client credentials authorizer
		c.Authorizer = &authorizers.NullAuthorizer{}
		return nil
	}
}

// Service-specific configuration helpers

// AuthConfig creates a configuration for the Auth service
func AuthConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://auth.globus.org/v2"),
		WithAPIVersion("v2"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("auth", allOptions...)
}

// TransferConfig creates a configuration for the Transfer service
func TransferConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://transfer.api.globus.org/v0.10"),
		WithAPIVersion("v0.10"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("transfer", allOptions...)
}

// GroupsConfig creates a configuration for the Groups service
func GroupsConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://groups.api.globus.org/v2"),
		WithAPIVersion("v2"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("groups", allOptions...)
}

// SearchConfig creates a configuration for the Search service
func SearchConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://search.api.globus.org/v1"),
		WithAPIVersion("v1"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("search", allOptions...)
}

// FlowsConfig creates a configuration for the Flows service
func FlowsConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://flows.globus.org/v1"),
		WithAPIVersion("v1"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("flows", allOptions...)
}

// ComputeConfig creates a configuration for the Compute service
func ComputeConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://compute.api.globus.org/v2"),
		WithAPIVersion("v2"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("compute", allOptions...)
}

// TimersConfig creates a configuration for the Timers service
func TimersConfig(options ...ConfigOption) (*Config, error) {
	baseOptions := []ConfigOption{
		WithBaseURL("https://timer.automate.globus.org/api/v1"),
		WithAPIVersion("v1"),
	}

	allOptions := append(baseOptions, options...)
	return NewConfig("timers", allOptions...)
}

// Legacy compatibility functions (with deprecation warnings)

// LegacyNewAuthClient creates an auth client using legacy patterns
func LegacyNewAuthClient(token string) error {
	deprecation.WarnLegacyClientInit("Auth")
	return fmt.Errorf("legacy client initialization is deprecated, use auth.NewClient() instead")
}

// LegacyNewTransferClient creates a transfer client using legacy patterns
func LegacyNewTransferClient(token string) error {
	deprecation.WarnLegacyClientInit("Transfer")
	return fmt.Errorf("legacy client initialization is deprecated, use transfer.NewClient() instead")
}

// LegacyNewGroupsClient creates a groups client using legacy patterns
func LegacyNewGroupsClient(token string) error {
	deprecation.WarnLegacyClientInit("Groups")
	return fmt.Errorf("legacy client initialization is deprecated, use groups.NewClient() instead")
}

// LegacyNewSearchClient creates a search client using legacy patterns
func LegacyNewSearchClient(token string) error {
	deprecation.WarnLegacyClientInit("Search")
	return fmt.Errorf("legacy client initialization is deprecated, use search.NewClient() instead")
}

// LegacyNewFlowsClient creates a flows client using legacy patterns
func LegacyNewFlowsClient(token string) error {
	deprecation.WarnLegacyClientInit("Flows")
	return fmt.Errorf("legacy client initialization is deprecated, use flows.NewClient() instead")
}

// LegacyNewComputeClient creates a compute client using legacy patterns
func LegacyNewComputeClient(token string) error {
	deprecation.WarnLegacyClientInit("Compute")
	return fmt.Errorf("legacy client initialization is deprecated, use compute.NewClient() instead")
}

// LegacyNewTimersClient creates a timers client using legacy patterns
func LegacyNewTimersClient(token string) error {
	deprecation.WarnLegacyClientInit("Timers")
	return fmt.Errorf("legacy client initialization is deprecated, use timers.NewClient() instead")
}

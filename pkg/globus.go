// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package pkg

import (
	"context"
	"fmt"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/config"
	httppool "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
	"os"
	"time"
)

// Version is the SDK version
const Version = "0.9.7"

// OAuth2 scopes for Globus services
const (
	// AuthScope is the scope for the Auth service
	AuthScope = auth.AuthScope

	// GroupsScope is the scope for the Groups service
	GroupsScope = groups.GroupsScope

	// TransferScope is the scope for the Transfer service
	TransferScope = transfer.TransferScope

	// SearchScope is the scope for the Search service
	SearchScope = search.SearchScope

	// FlowsScope is the scope for the Flows service
	FlowsScope = flows.FlowsScope

	// ComputeScope is the scope for the Compute service
	ComputeScope = compute.ComputeScope

	// TimersScope is the scope for the Timers service
	TimersScope = timers.TimersScope

	// TokensScope is the scope for token management (uses AuthScope)
	TokensScope = auth.AuthScope
)

// SDKConfig holds configuration for all services
type SDKConfig struct {
	Config       *config.Config
	ClientID     string
	ClientSecret string
}

// NewAuthClient creates a new Auth client with the SDK configuration
func (c *SDKConfig) NewAuthClient() (*auth.Client, error) {
	// Create auth client options
	options := []auth.ClientOption{
		auth.WithClientID(c.ClientID),
		auth.WithClientSecret(c.ClientSecret),
	}

	// Create the auth client
	authClient, err := auth.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(authClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("auth", nil)
		authClient.Client.HTTPClient = serviceClient
	}

	return authClient, nil
}

// NewGroupsClient creates a new Groups client with the SDK configuration
func (c *SDKConfig) NewGroupsClient(accessToken string) (*groups.Client, error) {
	// Create a simple static token authorizer directly
	authorizer := &simpleAuthorizer{token: accessToken}

	// Create options for the groups client
	options := []groups.Option{
		groups.WithAuthorizer(authorizer),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, groups.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, groups.WithHTTPTracing(true))
	}

	// Create the client
	groupsClient, err := groups.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating groups client: %w", err)
	}

	return groupsClient, nil
}

// NewTransferClient creates a new Transfer client with the SDK configuration
func (c *SDKConfig) NewTransferClient(accessToken string) (*transfer.Client, error) {
	// Create a simple static token authorizer directly
	// using a type that satisfies the auth.Authorizer interface
	authorizer := &simpleAuthorizer{token: accessToken}

	// Create options for the transfer client
	options := []transfer.Option{
		transfer.WithAuthorizer(authorizer),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, transfer.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, transfer.WithHTTPTracing(true))
	}

	// Create the client
	transferClient, err := transfer.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating transfer client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(transferClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("transfer", nil)
		transferClient.Client.HTTPClient = serviceClient
	}

	return transferClient, nil
}

// NewSearchClient creates a new Search client with the SDK configuration
func (c *SDKConfig) NewSearchClient(accessToken string) (*search.Client, error) {
	// Create options for the search client
	options := []search.ClientOption{
		search.WithAccessToken(accessToken),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, search.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, search.WithHTTPTracing(true))
	}

	// Create the client
	searchClient, err := search.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating search client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(searchClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("search", nil)
		searchClient.Client.HTTPClient = serviceClient
	}

	return searchClient, nil
}

// NewFlowsClient creates a new Flows client with the SDK configuration
func (c *SDKConfig) NewFlowsClient(accessToken string) (*flows.Client, error) {
	// Create options for the flows client
	options := []flows.ClientOption{
		flows.WithAccessToken(accessToken),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, flows.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, flows.WithHTTPTracing(true))
	}

	// Create the client
	flowsClient, err := flows.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating flows client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(flowsClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("flows", nil)
		flowsClient.Client.HTTPClient = serviceClient
	}

	return flowsClient, nil
}

// NewComputeClient creates a new Compute client with the SDK configuration
func (c *SDKConfig) NewComputeClient(accessToken string) (*compute.Client, error) {
	// Create options for the compute client
	options := []compute.ClientOption{
		compute.WithAccessToken(accessToken),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, compute.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, compute.WithHTTPTracing(true))
	}

	// Create the client
	computeClient, err := compute.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating compute client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(computeClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("compute", nil)
		computeClient.Client.HTTPClient = serviceClient
	}

	return computeClient, nil
}

// NewTimersClient creates a new Timers client with the SDK configuration
func (c *SDKConfig) NewTimersClient(accessToken string) (*timers.Client, error) {
	// Create options for the timers client
	options := []timers.ClientOption{
		timers.WithAccessToken(accessToken),
	}

	// Add debugging if configured
	if os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1" {
		options = append(options, timers.WithHTTPDebugging(true))
	}

	if os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1" {
		options = append(options, timers.WithHTTPTracing(true))
	}

	// Create the client
	timersClient, err := timers.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("error creating timers client: %w", err)
	}

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(timersClient.Client)
	}

	// Use service-specific connection pool if enabled
	if os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true" {
		serviceClient := httppool.GetHTTPClientForService("timers", nil)
		timersClient.Client.HTTPClient = serviceClient
	}

	return timersClient, nil
}

// NewTokenManager creates a new Token Manager with the SDK configuration
func (c *SDKConfig) NewTokenManager(opts ...tokens.ClientOption) (*tokens.Manager, error) {
	// Create a new token manager with the provided options
	tokenManager, err := tokens.NewManager(opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating token manager: %w", err)
	}

	return tokenManager, nil
}

// NewTokenManagerWithAuth creates a new Token Manager using an Auth client for token refreshing
func (c *SDKConfig) NewTokenManagerWithAuth(storageDirectory string) (*tokens.Manager, error) {
	// Create an auth client
	authClient, err := c.NewAuthClient()
	if err != nil {
		return nil, fmt.Errorf("error creating auth client for token manager: %w", err)
	}

	// Create token manager options
	options := []tokens.ClientOption{
		tokens.WithAuthClient(authClient),
	}

	// If a storage directory is provided, use file storage
	if storageDirectory != "" {
		options = append(options, tokens.WithFileStorage(storageDirectory))
	}

	// Create the token manager
	return c.NewTokenManager(options...)
}

// NewConfig creates a new SDK configuration
func NewConfig() *SDKConfig {
	return &SDKConfig{
		Config: config.DefaultConfig(),
	}
}

// NewConfigFromEnvironment creates a new SDK configuration from environment variables
func NewConfigFromEnvironment() *SDKConfig {
	// Initialize connection pooling based on environment variable
	useConnectionPool := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true"
	if useConnectionPool {
		initializeConnectionPools()
	}

	config := config.FromEnvironment()

	// If connection pooling is enabled, override HTTP clients
	if useConnectionPool {
		httpClient := httppool.GetHTTPClientForService("default", nil)
		config.HTTPClient = httpClient
	}

	return &SDKConfig{
		Config: config,
	}
}

// initializeConnectionPools sets up connection pools for all services
func initializeConnectionPools() {
	// Create the global pool manager if not already created
	httpPoolManager := httppool.NewHttpConnectionPoolManager(nil)

	// Create pools for each service with optimized settings
	serviceConfigs := map[string]*httppool.ConnectionPoolConfig{
		"auth": {
			MaxIdleConnsPerHost: 4,
			MaxConnsPerHost:     8,
			IdleConnTimeout:     60 * time.Second,
		},
		"transfer": {
			MaxIdleConnsPerHost: 8,
			MaxConnsPerHost:     16,
			IdleConnTimeout:     120 * time.Second,
		},
		"search": {
			MaxIdleConnsPerHost: 6,
			MaxConnsPerHost:     12,
			IdleConnTimeout:     90 * time.Second,
		},
		"flows": {
			MaxIdleConnsPerHost: 6,
			MaxConnsPerHost:     12,
			IdleConnTimeout:     90 * time.Second,
		},
		"groups": {
			MaxIdleConnsPerHost: 4,
			MaxConnsPerHost:     8,
			IdleConnTimeout:     60 * time.Second,
		},
		"compute": {
			MaxIdleConnsPerHost: 6,
			MaxConnsPerHost:     16,
			IdleConnTimeout:     120 * time.Second,
		},
		"timers": {
			MaxIdleConnsPerHost: 4,
			MaxConnsPerHost:     8,
			IdleConnTimeout:     60 * time.Second,
		},
		"default": nil, // Use defaults for the default pool
	}

	// Initialize all service pools using the HTTP adapter manager
	for service, poolConfig := range serviceConfigs {
		httpPoolManager.GetPool(service, poolConfig)
	}
}

// WithClientID sets the client ID
func (c *SDKConfig) WithClientID(clientID string) *SDKConfig {
	c.ClientID = clientID
	return c
}

// WithClientSecret sets the client secret
func (c *SDKConfig) WithClientSecret(clientSecret string) *SDKConfig {
	c.ClientSecret = clientSecret
	return c
}

// WithConfig sets the configuration
func (c *SDKConfig) WithConfig(config *config.Config) *SDKConfig {
	c.Config = config
	return c
}

// WithClientOption adds a client option to the configuration
func (c *SDKConfig) WithClientOption(option core.ClientOption) *SDKConfig {
	if c.Config == nil {
		c.Config = config.DefaultConfig()
	}

	// Create a temporary client to apply the option
	tempClient := &core.Client{}
	option(tempClient)

	return c
}

// GetScopesByService returns the OAuth2 scopes needed for the specified services
func GetScopesByService(services ...string) []string {
	scopes := make([]string, 0, len(services))

	for _, service := range services {
		switch service {
		case "auth":
			scopes = append(scopes, AuthScope)
		case "groups":
			scopes = append(scopes, GroupsScope)
		case "transfer":
			scopes = append(scopes, TransferScope)
		case "search":
			scopes = append(scopes, SearchScope)
		case "flows":
			scopes = append(scopes, FlowsScope)
		case "compute":
			scopes = append(scopes, ComputeScope)
		}
	}

	return scopes
}

// simpleAuthorizer is a simple implementation of the auth.Authorizer interface
type simpleAuthorizer struct {
	token string
}

// GetAuthorizationHeader returns the authorization header value
func (a *simpleAuthorizer) GetAuthorizationHeader(_ ...context.Context) (string, error) {
	if a.token == "" {
		return "", nil
	}
	return "Bearer " + a.token, nil
}

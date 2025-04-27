// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pkg

import (
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/config"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/compute"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/groups"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/timers"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// Version is the SDK version
const Version = "0.1.0"

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
)

// SDKConfig holds configuration for all services
type SDKConfig struct {
	Config       *config.Config
	ClientID     string
	ClientSecret string
}

// NewAuthClient creates a new Auth client with the SDK configuration
func (c *SDKConfig) NewAuthClient() *auth.Client {
	authClient := auth.NewClient(c.ClientID, c.ClientSecret)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(authClient.Client)
	}

	return authClient
}

// NewGroupsClient creates a new Groups client with the SDK configuration
func (c *SDKConfig) NewGroupsClient(accessToken string) *groups.Client {
	groupsClient := groups.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(groupsClient.Client)
	}

	return groupsClient
}

// NewTransferClient creates a new Transfer client with the SDK configuration
func (c *SDKConfig) NewTransferClient(accessToken string) *transfer.Client {
	transferClient := transfer.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(transferClient.Client)
	}

	return transferClient
}

// NewSearchClient creates a new Search client with the SDK configuration
func (c *SDKConfig) NewSearchClient(accessToken string) *search.Client {
	searchClient := search.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(searchClient.Client)
	}

	return searchClient
}

// NewFlowsClient creates a new Flows client with the SDK configuration
func (c *SDKConfig) NewFlowsClient(accessToken string) *flows.Client {
	flowsClient := flows.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(flowsClient.Client)
	}

	return flowsClient
}

// NewComputeClient creates a new Compute client with the SDK configuration
func (c *SDKConfig) NewComputeClient(accessToken string) *compute.Client {
	computeClient := compute.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(computeClient.Client)
	}

	return computeClient
}

// NewTimersClient creates a new Timers client with the SDK configuration
func (c *SDKConfig) NewTimersClient(accessToken string) *timers.Client {
	timersClient := timers.NewClient(accessToken)

	// Apply configuration
	if c.Config != nil {
		c.Config.ApplyToClient(timersClient.Client)
	}

	return timersClient
}

// NewConfig creates a new SDK configuration
func NewConfig() *SDKConfig {
	return &SDKConfig{
		Config: config.DefaultConfig(),
	}
}

// NewConfigFromEnvironment creates a new SDK configuration from environment variables
func NewConfigFromEnvironment() *SDKConfig {
	return &SDKConfig{
		Config: config.FromEnvironment(),
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

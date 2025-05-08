// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package config

import (
	"net/http"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// Config represents the configuration for the SDK
type Config struct {
	// HTTPClient is the HTTP client to use for API requests
	HTTPClient *http.Client

	// BaseURL is the base URL for API requests
	BaseURL string

	// UserAgent is the user agent for API requests
	UserAgent string

	// LogLevel is the logging level for the SDK
	LogLevel core.LogLevel

	// Timeout is the timeout for API requests
	Timeout time.Duration

	// RetryMax is the maximum number of retries for API requests
	RetryMax int

	// RetryWaitMin is the minimum time to wait between retries
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum time to wait between retries
	RetryWaitMax time.Duration

	// VersionCheck manages API version checking
	VersionCheck *core.VersionCheck

	// Debug enables debug mode for HTTP operations
	Debug bool

	// Trace enables distributed tracing
	Trace bool

	// CustomTransport allows for custom transport configuration
	CustomTransport interfaces.Transport
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	// Check if connection pooling is disabled
	disablePooling := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") == "true"

	var httpClient *http.Client

	if !disablePooling {
		// Use connection pooling by default
		// Import indirectly to avoid circular imports
		httpClient = core.NewHTTPClientWithConnectionPool("default", nil)
	} else {
		// Use standard HTTP client
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}

	return &Config{
		HTTPClient:   httpClient,
		UserAgent:    "globus-go-sdk/1.0",
		LogLevel:     core.LogLevelNone,
		Timeout:      time.Second * 30,
		RetryMax:     3,
		RetryWaitMin: time.Second,
		RetryWaitMax: time.Second * 5,
		VersionCheck: core.NewVersionCheck(),
	}
}

// FromEnvironment loads configuration from environment variables
func FromEnvironment() *Config {
	config := DefaultConfig()

	if val := os.Getenv("GLOBUS_SDK_BASE_URL"); val != "" {
		config.BaseURL = val
	}

	if val := os.Getenv("GLOBUS_SDK_USER_AGENT"); val != "" {
		config.UserAgent = val
	}

	return config
}

// ApplyToClient applies the configuration to a client
func (c *Config) ApplyToClient(client *core.Client) {
	if client == nil {
		return
	}

	if c.HTTPClient != nil {
		client.HTTPClient = c.HTTPClient
	}

	if c.BaseURL != "" {
		client.BaseURL = c.BaseURL
	}

	if c.UserAgent != "" {
		client.UserAgent = c.UserAgent
	}

	if c.LogLevel != core.LogLevelNone {
		client.Logger = core.NewDefaultLogger(nil, c.LogLevel)
	}

	// Apply VersionCheck if set
	if client.VersionCheck == nil && c.VersionCheck != nil {
		client.VersionCheck = c.VersionCheck
	}

	// Apply Debug and Tracing settings
	if c.Debug {
		client.Debug = true
	}

	if c.Trace {
		client.Trace = true
	}

	// Apply CustomTransport if set
	if c.CustomTransport != nil {
		client.Transport = c.CustomTransport
	}
}

// GetVersionCheck returns the version check manager
func (c *Config) GetVersionCheck() *core.VersionCheck {
	return c.VersionCheck
}

// SetVersionCheck sets the version check manager
func (c *Config) SetVersionCheck(vc *core.VersionCheck) {
	c.VersionCheck = vc
}

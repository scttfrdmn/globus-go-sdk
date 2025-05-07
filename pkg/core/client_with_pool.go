// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
)

// Global connection pool manager interface
var globalPoolManager interfaces.ConnectionPoolManager

// SetConnectionPoolManager sets the global connection pool manager
func SetConnectionPoolManager(manager interfaces.ConnectionPoolManager) {
	globalPoolManager = manager
}

// GetConnectionPool returns a connection pool for the given service
func GetConnectionPool(service string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	if globalPoolManager == nil {
		return nil
	}
	return globalPoolManager.GetPool(service, config)
}

// WithConnectionPool configures the client to use a connection pool
func WithConnectionPool(poolName string, config interfaces.ConnectionPoolConfig) ClientOption {
	return func(c *Client) {
		// Get or create a connection pool
		if globalPoolManager == nil {
			return
		}

		pool := globalPoolManager.GetPool(poolName, config)
		if pool != nil {
			// Use the HTTP client from the pool
			c.HTTPClient = pool.GetClient()
		}
	}
}

// WithPooledClient configures the client with a pre-existing HTTP client
// from a connection pool
func WithPooledClient(pool interfaces.ConnectionPool) ClientOption {
	return func(c *Client) {
		if pool != nil {
			c.HTTPClient = pool.GetClient()
		}
	}
}

// NewHTTPClientWithConnectionPool creates a new HTTP client with a connection pool
func NewHTTPClientWithConnectionPool(poolName string, config interfaces.ConnectionPoolConfig) *http.Client {
	if globalPoolManager == nil {
		return &http.Client{}
	}

	pool := globalPoolManager.GetPool(poolName, config)
	if pool == nil {
		return &http.Client{}
	}

	return pool.GetClient()
}

// EnableDefaultConnectionPool configures a default connection pool for all clients
// This should be called early in your application's initialization
func EnableDefaultConnectionPool() {
	if globalPoolManager == nil {
		return
	}

	// Create default connection pools for each service type
	serviceNames := []string{
		"auth",
		"transfer",
		"search",
		"flows",
		"groups",
		"compute",
		"timers",
	}

	for _, service := range serviceNames {
		// Use service-specific configurations
		config := pool.ForService(service)

		// Initialize the pool for the service
		globalPoolManager.GetPool(service, config)
	}
}

// ClientWithConnectionPool creates a new client with connection pooling enabled
func ClientWithConnectionPool(service string, options ...ClientOption) *Client {
	// Add connection pooling option to the provided options
	poolOption := WithConnectionPool(service, nil)
	allOptions := append([]ClientOption{poolOption}, options...)

	// Create the client with all options
	return NewClient(allOptions...)
}

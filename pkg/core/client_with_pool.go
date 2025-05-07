// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// WithConnectionPool configures the client to use a connection pool
func WithConnectionPool(poolName string, config *transport.ConnectionPoolConfig) ClientOption {
	return func(c *Client) {
		// Get or create a connection pool
		pool := transport.GetServicePool(poolName, config)
		
		// Use the HTTP client from the pool
		c.HTTPClient = pool.GetClient()
	}
}

// WithPooledClient configures the client with a pre-existing HTTP client
// from a connection pool
func WithPooledClient(pool *transport.ConnectionPool) ClientOption {
	return func(c *Client) {
		c.HTTPClient = pool.GetClient()
	}
}

// NewHTTPClientWithConnectionPool creates a new HTTP client with a connection pool
func NewHTTPClientWithConnectionPool(poolName string, config *transport.ConnectionPoolConfig) *http.Client {
	pool := transport.GetServicePool(poolName, config)
	return pool.GetClient()
}

// EnableDefaultConnectionPool configures a default connection pool for all clients
// This should be called early in your application's initialization
func EnableDefaultConnectionPool() {
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
		// Use slightly different settings based on expected service usage patterns
		config := transport.DefaultConnectionPoolConfig()
		
		switch service {
		case "transfer":
			// Transfer service may have higher throughput needs
			config.MaxIdleConnsPerHost = 8
			config.MaxConnsPerHost = 16
		case "auth":
			// Auth service typically needs fewer connections
			config.MaxIdleConnsPerHost = 4
			config.MaxConnsPerHost = 8
		default:
			// Use defaults for other services
		}
		
		// Initialize the pool for the service
		transport.GetServicePool(service, config)
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
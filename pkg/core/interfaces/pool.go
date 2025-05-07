// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package interfaces

import (
	"net/http"
	"time"
)

// ConnectionPool defines the interface for a connection pool
type ConnectionPool interface {
	// GetClient returns the HTTP client that uses this connection pool
	GetClient() *http.Client

	// SetTimeout sets the overall timeout for requests
	SetTimeout(timeout time.Duration)

	// CloseIdleConnections closes all idle connections in the pool
	CloseIdleConnections()

	// GetTransport returns the underlying HTTP transport
	GetTransport() *http.Transport
}

// ConnectionPoolConfig defines the interface for connection pool configuration
type ConnectionPoolConfig interface {
	// GetMaxIdleConnsPerHost returns the maximum number of idle connections to keep per host
	GetMaxIdleConnsPerHost() int

	// GetMaxIdleConns returns the maximum number of idle connections across all hosts
	GetMaxIdleConns() int

	// GetMaxConnsPerHost returns the maximum number of connections per host
	GetMaxConnsPerHost() int

	// GetIdleConnTimeout returns how long an idle connection will remain idle before being closed
	GetIdleConnTimeout() time.Duration
}

// ConnectionPoolManager defines the interface for a connection pool manager
type ConnectionPoolManager interface {
	// GetPool returns the connection pool for the given service, creating one if it doesn't exist
	GetPool(serviceName string, config ConnectionPoolConfig) ConnectionPool

	// CloseAllIdleConnections closes idle connections across all managed pools
	CloseAllIdleConnections()

	// GetAllStats returns stats for all managed connection pools
	GetAllStats() map[string]interface{}
}

// PooledHTTPClient defines the interface for an HTTP client that uses a connection pool
type PooledHTTPClient interface {
	// GetPool returns the connection pool used by this client
	GetPool() ConnectionPool

	// Do performs an HTTP request using the pooled client
	Do(req *http.Request) (*http.Response, error)
}

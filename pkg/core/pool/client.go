// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pool

import (
	"net/http"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// Client defines an HTTP client that uses a connection pool
type Client struct {
	// Client is the underlying HTTP client
	Client *http.Client

	// Pool is the connection pool this client uses
	Pool interfaces.ConnectionPool
}

// NewClient creates a new pooled client for the given service
func NewClient(serviceName string, config interfaces.ConnectionPoolConfig) *Client {
	pool := GetServicePool(serviceName, config)
	return &Client{
		Client: pool.GetClient(),
		Pool:   pool,
	}
}

// GetConnectionPool returns the connection pool used by this client
func (c *Client) GetConnectionPool() interfaces.ConnectionPool {
	return c.Pool
}

// SetTimeout sets the timeout for all requests made by this client
func (c *Client) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
}

// GetClient returns the underlying HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.Client
}

// CloseIdleConnections closes all idle connections in the connection pool
func (c *Client) CloseIdleConnections() {
	if c.Pool != nil {
		c.Pool.CloseIdleConnections()
	}
}
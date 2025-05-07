// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// ConnectionPoolConfig contains configuration options for the connection pool
type ConnectionPoolConfig struct {
	// MaxIdleConnsPerHost is the maximum number of idle connections to keep per host
	MaxIdleConnsPerHost int

	// MaxIdleConns is the maximum number of idle connections across all hosts
	MaxIdleConns int

	// MaxConnsPerHost limits the total number of connections per host
	MaxConnsPerHost int

	// IdleConnTimeout is how long an idle connection will remain idle before being closed
	IdleConnTimeout time.Duration

	// DisableKeepAlives disables HTTP keep-alives and will only use connections for a single request
	DisableKeepAlives bool

	// ResponseHeaderTimeout is the amount of time to wait for a server's response headers
	ResponseHeaderTimeout time.Duration

	// ExpectContinueTimeout is the amount of time to wait for a server's first response headers
	// after fully writing the request headers if the request has an "Expect: 100-continue" header
	ExpectContinueTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time waiting to wait for a TLS handshake
	TLSHandshakeTimeout time.Duration

	// TLSClientConfig specifies the TLS configuration to use with TLS connections
	TLSClientConfig *tls.Config
}

// DefaultConnectionPoolConfig returns a default configuration for the connection pool
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	cpuCount := runtime.NumCPU()
	return &ConnectionPoolConfig{
		MaxIdleConnsPerHost:   cpuCount * 2,
		MaxIdleConns:          100,
		MaxConnsPerHost:       cpuCount * 4,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		TLSClientConfig:       nil,
	}
}

// GetMaxIdleConnsPerHost returns the maximum number of idle connections to keep per host
func (c *ConnectionPoolConfig) GetMaxIdleConnsPerHost() int {
	return c.MaxIdleConnsPerHost
}

// GetMaxIdleConns returns the maximum number of idle connections across all hosts
func (c *ConnectionPoolConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

// GetMaxConnsPerHost returns the maximum number of connections per host
func (c *ConnectionPoolConfig) GetMaxConnsPerHost() int {
	return c.MaxConnsPerHost
}

// GetIdleConnTimeout returns how long an idle connection will remain idle before being closed
func (c *ConnectionPoolConfig) GetIdleConnTimeout() time.Duration {
	return c.IdleConnTimeout
}

// ConnectionPool manages a pool of HTTP connections to improve performance
type ConnectionPool struct {
	// Transport is the underlying HTTP transport
	Transport *http.Transport

	// Config contains the configuration for the connection pool
	Config *ConnectionPoolConfig

	// Client is the HTTP client that uses this connection pool
	Client *http.Client

	// mu protects shared data structures
	mu sync.Mutex

	// activeRequests tracks the number of active requests by host
	activeRequests map[string]int
}

// NewConnectionPool creates a new connection pool with the given configuration
func NewConnectionPool(config *ConnectionPoolConfig) *ConnectionPool {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}

	// Create the transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		DisableKeepAlives:     config.DisableKeepAlives,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ForceAttemptHTTP2:     true,
	}

	// Apply custom TLS config if provided
	if config.TLSClientConfig != nil {
		transport.TLSClientConfig = config.TLSClientConfig
	}

	// Create the client
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &ConnectionPool{
		Transport:      transport,
		Config:         config,
		Client:         client,
		activeRequests: make(map[string]int),
	}
}

// GetClient returns the HTTP client that uses this connection pool
func (p *ConnectionPool) GetClient() *http.Client {
	return p.Client
}

// SetTimeout sets the overall timeout for requests
func (p *ConnectionPool) SetTimeout(timeout time.Duration) {
	p.Client.Timeout = timeout
}

// CloseIdleConnections closes all idle connections in the pool
func (p *ConnectionPool) CloseIdleConnections() {
	p.Transport.CloseIdleConnections()
}

// GetTransport returns the underlying HTTP transport
func (p *ConnectionPool) GetTransport() *http.Transport {
	return p.Transport
}

// GetStats returns statistics about the connection pool
func (p *ConnectionPool) GetStats() ConnectionPoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	activeHosts := 0
	totalActive := 0
	activeByHost := make(map[string]int)

	for host, count := range p.activeRequests {
		if count > 0 {
			activeHosts++
			totalActive += count
			activeByHost[host] = count
		}
	}

	return ConnectionPoolStats{
		ActiveHosts:  activeHosts,
		TotalActive:  totalActive,
		ActiveByHost: activeByHost,
		Config:       *p.Config,
	}
}

// ConnectionPoolStats contains statistics about the connection pool
type ConnectionPoolStats struct {
	// ActiveHosts is the number of hosts with active connections
	ActiveHosts int

	// TotalActive is the total number of active connections
	TotalActive int

	// ActiveByHost shows the number of active connections per host
	ActiveByHost map[string]int

	// Config is the current configuration
	Config ConnectionPoolConfig
}

// ConnectionPoolManager provides a global connection pool manager
type ConnectionPoolManager struct {
	// pools maps service names to connection pools
	pools map[string]*ConnectionPool

	// defaultConfig is the default configuration to use for new pools
	defaultConfig *ConnectionPoolConfig

	// mu protects the pools map
	mu sync.Mutex
}

// NewConnectionPoolManager creates a new connection pool manager
func NewConnectionPoolManager(defaultConfig *ConnectionPoolConfig) *ConnectionPoolManager {
	if defaultConfig == nil {
		defaultConfig = DefaultConnectionPoolConfig()
	}

	return &ConnectionPoolManager{
		pools:         make(map[string]*ConnectionPool),
		defaultConfig: defaultConfig,
	}
}

// GetPool returns the connection pool for the given service, creating one if it doesn't exist
func (m *ConnectionPoolManager) GetPool(serviceName string, config *ConnectionPoolConfig) *ConnectionPool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we already have a pool for this service
	if pool, ok := m.pools[serviceName]; ok {
		return pool
	}

	// Use provided config or default
	if config == nil {
		config = m.defaultConfig
	}

	// Create a new pool
	pool := NewConnectionPool(config)
	m.pools[serviceName] = pool
	return pool
}

// CloseAllIdleConnections closes idle connections across all managed pools
func (m *ConnectionPoolManager) CloseAllIdleConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pool := range m.pools {
		pool.CloseIdleConnections()
	}
}

// GetAllStats returns stats for all managed connection pools
func (m *ConnectionPoolManager) GetAllStats() map[string]ConnectionPoolStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := make(map[string]ConnectionPoolStats)
	for name, pool := range m.pools {
		stats[name] = pool.GetStats()
	}
	return stats
}

// GlobalHttpPoolManager is the default global connection pool manager
var GlobalHttpPoolManager = NewConnectionPoolManager(nil)

// GetServicePool is a convenience function to get a connection pool from the global manager
func GetServicePool(serviceName string, config *ConnectionPoolConfig) *ConnectionPool {
	return GlobalHttpPoolManager.GetPool(serviceName, config)
}

// GetHTTPClientForService returns an HTTP client for the given service
func GetHTTPClientForService(serviceName string, config *ConnectionPoolConfig) *http.Client {
	pool := GetServicePool(serviceName, config)
	return pool.GetClient()
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pool

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// PoolManager implements the interfaces.ConnectionPoolManager interface
type PoolManager struct {
	// pools maps service names to connection pools
	pools map[string]*Pool

	// defaultConfig is the default configuration to use for new pools
	defaultConfig *Config

	// mu protects the pools map
	mu sync.Mutex
}

// NewPoolManager creates a new connection pool manager
func NewPoolManager(defaultConfig *Config) *PoolManager {
	if defaultConfig == nil {
		defaultConfig = DefaultConfig()
	}

	return &PoolManager{
		pools:         make(map[string]*Pool),
		defaultConfig: defaultConfig,
	}
}

// GetPool returns the connection pool for the given service
func (m *PoolManager) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we already have a pool for this service
	if pool, ok := m.pools[serviceName]; ok {
		return pool
	}

	// Convert config to our internal type if needed
	var poolConfig *Config
	if config == nil {
		poolConfig = m.defaultConfig
	} else if pc, ok := config.(*Config); ok {
		poolConfig = pc
	} else {
		// Create a new config from the interface
		poolConfig = &Config{
			MaxIdleConnsPerHost: config.GetMaxIdleConnsPerHost(),
			MaxIdleConns:        config.GetMaxIdleConns(),
			MaxConnsPerHost:     config.GetMaxConnsPerHost(),
			IdleConnTimeout:     config.GetIdleConnTimeout(),
		}
	}

	// Create a new pool
	pool := NewPool(poolConfig)
	m.pools[serviceName] = pool
	return pool
}

// CloseAllIdleConnections closes idle connections across all managed pools
func (m *PoolManager) CloseAllIdleConnections() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, pool := range m.pools {
		pool.CloseIdleConnections()
	}
}

// GetAllStats returns stats for all managed connection pools
func (m *PoolManager) GetAllStats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := make(map[string]interface{})
	for name, pool := range m.pools {
		stats[name] = pool.GetStats()
	}
	return stats
}

// Pool implements the interfaces.ConnectionPool interface
type Pool struct {
	// Transport is the underlying HTTP transport
	Transport *http.Transport

	// Config contains the configuration for the connection pool
	Config *Config

	// Client is the HTTP client that uses this connection pool
	Client *http.Client

	// mu protects shared data structures
	mu sync.Mutex

	// activeRequests tracks the number of active requests by host
	activeRequests map[string]int
}

// NewPool creates a new connection pool with the given configuration
func NewPool(config *Config) *Pool {
	if config == nil {
		config = DefaultConfig()
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
	if config.TLSHandshakeTimeout > 0 {
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	// Create the client
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Pool{
		Transport:      transport,
		Config:         config,
		Client:         client,
		activeRequests: make(map[string]int),
	}
}

// GetClient returns the HTTP client that uses this connection pool
func (p *Pool) GetClient() *http.Client {
	return p.Client
}

// SetTimeout sets the overall timeout for requests
func (p *Pool) SetTimeout(timeout time.Duration) {
	p.Client.Timeout = timeout
}

// CloseIdleConnections closes all idle connections in the pool
func (p *Pool) CloseIdleConnections() {
	p.Transport.CloseIdleConnections()
}

// GetTransport returns the underlying HTTP transport
func (p *Pool) GetTransport() *http.Transport {
	return p.Transport
}

// GetStats returns statistics about the connection pool
type PoolStats struct {
	// ActiveHosts is the number of hosts with active connections
	ActiveHosts int `json:"active_hosts"`

	// TotalActive is the total number of active connections
	TotalActive int `json:"total_active"`

	// ActiveByHost shows the number of active connections per host
	ActiveByHost map[string]int `json:"active_by_host"`

	// Config is the current configuration
	Config Config `json:"config"`
}

// GetStats returns statistics about the connection pool
func (p *Pool) GetStats() interface{} {
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

	return PoolStats{
		ActiveHosts:  activeHosts,
		TotalActive:  totalActive,
		ActiveByHost: activeByHost,
		Config:       *p.Config,
	}
}

// GlobalPoolManager is the default global connection pool manager
var GlobalPoolManager = NewPoolManager(nil)

// GetServicePool is a convenience function to get a connection pool from the global manager
func GetServicePool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	return GlobalPoolManager.GetPool(serviceName, config)
}
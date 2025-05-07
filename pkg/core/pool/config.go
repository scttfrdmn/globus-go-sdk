// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package pool

import (
	"runtime"
	"time"
)

// Config implements the interfaces.ConnectionPoolConfig interface
type Config struct {
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
}

// DefaultConfig returns a default configuration for the connection pool
func DefaultConfig() *Config {
	cpuCount := runtime.NumCPU()
	return &Config{
		MaxIdleConnsPerHost:   cpuCount * 2,
		MaxIdleConns:          100,
		MaxConnsPerHost:       cpuCount * 4,
		IdleConnTimeout:       90 * time.Second,
		DisableKeepAlives:     false,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
	}
}

// GetMaxIdleConnsPerHost returns the maximum number of idle connections to keep per host
func (c *Config) GetMaxIdleConnsPerHost() int {
	return c.MaxIdleConnsPerHost
}

// GetMaxIdleConns returns the maximum number of idle connections across all hosts
func (c *Config) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

// GetMaxConnsPerHost returns the maximum number of connections per host
func (c *Config) GetMaxConnsPerHost() int {
	return c.MaxConnsPerHost
}

// GetIdleConnTimeout returns how long an idle connection will remain idle before being closed
func (c *Config) GetIdleConnTimeout() time.Duration {
	return c.IdleConnTimeout
}

// ForService returns a configuration optimized for a specific service
func ForService(service string) *Config {
	config := DefaultConfig()
	
	switch service {
	case "transfer":
		// Transfer service may have higher throughput needs
		config.MaxIdleConnsPerHost = 8
		config.MaxConnsPerHost = 16
		config.IdleConnTimeout = 120 * time.Second
	case "auth":
		// Auth service typically needs fewer connections
		config.MaxIdleConnsPerHost = 4
		config.MaxConnsPerHost = 8
		config.IdleConnTimeout = 60 * time.Second
	case "compute":
		// Compute service has similar needs to Transfer
		config.MaxIdleConnsPerHost = 6
		config.MaxConnsPerHost = 16
		config.IdleConnTimeout = 120 * time.Second
	case "search":
	case "flows":
		// Medium usage services
		config.MaxIdleConnsPerHost = 6
		config.MaxConnsPerHost = 12
		config.IdleConnTimeout = 90 * time.Second
	case "groups":
	case "timers":
		// Low usage services
		config.MaxIdleConnsPerHost = 4
		config.MaxConnsPerHost = 8
		config.IdleConnTimeout = 60 * time.Second
	}
	
	return config
}
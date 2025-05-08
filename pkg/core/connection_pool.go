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

// GetConnectionPool returns a connection pool for the given service
func GetConnectionPool(service string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	if globalPoolManager == nil {
		return nil
	}
	return globalPoolManager.GetPool(service, config)
}

// GetHTTPClientForService returns an HTTP client configured for a specific service
func GetHTTPClientForService(service string) *http.Client {
	if globalPoolManager == nil {
		return &http.Client{}
	}

	pool := globalPoolManager.GetPool(service, nil)
	if pool == nil {
		return &http.Client{}
	}

	return pool.GetClient()
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"net/http"
	"os"
	"sync"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// ConnectionPoolProvider defines the interface for a provider of connection pools
type ConnectionPoolProvider interface {
	// GetPool returns a connection pool for the given service
	GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool

	// CloseAllIdleConnections closes all idle connections across all pools
	CloseAllIdleConnections()

	// GetAllStats returns statistics for all managed pools
	GetAllStats() map[string]interface{}
}

// globalConnectionPoolManager holds the current connection pool manager
var globalConnectionPoolManager ConnectionPoolProvider
var connectionPoolMutex sync.Mutex
var defaultConnectionPoolEnabled bool

// GetConnectionPoolManager returns the current global connection pool manager
func GetConnectionPoolManager() ConnectionPoolProvider {
	connectionPoolMutex.Lock()
	defer connectionPoolMutex.Unlock()
	return globalConnectionPoolManager
}

// SetConnectionPoolManager sets the global connection pool manager
func SetConnectionPoolManager(manager interface{}) {
	connectionPoolMutex.Lock()
	defer connectionPoolMutex.Unlock()

	// Handle both ConnectionPoolProvider and interfaces.ConnectionPoolManager
	switch m := manager.(type) {
	case ConnectionPoolProvider:
		globalConnectionPoolManager = m
	case interfaces.ConnectionPoolManager:
		// Adapt interfaces.ConnectionPoolManager to ConnectionPoolProvider
		globalConnectionPoolManager = connectionPoolManagerAdapter{m}
	}
}

// connectionPoolManagerAdapter adapts interfaces.ConnectionPoolManager to ConnectionPoolProvider
type connectionPoolManagerAdapter struct {
	manager interfaces.ConnectionPoolManager
}

func (a connectionPoolManagerAdapter) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	return a.manager.GetPool(serviceName, config)
}

func (a connectionPoolManagerAdapter) CloseAllIdleConnections() {
	a.manager.CloseAllIdleConnections()
}

func (a connectionPoolManagerAdapter) GetAllStats() map[string]interface{} {
	return a.manager.GetAllStats()
}

// EnableDefaultConnectionPool enables the default connection pool
func EnableDefaultConnectionPool() {
	connectionPoolMutex.Lock()
	defer connectionPoolMutex.Unlock()
	defaultConnectionPoolEnabled = true
}

// DisableDefaultConnectionPool disables the default connection pool
func DisableDefaultConnectionPool() {
	connectionPoolMutex.Lock()
	defer connectionPoolMutex.Unlock()
	defaultConnectionPoolEnabled = false
}

// IsDefaultConnectionPoolEnabled returns whether the default connection pool is enabled
func IsDefaultConnectionPoolEnabled() bool {
	connectionPoolMutex.Lock()
	defer connectionPoolMutex.Unlock()
	return defaultConnectionPoolEnabled && os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL") != "true"
}

// GetHTTPClientForService returns an HTTP client for the given service
func GetHTTPClientForService(serviceName string) *http.Client {
	if IsDefaultConnectionPoolEnabled() && globalConnectionPoolManager != nil {
		pool := globalConnectionPoolManager.GetPool(serviceName, nil)
		return pool.GetClient()
	}

	// Fallback to a standard client
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			MaxIdleConns:        100,
			IdleConnTimeout:     90,
		},
	}
}

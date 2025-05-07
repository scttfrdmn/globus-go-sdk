// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package http

import "github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"

// HttpConnectionPool embeds the standard ConnectionPool and implements the interfaces.ConnectionPool interface
type HttpConnectionPool struct {
	*ConnectionPool
}

func NewHttpConnectionPool(config *ConnectionPoolConfig) *HttpConnectionPool {
	return &HttpConnectionPool{
		ConnectionPool: NewConnectionPool(config),
	}
}

// Ensure HttpConnectionPool implements the interfaces.ConnectionPool interface
var _ interfaces.ConnectionPool = (*HttpConnectionPool)(nil)

// HttpConnectionPoolManager embeds the standard ConnectionPoolManager and implements the interfaces.ConnectionPoolManager interface
type HttpConnectionPoolManager struct {
	*ConnectionPoolManager
}

func NewHttpConnectionPoolManager(defaultConfig *ConnectionPoolConfig) *HttpConnectionPoolManager {
	return &HttpConnectionPoolManager{
		ConnectionPoolManager: NewConnectionPoolManager(defaultConfig),
	}
}

// GetPool implements the interfaces.ConnectionPoolManager interface
func (m *HttpConnectionPoolManager) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	// Convert the interface to our specific config type
	var poolConfig *ConnectionPoolConfig
	if config != nil {
		if pc, ok := config.(*ConnectionPoolConfig); ok {
			poolConfig = pc
		} else {
			// Create a new config with the interface values
			poolConfig = &ConnectionPoolConfig{
				MaxIdleConnsPerHost: config.GetMaxIdleConnsPerHost(),
				MaxIdleConns:        config.GetMaxIdleConns(),
				MaxConnsPerHost:     config.GetMaxConnsPerHost(),
				IdleConnTimeout:     config.GetIdleConnTimeout(),
			}
		}
	}

	pool := m.ConnectionPoolManager.GetPool(serviceName, poolConfig)
	return &HttpConnectionPool{ConnectionPool: pool}
}

// CloseAllIdleConnections implements the interfaces.ConnectionPoolManager interface
func (m *HttpConnectionPoolManager) CloseAllIdleConnections() {
	m.ConnectionPoolManager.CloseAllIdleConnections()
}

// GetAllStats implements the interfaces.ConnectionPoolManager interface
func (m *HttpConnectionPoolManager) GetAllStats() map[string]interface{} {
	stats := m.ConnectionPoolManager.GetAllStats()
	result := make(map[string]interface{})
	for k, v := range stats {
		result[k] = v
	}
	return result
}

// Ensure HttpConnectionPoolManager implements the interfaces.ConnectionPoolManager interface
var _ interfaces.ConnectionPoolManager = (*HttpConnectionPoolManager)(nil)

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"

// TransportConnectionPool embeds the standard ConnectionPool and implements the interfaces.ConnectionPool interface
type TransportConnectionPool struct {
	*ConnectionPool
}

func NewTransportConnectionPool(config *ConnectionPoolConfig) *TransportConnectionPool {
	return &TransportConnectionPool{
		ConnectionPool: NewConnectionPool(config),
	}
}

// Ensure TransportConnectionPool implements the interfaces.ConnectionPool interface
var _ interfaces.ConnectionPool = (*TransportConnectionPool)(nil)

// TransportConnectionPoolManager embeds the standard ConnectionPoolManager and implements the interfaces.ConnectionPoolManager interface
type TransportConnectionPoolManager struct {
	*ConnectionPoolManager
}

func NewTransportConnectionPoolManager(defaultConfig *ConnectionPoolConfig) *TransportConnectionPoolManager {
	return &TransportConnectionPoolManager{
		ConnectionPoolManager: NewConnectionPoolManager(defaultConfig),
	}
}

// GetPool implements the interfaces.ConnectionPoolManager interface
func (m *TransportConnectionPoolManager) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
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
	return &TransportConnectionPool{ConnectionPool: pool}
}

// GetAllStats implements the interfaces.ConnectionPoolManager interface
func (m *TransportConnectionPoolManager) GetAllStats() map[string]interface{} {
	stats := m.ConnectionPoolManager.GetAllStats()
	result := make(map[string]interface{})
	for k, v := range stats {
		result[k] = v
	}
	return result
}

// Ensure TransportConnectionPoolManager implements the interfaces.ConnectionPoolManager interface
var _ interfaces.ConnectionPoolManager = (*TransportConnectionPoolManager)(nil)

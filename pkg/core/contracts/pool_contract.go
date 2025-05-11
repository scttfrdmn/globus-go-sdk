// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"net/http"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyConnectionPoolContract verifies that a ConnectionPool implementation
// satisfies the behavioral contract of the interface.
func VerifyConnectionPoolContract(t *testing.T, pool interfaces.ConnectionPool) {
	t.Helper()

	t.Run("GetClient method", func(t *testing.T) {
		verifyGetClientMethod(t, pool)
	})

	t.Run("SetTimeout method", func(t *testing.T) {
		verifySetTimeoutMethod(t, pool)
	})

	t.Run("CloseIdleConnections method", func(t *testing.T) {
		verifyCloseIdleConnectionsMethod(t, pool)
	})

	t.Run("GetTransport method", func(t *testing.T) {
		verifyGetTransportMethod(t, pool)
	})
}

// verifyGetClientMethod tests the behavior of the GetClient method
func verifyGetClientMethod(t *testing.T, pool interfaces.ConnectionPool) {
	t.Helper()

	// GetClient should return a non-nil HTTP client
	client := pool.GetClient()
	if client == nil {
		t.Error("GetClient returned nil")
	}

	// Subsequent calls should return the same HTTP client instance
	client2 := pool.GetClient()
	if client != client2 {
		t.Error("GetClient returned different instances on consecutive calls")
	}
}

// verifySetTimeoutMethod tests the behavior of the SetTimeout method
func verifySetTimeoutMethod(t *testing.T, pool interfaces.ConnectionPool) {
	t.Helper()

	// Get the original client to compare timeout changes
	originalClient := pool.GetClient()
	originalTimeout := originalClient.Timeout

	// Set a new timeout
	newTimeout := originalTimeout + 10*time.Second
	pool.SetTimeout(newTimeout)

	// The client should have the new timeout
	updatedClient := pool.GetClient()
	if updatedClient.Timeout != newTimeout {
		t.Errorf("SetTimeout did not update client timeout. Expected %v, got %v",
			newTimeout, updatedClient.Timeout)
	}

	// Restore the original timeout
	pool.SetTimeout(originalTimeout)
}

// verifyCloseIdleConnectionsMethod tests the behavior of the CloseIdleConnections method
func verifyCloseIdleConnectionsMethod(t *testing.T, pool interfaces.ConnectionPool) {
	t.Helper()

	// This method doesn't have a testable result, but we can verify it doesn't panic
	pool.CloseIdleConnections()
}

// verifyGetTransportMethod tests the behavior of the GetTransport method
func verifyGetTransportMethod(t *testing.T, pool interfaces.ConnectionPool) {
	t.Helper()

	// GetTransport should return a non-nil transport
	transport := pool.GetTransport()
	if transport == nil {
		t.Error("GetTransport returned nil")
	}

	// Subsequent calls should return the same transport instance
	transport2 := pool.GetTransport()
	if transport != transport2 {
		t.Error("GetTransport returned different instances on consecutive calls")
	}
}

// VerifyConnectionPoolConfigContract verifies that a ConnectionPoolConfig implementation
// satisfies the behavioral contract of the interface.
func VerifyConnectionPoolConfigContract(t *testing.T, config interfaces.ConnectionPoolConfig) {
	t.Helper()

	t.Run("GetMaxIdleConnsPerHost method", func(t *testing.T) {
		verifyGetMaxIdleConnsPerHostMethod(t, config)
	})

	t.Run("GetMaxIdleConns method", func(t *testing.T) {
		verifyGetMaxIdleConnsMethod(t, config)
	})

	t.Run("GetMaxConnsPerHost method", func(t *testing.T) {
		verifyGetMaxConnsPerHostMethod(t, config)
	})

	t.Run("GetIdleConnTimeout method", func(t *testing.T) {
		verifyGetIdleConnTimeoutMethod(t, config)
	})
}

// verifyGetMaxIdleConnsPerHostMethod tests the behavior of the GetMaxIdleConnsPerHost method
func verifyGetMaxIdleConnsPerHostMethod(t *testing.T, config interfaces.ConnectionPoolConfig) {
	t.Helper()

	// GetMaxIdleConnsPerHost should return a non-negative value
	maxIdleConnsPerHost := config.GetMaxIdleConnsPerHost()
	if maxIdleConnsPerHost < 0 {
		t.Errorf("GetMaxIdleConnsPerHost returned negative value: %d", maxIdleConnsPerHost)
	}

	// Subsequent calls should return the same value
	maxIdleConnsPerHost2 := config.GetMaxIdleConnsPerHost()
	if maxIdleConnsPerHost != maxIdleConnsPerHost2 {
		t.Errorf("GetMaxIdleConnsPerHost returned different values: %d and %d",
			maxIdleConnsPerHost, maxIdleConnsPerHost2)
	}
}

// verifyGetMaxIdleConnsMethod tests the behavior of the GetMaxIdleConns method
func verifyGetMaxIdleConnsMethod(t *testing.T, config interfaces.ConnectionPoolConfig) {
	t.Helper()

	// GetMaxIdleConns should return a non-negative value
	maxIdleConns := config.GetMaxIdleConns()
	if maxIdleConns < 0 {
		t.Errorf("GetMaxIdleConns returned negative value: %d", maxIdleConns)
	}

	// Subsequent calls should return the same value
	maxIdleConns2 := config.GetMaxIdleConns()
	if maxIdleConns != maxIdleConns2 {
		t.Errorf("GetMaxIdleConns returned different values: %d and %d",
			maxIdleConns, maxIdleConns2)
	}
}

// verifyGetMaxConnsPerHostMethod tests the behavior of the GetMaxConnsPerHost method
func verifyGetMaxConnsPerHostMethod(t *testing.T, config interfaces.ConnectionPoolConfig) {
	t.Helper()

	// GetMaxConnsPerHost should return a non-negative value
	maxConnsPerHost := config.GetMaxConnsPerHost()
	if maxConnsPerHost < 0 {
		t.Errorf("GetMaxConnsPerHost returned negative value: %d", maxConnsPerHost)
	}

	// Subsequent calls should return the same value
	maxConnsPerHost2 := config.GetMaxConnsPerHost()
	if maxConnsPerHost != maxConnsPerHost2 {
		t.Errorf("GetMaxConnsPerHost returned different values: %d and %d",
			maxConnsPerHost, maxConnsPerHost2)
	}
}

// verifyGetIdleConnTimeoutMethod tests the behavior of the GetIdleConnTimeout method
func verifyGetIdleConnTimeoutMethod(t *testing.T, config interfaces.ConnectionPoolConfig) {
	t.Helper()

	// GetIdleConnTimeout should return a non-negative value
	idleConnTimeout := config.GetIdleConnTimeout()
	if idleConnTimeout < 0 {
		t.Errorf("GetIdleConnTimeout returned negative value: %v", idleConnTimeout)
	}

	// Subsequent calls should return the same value
	idleConnTimeout2 := config.GetIdleConnTimeout()
	if idleConnTimeout != idleConnTimeout2 {
		t.Errorf("GetIdleConnTimeout returned different values: %v and %v",
			idleConnTimeout, idleConnTimeout2)
	}
}

// VerifyConnectionPoolManagerContract verifies that a ConnectionPoolManager implementation
// satisfies the behavioral contract of the interface.
//
// The configFactory parameter can be either:
// - func() interfaces.ConnectionPoolConfig: A function that returns a pool config
// - func() interface{}: A function that returns any type that implements interfaces.ConnectionPoolConfig
func VerifyConnectionPoolManagerContract(t *testing.T, manager interfaces.ConnectionPoolManager,
	configFactoryOrInterface interface{}) {

	t.Helper()

	if configFactoryOrInterface == nil {
		t.Fatal("config factory is required for testing ConnectionPoolManager")
	}

	// Handle different types of configFactory
	var configFactory func() interfaces.ConnectionPoolConfig

	// Try direct type assertion first
	if typedFactory, ok := configFactoryOrInterface.(func() interfaces.ConnectionPoolConfig); ok {
		configFactory = typedFactory
	} else if genericFactory, ok := configFactoryOrInterface.(func() interface{}); ok {
		// Wrap the generic factory to return interfaces.ConnectionPoolConfig
		configFactory = func() interfaces.ConnectionPoolConfig {
			result := genericFactory()
			// Try to convert the result to interfaces.ConnectionPoolConfig
			if config, ok := result.(interfaces.ConnectionPoolConfig); ok {
				return config
			}
			t.Fatalf("config factory returned %T which does not implement interfaces.ConnectionPoolConfig", result)
			return nil // never reached due to t.Fatalf above
		}
	} else {
		t.Fatalf("config factory must be either func() interfaces.ConnectionPoolConfig or func() interface{}, got %T", configFactoryOrInterface)
	}

	t.Run("GetPool method", func(t *testing.T) {
		verifyGetPoolMethod(t, manager, configFactory)
	})

	t.Run("CloseAllIdleConnections method", func(t *testing.T) {
		verifyCloseAllIdleConnectionsMethod(t, manager)
	})

	t.Run("GetAllStats method", func(t *testing.T) {
		verifyGetAllStatsMethod(t, manager)
	})
}

// verifyGetPoolMethod tests the behavior of the GetPool method
func verifyGetPoolMethod(t *testing.T, manager interfaces.ConnectionPoolManager,
	configFactory func() interfaces.ConnectionPoolConfig) {

	t.Helper()

	// Create a config for testing
	config := configFactory()

	// GetPool should return a non-nil pool
	pool := manager.GetPool("test-service", config)
	if pool == nil {
		t.Error("GetPool returned nil")
	}

	// Getting the same service should return the same pool instance
	pool2 := manager.GetPool("test-service", config)
	if pool != pool2 {
		t.Error("GetPool returned different instances for the same service")
	}

	// Getting a different service should return a different pool instance
	pool3 := manager.GetPool("test-service-2", config)
	if pool == pool3 {
		t.Error("GetPool returned the same instance for different services")
	}
}

// verifyCloseAllIdleConnectionsMethod tests the behavior of the CloseAllIdleConnections method
func verifyCloseAllIdleConnectionsMethod(t *testing.T, manager interfaces.ConnectionPoolManager) {
	t.Helper()

	// This method doesn't have a testable result, but we can verify it doesn't panic
	manager.CloseAllIdleConnections()
}

// verifyGetAllStatsMethod tests the behavior of the GetAllStats method
func verifyGetAllStatsMethod(t *testing.T, manager interfaces.ConnectionPoolManager) {
	t.Helper()

	// GetAllStats should return a non-nil map
	stats := manager.GetAllStats()
	if stats == nil {
		t.Error("GetAllStats returned nil map")
	}

	// The map should be safe to iterate
	for key, value := range stats {
		// Just ensure we can access the data
		if key == "" && value == nil {
			// This condition should never happen, but we need to use the variables
			t.Error("Empty key and nil value in stats")
		}
	}
}

// VerifyPooledHTTPClientContract verifies that a PooledHTTPClient implementation
// satisfies the behavioral contract of the interface.
func VerifyPooledHTTPClientContract(t *testing.T, client interfaces.PooledHTTPClient) {
	t.Helper()

	t.Run("GetPool method", func(t *testing.T) {
		verifyGetPooledClientPoolMethod(t, client)
	})

	t.Run("Do method", func(t *testing.T) {
		verifyPooledClientDoMethod(t, client)
	})
}

// verifyGetPooledClientPoolMethod tests the behavior of the GetPool method
func verifyGetPooledClientPoolMethod(t *testing.T, client interfaces.PooledHTTPClient) {
	t.Helper()

	// GetPool should return a non-nil pool
	pool := client.GetPool()
	if pool == nil {
		t.Error("GetPool returned nil")
	}

	// Subsequent calls should return the same pool instance
	pool2 := client.GetPool()
	if pool != pool2 {
		t.Error("GetPool returned different instances on consecutive calls")
	}
}

// verifyPooledClientDoMethod tests the behavior of the Do method
func verifyPooledClientDoMethod(t *testing.T, client interfaces.PooledHTTPClient) {
	t.Helper()

	// Create a test request
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test with valid request
	resp, err := client.Do(req)

	// We can't assume the request will succeed since we don't control the URL
	// Instead, we verify that the method doesn't panic and returns a response or error
	if err != nil {
		t.Logf("Request failed with error: %v (this may be expected)", err)
	} else if resp == nil {
		t.Error("Request succeeded but returned nil response")
	} else {
		// Clean up
		resp.Body.Close()
	}

	// Test with nil request (should return error)
	_, err = client.Do(nil)
	if err == nil {
		t.Error("Do with nil request should return error")
	}
}

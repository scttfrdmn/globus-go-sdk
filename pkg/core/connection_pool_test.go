// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
)

// TestMissingConnectionPoolFunctions tests all the functions that were reported
// as missing in issue #13 to ensure they exist and work correctly
func TestMissingConnectionPoolFunctions(t *testing.T) {
	// Create a pool manager to use in tests
	manager := pool.NewPoolManager(nil)

	// Test SetConnectionPoolManager exists and can be called
	t.Run("SetConnectionPoolManager", func(t *testing.T) {
		// This should not panic
		core.SetConnectionPoolManager(manager)
	})

	// Test EnableDefaultConnectionPool exists and can be called
	t.Run("EnableDefaultConnectionPool", func(t *testing.T) {
		// This should not panic
		core.EnableDefaultConnectionPool()
	})

	// Test GetConnectionPool exists and returns expected values
	t.Run("GetConnectionPool", func(t *testing.T) {
		pool := core.GetConnectionPool("test-service", nil)
		if pool == nil {
			t.Fatal("Expected non-nil pool but got nil")
		}
	})

	// Test GetHTTPClientForService exists and returns expected values
	t.Run("GetHTTPClientForService", func(t *testing.T) {
		client := core.GetHTTPClientForService("test-service")
		if client == nil {
			t.Fatal("Expected non-nil HTTP client but got nil")
		}
	})
}

// TestConnectionPoolIntegration tests that all the connection pool functions
// work together correctly to ensure the fix is comprehensive
func TestConnectionPoolIntegration(t *testing.T) {
	// First reset any global state
	core.SetConnectionPoolManager(nil)

	// Create a test manager
	manager := pool.NewPoolManager(nil)
	core.SetConnectionPoolManager(manager)

	// Test that we can enable the default connection pool
	core.EnableDefaultConnectionPool()

	// Verify we can get connection pools for various services
	services := []string{
		"auth",
		"transfer",
		"search",
		"flows",
		"groups",
		"compute",
		"timers",
	}

	for _, service := range services {
		// Get a connection pool
		pool := core.GetConnectionPool(service, nil)
		if pool == nil {
			t.Fatalf("Failed to get pool for service %s", service)
		}

		// Verify we can get an HTTP client
		client := pool.GetClient()
		if client == nil {
			t.Fatalf("Failed to get HTTP client for service %s", service)
		}

		// Get an HTTP client directly
		directClient := core.GetHTTPClientForService(service)
		if directClient == nil {
			t.Fatalf("Failed to get HTTP client directly for service %s", service)
		}
	}
}

// TestVerifyReleaseContainsRequiredFunctions verifies the functions are exported
// and accessible, similar to how downstream projects would use them
func TestVerifyReleaseContainsRequiredFunctions(t *testing.T) {
	// Expected function signatures
	var _ func(interfaces.ConnectionPoolManager) = core.SetConnectionPoolManager
	var _ func() = core.EnableDefaultConnectionPool
	var _ func(string, interfaces.ConnectionPoolConfig) interfaces.ConnectionPool = core.GetConnectionPool
	var _ func(string) *http.Client = core.GetHTTPClientForService

	// This test passes just by compiling - it verifies the function signatures
	// match what we expect and are exported from the package
}

// mockPoolManager implements interfaces.ConnectionPoolManager for testing
type mockPoolManager struct{}

func (m *mockPoolManager) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	return &mockPool{}
}

func (m *mockPoolManager) CloseAllIdleConnections() {}

func (m *mockPoolManager) GetAllStats() map[string]interface{} {
	return nil
}

// mockPool implements interfaces.ConnectionPool for testing
type mockPool struct{}

func (p *mockPool) GetClient() *http.Client {
	return &http.Client{}
}

func (p *mockPool) SetTimeout(timeout time.Duration) {}

func (p *mockPool) CloseIdleConnections() {}

func (p *mockPool) GetTransport() *http.Transport {
	return &http.Transport{}
}

func (p *mockPool) GetStats() interface{} {
	return nil
}

// TestWithMockImplementations tests the connection pool functions with mock implementations
// This is similar to how a downstream project might implement and use these interfaces
func TestWithMockImplementations(t *testing.T) {
	// Reset global state
	core.SetConnectionPoolManager(nil)
	
	// Create and set a mock pool manager
	mockManager := &mockPoolManager{}
	core.SetConnectionPoolManager(mockManager)
	
	// Test EnableDefaultConnectionPool with mock
	core.EnableDefaultConnectionPool()
	
	// Test GetConnectionPool with mock
	pool := core.GetConnectionPool("test-service", nil)
	if pool == nil {
		t.Fatal("Expected non-nil pool from mock but got nil")
	}
	
	// Test GetHTTPClientForService with mock
	client := core.GetHTTPClientForService("test-service")
	if client == nil {
		t.Fatal("Expected non-nil HTTP client from mock but got nil")
	}
}
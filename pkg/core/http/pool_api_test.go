// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package http_test

import (
	"reflect"
	"testing"

	httppool "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// TestHTTPPoolFunctionAvailability verifies that all required HTTP pool functions are
// properly defined and exported. This test ensures that dependent packages can access
// these functions and that no regressions occur in future updates.
func TestHTTPPoolFunctionAvailability(t *testing.T) {
	// Test GetServicePool - this is the critical function that was missing in issue #11
	if reflect.ValueOf(httppool.GetServicePool).IsNil() {
		t.Error("GetServicePool function is nil")
	}

	testFunc := func(name string, fn interface{}) {
		if fn == nil {
			t.Errorf("Function %s is nil", name)
		}
		if reflect.ValueOf(fn).IsNil() {
			t.Errorf("Function %s is defined but nil", name)
		}
	}

	// Test all essential functions used by the SDK
	testFunc("GetServicePool", httppool.GetServicePool)
	testFunc("GetHTTPClientForService", httppool.GetHTTPClientForService)
	testFunc("NewConnectionPool", httppool.NewConnectionPool)
	testFunc("NewConnectionPoolManager", httppool.NewConnectionPoolManager)
	testFunc("DefaultConnectionPoolConfig", httppool.DefaultConnectionPoolConfig)
	testFunc("NewHttpConnectionPool", httppool.NewHttpConnectionPool)
	testFunc("NewHttpConnectionPoolManager", httppool.NewHttpConnectionPoolManager)
}

// TestHTTPPoolImplementsInterfaces verifies that our pool types correctly implement
// the interfaces defined in the interfaces package
func TestHTTPPoolImplementsInterfaces(t *testing.T) {
	// Static type assertions (compile-time checks)
	var _ interfaces.ConnectionPool = (*httppool.HttpConnectionPool)(nil)
	var _ interfaces.ConnectionPoolManager = (*httppool.HttpConnectionPoolManager)(nil)

	// Runtime type checks
	pool := httppool.NewHttpConnectionPool(nil)
	if _, ok := interface{}(pool).(interfaces.ConnectionPool); !ok {
		t.Error("HttpConnectionPool does not implement interfaces.ConnectionPool")
	}

	manager := httppool.NewHttpConnectionPoolManager(nil)
	if _, ok := interface{}(manager).(interfaces.ConnectionPoolManager); !ok {
		t.Error("HttpConnectionPoolManager does not implement interfaces.ConnectionPoolManager")
	}
}

// TestGlobalHTTPPoolManagerAvailability verifies that the global HTTP pool manager
// is properly initialized
func TestGlobalHTTPPoolManagerAvailability(t *testing.T) {
	if httppool.GlobalHttpPoolManager == nil {
		t.Error("GlobalHttpPoolManager is nil")
	}

	// Test basic functionality to ensure it's properly initialized
	pool := httppool.GlobalHttpPoolManager.GetPool("test", nil)
	if pool == nil {
		t.Error("GlobalHttpPoolManager.GetPool returned nil")
	}
}

// TestHTTPPoolUsage verifies the full connection pool initialization flow that's
// used in the Globus SDK
func TestHTTPPoolUsage(t *testing.T) {
	// This test verifies the exact usage pattern found in globus.go
	// to make sure any future changes don't break compatibility

	serviceConfig := &httppool.ConnectionPoolConfig{
		MaxIdleConnsPerHost: 8,
		MaxConnsPerHost:     16,
		IdleConnTimeout:     90,
	}

	// Get a pool - this is called in initializeConnectionPools()
	pool := httppool.GetServicePool("test-service", serviceConfig)
	if pool == nil {
		t.Fatal("GetServicePool returned nil")
	}

	// Get an HTTP client - this is called in the service client constructors
	client := httppool.GetHTTPClientForService("test-service", serviceConfig)
	if client == nil {
		t.Fatal("GetHTTPClientForService returned nil")
	}

	// Verify the client has a transport from our pool
	if client.Transport == nil {
		t.Error("Client's Transport is nil")
	}
}

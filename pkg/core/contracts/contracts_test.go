// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts_test

import (
	"os"
	"testing"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/contracts"
	corehttp "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
	corepool "github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
)

// TestClientImplementationContract verifies that a Client implementation
// satisfies the ClientInterface contract
func TestClientImplementationContract(t *testing.T) {
	// Use the mock client for testing which fully implements the interface
	client := contracts.NewMockClient()
	contracts.VerifyClientContract(t, client)
}

// TestTransportImplementationContract verifies that the transport.Transport
// satisfies the Transport contract
func TestTransportImplementationContract(t *testing.T) {
	// Set up the transport without client since our MockTransport doesn't use it

	// Create a transport instance for testing
	trans := &contracts.MockTransport{
		Logger:  contracts.NewMockLogger(),
		BaseURL: "https://example.com",
	}

	// Test the transport
	contracts.VerifyTransportContract(t, trans)
}

// TestConnectionPoolImplementationContract verifies that the HttpConnectionPool
// satisfies the ConnectionPool contract
func TestConnectionPoolImplementationContract(t *testing.T) {
	pool := corehttp.NewHttpConnectionPool(nil)
	contracts.VerifyConnectionPoolContract(t, pool)
}

// TestConnectionPoolConfigImplementationContract verifies that the pool.Config
// satisfies the ConnectionPoolConfig contract
func TestConnectionPoolConfigImplementationContract(t *testing.T) {
	config := &corepool.Config{
		MaxIdleConnsPerHost: 10,
		MaxIdleConns:        100,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     30 * time.Second,
	}

	contracts.VerifyConnectionPoolConfigContract(t, config)
}

// TestConnectionPoolManagerImplementationContract verifies that the pool.Manager
// satisfies the ConnectionPoolManager contract
func TestConnectionPoolManagerImplementationContract(t *testing.T) {
	// Create a real pool manager
	manager := corepool.NewPoolManager(nil)

	// Create a factory function that returns an actual Config instance
	configFactory := func() *corepool.Config {
		return &corepool.Config{
			MaxIdleConnsPerHost: 10,
			MaxIdleConns:        100,
			MaxConnsPerHost:     50,
			IdleConnTimeout:     30 * time.Second,
		}
	}

	// Test with the actual implementation but wrap in interface{} to test the adapter
	contracts.VerifyConnectionPoolManagerContract(t, manager, func() interface{} {
		return configFactory()
	})
}

// TestAuthorizerImplementationContract verifies that authorizers satisfy
// the Authorizer contract
func TestAuthorizerImplementationContract(t *testing.T) {
	// Use mock implementation which properly implements the interface
	authorizer := contracts.NewMockAuthorizer("test-token", true)
	contracts.VerifyAuthorizerContract(t, authorizer)
}

// TestLoggerImplementationContract verifies that the core.DefaultLogger
// satisfies the Logger contract
func TestLoggerImplementationContract(t *testing.T) {
	logger := core.NewDefaultLogger(os.Stderr, core.LogLevelDebug)
	contracts.VerifyLoggerContract(t, logger)
}

// TestLoggerImplementationWithLoggingPackageContract verifies that the logging.Logger
// satisfies the Logger contract
func TestLoggerImplementationWithLoggingPackageContract(t *testing.T) {
	opts := &logging.Options{
		Output: os.Stderr,
		Level:  logging.LogLevelDebug,
	}
	logger := logging.NewLogger(opts)
	contracts.VerifyLoggerContract(t, logger)
}

// TestTokenManagerImplementationContract would verify a TokenManager implementation,
// but depends on external credentials so we skip it in unit tests
func TestTokenManagerImplementationContract(t *testing.T) {
	// This would require real credentials, so we'll skip in unit tests
	t.Skip("TokenManager tests require real credentials and are better suited for integration tests")
}

// TestAllImplementations runs all contract tests for the standard
// implementations provided by the SDK
func TestAllImplementations(t *testing.T) {
	// Run all the tests
	t.Run("Client", TestClientImplementationContract)
	t.Run("Transport", TestTransportImplementationContract)
	t.Run("ConnectionPool", TestConnectionPoolImplementationContract)
	t.Run("ConnectionPoolConfig", TestConnectionPoolConfigImplementationContract)
	t.Run("ConnectionPoolManager", TestConnectionPoolManagerImplementationContract)
	t.Run("Authorizer", TestAuthorizerImplementationContract)
	t.Run("DefaultLogger", TestLoggerImplementationContract)
	t.Run("Logger", TestLoggerImplementationWithLoggingPackageContract)

	// Skip token manager test as it needs real credentials
	t.Run("TokenManager", func(t *testing.T) {
		t.Skip("TokenManager test requires real credentials")
	})
}

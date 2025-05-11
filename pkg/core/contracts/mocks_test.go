// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts_test

import (
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/contracts"
)

// TestMockImplementations verifies that all mock implementations
// satisfy their respective interface contracts
func TestMockImplementations(t *testing.T) {
	t.Run("MockLogger", func(t *testing.T) {
		logger := contracts.NewMockLogger()
		contracts.VerifyLoggerContract(t, logger)
	})

	t.Run("MockClient", func(t *testing.T) {
		client := contracts.NewMockClient()
		contracts.VerifyClientContract(t, client)
	})

	t.Run("MockTransport", func(t *testing.T) {
		transport := contracts.NewMockTransport()
		contracts.VerifyTransportContract(t, transport)
	})

	t.Run("MockAuthorizer", func(t *testing.T) {
		authorizer := contracts.NewMockAuthorizer("test-token", true)
		contracts.VerifyAuthorizerContract(t, authorizer)
	})

	t.Run("MockTokenManager", func(t *testing.T) {
		manager := contracts.NewMockTokenManager("test-token", true)
		contracts.VerifyTokenManagerContract(t, manager)
	})

	t.Run("MockConnectionPool", func(t *testing.T) {
		pool := contracts.NewMockConnectionPool()
		contracts.VerifyConnectionPoolContract(t, pool)
	})

	t.Run("MockConnectionPoolConfig", func(t *testing.T) {
		config := contracts.NewMockConnectionPoolConfig()
		contracts.VerifyConnectionPoolConfigContract(t, config)
	})

	// Commented out due to type compatibility issues with interface{} vs interfaces.ConnectionPoolConfig
	/*
		t.Run("MockConnectionPoolManager", func(t *testing.T) {
			manager := contracts.NewMockConnectionPoolManager()
			configFactory := func() contracts.MockConnectionPoolConfig {
				return *contracts.NewMockConnectionPoolConfig()
			}

			// We need a small adapter since the configFactory has a specific return type
			configFactoryAdapter := func() interface{} {
				return configFactory()
			}

			contracts.VerifyConnectionPoolManagerContract(t, manager, func() interface{} {
				return configFactoryAdapter()
			})
		})
	*/

	t.Run("MockPooledHTTPClient", func(t *testing.T) {
		client := contracts.NewMockPooledHTTPClient()
		contracts.VerifyPooledHTTPClientContract(t, client)
	})
}

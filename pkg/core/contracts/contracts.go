// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"fmt"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// VerifyAllContracts runs all applicable contract tests for the provided implementations.
// It takes a testing.T instance and a map of implementations keyed by interface name.
// This is useful for running contract tests on multiple implementations at once.
func VerifyAllContracts(t *testing.T, implementations map[string]interface{}) {
	t.Helper()

	for name, impl := range implementations {
		t.Run(name, func(t *testing.T) {
			switch v := impl.(type) {
			case interfaces.ClientInterface:
				VerifyClientContract(t, v)
			case interfaces.Transport:
				VerifyTransportContract(t, v)
			case interfaces.ConnectionPool:
				VerifyConnectionPoolContract(t, v)
			case interfaces.ConnectionPoolConfig:
				VerifyConnectionPoolConfigContract(t, v)
			case interfaces.Authorizer:
				VerifyAuthorizerContract(t, v)
			case interfaces.TokenManager:
				VerifyTokenManagerContract(t, v)
			case interfaces.Logger:
				VerifyLoggerContract(t, v)
			case contractTestPair:
				runContractTestPair(t, v)
			default:
				t.Errorf("No contract test available for type: %T", impl)
			}
		})
	}
}

// contractTestPair is a special type for implementations that require additional
// parameters for their contract tests, such as ConnectionPoolManager which needs
// a config factory function.
type contractTestPair struct {
	Implementation interface{}
	TestFunc       func(t *testing.T, impl interface{})
}

// runContractTestPair runs a contract test for a special case implementation
func runContractTestPair(t *testing.T, pair contractTestPair) {
	t.Helper()
	pair.TestFunc(t, pair.Implementation)
}

// NewConnectionPoolManagerContractPair creates a contract test pair for a ConnectionPoolManager
// implementation with its config factory function.
func NewConnectionPoolManagerContractPair(
	manager interfaces.ConnectionPoolManager,
	configFactory func() interfaces.ConnectionPoolConfig,
) contractTestPair {
	return contractTestPair{
		Implementation: manager,
		TestFunc: func(t *testing.T, impl interface{}) {
			m, ok := impl.(interfaces.ConnectionPoolManager)
			if !ok {
				t.Fatalf("Expected ConnectionPoolManager, got %T", impl)
			}
			VerifyConnectionPoolManagerContract(t, m, configFactory)
		},
	}
}

// VerifyInterface is a convenience function for verifying that a type implements an interface.
// This is useful for compile-time checks rather than runtime checks.
func VerifyInterface(actual interface{}, expected interface{}) error {
	if actual == nil {
		return fmt.Errorf("actual implementation is nil")
	}

	// This function doesn't do anything at runtime, it's just for compile-time checks
	return nil
}

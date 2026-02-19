// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package pool provides connection pool management for the Globus Go SDK.

This package implements a higher-level connection pool manager that satisfies the
interfaces defined in pkg/core/interfaces and provides service-specific configuration
presets for Globus services. It is the primary connection pooling backend used by
the SDK's core package and service clients.

# STABILITY: BETA

This package is in beta. The core abstractions are stable, but some configuration
details and service presets may change in minor releases as Globus service
characteristics evolve. Changes will be documented in the CHANGELOG with
migration guidance.

The following components are considered beta-stable:

  - Config struct and all exported fields (MaxIdleConnsPerHost, MaxIdleConns,
    MaxConnsPerHost, IdleConnTimeout, DisableKeepAlives, ResponseHeaderTimeout,
    ExpectContinueTimeout, TLSHandshakeTimeout)
  - Config methods (GetMaxIdleConnsPerHost, GetMaxIdleConns, GetMaxConnsPerHost,
    GetIdleConnTimeout)
  - DefaultConfig constructor
  - ForService function (returns service-optimized Config)
  - Pool type and constructor (NewPool)
  - Pool methods (GetClient, SetTimeout, CloseIdleConnections, GetTransport, GetStats)
  - PoolStats struct and fields
  - PoolManager type and constructor (NewPoolManager)
  - PoolManager methods (GetPool, CloseAllIdleConnections, GetAllStats)
  - GlobalPoolManager package-level variable
  - GetServicePool convenience function

# Compatibility Guarantees

For beta components:
  - Minor backward-incompatible changes may still occur in minor releases
  - Service-specific configuration presets in ForService may be tuned in minor releases
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Get a connection pool for a specific service using the global manager:

	pool := pool.GetServicePool("transfer", nil)
	httpClient := pool.GetClient()

Get a pool with service-optimized configuration:

	config := pool.ForService("transfer")
	p := pool.NewPool(config)

Create a custom pool manager with a non-default configuration:

	defaultConfig := pool.DefaultConfig()
	defaultConfig.MaxIdleConns = 200

	manager := pool.NewPoolManager(defaultConfig)
	transferPool := manager.GetPool("transfer", nil)

	// Close idle connections when shutting down
	manager.CloseAllIdleConnections()
*/
package pool

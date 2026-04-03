// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package http provides HTTP connection pooling utilities for the Globus Go SDK.

This package implements HTTP connection pool management, providing efficient
reuse of TCP connections across requests to Globus services. It implements the
connection pool interfaces defined in pkg/core/interfaces and integrates with
the global pool manager used by the SDK's service clients.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - ConnectionPoolConfig struct and all exported fields (MaxIdleConnsPerHost,
    MaxIdleConns, MaxConnsPerHost, IdleConnTimeout, DisableKeepAlives,
    ResponseHeaderTimeout, ExpectContinueTimeout, TLSHandshakeTimeout,
    TLSClientConfig)
  - ConnectionPoolConfig methods (GetMaxIdleConnsPerHost, GetMaxIdleConns,
    GetMaxConnsPerHost, GetIdleConnTimeout)
  - DefaultConnectionPoolConfig constructor
  - ConnectionPool type and its constructor (NewConnectionPool)
  - ConnectionPool methods (GetClient, SetTimeout, CloseIdleConnections, GetTransport, GetStats)
  - ConnectionPoolStats struct and fields
  - ConnectionPoolManager type and constructor (NewConnectionPoolManager)
  - ConnectionPoolManager methods (GetPool, CloseAllIdleConnections, GetAllStats)
  - GlobalHttpPoolManager package-level variable
  - GetServicePool and GetHTTPClientForService convenience functions
  - HttpConnectionPool and HttpConnectionPoolManager adapter types
    (implement pkg/core/interfaces connection pool interfaces)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Get an HTTP client for a specific service using the global pool manager:

	httpClient := http.GetHTTPClientForService("transfer", nil)

Create a custom connection pool:

	config := http.DefaultConnectionPoolConfig()
	config.MaxIdleConnsPerHost = 8
	config.IdleConnTimeout = 120 * time.Second

	pool := http.NewConnectionPool(config)
	client := pool.GetClient()

Manage connection pools per service:

	manager := http.NewConnectionPoolManager(nil)
	transferPool := manager.GetPool("transfer", nil)
	authPool := manager.GetPool("auth", nil)

	// Get statistics for all pools
	stats := manager.GetAllStats()

Close idle connections when done:

	manager.CloseAllIdleConnections()
*/
package http

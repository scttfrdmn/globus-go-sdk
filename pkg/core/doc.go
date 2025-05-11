// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package core provides the foundational components for the Globus Go SDK.

# STABILITY: BETA

This package is approaching stability but may still undergo minor changes.
Components listed below are considered relatively stable, but may have
minor signature changes before the package is marked as stable:

  - Client interface and implementation - Core client behaviors
  - Core functional options (WithContext, WithHTTPClient, etc.)
  - Logger interface - Standard logging behaviors
  - Error handling utilities - API error types and helpers

The following components are less stable and more likely to evolve:

  - Connection pool interfaces and implementations
  - HTTP transport layer and configuration
  - Rate limiting and backoff mechanisms

# Connection Pooling API

The connection pooling API (introduced in v0.9.11) provides functionality for
managing HTTP connection pools across service clients:

  - SetConnectionPoolManager - Sets the global connection pool manager
  - EnableDefaultConnectionPool - Configures and enables the default connection pool
  - GetConnectionPool - Retrieves the current connection pool
  - GetHTTPClientForService - Gets an HTTP client for a specific service

These connection pool functions were previously defined in client_with_pool.go
and are now maintained in connection_pool.go to ensure backward compatibility.

# Compatibility Notes

For beta packages:
  - Some efforts are made to maintain backward compatibility
  - Breaking changes are documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices
  - Migration paths will be provided for any breaking changes

This package is expected to reach stable status in version v1.0.0.
Until then, users should review the CHANGELOG when upgrading.

# Basic Usage

Creating a client with default options:

	client := core.NewClient("https://transfer.api.globus.org")

Creating a client with custom options:

	httpClient := &http.Client{Timeout: 30 * time.Second}
	client := core.NewClient(
		"https://transfer.api.globus.org",
		core.WithHTTPClient(httpClient),
		core.WithUserAgent("MyApp/1.0"),
	)

Using connection pooling:

	// Enable default connection pool for all clients
	core.EnableDefaultConnectionPool()

	// Create client that will use the connection pool
	client := core.NewClient("https://transfer.api.globus.org")
*/
package core

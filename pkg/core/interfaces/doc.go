// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package interfaces provides core interface definitions for the Globus Go SDK.

This package defines the fundamental contracts between components of the SDK.
All other packages depend on these interfaces rather than concrete types,
enabling testability, extensibility, and clean separation of concerns.
Custom implementations of these interfaces can be substituted anywhere the
SDK accepts interface values.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Authorizer interface (GetAuthorizationHeader, IsValid, GetToken)
  - TokenManager interface (GetToken, RefreshToken, RevokeToken, IsValid)
  - ClientInterface interface (Do, GetHTTPClient, GetBaseURL, GetUserAgent, GetLogger)
  - Logger interface (Debug, Info, Warn, Error)
  - HTTPClient interface (Do)
  - HTTPDoer interface (Do with context)
  - Transport interface (Request, Get, Post, Put, Delete, Patch, RoundTrip)
  - ConnectionPool interface (GetClient, SetTimeout, CloseIdleConnections, GetTransport)
  - ConnectionPoolConfig interface (GetMaxIdleConnsPerHost, GetMaxIdleConns,
    GetMaxConnsPerHost, GetIdleConnTimeout)
  - ConnectionPoolManager interface (GetPool, CloseAllIdleConnections, GetAllStats)
  - PooledHTTPClient interface (GetPool, Do)

# Compatibility Guarantees

For stable components:
  - Interface method signatures will not change incompatibly in minor or patch releases
  - New methods will not be added to existing interfaces in minor releases (this would
    break existing implementations); new interfaces will be introduced instead
  - Deprecated interfaces will be marked with appropriate notices
  - Deprecated interfaces will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Implement the Authorizer interface for a custom authorization scheme:

	type MyAuthorizer struct{ token string }

	func (a *MyAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
		return "Bearer " + a.token, nil
	}

	func (a *MyAuthorizer) IsValid() bool {
		return a.token != ""
	}

	func (a *MyAuthorizer) GetToken() string {
		return a.token
	}

Implement the Transport interface for testing:

	type MockTransport struct{}

	func (t *MockTransport) Get(ctx context.Context, path string,
		query url.Values, headers http.Header) (*http.Response, error) {
		// Return a mock response
	}
	// ... implement remaining interface methods
*/
package interfaces

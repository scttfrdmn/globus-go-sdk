// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package transport provides the HTTP transport layer for the Globus Go SDK.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Transport type and its constructor (NewTransport)
  - DeferredTransport type and its constructor (NewDeferredTransport)
  - DeferredTransport.AttachClient method
  - Transport.Request method (GET, POST, PUT, DELETE, PATCH, Request)
  - Transport.Get, Post, Put, Delete, Patch convenience methods
  - Transport.RoundTrip method (implements http.RoundTripper)
  - Options configuration struct and its fields (Debug, Trace, Logger)
  - DecodeResponse helper function
  - Environment variable support (GLOBUS_SDK_HTTP_DEBUG, GLOBUS_SDK_HTTP_TRACE)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create a transport for a service client:

	client := &MyClient{}
	transport := transport.NewTransport(client, &transport.Options{
		Debug: true,
	})

Make an HTTP GET request:

	resp, err := transport.Get(ctx, "/v2/some/resource", nil, nil)
	if err != nil {
		// Handle error
	}
	defer resp.Body.Close()

Decode a JSON response:

	var result MyResult
	if err := transport.DecodeResponse(resp, &result); err != nil {
		// Handle error
	}

Use DeferredTransport when the client is not yet available:

	dt := transport.NewDeferredTransport(&transport.Options{Debug: true})
	// Later, once the client is available:
	t := dt.AttachClient(client)
*/
package transport

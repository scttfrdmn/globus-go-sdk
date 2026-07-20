// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package compute

// The upstream Globus Compute SDK defines no request or response models and no
// pagination: every request body is a passthrough document and every response is
// a passthrough JSON object. The Go client therefore accepts and returns
// map[string]interface{}. Only the two query-param option structs are typed.

// GetVersionOptions carries the optional "service" query param for GetVersion.
type GetVersionOptions struct {
	Service string
}

// GetEndpointsOptions carries the optional "role" query param for GetEndpoints.
type GetEndpointsOptions struct {
	Role string
}

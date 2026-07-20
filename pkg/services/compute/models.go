// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package compute

// The upstream Globus Compute SDK (3.65.0) defines no request or response models
// and no pagination: every request body is a passthrough document and every
// response is a passthrough JSON object. The Go client therefore accepts and
// returns map[string]interface{}; only the query-param option struct is typed
// (GetEndpointsOptions, defined in client.go).

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package auth

import "context"

// Authorizer defines the interface for components that can authorize HTTP requests
type Authorizer interface {
	// GetAuthorizationHeader returns the authorization header value
	GetAuthorizationHeader(ctx ...context.Context) (string, error)
}
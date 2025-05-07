// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package interfaces

import (
	"context"
)

// Authorizer defines the interface for authorization
type Authorizer interface {
	// GetAuthorizationHeader returns the authorization header value
	GetAuthorizationHeader(ctx context.Context) (string, error)

	// IsValid returns whether the current authorization is valid
	IsValid() bool

	// GetToken returns the current token
	GetToken() string
}

// TokenManager defines the interface for token management
type TokenManager interface {
	// GetToken returns the current token
	GetToken(ctx context.Context) (string, error)

	// RefreshToken refreshes the current token if necessary
	RefreshToken(ctx context.Context) error

	// RevokeToken revokes the current token
	RevokeToken(ctx context.Context) error

	// IsValid returns whether the current token is valid
	IsValid() bool
}

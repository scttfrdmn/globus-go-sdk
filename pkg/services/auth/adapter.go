// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"

	coreauthlib "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// AuthorizerAdapter adapts between the core.auth.Authorizer and interfaces.Authorizer
type AuthorizerAdapter struct {
	authorizer interfaces.Authorizer
}

// NewAuthorizerAdapter creates a new adapter for the given authorizer
func NewAuthorizerAdapter(authorizer interfaces.Authorizer) *AuthorizerAdapter {
	return &AuthorizerAdapter{
		authorizer: authorizer,
	}
}

// GetAuthorizationHeader implements the core.auth.Authorizer interface
func (a *AuthorizerAdapter) GetAuthorizationHeader(ctx ...context.Context) (string, error) {
	// Use the first context if available, otherwise use background
	var c context.Context
	if len(ctx) > 0 && ctx[0] != nil {
		c = ctx[0]
	} else {
		c = context.Background()
	}

	return a.authorizer.GetAuthorizationHeader(c)
}

// IsValid implements the core.auth.Authorizer interface
func (a *AuthorizerAdapter) IsValid() bool {
	return a.authorizer.IsValid()
}

// GetToken implements the core.auth.Authorizer interface
func (a *AuthorizerAdapter) GetToken() string {
	return a.authorizer.GetToken()
}

// Ensure AuthorizerAdapter implements the core.auth.Authorizer interface
var _ coreauthlib.Authorizer = (*AuthorizerAdapter)(nil)

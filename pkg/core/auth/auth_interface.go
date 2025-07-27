// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package auth

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/interfaces"
)

// BasicAuthorizer implements the interfaces.Authorizer interface for static tokens
type BasicAuthorizer struct {
	token string
}

// NewBasicAuthorizer creates a new BasicAuthorizer
func NewBasicAuthorizer(token string) *BasicAuthorizer {
	return &BasicAuthorizer{
		token: token,
	}
}

// GetAuthorizationHeader returns the authorization header value
func (a *BasicAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	return "Bearer " + a.token, nil
}

// IsValid returns whether the current authorization is valid
func (a *BasicAuthorizer) IsValid() bool {
	return a.token != ""
}

// GetToken returns the current token
func (a *BasicAuthorizer) GetToken() string {
	return a.token
}

// Ensure BasicAuthorizer implements the interfaces.Authorizer interface
var _ interfaces.Authorizer = (*BasicAuthorizer)(nil)

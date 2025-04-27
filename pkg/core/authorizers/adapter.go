// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package authorizers

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

// CoreAuthorizer adapts an Authorizer to the auth.Authorizer interface
type CoreAuthorizer struct {
	Authorizer Authorizer
}

// NewCoreAuthorizer creates a new CoreAuthorizer that wraps an Authorizer
func NewCoreAuthorizer(authorizer Authorizer) auth.Authorizer {
	return &CoreAuthorizer{Authorizer: authorizer}
}

// GetAuthorizationHeader implements the auth.Authorizer interface
func (a *CoreAuthorizer) GetAuthorizationHeader(ctxArr ...context.Context) (string, error) {
	var ctx context.Context
	if len(ctxArr) > 0 {
		ctx = ctxArr[0]
	} else {
		ctx = context.Background()
	}
	return a.Authorizer.GetAuthorizationHeader(ctx)
}

// ToCore adapts any Authorizer to an auth.Authorizer
func ToCore(a Authorizer) auth.Authorizer {
	return NewCoreAuthorizer(a)
}

// NullCoreAuthorizer returns a null authorizer that implements auth.Authorizer
func NullCoreAuthorizer() auth.Authorizer {
	return ToCore(&NullAuthorizer{})
}

// StaticTokenCoreAuthorizer creates a static token authorizer that implements auth.Authorizer
func StaticTokenCoreAuthorizer(token string) auth.Authorizer {
	return ToCore(NewStaticTokenAuthorizer(token))
}

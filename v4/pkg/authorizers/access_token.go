// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// AccessTokenAuthorizer provides a static bearer token that never refreshes.
// It implements core.Authorizer.
type AccessTokenAuthorizer struct {
	accessToken string
}

// NewAccessTokenAuthorizer creates an authorizer that always sends the given
// access token as a Bearer header.
func NewAccessTokenAuthorizer(accessToken string) core.Authorizer {
	return &AccessTokenAuthorizer{accessToken: accessToken}
}

// GetAuthorizationHeader returns the Bearer authorization header.
func (a *AccessTokenAuthorizer) GetAuthorizationHeader(_ context.Context) (string, error) {
	if a.accessToken == "" {
		return "", &core.ValidationError{
			Field:   "accessToken",
			Message: "access token is empty",
		}
	}
	return fmt.Sprintf("Bearer %s", a.accessToken), nil
}

// HandleMissingAuthorization always returns false — a static token cannot be refreshed.
func (a *AccessTokenAuthorizer) HandleMissingAuthorization(_ context.Context) bool {
	return false
}

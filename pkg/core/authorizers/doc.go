// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package authorizers provides authorization mechanisms for the Globus Go SDK.

This package implements the Authorizer interface for several common OAuth2
authorization patterns used across Globus services. Authorizers are attached
to service clients and are responsible for supplying Authorization headers,
detecting expired credentials, and refreshing tokens when possible.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Authorizer interface (GetAuthorizationHeader, HandleMissingAuthorization, IsExpired)
  - NullAuthorizer type (no-op authorizer for unauthenticated requests)
  - StaticTokenAuthorizer type and constructor (NewStaticTokenAuthorizer)
  - StaticTokenAuthorizer methods (GetAuthorizationHeader, HandleMissingAuthorization,
    IsExpired, IsValid, GetToken)
  - RefreshableTokenAuthorizer type and constructor (NewRefreshableTokenAuthorizer)
  - RefreshableTokenAuthorizer methods (GetAuthorizationHeader, HandleMissingAuthorization,
    IsExpired)
  - ClientCredentialsAuthorizer type and constructor (NewClientCredentialsAuthorizer)
  - ClientCredentialsAuthorizer methods (GetAuthorizationHeader, HandleMissingAuthorization,
    IsExpired)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create a static token authorizer for a long-lived token:

	authorizer := authorizers.NewStaticTokenAuthorizer("my-access-token")
	header, err := authorizer.GetAuthorizationHeader(ctx)
	if err != nil {
		// Handle error
	}

Create a refreshable token authorizer that automatically renews expired tokens:

	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		// Call the Globus Auth service to refresh the token
		return newAccessToken, newRefreshToken, expiresAt, nil
	}

	authorizer := authorizers.NewRefreshableTokenAuthorizer(
		accessToken,
		refreshToken,
		expiresIn, // seconds until expiry
		refreshFunc,
	)

Create a client credentials authorizer for machine-to-machine access:

	authFunc := func(ctx context.Context, clientID, clientSecret string, scopes []string) (string, time.Time, error) {
		// Call the Globus Auth client credentials endpoint
		return accessToken, expiresAt, nil
	}

	authorizer := authorizers.NewClientCredentialsAuthorizer(
		"client-id",
		"client-secret",
		[]string{"urn:globus:auth:scope:transfer.api.globus.org:all"},
		authFunc,
	)
*/
package authorizers

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package auth provides a client for interacting with the Globus Auth service.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Client interface and implementation
  - OAuth2 flow methods (GetAuthorizationURL, ExchangeAuthorizationCode, etc.)
  - Token management methods (IntrospectToken, RevokeToken, etc.)
  - Authorizer factory methods (CreateClientCredentialsAuthorizer, etc.)
  - Token utility methods (IsTokenValid, GetTokenExpiry, etc.)
  - Core model types (TokenResponse, TokenInfo, UserInfo)
  - Error checking functions (IsInvalidGrant, IsInvalidClient, etc.)
  - Client configuration options (WithClientID, WithClientSecret, etc.)

MFA-related components are considered BETA:
  - MFAChallenge and MFAResponse types
  - MFA-related methods (ExchangeAuthorizationCodeWithMFA, etc.)
  - MFARequiredError type

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

For beta components (MFA-related):
  - Minor backward-incompatible changes may still occur in minor releases
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Create a new auth client:

	authClient := auth.NewClient(
		auth.WithClientID("your-client-id"),
		auth.WithClientSecret("your-client-secret"),
		auth.WithRedirectURL("https://your-app.example.com/callback"),
	)

OAuth2 authorization code flow:

	// Get authorization URL for user to visit
	authURL := authClient.GetAuthorizationURL(
		[]string{"openid", "profile", "email"},
		auth.WithState("random-state-value"),
	)

	// After user is redirected to your callback URL with a code...
	tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
	if err != nil {
		// Handle error
	}

	// Use the tokens
	accessToken := tokenResponse.AccessToken
	refreshToken := tokenResponse.RefreshToken

Client credentials flow:

	// Get a token using client credentials
	tokenResponse, err := authClient.GetClientCredentialsToken(
		ctx, []string{"https://auth.globus.org/scopes/transfer.api.globus.org/all"},
	)
	if err != nil {
		// Handle error
	}

Token management:

	// Refresh a token
	newTokenResponse, err := authClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		// Handle error
	}

	// Introspect a token
	tokenInfo, err := authClient.IntrospectToken(ctx, accessToken)
	if err != nil {
		// Handle error
	}
	if !tokenInfo.Active {
		// Token is no longer valid
	}

	// Revoke a token
	err = authClient.RevokeToken(ctx, accessToken)
	if err != nil {
		// Handle error
	}

Creating authorizers:

	// Create a static token authorizer
	authorizer := authClient.CreateStaticTokenAuthorizer(accessToken)

	// Create a refreshable token authorizer
	authorizer, err := authClient.CreateRefreshableTokenAuthorizer(ctx, refreshToken)
	if err != nil {
		// Handle error
	}

	// Create a client credentials authorizer
	authorizer, err := authClient.CreateClientCredentialsAuthorizer(
		ctx, []string{"https://auth.globus.org/scopes/transfer.api.globus.org/all"},
	)
	if err != nil {
		// Handle error
	}
*/
package auth

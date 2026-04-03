// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

/*
Package auth provides core OAuth2 primitives for the Globus Go SDK.

This package contains low-level OAuth2 abstractions used by the higher-level
auth service client (pkg/services/auth). It is primarily intended for use by
SDK developers building service clients, or for advanced use cases requiring
direct control over token storage and lifecycle management.

# STABILITY: STABLE

This package follows semantic versioning. Components listed below are
considered part of the public API and will not change incompatibly
within a major version:

  - Authorizer interface (core OAuth2 authorizer contract)
  - BasicAuthorizer type and constructor (NewBasicAuthorizer)
  - BasicAuthorizer methods (GetAuthorizationHeader, IsValid, GetToken)
  - TokenInfo struct and its fields (AccessToken, RefreshToken, ExpiresAt, Scopes, ResourceID)
  - TokenInfo methods (IsValid, CanRefresh)
  - TokenStorage interface (StoreToken, GetToken, DeleteToken, ListTokens)
  - MemoryTokenStorage type and constructor (NewMemoryTokenStorage)
  - FileTokenStorage type and constructor (NewFileTokenStorage)
  - TokenManager type and constructor (NewTokenManager)
  - TokenManager methods (GetToken, StoreToken, RefreshToken, SetRefreshThreshold, StartBackgroundRefresh)
  - RefreshFunc type alias
  - Sentinel errors (ErrTokenNotFound, ErrStorageCorrupt)

# Compatibility Guarantees

For stable components:
  - Public API signatures will not change incompatibly in minor or patch releases
  - New functionality will be added in backward-compatible ways
  - Deprecated functionality will be marked with appropriate notices
  - Deprecated functionality will be maintained for at least one major release cycle
  - Any breaking changes will only occur in major version bumps (e.g., v1.0.0 to v2.0.0)

# Basic Usage

Create an in-memory token storage and token manager:

	storage := auth.NewMemoryTokenStorage()

	refreshFunc := func(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
		// Call the Globus Auth service to refresh the token
		return newAccessToken, newRefreshToken, expiresAt, nil
	}

	manager := auth.NewTokenManager(storage, refreshFunc)

Store and retrieve tokens:

	token := auth.TokenInfo{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := manager.StoreToken(ctx, "my-resource", token); err != nil {
		// Handle error
	}

	retrieved, err := manager.GetToken(ctx, "my-resource")
	if err != nil {
		// Handle error
	}

Use a static token authorizer:

	authorizer := auth.NewBasicAuthorizer("my-access-token")
	header, err := authorizer.GetAuthorizationHeader(ctx)
	if err != nil {
		// Handle error
	}

Persist tokens to disk:

	storage, err := auth.NewFileTokenStorage("/path/to/token/dir")
	if err != nil {
		// Handle error
	}
*/
package auth

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package app provides high-level application abstractions for the Globus SDK,
// mirroring the Python SDK's globus_sdk.globus_app module.
//
// # Available Application Types
//
//   - UserApp: for applications that authenticate on behalf of a human user
//     via an interactive login flow (browser + auth code).
//   - ClientApp: for service accounts / machine-to-machine applications that
//     authenticate using the client credentials grant (no user interaction).
//
// # Basic usage — service account
//
//	a, err := app.NewClientApp("my-client-id", "my-client-secret", nil)
//	a.AddScopeRequirements("transfer.api.globus.org", core.Scopes.TransferAll)
//	auth, err := a.GetAuthorizer(ctx, "transfer.api.globus.org")
//	cfg := &core.Config{Authorizer: auth, Scopes: []string{core.Scopes.TransferAll}}
//	client, err := transfer.NewClient(ctx, cfg)
//
// # Stability: BETA
//
// The API of this package may have minor changes in minor releases.
package app

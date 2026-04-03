// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package tokenstorage

// TokenStorage is the interface for persisting OAuth2 token data.
// Implementations must be safe for concurrent use.
type TokenStorage interface {
	// Store saves or replaces token data for the given resource server.
	Store(data *TokenData) error

	// Get retrieves token data for the given resource server.
	// Returns nil, nil if no data is found.
	Get(resourceServer string) (*TokenData, error)

	// Remove deletes token data for the given resource server.
	// Is a no-op if the resource server is not found.
	Remove(resourceServer string) error

	// GetAll returns all stored token data.
	GetAll() ([]*TokenData, error)

	// Close releases any resources held by the storage implementation.
	Close() error
}

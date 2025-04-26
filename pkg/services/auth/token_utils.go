// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package auth

import (
	"context"
	"errors"
	"time"
)

// ErrTokenInvalid is returned when a token is invalid or expired
var ErrTokenInvalid = errors.New("token is invalid or expired")

// ErrTokenExpired is returned when a token is expired
var ErrTokenExpired = errors.New("token is expired")

// ValidateToken validates the given token using introspection
// Returns nil if the token is valid, or an error if invalid or expired
func (c *Client) ValidateToken(ctx context.Context, token string) error {
	// Introspect the token
	info, err := c.IntrospectToken(ctx, token)
	if err != nil {
		return err
	}

	// Check if the token is active
	if !info.Active {
		return ErrTokenInvalid
	}

	// Check if the token is expired
	if info.IsExpired() {
		return ErrTokenExpired
	}

	return nil
}

// GetTokenExpiry extracts the expiry time for a token
// Returns the expiry time and whether the token is valid
func (c *Client) GetTokenExpiry(ctx context.Context, token string) (time.Time, bool, error) {
	// Introspect the token
	info, err := c.IntrospectToken(ctx, token)
	if err != nil {
		return time.Time{}, false, err
	}

	// Check if the token is active
	if !info.Active {
		return time.Time{}, false, nil
	}

	return info.ExpiresAt(), true, nil
}

// IsTokenValid checks if a token is valid
// A convenience wrapper around ValidateToken
func (c *Client) IsTokenValid(ctx context.Context, token string) bool {
	err := c.ValidateToken(ctx, token)
	return err == nil
}

// GetRemainingValidity returns the duration until a token expires
// Returns 0 if the token is already expired or invalid
func (c *Client) GetRemainingValidity(ctx context.Context, token string) (time.Duration, error) {
	// Get token expiry
	expiry, valid, err := c.GetTokenExpiry(ctx, token)
	if err != nil {
		return 0, err
	}

	// If token is not valid, return 0 duration
	if !valid {
		return 0, nil
	}

	// Calculate remaining duration
	remaining := time.Until(expiry)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// ShouldRefresh checks if a token should be refreshed based on a threshold
// Returns true if the token will expire within the given threshold
func (c *Client) ShouldRefresh(ctx context.Context, token string, threshold time.Duration) (bool, error) {
	// Get remaining validity
	remaining, err := c.GetRemainingValidity(ctx, token)
	if err != nil {
		return false, err
	}

	// If already expired or will expire within threshold, it should be refreshed
	return remaining <= threshold, nil
}
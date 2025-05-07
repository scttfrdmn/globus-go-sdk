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

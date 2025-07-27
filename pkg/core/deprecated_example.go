// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"context"
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/deprecation"
)

// WithDebugMode enables debug mode for the client.
//
// Deprecated: This function was deprecated in v0.9.15 and will be removed in v1.0.0.
// Use WithDebug(true) instead.
func WithDebugMode() ClientOption {
	return func(c *Client) {
		// Log a deprecation warning when this option is used
		deprecation.LogWarning(
			c.Logger,
			"WithDebugMode",
			"v0.9.15",
			"v1.0.0",
			"Use WithDebug(true) instead.",
		)
		c.Debug = true
	}
}

// DoWithRetry performs an HTTP request with automatic retries.
//
// Deprecated: This function was deprecated in v0.9.15 and will be removed in v1.0.0.
// Use the rate limiter's DoWithRetry method instead.
func (c *Client) DoWithRetry(ctx context.Context, req *http.Request, maxRetries int) (*http.Response, error) {
	// Log a deprecation warning
	deprecation.LogWarning(
		c.Logger,
		"Client.DoWithRetry",
		"v0.9.15",
		"v1.0.0",
		"Use the rate limiter's DoWithRetry method instead.",
	)

	// Simple implementation for demonstration
	var resp *http.Response
	var err error

	for i := 0; i <= maxRetries; i++ {
		resp, err = c.Do(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Check if context is canceled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue retrying
		}
	}

	return resp, err
}

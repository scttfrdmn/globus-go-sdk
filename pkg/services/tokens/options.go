// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package tokens

import (
	"fmt"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
)

// ClientOption configures a Token Manager
type ClientOption func(*clientOptions)

// clientOptions represents options for configuring a Token Manager
type clientOptions struct {
	storage          Storage
	refreshHandler   RefreshHandler
	refreshThreshold time.Duration
}

// defaultOptions returns the default client options
func defaultOptions() *clientOptions {
	return &clientOptions{
		storage:          NewMemoryStorage(),
		refreshThreshold: 5 * time.Minute,
	}
}

// WithStorage sets the token storage mechanism
func WithStorage(storage Storage) ClientOption {
	return func(o *clientOptions) {
		o.storage = storage
	}
}

// WithRefreshHandler sets the refresh handler for token refreshing
func WithRefreshHandler(refreshHandler RefreshHandler) ClientOption {
	return func(o *clientOptions) {
		o.refreshHandler = refreshHandler
	}
}

// WithAuthClient sets an auth client as the refresh handler
func WithAuthClient(authClient *auth.Client) ClientOption {
	return func(o *clientOptions) {
		o.refreshHandler = authClient
	}
}

// WithRefreshThreshold sets the threshold for automatic token refreshing
func WithRefreshThreshold(threshold time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.refreshThreshold = threshold
	}
}

// WithFileStorage sets file-based storage with the specified directory
func WithFileStorage(directory string) ClientOption {
	return func(o *clientOptions) {
		// Create file storage
		fileStorage, err := NewFileStorage(directory)
		if err != nil {
			// Log the error but don't fail - fall back to memory storage
			fmt.Fprintf(os.Stderr, "Failed to create file storage: %v, falling back to memory storage\n", err)
			return
		}
		o.storage = fileStorage
	}
}
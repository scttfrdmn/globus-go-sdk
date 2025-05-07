// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package interfaces

import (
	"context"
	"net/http"
)

// ClientInterface defines the interface for a base client
type ClientInterface interface {
	// Do performs an HTTP request and handles common error cases
	Do(ctx context.Context, req *http.Request) (*http.Response, error)

	// GetHTTPClient returns the underlying HTTP client
	GetHTTPClient() *http.Client

	// GetBaseURL returns the base URL for the client
	GetBaseURL() string

	// GetUserAgent returns the user agent string for the client
	GetUserAgent() string

	// GetLogger returns the logger for the client
	GetLogger() Logger
}

// Logger defines the interface for logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

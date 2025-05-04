// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package interfaces

import (
	"context"
	"net/http"
	"net/url"
)

// HTTPClient defines the interface for an HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPDoer defines an interface for making HTTP requests
type HTTPDoer interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// Transport defines the interface for making API requests
type Transport interface {
	Request(ctx context.Context, method, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error)
	Get(ctx context.Context, path string, query url.Values, headers http.Header) (*http.Response, error)
	Post(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error)
	Put(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error)
	Delete(ctx context.Context, path string, query url.Values, headers http.Header) (*http.Response, error)
	Patch(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error)
	RoundTrip(req *http.Request) (*http.Response, error)
}
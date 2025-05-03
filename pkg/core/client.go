// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"context"
	"net/http"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/ratelimit"
)

// Client defines the base client used by all service-specific clients
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	UserAgent   string
	Logger      interfaces.Logger
	Authorizer  auth.Authorizer
	RateLimiter ratelimit.RateLimiter
	Transport   interfaces.Transport
	Debug       bool
	Trace       bool
}

// NewClient creates a new base client with default settings
func NewClient(options ...ClientOption) *Client {
	// Initialize with defaults
	client := &Client{
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
		UserAgent: "globus-go-sdk/1.0",
		Logger:    NewDefaultLogger(nil, LogLevelNone),
		Debug:     false,
		Trace:     false,
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Create Transport if it hasn't been created yet
	if client.Transport == nil {
		// Use the helper function to initialize the transport
		client.Transport = InitTransport(client, client.Debug, client.Trace)
	}

	return client
}

// ClientOption defines a function that configures the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithAuthorizer sets the authorizer to use for authentication
func WithAuthorizer(authorizer auth.Authorizer) ClientOption {
	return func(c *Client) {
		c.Authorizer = authorizer
	}
}

// WithRateLimiter sets the rate limiter to use for API requests
func WithRateLimiter(limiter ratelimit.RateLimiter) ClientOption {
	return func(c *Client) {
		c.RateLimiter = limiter
	}
}

// WithHTTPDebugging enables HTTP request/response logging
func WithHTTPDebugging(enable bool) ClientOption {
	return func(c *Client) {
		c.Debug = enable
		// Transport will be initialized in NewClient if it's nil
	}
}

// WithHTTPTracing enables detailed HTTP tracing including headers and bodies
func WithHTTPTracing(enable bool) ClientOption {
	return func(c *Client) {
		c.Trace = enable
		// Transport will be initialized in NewClient if it's nil
		// Make sure debug is also enabled if tracing is
		if enable {
			c.Debug = true // Tracing requires debug mode
		}
	}
}

// Do performs an HTTP request and handles common error cases
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Set common headers
	req.Header.Set("User-Agent", c.UserAgent)

	// Apply authorization if available
	if c.Authorizer != nil {
		header, err := c.Authorizer.GetAuthorizationHeader(ctx)
		if err != nil {
			return nil, err
		}
		if header != "" {
			req.Header.Set("Authorization", header)
		}
	}

	// Apply rate limiting if configured
	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			c.Logger.Error("Rate limiting failed: %v", err)
			return nil, err
		}
	}

	// Log the request
	c.Logger.Debug("Making request to %s %s", req.Method, req.URL.String())

	// Execute request with context
	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		c.Logger.Error("Request failed: %v", err)
		return nil, err
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		err = NewAPIError(resp)
		c.Logger.Error("API error: %v", err)
		return resp, err
	}

	return resp, nil
}


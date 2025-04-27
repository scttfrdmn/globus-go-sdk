// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"context"
	"net/http"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

// Client defines the base client used by all service-specific clients
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	Logger     Logger
	Authorizer auth.Authorizer
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
	}

	// Apply options
	for _, option := range options {
		option(client)
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

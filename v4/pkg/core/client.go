// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the base client for all Globus SDK v4 service clients
// It implements context-first design and enhanced error handling
type Client struct {
	config            *Config
	httpClientCreated bool // tracks if we created the HTTP client internally
}

// NewClient creates a new base client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, &ValidationError{
			Message: "config cannot be nil",
		}
	}

	// Track if we'll create the HTTP client
	httpClientCreated := config.HTTPClient == nil

	// Apply defaults
	config = config.WithDefaults()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		config:            config,
		httpClientCreated: httpClientCreated,
	}, nil
}

// DoRequest performs an HTTP request with context, retry logic, and error handling
// This is the core method used by all service clients. The endpoint is joined onto
// the configured base URL.
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, query url.Values, body interface{}, result interface{}) error {
	return c.DoRequestURL(ctx, method, c.buildURL(endpoint, query), body, result)
}

// DoRequestNoAuth is like DoRequest but omits the Authorization header. It is
// used for the few endpoints that must be called unauthenticated (e.g. the GCS
// manager's GET /info).
func (c *Client) DoRequestNoAuth(ctx context.Context, method, endpoint string, query url.Values, body interface{}, result interface{}) error {
	return c.doRequestURL(ctx, method, c.buildURL(endpoint, query), body, result, true)
}

// DoRequestURL performs an HTTP request against a fully-formed URL, bypassing the
// base-URL join. It is used for the handful of endpoints that live outside a
// client's base path (e.g. Auth's host-root OIDC discovery / JWKS URIs). The
// retry, auth, and decoding behavior matches DoRequest.
func (c *Client) DoRequestURL(ctx context.Context, method, reqURL string, body interface{}, result interface{}) error {
	return c.doRequestURL(ctx, method, reqURL, body, result, false)
}

func (c *Client) doRequestURL(ctx context.Context, method, reqURL string, body interface{}, result interface{}, skipAuth bool) error {
	// Marshal request body if provided.
	// A url.Values body is sent as application/x-www-form-urlencoded (required
	// by the OAuth2 token/introspect/revoke endpoints); anything else is JSON.
	var bodyReader io.Reader
	contentType := ""
	if body != nil {
		if form, ok := body.(url.Values); ok {
			bodyReader = strings.NewReader(form.Encode())
			contentType = "application/x-www-form-urlencoded"
		} else {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return &ValidationError{
					Message: fmt.Sprintf("failed to marshal request body: %v", err),
					Value:   body,
				}
			}
			bodyReader = bytes.NewReader(bodyBytes)
			contentType = "application/json"
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return &NetworkError{
			Operation: "create_request",
			Message:   "failed to create HTTP request",
			Err:       err,
		}
	}

	// Set headers
	req.Header.Set("User-Agent", c.config.UserAgent)
	if !skipAuth {
		if c.config.Authorizer != nil {
			authHeader, authErr := c.config.Authorizer.GetAuthorizationHeader(ctx)
			if authErr != nil {
				return &NetworkError{
					Operation: "get_auth_header",
					Message:   "authorizer failed to provide authorization header",
					Err:       authErr,
				}
			}
			req.Header.Set("Authorization", authHeader)
		} else {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))
		}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	// Perform request with retry logic
	resp, err := c.doWithRetry(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return NewAPIError(resp, string(bodyBytes))
	}

	// Decode response if result is provided
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return &ValidationError{
				Message: fmt.Sprintf("failed to decode response: %v", err),
			}
		}
	}

	return nil
}

// doWithRetry performs an HTTP request with retry logic
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	retryConfig := c.config.RetryConfig
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		// Check context before making request
		if err := ctx.Err(); err != nil {
			return nil, &NetworkError{
				Operation: "check_context",
				Message:   "context cancelled or expired",
				Err:       err,
			}
		}

		// Make the request
		resp, err := c.config.HTTPClient.Do(req)

		// If no error and status is not retryable, return immediately
		if err == nil {
			shouldRetry := false
			for _, code := range retryConfig.RetryableStatusCodes {
				if resp.StatusCode == code {
					shouldRetry = true
					break
				}
			}
			if !shouldRetry {
				return resp, nil
			}
			// Close response body for retryable errors
			resp.Body.Close()
			lastErr = fmt.Errorf("retryable HTTP status: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		// Don't sleep after the last attempt
		if attempt < retryConfig.MaxRetries {
			backoff := c.calculateBackoff(attempt, retryConfig)
			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, &NetworkError{
					Operation: "retry_backoff",
					Message:   "context cancelled during retry backoff",
					Err:       ctx.Err(),
				}
			}
		}
	}

	// All retries exhausted
	return nil, &NetworkError{
		Operation: "http_request",
		Message:   fmt.Sprintf("all %d retry attempts exhausted", retryConfig.MaxRetries),
		Err:       lastErr,
	}
}

// calculateBackoff calculates the backoff duration for a given attempt
func (c *Client) calculateBackoff(attempt int, config *RetryConfig) time.Duration {
	backoff := float64(config.InitialBackoff)
	for i := 0; i < attempt; i++ {
		backoff *= config.BackoffMultiplier
	}

	duration := time.Duration(backoff)
	if duration > config.MaxBackoff {
		duration = config.MaxBackoff
	}

	return duration
}

// buildURL builds a complete URL from the base URL, endpoint, and query parameters
func (c *Client) buildURL(endpoint string, query url.Values) string {
	baseURL := c.config.BaseURL
	if baseURL == "" {
		// This will be overridden by service-specific clients
		baseURL = "https://api.globus.org"
	}

	u, _ := url.Parse(baseURL)
	// Join the endpoint onto the base URL's path rather than overwriting it, so
	// a base URL carrying a version prefix (e.g. .../v2, .../v0.10) is preserved.
	// Endpoints are written as absolute-looking paths ("/oauth2/token"); trim the
	// leading slash so JoinPath appends instead of treating it as already-joined.
	u = u.JoinPath(strings.TrimPrefix(endpoint, "/"))
	if query != nil {
		u.RawQuery = query.Encode()
	}

	return u.String()
}

// GetConfig returns the client configuration (read-only)
func (c *Client) GetConfig() *Config {
	return c.config
}

// Close closes the client and releases resources
// This implements the v4.2.0 context manager pattern from Python SDK
// It's safe to call Close multiple times
func (c *Client) Close() error {
	// Only close the HTTP client if we created it internally
	// If the user provided their own HTTP client, they're responsible for closing it
	if c.httpClientCreated && c.config.HTTPClient != nil {
		c.config.HTTPClient.CloseIdleConnections()
	}
	return nil
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// Transport handles HTTP communication with the API
type Transport struct {
	Client *core.Client
}

// NewTransport creates a new Transport
func NewTransport(client *core.Client) *Transport {
	return &Transport{
		Client: client,
	}
}

// Request makes an HTTP request to the API
func (t *Transport) Request(
	ctx context.Context,
	method string,
	path string,
	body interface{},
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	// Build the full URL
	urlStr := t.Client.BaseURL
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}
	urlStr += strings.TrimPrefix(path, "/")

	if query != nil && len(query) > 0 {
		urlStr += "?" + query.Encode()
	}

	// Prepare the request body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if headers != nil {
		for key, values := range headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Set content type for JSON requests with body
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Accept JSON responses
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// Send the request
	resp, err := t.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Get makes a GET request to the API
func (t *Transport) Get(
	ctx context.Context,
	path string,
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	return t.Request(ctx, http.MethodGet, path, nil, query, headers)
}

// Post makes a POST request to the API
func (t *Transport) Post(
	ctx context.Context,
	path string,
	body interface{},
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	return t.Request(ctx, http.MethodPost, path, body, query, headers)
}

// Put makes a PUT request to the API
func (t *Transport) Put(
	ctx context.Context,
	path string,
	body interface{},
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	return t.Request(ctx, http.MethodPut, path, body, query, headers)
}

// Delete makes a DELETE request to the API
func (t *Transport) Delete(
	ctx context.Context,
	path string,
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	return t.Request(ctx, http.MethodDelete, path, nil, query, headers)
}

// Patch makes a PATCH request to the API
func (t *Transport) Patch(
	ctx context.Context,
	path string,
	body interface{},
	query url.Values,
	headers http.Header,
) (*http.Response, error) {
	return t.Request(ctx, http.MethodPatch, path, body, query, headers)
}

// DecodeResponse decodes the response body into the specified type
func DecodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w (status: %d, body: %s)",
			err, resp.StatusCode, string(body))
	}

	return nil
}

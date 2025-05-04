// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// Transport handles HTTP communication with the API
type Transport struct {
	Client interfaces.ClientInterface
	Debug  bool
	Trace  bool
	Logger *log.Logger
}

// Options for configuring the transport
type Options struct {
	// Debug enables HTTP request/response logging
	Debug bool
	
	// Trace enables detailed HTTP tracing (including bodies)
	Trace bool
	
	// Logger is the logger to use for debug output
	Logger *log.Logger
}

// NewTransport creates a new Transport
func NewTransport(client interfaces.ClientInterface, options *Options) *Transport {
	// Check environment variables for debug settings
	envDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"
	envTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1"
	
	debug := envDebug
	trace := envTrace
	logger := log.New(os.Stderr, "", log.LstdFlags)
	
	// Override with options if provided
	if options != nil {
		if options.Debug {
			debug = true
		}
		if options.Trace {
			trace = true
		}
		if options.Logger != nil {
			logger = options.Logger
		}
	}
	
	// If trace is enabled, debug is also enabled
	if trace {
		debug = true
	}
	
	return &Transport{
		Client: client,
		Debug:  debug,
		Trace:  trace,
		Logger: logger,
	}
}

// DeferredTransport holds the transport configuration until a client is available
type DeferredTransport struct {
	Debug  bool
	Trace  bool
	Logger *log.Logger
}

// NewDeferredTransport creates a configuration for a transport that can be attached to a client later
func NewDeferredTransport(options *Options) *DeferredTransport {
	// Check environment variables for debug settings
	envDebug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"
	envTrace := os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1"
	
	debug := envDebug
	trace := envTrace
	logger := log.New(os.Stderr, "", log.LstdFlags)
	
	// Override with options if provided
	if options != nil {
		if options.Debug {
			debug = true
		}
		if options.Trace {
			trace = true
		}
		if options.Logger != nil {
			logger = options.Logger
		}
	}
	
	// If trace is enabled, debug is also enabled
	if trace {
		debug = true
	}
	
	return &DeferredTransport{
		Debug:  debug,
		Trace:  trace,
		Logger: logger,
	}
}

// AttachClient creates a Transport by attaching a client to a DeferredTransport
func (dt *DeferredTransport) AttachClient(client interfaces.ClientInterface) *Transport {
	return &Transport{
		Client: client,
		Debug:  dt.Debug,
		Trace:  dt.Trace,
		Logger: dt.Logger,
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
	urlStr := t.Client.GetBaseURL()
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}
	urlStr += strings.TrimPrefix(path, "/")

	if query != nil && len(query) > 0 {
		urlStr += "?" + query.Encode()
	}

	// Prepare the request body
	var bodyBytes []byte
	var bodyReader io.Reader
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
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

	// Log the request if debug is enabled
	if t.Debug {
		t.logRequest(method, urlStr, req.Header, bodyBytes)
	}

	// Send the request
	startTime := time.Now()
	resp, err := t.Client.Do(ctx, req)
	duration := time.Since(startTime)

	// Log the error if there is one
	if err != nil {
		if t.Debug {
			t.Logger.Printf("HTTP Error: %v (%s)", err, duration.Round(time.Millisecond))
		}
		return nil, err
	}

	// Log the response if debug is enabled
	if t.Debug {
		t.logResponse(resp, duration)
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

// RoundTrip implements the http.RoundTripper interface, allowing the Transport to be used
// directly with low-level HTTP operations. This is useful for debugging and testing.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use the context from the request
	ctx := req.Context()
	
	// Log the request if debug is enabled
	if t.Debug {
		var bodyBytes []byte
		if req.Body != nil {
			// Read the body and replace it
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		t.logRequest(req.Method, req.URL.String(), req.Header, bodyBytes)
	}
	
	// Send the request
	startTime := time.Now()
	resp, err := t.Client.Do(ctx, req)
	duration := time.Since(startTime)
	
	// Log the error if there is one
	if err != nil {
		if t.Debug {
			t.Logger.Printf("HTTP Error: %v (%s)", err, duration.Round(time.Millisecond))
		}
		return nil, err
	}
	
	// Log the response if debug is enabled
	if t.Debug {
		t.logResponse(resp, duration)
	}
	
	return resp, nil
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

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transport

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

// DebugTransport is an http.RoundTripper that logs requests and responses
type DebugTransport struct {
	Transport http.RoundTripper
	Debug     bool
	Trace     bool
	Logger    *log.Logger
}

// NewDebugTransport creates a new DebugTransport
func NewDebugTransport(transport http.RoundTripper) *DebugTransport {
	debug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"
	trace := os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1"

	// If trace is enabled, debug is also enabled
	if trace {
		debug = true
	}

	return &DebugTransport{
		Transport: transport,
		Debug:     debug,
		Trace:     trace,
		Logger:    log.New(os.Stderr, "", log.LstdFlags),
	}
}

// RoundTrip logs the request and response
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.Debug {
		return t.Transport.RoundTrip(req)
	}

	// Clone the request body if we're going to read it
	var reqBody []byte
	if t.Trace && req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Log the request
	if t.Trace {
		dump, _ := httputil.DumpRequestOut(req, false)
		t.Logger.Printf("HTTP Request:\n%s", sanitizeHeaders(string(dump)))

		if len(reqBody) > 0 {
			t.Logger.Printf("Request Body:\n%s", sanitizeBody(string(reqBody)))
		}
	} else {
		t.Logger.Printf("HTTP Request: %s %s", req.Method, req.URL.String())
	}

	// Make the request and time it
	start := time.Now()
	resp, err := t.Transport.RoundTrip(req)
	duration := time.Since(start)

	// Return early if there was an error
	if err != nil {
		t.Logger.Printf("HTTP Error: %v", err)
		return resp, err
	}

	// Log the response
	if t.Trace {
		dump, _ := httputil.DumpResponse(resp, false)
		t.Logger.Printf("HTTP Response:\n%s", sanitizeHeaders(string(dump)))

		// Clone and log the response body
		if resp.Body != nil {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

			if len(respBody) > 0 {
				t.Logger.Printf("Response Body:\n%s", sanitizeBody(string(respBody)))
			}
		}
	} else {
		t.Logger.Printf("HTTP Response: %d %s (%s)",
			resp.StatusCode, resp.Status, duration.Round(time.Millisecond))
	}

	// Log rate limit information if present
	t.logRateLimitInfo(resp)

	return resp, err
}

// logRateLimitInfo logs rate limit headers from the response
func (t *DebugTransport) logRateLimitInfo(resp *http.Response) {
	if !t.Debug || resp == nil {
		return
	}

	// Look for rate limit headers
	limit := resp.Header.Get("X-RateLimit-Limit")
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	if limit != "" || remaining != "" || reset != "" {
		t.Logger.Printf("Rate Limits: limit=%s remaining=%s reset=%s",
			limit, remaining, reset)
	}
}

// sanitizeHeaders removes sensitive information from headers
func sanitizeHeaders(dump string) string {
	lines := strings.Split(dump, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Authorization:") {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 3 {
				lines[i] = fmt.Sprintf("Authorization: %s [REDACTED]", parts[1])
			}
		}
	}
	return strings.Join(lines, "\n")
}

// sanitizeBody limits the length of bodies for logging
func sanitizeBody(body string) string {
	maxLen := 4096
	if len(body) > maxLen {
		return body[:maxLen] + "... [truncated]"
	}
	return body
}

// EnableHTTPDebugging enables HTTP debugging on the default HTTP client
func EnableHTTPDebugging() {
	http.DefaultTransport = NewDebugTransport(http.DefaultTransport)
}

// SetHTTPDebugOutput sets the output writer for HTTP debug logging
func SetHTTPDebugOutput(w io.Writer) {
	if transport, ok := http.DefaultTransport.(*DebugTransport); ok {
		transport.Logger = log.New(w, "", log.LstdFlags)
	}
}

// IsDebugEnabled returns true if HTTP debugging is enabled
func IsDebugEnabled() bool {
	if transport, ok := http.DefaultTransport.(*DebugTransport); ok {
		return transport.Debug
	}

	// Check environment variables
	debug := os.Getenv("GLOBUS_SDK_HTTP_DEBUG") == "1"
	trace := os.Getenv("GLOBUS_SDK_HTTP_TRACE") == "1"
	return debug || trace
}

// SetDebugLevel sets the HTTP debug level
func SetDebugLevel(debug, trace bool) {
	if transport, ok := http.DefaultTransport.(*DebugTransport); ok {
		transport.Debug = debug
		transport.Trace = trace
	} else {
		// Create a new debug transport
		http.DefaultTransport = &DebugTransport{
			Transport: http.DefaultTransport,
			Debug:     debug,
			Trace:     trace,
			Logger:    log.New(os.Stderr, "", log.LstdFlags),
		}
	}
}

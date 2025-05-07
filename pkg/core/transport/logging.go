// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package transport

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// logRequest logs an HTTP request
func (t *Transport) logRequest(method, url string, headers http.Header, body []byte) {
	if !t.Debug {
		return
	}

	if t.Trace {
		// Log detailed request
		t.Logger.Printf("HTTP Request: %s %s", method, url)
		t.Logger.Printf("Request Headers:")
		for key, values := range headers {
			// Redact Authorization header value
			if key == "Authorization" {
				t.Logger.Printf("  %s: [REDACTED]", key)
			} else {
				for _, value := range values {
					t.Logger.Printf("  %s: %s", key, value)
				}
			}
		}

		if len(body) > 0 {
			t.Logger.Printf("Request Body:")
			t.logBody(body)
		}
	} else {
		// Log basic request
		t.Logger.Printf("HTTP Request: %s %s", method, url)
	}
}

// logResponse logs an HTTP response
func (t *Transport) logResponse(resp *http.Response, duration time.Duration) {
	if !t.Debug || resp == nil {
		return
	}

	// Always log the status and duration
	t.Logger.Printf("HTTP Response: %d %s (%s)",
		resp.StatusCode, resp.Status, duration.Round(time.Millisecond))

	// Log rate limit headers if present
	t.logRateLimitHeaders(resp)

	if t.Trace {
		// Log detailed response
		t.Logger.Printf("Response Headers:")
		for key, values := range resp.Header {
			for _, value := range values {
				t.Logger.Printf("  %s: %s", key, value)
			}
		}

		// Only log the body for trace mode and if the response has a body
		if resp.Body != nil {
			// Read the body and replace it with a new reader
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

			if len(respBody) > 0 {
				t.Logger.Printf("Response Body:")
				t.logBody(respBody)
			}
		}
	}
}

// logBody logs a request or response body, truncating if it's too large
func (t *Transport) logBody(body []byte) {
	if !t.Debug || body == nil {
		return
	}

	// Truncate large bodies
	maxLen := 2048
	bodyStr := string(body)
	if len(bodyStr) > maxLen {
		t.Logger.Printf("%s... [truncated %d/%d bytes]", 
			bodyStr[:maxLen], maxLen, len(bodyStr))
	} else {
		t.Logger.Printf("%s", bodyStr)
	}
}

// logRateLimitHeaders logs rate limit headers from the response
func (t *Transport) logRateLimitHeaders(resp *http.Response) {
	if !t.Debug || resp == nil {
		return
	}

	// Look for rate limit headers
	limit := resp.Header.Get("X-RateLimit-Limit")
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	reset := resp.Header.Get("X-RateLimit-Reset")

	// Check alternative header formats
	if limit == "" {
		limit = resp.Header.Get("X-Rate-Limit-Limit")
	}
	if remaining == "" {
		remaining = resp.Header.Get("X-Rate-Limit-Remaining")
	}
	if reset == "" {
		reset = resp.Header.Get("X-Rate-Limit-Reset")
	}

	if limit != "" || remaining != "" || reset != "" {
		t.Logger.Printf("Rate Limits: limit=%s remaining=%s reset=%s",
			limit, remaining, reset)
	}
}
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package logging

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// TracingTransport is an http.RoundTripper that adds tracing capabilities
type TracingTransport struct {
	Base         http.RoundTripper
	Logger       *EnhancedLogger
	GenerateID   func() string
	RequestHook  func(*http.Request)
	ResponseHook func(*http.Response, time.Duration)
}

// RoundTrip implements the http.RoundTripper interface
func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Generate trace ID if not already set
	traceID := req.Header.Get("X-Trace-ID")
	if traceID == "" {
		if t.GenerateID != nil {
			traceID = t.GenerateID()
		} else {
			traceID = GenerateTraceID()
		}
		req.Header.Set("X-Trace-ID", traceID)
	}
	
	// Get logger with trace ID
	logger := t.Logger
	if logger != nil {
		logger = logger.WithTraceID(traceID)
	}
	
	// Log the request
	if logger != nil && logger.HasTraceEnabled() {
		logger.LogHTTPRequest(req.Method, req.URL.String(), req.Header)
	}
	
	// Call pre-request hook if provided
	if t.RequestHook != nil {
		t.RequestHook(req)
	}
	
	// Record start time
	start := time.Now()
	
	// Send the request using the base transport
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	resp, err := base.RoundTrip(req)
	
	// Record elapsed time
	elapsed := time.Since(start)
	
	if err != nil {
		// Log error
		if logger != nil {
			logger.WithField("error", err.Error()).
				WithField("elapsed_ms", elapsed.Milliseconds()).
				Error("HTTP request failed")
		}
		return resp, err
	}
	
	// Add trace ID to response
	if resp != nil && resp.Header != nil {
		resp.Header.Set("X-Trace-ID", traceID)
	}
	
	// Log the response
	if logger != nil && logger.HasTraceEnabled() {
		if resp != nil {
			logger.LogHTTPResponse(resp.StatusCode, resp.Header, elapsed)
		}
	}
	
	// Call post-response hook if provided
	if t.ResponseHook != nil && resp != nil {
		t.ResponseHook(resp, elapsed)
	}
	
	return resp, err
}

// GenerateTraceID generates a random trace ID
func GenerateTraceID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if random generation fails
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

// NewTracingTransport creates a new tracing transport
func NewTracingTransport(base http.RoundTripper, logger *EnhancedLogger) *TracingTransport {
	return &TracingTransport{
		Base:   base,
		Logger: logger,
	}
}
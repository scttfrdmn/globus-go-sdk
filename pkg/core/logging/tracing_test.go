// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package logging

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracingTransport(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create a logger with trace level
	logger := NewLogger(&Options{
		Output: &buf,
		Level:  LogLevelTrace,
		Format: FormatText,
	})

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the trace ID was added to the request
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			t.Error("Trace ID was not added to the request")
		}
		
		// Add a test header to the response
		w.Header().Set("X-Test", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a tracing transport
	transport := NewTracingTransport(http.DefaultTransport, logger)

	// Create a client with the tracing transport
	client := &http.Client{
		Transport: transport,
	}

	// Send a request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add an authorization header to test redaction
	req.Header.Set("Authorization", "Bearer secret-token")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check if the trace ID was added to the response
	traceID := resp.Header.Get("X-Trace-ID")
	if traceID == "" {
		t.Error("Trace ID was not added to the response")
	}

	// Check the log output
	output := buf.String()
	
	// Verify request logging
	if !bytes.Contains(buf.Bytes(), []byte("HTTP Request")) {
		t.Error("Request was not logged")
	}
	
	// Verify response logging
	if !bytes.Contains(buf.Bytes(), []byte("HTTP Response")) {
		t.Error("Response was not logged")
	}
	
	// Verify sensitive headers are redacted
	if bytes.Contains(buf.Bytes(), []byte("secret-token")) {
		t.Error("Authorization header was not redacted")
	}
	if !bytes.Contains(buf.Bytes(), []byte("[REDACTED]")) {
		t.Error("Authorization header was not properly redacted")
	}
}

func TestGenerateTraceID(t *testing.T) {
	// Generate a trace ID
	traceID := GenerateTraceID()
	
	// Check that it's not empty
	if traceID == "" {
		t.Error("Generated trace ID is empty")
	}
	
	// Check that it's 32 characters (16 bytes as hex)
	if len(traceID) != 32 {
		t.Errorf("Expected trace ID to be 32 characters, got %d", len(traceID))
	}
	
	// Generate another one and make sure it's different
	traceID2 := GenerateTraceID()
	if traceID == traceID2 {
		t.Error("Generated trace IDs should be unique")
	}
}
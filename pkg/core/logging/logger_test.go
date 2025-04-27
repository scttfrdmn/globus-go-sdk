// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnhancedLoggerTextFormat(t *testing.T) {
	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Create a logger with text format
	logger := NewLogger(&Options{
		Output: &buf,
		Level:  LogLevelDebug,
		Format: FormatText,
	})

	// Log some messages
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	// Check the output
	output := buf.String()
	if !strings.Contains(output, "[DEBUG] This is a debug message") {
		t.Errorf("Missing debug message in output: %s", output)
	}
	if !strings.Contains(output, "[INFO] This is an info message") {
		t.Errorf("Missing info message in output: %s", output)
	}
	if !strings.Contains(output, "[WARN] This is a warning message") {
		t.Errorf("Missing warning message in output: %s", output)
	}
	if !strings.Contains(output, "[ERROR] This is an error message") {
		t.Errorf("Missing error message in output: %s", output)
	}
}

func TestEnhancedLoggerJSONFormat(t *testing.T) {
	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Create a logger with JSON format
	logger := NewLogger(&Options{
		Output: &buf,
		Level:  LogLevelDebug,
		Format: FormatJSON,
	})

	// Log a message
	logger.Info("This is a JSON message")

	// Check the output
	output := buf.String()
	if !strings.Contains(output, "This is a JSON message") {
		t.Errorf("Missing message in output: %s", output)
	}

	// Try to parse the JSON
	var entry LogEntry
	jsonStr := strings.TrimSpace(output)
	err := json.Unmarshal([]byte(jsonStr), &entry)
	if err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	// Check the parsed JSON
	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != "This is a JSON message" {
		t.Errorf("Expected message 'This is a JSON message', got %s", entry.Message)
	}
}

func TestEnhancedLoggerWithFields(t *testing.T) {
	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Create a logger with text format
	logger := NewLogger(&Options{
		Output: &buf,
		Level:  LogLevelDebug,
		Format: FormatText,
	})

	// Add fields and log a message
	logger.WithField("user", "test").WithField("action", "login").Info("User logged in")

	// Check the output
	output := buf.String()
	if !strings.Contains(output, "[INFO] user=test action=login User logged in") {
		t.Errorf("Missing fields in output: %s", output)
	}
}

func TestEnhancedLoggerWithTraceID(t *testing.T) {
	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Create a logger with JSON format and trace ID
	logger := NewLogger(&Options{
		Output:  &buf,
		Level:   LogLevelTrace,
		Format:  FormatJSON,
		TraceID: "test-trace-id",
	})

	// Log a message
	logger.Trace("This is a traced message")

	// Check the output
	output := buf.String()
	
	// Try to parse the JSON
	var entry LogEntry
	jsonStr := strings.TrimSpace(output)
	err := json.Unmarshal([]byte(jsonStr), &entry)
	if err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	// Check the trace ID
	if entry.TraceID != "test-trace-id" {
		t.Errorf("Expected trace_id 'test-trace-id', got %s", entry.TraceID)
	}
}

func TestEnhancedLoggerLevelFiltering(t *testing.T) {
	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Create a logger with info level
	logger := NewLogger(&Options{
		Output: &buf,
		Level:  LogLevelInfo,
		Format: FormatText,
	})

	// Log messages at different levels
	logger.Debug("This should be filtered out")
	logger.Info("This should appear")
	logger.Error("This should also appear")

	// Check the output
	output := buf.String()
	if strings.Contains(output, "This should be filtered out") {
		t.Errorf("Debug message should be filtered out: %s", output)
	}
	if !strings.Contains(output, "This should appear") {
		t.Errorf("Missing info message in output: %s", output)
	}
	if !strings.Contains(output, "This should also appear") {
		t.Errorf("Missing error message in output: %s", output)
	}
}
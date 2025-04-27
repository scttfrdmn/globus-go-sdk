// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// LogLevel defines the level of logging
type LogLevel int

const (
	// LogLevelNone disables logging
	LogLevelNone LogLevel = iota
	// LogLevelError logs only errors
	LogLevelError
	// LogLevelWarn logs warnings and errors
	LogLevelWarn
	// LogLevelInfo logs information, warnings, and errors
	LogLevelInfo
	// LogLevelDebug logs debug information and all above
	LogLevelDebug
	// LogLevelTrace logs trace information and all above (including request/response details)
	LogLevelTrace
)

// Format defines the log format
type Format int

const (
	// FormatText outputs logs in plain text format
	FormatText Format = iota
	// FormatJSON outputs logs in JSON format
	FormatJSON
)

// EnhancedLogger implements core.Logger interface with additional features
type EnhancedLogger struct {
	logger       *log.Logger
	level        LogLevel
	format       Format
	traceEnabled bool
	fields       map[string]interface{}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Options defines options for creating an EnhancedLogger
type Options struct {
	Output  io.Writer
	Level   LogLevel
	Format  Format
	TraceID string
	Fields  map[string]interface{}
}

// NewLogger creates a new enhanced logger
func NewLogger(opts *Options) *EnhancedLogger {
	if opts == nil {
		opts = &Options{
			Output: os.Stderr,
			Level:  LogLevelInfo,
			Format: FormatText,
		}
	}

	if opts.Output == nil {
		opts.Output = os.Stderr
	}

	fields := make(map[string]interface{})
	if opts.Fields != nil {
		for k, v := range opts.Fields {
			fields[k] = v
		}
	}

	if opts.TraceID != "" {
		fields["trace_id"] = opts.TraceID
	}

	return &EnhancedLogger{
		logger:       log.New(opts.Output, "", log.LstdFlags),
		level:        opts.Level,
		format:       opts.Format,
		traceEnabled: opts.TraceID != "",
		fields:       fields,
	}
}

// Debug logs debug information
func (l *EnhancedLogger) Debug(format string, v ...interface{}) {
	if l.level >= LogLevelDebug {
		l.log("DEBUG", format, v...)
	}
}

// Info logs information
func (l *EnhancedLogger) Info(format string, v ...interface{}) {
	if l.level >= LogLevelInfo {
		l.log("INFO", format, v...)
	}
}

// Warn logs warnings
func (l *EnhancedLogger) Warn(format string, v ...interface{}) {
	if l.level >= LogLevelWarn {
		l.log("WARN", format, v...)
	}
}

// Error logs errors
func (l *EnhancedLogger) Error(format string, v ...interface{}) {
	if l.level >= LogLevelError {
		l.log("ERROR", format, v...)
	}
}

// Trace logs trace information
func (l *EnhancedLogger) Trace(format string, v ...interface{}) {
	if l.level >= LogLevelTrace {
		l.log("TRACE", format, v...)
	}
}

// WithField adds a field to the logger
func (l *EnhancedLogger) WithField(key string, value interface{}) *EnhancedLogger {
	newLogger := &EnhancedLogger{
		logger:       l.logger,
		level:        l.level,
		format:       l.format,
		traceEnabled: l.traceEnabled,
		fields:       make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *EnhancedLogger) WithFields(fields map[string]interface{}) *EnhancedLogger {
	newLogger := &EnhancedLogger{
		logger:       l.logger,
		level:        l.level,
		format:       l.format,
		traceEnabled: l.traceEnabled,
		fields:       make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithTraceID creates a new logger with a trace ID
func (l *EnhancedLogger) WithTraceID(traceID string) *EnhancedLogger {
	newLogger := &EnhancedLogger{
		logger:       l.logger,
		level:        l.level,
		format:       l.format,
		traceEnabled: true,
		fields:       make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add trace ID
	newLogger.fields["trace_id"] = traceID

	return newLogger
}

// SetLevel sets the log level
func (l *EnhancedLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetFormat sets the log format
func (l *EnhancedLogger) SetFormat(format Format) {
	l.format = format
}

// log outputs a log message
func (l *EnhancedLogger) log(level string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)

	if l.format == FormatText {
		// Plain text format
		prefix := fmt.Sprintf("[%s]", level)
		
		// Add fields if available
		if len(l.fields) > 0 {
			prefix += " "
			for k, v := range l.fields {
				prefix += fmt.Sprintf("%s=%v ", k, v)
			}
		}
		
		l.logger.Printf("%s %s", prefix, message)
	} else {
		// JSON format
		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     level,
			Message:   message,
			Fields:    l.fields,
		}

		if traceID, ok := l.fields["trace_id"]; ok {
			entry.TraceID = fmt.Sprintf("%v", traceID)
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			// Fall back to plain text if JSON marshaling fails
			l.logger.Printf("[%s] Error marshaling log entry to JSON: %v", level, err)
			return
		}

		l.logger.Println(string(jsonData))
	}
}

// AsCore converts the EnhancedLogger to a core.Logger
// This is needed for compatibility with the existing interfaces
func (l *EnhancedLogger) AsCore() core.Logger {
	return l
}

// GetTraceID returns the trace ID if set
func (l *EnhancedLogger) GetTraceID() string {
	if traceID, ok := l.fields["trace_id"]; ok {
		return fmt.Sprintf("%v", traceID)
	}
	return ""
}

// HasTraceEnabled returns true if tracing is enabled
func (l *EnhancedLogger) HasTraceEnabled() bool {
	return l.traceEnabled && l.level >= LogLevelTrace
}

// LogHTTPRequest logs an HTTP request if tracing is enabled
func (l *EnhancedLogger) LogHTTPRequest(method, url string, headers map[string][]string) {
	if !l.HasTraceEnabled() {
		return
	}

	// Create a copy of headers with sensitive information redacted
	redactedHeaders := make(map[string][]string)
	for k, v := range headers {
		redactedValues := make([]string, len(v))
		for i, value := range v {
			if k == "Authorization" || k == "X-Auth-Token" {
				redactedValues[i] = "[REDACTED]"
			} else {
				redactedValues[i] = value
			}
		}
		redactedHeaders[k] = redactedValues
	}

	fields := map[string]interface{}{
		"http_method": method,
		"http_url":    url,
		"headers":     redactedHeaders,
	}

	l.WithFields(fields).Trace("HTTP Request")
}

// LogHTTPResponse logs an HTTP response if tracing is enabled
func (l *EnhancedLogger) LogHTTPResponse(statusCode int, headers map[string][]string, elapsed time.Duration) {
	if !l.HasTraceEnabled() {
		return
	}

	fields := map[string]interface{}{
		"http_status":     statusCode,
		"headers":         headers,
		"elapsed_ms":      elapsed.Milliseconds(),
		"elapsed_seconds": elapsed.Seconds(),
	}

	l.WithFields(fields).Trace("HTTP Response")
}
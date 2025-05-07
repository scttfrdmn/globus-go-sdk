// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package logging

import (
	"io"
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
)

// WithEnhancedLogger sets an enhanced logger for the client
func WithEnhancedLogger(logger *EnhancedLogger) core.ClientOption {
	return func(c *core.Client) {
		c.Logger = logger
	}
}

// WithLogOutput sets the output for the enhanced logger
func WithLogOutput(output io.Writer) core.ClientOption {
	return func(c *core.Client) {
		// If we already have an enhanced logger, update its output
		if logger, ok := c.Logger.(*EnhancedLogger); ok {
			// Create a new logger with the same settings but different output
			newLogger := NewLogger(&Options{
				Output:  output,
				Level:   LogLevel(logger.level),
				Format:  logger.format,
				TraceID: logger.GetTraceID(),
				Fields:  logger.fields,
			})
			c.Logger = newLogger
		} else {
			// Create a new enhanced logger with the specified output
			c.Logger = NewLogger(&Options{
				Output: output,
				Level:  LogLevelInfo,
				Format: FormatText,
			})
		}
	}
}

// WithLogLevel sets the log level for the client's logger
func WithLogLevel(level LogLevel) core.ClientOption {
	return func(c *core.Client) {
		// If we already have an enhanced logger, update its level
		if logger, ok := c.Logger.(*EnhancedLogger); ok {
			logger.SetLevel(level)
		} else if _, ok := c.Logger.(*core.DefaultLogger); ok {
			// If it's a default logger, set its level
			c.Logger = core.NewDefaultLogger(nil, core.LogLevel(level))
		} else {
			// Create a new enhanced logger with the specified level
			c.Logger = NewLogger(&Options{
				Level:  level,
				Format: FormatText,
			})
		}
	}
}

// WithJSONLogging enables JSON-formatted logging
func WithJSONLogging() core.ClientOption {
	return func(c *core.Client) {
		// If we already have an enhanced logger, update its format
		if logger, ok := c.Logger.(*EnhancedLogger); ok {
			logger.SetFormat(FormatJSON)
		} else {
			// Create a new enhanced logger with JSON format
			c.Logger = NewLogger(&Options{
				Level:  LogLevelInfo,
				Format: FormatJSON,
			})
		}
	}
}

// WithTracing enables request/response tracing
func WithTracing(traceID string) core.ClientOption {
	return func(c *core.Client) {
		// If no trace ID is provided, generate one
		if traceID == "" {
			traceID = GenerateTraceID()
		}

		// Create or get the enhanced logger
		var logger *EnhancedLogger
		if existing, ok := c.Logger.(*EnhancedLogger); ok {
			logger = existing.WithTraceID(traceID)
		} else {
			logger = NewLogger(&Options{
				Level:   LogLevelTrace,
				Format:  FormatText,
				TraceID: traceID,
			})
		}

		// Set up the HTTP client with tracing
		transport := c.HTTPClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}

		// Create a tracing transport
		tracingTransport := NewTracingTransport(transport, logger)
		c.HTTPClient.Transport = tracingTransport

		// Update the logger
		c.Logger = logger
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

/*
Package logging provides structured logging and HTTP tracing utilities for the Globus Go SDK.

This package extends the basic logging support in pkg/core with structured,
leveled logging in both plain-text and JSON formats, distributed trace ID
propagation, and an http.RoundTripper wrapper that automatically logs HTTP
request/response details while redacting sensitive header values.

# STABILITY: BETA

This package is in beta. The core logging abstractions are stable, but
output format details and some tracing-related APIs may change in minor
releases. Changes will be documented in the CHANGELOG with migration guidance.

The following components are considered beta-stable:

  - LogLevel constants (LogLevelNone, LogLevelError, LogLevelWarn, LogLevelInfo,
    LogLevelDebug, LogLevelTrace)
  - Format constants (FormatText, FormatJSON)
  - Options struct and all exported fields (Output, Level, Format, TraceID, Fields)
  - LogEntry struct and fields (Timestamp, Level, Message, TraceID, Fields)
  - EnhancedLogger type and constructor (NewLogger)
  - EnhancedLogger logging methods (Debug, Info, Warn, Error, Trace)
  - EnhancedLogger builder methods (WithField, WithFields, WithTraceID)
  - EnhancedLogger configuration methods (SetLevel, SetFormat)
  - EnhancedLogger utility methods (AsCore, GetTraceID, HasTraceEnabled)
  - EnhancedLogger HTTP tracing methods (LogHTTPRequest, LogHTTPResponse)
  - TracingTransport type and constructor (NewTracingTransport)
  - TracingTransport fields (Base, Logger, GenerateID, RequestHook, ResponseHook)
  - TracingTransport.RoundTrip method (implements http.RoundTripper)
  - GenerateTraceID utility function

# Compatibility Guarantees

For beta components:
  - Minor backward-incompatible changes may still occur in minor releases
  - JSON log output format details may be refined
  - Significant efforts will be made to maintain backward compatibility
  - Changes will be clearly documented in the CHANGELOG
  - Deprecated functionality will be marked with appropriate notices

# Basic Usage

Create a structured logger with JSON output:

	logger := logging.NewLogger(&logging.Options{
		Output: os.Stderr,
		Level:  logging.LogLevelInfo,
		Format: logging.FormatJSON,
	})
	logger.Info("SDK initialized")

Add structured fields to log entries:

	requestLogger := logger.WithField("service", "transfer").
		WithField("endpoint_id", endpointID)
	requestLogger.Debug("submitting transfer task")

Enable distributed tracing:

	tracedLogger := logger.WithTraceID(logging.GenerateTraceID())
	tracedLogger.Info("handling request")

Wrap an HTTP transport with automatic request/response logging:

	tracingTransport := logging.NewTracingTransport(http.DefaultTransport, tracedLogger)
	httpClient := &http.Client{Transport: tracingTransport}
*/
package logging

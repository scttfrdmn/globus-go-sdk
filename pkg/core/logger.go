// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"io"
	"log"
	"os"
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
)

// Logger defines the interface for logging
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// DefaultLogger implements the Logger interface
type DefaultLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(out io.Writer, level LogLevel) *DefaultLogger {
	if out == nil {
		out = os.Stderr
	}
	return &DefaultLogger{
		logger: log.New(out, "", log.LstdFlags),
		level:  level,
	}
}

// Debug logs debug information
func (l *DefaultLogger) Debug(format string, v ...interface{}) {
	if l.level >= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs information
func (l *DefaultLogger) Info(format string, v ...interface{}) {
	if l.level >= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warnings
func (l *DefaultLogger) Warn(format string, v ...interface{}) {
	if l.level >= LogLevelWarn {
		l.logger.Printf("[WARN] "+format, v...)
	}
}

// Error logs errors
func (l *DefaultLogger) Error(format string, v ...interface{}) {
	if l.level >= LogLevelError {
		l.logger.Printf("[ERROR] "+format, v...)
	}
}

// WithLogger sets a custom logger
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithLogLevel sets the log level for the default logger
func WithLogLevel(level LogLevel) ClientOption {
	return func(c *Client) {
		if c.Logger == nil {
			c.Logger = NewDefaultLogger(nil, level)
		} else if dl, ok := c.Logger.(*DefaultLogger); ok {
			dl.level = level
		}
	}
}

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Logging and Tracing Guide

This guide explains how to use the enhanced logging and tracing capabilities in the Globus Go SDK.

## Overview

The Globus Go SDK provides a flexible and powerful logging system that supports:

- Multiple log formats (text and JSON)
- Various log levels for controlling verbosity
- Structured logging with arbitrary fields
- HTTP request and response tracing
- Distributed tracing with trace IDs

## Basic Usage

### Enabling Logging

You can enable logging when creating a client by using the appropriate client options:

```go
import (
    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
)

// Create a configuration with logging enabled
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelDebug))

// Create a client that will use this logging configuration
client := config.NewTransferClient(accessToken)
```

### Log Levels

The SDK supports the following log levels, from least to most verbose:

- `LogLevelNone`: Disables all logging
- `LogLevelError`: Logs only errors
- `LogLevelWarn`: Logs warnings and errors
- `LogLevelInfo`: Logs information, warnings, and errors (default)
- `LogLevelDebug`: Logs debug information and all above
- `LogLevelTrace`: Logs trace information, including HTTP requests and responses

Example:

```go
// Enable debug logging
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelDebug))
```

### Log Formats

The SDK supports two log formats:

1. **Text format** (default): Simple text logs with level prefixes
2. **JSON format**: Structured JSON logs for machine processing

To enable JSON logging:

```go
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithJSONLogging())
```

## Advanced Logging

### Creating a Custom Logger

You can create a fully customized logger:

```go
import (
    "os"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
)

// Create a custom logger
logger := logging.NewLogger(&logging.Options{
    Output:  os.Stdout,               // Output destination
    Level:   logging.LogLevelDebug,   // Log level
    Format:  logging.FormatJSON,      // Log format
    TraceID: "my-custom-trace-id",    // Optional trace ID
    Fields: map[string]interface{}{   // Additional fields for all log entries
        "application": "my-app",
        "version":     "1.0.0",
    },
})

// Use the custom logger with a client
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithEnhancedLogger(logger))
```

### Contextual Logging with Fields

You can add contextual information to logs using fields:

```go
// Create a basic logger
logger := logging.NewLogger(&logging.Options{
    Output: os.Stdout,
    Level:  logging.LogLevelDebug,
})

// Add context for a specific operation
operationLogger := logger.WithFields(map[string]interface{}{
    "operation": "file-transfer",
    "user_id":   "user-123",
})

// Log with context
operationLogger.Info("Starting file transfer")
```

This will produce logs with the additional fields included:

Text format:
```
[INFO] operation=file-transfer user_id=user-123 Starting file transfer
```

JSON format:
```json
{
  "timestamp": "2025-04-26T15:30:45Z",
  "level": "INFO",
  "message": "Starting file transfer",
  "fields": {
    "operation": "file-transfer",
    "user_id": "user-123"
  }
}
```

### Logging to File

You can send logs to a file or any other `io.Writer`:

```go
// Open a log file
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Create a config that logs to the file
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogOutput(file))
```

### Multi-target Logging

For more advanced logging scenarios, you can log to multiple targets:

```go
// Create a multi-writer
multiWriter := io.MultiWriter(os.Stdout, logFile)

// Create a logger that writes to both stdout and a file
logger := logging.NewLogger(&logging.Options{
    Output: multiWriter,
    Level:  logging.LogLevelInfo,
})
```

## Tracing

Tracing provides detailed information about HTTP requests and responses, which is particularly useful for debugging API interactions and performance issues.

### Enabling Tracing

You can enable tracing using the `WithTracing` client option:

```go
// Enable tracing with an automatically generated trace ID
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelTrace)).
    WithClientOption(logging.WithTracing(""))

// Or with a specific trace ID
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelTrace)).
    WithClientOption(logging.WithTracing("my-trace-id"))
```

When tracing is enabled, the SDK will log:
- HTTP request method, URL, and headers (with sensitive headers redacted)
- HTTP response status code and headers
- Request/response timing information

### Trace IDs and Distributed Tracing

Trace IDs allow you to track a single operation across multiple components or services. The SDK automatically:

1. Generates a unique trace ID if one isn't provided
2. Adds the trace ID to HTTP request headers as `X-Trace-ID`
3. Extracts trace IDs from response headers
4. Includes the trace ID in all log entries related to the operation

To manually work with trace IDs:

```go
// Generate a trace ID
traceID := logging.GenerateTraceID()

// Create a logger with this trace ID
logger := logging.NewLogger(&logging.Options{
    TraceID: traceID,
    Level:   logging.LogLevelTrace,
})

// Use this logger in a component
component.DoWork(logger)

// Pass the trace ID to another service
headers.Set("X-Trace-ID", traceID)
```

## Best Practices

### Production Settings

For production environments:

```go
// Production logging setup
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelInfo)). // Less verbose
    WithClientOption(logging.WithJSONLogging())                   // Structured format
```

### Development Settings

For development:

```go
// Development logging setup
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelDebug)). // More verbose
    WithClientOption(logging.WithTracing(""))                      // Enable tracing
```

### Security Considerations

- The SDK automatically redacts sensitive information from logs (like auth tokens)
- Still, be careful not to log sensitive data in your application
- Consider where logs are stored and who has access to them

## Example

Here's a complete example demonstrating various logging and tracing features:

```go
package main

import (
    "context"
    "os"

    "github.com/scttfrdmn/globus-go-sdk/pkg"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
)

func main() {
    // Get access token
    accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
    if accessToken == "" {
        panic("GLOBUS_ACCESS_TOKEN environment variable is required")
    }

    // Create a logger with custom fields
    logger := logging.NewLogger(&logging.Options{
        Output: os.Stdout,
        Level:  logging.LogLevelTrace,
        Format: logging.FormatJSON,
        Fields: map[string]interface{}{
            "application": "my-transfer-app",
            "version":     "1.0.0",
        },
    })

    // Enable tracing with a fixed trace ID for this operation
    traceID := "transfer-operation-123"
    tracingLogger := logger.WithTraceID(traceID)

    // Create a configuration with the logger
    config := pkg.NewConfigFromEnvironment().
        WithClientOption(logging.WithEnhancedLogger(tracingLogger)).
        WithClientOption(logging.WithTracing(traceID))

    // Create transfer client
    transferClient := config.NewTransferClient(accessToken)

    // Add operation context to the logger
    opLogger := tracingLogger.WithFields(map[string]interface{}{
        "operation": "list-endpoints",
    })
    opLogger.Info("Starting endpoint listing operation")

    // Perform the operation
    ctx := context.Background()
    endpoints, err := transferClient.ListEndpoints(ctx, nil)
    
    if err != nil {
        opLogger.WithField("error", err.Error()).Error("Failed to list endpoints")
        return
    }

    // Log success
    opLogger.WithField("endpoint_count", len(endpoints.DATA)).Info("Successfully listed endpoints")
}
```

## Related Documentation

- [User Guide](user-guide.md)
- [Error Handling](error-handling.md)
- [Examples](../examples/logging/README.md)
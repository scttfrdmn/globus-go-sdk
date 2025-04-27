<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Logging and Tracing Example

This example demonstrates the enhanced logging and tracing capabilities of the Globus Go SDK.

## Features

- **Text Logging**: Simple text-based logging with different log levels
- **JSON Logging**: Structured JSON logging for machine processing
- **Request/Response Tracing**: Detailed HTTP request and response tracing
- **Trace IDs**: Distributed tracing across multiple operations

## Prerequisites

- A Globus account
- A Globus access token with the appropriate scopes

## Running the Example

1. Set the access token as an environment variable:

```bash
export GLOBUS_ACCESS_TOKEN=your-access-token
```

2. Run the example:

```bash
go run main.go
```

## Understanding the Example

### Text Logging

The text logging example demonstrates how to enable debug-level logging with a simple text format. This is useful for development and debugging.

```go
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelDebug))
```

### JSON Logging

The JSON logging example shows how to use structured JSON logging, which is better for machine processing, log aggregation, and analysis tools.

```go
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelDebug)).
    WithClientOption(logging.WithJSONLogging())
```

### Request/Response Tracing

The tracing example demonstrates how to enable detailed HTTP request and response tracing. This is particularly useful for debugging API interactions and performance issues.

```go
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogLevel(logging.LogLevelTrace)).
    WithClientOption(logging.WithTracing("example-trace-id"))
```

## Available Log Levels

The SDK supports the following log levels:

- `LogLevelNone`: Disables all logging
- `LogLevelError`: Logs only errors
- `LogLevelWarn`: Logs warnings and errors
- `LogLevelInfo`: Logs information, warnings, and errors
- `LogLevelDebug`: Logs debug information and all above
- `LogLevelTrace`: Logs trace information, including HTTP requests and responses

## Custom Logger Configuration

You can create a fully customized logger:

```go
// Create a custom logger
logger := logging.NewLogger(&logging.Options{
    Output:  os.Stdout,       // Output destination
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

## Contextual Logging

You can add contextual information to logs:

```go
// Create a logger with context
logger := logging.NewLogger(&logging.Options{
    Output: os.Stdout,
    Level:  logging.LogLevelDebug,
})

// Add fields for a specific operation
operationLogger := logger.WithFields(map[string]interface{}{
    "operation": "file-transfer",
    "user_id":   "user-123",
})

// Log with context
operationLogger.Info("Starting file transfer")
```

## Log Output Customization

You can customize where logs are sent:

```go
// Send logs to a file
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Create a config that logs to the file
config := pkg.NewConfigFromEnvironment().
    WithClientOption(logging.WithLogOutput(file))
```
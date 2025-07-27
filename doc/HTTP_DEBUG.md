<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# HTTP Debugging for Integration Tests

This document provides guidance on how to enable and use HTTP debugging when running integration tests against the Globus API.

## Enabling HTTP Debug Logging

The SDK supports HTTP debug logging to help troubleshoot API communication issues. To enable debug logging, you can use the following environment variables:

```bash
export GLOBUS_SDK_HTTP_DEBUG=1           # Enable basic HTTP logging
export GLOBUS_SDK_HTTP_TRACE=1           # Enable detailed HTTP tracing (including request/response bodies)
export GLOBUS_TEST_LOG_LEVEL=debug       # Set log level to debug for tests
```

### Debug Levels

1. **Basic HTTP Logging (`GLOBUS_SDK_HTTP_DEBUG=1`)**
   - Logs request method, URL, and response status code
   - Does not log request or response bodies
   - Useful for tracking API call order and response codes

2. **Detailed HTTP Tracing (`GLOBUS_SDK_HTTP_TRACE=1`)**
   - Logs everything in basic mode plus:
   - Full request and response headers
   - Request and response bodies (truncated for large payloads)
   - Timing information for each request

## Adding HTTP Debug to the Core Client

The core HTTP client has been enhanced to support debugging. The implementation can be found in `pkg/core/transport/transport.go`.

Here's how to enhance an existing client to add HTTP debugging:

```go
// Create a transport with debugging
transportOptions := &transport.Options{
    Debug:  true,           // Basic debugging
    Trace:  true,           // Full request/response tracing
    Logger: myLoggerInstance,
}
httpTransport := transport.NewTransport(transportOptions)

// Apply it to an existing client
clientOptions := []core.ClientOption{
    core.WithHTTPTransport(httpTransport),
}

// Create client with debug transport
client := core.NewClient(clientOptions...)
```

## Integration Test Debug Configuration

When running integration tests, you can enable HTTP debugging by setting environment variables:

```bash
# Run with basic HTTP debugging
GLOBUS_SDK_HTTP_DEBUG=1 go test -tags=integration ./pkg/services/transfer

# Run with full HTTP tracing
GLOBUS_SDK_HTTP_TRACE=1 go test -tags=integration ./pkg/services/transfer

# Run a specific test with debugging
GLOBUS_SDK_HTTP_DEBUG=1 go test -tags=integration -run=TestIntegration_ListEndpoints ./pkg/services/transfer
```

## Analyzing Debug Output

When HTTP debugging is enabled, you'll see output similar to:

```
2025/04/29 21:45:02 HTTP Request: GET https://transfer.api.globus.org/v0.10/endpoint_search?limit=5
2025/04/29 21:45:02 HTTP Response: 400 Bad Request (398 ms)
```

With tracing enabled, you'll see full headers and bodies:

```
2025/04/29 21:45:02 HTTP Request: GET https://transfer.api.globus.org/v0.10/endpoint_search?limit=5
2025/04/29 21:45:02 Request Headers:
  Authorization: Bearer [REDACTED]
  Accept: application/json
  User-Agent: globus-go-sdk/0.2.0

2025/04/29 21:45:02 HTTP Response: 400 Bad Request (398 ms)
2025/04/29 21:45:02 Response Headers:
  Content-Type: application/json
  Server: nginx
  x-transfer-api-error: ClientError.BadRequest

2025/04/29 21:45:02 Response Body:
{
  "code": "BadRequest",
  "message": "Invalid filter parameter",
  "resource": "/endpoint_search",
  "request_id": "kUUJqMoH9"
}
```

## Common API Issues

When debugging integration tests, look for these common issues:

1. **400 Bad Request**
   - Incorrect URL path or query parameters
   - Missing required fields in request body
   - Malformed JSON in request body

2. **401 Unauthorized**
   - Missing or expired access token
   - Token does not have required scopes

3. **403 Forbidden**
   - Token has insufficient permissions for the requested resource
   - Resource access is restricted

4. **404 Not Found**
   - Incorrect API endpoint path
   - Resource does not exist
   - API version mismatch

5. **429 Too Many Requests**
   - Rate limit exceeded
   - Check the `x-ratelimit-remaining` and `x-ratelimit-reset` headers

## Next Steps

1. Enhance the core client to automatically log API errors with proper context
2. Add support for selectively enabling debug logging per service client
3. Implement sensitive data redaction for logged values
4. Add structured logging with request correlation IDs
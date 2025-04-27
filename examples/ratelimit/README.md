# Rate Limiting and Backoff Example

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This example demonstrates the rate limiting, backoff, and circuit breaker functionality in the Globus Go SDK.

## Features

- Token bucket rate limiting
- Adaptive rate limiting based on API response headers
- Exponential backoff with jitter for retries
- Circuit breaker pattern for fault tolerance
- Comprehensive testing modes

## Usage

### Prerequisites

Before running the example, you need:

1. Globus account credentials
2. Globus client ID and client secret from the Globus Developers portal

Set the required environment variables:

```bash
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
```

### Running the Demo Mode

```bash
go run main.go --mode=demo --concurrency=10 --duration=30 --rate=5.0
```

This runs a comprehensive demo that shows all rate limiting features working together.

### Running the Rate Limit Test

```bash
go run main.go --mode=ratelimit --concurrency=5 --duration=60 --rate=3.0
```

This tests the rate limiter against the actual Globus API, showing how it adapts to server-side rate limits.

### Running the Backoff Test

```bash
go run main.go --mode=backoff
```

This demonstrates exponential backoff with simulated API failures.

### Running the Circuit Breaker Test

```bash
go run main.go --mode=circuit
```

This demonstrates the circuit breaker pattern for handling service degradation.

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--mode` | "demo" | Mode to run: demo, ratelimit, backoff, circuit |
| `--concurrency` | 10 | Number of concurrent requests |
| `--duration` | 10 | Test duration in seconds |
| `--rate` | 5.0 | Requests per second |
| `--token` | "" | Globus access token (if not provided, will use auth flow) |

## Understanding the Output

### Demo Mode

The demo mode shows:

1. Rate limiting in action with multiple concurrent requests
2. Automatic retries with exponential backoff
3. Circuit breaker opening and closing based on failure patterns
4. Real-time statistics on request rates and successes

Example output:

```
=== Running Rate Limit and Backoff Demo ===
Concurrency: 10, Duration: 10s, Rate: 5.0 req/sec
[5s] Requests: 20, Success: 18, Rate: 4.0 req/sec, Retries: 2, Circuit Opens: 0
  Rate Limiter: Limit=5.0, Remaining=3.2, Throttled=4
```

### Backoff Test

The backoff test shows the progression of retry delays:

```
=== Running Backoff Strategy Test ===
Attempt 1/4...
  Failed with temporary error
Attempt 2/4...
  Failed with temporary error
Attempt 3/4...
  Failed with temporary error
Attempt 4/4...
  Success!
Final result: Success after 4 attempts
```

### Circuit Breaker Test

The circuit breaker test shows state transitions:

```
=== Running Circuit Breaker Test ===
Phase 1: Normal operation
Request 1: Normal operation (success)

*** Circuit state changed: CLOSED -> OPEN ***

Phase 3: Circuit is open, requests should fail fast
Request 1: Correctly rejected (circuit open), took 215.833µs

*** Circuit state changed: OPEN -> HALF-OPEN ***

Phase 5: Testing half-open state
First request in half-open state (success)

*** Circuit state changed: HALF-OPEN -> CLOSED ***
```

## How It Works

### Rate Limiting

The rate limiter uses a token bucket algorithm:

1. Tokens accumulate at a fixed rate (e.g., 5 per second)
2. Each request consumes one token
3. When no tokens are available, requests wait
4. Adaptive mode adjusts based on server response headers

### Backoff Strategy

The exponential backoff strategy:

1. Starts with a small initial delay (e.g., 100ms)
2. Doubles the delay after each failed attempt
3. Adds random jitter to prevent thundering herd problems
4. Caps at a maximum delay (e.g., 60 seconds)

### Circuit Breaker

The circuit breaker:

1. Starts in the CLOSED state (allows all requests)
2. Opens after a threshold of consecutive failures
3. In the OPEN state, fails fast without making API calls
4. After a timeout, enters HALF-OPEN state to test
5. Returns to CLOSED after successful test requests

## Best Practices

When using these features in your application:

1. Configure rate limits conservatively to avoid hitting service limits
2. Use appropriate retry strategies based on the API's characteristics
3. Set appropriate thresholds for circuit breakers
4. Monitor and log rate limiting statistics
5. Consider service-specific customization of retry policies
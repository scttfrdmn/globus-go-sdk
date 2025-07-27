<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Error Recovery Patterns

This example demonstrates advanced error recovery patterns for applications using the Globus Go SDK. It showcases robust error handling techniques suitable for production deployments.

## Overview

Error handling in distributed systems is complex due to various failure modes:
- Transient network issues
- Service outages and maintenance
- Rate limiting
- Resource exhaustion
- Authentication/authorization failures
- Context cancellations and timeouts

This example demonstrates practical patterns for building resilient applications that can gracefully handle these failure scenarios.

## Implemented Patterns

### 1. Circuit Breaker

The circuit breaker pattern prevents cascading failures by "opening the circuit" when a service appears to be failing. This example implements:

- Configurable failure thresholds
- Half-open state for testing service recovery
- Automatic reset after configurable cool-down period
- Manual override capabilities

### 2. Exponential Backoff with Jitter

This example demonstrates sophisticated retry logic:

- Configurable initial retry delay
- Exponential increase in delay between retries
- Jitter to prevent thundering herd problem
- Maximum retry limit and timeout
- Error classification to determine if retry is appropriate

### 3. Graceful Degradation

The example shows how to maintain partial functionality during service disruptions:

- Feature flags to disable non-critical functionality
- Fallback to local caching when services are unavailable
- Progressive enhancement based on service availability
- Clear user communication during degraded operation

### 4. Authentication Failure Recovery

Robust handling of authentication issues:

- Automatic token refresh on expiration
- Fallback authentication methods
- Token revocation and reacquisition
- Session recovery after credential refresh

### 5. Resource Management

Proper management of resources during failures:

- Connection pool health monitoring
- Idle connection cleanup
- Background health checks
- Proactive resource management

## Usage

To run this example:

```bash
cd examples/error-recovery
go run .
```

By default, the example will simulate various error conditions to demonstrate recovery patterns. You can modify the behavior with flags:

```bash
# Run with specific error simulation
go run . -simulate network-partition

# Run with real services (requires Globus credentials)
go run . -use-real-services

# Set failure probability (0.0-1.0)
go run . -failure-rate 0.3
```

## Code Structure

- `main.go` - Main application
- `circuit.go` - Circuit breaker implementation
- `backoff.go` - Retry logic
- `degradation.go` - Graceful degradation strategies
- `auth_recovery.go` - Authentication failure recovery
- `resource_mgmt.go` - Resource management during failures
- `monitor.go` - Error monitoring and metrics collection

## Further Reading

- [Resilience in Distributed Systems](https://docs.globus.org/developer-tools/go-sdk/resilience) - Globus documentation
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html) - Martin Fowler
- [Exponential Backoff and Jitter](https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/) - AWS Architecture Blog
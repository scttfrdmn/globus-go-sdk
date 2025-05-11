# Contract Testing in the Globus Go SDK

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

This document explains the contract testing system implemented as part of Phase 2 of the API Stability Implementation Plan.

## Overview

Contract testing verifies that implementations of interfaces adhere to their behavioral expectations (the "contract") beyond just satisfying the method signatures. While type checking ensures that implementations provide the required methods with the correct signatures, contract testing ensures they behave correctly.

## Why Contract Testing?

1. **API Stability**: Ensures consistent behavior across releases
2. **Documentation**: Serves as executable documentation for interface behavioral requirements
3. **Alternative Implementations**: Helps users create compatible alternative implementations
4. **Breaking Change Detection**: Identifies behavioral changes that might not be caught by type checking
5. **Interface Evolution**: Provides a framework for evolving interfaces while maintaining compatibility

## Contract Test Structure

Each core interface in the SDK has a corresponding contract test that verifies:

1. **Method Behavior**: Does each method behave according to its contract?
2. **Error Handling**: Are errors returned in the expected situations?
3. **Thread Safety**: Does the implementation handle concurrent access correctly?
4. **Resource Management**: Are resources properly managed and released?
5. **Context Handling**: Does the implementation respect context cancellation?

### Example: ClientInterface Contract

The `ClientInterface` contract test verifies that any implementation:

- Returns a non-nil HTTP client from `GetHTTPClient()`
- Returns a non-empty string from `GetBaseURL()`
- Returns a non-empty string from `GetUserAgent()`
- Returns a non-nil logger from `GetLogger()`
- Respects context cancellation in the `Do` method
- Returns an error when given a nil request
- Returns consistent values for configuration getters

## Using Contract Tests

### 1. Testing SDK Implementations

The SDK's own implementations are tested against these contracts to ensure they behave correctly:

```go
func TestClientImplementationContract(t *testing.T) {
    client := core.NewClient(/* options... */)
    contracts.VerifyClientContract(t, client)
}
```

### 2. Testing Alternative Implementations

Users who create alternative implementations can use these contract tests to verify compatibility:

```go
func TestMyCustomClientContract(t *testing.T) {
    client := NewMyCustomClient()
    contracts.VerifyClientContract(t, client)
}
```

### 3. Creating Custom Contract Tests

Users can create their own contract tests for custom interfaces following the same pattern:

```go
// Define contract test for custom interface
func VerifyMyServiceContract(t *testing.T, service MyServiceInterface) {
    t.Helper()
    
    t.Run("MethodA", func(t *testing.T) {
        // Test method behavior
    })
    
    // ... test other methods
}

// Use it in tests
func TestMyServiceImplementation(t *testing.T) {
    service := NewMyService()
    VerifyMyServiceContract(t, service)
}
```

## Mock Implementations

The `contracts` package includes mock implementations of all core interfaces for use in testing:

- `MockClient`: A simple `ClientInterface` implementation
- `MockTransport`: A simple `Transport` implementation
- `MockAuthorizer`: A simple `Authorizer` implementation
- `MockLogger`: A simple `Logger` implementation
- etc.

These mocks implement the interface contracts and can be used in your own tests:

```go
func TestMyCode(t *testing.T) {
    client := contracts.NewMockClient()
    // Use the mock client in your tests
}
```

## Covered Interfaces

Contract tests are provided for the following interfaces:

| Interface | Description | Contract Test |
|-----------|-------------|--------------|
| `ClientInterface` | Core client functionality | `VerifyClientContract` |
| `Transport` | HTTP transport functionality | `VerifyTransportContract` |
| `ConnectionPool` | Connection pool functionality | `VerifyConnectionPoolContract` |
| `ConnectionPoolManager` | Connection pool management | `VerifyConnectionPoolManagerContract` |
| `Authorizer` | Authorization functionality | `VerifyAuthorizerContract` |
| `TokenManager` | Token management and refresh | `VerifyTokenManagerContract` |
| `Logger` | Logging functionality | `VerifyLoggerContract` |

## Best Practices

### When Writing Contract Tests

1. **Focus on Behavior**: Test the behavior described in the documentation, not implementation details
2. **Test Edge Cases**: Include tests for error conditions, resource exhaustion, etc.
3. **Use Clear Test Names**: Make it obvious what behavior is being tested
4. **Minimize Dependencies**: Avoid dependencies on external services
5. **Document Requirements**: Clearly document the behavioral requirements being tested

### When Implementing Interfaces

1. **Read the Contract Tests**: Understand the behavioral expectations
2. **Test Early and Often**: Run contract tests during development
3. **Focus on Compatibility**: Ensure compatibility with existing clients
4. **Document Deviations**: If you must deviate from the contract, document it clearly
5. **Extend, Don't Modify**: Add new methods rather than changing existing ones

## Integration with API Stability Plan

Contract testing is part of Phase 2 of the API Stability Implementation Plan and works together with:

1. **API Stability Indicators**: Package stability levels (STABLE, BETA, etc.)
2. **API Compatibility Tools**: Tools to detect API changes between versions
3. **Deprecation System**: System for marking and tracking deprecated features
4. **CI Integration**: Automated verification of API compatibility

By enforcing behavioral contracts, we ensure that the SDK's API remains stable and predictable across releases, while still allowing for evolution and improvement.

## Future Enhancements

Planned enhancements to the contract testing system include:

1. **Automated Contract Generation**: Generate contract tests from interface documentation
2. **Performance Contracts**: Verify performance characteristics
3. **Contract Visualization**: Visualize interface contracts and relationships
4. **Test Coverage Analysis**: Analyze coverage of interface contracts
5. **Contract Versioning**: Support for versioned contracts to track evolution

## Conclusion

Contract testing provides a foundation for API stability by ensuring that implementations adhere to their behavioral contracts. By documenting and testing these contracts, we can maintain compatibility while evolving the SDK to meet new requirements.

The contract testing system is available in the `pkg/core/contracts` package and is designed to be used by both SDK developers and users implementing custom interfaces.
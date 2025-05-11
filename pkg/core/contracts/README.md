# Contract Tests for Core Interfaces

This package provides contract tests for the core interfaces defined in the Globus Go SDK. Contract tests verify that implementations of these interfaces adhere to the behavioral contracts required by the interfaces.

## Purpose

Interface contracts extend beyond method signatures to include behavioral expectations. These tests ensure that any implementation of an interface not only provides the required methods but also behaves according to specifications. This is crucial for API stability, as users may depend on these behaviors.

## Usage

Each interface has a dedicated contract test that can be used to verify any implementation. To use these contract tests:

1. Create an implementation of the interface you want to test
2. Call the appropriate test function, passing your implementation and a testing.T instance

Example:

```go
// In your test file
func TestMyClientImplementationContract(t *testing.T) {
    myClient := NewMyClient()
    client_contract.VerifyClientContract(t, myClient)
}
```

## Covered Interfaces

Contract tests are provided for the following interfaces:

- `ClientInterface`: Core client behavior
- `ConnectionPool`: HTTP connection pool behavior
- `ConnectionPoolManager`: Connection pool management
- `Transport`: HTTP transport behavior
- `Authorizer`: Authorization mechanism
- `TokenManager`: Token management and refresh
- `Logger`: Logging functionality

## Contract Definitions

Each contract test explicitly defines the behavioral expectations for the interface. These expectations include:

- Error handling behavior
- Thread safety requirements
- Timeout handling
- Resource management
- Response format validation
- State transitions
- Side effects

## Mocks and Fixtures

The package includes mock implementations and test fixtures to facilitate testing. These mocks can also be used by SDK users when writing their own tests.

## Integration with API Stability

These contract tests are part of the API Stability Implementation Plan and serve several purposes:

1. Document behavioral expectations for interface implementers
2. Verify SDK implementations adhere to their contracts
3. Prevent accidental breaking changes to interface behavior
4. Provide a reference for users implementing interfaces

Contract tests run as part of the CI pipeline to ensure continued compliance across releases.
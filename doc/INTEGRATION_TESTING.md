# Integration Testing Guide for Globus Go SDK

This document outlines the approach for integration testing the Globus Go SDK, which verifies that the SDK correctly interacts with the actual Globus API services.

## Prerequisites

Before running integration tests, you'll need:

1. A Globus account with access to the services being tested
2. A registered client application in Globus Auth with appropriate scopes
3. Environment variables set up with credentials and test resources

## Environment Setup

Set the following environment variables before running integration tests:

```bash
# Required for all tests
export GLOBUS_TEST_CLIENT_ID="your-client-id"
export GLOBUS_TEST_CLIENT_SECRET="your-client-secret"

# Required for Transfer tests
export GLOBUS_TEST_SOURCE_ENDPOINT_ID="your-source-endpoint-id"
export GLOBUS_TEST_DEST_ENDPOINT_ID="your-destination-endpoint-id"

# Required for Groups tests
export GLOBUS_TEST_GROUP_ID="existing-group-id"  # Optional, will create one if not provided
```

## Running Integration Tests

Integration tests are in a separate package with the suffix `_integration_test.go` and are organized by service. They're skipped by default in normal test runs to avoid external dependencies.

To run integration tests:

```bash
# Run all integration tests
go test ./... -tags=integration

# Run just Auth integration tests
go test ./pkg/services/auth -tags=integration

# Run specific test
go test ./pkg/services/transfer -run TestIntegration_SubmitTransfer -tags=integration
```

## Test Organization

Integration tests follow these principles:

1. **Self-contained**: Each test should create and clean up its own resources when possible
2. **Graceful degradation**: Tests should skip (not fail) if credentials are missing
3. **Realistic scenarios**: Tests should model real-world usage patterns
4. **Comprehensive coverage**: Test all key API interactions
5. **Idempotency**: Tests should be repeatable without side effects

## Test Structure

Each service has an integration test file with the following structure:

1. **Setup code**: Functions to create test resources and check environment
2. **Teardown code**: Functions to clean up resources after tests
3. **Helper functions**: Common code used by multiple tests
4. **Test cases**: Individual tests for each API feature

## Writing Integration Tests

When writing integration tests:

1. Follow the naming convention `TestIntegration_<FunctionName>`
2. Check for required environment variables and skip if missing
3. Use descriptive error messages
4. Clean up resources in defer statements
5. Allow for reasonable timing variations in asynchronous operations
6. Add meaningful assertions that verify correct behavior

## Example Test Structure

```go
func TestIntegration_TransferClient(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Check for required env vars
    clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
    clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
    if clientID == "" || clientSecret == "" {
        t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
    }
    
    // Create test resources
    // ...
    
    // Clean up afterwards
    defer func() {
        // Cleanup code
    }()
    
    // Run test
    // ...
    
    // Verify results
    // ...
}
```

## Continuous Integration

Integration tests are marked with build tags so they can be selectively run in CI environments where credentials are available. They are not run on every PR but can be manually triggered for significant changes.

## Best Practices

1. **Be respectful of API rate limits**
2. **Never hardcode credentials**
3. **Avoid testing in production when possible**
4. **Consider graceful retries for intermittent failures**
5. **Focus on end-to-end workflows rather than individual calls**
6. **Maintain test independence**

## Supporting Files

Look for helper files in the `testdata` directory that provide:

1. Template requests for creating test resources
2. Expected response structures 
3. Test file contents for transfer tests

## Development Workflow

1. Start with unit tests for all components
2. Add integration tests for critical paths
3. Run integration tests on significant changes
4. Verify real-world usage with example applications
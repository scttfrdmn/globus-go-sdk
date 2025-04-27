<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Integration Testing Guide

This document provides guidance on setting up and running integration tests for the Globus Go SDK.

## Overview

Integration tests for the Globus Go SDK verify that our code works correctly against the actual Globus API endpoints. These tests ensure that:

1. Our API implementations match the expected behavior of Globus services
2. Authentication and token handling work correctly
3. Transfer operations complete as expected
4. Error handling captures and processes API errors properly

## Prerequisites

To run the integration tests, you need:

1. A Globus account with appropriate permissions
2. A Globus client application with the necessary scopes
3. Access to at least one Globus endpoint for transfer tests
4. Go installed on your system (version 1.18 or later)

## Setting Up Credentials

Integration tests require real Globus API credentials. These should be provided as environment variables:

### Required Variables

- `GLOBUS_TEST_CLIENT_ID`: Your Globus client ID
- `GLOBUS_TEST_CLIENT_SECRET`: Your Globus client secret

### Optional Variables (Recommended for Transfer Tests)

- `GLOBUS_TEST_SOURCE_ENDPOINT_ID`: ID of a source endpoint for transfer tests
- `GLOBUS_TEST_DESTINATION_ENDPOINT_ID`: ID of a destination endpoint for transfer tests
- `GLOBUS_TEST_SOURCE_PATH`: Path on the source endpoint (default: `/globus-test`)
- `GLOBUS_TEST_DESTINATION_PATH`: Path on the destination endpoint (default: `/globus-test`)

### Optional Variables (For Group Tests)

- `GLOBUS_TEST_GROUP_ID`: ID of a Globus group for testing group operations
- `GLOBUS_TEST_USER_ID`: ID of a user for testing group membership operations

## Using an Environment File

You can create a `.env.test` file in the project root to store your credentials:

```
GLOBUS_TEST_CLIENT_ID=your-client-id
GLOBUS_TEST_CLIENT_SECRET=your-client-secret
GLOBUS_TEST_SOURCE_ENDPOINT_ID=your-source-endpoint-id
GLOBUS_TEST_DESTINATION_ENDPOINT_ID=your-destination-endpoint-id
GLOBUS_TEST_SOURCE_PATH=/globus-test
GLOBUS_TEST_DESTINATION_PATH=/globus-test
GLOBUS_TEST_USER_ID=your-user-id
GLOBUS_TEST_GROUP_ID=your-group-id
```

⚠️ **IMPORTANT: Never commit this file to the repository.** The `.env.test` file is included in `.gitignore` to prevent accidental exposure of credentials.

## Running the Tests

### Using the Script

The easiest way to run integration tests is using the provided script:

```bash
# Run all integration tests
./scripts/run_integration_tests.sh

# Run tests for a specific package
./scripts/run_integration_tests.sh pkg/services/transfer

# Run tests for a specific test function
./scripts/run_integration_tests.sh pkg/services/transfer TestIntegration_ResumableTransfer
```

The script automatically:
1. Loads environment variables from `.env.test` if present
2. Checks that required variables are set
3. Runs the specified tests with the appropriate tags

### Manual Execution

Alternatively, you can run the tests manually:

```bash
# Set environment variables (if not using .env.test)
export GLOBUS_TEST_CLIENT_ID=your-client-id
export GLOBUS_TEST_CLIENT_SECRET=your-client-secret
# ... set other variables as needed

# Run all integration tests
go test -v -tags=integration ./...

# Run tests for a specific package
go test -v -tags=integration ./pkg/services/transfer/...

# Run a specific test
go test -v -tags=integration ./pkg/services/transfer -run TestIntegration_ResumableTransfer
```

## Writing Integration Tests

When writing new integration tests:

1. Use the `_test` suffix in the test file name
2. Use the `//go:build integration` build tag at the top of the file
3. Name test functions with the `TestIntegration_` prefix
4. Use the `getTestCredentials` function to get credentials
5. Include cleanup code to remove any created resources
6. Make tests skip gracefully if required credentials are missing

Example structure:

```go
//go:build integration
package mypackage_test

import (
    "testing"
    "context"
    "os"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/myservice"
)

func TestIntegration_MyFeature(t *testing.T) {
    // Get credentials
    clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
    clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
    
    if clientID == "" || clientSecret == "" {
        t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
    }
    
    // Create client and run test
    // ...
    
    // Clean up resources
    defer func() {
        // Delete any created resources
    }()
}
```

## Test Data Safety

To avoid data loss or unintended consequences:

1. Always use test-specific paths and resources
2. Include timestamps in test resource names to avoid conflicts
3. Clean up all created resources when tests complete
4. Use read-only operations where possible
5. Never modify or delete data outside of your test directories

## CI/CD Integration

When running in CI/CD environments:

1. Store credentials as secure environment variables
2. Consider using dedicated test credentials with limited permissions
3. Set up separate test endpoints for automated testing
4. Ensure cleanup runs even if tests fail

## Troubleshooting

If integration tests fail:

1. Check that your credentials are correct and not expired
2. Verify that endpoints are accessible and activated
3. Check if you have the necessary permissions
4. Look for rate limiting or service availability issues
5. Check the Globus status page for service disruptions

## Related Documentation

- [Globus API Documentation](https://docs.globus.org/api/)
- [Go Testing Package](https://golang.org/pkg/testing/)
- [Development Guide](DEVELOPMENT.md)
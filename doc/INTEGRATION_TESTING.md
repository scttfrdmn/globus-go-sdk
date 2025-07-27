<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Integration Testing with the Globus Go SDK

This document explains how to set up and run integration tests for the Globus Go SDK. Integration tests allow you to validate that the SDK functions correctly against the real Globus APIs.

## Overview

The Globus Go SDK includes integration tests for:
- Auth Service
- Transfer Service
- Groups Service
- Search Service
- Flows Service
- Compute Service

These tests verify that the SDK can perform common operations against the actual Globus services using real credentials.

## Prerequisites

Before running integration tests, you need:

1. A Globus account with access to the services you want to test
2. A Globus client ID and client secret (for authenticating to Globus services)
3. Access to Globus endpoints for testing transfer functionality
4. Go 1.18 or higher installed

## Setting Up Credentials

Integration tests require credentials to access Globus services. You can provide these in two ways:

### Option 1: Environment Variables

Set the following environment variables:

```bash
export GLOBUS_TEST_CLIENT_ID="your-client-id"
export GLOBUS_TEST_CLIENT_SECRET="your-client-secret"

# Required variables for transfer tests
export GLOBUS_TEST_SOURCE_ENDPOINT_ID="source-endpoint-id"
export GLOBUS_TEST_DEST_ENDPOINT_ID="destination-endpoint-id"

# Optional variables for group tests
export GLOBUS_TEST_GROUP_ID="group-id"
export GLOBUS_TEST_USER_ID="user-id"
```

### Option 2: .env.test File (Recommended)

Create a `.env.test` file in the project root with these variables:

```
GLOBUS_TEST_CLIENT_ID=your-client-id
GLOBUS_TEST_CLIENT_SECRET=your-client-secret

# Required variables for transfer tests
GLOBUS_TEST_SOURCE_ENDPOINT_ID=source-endpoint-id
GLOBUS_TEST_DEST_ENDPOINT_ID=destination-endpoint-id

# Optional variables for group tests
GLOBUS_TEST_GROUP_ID=group-id
GLOBUS_TEST_USER_ID=user-id
```

> **IMPORTANT**: Add `.env.test` to your `.gitignore` file to prevent accidentally committing credentials to version control.

### Setting Up Test System Information

For transfer tests, the SDK uses the following:

1. Test system (named "terror") with UUID: 20b46e7f-230d-11f0-9913-0affeb91e4e5
2. Local test data directory: `/Users/scttfrdmn/globus-test`

Create and populate this directory with test files before running transfer integration tests.

## Verifying Your Setup

Before running all integration tests, verify that your setup is correct by running:

```bash
./scripts/run_integration_tests.sh --verify
```

This will:

1. Verify that your credentials are valid
2. Check that the specified endpoints are accessible (if provided)
3. Confirm that test directories exist
4. Validate that your environment is properly configured

If the verification passes, you're ready to run the full test suite.

## Running Integration Tests

### Run All Integration Tests

To run all integration tests:

```bash
./scripts/run_integration_tests.sh
```

### Run Tests for a Specific Service

To run tests for a specific service:

```bash
./scripts/run_integration_tests.sh pkg/services/auth
./scripts/run_integration_tests.sh pkg/services/transfer
./scripts/run_integration_tests.sh pkg/services/groups
./scripts/run_integration_tests.sh pkg/services/search
./scripts/run_integration_tests.sh pkg/services/flows
./scripts/run_integration_tests.sh pkg/services/compute
```

### Run Specific Tests

To run tests that match a specific pattern:

```bash
./scripts/run_integration_tests.sh pkg/services/transfer Transfer
```

## Integration Test Coverage and Release Requirements

The following tests must pass against the real Globus API for a production release:

### Auth Package
- `TestIntegration_ClientCredentialsFlow`: Verifies token acquisition using client credentials
- `TestIntegration_TokenUtils`: Tests token validation and expiry checking
- `TestIntegration_ClientCredentialsAuthorizer`: Tests authorization header generation

### Transfer Package
- `TestIntegration_ListEndpoints`: Tests endpoint listing functionality
- `TestIntegration_TransferFlow`: Tests a complete transfer workflow (directory creation, file transfer, cleanup)
- `TestIntegration_RecursiveTransfer`: Tests recursive directory transfer
- `TestIntegration_TaskManagement`: Tests task creation, monitoring, and cancellation
- `TestComprehensiveTransfer`: End-to-end test of all transfer functionality

The integration tests cover the following services:

| Service | Coverage Status | Required Variables | Tests |
|---------|----------------|-------------------|-------|
| Auth    | Complete       | `GLOBUS_TEST_CLIENT_ID`, `GLOBUS_TEST_CLIENT_SECRET` | Client credentials, token introspection, authorization |
| Transfer | Complete      | Above + `GLOBUS_TEST_SOURCE_ENDPOINT_ID`, `GLOBUS_TEST_DEST_ENDPOINT_ID` | Endpoint operations, directory management, file transfers, task monitoring |
| Groups  | Complete       | Above + `GLOBUS_TEST_GROUP_ID` (optional) | Group listing, creation, membership management |
| Search  | Complete       | Above + specific index permission | Index operations, search queries, result handling |
| Flows   | Complete       | Above | Flow definition, execution, monitoring |
| Compute | Complete       | Above | Compute job submission and monitoring |
| Timers  | Partial        | Above | Basic timer operations |

## Writing New Integration Tests

When writing new integration tests:

1. Add the integration build tag to your test file:
```go
//go:build integration
// +build integration
```

2. Make sure your tests check for required environment variables
3. Skip tests gracefully when optional variables are missing
4. Clean up any resources created during tests
5. Add appropriate timeouts to avoid tests hanging

## Integration Test Implementation Details

Integration tests run against the live Globus services, so they:

1. Are slower than unit tests
2. May incur usage charges in some cases
3. Require valid credentials
4. Can be affected by network conditions or service status

For these reasons, integration tests are not run in CI pipelines by default.

### Enhanced Error Handling

Our integration tests include robust error handling features:

1. **Rate limiting**: Tests handle API rate limits using exponential backoff
2. **Authentication errors**: Clear messaging for credential issues
3. **Network resilience**: Retries for transient network failures
4. **Proper cleanup**: Ensures test resources are removed even when tests fail
5. **Contextualized errors**: Detailed error messages with context for troubleshooting

### Test Data Creation

For Transfer tests, the SDK:

1. Creates timestamped directories to avoid conflicts between test runs
2. Generates unique test files with verifiable content
3. Properly cleans up all created resources after tests complete
4. Verifies successful operations before proceeding to the next step

## Troubleshooting

If integration tests fail:

1. Verify your credentials using the `--verify` flag
2. Check that you have the necessary permissions for the resources you're trying to access
3. Ensure your Globus endpoints are activated
4. Verify that test paths exist and are readable/writable
5. Check network connectivity to Globus services
6. Look for rate limit errors (may need to wait before retrying)
7. Check for endpoint activation requirements

For detailed logs, set the following environment variable:

```bash
export GLOBUS_TEST_LOG_LEVEL=debug
```

## Status for v0.2.0 Release

For the v0.2.0 release:

1. All core integration tests are implemented and passing against real Globus services
2. Authentication uses client credentials flow with proper error handling
3. Transfer operations include robust rate limiting and retry logic
4. Error handling has been standardized across all service clients
5. All tests include proper cleanup of resources
6. The SDK handles API-specified rate limits appropriately

Integration testing with real credentials is a release requirement to ensure the SDK functions correctly in production environments.
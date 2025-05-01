<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Comprehensive Testing Guide

This document outlines the comprehensive testing approach for the Globus Go SDK. It explains how to properly validate the SDK with real Globus credentials before release.

## Overview

Testing the Globus Go SDK thoroughly requires a multi-faceted approach:

1. **Unit Testing**: Tests individual components without external dependencies
2. **Integration Testing**: Tests interactions with real Globus services
3. **Example Validation**: Ensures examples work with real credentials
4. **Credential Verification**: Validates all required scopes and permissions

## Prerequisites

To run comprehensive tests, you'll need:

1. **Globus Credentials**:
   - Client ID
   - Client Secret
   - Appropriate scopes for all services
   - Access to test endpoints for Transfer

2. **Environment Setup**:
   - Go 1.18 or later
   - A `.env.test` file with your credentials
   - Sufficient permissions on your filesystem

## Setting Up the .env.test File

Create a `.env.test` file in the project root with the following structure:

```
GLOBUS_CLIENT_ID=your_client_id
GLOBUS_CLIENT_SECRET=your_client_secret
GLOBUS_ACCESS_TOKEN=optional_access_token
GLOBUS_REFRESH_TOKEN=optional_refresh_token
SOURCE_ENDPOINT_ID=optional_source_endpoint_for_transfer_tests
DEST_ENDPOINT_ID=optional_destination_endpoint_for_transfer_tests
GLOBUS_SEARCH_INDEX_ID=optional_search_index_id
GLOBUS_FLOW_ID=optional_flow_id
```

## Running Comprehensive Tests

The SDK provides a comprehensive testing script that runs all tests, examples, and validation:

```bash
./scripts/comprehensive_testing.sh
```

This script will:

1. Run all unit tests
2. Verify all examples compile
3. Test token management with real credentials
4. Run the credential verification tool
5. Run integration tests for all services
6. Test specific examples with real credentials
7. Run linting checks
8. Perform security scanning

All results are logged to `comprehensive_testing.log`.

## Manual Testing of Key Components

In addition to the automated tests, you should manually test the following key components:

### 1. Authentication Flows

Test different authentication flows:

```bash
# Test client credentials flow
go run cmd/examples/auth/main.go --client-credentials

# Test authorization code flow (requires browser)
go run cmd/examples/auth/main.go --auth-code
```

### 2. Transfer Operations

Test file transfer:

```bash
# Test endpoint listing
go run cmd/examples/transfer/main.go --list-endpoints

# Test file transfer (requires endpoint IDs)
go run cmd/examples/transfer/main.go --source-endpoint $SOURCE_ENDPOINT_ID --destination-endpoint $DEST_ENDPOINT_ID --transfer
```

### 3. Search Operations

Test search functionality:

```bash
# Test index listing
go run cmd/examples/search/main.go --list-indexes

# Test search query (requires index ID)
go run cmd/examples/search/main.go --index $GLOBUS_SEARCH_INDEX_ID --query "test"
```

### 4. Groups Operations

Test groups functionality:

```bash
# Test group listing
go run cmd/examples/groups/main.go --list
```

### 5. Flows Operations

Test flows functionality:

```bash
# Test flow listing
go run cmd/examples/flows/main.go --list

# Test flow running (requires flow ID)
go run cmd/examples/flows/main.go --run-flow $GLOBUS_FLOW_ID
```

### 6. Token Management

Test token management:

```bash
# Test token management with real credentials
cd examples/token-management && ./test_tokens.sh
```

## Testing Error Handling

It's important to test error scenarios:

1. Invalid credentials
2. Network disconnections
3. Rate limiting
4. Timeouts
5. Permission issues

## Reporting Test Results

When reporting test results:

1. Document any failures in detail
2. Include the environment used for testing
3. Note any workarounds or special configurations
4. Track performance metrics where relevant

## Pre-Release Testing Checklist

Before each release:

1. Run the comprehensive testing script
2. Verify all examples work with real credentials
3. Test any new features or changed functionality manually
4. Verify all documentation is up-to-date
5. Ensure all integration tests pass
6. Check that all linting passes without warnings

## Continuous Integration

While the comprehensive tests require real credentials and are typically run manually before releases, a subset of tests should be run in CI:

1. Unit tests
2. Linting
3. Compilation checks
4. Code coverage analysis

## Adding New Tests

When adding new functionality:

1. Add unit tests for all new code
2. Add integration tests for service interactions
3. Update relevant examples
4. Add manual testing instructions to this guide
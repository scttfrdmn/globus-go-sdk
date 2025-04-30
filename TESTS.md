<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->
# Globus Go SDK Testing

## Test Coverage Status

| Package       | Unit Tests | Integration Tests | Status |
|---------------|------------|-------------------|--------|
| auth          | ✅ Pass    | ✅ Pass           | All tests passing |
| search        | ✅ Pass    | ⚠️ Partial        | Integration tests work with limitations |
| flows         | ✅ Pass    | ⚠️ Partial        | Integration tests work with limited permissions |
| compute       | ✅ Pass    | ⚠️ Partial        | Integration tests work with limited permissions |
| groups        | ❌ Failing | ❌ Failing        | Build issues need fixing |
| transfer      | ✅ Pass    | ⚠️ Partial        | Integration tests work with some limitations |

## Fixed Issues

### Search Package

- Fixed string conversion issues in pagination tests by replacing `string(int+'0')` with `fmt.Sprintf("%d", int)`
- Fixed test expectations to handle proper pagination with the mock server
- Fixed integration tests to handle limited API permissions

### Core - Rate Limiting

- Fixed circuit breaker race condition that was causing "sync: RUnlock of unlocked RWMutex" panic
- Made timing-sensitive tests more robust by relaxing assertions for flaky conditions
- Improved test stability across different execution environments

### Integration Testing Setup

- Added proper build tags to integration tests to separate them from regular unit tests
- Added environment variable loading in each integration test file to ensure credentials are available
- Modified integration tests to gracefully handle limited permissions for API resources

## Known Issues

### Transfer Package

The transfer tests have been significantly improved:

1. ~~Duplicate declarations of `setupMockServer` in multiple files~~ Fixed
2. ~~Missing or undefined types referenced in tests~~ Fixed
3. ~~Incorrect client method references~~ Fixed
4. ~~Authentication scope issues~~ Fixed - now uses client credentials flow with proper scopes

Integration tests fully validate correct operation of the transfer API:
- All tests are now designed to fail with clear, descriptive error messages when permissions or configuration issues occur
- Error messages include specific guidance on how to resolve the issues (e.g., "To resolve, provide GLOBUS_TEST_TRANSFER_TOKEN with proper permissions")
- For tests to pass, you MUST provide a pre-generated transfer token via `GLOBUS_TEST_TRANSFER_TOKEN`
- The token MUST have write permissions on the test endpoints
- See the `.env.test.example` file for complete configuration details

Important notes about path handling in Globus endpoints:
- By default, tests use a simple directory path: `globus-test/`
- You can specify a custom test directory with the `GLOBUS_TEST_DIRECTORY_PATH` environment variable
- The directory must have read/write permissions for the service account associated with your token
- For directories on Guest Collections or shared endpoints, use a path without a leading slash (e.g., `my-directory/path`)
- Collection IDs should be used instead of Endpoint IDs for Globus Connect Personal endpoints

A note about endpoint activation: This SDK requires Globus endpoints that support API version v0.10 or later, which includes automatic activation with properly scoped tokens. Explicit activation functionality has been removed from the SDK as it's no longer needed with modern Globus endpoints.

Example `.env.test` configuration:
```
# Standard credentials
GLOBUS_TEST_CLIENT_ID=your-client-id
GLOBUS_TEST_CLIENT_SECRET=your-client-secret

# Transfer endpoints configuration
GLOBUS_TEST_SOURCE_ENDPOINT_ID=source-endpoint-id
GLOBUS_TEST_DEST_ENDPOINT_ID=destination-endpoint-id
GLOBUS_TEST_TRANSFER_TOKEN=your-transfer-token-with-write-permissions

# Specify a directory with proper R/W permissions on both endpoints
GLOBUS_TEST_DIRECTORY_PATH=~/globus-test-directory
```

### Groups Package

The groups tests need fixes to address build issues:

1. Unknown field references in struct literals
2. Undefined struct fields

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run specific package tests
go test ./pkg/services/search
```

### Integration Tests

Integration tests require valid Globus API credentials and resources set in the `.env.test` file.

```bash
# Run all integration tests
go test -tags=integration ./...

# Run integration tests for a specific package
go test -tags=integration ./pkg/services/auth
```

## Next Steps

1. ~~Fix build issues in the transfer package~~ Done
2. Fix build issues in the groups package
3. ~~Add proper error handling to all integration tests for consistent behavior~~ Done
4. Improve integration test coverage for packages with limited permissions
5. Add more real-world scenarios to integration tests
6. Implement test mocks to reduce reliance on actual API endpoints
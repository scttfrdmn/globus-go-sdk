<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

# Disabled Tests Summary Report

This document provides a summary of all remaining disabled test files in the Globus Go SDK codebase as of May 2, 2025. It includes analysis of each disabled file, patterns in the disabled tests, and recommendations for fixing them.

## Overview of Disabled Test Files

The following `.disabled` test files remain in the codebase:

| Package | Disabled Test File | Type | Issue Pattern |
|---------|-------------------|------|---------------|
| Transfer | `resumable_test.go.disabled` | Unit | Checkpoint storage implementation |
| Transfer | `streaming_iterator_test.go.disabled` | Unit | Iterator implementation |
| Transfer | `memory_optimized_test.go.disabled` | Unit | Memory optimization features |
| Transfer | `resumable_integration_test.go.disabled` | Integration | Resumable transfer against real API |
| Transfer | `integration_test.go.disabled` | Integration | Core transfer operations against real API |
| Auth | `integration_test.go.disabled` | Integration | Authentication against real Auth API |

## Analysis By Pattern

### 1. Unit Tests with External Dependencies

Several disabled tests have dependencies on external resources:

- `resumable_test.go.disabled` tests checkpoint file storage with filesystem operations
- `memory_optimized_test.go.disabled` tests memory optimization features with variable load

**Common Issues:**
- File system operations that may not be portable across environments
- Reliance on local temporary directories 
- Tests that depend on specific system resources

### 2. Integration Tests Requiring Real Credentials

Several disabled integration tests need real Globus credentials:

- `auth/integration_test.go.disabled`
- `transfer/integration_test.go.disabled`
- `transfer/resumable_integration_test.go.disabled`

**Common Issues:**
- Tests require client credentials for access to Globus services
- Tests need specific endpoints with appropriate permissions
- Tests create and manipulate real resources that may incur charges

### 3. Complex API Behavior Tests

Some tests deal with complex API behaviors:

- `streaming_iterator_test.go.disabled` manages paged API responses
- `resumable_integration_test.go.disabled` handles resumable transfers with checkpointing

**Common Issues:**
- Difficult to mock complex API behavior
- Tests may be flaky due to timing issues or network conditions
- May require specific test data setup

## Common Patterns in Fixes Applied to Other Tests

Based on the FIXES.md documents and other fixed code, I observed these common patterns in previously fixed tests:

1. **Error Handling Improvements**:
   - Enhanced error checking functions to be aware of both specific error types and generic core errors
   - Added capability to recognize errors based on HTTP status codes
   - Added support for `core.IsNotFound` for 404 errors

2. **URL Path Handling**:
   - More robust path matching in test servers
   - Improved URL path handling by explicitly checking for specific paths
   - Added debugging information to track requests and responses

3. **Test Flakiness Fixes**:
   - Made timing-sensitive tests more lenient in their assertions
   - Replaced hard assertions with logging for cases where exact timing can't be guaranteed
   - Modified state change validation to be more flexible

4. **String Conversion Fixes**:
   - Replaced expressions using `string(int+'0')` with proper string formatting using `fmt.Sprintf()`
   - Fixed patterns that generated compiler warnings

5. **Race Condition Fixes**:
   - Simplified implementations to use more straightforward locking patterns
   - Removed complex locking patterns that led to race conditions
   - Ensured proper mutex handling in all state transition code paths

## Recommendations for Each Disabled Test

### 1. Transfer Service

#### `resumable_test.go.disabled`
**Recommendation:** Enable this test by:
- Moving file operations to a mock filesystem implementation that can be used in tests
- Using a test-specific temporary directory pattern as seen in other tests
- Implementing proper cleanup in defer blocks

#### `streaming_iterator_test.go.disabled`
**Recommendation:** Enable this test by:
- Creating a more robust mock server for API pagination
- Implementing the fixes related to URL path handling found in the flows package
- Making timeout-related assertions more flexible

#### `memory_optimized_test.go.disabled`
**Recommendation:** Enable this test by:
- Creating a more controlled test environment for memory testing
- Using benchmarks instead of standard tests for memory optimization validation
- Implementing the changes from `memory_optimized.go.fixed` file

#### `resumable_integration_test.go.disabled` and `integration_test.go.disabled`
**Recommendation:** Enable these tests by:
- Following the patterns in other successfully enabled integration tests
- Using the retry with backoff pattern for rate-limited operations
- Implementing environment variable checks and fallbacks as seen in other integration tests
- Adding proper cleanup of test resources

### 2. Auth Service

#### `auth/integration_test.go.disabled`
**Recommendation:** Enable this test by:
- Fixing the compilation error with the undefined `err` variable on line 50
- Adding proper imports for the rate limiting package
- Implementing the retry with backoff pattern as seen in other integration tests
- Using the same credential verification pattern as other enabled integration tests

## Implementation Plan

1. **First Priority**: Fix auth integration tests, as they're the foundation for other service tests
   - Fix undefined variable errors
   - Implement proper retry patterns
   - Test with client credentials flow

2. **Second Priority**: Enable transfer unit tests
   - Start with resumable_test.go and streaming_iterator_test.go
   - Apply lessons from flows and search fixes
   - Fix filesystem dependencies

3. **Third Priority**: Enable transfer integration tests
   - Apply patterns from other successful integration tests
   - Implement proper resource cleanup
   - Add fallback mechanisms for credentials

## Conclusion

The disabled tests fall into patterns related to external dependencies, credential requirements, and complex API behaviors. By applying the fix patterns observed in previously fixed tests, we can systematically enable these remaining tests. This will improve the overall test coverage of the SDK and ensure more reliable behavior when interacting with Globus services.

The most critical tests to fix are the auth integration tests since authentication is a fundamental requirement for all other services. Once those are working reliably, we can proceed with enabling the transfer tests, which make up the majority of the remaining disabled tests.
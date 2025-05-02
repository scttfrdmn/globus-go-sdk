# Next Tasks for Globus Go SDK v0.8.0

Based on our investigation, here are the priority tasks for fixing the remaining issues with the Globus Go SDK:

## 1. Fix Transfer Package Tests

There are several disabled test files that need to be fixed and re-enabled:

- ✅ `pkg/services/transfer/resumable_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/resumable_integration_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/streaming_iterator_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/memory_optimized_test.go.disabled` - Fixed

The main issues in these files are:

1. Client initialization pattern - these tests use the old pattern:
   ```go
   client := NewClient("fake-token", WithBaseURL(server.URL+"/"))
   ```
   
   Need to update to the new pattern:
   ```go
   authorizer := &testAuthorizer{token: "fake-token"}
   client, err := NewClient(
       WithAuthorizer(authorizer),
       WithCoreOptions(core.WithBaseURL(server.URL+"/")),
   )
   ```

2. Need to add proper DATA_TYPE field to all TransferItem, DeleteItem, etc. structs in these tests
   
3. Update field references (e.g., `.DATA` -> `.Data`) for proper case sensitivity

## 2. Fix Import Cycle Issues - ✅ Completed

The main import cycle issues have been resolved by:

1. **Core Interface Dependencies**
   - ✅ Created `pkg/core/interfaces` package with interface definitions
   - ✅ Added adapter implementations for interface verification
   - ✅ Updated code to use interfaces rather than concrete implementations

2. **Service Dependencies**
   - ✅ Updated auth service integration tests to use the new pattern
   - ✅ Fixed all import cycles by using proper interfaces and dependency inversion

## 3. Fix Duplication Issues

1. Resolve duplicate type definitions and method implementations:
   - DeleteItem in test_helpers.go vs models.go
   - Ensure consistent naming and field structure for all types

2. Fix test helpers to use consistent naming and avoid conflicts:
   - Prefix all test-specific types with "Test..." to avoid collision
   - Move test helpers to a separate testutils package if needed

## 4. Integration Testing - In Progress

1. ✅ Updated auth integration tests to work with the new interface pattern
2. ⏳ Update transfer integration tests to work with the new interface pattern
3. ⏳ Implement the credential verification checks in the integration test script
4. ⏳ Create proper environment setup for running integration tests

## Implementation Strategy

1. ✅ Fix the Transfer package tests first, as this is the most critical service
2. ✅ Fix import cycles by implementing proper interfaces
3. ⏳ Update integration testing infrastructure 
4. ⏳ Address any remaining duplications
5. ⏳ Verify all tests pass and clean up any remaining issues

This approach has allowed us to make significant progress toward the v0.8.0 release incrementally, with each step improving the build and test process for the SDK.